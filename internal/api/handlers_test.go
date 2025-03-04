package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"open-sonar/internal/models"
)

func TestMain(m *testing.M) {
	// Set test mode environment variable
	os.Setenv("TEST_MODE", "true")

	// Run tests
	exitCode := m.Run()

	// Exit with the same code
	os.Exit(exitCode)
}

func TestChatHandler(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      models.ChatRequest
		expectedStatus   int
		expectedDecision string
	}{
		{
			name: "test direct",
			requestBody: models.ChatRequest{
				Query:      "Bulbasaur",
				NeedSearch: false,
				Pages:      0,
				Retries:    0,
				Provider:   "mock", // Specify the mock provider
			},
			expectedStatus:   http.StatusOK,
			expectedDecision: "direct LLM call",
		},
		{
			name: "test web search",
			requestBody: models.ChatRequest{
				Query:      "Charmander",
				NeedSearch: true,
				Pages:      2,
				Retries:    1,
				Provider:   "mock",
			},
			expectedStatus:   http.StatusOK,
			expectedDecision: "search + LLM call",
		},
		{
			name: "test no query",
			requestBody: models.ChatRequest{
				NeedSearch: true,
				Pages:      2,
				Retries:    1,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "test invalid",
			requestBody:    models.ChatRequest{}, // invalid JSON scenario
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to compile request: %v", err)
			}

			// Simulate a truly invalid JSON payload
			if tt.name == "test invalid" {
				reqBody = []byte(`{invalid json}`)
			}

			req, err := http.NewRequest("POST", "/chat", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(ChatHandler)

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Code == http.StatusOK {
				var resp map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Could not parse success response: %v", err)
				} else {
					// Check for the decision field if expected
					if tt.expectedDecision != "" && resp["decision"] != tt.expectedDecision {
						t.Errorf("Expected decision: %s, got: %s", tt.expectedDecision, resp["decision"])
					}
				}
			} else {
				// parse the structured JSON error
				var errResp struct {
					Error struct {
						Code    int    `json:"code"`
						Message string `json:"message"`
					} `json:"error"`
				}
				if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
					t.Errorf("Expected JSON error response, got: %s", rr.Body.String())
				} else {
					if errResp.Error.Code != rr.Code {
						t.Errorf("Expected error code %d, got %d", rr.Code, errResp.Error.Code)
					}
					if errResp.Error.Message == "" {
						t.Error("Expected a non-empty error message")
					}
				}
			}
		})
	}
}

func TestChatCompletionsHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		bearerToken    string
		expectedStatus int
	}{
		{
			name: "valid request with sonar model",
			requestBody: models.ChatCompletionRequest{
				Model: "mock", // Use mock model for testing
				Messages: []models.Message{
					{Role: "system", Content: "Be helpful."},
					{Role: "user", Content: "How many planets are in the solar system?"},
				},
			},
			bearerToken:    "valid-token",
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing token",
			requestBody: models.ChatCompletionRequest{
				Model: "sonar",
				Messages: []models.Message{
					{Role: "user", Content: "Hello world"},
				},
			},
			bearerToken:    "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "empty messages",
			requestBody: models.ChatCompletionRequest{
				Model:    "sonar",
				Messages: []models.Message{},
			},
			bearerToken:    "valid-token",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			requestBody:    "invalid json",
			bearerToken:    "valid-token",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			var err error

			// If the test scenario is "invalid json"
			switch body := tt.requestBody.(type) {
			case string:
				reqBody = []byte(body)
			default:
				reqBody, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to marshal request body: %v", err)
				}
			}

			req, err := http.NewRequest("POST", "/chat/completions", bytes.NewBuffer(reqBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Add bearer token if provided
			if tt.bearerToken != "" {
				req.Header.Set("Authorization", "Bearer "+tt.bearerToken)
			}

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(ChatCompletionsHandler)

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Code == http.StatusOK {
				var resp models.ChatCompletionResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Errorf("Invalid JSON response: %v", err)
				} else {
					// Check for required fields
					if resp.ID == "" {
						t.Error("Missing ID in response")
					}
					if resp.Model == "" {
						t.Error("Missing Model in response")
					}
					if resp.Object != "chat.completion" {
						t.Errorf("Expected object 'chat.completion', got '%s'", resp.Object)
					}
					if len(resp.Choices) == 0 {
						t.Error("No choices returned in response")
					} else if resp.Choices[0].Message.Role != "assistant" {
						t.Errorf("Expected role 'assistant', got '%s'", resp.Choices[0].Message.Role)
					}
				}
			} else {
				// Error path: parse JSON error
				var errResp struct {
					Error struct {
						Code    int    `json:"code"`
						Message string `json:"message"`
					} `json:"error"`
				}
				if err := json.Unmarshal(rr.Body.Bytes(), &errResp); err != nil {
					t.Errorf("Expected JSON error response, got: %s", rr.Body.String())
				} else {
					if errResp.Error.Code != rr.Code {
						t.Errorf("Expected error code %d, got %d", rr.Code, errResp.Error.Code)
					}
					if errResp.Error.Message == "" {
						t.Error("Expected a non-empty error message")
					}
				}
			}
		})
	}
}
