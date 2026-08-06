// Package ac provides an implementation of the Aho-Corasick string matching
// algorithm. Throughout this code []byte is referred to
// as a blice.
//
// http://en.wikipedia.org/wiki/Aho%E2%80%93Corasick_string_matching_algorithm
//
// Copyright (c) 2013 CloudFlare, Inc.
//
// Originally from https://github.com/cloudflare/ahocorasick
package ac

import (
	"bytes"
	"errors"
	"math"
	"sort"
)

const maxchar = 256

// ErrTooLarge is returned when a dictionary needs an automaton whose row
// offsets would not fit in the int32 the transition table stores.
var ErrTooLarge = errors.New("dictionary too large")

// metaRow is the number of int32 slots each row carries after its transition
// columns: the dictionary-suffix link, the length of the entry ending at the
// state, and the duplicate-suppression counter.
const metaRow = 3

// Matcher contains a list of blices to match against
type Matcher struct {
	// table holds the entire automaton. Every state owns one row of
	// width+metaRow int32s, laid out as
	//
	//	[ transition columns ... | suffix | outLen | counter ]
	//
	// A state is identified by the offset of its row rather than by an index,
	// so a transition is table[s+column]: one add and one load, with no
	// multiply on the dependency chain that the scan loops are bound by.
	//
	// Folding the child pointers and the fail transitions into a single dense
	// goto table, and folding the per-state data into the same slice, is what
	// keeps the whole automaton down to one allocation.
	table []int32

	// width is the number of transition columns in a row.
	width int

	// alphabet maps an input byte to its transition column. Bytes occurring
	// in no dictionary entry share column 0, whose transition is always back
	// to the root, so the table only pays for the bytes actually used.
	alphabet [maxchar]uint16

	// starts reports whether a byte can begin a dictionary entry. While no
	// match is in progress the scanners use it to skip input without walking
	// the state machine.
	starts [maxchar]bool

	// counter counts the number of matches done, and is used to prevent
	// output of multiple matches of the same string
	counter int32
}

// setAlphabet gives every byte the dictionary uses its own transition column.
// Unused bytes keep column 0.
func (m *Matcher) setAlphabet(present *[maxchar]bool) {
	width := 1
	for b := 0; b < maxchar; b++ {
		if present[b] {
			m.alphabet[b] = uint16(width)
			width++
		}
	}
	m.width = width
}

// countNodesString returns the exact number of states a sorted dictionary
// needs: its number of distinct prefixes, plus the root. Sorting is what
// makes the count exact -- in lexicographic order an entry's longest common
// prefix with all earlier entries is its longest common prefix with its
// immediate predecessor.
func countNodesString(sorted []string) int {
	count := 1
	for i, s := range sorted {
		p := 0
		if i > 0 {
			prev := sorted[i-1]
			for p < len(s) && p < len(prev) && s[p] == prev[p] {
				p++
			}
		}
		count += len(s) - p
	}
	return count
}

// countNodesBytes is countNodesString for a sorted [][]byte dictionary.
func countNodesBytes(sorted [][]byte) int {
	count := 1
	for i, s := range sorted {
		p := 0
		if i > 0 {
			prev := sorted[i-1]
			for p < len(s) && p < len(prev) && s[p] == prev[p] {
				p++
			}
		}
		count += len(s) - p
	}
	return count
}

// blices orders a [][]byte dictionary lexicographically.
type blices [][]byte

func (b blices) Len() int           { return len(b) }
func (b blices) Less(i, j int) bool { return bytes.Compare(b[i], b[j]) < 0 }
func (b blices) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// buildTrie builds the fundamental trie structure from a set of blices.
//
// While the trie is being built the transition columns hold plain child
// links, where 0 means "no child" -- unambiguous, because the root's row sits
// at offset 0 and the root can never be a child. link then rewrites them into
// a complete goto table.
func (m *Matcher) buildTrie(dictionary [][]byte) error {
	var present [maxchar]bool
	for _, blice := range dictionary {
		for _, b := range blice {
			present[b] = true
		}
	}
	m.setAlphabet(&present)

	sorted := append(make(blices, 0, len(dictionary)), dictionary...)
	sort.Sort(sorted)

	stride := m.width + metaRow
	size := countNodesBytes(sorted) * stride
	if size > math.MaxInt32 {
		return ErrTooLarge
	}
	m.table = make([]int32, size)

	free := int32(stride)
	for _, blice := range sorted {
		cur := 0
		for _, b := range blice {
			i := cur + int(m.alphabet[b])
			t := m.table[i]
			if t == 0 {
				t = free
				free += int32(stride)
				m.table[i] = t
			}
			cur = int(t)
		}

		// cur now points at the state representing a dictionary entry. Empty
		// entries land on the root, which is never reported, so they are
		// skipped -- as they were before.
		if len(blice) > 0 {
			m.table[cur+m.width+1] = int32(len(blice))
		}
	}

	m.link()
	return nil
}

