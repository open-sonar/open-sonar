package llm

import (
	"fmt"
	"strings"
)

type LLMProviderFunc func(provider string) (LLMProvider, error)

// registry holds the mapping from provider names to constructor functions.
var registry = map[string]LLMProviderFunc{}

// allows new providers to be registered dynamically
func RegisterProvider(name string, constructor LLMProviderFunc) {
	registry[name] = constructor
}

func GetProvider(provider string) (LLMProvider, error) {
	if constructor, ok := registry[provider]; ok {
		return constructor(provider)
	}

	// Fallback
	if strings.HasPrefix(provider, "deepseek") ||
		strings.Contains(provider, ":") ||
		strings.Contains(provider, "llama") {
		if constructor, ok := registry["ollama"]; ok {
			return constructor("ollama")
		}
		return nil, fmt.Errorf("provider looks like an Ollama model but 'ollama' not registered")
	}

	return nil, fmt.Errorf("unsupported or unregistered provider: %s", provider)
}
