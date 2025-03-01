package llm

import "fmt"

// LLMProviderFunc is a function type for creating LLM providers
type LLMProviderFunc func(provider string) (LLMProvider, error)

// DefaultLLMProviderFunc is the default implementation
var DefaultLLMProviderFunc LLMProviderFunc = func(provider string) (LLMProvider, error) {
	switch provider {
	case "openai":
		return NewOpenAIClient()
	case "anthropic":
		return NewAnthropicClient()
	case "ollama":
		return nil, fmt.Errorf("ollama provider not implemented yet")
	case "sonar", "sonar-small-online", "sonar-medium-online", "sonar-large-online":
		// Default to OpenAI for sonar models
		return NewOpenAIClient()
	default:
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
