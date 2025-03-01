package webscrape

import (
	"time"
)

// PageInfo represents information about a web page
type PageInfo struct {
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Summary   string    `json:"summary"`
	Published time.Time `json:"published"`
}

// Scrape performs a web search and returns the results
func Scrape(query string, maxPages, maxRetries int) []PageInfo {
	provider, _ := GetSearchProvider("")

	options := SearchOptions{
		MaxPages:   maxPages,
		MaxRetries: maxRetries,
	}

	results, err := provider.Search(query, options)
	if err != nil {
		return []PageInfo{}
	}

	return results
}

// ScrapeWithOptions performs a web search with additional options
func ScrapeWithOptions(query string, options SearchOptions) []PageInfo {
	provider, _ := GetSearchProvider("")

	results, err := provider.Search(query, options)
	if err != nil {
		return []PageInfo{}
	}

	return FilterResults(results, options)
}
