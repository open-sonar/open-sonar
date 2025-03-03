package webscrape

import "time"

type PageInfo struct {
	URL       string
	Title     string
	Content   string
	Summary   string
	Published time.Time
}

type SearchOptions struct {
	MaxPages            int
	MaxRetries          int
	SearchDomainFilter  []string
	SearchRecencyFilter string
}
