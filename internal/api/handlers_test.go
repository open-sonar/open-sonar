package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"open-sonar/internal/models"
)

func TestChatHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    models.ChatRequest
		expectedStatus int
		expectedDecision string
	}{
		{
			name: "test direct",
			requestBody: models.ChatRequest{
				Query:      "Bulbasaur",
				NeedSearch: false,
				Pages:      0,
				Retries:    0,
			},
			expectedStatus:  http.StatusOK,
			expectedDecision: "direct LLM call",
		},
		{
			name: "test web search",
			requestBody: models.ChatRequest{
				Query:      "Charmander",
				NeedSearch: true,
				Pages:      2,
				Retries:    1,
			},
			expectedStatus:  http.StatusOK,
			expectedDecision: "search + LLM call",
		},
		{
			name: "test no query",
			requestBody: models.ChatRequest{
				NeedSearch: true,
				Pages:      2,
				Retries:    1,
			},
			expectedStatus:  http.StatusBadRequest,
		},
		{
			name: "test invalid",
			requestBody: models.ChatRequest{}, // invalid JSON
			expectedStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to compile request: %v", err)
			}

			if tt.name == "Invalid JSON" {
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

			var resp map[string]interface{}
			if err := json.Unmarshal(rr.Body.Bytes(), &resp); err == nil {
				if tt.expectedDecision != "" && resp["decision"] != tt.expectedDecision {
					t.Errorf("Expected decision: %s, got: %s", tt.expectedDecision, resp["decision"])
				}
			} else {
				t.Logf("Couldn't parse response: %s", rr.Body.String())
			}
		})
	}
}
