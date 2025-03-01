package webscrape

import (
	"fmt"
	"strings"
	"time"
)

// SearchOptions represents options for search providers
type SearchOptions struct {
	MaxPages            int
	MaxRetries          int
	SearchDomainFilter  []string
	SearchRecencyFilter string
}

// SearchProvider is an interface for different search engines
type SearchProvider interface {
	Search(query string, options SearchOptions) ([]PageInfo, error)
}

// GetSearchProviderFunc is a function type for creating search providers
type GetSearchProviderFunc func(provider string) (SearchProvider, error)

// DefaultGetSearchProvider is the default implementation
var DefaultGetSearchProvider GetSearchProviderFunc = func(provider string) (SearchProvider, error) {
	switch strings.ToLower(provider) {
	case "duckduckgo":
		return &DuckDuckGoSearchProvider{}, nil
	// Add other search providers as needed
	default:
		return &DuckDuckGoSearchProvider{}, nil
	}
}

// CurrentGetSearchProvider holds the current provider implementation
var CurrentGetSearchProvider GetSearchProviderFunc = DefaultGetSearchProvider

// GetSearchProvider returns a search provider based on provider name
func GetSearchProvider(provider string) (SearchProvider, error) {
	return CurrentGetSearchProvider(provider)
}

// SetGetSearchProvider allows tests to replace the provider lookup function
func SetGetSearchProvider(fn GetSearchProviderFunc) GetSearchProviderFunc {
	old := CurrentGetSearchProvider
	CurrentGetSearchProvider = fn
	return old
}

// RestoreDefaultGetSearchProvider restores the default provider lookup
func RestoreDefaultGetSearchProvider() {
	CurrentGetSearchProvider = DefaultGetSearchProvider
}

// RecencyToTime converts a recency filter to a time.Time
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
		return time.Time{}, nil // No filter
	default:
		return time.Time{}, fmt.Errorf("invalid recency filter: %s", recency)
	}
}

// FilterResults filters results based on domain filters and recency
func FilterResults(results []PageInfo, options SearchOptions) []PageInfo {
	if len(options.SearchDomainFilter) == 0 && options.SearchRecencyFilter == "" {
		return results
	}

	var filtered []PageInfo

	// Process domain filters
	allowDomains := make(map[string]bool)
	blockedDomains := make(map[string]bool)

	for _, domain := range options.SearchDomainFilter {
		if strings.HasPrefix(domain, "-") {
			blockedDomains[strings.TrimPrefix(domain, "-")] = true
		} else {
			allowDomains[domain] = true
		}
	}

	// Process recency filter
	var minTime time.Time
	if options.SearchRecencyFilter != "" {
		var err error
		minTime, err = RecencyToTime(options.SearchRecencyFilter)
		if err != nil {
			// If invalid recency filter, ignore it
			minTime = time.Time{}
		}
	}

	for _, result := range results {
		// Skip if the domain is blocked
		domain := extractDomain(result.URL)
		if blockedDomains[domain] {
			continue
		}

		// Check if we should only allow specific domains
		if len(allowDomains) > 0 && !allowDomains[domain] {
			continue
		}

		// Check recency if specified
		if !minTime.IsZero() && result.Published.Before(minTime) {
			continue
		}

		filtered = append(filtered, result)
	}

	return filtered
}

// extractDomain extracts the domain from a URL
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
