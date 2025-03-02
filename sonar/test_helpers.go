package sonar

import (
	"open-sonar/internal/models"
)

// CreateTestCompletionRequest creates a test chat completion request
func CreateTestCompletionRequest() models.ChatCompletionRequest {
	return models.ChatCompletionRequest{
		Model: "sonar",
		Messages: []models.Message{
			{Role: "system", Content: "Be helpful and concise."},
			{Role: "user", Content: "What is quantum computing?"},
		},
		Temperature: PtrFloat64(0.2),
		MaxTokens:   150,
	}
}

// PtrFloat64 creates a pointer to a float64 (exported for tests)
func PtrFloat64(v float64) *float64 {
	return &v
}
