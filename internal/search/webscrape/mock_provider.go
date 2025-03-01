package webscrape

import "time"

// MockSearchProvider is a search provider that returns mock results for testing
type MockSearchProvider struct{}

// Search returns mock search results
func (p *MockSearchProvider) Search(query string, options SearchOptions) ([]PageInfo, error) {
	mockTime := time.Now()

	switch query {
	case "empty":
		return []PageInfo{}, nil
	case "error":
		return nil, mockSearchError("mock search error")
	default:
		return []PageInfo{
			{
				URL:       "https://example.com/result1",
				Title:     "Example Result 1",
				Content:   "This is the first example result for " + query,
				Summary:   "First result summary about " + query,
				Published: mockTime,
			},
			{
				URL:       "https://example.org/result2",
				Title:     "Example Result 2",
				Content:   "This is the second example result for " + query,
				Summary:   "Second result summary about " + query,
				Published: mockTime.Add(-24 * time.Hour),
			},
			{
				URL:       "https://wikipedia.org/wiki/" + query,
				Title:     query + " - Wikipedia",
				Content:   "Wikipedia article about " + query,
				Summary:   "Encyclopedia entry for " + query,
				Published: mockTime.Add(-30 * 24 * time.Hour),
			},
		}, nil
	}
}

// mockSearchError is a custom error type for search failures
type mockSearchError string

func (m mockSearchError) Error() string {
	return string(m)
}