// buildTrieString builds the fundamental trie structure from a []string
func (m *Matcher) buildTrieString(dictionary []string) error {
	var present [maxchar]bool
	for _, s := range dictionary {
		for i := 0; i < len(s); i++ {
			present[s[i]] = true
		}
	}
	m.setAlphabet(&present)

	sorted := append(make([]string, 0, len(dictionary)), dictionary...)
	sort.Strings(sorted)

	stride := m.width + metaRow
	size := countNodesString(sorted) * stride
	if size > math.MaxInt32 {
		return ErrTooLarge
	}
	m.table = make([]int32, size)

	free := int32(stride)
	for _, s := range sorted {
		cur := 0
		for j := 0; j < len(s); j++ {
			i := cur + int(m.alphabet[s[j]])
			t := m.table[i]
			if t == 0 {
				t = free
				free += int32(stride)
				m.table[i] = t
			}
			cur = int(t)
		}

		if len(s) > 0 {
			m.table[cur+m.width+1] = int32(len(s))
		}
	}

	m.link()
	return nil
}

// link rewrites the child links left by the trie build into a complete goto
// table and fills in the dictionary-suffix links, in one breadth-first pass.
//
// Because the pass visits states in order of increasing depth, the row of a
// state's fail target is always fully converted by the time it is needed, so
// a missing child can simply inherit the fail target's transition. That is
// what removes the need for a separate fail table at match time.
func (m *Matcher) link() {
	w := m.width
	stride := int32(w + metaRow)
	count := len(m.table) / int(stride)

	// fail maps a state index to the row offset of its fail target. The
	// root's own row offset is 0, which is also the zero value, so states one
	// byte deep need no initialisation.
	fail := make([]int32, count)
	queue := make([]int32, 0, count)

	// The root's row already is a valid goto row: a missing child reads as 0,
	// which is the root itself.
	for c := 0; c < w; c++ {
		if t := m.table[c]; t != 0 {
			queue = append(queue, t/stride)
		}
	}

	for qi := 0; qi < len(queue); qi++ {
		row := int(queue[qi] * stride)
		frow := int(fail[queue[qi]])

		if m.table[frow+w+1] != 0 {
			// the fail target is itself a dictionary entry
			m.table[row+w] = int32(frow)
		} else {
			m.table[row+w] = m.table[frow+w]
		}

		for c := 0; c < w; c++ {
			if t := m.table[row+c]; t == 0 {
				m.table[row+c] = m.table[frow+c]
			} else {
				fail[t/stride] = m.table[frow+c]
				queue = append(queue, t/stride)
			}
		}
	}

	for b := 0; b < maxchar; b++ {
		m.starts[b] = m.table[int(m.alphabet[b])] != 0
	}
}

// nextCounter advances the duplicate-suppression counter. On the practically
// unreachable wrap it clears the per-state markers, so that a stale marker
// can never suppress a real match.
func (m *Matcher) nextCounter() int32 {
	m.counter++
	if m.counter <= 0 {
		stride := m.width + metaRow
		for i := m.width + 2; i < len(m.table); i += stride {
			m.table[i] = 0
		}
		m.counter = 1
	}
	return m.counter
}

// Compile creates a new Matcher using a list of []byte
func Compile(dictionary [][]byte) (*Matcher, error) {
	m := new(Matcher)
	if err := m.buildTrie(dictionary); err != nil {
		return nil, err
	}
	return m, nil
}

// MustCompile returns a Matcher or panics
func MustCompile(dictionary [][]byte) *Matcher {
	m, err := Compile(dictionary)
	if err != nil {
		panic(err)
	}
	return m
}

