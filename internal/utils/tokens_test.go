package utils

import "testing"

func TestSimpleTokenCount(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{
			name:     "empty string",
			text:     "",
			expected: 0,
		},
		{
			name:     "single word",
			text:     "hello",
			expected: 1,
		},
		{
			name:     "multiple words",
			text:     "hello world how are you",
			expected: 7, // 5 words * 1.33 ≈ 7
		},
		{
			name:     "with punctuation",
			text:     "hello, world! how are you?",
			expected: 7, // 5 words * 1.33 ≈ 7
		},
		{
			name:     "long text",
			text:     "This is a longer text that should be counted as approximately twenty tokens when using our simple token counting algorithm.",
			expected: 20, // 15 words * 1.33 ≈ 20
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := SimpleTokenCount(tt.text)
			if count != tt.expected {
				t.Errorf("Expected %d tokens, got %d for text: %s", tt.expected, count, tt.text)
			}
		})
	}
}
