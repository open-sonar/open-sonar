package citations

import (
	"open-sonar/internal/search/webscrape"
)

// ExtractCitationURLs extracts citation URLs from search results
func ExtractCitationURLs(results []webscrape.PageInfo) []string {
	// Return nil if results is empty
	if len(results) == 0 {
		return nil
	}

	urls := make([]string, 0, len(results))

	// Extract URLs from page info
	for _, result := range results {
		if result.URL != "" {
			urls = append(urls, result.URL)
		}
	}

	// Return nil if no URLs were extracted
	if len(urls) == 0 {
		return nil
	}

	return urls
}

// ExtractFullCitations extracts full citation information including title and summary
func ExtractFullCitations(results []webscrape.PageInfo) []map[string]string {
	if len(results) == 0 {
		return nil
	}

	citations := make([]map[string]string, 0, len(results))

	for _, result := range results {
		if result.URL != "" {
			citation := map[string]string{
				"URL":     result.URL,
				"Title":   result.Title,
				"Summary": result.Summary,
			}
			citations = append(citations, citation)
		}
	}

	if len(citations) == 0 {
		return nil
	}

	return citations
}
