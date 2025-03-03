package webscrape

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Note: PageInfo and SearchOptions are defined in types.go

// SearchProvider is an interface for different search engines.
type SearchProvider interface {
	Search(query string, options SearchOptions) ([]PageInfo, error)
}

// GetSearchProviderFunc is a function type for creating search providers.
type GetSearchProviderFunc func(provider string) (SearchProvider, error)

// DefaultGetSearchProvider is the default implementation.
var DefaultGetSearchProvider GetSearchProviderFunc = func(provider string) (SearchProvider, error) {
	switch strings.ToLower(provider) {
	case "duckduckgo":
		return NewDuckDuckGoSearchProvider(), nil
	case "mock":
		return NewMockSearchProvider(), nil
	// Add other search providers as needed.
	default:
		return NewDuckDuckGoSearchProvider(), nil
	}
}

// CurrentGetSearchProvider holds the current provider implementation.
var CurrentGetSearchProvider GetSearchProviderFunc = DefaultGetSearchProvider

// GetSearchProvider returns a search provider based on provider name.
func GetSearchProvider(provider string) (SearchProvider, error) {
	return CurrentGetSearchProvider(provider)
}

// SetGetSearchProvider allows tests to replace the provider lookup function.
func SetGetSearchProvider(fn GetSearchProviderFunc) GetSearchProviderFunc {
	old := CurrentGetSearchProvider
	CurrentGetSearchProvider = fn
	return old
}

// RestoreDefaultGetSearchProvider restores the default provider lookup.
func RestoreDefaultGetSearchProvider() {
	CurrentGetSearchProvider = DefaultGetSearchProvider
}

// RecencyToTime converts a recency filter to a time.Time.
func RecencyToTime(recency string) (time.Time, error) {
	now := time.Now()
	switch strings.ToLower(recency) {
	case "hour":
		return now.Add(-1 * time.Hour), nil
	case "day":
		return now.AddDate(0, 0, -1), nil
	case "week":
		return now.AddDate(0, 0, -7), nil
	case "month":
		return now.AddDate(0, -1, 0), nil
	case "":
		return time.Time{}, nil
	default:
		return time.Time{}, fmt.Errorf("invalid recency filter: %s", recency)
	}
}

// FilterResults filters results based on domain filters and recency.
func FilterResults(results []PageInfo, options SearchOptions) []PageInfo {
	if len(options.SearchDomainFilter) == 0 && options.SearchRecencyFilter == "" {
		return results
	}

	var filtered []PageInfo
	var allowFilters []string
	var blockedFilters []string
	for _, filter := range options.SearchDomainFilter {
		if strings.HasPrefix(filter, "-") {
			blockedFilters = append(blockedFilters, strings.TrimPrefix(filter, "-"))
		} else {
			allowFilters = append(allowFilters, filter)
		}
	}

	var minTime time.Time
	if options.SearchRecencyFilter != "" {
		var err error
		minTime, err = RecencyToTime(options.SearchRecencyFilter)
		if err != nil {
			minTime = time.Time{}
		}
	}

	for _, result := range results {
		domain := extractDomain(result.URL)
		// Check blocked filters (substring match).
		blocked := false
		for _, bf := range blockedFilters {
			if strings.Contains(domain, bf) {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}

		// If allow filters are specified, at least one must match.
		if len(allowFilters) > 0 {
			allowed := false
			for _, af := range allowFilters {
				if strings.Contains(domain, af) {
					allowed = true
					break
				}
			}
			if !allowed {
				continue
			}
		}

		// Check recency.
		if !minTime.IsZero() && result.Published.Before(minTime) {
			continue
		}
		filtered = append(filtered, result)
	}
	return filtered
}

// extractDomain extracts the domain from a URL.
func extractDomain(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "www.")
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return url
}

// CreateMockResults creates sample PageInfo instances for testing.
func CreateMockResults() []PageInfo {
	mockTime := time.Now()
	return []PageInfo{
		{
			URL:       "https://example.com/page1",
			Title:     "Example Page 1",
			Content:   "This is the content of page 1. It contains sample text.",
			Summary:   "Summary of page 1",
			Published: mockTime,
		},
		{
			URL:       "https://example.com/page2",
			Title:     "Example Page 2",
			Content:   "This is the content of page 2. More sample text here.",
			Summary:   "Summary of page 2",
			Published: mockTime.Add(-24 * time.Hour),
		},
		{
			URL:       "https://example.com/page3",
			Title:     "Example Page 3",
			Content:   "This is the content of page 3. Even more sample text.",
			Summary:   "Summary of page 3",
			Published: mockTime.Add(-30 * 24 * time.Hour),
		},
	}
}

// generateDummyResults returns a slice of n dummy PageInfo results.
func generateDummyResults(n int) []PageInfo {
	results := make([]PageInfo, n)
	for i := 0; i < n; i++ {
		results[i] = PageInfo{
			URL:       "http://example.com/" + strconv.Itoa(i),
			Title:     "Title " + strconv.Itoa(i),
			Content:   "Content " + strconv.Itoa(i),
			Summary:   "Summary " + strconv.Itoa(i),
			Published: time.Now(),
		}
	}
	return results
}

// NewDuckDuckGoSearchProvider returns a new DuckDuckGoSearchProvider.
func NewDuckDuckGoSearchProvider() SearchProvider {
	return &DuckDuckGoSearchProvider{}
}

// NewMockSearchProvider returns a new MockSearchProvider.
func NewMockSearchProvider() SearchProvider {
	return &MockSearchProvider{}
}
