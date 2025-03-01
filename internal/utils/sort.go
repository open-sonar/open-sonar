package utils

import "sort"

// SortScored sorts a slice using the provided less function
func SortScored[T any](data []T, less func(i, j int) bool) {
	sort.Slice(data, less)
}
