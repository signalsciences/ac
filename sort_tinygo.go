//go:build tinygo

package ac

import (
	"bytes"
	"sort"
)

// sortStrings orders a dictionary lexicographically.
//
// TinyGo evaluates package-level variables at compile time, so a matcher
// declared in one costs nothing at startup. Only sort.Slice keeps that
// working: TinyGo's interp phase gives up on the pdqsort implementation behind
// sort.Strings, slices.Sort and slices.SortFunc, and then builds the whole
// matcher at startup instead.
func sortStrings(a []string) {
	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
}

// sortBlices orders a dictionary lexicographically.
func sortBlices(a [][]byte) {
	sort.Slice(a, func(i, j int) bool { return bytes.Compare(a[i], a[j]) < 0 })
}
