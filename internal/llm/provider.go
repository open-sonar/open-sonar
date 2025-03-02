package llm

import "fmt"

var overrideProviderFunc func(provider string) (LLMProvider, error)

func NewLLMProvider(provider string) (LLMProvider, error) {
	if overrideProviderFunc != nil {
		return overrideProviderFunc(provider)
	}
	p, err := GetProvider(provider)
	if err != nil {
		return nil, fmt.Errorf("NewLLMProvider: %w", err)
	}
	return p, nil
}

func SetLLMProvider(fn func(provider string) (LLMProvider, error)) func(provider string) (LLMProvider, error) {
	old := overrideProviderFunc
	overrideProviderFunc = fn
	return old
}

// clears any override, restoring the default lookup
func RestoreDefaultLLMProvider() {
	overrideProviderFunc = nil
}