// CompileString creates a new Matcher used to match against a set
// of strings (this is a helper to make initialization easy)
func CompileString(dictionary []string) (*Matcher, error) {
	m := new(Matcher)
	if err := m.buildTrieString(dictionary); err != nil {
		return nil, err
	}
	return m, nil
}

// MustCompileString returns a Matcher or panics
func MustCompileString(dictionary []string) *Matcher {
	m, err := CompileString(dictionary)
	if err != nil {
		panic(err)
	}
	return m
}

// FindAll searches in for blices and returns all the blices found
// in the original dictionary.
//
// It is not safe to call concurrently on a shared Matcher.
func (m *Matcher) FindAll(in []byte) [][]byte {
	counter := m.nextCounter()
	var hits [][]byte

	table, w := m.table, m.width
	alphabet, starts := &m.alphabet, &m.starts

	s := 0
	for idx := 0; idx < len(in); {
		if s == 0 {
			// Nothing is partially matched, so any byte that cannot begin an
			// entry can be stepped over. These loads do not depend on one
			// another, unlike the transitions below, so this runs several
			// times faster than driving the state machine.
			for idx < len(in) && !starts[in[idx]] {
				idx++
			}
			if idx == len(in) {
				break
			}
		}

		s = int(table[s+int(alphabet[in[idx]])])
		idx++
		if s == 0 {
			continue
		}

		o := s + w
		if table[o+1] != 0 && table[o+2] != counter {
			table[o+2] = counter
			hits = append(hits, in[idx-int(table[o+1]):idx])
		}

		for table[o] != 0 {
			o = int(table[o]) + w
			if table[o+2] == counter {
				// There's no point working our way up the suffixes if
				// it's been done before for this call to Match. The
				// matches are already in hits.
				break
			}
			table[o+2] = counter
			hits = append(hits, in[idx-int(table[o+1]):idx])
		}
	}

	return hits
}

// FindAllString searches in for blices and returns all the blices (as strings) found as
// in the original dictionary.
//
// It is not safe to call concurrently on a shared Matcher.
func (m *Matcher) FindAllString(in string) []string {
	counter := m.nextCounter()
	var hits []string

	table, w := m.table, m.width
	alphabet, starts := &m.alphabet, &m.starts

	s := 0
	for idx := 0; idx < len(in); {
		if s == 0 {
			for idx < len(in) && !starts[in[idx]] {
				idx++
			}
			if idx == len(in) {
				break
			}
		}

		s = int(table[s+int(alphabet[in[idx]])])
		idx++
		if s == 0 {
			continue
		}

		o := s + w
		if table[o+1] != 0 && table[o+2] != counter {
			table[o+2] = counter
			hits = append(hits, in[idx-int(table[o+1]):idx])
		}

		for table[o] != 0 {
			o = int(table[o]) + w
			if table[o+2] == counter {
				break
			}
			table[o+2] = counter
			hits = append(hits, in[idx-int(table[o+1]):idx])
		}
	}

	return hits
}

// Match returns true if the input slice contains any subslices
func (m *Matcher) Match(in []byte) bool {
	table, w := m.table, m.width
	alphabet, starts := &m.alphabet, &m.starts

	s := 0
	for idx := 0; idx < len(in); {
		if s == 0 {
			for idx < len(in) && !starts[in[idx]] {
				idx++
			}
			if idx == len(in) {
				break
			}
		}

		s = int(table[s+int(alphabet[in[idx]])])
		idx++
		if s != 0 && (table[s+w+1] != 0 || table[s+w] != 0) {
			return true
		}
	}
	return false
}

// MatchString returns true if the input slice contains any subslices
func (m *Matcher) MatchString(in string) bool {
	table, w := m.table, m.width
	alphabet, starts := &m.alphabet, &m.starts

	s := 0
	for idx := 0; idx < len(in); {
		if s == 0 {
			for idx < len(in) && !starts[in[idx]] {
				idx++
			}
			if idx == len(in) {
				break
			}
		}

		s = int(table[s+int(alphabet[in[idx]])])
		idx++
		if s != 0 && (table[s+w+1] != 0 || table[s+w] != 0) {
			return true
		}
	}
	return false
}
