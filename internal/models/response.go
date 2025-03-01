package models

// ChatCompletionResponse represents a response from the chat completions API
type ChatCompletionResponse struct {
	ID        string   `json:"id"`
	Model     string   `json:"model"`
	Object    string   `json:"object"`
	Created   int64    `json:"created"`
	Citations []string `json:"citations"`
	Choices   []Choice `json:"choices"`
	Usage     Usage    `json:"usage"`
}

// Choice represents a completion choice
type Choice struct {
	Index        int     `json:"index"`
	FinishReason string  `json:"finish_reason"`
	Message      Message `json:"message"`
}

// Usage represents token usage statistics
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
