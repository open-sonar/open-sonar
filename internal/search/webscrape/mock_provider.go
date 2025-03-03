package webscrape

import (
	"strings"
	"time"
)

type MockSearchProvider struct{}

func (p *MockSearchProvider) Search(query string, options SearchOptions) ([]PageInfo, error) {
	// Normalize the query
	lQuery := strings.TrimSpace(strings.ToLower(query))

	if lQuery == "" || lQuery == "empty" {
		return []PageInfo{}, nil
	}
	if lQuery == "test query" {
		mockTime := time.Now()
		return []PageInfo{
			{
				URL:       "https://example.com/page1",
				Title:     "Example Page 1",
				Content:   "Content for page 1.",
				Summary:   "Summary 1",
				Published: mockTime,
			},
			{
				URL:       "https://example.com/page2",
				Title:     "Example Page 2",
				Content:   "Content for page 2.",
				Summary:   "Summary 2",
				Published: mockTime.Add(-24 * time.Hour),
			},
			{
				URL:       "https://example.com/page3",
				Title:     "Example Page 3",
				Content:   "Content for page 3.",
				Summary:   "Summary 3",
				Published: mockTime.Add(-48 * time.Hour),
			},
		}, nil
	}
	if lQuery == "government data" && len(options.SearchDomainFilter) > 0 {
		for _, filter := range options.SearchDomainFilter {
			if strings.Contains(strings.ToLower(filter), ".gov") {
				mockTime := time.Now()
				return []PageInfo{
					{
						URL:       "https://example.gov/page",
						Title:     "Government Example",
						Content:   "Content from a government site.",
						Summary:   "Gov summary",
						Published: mockTime,
					},
				}, nil
			}
		}
		return []PageInfo{}, nil
	}
	// Default: return 10 dummy results
	return generateDummyResults(10), nil
}
