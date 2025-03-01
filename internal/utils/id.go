package utils

import (
	"crypto/rand"
	"fmt"
)

// GenerateUUID generates a simple UUID v4
func GenerateUUID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fall back to a hardcoded format if random fails
		return "00000000-0000-4000-0000-000000000000"
	}

	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0F) | 0x40 // Version 4
	b[8] = (b[8] & 0x3F) | 0x80 // Variant 1

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
