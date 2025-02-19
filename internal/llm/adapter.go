package llm

import "fmt"

// common interface
type LLMProvider interface {
	GenerateResponse(query string) (string, error)
}

func NewLLMProvider(provider string) (LLMProvider, error) {
	switch provider {
	case "openai":
		return NewOpenAIClient()
	case "anthropic":
		return NewAnthropicClient()
	case "ollama":
		return nil, fmt.Errorf("ollama provider not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}
