package llm

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Set test mode environment variable if not already set
	if os.Getenv("TEST_MODE") == "" {
		os.Setenv("TEST_MODE", "true")
	}

	// Run tests
	exitCode := m.Run()

	// Exit with the same code
	os.Exit(exitCode)
}

func TestMockLLMProvider(t *testing.T) {
	// Ensure we're using the mock provider
	os.Setenv("TEST_MODE", "true")

	provider, err := NewLLMProvider("mock")
	if err != nil {
		t.Fatalf("Failed to create mock provider: %v", err)
	}

	// Test simple generation
	response, err := provider.GenerateResponse("test query")
	if err != nil {
		t.Errorf("Error generating response: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response")
	}

	// Test with options
	messages := []string{
		"system: Be helpful.",
		"user: How many planets are in the solar system?",
	}

	options := DefaultLLMOptions()
	response, err = provider.GenerateResponseWithOptions(messages, options)
	if err != nil {
		t.Errorf("Error generating response with options: %v", err)
	}
	if response == "" {
		t.Error("Expected non-empty response with options")
	}

	// Test error case
	_, err = provider.GenerateResponse("ERROR_TEST")
	if err == nil {
		t.Error("Expected error for ERROR_TEST, got nil")
	}
}

func TestLLMOptions(t *testing.T) {
	options := DefaultLLMOptions()

	// Verify default values
	if options.Temperature != 0.2 {
		t.Errorf("Expected default temperature 0.2, got %f", options.Temperature)
	}
	if options.TopP != 0.9 {
		t.Errorf("Expected default topP 0.9, got %f", options.TopP)
	}
	if options.MaxTokens != 1024 {
		t.Errorf("Expected default maxTokens 1024, got %d", options.MaxTokens)
	}
}
