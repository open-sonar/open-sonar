package llm

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"open-sonar/internal/utils"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// OllamaClient implements the LLMProvider interface using Ollama
type OllamaClient struct {
	model string
	llm   *ollama.LLM
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient() (*OllamaClient, error) {
	// Check for environment variables
	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "deepseek-r1:1.5b" // Default model
	}

	// Get base URL from environment if provided
	baseURL := os.Getenv("OLLAMA_HOST")
	if baseURL == "" {
		baseURL = "http://localhost:11434" // Default Ollama server
	}

	// Create options
	options := []ollama.Option{
		ollama.WithModel(model),
		ollama.WithServerURL(baseURL),
	}

	// Initialize the Ollama client
	llm, err := ollama.New(options...)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Ollama client: %w", err)
	}

	return &OllamaClient{
		model: model,
		llm:   llm,
	}, nil
}

// GenerateResponse generates a response to the given query
func (c *OllamaClient) GenerateResponse(query string) (string, error) {
	// Create timer for performance tracking
	timer := utils.NewTimer("Ollama-GenerateResponse")
	defer timer.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	completion, err := llms.GenerateFromSinglePrompt(ctx, c.llm, query)
	if err != nil {
		return "", fmt.Errorf("Ollama generation failed: %w", err)
	}

	// Return the trimmed completion
	return strings.TrimSpace(completion), nil
}

// GenerateResponseWithOptions generates a response with the given options
func (c *OllamaClient) GenerateResponseWithOptions(messages []string, options LLMOptions) (string, error) {
	// Create timer for performance tracking
	timer := utils.NewTimer("Ollama-GenerateResponseWithOptions")
	defer timer.Stop()

	// Convert our messages to a simplified prompt
	prompt := formatMessagesToPrompt(messages)

	// Set Ollama options
	ctx := context.Background()
	ollamaOpts := []llms.CallOption{
		llms.WithTemperature(options.Temperature),
		llms.WithTopP(options.TopP),
	}

	if options.MaxTokens > 0 {
		ollamaOpts = append(ollamaOpts, llms.WithMaxTokens(options.MaxTokens))
	}

	// Generate the completion - using GenerateFromSinglePrompt because Generate doesn't exist
	completion, err := llms.GenerateFromSinglePrompt(ctx, c.llm, prompt, ollamaOpts...)
	if err != nil {
		return "", fmt.Errorf("Ollama generation failed: %w", err)
	}

	// Return the trimmed completion
	return strings.TrimSpace(completion), nil
}

// formatMessagesToPrompt formats message array into a prompt string
func formatMessagesToPrompt(messages []string) string {
	var formattedMessages []string

	for _, message := range messages {
		parts := strings.SplitN(message, ": ", 2)
		if len(parts) == 2 {
			role := parts[0]
			content := parts[1]

			switch role {
			case "system":
				formattedMessages = append(formattedMessages, fmt.Sprintf("[SYSTEM]\n%s\n", content))
			case "user":
				formattedMessages = append(formattedMessages, fmt.Sprintf("[USER]\n%s\n", content))
			case "assistant":
				formattedMessages = append(formattedMessages, fmt.Sprintf("[ASSISTANT]\n%s\n", content))
			default:
				formattedMessages = append(formattedMessages, content)
			}
		} else {
			formattedMessages = append(formattedMessages, message)
		}
	}

	// Add final [ASSISTANT] prefix to prompt for completion
	formattedMessages = append(formattedMessages, "[ASSISTANT]\n")

	return strings.Join(formattedMessages, "\n")
}

// CountTokens estimates the number of tokens in the given text
func (c *OllamaClient) CountTokens(text string) (int, error) {
	// Using simple approximation since Ollama doesn't have a tokenization API
	return utils.SimpleTokenCount(text), nil
}
