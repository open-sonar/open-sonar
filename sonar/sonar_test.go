package sonar_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"open-sonar/internal/models"
	"open-sonar/internal/utils"
	"open-sonar/sonar"
)

const TEST_PORT = 9876

// TestSonarPackage is a comprehensive test of the sonar package functionality
func TestSonarPackage(t *testing.T) {
	// Skip test if SKIP_OLLAMA_TESTS is set
	if os.Getenv("SKIP_OLLAMA_TESTS") == "true" {
		t.Skip("Skipping Ollama-dependent test")
	}

	// Check if Ollama is available
	if !isOllamaRunning() {
		t.Skip("Ollama is not running, skipping test")
	}

	// Create a test server with explicit configuration
	server := sonar.NewServer(
		sonar.WithPort(TEST_PORT),
		sonar.WithLogLevel(utils.InfoLevel), // This is already correct
		sonar.WithOllama("deepseek-r1:1.5b", "http://localhost:11434"),
		sonar.WithAuthToken("test-token"),
		sonar.WithoutEnvFile(),
	)

	// Start the server
	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Ensure server is running
	if !server.IsRunning() {
		t.Fatal("Server reports not running after Start()")
	}

	// Allow time for server to initialize
	time.Sleep(500 * time.Millisecond)

	// Create a client to access the API
	client := sonar.NewClient(fmt.Sprintf("http://localhost:%d", TEST_PORT), "test-token")

	// Test the /chat endpoint
	t.Run("Simple Chat API", func(t *testing.T) {
		response, err := client.Chat("What is 2+2?", false) // No search for simple questions
		if err != nil {
			t.Fatalf("Chat request failed: %v", err)
		}

		// Verify response structure
		if _, ok := response["decision"]; !ok {
			t.Error("Response missing 'decision' field")
		}
		if _, ok := response["response"]; !ok {
			t.Error("Response missing 'response' field")
		}
	})

	// Test the Perplexity API compatible endpoint
	t.Run("Perplexity API compatible endpoint", func(t *testing.T) {
		// Create a request similar to Perplexity API format
		request := models.ChatCompletionRequest{
			Model: "sonar",
			Messages: []models.Message{
				{Role: "system", Content: "Be precise and concise."},
				{Role: "user", Content: "What is quantum computing in simple terms?"},
			},
			Temperature:      sonar.PtrFloat64(0.2), // Use the exported helper
			MaxTokens:        100,
			TopP:             sonar.PtrFloat64(0.9),
			FrequencyPenalty: sonar.PtrFloat64(1.0),
		}

		// Send the request
		resp, err := client.ChatCompletions(request)
		if err != nil {
			t.Fatalf("ChatCompletions request failed: %v", err)
		}

		// Verify response format matches Perplexity API
		verifyPerplexityAPIFormat(t, resp)
	})

	// Test web search integration
	t.Run("Web search enhanced query", func(t *testing.T) {
		request := models.ChatCompletionRequest{
			Model: "sonar", // sonar model triggers web search
			Messages: []models.Message{
				{Role: "user", Content: "Who won the most recent Super Bowl?"},
			},
			SearchDomainFilter: []string{".com"},
			Temperature:        sonar.PtrFloat64(0.3), // Use the exported helper
			MaxTokens:          150,
		}

		resp, err := client.ChatCompletions(request)
		if err != nil {
			t.Fatalf("Web search request failed: %v", err)
		}

		// Verify response format
		verifyPerplexityAPIFormat(t, resp)

		// In web search mode, there should be citations
		if resp.Citations == nil {
			t.Log("Warning: No citations found in web search response")
		}
	})

	// Test raw HTTP call to match exact Perplexity API format
	t.Run("Raw HTTP Perplexity API format", func(t *testing.T) {
		// Create request body exactly like Perplexity's API
		reqBody := `{
			"model": "sonar",
			"messages": [
				{"role": "system", "content": "Be precise and concise."},
				{"role": "user", "content": "How many planets are in our solar system?"}
			],
			"max_tokens": 123,
			"temperature": 0.2,
			"top_p": 0.9,
			"stream": false,
			"presence_penalty": 0,
			"frequency_penalty": 1
		}`

		// Create HTTP request
		req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%d/chat/completions", TEST_PORT),
			strings.NewReader(reqBody))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Add headers
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer test-token")

		// Send request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		// Check status code
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var result models.ChatCompletionResponse
		err = json.NewDecoder(resp.Body).Decode(&result)
		if err != nil {
			t.Fatalf("Failed to parse response: %v", err)
		}

		// Verify response format
		verifyPerplexityAPIFormat(t, &result)
	})

	// Stop the server
	err = server.Stop()
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}

	// Ensure server is stopped
	if server.IsRunning() {
		t.Fatal("Server reports running after Stop()")
	}
}

// Test helper functions
func verifyPerplexityAPIFormat(t *testing.T, resp *models.ChatCompletionResponse) {
	// Check required fields
	if resp.ID == "" {
		t.Error("Response missing ID")
	}
	if resp.Model == "" {
		t.Error("Response missing Model")
	}
	if resp.Object != "chat.completion" {
		t.Errorf("Expected object 'chat.completion', got '%s'", resp.Object)
	}
	if resp.Created == 0 {
		t.Error("Response missing Created timestamp")
	}

	// Check choices
	if len(resp.Choices) == 0 {
		t.Error("Response has no choices")
	} else {
		choice := resp.Choices[0]
		if choice.Message.Role != "assistant" {
			t.Errorf("Expected role 'assistant', got '%s'", choice.Message.Role)
		}
		if choice.Message.Content == "" {
			t.Error("Response content is empty")
		}
		if choice.FinishReason == "" {
			t.Error("Response missing finish_reason")
		}
	}

	// Check usage
	if resp.Usage.PromptTokens == 0 {
		t.Error("Response missing prompt_tokens")
	}
	if resp.Usage.CompletionTokens == 0 {
		t.Error("Response missing completion_tokens")
	}
	if resp.Usage.TotalTokens == 0 {
		t.Error("Response missing total_tokens")
	}
}

func isOllamaRunning() bool {
	resp, err := http.Get("http://localhost:11434/api/version")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
