//go:build tinygo

package ac

import "bytes"

// The dictionary is sorted by the Shellsort implementations below rather than
// through the Go sort package.
//
// TinyGo evaluates package-level variables at compile time, so a matcher
// declared in one costs nothing at startup. However, TinyGo's interp phase
// gives up when the sort package's recursive pdqsort implementation kicks in,
// and then builds the whole matcher at startup instead. Plain loops it can
// evaluate.
//
// https://en.wikipedia.org/wiki/Shellsort
//
// The gaps are from Ciura, "Best Increments for the Average Case of Shellsort"
// (2001), extended above 701 by his recurrence h = floor(2.25h).
var shellGaps = [...]int{1750, 701, 301, 132, 57, 23, 10, 4, 1}

// sortStrings orders a dictionary lexicographically.
func sortStrings(a []string) {
	for _, gap := range shellGaps {
		if gap > len(a) {
			continue
		}
		for i := gap; i < len(a); i++ {
			v := a[i]
			j := i
			for ; j >= gap && a[j-gap] > v; j -= gap {
				a[j] = a[j-gap]
			}
			a[j] = v
		}
	}
}

// sortBlices orders a dictionary lexicographically.
func sortBlices(a [][]byte) {
	for _, gap := range shellGaps {
		if gap > len(a) {
			continue
		}
		for i := gap; i < len(a); i++ {
			v := a[i]
			j := i
			for ; j >= gap && bytes.Compare(a[j-gap], v) > 0; j -= gap {
				a[j] = a[j-gap]
			}
			a[j] = v
		}
	}
}
