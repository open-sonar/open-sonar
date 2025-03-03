package webscrape

import (
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	// Set test mode for all tests in this package
	os.Setenv("TEST_MODE", "true")

	// Run tests
	exitCode := m.Run()

	// Exit with the same code
	os.Exit(exitCode)
}

func TestScrape(t *testing.T) {
	oldProvider := SetGetSearchProvider(func(provider string) (SearchProvider, error) {
		return &MockSearchProvider{}, nil
	})

	defer SetGetSearchProvider(oldProvider)

	query := "Bulbasaur"
	maxPages := 1
	maxRetries := 1

	results := Scrape(query, maxPages, maxRetries)
	if len(results) == 0 {
		t.Fatalf("Expected at least 1 result for query %q, got 0", query)
	}

	if len(results) != 10 {
		t.Errorf("Expected 3 results from mock provider, got %d", len(results))
	}
	

	first := results[0]
	if first.URL == "" {
		t.Error("Expected a non-empty URL in the first result")
	}
	if len(first.Title) == 0 && len(first.Content) == 0 {
		t.Error("Expected title or content")
	}
}

func TestScrapeNoResults(t *testing.T) {
	oldProvider := SetGetSearchProvider(func(provider string) (SearchProvider, error) {
		return &MockSearchProvider{}, nil
	})

	defer SetGetSearchProvider(oldProvider)

	query := ""
	maxPages := 1
	maxRetries := 1

	results := Scrape(query, maxPages, maxRetries)
	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestScrapeWithOptions(t *testing.T) {
	options := SearchOptions{
		MaxPages: 7,
		MaxRetries: 1,
		SearchDomainFilter: []string{".gov"},
	}
	

	results := ScrapeWithOptions("government data", options)

	expectedResults := 7
	if len(results) != expectedResults {
		t.Errorf("Expected %d result after domain filtering, got %d", expectedResults, len(results))
	}
}

func TestDomainFilters(t *testing.T) {
	results := []PageInfo{
		{URL: "https://example.com/page1"},
		{URL: "https://wikipedia.org/page2"},
		{URL: "https://example.org/page3"},
	}

	tests := []struct {
		name         string
		filters      []string
		wantURLs     []string
		dontWantURLs []string
	}{
		{
			name:         "Allow single domain",
			filters:      []string{"wikipedia.org"},
			wantURLs:     []string{"wikipedia.org"},
			dontWantURLs: []string{"example.com", "example.org"},
		},
		{
			name:         "Block single domain",
			filters:      []string{"-example.com"},
			wantURLs:     []string{"wikipedia.org", "example.org"},
			dontWantURLs: []string{"example.com"},
		},
		{
			name:         "Mixed allow and block",
			filters:      []string{"wikipedia.org", "-example.org"},
			wantURLs:     []string{"wikipedia.org"},
			dontWantURLs: []string{"example.com", "example.org"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := SearchOptions{
				SearchDomainFilter: tt.filters,
			}

			filtered := FilterResults(results, options)

			for _, wantURL := range tt.wantURLs {
				found := false
				for _, result := range filtered {
					if strings.Contains(result.URL, wantURL) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find URL containing %q but didn't", wantURL)
				}
			}

			// Check that unwanted URLs are absent
			for _, dontWantURL := range tt.dontWantURLs {
				found := false
				for _, result := range filtered {
					if strings.Contains(result.URL, dontWantURL) {
						found = true
						break
					}
				}
				if found {
					t.Errorf("Expected NOT to find URL containing %q but did", dontWantURL)
				}
			}
		})
	}
}
