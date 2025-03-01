package models

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestChatRequestSerialization(t *testing.T) {
	original := ChatRequest{
		Query:      "What is the weather like today?",
		NeedSearch: true,
		Pages:      3,
		Retries:    2,
		Provider:   "openai",
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to serialize ChatRequest: %v", err)
	}

	// Deserialize from JSON
	var deserialized ChatRequest
	err = json.Unmarshal(jsonData, &deserialized)
	if err != nil {
		t.Fatalf("Failed to deserialize ChatRequest: %v", err)
	}

	// Compare original and deserialized
	if !reflect.DeepEqual(original, deserialized) {
		t.Errorf("Deserialized ChatRequest doesn't match original: %+v vs %+v", original, deserialized)
	}
}

func TestChatCompletionRequestSerialization(t *testing.T) {
	temperature := 0.7
	topP := 0.9
	frequencyPenalty := 1.0

	original := ChatCompletionRequest{
		Model: "sonar",
		Messages: []Message{
			{Role: "system", Content: "You are a helpful assistant."},
			{Role: "user", Content: "What is the weather like today?"},
		},
		MaxTokens:              100,
		Temperature:            &temperature,
		TopP:                   &topP,
		SearchDomainFilter:     []string{"weather.com", "-twitter.com"},
		ReturnImages:           false,
		ReturnRelatedQuestions: true,
		SearchRecencyFilter:    "day",
		TopK:                   10,
		Stream:                 false,
		FrequencyPenalty:       &frequencyPenalty,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to serialize ChatCompletionRequest: %v", err)
	}

	// Deserialize from JSON
	var deserialized ChatCompletionRequest
	err = json.Unmarshal(jsonData, &deserialized)
	if err != nil {
		t.Fatalf("Failed to deserialize ChatCompletionRequest: %v", err)
	}

	// Compare original and deserialized
	if original.Model != deserialized.Model ||
		len(original.Messages) != len(deserialized.Messages) ||
		original.MaxTokens != deserialized.MaxTokens ||
		*original.Temperature != *deserialized.Temperature ||
		*original.TopP != *deserialized.TopP ||
		len(original.SearchDomainFilter) != len(deserialized.SearchDomainFilter) ||
		original.ReturnImages != deserialized.ReturnImages ||
		original.ReturnRelatedQuestions != deserialized.ReturnRelatedQuestions ||
		original.SearchRecencyFilter != deserialized.SearchRecencyFilter {
		t.Errorf("Deserialized ChatCompletionRequest doesn't match original")
	}
}

func TestChatCompletionResponseSerialization(t *testing.T) {
	original := ChatCompletionResponse{
		ID:      "resp-123456",
		Model:   "sonar",
		Object:  "chat.completion",
		Created: 1646156262,
		Citations: []string{
			"https://example.com/doc1",
			"https://example.com/doc2",
		},
		Choices: []Choice{
			{
				Index:        0,
				FinishReason: "stop",
				Message: Message{
					Role:    "assistant",
					Content: "The weather today is sunny with a high of 75°F.",
				},
			},
		},
		Usage: Usage{
			PromptTokens:     15,
			CompletionTokens: 12,
			TotalTokens:      27,
		},
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to serialize ChatCompletionResponse: %v", err)
	}

	// Deserialize from JSON
	var deserialized ChatCompletionResponse
	err = json.Unmarshal(jsonData, &deserialized)
	if err != nil {
		t.Fatalf("Failed to deserialize ChatCompletionResponse: %v", err)
	}

	// Compare original and deserialized
	if !reflect.DeepEqual(original, deserialized) {
		t.Errorf("Deserialized ChatCompletionResponse doesn't match original")
	}
}
