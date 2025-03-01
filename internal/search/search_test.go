package search

import (
	"testing"

	"open-sonar/internal/search/webscrape"
)

func TestSearch(t *testing.T) {
	// Save original function to restore later
	originalGetSearchProvider := webscrape.CurrentGetSearchProvider

	// Override with mock provider for testing
	webscrape.SetGetSearchProvider(func(provider string) (webscrape.SearchProvider, error) {
		return &webscrape.MockSearchProvider{}, nil
	})

	// Restore the original function when the test completes
	defer func() {
		webscrape.CurrentGetSearchProvider = originalGetSearchProvider
	}()

	// Test cases
	testCases := []struct {
		name          string
		query         string
		options       SearchOptions
		expectedCount int
	}{
		{
			name:  "Basic search",
			query: "test query",
			options: SearchOptions{
				MaxPages:   1,
				MaxRetries: 1,
			},
			expectedCount: 3, // MockSearchProvider returns 3 results
		},
		{
			name:  "Empty search",
			query: "empty",
			options: SearchOptions{
				MaxPages:   1,
				MaxRetries: 1,
			},
			expectedCount: 0,
		},
		{
			name:  "With domain filter",
			query: "test query",
			options: SearchOptions{
				MaxPages:      1,
				MaxRetries:    1,
				DomainFilters: []string{"wikipedia.org"},
			},
			expectedCount: 1, // Only Wikipedia result should be included
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := Search(tc.query, tc.options)

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(results) != tc.expectedCount {
				t.Errorf("Expected %d results, got %d", tc.expectedCount, len(results))
			}
		})
	}
}
