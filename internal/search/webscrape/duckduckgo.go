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

type DuckDuckGoSearchProvider struct{}

func (p *DuckDuckGoSearchProvider) Search(query string, options SearchOptions) ([]PageInfo, error) {
	var results []PageInfo

	if options.MaxPages <= 0 {
		options.MaxPages = 1
	}
	if options.MaxRetries <= 0 {
		options.MaxRetries = 2
	}

	searchTimer := utils.NewTimer("DuckDuckGo search")
	defer searchTimer.Stop()

	baseURL := "https://html.duckduckgo.com/html/"
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("%s?q=%s", baseURL, encodedQuery)

	resultsMap := make(map[string]bool)

	for page := 0; page < options.MaxPages; page++ {
		pageResults, nextURL, err := p.scrapePage(searchURL, options.MaxRetries)
		if err != nil {
			utils.Warn(fmt.Sprintf("Error scraping page %d: %v", page+1, err))
			break
		}

		for _, result := range pageResults {
			if !resultsMap[result.URL] {
				results = append(results, result)
				resultsMap[result.URL] = true
				go p.enrichResultContent(&result)
			}
		}

		if nextURL == "" {
			break
		}

		searchURL = nextURL
		time.Sleep(200 * time.Millisecond)
	}

	return results, nil
}

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
	doc.Find(".result, .web-result").Each(func(i int, s *goquery.Selection) {
		titleSel := s.Find(".result__title, .result__a")
		title := titleSel.Text()
		snippet := s.Find(".result__snippet").Text()
		href, exists := s.Find("a.result__url").Attr("href")
		if !exists {
			href, exists = s.Find("a.result__a").Attr("href")
			if !exists {
				href, _ = titleSel.Attr("href")
			}
		}
		if href != "" {
			href = strings.TrimSpace(href)
			href = p.cleanUrl(href)
			dateStr := s.Find(".result__timestamp").Text()
			var pubDate time.Time
			if dateStr != "" {
				pubDate, _ = time.Parse("2006-01-02", dateStr)
			} else {
				pubDate = time.Now()
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

func (p *DuckDuckGoSearchProvider) cleanUrl(href string) string {
	if strings.Contains(href, "duckduckgo.com/l/?uddg=") {
		parsed, err := url.Parse(href)
		if err == nil {
			if uddg := parsed.Query().Get("uddg"); uddg != "" {
				if decoded, err := url.QueryUnescape(uddg); err == nil {
					href = decoded
				}
			}
		}
	}
	if !strings.HasPrefix(href, "http") {
		href = "https://" + strings.TrimPrefix(href, "//")
	}
	return href
}

func (p *DuckDuckGoSearchProvider) enrichResultContent(result *PageInfo) {
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
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return
	}
	baseURL, _ := url.Parse(result.URL)
	limitedReader := io.LimitReader(resp.Body, 1024*1024)
	article, err := readability.FromReader(limitedReader, baseURL)
	if err != nil {
		return
	}
	if article.Title != "" {
		result.Title = article.Title
	}
	content := p.cleanText(article.TextContent)
	result.Content = content
	if len(content) > 0 {
		result.Summary = p.generateSummary(content)
	}
	if result.Published.IsZero() || result.Published.Equal(time.Now()) {
		if lastMod := resp.Header.Get("Last-Modified"); lastMod != "" {
			if pubTime, err := time.Parse(time.RFC1123, lastMod); err == nil {
				result.Published = pubTime
			}
		} else {
			result.Published = time.Now()
		}
	}
}

func (p *DuckDuckGoSearchProvider) generateSummary(content string) string {
	sentences := splitToSentences(content)
	if len(sentences) == 0 {
		return ""
	}
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

func splitToSentences(text string) []string {
	sentenceEnders := regexp.MustCompile(`[.!?]`)
	parts := sentenceEnders.Split(text, -1)
	var sentences []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) > 10 {
			sentences = append(sentences, part+".")
		}
	}
	return sentences
}

func (p *DuckDuckGoSearchProvider) cleanText(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	multipleSpaces := regexp.MustCompile(`\s+`)
	text = multipleSpaces.ReplaceAllString(text, " ")
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
