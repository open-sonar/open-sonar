package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"open-sonar/internal/utils"
)

// OllamaProvider implements the LLMProvider interface using the Ollama API.
type OllamaProvider struct {
	model string
	host  string
}

// init registers the Ollama provider (and alias "sonar") with the pluggable registry.
func init() {
	RegisterProvider("ollama", func(_ string) (LLMProvider, error) {
		return NewOllamaClient()
	})
	RegisterProvider("sonar", func(_ string) (LLMProvider, error) {
		return NewOllamaClient()
	})
}

// NewOllamaClient is a factory function that creates a new Ollama provider.
func NewOllamaClient() (LLMProvider, error) {
	host := os.Getenv("OLLAMA_HOST")
	if host == "" {
		host = "http://localhost:11434"
	}

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "deepseek-r1:1.5b" // Default model
	}

	provider := &OllamaProvider{
		model: model,
		host:  host,
	}

	// Verify connection and model availability
	err := provider.verifyModelAvailability(true)
	if err != nil {
		return nil, fmt.Errorf("ollama provider creation failed: %w", err)
	}

	utils.Info(fmt.Sprintf("Ollama provider initialized with model: %s", model))
	return provider, nil
}

// ollamaRequest represents the request structure for the Ollama API.
type ollamaRequest struct {
	Model     string    `json:"model"`
	Prompt    string    `json:"prompt"`
	Stream    bool      `json:"stream"`
	Options   *options  `json:"options,omitempty"`
	System    string    `json:"system,omitempty"`
	Templates []message `json:"templates,omitempty"`
}

type options struct {
	Temperature      float64 `json:"temperature,omitempty"`
	TopP             float64 `json:"top_p,omitempty"`
	TopK             int     `json:"top_k,omitempty"`
	MaxTokens        int     `json:"num_predict,omitempty"`
	PresencePenalty  float64 `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ollamaResponse represents the response structure from the Ollama API.
type ollamaResponse struct {
	Model     string `json:"model"`
	Response  string `json:"response"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at"`
}

// ollamaListResponse represents the response for listing models.
type ollamaListResponse struct {
	Models []ollamaModel `json:"models"`
}

type ollamaModel struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	ModifiedAt string `json:"modified_at"`
}

// ollamaPullResponse represents the pull response from the Ollama registry.
type ollamaPullResponse struct {
	Status string `json:"status"`
	Digest string `json:"digest"`
	Total  int64  `json:"total"`
	Error  string `json:"error"`
}

// GenerateResponse implements the LLMProvider interface.
func (p *OllamaProvider) GenerateResponse(prompt string) (string, error) {
	timer := utils.NewTimer("Ollama-GenerateResponse")
	defer timer.Stop()

	// Create request payload.
	payload := ollamaRequest{
		Model:  p.model,
		Prompt: prompt,
	}

	return p.sendRequest(payload)
}

// GenerateResponseWithOptions implements the LLMProvider interface.
func (p *OllamaProvider) GenerateResponseWithOptions(messages []string, opts LLMOptions) (string, error) {
	timer := utils.NewTimer("Ollama-GenerateResponseWithOptions")
	defer timer.Stop()

	// Extract system message and user prompt.
	var system string
	var prompt string

	for _, msg := range messages {
		if strings.HasPrefix(msg, "system: ") {
			system = strings.TrimPrefix(msg, "system: ")
		} else if strings.HasPrefix(msg, "user: ") {
			prompt = strings.TrimPrefix(msg, "user: ")
		}
	}

	// If no specific user prompt is found, combine all messages.
	if prompt == "" {
		prompt = strings.Join(messages, "\n")
	}

	// Create request payload.
	payload := ollamaRequest{
		Model:  p.model,
		Prompt: prompt,
		System: system,
		Options: &options{
			Temperature:      opts.Temperature,
			TopP:             opts.TopP,
			TopK:             opts.TopK,
			MaxTokens:        opts.MaxTokens,
			PresencePenalty:  opts.PresencePenalty,
			FrequencyPenalty: opts.FrequencyPenalty,
		},
	}

	return p.sendRequest(payload)
}

