package webscrape

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
)

type SearchProvider interface {
	Search(query string, maxPages, maxRetries int) ([]PageInfo, error)
}

type DuckDuckGoSearchProvider struct{}

func (p *DuckDuckGoSearchProvider) Search(query string, maxPages, maxRetries int) ([]PageInfo, error) {
	results := Scrape(query, maxPages, maxRetries)
	if len(results) == 0 {
		return nil, fmt.Errorf("no results found for query: %s", query)
	}
	return results, nil
}

type PageInfo struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Summary string `json:"summary,omitempty"`
	Content string `json:"content,omitempty"`
}

func Scrape(query string, maxPages, maxRetries int) []PageInfo {
	initialLinks := searchDuckDuckGo(query, maxPages*2)

	// successful results channel
	resultsChan := make(chan PageInfo, len(initialLinks))

	// goroutine for each link in initialLinks
	var wg sync.WaitGroup
	for _, link := range initialLinks {
		wg.Add(1)
		go func(link string) {
			defer wg.Done()

			title, content := crawlPageWithRetry(link, maxRetries)
			if content == "" || isPlaceholder(title, content) {
				newLinks := searchDuckDuckGo(query, 1)
				if len(newLinks) == 0 {
					return
				}
				newTitle, newContent := crawlPageWithRetry(newLinks[0], maxRetries)
				if newContent == "" || isPlaceholder(newTitle, newContent) {
					return
				}
				link = newLinks[0]
				title, content = newTitle, newContent
			}

			if content == "" || isPlaceholder(title, content) {
				return
			}

			summary := summarize(content)
			resultsChan <- PageInfo{
				URL:     link,
				Title:   title,
				Content: content,
				Summary: summary,
			}
		}(link)
	}
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	results := make([]PageInfo, 0, maxPages)
	for page := range resultsChan {
		results = append(results, page)
		if len(results) >= maxPages {
			break
		}
	}

	return results
}

func searchDuckDuckGo(query string, maxLinks int) []string {
	if maxLinks < 1 {
		return nil
	}
	searchURL := "https://duckduckgo.com/html/?q=" + strings.ReplaceAll(query, " ", "+")
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoScraper/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil
	}

	var links []string
	doc.Find("a.result__a").Each(func(i int, s *goquery.Selection) {
		if i >= maxLinks {
			return
		}
		if href, exists := s.Attr("href"); exists {
			if strings.HasPrefix(href, "//") {
				href = "https:" + href
			}
			if parsed, parseErr := url.Parse(href); parseErr == nil {
				qparams := parsed.Query()
				if uddg := qparams.Get("uddg"); uddg != "" {
					if decoded, decErr := url.QueryUnescape(uddg); decErr == nil {
						href = decoded
					}
				}
			}
			links = append(links, href)
		}
	})
	return links
}

func crawlPageWithRetry(urlStr string, maxRetries int) (string, string) {
	var title, content string
	for attempt := 1; attempt <= maxRetries; attempt++ {
		title, content = crawlPage(urlStr)
		if content != "" {
			break
		}
		time.Sleep(time.Second * time.Duration(attempt))
	}
	return title, content
}

func crawlPage(urlStr string) (string, string) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", ""
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoScraper/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return "", ""
	}
	defer resp.Body.Close()

	baseURL, parseErr := url.Parse(urlStr)
	if parseErr != nil {
		return "", ""
	}
	article, readErr := readability.FromReader(resp.Body, baseURL)
	if readErr != nil {
		return "", ""
	}

	title := strings.TrimSpace(article.Title)
	content := cleanText(strings.TrimSpace(article.TextContent))
	return title, content
}

func isPlaceholder(title, content string) bool {
	tl := strings.ToLower(title)
	switch {
	case strings.Contains(tl, "just a moment"),
		strings.Contains(tl, "one moment"),
		strings.Contains(tl, "please wait"),
		strings.Contains(tl, "checking your browser"),
		strings.Contains(tl, "redirecting"):
		return true
	}
	return len(content) < 50
}
//removes whitespace and irrelevant patterns
func cleanText(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.Join(strings.Fields(text), " ")

	unwantedPatterns := []string{
		`(?i)related articles:.*`,          // remove related articles
		`(?i)see also:.*`,                  // remove see also
		`(?i)references\s*\d*.*`,           // remove references
		`(?i)external links:.*`,            // remove external links
		`(?i)share this article.*`,         // remove share
		`(?i)you might also like.*`,        // remove suggested
		`(?i)advertisement.*`,              // remove ads
		`(?i)subscribe for more.*`,         // remove subscription prompts
		`(?i)trending now:.*`,              // remove trending
		`(?i)follow us on.*`,               // remove social media
		`(?i)comments.*`,                   // remove comment
		`(?i)leave a reply.*`,              // remove comment submission
		`(?i)watch now:.*`,                 // remove embedded videos
		`(?i)click here.*`,                 // remove clickbait phrases
		`(?i)continue reading.*`,           // remove pagination prompts
		`(?i)skip to main content.*`,       // remove accessibility prompts
		`(?i)privacy policy.*`,             // remove privacy policy links
		`(?i)terms of service.*`,           // remove terms and conditions
	}
	for _, pattern := range unwantedPatterns {
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllString(text, "")
	}

	return text
}

// summarize returns the first two sentences as a summary
func summarize(content string) string {
	sentences := strings.Split(content, ".")
	if len(sentences) < 2 {
		return content
	}
	return strings.TrimSpace(sentences[0]) + ". " + strings.TrimSpace(sentences[1]) + "."
}

func main() {
	query := flag.String("query", "", "Search query")
	pages := flag.Int("pages", 3, "# pages to retrieve (default=3)")
	retries := flag.Int("retries", 2, "# retries per page (default=2)")
	flag.Parse()

	fmt.Printf("Searching for: %s, retrieving %d pages\n", *query, *pages)

	provider := &DuckDuckGoSearchProvider{}
	results, err := provider.Search(*query, *pages, *retries)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	jsonData, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(jsonData))
}
