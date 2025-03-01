package llm

import (
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestOllamaIntegration(t *testing.T) {
	// Skip this test if running in TEST_MODE
	if os.Getenv("TEST_MODE") == "true" {
		t.Skip("Skipping integration test in TEST_MODE")
	}

	// Check if Ollama server is available
	ollamaHost := os.Getenv("OLLAMA_HOST")
	if ollamaHost == "" {
		ollamaHost = "http://localhost:11434"
	}

	resp, err := http.Get(ollamaHost + "/api/version")
	if err != nil || resp.StatusCode != 200 {
		t.Skip("Ollama server not available, skipping integration test")
	}

	// Set environment variable for test
	os.Setenv("OLLAMA_MODEL", "deepseek-r1:1.5b")

	// Initialize the Ollama client
	client, err := NewOllamaClient()
	if err != nil {
		t.Fatalf("Failed to initialize Ollama client: %v", err)
	}

	// Test simple generation
	t.Run("SimpleGeneration", func(t *testing.T) {
		query := "What is the capital of France? Answer in one word."

		// Use a timeout to prevent test from hanging
		resultCh := make(chan string, 1)
		errCh := make(chan error, 1)

		go func() {
			result, err := client.GenerateResponse(query)
			if err != nil {
				errCh <- err
				return
			}
			resultCh <- result
		}()

		// Wait with timeout
		select {
		case result := <-resultCh:
			t.Logf("Ollama response: %s", result)

			// Check if response looks reasonable (contains "Paris")
			if !strings.Contains(strings.ToLower(result), "paris") {
				t.Errorf("Expected response to contain 'Paris', got: %s", result)
			}
		case err := <-errCh:
			t.Fatalf("Error during generation: %v", err)
		case <-time.After(30 * time.Second):
			t.Fatal("Test timed out waiting for Ollama response")
		}
	})

	// Test chat completion with options
	t.Run("ChatCompletionWithOptions", func(t *testing.T) {
		messages := []string{
			"system: You are a helpful, concise assistant.",
			"user: What is the tallest mountain in the world? Just name it.",
		}

		options := LLMOptions{
			Temperature: 0.2,
			MaxTokens:   50,
		}

		// Use a timeout to prevent test from hanging
		resultCh := make(chan string, 1)
		errCh := make(chan error, 1)

		go func() {
			result, err := client.GenerateResponseWithOptions(messages, options)
			if err != nil {
				errCh <- err
				return
			}
			resultCh <- result
		}()

		// Wait with timeout
		select {
		case result := <-resultCh:
			t.Logf("Ollama response: %s", result)

			// Check if response contains expected content
			lowercaseResult := strings.ToLower(result)
			if !strings.Contains(lowercaseResult, "everest") {
				t.Errorf("Expected response to contain 'Everest', got: %s", result)
			}
		case err := <-errCh:
			t.Fatalf("Error during generation with options: %v", err)
		case <-time.After(30 * time.Second):
			t.Fatal("Test timed out waiting for Ollama response")
		}
	})
}
