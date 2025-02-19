package citations

import (
	"fmt"
	"testing"

	"open-sonar/internal/search/webscrape"
)


func TestScraperAndCitations(t *testing.T) {
	query := "I'm hungry"
	maxPages := 2
	maxRetries := 2

	fmt.Printf("Running scraper test for query: %q with maxPages=%d, maxRetries=%d\n", query, maxPages, maxRetries)

	pages := webscrape.Scrape(query, maxPages, maxRetries)
	if len(pages) == 0 {
		t.Fatal("Scraper returned zero pages.")
	}

	fmt.Printf("Scraper retrieved %d pages.\n", len(pages))

	fullCitations := ExtractCitations(pages, "full")
	if len(fullCitations.Citations) == 0 {
		t.Fatal("Expected at least one citation in 'full' mode.")
	}

	fmt.Printf("Extracted %d full citations.\n", len(fullCitations.Citations))

	for i, c := range fullCitations.Citations {
		fmt.Printf("Full Citation [%d]: URL=%s, Title=%q, Summary=%q\n", i+1, c.URL, c.Title, c.Summary)

		if c.URL == "" {
			t.Errorf("Citation [%d] has an empty URL", i+1)
		}
		if c.Title == "" {
			t.Errorf("Citation [%d] has an empty title", i+1)
		}
		if c.Summary == "" {
			t.Errorf("Citation [%d] has an empty summary", i+1)
		}
	}

	shortCitations := ExtractCitations(pages, "short")
	if len(shortCitations.Citations) == 0 {
		t.Fatal("Expected at least one citation in 'short' mode.")
	}

	fmt.Printf("Extracted %d short citations.\n", len(shortCitations.Citations))

	for i, c := range shortCitations.Citations {
		fmt.Printf("Short Citation [%d]: URL=%s\n", i+1, c.URL)

		if c.URL == "" {
			t.Errorf("Short Citation [%d] has an empty URL", i+1)
		}
	}
}
