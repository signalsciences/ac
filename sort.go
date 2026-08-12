//go:build !tinygo

package ac

import (
	"bytes"
	"slices"
)

// sortStrings orders a dictionary lexicographically.
func sortStrings(a []string) {
	slices.Sort(a)
}

// sortBlices orders a dictionary lexicographically.
func sortBlices(a [][]byte) {
	slices.SortFunc(a, bytes.Compare)
}
