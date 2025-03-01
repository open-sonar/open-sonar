package models

// ChatRequest represents a request to the chat endpoint
type ChatRequest struct {
	Query      string `json:"query"`
	NeedSearch bool   `json:"needSearch"`
	Pages      int    `json:"pages"`
	Retries    int    `json:"retries"`
	Provider   string `json:"provider"`
}

// Note: Message struct is now only defined in completions.go
