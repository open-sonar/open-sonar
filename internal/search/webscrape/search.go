package webscrape

import (
	"fmt"
	"open-sonar/internal/utils"
	"os"
)

var testMode = os.Getenv("TEST_MODE") == "true"

func ScrapeWithOptions(query string, options SearchOptions) []PageInfo {
	searchTimer := utils.NewTimer("DuckDuckGo search")
	defer searchTimer.Stop()

	var provider SearchProvider
	var err error
	if testMode {
		provider, err = GetSearchProvider("mock")
	} else {
		provider, err = GetSearchProvider("duckduckgo")
	}
	if err != nil {
		utils.Error(fmt.Sprintf("Search provider error: %v", err))
		return []PageInfo{}
	}

	results, err := provider.Search(query, options)
	if err != nil {
		utils.Error(fmt.Sprintf("Search error: %v", err))
		return []PageInfo{}
	}

	if len(results) > 0 {
		utils.Info(fmt.Sprintf("Search returned %d results", len(results)))
		for i, result := range results {
			if i < 3 {
				utils.Debug(fmt.Sprintf("Result %d: URL=%s, Title=%s", i+1, result.URL, result.Title))
			}
		}
	} else {
		utils.Warn("Search returned no results")
	}

	if len(options.SearchDomainFilter) > 0 {
		results = FilterResults(results, options)
		utils.Info(fmt.Sprintf("After domain filtering: %d results remain", len(results)))
	}

	return results
}

func Scrape(query string, maxPages int, maxRetries int) []PageInfo {
	return ScrapeWithOptions(query, SearchOptions{
		MaxPages:   maxPages,
		MaxRetries: maxRetries,
	})
}
