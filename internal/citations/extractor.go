package citations

import (
	"time"

	"open-sonar/internal/search/webscrape"
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

// ExtractCitationURLs returns just the URLs from PageInfo objects
func ExtractCitationURLs(pages []webscrape.PageInfo) []string {
	var urls []string
	for _, page := range pages {
		urls = append(urls, page.URL)
	}
	return urls
}
