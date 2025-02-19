package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms/anthropic"
)

type AnthropicClient struct {
	client *anthropic.LLM
}

func NewAnthropicClient() (*AnthropicClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing ANTHROPIC_API_KEY")
	}

	c, err := anthropic.New(anthropic.WithToken(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Anthropic client: %v", err)
	}

	return &AnthropicClient{client: c}, nil
}

func (a *AnthropicClient) GenerateResponse(query string) (string, error) {
	resp, err := a.client.Call(context.Background(), query)
	if err != nil {
		return "", fmt.Errorf("Anthropic call failed: %v", err)
	}
	return resp, nil
}
