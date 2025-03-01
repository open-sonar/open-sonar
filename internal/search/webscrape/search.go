package webscrape

import (
	"fmt"
	"open-sonar/internal/utils"
	"strings"
)

// ScrapeWithOptions performs web scraping with additional options
func ScrapeWithOptions(query string, options SearchOptions) []PageInfo {
	// Create a search timer
	searchTimer := utils.NewTimer("DuckDuckGo search")
	defer searchTimer.Stop()

	// Create a new provider
	provider := &DuckDuckGoSearchProvider{}

	// Perform the search
	results, err := provider.Search(query, options)
	if err != nil {
		utils.Error(fmt.Sprintf("Search error: %v", err))
		return []PageInfo{}
	}

	// Add debug logging to verify we have URLs
	if len(results) > 0 {
		utils.Info(fmt.Sprintf("Search returned %d results", len(results)))
		for i, result := range results {
			if i < 3 { // Just log the first 3 for brevity
				utils.Debug(fmt.Sprintf("Result %d: URL=%s, Title=%s", i+1, result.URL, result.Title))
			}
		}
	} else {
		utils.Warn("Search returned no results")
	}

	// Filter by domain if needed
	if len(options.SearchDomainFilter) > 0 {
		results = filterByDomain(results, options.SearchDomainFilter)
		utils.Info(fmt.Sprintf("After domain filtering: %d results remain", len(results)))
	}

	return results
}

// Scrape performs web scraping with default options
func Scrape(query string, maxPages int, maxRetries int) []PageInfo {
	return ScrapeWithOptions(query, SearchOptions{
		MaxPages:   maxPages,
		MaxRetries: maxRetries,
	})
}

// filterByDomain filters search results by domain patterns
func filterByDomain(results []PageInfo, domainFilters []string) []PageInfo {
	if len(domainFilters) == 0 {
		return results
	}

	var filteredResults []PageInfo
	for _, result := range results {
		// Check if any domain filter matches the URL
		for _, filter := range domainFilters {
			// Check both includes and excludes
			if strings.HasPrefix(filter, "!") {
				// This is an exclude filter
				excludePattern := filter[1:]
				if !strings.Contains(result.URL, excludePattern) {
					filteredResults = append(filteredResults, result)
					break
				}
			} else {
				// This is an include filter
				if strings.Contains(result.URL, filter) {
					filteredResults = append(filteredResults, result)
					break
				}
			}
		}
	}
	return filteredResults
}
