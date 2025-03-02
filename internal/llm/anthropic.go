package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"open-sonar/internal/utils"
)

// AnthropicClient implements the LLM provider interface for Anthropic.
type AnthropicClient struct {
	apiKey string
	model  string
}

type anthropicRequest struct {
	Model       string             `json:"model"`
	Messages    []anthropicMessage `json:"messages"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
	TopP        float64            `json:"top_p,omitempty"`
	TopK        int                `json:"top_k,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// init registers the Anthropic provider with the pluggable registry.
func init() {
	RegisterProvider("anthropic", func(_ string) (LLMProvider, error) {
		return NewAnthropicClient()
	})
}

// NewAnthropicClient creates a new Anthropic client.
func NewAnthropicClient() (*AnthropicClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = "claude-3-opus-20240229" // Default model.
	}

	return &AnthropicClient{
		apiKey: apiKey,
		model:  model,
	}, nil
}

// GenerateResponse implements the LLMProvider interface.
func (c *AnthropicClient) GenerateResponse(query string) (string, error) {
	options := DefaultLLMOptions()
	messages := []string{fmt.Sprintf("user: %s", query)}
	return c.GenerateResponseWithOptions(messages, options)
}

// GenerateResponseWithOptions implements the LLMProvider interface with options.
func (c *AnthropicClient) GenerateResponseWithOptions(messages []string, options LLMOptions) (string, error) {
	// Convert messages format.
	var anthropicMessages []anthropicMessage
	for _, message := range messages {
		parts := strings.SplitN(message, ": ", 2)
		if len(parts) < 2 {
			// Handle malformed messages.
			anthropicMessages = append(anthropicMessages, anthropicMessage{
				Role:    "user",
				Content: message,
			})
			continue
		}

		role := parts[0]
		content := parts[1]

		// Convert to Anthropic expected roles.
		switch role {
		case "system":
			// Skip system message - add it at the beginning of the first user message.
			if len(anthropicMessages) > 0 && anthropicMessages[0].Role == "user" {
				anthropicMessages[0].Content = content + "\n\n" + anthropicMessages[0].Content
			} else {
				anthropicMessages = append(anthropicMessages, anthropicMessage{
					Role:    "user",
					Content: content,
				})
			}
			continue
		case "assistant":
			role = "assistant"
		case "user":
			role = "user"
		default:
			role = "user" // Default to user for unknown roles.
		}

		anthropicMessages = append(anthropicMessages, anthropicMessage{
			Role:    role,
			Content: content,
		})
	}

	// Create request.
	reqBody := anthropicRequest{
		Model:       c.model,
		Messages:    anthropicMessages,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		TopP:        options.TopP,
		TopK:        options.TopK,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	// Set headers.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Make the request.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	// Check status.
	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return "", fmt.Errorf("Anthropic API error: %v", errorResponse)
		}
		return "", fmt.Errorf("Anthropic API error: %s", resp.Status)
	}

	// Parse response.
	var result anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("no response from Anthropic")
	}

	// Extract text from response.
	var text string
	for _, content := range result.Content {
		if content.Type == "text" {
			text += content.Text
		}
	}

	return text, nil
}

// CountTokens implements the LLMProvider interface for token counting.
func (c *AnthropicClient) CountTokens(text string) (int, error) {
	// For MVP, use a simple heuristic.
	return utils.SimpleTokenCount(text), nil
}
