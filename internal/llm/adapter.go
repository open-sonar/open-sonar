package llm

// LLMOptions represents options for LLM generation
type LLMOptions struct {
	MaxTokens        int
	Temperature      float64
	TopP             float64
	TopK             int
	PresencePenalty  float64
	FrequencyPenalty float64
}

// DefaultLLMOptions returns default LLM options
func DefaultLLMOptions() LLMOptions {
	return LLMOptions{
		MaxTokens:        1024,
		Temperature:      0.2,
		TopP:             0.9,
		TopK:             0,
		PresencePenalty:  0.0,
		FrequencyPenalty: 1.0,
	}
}

// common interface
type LLMProvider interface {
	GenerateResponse(query string) (string, error)
	GenerateResponseWithOptions(messages []string, options LLMOptions) (string, error)
	CountTokens(text string) (int, error)
}
