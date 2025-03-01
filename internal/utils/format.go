package utils

import (
	"fmt"
	"strings"
)

// FormatInt formats an integer to a string with comma separators
func FormatInt(n int) string {
	// Simple implementation for numbers under 1000
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}

	// For larger numbers, add comma separators
	s := fmt.Sprintf("%d", n)
	result := ""
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}
	return result
}

// TruncateText truncates text to specified length with ellipsis
func TruncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return strings.TrimSpace(text[:maxLength]) + "..."
}

// SortScored sorts a slice by the given comparison function
func SortScored[T any](items []T, less func(i, j int) bool) {
	n := len(items)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if !less(j, j+1) {
				items[j], items[j+1] = items[j+1], items[j]
			}
		}
	}
}
