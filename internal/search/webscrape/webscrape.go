package webscrape

import "time"

// PageInfo represents the data for a search result
type PageInfo struct {
	URL       string
	Title     string
	Content   string
	Summary   string
	Published time.Time
}

// Note: Other declarations moved to search_provider.go
