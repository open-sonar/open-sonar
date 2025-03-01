package utils

import (
	"regexp"
	"testing"
)

func TestGenerateUUID(t *testing.T) {
	// UUID v4 format: 8-4-4-4-12 hex digits
	uuidPattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	// Generate multiple UUIDs to ensure they're unique and correctly formatted
	uuids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		uuid := GenerateUUID()

		// Verify format
		if !uuidPattern.MatchString(uuid) {
			t.Errorf("UUID %s doesn't match expected format", uuid)
		}

		// Verify uniqueness
		if uuids[uuid] {
			t.Errorf("Duplicate UUID generated: %s", uuid)
		}
		uuids[uuid] = true
	}
}
