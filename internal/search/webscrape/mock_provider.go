package webscrape

import (
	"time"
)

// MockSearchProvider provides deterministic responses for testing
type MockSearchProvider struct{}

// Search provides deterministic search results for testing
func (p *MockSearchProvider) Search(query string, options SearchOptions) ([]PageInfo, error) {
	// For empty query or "empty" query, return empty results
	if query == "" || query == "empty" {
		return []PageInfo{}, nil
	}

	// For domain filter test, return specific result if matching filter
	if len(options.SearchDomainFilter) > 0 {
		if contains(options.SearchDomainFilter, ".gov") {
			// Create a single result that matches the domain filter
			mockTime := time.Now()
			return []PageInfo{
				{
					URL:       "https://example.gov/page",
					Title:     "Government Example",
					Content:   "This is content from a government site.",
					Summary:   "Summary from government site",
					Published: mockTime,
				},
			}, nil
		}
		// Otherwise return empty for domain filter test
		return []PageInfo{}, nil
	}

	// For regular tests, return exactly 3 results as expected by tests
	mockTime := time.Now()
	return []PageInfo{
		{
			URL:       "https://example.com/page1",
			Title:     "Example Page 1",
			Content:   "This is the content of page 1. It contains sample text.",
			Summary:   "Summary of page 1",
			Published: mockTime,
		},
		{
			URL:       "https://example.com/page2",
			Title:     "Example Page 2",
			Content:   "This is the content of page 2. More sample text here.",
			Summary:   "Summary of page 2",
			Published: mockTime.Add(-24 * time.Hour),
		},
		{
			URL:       "https://example.com/page3",
			Title:     "Example Page 3",
			Content:   "This is the content of page 3. Even more sample text.",
			Summary:   "Summary of page 3",
			Published: mockTime.Add(-48 * time.Hour),
		},
	}, nil
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
