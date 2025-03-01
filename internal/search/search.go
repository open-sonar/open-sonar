package search

import (
	"open-sonar/internal/search/webscrape"
)

// SearchOptions specifies search parameters
type SearchOptions struct {
	MaxPages       int
	MaxRetries     int
	DomainFilters  []string
	RecencyFilter  string
	Provider       string
	IncludeContent bool
}

// Result represents a search result
type Result struct {
	URL       string
	Title     string
	Content   string
	Summary   string
	Published interface{}
}

// Search performs a web search for the given query
func Search(query string, options SearchOptions) ([]Result, error) {
	webscrapeOptions := webscrape.SearchOptions{
		MaxPages:            options.MaxPages,
		MaxRetries:          options.MaxRetries,
		SearchDomainFilter:  options.DomainFilters,
		SearchRecencyFilter: options.RecencyFilter,
	}

	results := webscrape.ScrapeWithOptions(query, webscrapeOptions)

	searchResults := make([]Result, 0, len(results))
	for _, r := range results {
		searchResults = append(searchResults, Result{
			URL:       r.URL,
			Title:     r.Title,
			Content:   r.Content,
			Summary:   r.Summary,
			Published: r.Published,
		})
	}

	return searchResults, nil
}
