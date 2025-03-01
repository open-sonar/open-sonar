package webscrape

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"open-sonar/internal/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
)

// randomUserAgent returns a random user agent to avoid bot detection
func randomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/115.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.1 Safari/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36 Edg/116.0.1938.62",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/117.0",
	}
	return userAgents[rand.Intn(len(userAgents))]
}

// DuckDuckGoSearchProvider implements the SearchProvider interface for DuckDuckGo
type DuckDuckGoSearchProvider struct{}

// Search performs a web search using DuckDuckGo
func (p *DuckDuckGoSearchProvider) Search(query string, options SearchOptions) ([]PageInfo, error) {
	var results []PageInfo

	// Ensure reasonable defaults
	if options.MaxPages <= 0 {
		options.MaxPages = 1
	}
	if options.MaxRetries <= 0 {
		options.MaxRetries = 2
	}

	// Create a search timer
	searchTimer := utils.NewTimer("DuckDuckGo search")
	defer searchTimer.Stop()

	baseURL := "https://html.duckduckgo.com/html/"
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("%s?q=%s", baseURL, encodedQuery)

	resultsMap := make(map[string]bool) // Track URLs to avoid duplicates

	for page := 0; page < options.MaxPages; page++ {
		pageResults, nextURL, err := p.scrapePage(searchURL, options.MaxRetries)
		if err != nil {
			utils.Warn(fmt.Sprintf("Error scraping page %d: %v", page+1, err))
			break
		}

		// Add non-duplicate results
		for _, result := range pageResults {
			if !resultsMap[result.URL] {
				results = append(results, result)
				resultsMap[result.URL] = true

				// Fetch full content for each result
				go p.enrichResultContent(&result)
			}
		}

		if nextURL == "" {
			break
		}

		searchURL = nextURL
		time.Sleep(200 * time.Millisecond) // Polite delay between pages
	}

	return results, nil
}

// scrapePage scrapes a single page of DuckDuckGo search results
func (p *DuckDuckGoSearchProvider) scrapePage(url string, maxRetries int) ([]PageInfo, string, error) {
	var results []PageInfo

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	var resp *http.Response
	var err error

	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			continue
		}

		req.Header.Set("User-Agent", randomUserAgent())
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Accept-Language", "en-US,en;q=0.5")

		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == 200 {
			break
		}

		// Progressive backoff
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	if err != nil || resp == nil || resp.StatusCode != 200 {
		return nil, "", errors.New("failed to fetch search results")
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, "", err
	}

	// Extract results with detailed selectors
	doc.Find(".result, .web-result").Each(func(i int, s *goquery.Selection) {
		// Extract title
		titleSel := s.Find(".result__title, .result__a")
		title := titleSel.Text()

		// Extract snippet/content
		snippet := s.Find(".result__snippet").Text()

		// Extract URL
		href, exists := s.Find("a.result__url").Attr("href")
		if !exists {
			href, exists = s.Find("a.result__a").Attr("href")
			if !exists {
				href, _ = titleSel.Attr("href")
			}
		}

		// Process URL if found
		if href != "" {
			href = strings.TrimSpace(href)
			href = p.cleanUrl(href)

			// Extract additional metadata where available
			dateStr := s.Find(".result__timestamp").Text()
			var pubDate time.Time
			if dateStr != "" {
				pubDate, _ = time.Parse("2006-01-02", dateStr)
			} else {
				pubDate = time.Now() // Default to current time if not available
			}

			results = append(results, PageInfo{
				URL:       href,
				Title:     strings.TrimSpace(title),
				Content:   strings.TrimSpace(snippet),
				Summary:   strings.TrimSpace(snippet),
				Published: pubDate,
			})
		}
	})

	// Check if there's a next page
	nextURL := ""
	doc.Find(".nav-link").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(strings.ToLower(s.Text()), "next") {
			nextURL, _ = s.Attr("href")
			if !strings.HasPrefix(nextURL, "http") {
				nextURL = "https://html.duckduckgo.com" + nextURL
			}
		}
	})

	return results, nextURL, nil
}

