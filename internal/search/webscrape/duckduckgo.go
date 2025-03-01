package webscrape

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

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

	baseURL := "https://html.duckduckgo.com/html/"
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("%s?q=%s", baseURL, encodedQuery)

	for page := 0; page < options.MaxPages; page++ {
		pageResults, nextURL, err := p.scrapePage(searchURL, options.MaxRetries)
		if err != nil {
			break
		}

		results = append(results, pageResults...)

		if nextURL == "" {
			break
		}

		searchURL = nextURL
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

		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.110 Safari/537.36")

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

	doc.Find(".result").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".result__title").Text()
		snippet := s.Find(".result__snippet").Text()

		href, exists := s.Find(".result__url").Attr("href")
		if !exists {
			href, _ = s.Find(".result__a").Attr("href")
		}

		if href != "" {
			href = strings.TrimSpace(href)
			if !strings.HasPrefix(href, "http") {
				href = "https://" + strings.TrimPrefix(href, "//")
			}

			results = append(results, PageInfo{
				URL:       href,
				Title:     strings.TrimSpace(title),
				Content:   strings.TrimSpace(snippet),
				Summary:   strings.TrimSpace(snippet),
				Published: time.Now(), // We don't have actual publish dates from DDG
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
