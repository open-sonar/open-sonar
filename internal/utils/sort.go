package utils

// Note: The SortScored function has been moved to format.go
// This file is being kept as a placeholder in case we need to add
// more specialized sorting functions in the future

// SortByRelevance sorts items by their relevance score (high to low)
func SortByRelevance[T any](items []T, getScore func(T) float64) {
	SortScored(items, func(i, j int) bool {
		return getScore(items[i]) > getScore(items[j])
	})
}

// SortByRecency sorts items by their timestamp (newest first)
func SortByRecency[T any](items []T, getTimestamp func(T) int64) {
	SortScored(items, func(i, j int) bool {
		return getTimestamp(items[i]) > getTimestamp(items[j])
	})
}
