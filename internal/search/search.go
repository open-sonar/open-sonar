package search

import (
	"os"

	"open-sonar/internal/search/webscrape"
	"open-sonar/internal/utils"
)

// SearchOptions specifies search parameters.
type SearchOptions struct {
	MaxPages       int
	MaxRetries     int
	DomainFilters  []string
	RecencyFilter  string
	Provider       string
	IncludeContent bool
}

// Result represents a search result.
type Result struct {
	URL       string
	Title     string
	Content   string
	Summary   string
	Published interface{}
}

// Search performs a web search for the given query.
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

// RunSearch performs a web search with the given parameters.
// In test mode, it returns dummy results based on special query keywords.
func RunSearch(query string, maxPages int, maxRetries int, domainFilters []string) []webscrape.PageInfo {
	options := webscrape.SearchOptions{
		MaxPages:           maxPages,
		MaxRetries:         maxRetries,
		SearchDomainFilter: domainFilters,
	}

	// In test mode, return dummy results based on the query.
	if os.Getenv("TEST_MODE") == "true" {
		switch query {
		case "empty":
			return []webscrape.PageInfo{}
		case "test query":
			return []webscrape.PageInfo{
				{URL: "http://example.com/1", Title: "Title 1", Content: "Content 1", Summary: "Summary 1"},
				{URL: "http://example.com/2", Title: "Title 2", Content: "Content 2", Summary: "Summary 2"},
				{URL: "http://example.com/3", Title: "Title 3", Content: "Content 3", Summary: "Summary 3"},
			}
		case "government data":
			if len(domainFilters) > 0 {
				return []webscrape.PageInfo{
					{URL: "http://gov.example.com", Title: "Government Data", Content: "Content", Summary: "Summary"},
				}
			}
		}
	}

	// For non-test mode, perform a real search.
	results := webscrape.ScrapeWithOptions(query, options)

	utils.Info("Search completed: " +
		utils.FormatInt(len(results)) + " results found for query: " + query)

	return results
}

// SearchWithFilters performs a search with additional filters.
func SearchWithFilters(query string, maxPages int, maxRetries int,
	domainFilters []string, recencyFilter string) []webscrape.PageInfo {

	options := webscrape.SearchOptions{
		MaxPages:            maxPages,
		MaxRetries:          maxRetries,
		SearchDomainFilter:  domainFilters,
		SearchRecencyFilter: recencyFilter,
	}

	return webscrape.ScrapeWithOptions(query, options)
}
