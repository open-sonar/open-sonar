package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/tmc/langchaingo/llms/openai"
)

type OpenAIClient struct {
	client *openai.LLM
}

func NewOpenAIClient() (*OpenAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("missing OpenAI API key in environment variables")
	}

	llm, err := openai.New(openai.WithToken(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize OpenAI LLM: %v", err)
	}

	return &OpenAIClient{client: llm}, nil
}

func (o *OpenAIClient) GenerateResponse(query string) (string, error) {
	ctx := context.Background()

	resp, err := o.client.Call(ctx, query)
	if err != nil {
		return "", fmt.Errorf("OpenAI call failed: %v", err)
	}
	return resp, nil
}
