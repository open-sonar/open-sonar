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

type OpenAIClient struct {
	apiKey string
	model  string
}

type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message      openaiMessage `json:"message"`
		FinishReason string        `json:"finish_reason"`
	} `json:"choices"`
}

func NewOpenAIClient() (*OpenAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-3.5-turbo" // Default model
	}

	return &OpenAIClient{
		apiKey: apiKey,
		model:  model,
	}, nil
}

func (c *OpenAIClient) GenerateResponse(query string) (string, error) {
	options := DefaultLLMOptions()
	messages := []string{fmt.Sprintf("user: %s", query)}
	return c.GenerateResponseWithOptions(messages, options)
}

func (c *OpenAIClient) GenerateResponseWithOptions(messages []string, options LLMOptions) (string, error) {
	var openaiMessages []openaiMessage
	for _, message := range messages {
		parts := strings.SplitN(message, ": ", 2)
		if len(parts) < 2 {
			openaiMessages = append(openaiMessages, openaiMessage{
				Role:    "user",
				Content: message,
			})
			continue
		}

		role := parts[0]
		content := parts[1]

		switch role {
		case "system", "user", "assistant":
		default:
			role = "user"
		}

		openaiMessages = append(openaiMessages, openaiMessage{
			Role:    role,
			Content: content,
		})
	}

	reqBody := openaiRequest{
		Model:       c.model,
		Messages:    openaiMessages,
		MaxTokens:   options.MaxTokens,
		Temperature: options.Temperature,
		TopP:        options.TopP,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshalling request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err == nil {
			return "", fmt.Errorf("OpenAI API error: %v", errorResponse)
		}
		return "", fmt.Errorf("OpenAI API error: %s", resp.Status)
	}

	var result openaiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return result.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) CountTokens(text string) (int, error) {
	return utils.SimpleTokenCount(text), nil
}
