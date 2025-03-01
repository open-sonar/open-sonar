package models

// Message represents a message in a chat conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents a request to the chat completions API
type ChatCompletionRequest struct {
	Model                  string    `json:"model"`
	Messages               []Message `json:"messages"`
	MaxTokens              int       `json:"max_tokens,omitempty"`
	Temperature            *float64  `json:"temperature,omitempty"`
	TopP                   *float64  `json:"top_p,omitempty"`
	SearchDomainFilter     []string  `json:"search_domain_filter,omitempty"`
	ReturnImages           bool      `json:"return_images,omitempty"`
	ReturnRelatedQuestions bool      `json:"return_related_questions,omitempty"`
	SearchRecencyFilter    string    `json:"search_recency_filter,omitempty"`
	TopK                   int       `json:"top_k,omitempty"`
	Stream                 bool      `json:"stream,omitempty"`
	PresencePenalty        *float64  `json:"presence_penalty,omitempty"`
	FrequencyPenalty       *float64  `json:"frequency_penalty,omitempty"`
	ResponseFormat         *any      `json:"response_format,omitempty"`
}

// Legacy request format
type ChatRequest struct {
	Query      string `json:"query"`
	NeedSearch bool   `json:"need_search"`
	Pages      int    `json:"pages"`
	Retries    int    `json:"retries"`
	Provider   string `json:"provider"`
}
