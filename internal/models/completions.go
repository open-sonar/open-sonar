package models

// ChatCompletionRequest is the request for chat completions
type ChatCompletionRequest struct {
	Model               string    `json:"model"`
	Messages            []Message `json:"messages"`
	Temperature         *float64  `json:"temperature,omitempty"`
	TopP                *float64  `json:"top_p,omitempty"`
	TopK                int       `json:"top_k,omitempty"`
	MaxTokens           int       `json:"max_tokens,omitempty"`
	Stream              bool      `json:"stream,omitempty"`
	PresencePenalty     *float64  `json:"presence_penalty,omitempty"`
	FrequencyPenalty    *float64  `json:"frequency_penalty,omitempty"`
	SearchDomainFilter  []string  `json:"search_domain_filter,omitempty"`
	SearchRecencyFilter string    `json:"search_recency_filter,omitempty"`
	ResponseFormat      *string   `json:"response_format,omitempty"`
}

// ChatCompletionResponse is the response object for chat completions
type ChatCompletionResponse struct {
	ID        string   `json:"id"`
	Model     string   `json:"model"`
	Object    string   `json:"object"`
	Created   int64    `json:"created"`
	Citations []string `json:"citations"` // Make sure this is properly tagged
	Choices   []Choice `json:"choices"`
	Usage     Usage    `json:"usage"`
}

// Message represents an individual message in the conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Choice represents a generation choice
type Choice struct {
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
	Delta        *Delta  `json:"delta,omitempty"`
}

// Delta represents incremental content when streaming
type Delta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// Usage contains token statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// VerifyCitations checks if citations are properly included
func (resp *ChatCompletionResponse) VerifyCitations() bool {
	return resp.Citations != nil && len(resp.Citations) > 0
}