// cleanUrl processes and standardizes URLs from DuckDuckGo
func (p *DuckDuckGoSearchProvider) cleanUrl(href string) string {
	// Handle DuckDuckGo's redirection URLs
	if strings.Contains(href, "duckduckgo.com/l/?uddg=") {
		// Extract the actual URL from DuckDuckGo's redirect
		parsed, err := url.Parse(href)
		if err == nil {
			if uddg := parsed.Query().Get("uddg"); uddg != "" {
				if decoded, err := url.QueryUnescape(uddg); err == nil {
					href = decoded
				}
			}
		}
	}

	// Ensure URL starts with a protocol
	if !strings.HasPrefix(href, "http") {
		href = "https://" + strings.TrimPrefix(href, "//")
	}

	return href
}

// enrichResultContent fetches the full content of a page and updates the PageInfo
func (p *DuckDuckGoSearchProvider) enrichResultContent(result *PageInfo) {
	// Skip certain file types
	if strings.HasSuffix(result.URL, ".pdf") || strings.HasSuffix(result.URL, ".doc") ||
		strings.HasSuffix(result.URL, ".docx") || strings.HasSuffix(result.URL, ".xlsx") {
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", result.URL, nil)
	if err != nil {
		return
	}

	req.Header.Set("User-Agent", randomUserAgent())
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

	// Check content type to make sure it's HTML
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return
	}

	// Use readability to extract the article content
	baseURL, _ := url.Parse(result.URL)

	// Read up to 1MB to avoid large files
	limitedReader := io.LimitReader(resp.Body, 1024*1024)
	article, err := readability.FromReader(limitedReader, baseURL)
	if err != nil {
		return
	}

	// Extract better metadata
	if article.Title != "" {
		result.Title = article.Title
	}

	// Clean text content
	content := p.cleanText(article.TextContent)
	result.Content = content

	// Generate a better summary if available
	if len(content) > 0 {
		result.Summary = p.generateSummary(content)
	}

	// Try to extract publication date from HTML meta tags
	// Note: The readability package doesn't directly provide publication date
	// We'll use the response headers as a fallback
	if result.Published.IsZero() || result.Published.Equal(time.Now()) {
		// Try to get date from Last-Modified header as a fallback
		if lastMod := resp.Header.Get("Last-Modified"); lastMod != "" {
			if pubTime, err := time.Parse(time.RFC1123, lastMod); err == nil {
				result.Published = pubTime
			}
		} else {
			// Use article extraction time as an approximation
			result.Published = time.Now()
		}
	}
}

// generateSummary creates a summary from the content
func (p *DuckDuckGoSearchProvider) generateSummary(content string) string {
	// Simple summarization: first 2-3 sentences
	sentences := splitToSentences(content)

	if len(sentences) == 0 {
		return ""
	}

	// Take first 2-3 sentences depending on length
	var summary strings.Builder
	totalLength := 0
	maxLength := 300

	for i, sentence := range sentences {
		if i >= 3 || totalLength+len(sentence) > maxLength {
			break
		}

		if summary.Len() > 0 {
			summary.WriteString(" ")
		}
		summary.WriteString(sentence)
		totalLength += len(sentence)
	}

	return summary.String()
}

// splitToSentences splits text into sentences
func splitToSentences(text string) []string {
	// Simple sentence splitter - can be improved
	sentenceEnders := regexp.MustCompile(`[.!?]`)
	parts := sentenceEnders.Split(text, -1)

	var sentences []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) > 10 { // Ignore very short segments
			sentences = append(sentences, part+".")
		}
	}

	return sentences
}

// cleanText removes whitespace and irrelevant patterns from text
func (p *DuckDuckGoSearchProvider) cleanText(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// Replace multiple spaces with single space
	multipleSpaces := regexp.MustCompile(`\s+`)
	text = multipleSpaces.ReplaceAllString(text, " ")

	// List of unwanted patterns
	unwantedPatterns := []string{
		`(?i)related articles:.*`,
		`(?i)see also:.*`,
		`(?i)references\s*\d*.*`,
		`(?i)external links:.*`,
		`(?i)share this article.*`,
		`(?i)you might also like.*`,
		`(?i)advertisement.*`,
		`(?i)subscribe for more.*`,
		`(?i)trending now:.*`,
		`(?i)follow us on.*`,
		`(?i)comments.*`,
		`(?i)leave a reply.*`,
		`(?i)watch now:.*`,
		`(?i)click here.*`,
		`(?i)continue reading.*`,
		`(?i)skip to main content.*`,
		`(?i)privacy policy.*`,
		`(?i)terms of service.*`,
	}

	for _, pattern := range unwantedPatterns {
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllString(text, "")
	}

	return strings.TrimSpace(text)
}
