package citations

import (
	"fmt"
	"open-sonar/internal/search/webscrape"
	"strings"
	"time"
)

type Citation struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Summary string `json:"summary,omitempty"`
}

type CitationResponse struct {
	Created   int64      `json:"created"`
	Citations []Citation `json:"citations"`
}

// JSON schema for CitationResponse
const CitationSchema = `{
  "type": "object",
  "properties": {
    "created": { "type": "number" },
    "citations": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "url": { "type": "string" },
          "title": { "type": "string" },
          "summary": { "type": "string" }
        },
        "required": ["url"]
      }
    }
  },
  "required": ["created", "citations"]
}`

// ExtractCitations converts PageInfo objects to citations
func ExtractCitations(pages []webscrape.PageInfo, style string) CitationResponse {
	var citations []Citation
	for _, page := range pages {
		if style == "full" {
			citations = append(citations, Citation{
				URL:     page.URL,
				Title:   page.Title,
				Summary: page.Summary,
			})
		} else {
			citations = append(citations, Citation{
				URL: page.URL,
			})
		}
	}
	return CitationResponse{
		Created:   time.Now().Unix(),
		Citations: citations,
	}
}

// NOTE: ExtractCitationURLs is now in citations.go

// CitationExtractor handles extraction and formatting of citations
type CitationExtractor struct {
	MaxResults int
}

// NewCitationExtractor creates a new citation extractor
func NewCitationExtractor() *CitationExtractor {
	return &CitationExtractor{
		MaxResults: 10,
	}
}

// Extract extracts full and short citations from search results
func (e *CitationExtractor) Extract(query string, maxPages int, maxRetries int) ([]map[string]string, []string) {
	fmt.Printf("Running citation extraction test for query: %q with maxPages=%d, maxRetries=%d\n",
		query, maxPages, maxRetries)

	// Use mock results for testing
	results := webscrape.CreateMockResults()

	// Extract full citations (with title and summary)
	fullCitations := e.extractFullCitations(results)

	fmt.Printf("Extracted %d full citations.\n", len(fullCitations))
	for i, citation := range fullCitations {
		if i < 3 { // Show just the first 3 for brevity
			fmt.Printf("Full Citation [%d]: URL=%s, Title=%q, Summary=%q\n",
				i+1, citation["URL"], citation["Title"], citation["Summary"])
		}
	}

	// Extract short citations (URLs only) - use the function from citations.go
	shortCitations := ExtractCitationURLs(results)

	fmt.Printf("Extracted %d short citations.\n", len(shortCitations))
	for i, url := range shortCitations {
		if i < 3 { // Show just the first 3 for brevity
			fmt.Printf("Short Citation [%d]: URL=%s\n", i+1, url)
		}
	}

	return fullCitations, shortCitations
}

// extractFullCitations extracts detailed citation information
func (e *CitationExtractor) extractFullCitations(results []webscrape.PageInfo) []map[string]string {
	// Limit number of results if needed
	if len(results) > e.MaxResults {
		results = results[:e.MaxResults]
	}

	citations := make([]map[string]string, 0, len(results))
	for _, result := range results {
		// Clean up citation text
		title := strings.TrimSpace(result.Title)
		content := strings.TrimSpace(result.Content)
		summary := strings.TrimSpace(result.Summary)

		// If summary is empty, generate one from content
		if summary == "" && content != "" {
			if len(content) > 200 {
				summary = content[:200] + "..."
			} else {
				summary = content
			}
		}

		citation := map[string]string{
			"URL":     result.URL,
			"Title":   title,
			"Summary": summary,
		}

		citations = append(citations, citation)
	}

	return citations
}
