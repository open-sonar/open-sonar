package llm

import (
	"fmt"
	"strings"
)

// LLMProviderFunc is a function type for creating LLM providers
type LLMProviderFunc func(provider string) (LLMProvider, error)

// DefaultLLMProviderFunc is the default implementation
var DefaultLLMProviderFunc LLMProviderFunc = func(provider string) (LLMProvider, error) {
	switch provider {
	case "openai":
		return NewOpenAIClient()
	case "anthropic":
		return NewAnthropicClient()
	case "ollama", "":
		// Default to Ollama if no provider specified
		return NewOllamaClient()
	case "sonar", "sonar-small-online", "sonar-medium-online", "sonar-large-online":
		// Use Ollama for sonar models
		return NewOllamaClient()
	case "mock":
		// Only used in tests
		return NewMockLLMProvider()
	default:
		// Check if this is an Ollama model name
		if strings.HasPrefix(provider, "deepseek") ||
			strings.Contains(provider, ":") || // Catches format like "model:tag"
			strings.Contains(provider, "llama") {
			// Treat as an Ollama model name
			return NewOllamaClient()
		}

		return nil, fmt.Errorf("unsupported LLM provider: %s", provider)
	}
}

// CurrentLLMProviderFunc holds the current provider implementation
var CurrentLLMProviderFunc LLMProviderFunc = DefaultLLMProviderFunc

// NewLLMProvider returns an LLM provider based on provider name
func NewLLMProvider(provider string) (LLMProvider, error) {
	return CurrentLLMProviderFunc(provider)
}

// SetLLMProvider allows tests to replace the provider lookup function
func SetLLMProvider(fn LLMProviderFunc) LLMProviderFunc {
	old := CurrentLLMProviderFunc
	CurrentLLMProviderFunc = fn
	return old
}

// RestoreDefaultLLMProvider restores the default provider lookup
func RestoreDefaultLLMProvider() {
	CurrentLLMProviderFunc = DefaultLLMProviderFunc
}
