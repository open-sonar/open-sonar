package webscrape

import (
	"testing"
)

func TestScrape(t *testing.T) {
	query := "Bulbasaur"
	maxPages := 1
	maxRetries := 1

	results := Scrape(query, maxPages, maxRetries)
	if len(results) == 0 {
		t.Fatalf("Expected at least 1 result for query %q, got 0", query)
	}

	if len(results) > 1 {
		t.Errorf("Expected at most 1 result, got %d", len(results))
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
	query := "asdlkfjqwertyuiopzxcvbn" 
	maxPages := 1
	maxRetries := 1

	results := Scrape(query, maxPages, maxRetries)
	if len(results) > 2 {
		t.Errorf("Expected 0–2 results, got %d", len(results))
	}
}

func TestDuckDuckGoSearchProvider(t *testing.T) {
	provider := &DuckDuckGoSearchProvider{}

	query := "arrested development"
	maxPages := 1
	maxRetries := 1

	results, err := provider.Search(query, maxPages, maxRetries)
	if err != nil {
		t.Fatalf("Expected no error for query %q, got %v", query, err)
	}
	if len(results) == 0 {
		t.Logf("No results found for query %q, as expected", query)
	} else {
		first := results[0]
		if first.URL == "" {
			t.Error("Expected a non-empty URL in the first result")
		}
		if len(first.Title) == 0 && len(first.Content) == 0 {
			t.Error("Expected title or content")
		}
	}
}

