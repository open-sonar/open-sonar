//preliminary web scraper

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-shiori/go-readability"
)

type PageInfo struct {
	URL     string `json:"url"`
	Title   string `json:"title,omitempty"`
	Summary string `json:"summary,omitempty"`
	Content string `json:"content,omitempty"`
	Error   string `json:"error,omitempty"`
}

func searchDuckDuckGo(query string, maxPages int) ([]string, error) {
	// build search URL
	searchURL := "https://duckduckgo.com/html/?q=" + strings.ReplaceAll(query, " ", "+")
	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoScraper/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var links []string
	doc.Find("a.result__a").Each(func(i int, s *goquery.Selection) {
		if i >= maxPages {
			return
		}
		if href, exists := s.Attr("href"); exists {
			// if link protocol relative
			if strings.HasPrefix(href, "//") {
				href = "https:" + href
			}
			// uddg check
			parsed, err := url.Parse(href)
			if err == nil {
				qparams := parsed.Query()
				if uddg := qparams.Get("uddg"); uddg != "" {
					if decoded, err := url.QueryUnescape(uddg); err == nil {
						href = decoded
					}
				}
			}
			links = append(links, href)
		}
	})
	return links, nil
}

// get rid of whitespace/newlines and remove irrelevant patterns
func cleanText(text string) string {
	// get rid of whitespace and newlines
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
		`(?i)continue reading.*`,           // remove pagination
		`(?i)skip to main content.*`,       // remove accessibility prompts
		`(?i)privacy policy.*`,             // remove privacy policy
		`(?i)terms of service.*`,           // remove terms and conditions
	}

	for _, pattern := range unwantedPatterns {
		re := regexp.MustCompile(pattern)
		text = re.ReplaceAllString(text, "")
	}

	return text
}

func crawlPage(urlStr string) (string, string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; GoScraper/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	baseURL, err := url.Parse(urlStr)
	if err != nil {
		return "", "", err
	}

	article, err := readability.FromReader(resp.Body, baseURL)
	if err != nil {
		return "", "", err
	}

	title := strings.TrimSpace(article.Title)
	content := cleanText(strings.TrimSpace(article.TextContent))
	return title, content, nil
}

// return first two sentences of the content for summary
func summarize(content string) string {
	sentences := strings.Split(content, ".")
	if len(sentences) < 2 {
		return content
	}
	return strings.TrimSpace(sentences[0]) + ". " + strings.TrimSpace(sentences[1]) + "."
}

func main() {
	query := flag.String("query", "", "Search query for web scraping")

	pages := flag.Int("pages", 3, "# pages to scrape (default 3)")

	flag.Parse()
	if *query == "" {
		log.Fatal("Provide search query using -query flag.")
	}

	fmt.Printf("Searching for: %s\n", *query)

	links, err := searchDuckDuckGo(*query, *pages)
	if err != nil {
		log.Fatalf("Search error: %v", err)
	}
	if len(links) == 0 {
		log.Fatal("No search results found.")
	}

	// link crawl
	results := make([]PageInfo, len(links))
	var wg sync.WaitGroup
	for i, link := range links {
		wg.Add(1)
		go func(i int, link string) {
			defer wg.Done()
			title, content, err := crawlPage(link)
			if err != nil {
				results[i] = PageInfo{URL: link, Error: err.Error()}
				return
			}
			summary := summarize(content)
			results[i] = PageInfo{
				URL:     link,
				Title:   title,
				Content: content,
				Summary: summary,
			}
		}(i, link)
	}
	wg.Wait()

	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		log.Fatalf("Error compiling JSON: %v", err)
	}

	fmt.Println(string(jsonData))
}
