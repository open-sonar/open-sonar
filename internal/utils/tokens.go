package utils

import (
	"strings"
)

// SimpleTokenCount provides a very rough estimate of token count
// Note: This is a simplified version; production code should use a proper tokenizer
func SimpleTokenCount(text string) int {
	words := strings.Fields(text)
	// Rough approximation: 1 token ≈ 0.75 words
	return int(float64(len(words)) * 1.33)
}
