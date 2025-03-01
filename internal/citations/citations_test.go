package citations

import (
	"fmt"
	"testing"
	"time"

	"open-sonar/internal/search/webscrape"
)

// TestCitationExtraction tests the citation extraction functionality
func TestCitationExtraction(t *testing.T) {
	// Create sample data
	query := "I'm hungry"
	maxPages := 2
	maxRetries := 2

	fmt.Printf("Running citation extraction test for query: %q with maxPages=%d, maxRetries=%d\n",
		query, maxPages, maxRetries)

	// Create mock pages
	mockPages := []webscrape.PageInfo{}
	for i := 1; i <= 10; i++ {
		mockPages = append(mockPages, webscrape.PageInfo{
			URL:       fmt.Sprintf("https://example.com/page%d", i),
			Title:     fmt.Sprintf("Example Page %d", i),
			Content:   fmt.Sprintf("This is the content of page %d", i),
			Summary:   fmt.Sprintf("Summary of page %d", i),
			Published: time.Now().Add(-time.Duration(i) * time.Hour),
		})
	}

	// Test ExtractCitations - full style
	fullCitations := ExtractCitations(mockPages, "full")
	fmt.Printf("Extracted %d full citations.\n", len(fullCitations.Citations))

	if len(fullCitations.Citations) != len(mockPages) {
		t.Errorf("Expected %d full citations, got %d", len(mockPages), len(fullCitations.Citations))
	}

	// Print some sample full citations
	for i := 0; i < 3 && i < len(fullCitations.Citations); i++ {
		c := fullCitations.Citations[i]
		fmt.Printf("Full Citation [%d]: URL=%s, Title=%q, Summary=%q\n",
			i+1, c.URL, c.Title, c.Summary)
	}

	// Test ExtractCitations - short style
	shortCitations := ExtractCitations(mockPages, "short")
	fmt.Printf("Extracted %d short citations.\n", len(shortCitations.Citations))

	if len(shortCitations.Citations) != len(mockPages) {
		t.Errorf("Expected %d short citations, got %d", len(mockPages), len(shortCitations.Citations))
	}

	// Print some sample short citations
	for i := 0; i < 3 && i < len(shortCitations.Citations); i++ {
		c := shortCitations.Citations[i]
		fmt.Printf("Short Citation [%d]: URL=%s\n", i+1, c.URL)
	}

	// Test ExtractCitationURLs
	urls := ExtractCitationURLs(mockPages)

	if len(urls) != len(mockPages) {
		t.Errorf("Expected %d URLs, got %d", len(mockPages), len(urls))
	}
}
