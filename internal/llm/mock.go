package llm

import (
	"fmt"
	"os"
	"strings"

	"open-sonar/internal/utils"
)

// MockLLMProvider provides mock responses for testing
type MockLLMProvider struct{}

// NewMockLLMProvider creates a new mock LLM provider
func NewMockLLMProvider() (*MockLLMProvider, error) {
	return &MockLLMProvider{}, nil
}

// GenerateResponse returns a canned response
func (p *MockLLMProvider) GenerateResponse(query string) (string, error) {
	if strings.Contains(query, "ERROR_TEST") {
		return "", fmt.Errorf("mock error for testing")
	}
	return fmt.Sprintf("This is a mock response from deepseek-r1:1.5b to your query about: %s", query), nil
}

// GenerateResponseWithOptions returns a canned response with options
func (p *MockLLMProvider) GenerateResponseWithOptions(messages []string, options LLMOptions) (string, error) {
	if len(messages) == 0 {
		return "No input provided", nil
	}

	// Check for error testing case
	for _, msg := range messages {
		if strings.Contains(msg, "ERROR_TEST") {
			return "", fmt.Errorf("mock error for testing")
		}
	}

	// Extract user query from messages
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

	// Include message content
	response := fmt.Sprintf("This is a mock response from deepseek-r1:1.5b to '%s'.", userQuery)

	return response, nil
}

// CountTokens returns a simple token count
func (p *MockLLMProvider) CountTokens(text string) (int, error) {
	return utils.SimpleTokenCount(text), nil
}

// Initialize mock provider for testing
func init() {
	// Override the provider function in test mode
	if os.Getenv("TEST_MODE") == "true" {
		// Replace the provider function with one that returns our mock
		SetLLMProvider(func(provider string) (LLMProvider, error) {
			return NewMockLLMProvider()
		})

		utils.Info("LLM provider mocked for testing")
	}
}
