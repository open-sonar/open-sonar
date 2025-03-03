package llm

import (
	"fmt"
	"os"
	"strings"

	"open-sonar/internal/utils"
)

// MockLLMProvider provides mock responses for testing.
type MockLLMProvider struct{}

func init() {
	// Register the mock provider in the registry.
	RegisterProvider("mock", func(_ string) (LLMProvider, error) {
		return NewMockLLMProvider()
	})
}

// NewMockLLMProvider creates a new mock LLM provider.
func NewMockLLMProvider() (*MockLLMProvider, error) {
	return &MockLLMProvider{}, nil
}

// GenerateResponse returns a mock response.
func (p *MockLLMProvider) GenerateResponse(query string) (string, error) {
	if strings.Contains(query, "ERROR_TEST") {
		return "", fmt.Errorf("mock error for testing")
	}
	return fmt.Sprintf("Mock response: deepseek-r1:1.5b -> '%s'", query), nil
}

// GenerateResponseWithOptions returns a mock response with options.
func (p *MockLLMProvider) GenerateResponseWithOptions(messages []string, options LLMOptions) (string, error) {
	if len(messages) == 0 {
		return "No input provided", nil
	}

	for _, msg := range messages {
		if strings.Contains(msg, "ERROR_TEST") {
			return "", fmt.Errorf("mock error for testing")
		}
	}

	userQuery := "your query"
	for _, msg := range messages {
		if strings.HasPrefix(msg, "user: ") {
			userQuery = strings.TrimPrefix(msg, "user: ")
			if len(userQuery) > 30 {
				userQuery = userQuery[:30] + "..."
			}
			break
		}
	}

	response := fmt.Sprintf("Mock response from deepseek-r1:1.5b -> '%s'.", userQuery)
	return response, nil
}

// CountTokens returns a simple token count.
func (p *MockLLMProvider) CountTokens(text string) (int, error) {
	return utils.SimpleTokenCount(text), nil
}

func init() {
	// Override the provider function in test mode.
	if os.Getenv("TEST_MODE") == "true" {
		SetLLMProvider(func(provider string) (LLMProvider, error) {
			return NewMockLLMProvider()
		})
		utils.Info("LLM provider mocked for testing")
	}
}