// CountTokens implements the LLMProvider interface.
func (p *OllamaProvider) CountTokens(text string) (int, error) {
	// Using simple approximation since Ollama doesn't have a tokenization API.
	return utils.SimpleTokenCount(text), nil
}

// sendRequest sends a request to the Ollama API.
func (p *OllamaProvider) sendRequest(payload ollamaRequest) (string, error) {
	// Convert payload to JSON.
	reqBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request.
	req, err := http.NewRequest("POST", p.host+"/api/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request.
	client := &http.Client{
		Timeout: 300 * time.Second, // 5 minutes timeout for long generations.
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response.
	var ollamaResp ollamaResponse
	err = json.NewDecoder(resp.Body).Decode(&ollamaResp)
	if err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for error in response.
	if ollamaResp.Error != "" {
		return "", fmt.Errorf("ollama returned error: %s", ollamaResp.Error)
	}

	return ollamaResp.Response, nil
}

// verifyModelAvailability checks if the model is available in Ollama.
// If autoPull is true, it will attempt to pull the model if not available.
func (p *OllamaProvider) verifyModelAvailability(autoPull bool) error {
	// First, check if we can reach the Ollama server.
	_, err := http.Get(p.host + "/api/version")
	if err != nil {
		return fmt.Errorf("cannot connect to Ollama server at %s: %w", p.host, err)
	}

	// List available models.
	req, err := http.NewRequest("GET", p.host+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}
	defer resp.Body.Close()

	var listResp ollamaListResponse
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	if err != nil {
		return fmt.Errorf("failed to decode model list: %w", err)
	}

	// Check if our model is in the list.
	modelFound := false
	for _, model := range listResp.Models {
		if model.Name == p.model {
			modelFound = true
			break
		}
	}

	// If the model is not found and autoPull is enabled, attempt to pull it.
	if !modelFound && autoPull {
		utils.Warn(fmt.Sprintf("Model %s not found, attempting to pull it...", p.model))
		if err := p.pullModel(); err != nil {
			return fmt.Errorf("failed to pull model %s: %w", p.model, err)
		}
		utils.Info(fmt.Sprintf("Successfully pulled model: %s", p.model))
		return nil
	} else if !modelFound {
		return fmt.Errorf("model %s is not available on this Ollama server", p.model)
	}

	return nil
}

// pullModel pulls a model from the Ollama registry.
func (p *OllamaProvider) pullModel() error {
	// Create pull request.
	pullReq := map[string]string{
		"name": p.model,
	}
	reqBody, err := json.Marshal(pullReq)
	if err != nil {
		return fmt.Errorf("failed to marshal pull request: %w", err)
	}

	// Create request.
	req, err := http.NewRequest("POST", p.host+"/api/pull", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create pull request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request.
	client := &http.Client{
		Timeout: 30 * time.Minute, // Model pulls can take a very long time.
	}

	utils.Info(fmt.Sprintf("Pulling model %s (this may take several minutes)...", p.model))
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send pull request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status.
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pull request failed with status: %d", resp.StatusCode)
	}

	// Pull requests stream responses - log progress.
	decoder := json.NewDecoder(resp.Body)
	for {
		var pullResp ollamaPullResponse
		if err := decoder.Decode(&pullResp); err != nil {
			break // End of stream.
		}

		if pullResp.Error != "" {
			return fmt.Errorf("pull error: %s", pullResp.Error)
		}

		if strings.Contains(pullResp.Status, "success") {
			return nil // Pull completed successfully.
		}
	}

	return nil // Pull completed.
}

// SetModel changes the model used by this provider.
func (p *OllamaProvider) SetModel(model string) error {
	p.model = model
	return p.verifyModelAvailability(true)
}
