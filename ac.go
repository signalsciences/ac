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

// ErrTooLarge is returned when the dictionary is too large to compile
var ErrTooLarge = errors.New("dictionary too large")

// Config describes the alphabet a Matcher accepts. It exists so that the
// acascii package can restrict the dictionary to ASCII and fold higher input
// bytes onto byte 0, without duplicating the automaton. Callers of this
// package want Compile or CompileString instead.
type Config struct {
	// Limit is one past the highest byte a dictionary entry may hold
	Limit int

	// ErrRange is returned for a dictionary byte at or above Limit
	ErrRange error

	// FoldHigh folds input bytes at or above Limit onto byte 0
	FoldHigh bool
}

// fullByte allows every byte, which is what this package itself uses.
var fullByte = Config{Limit: maxchar}

// metaRow is the number of int32 slots each row carries after its transition
// columns: the suffix link, the length of any entry ending here, and the
// counter.
const metaRow = 3

// Matcher contains a list of blices to match against
type Matcher struct {
	cfg Config

	// table holds the entire automaton, one row per state:
	//
	//	[ transition columns ... | suffix | outLen | counter ]
	//
	// A state is the offset of its row, so a transition is
	// table[s+column]: one add and one load, with no multiply.
	table []int32

	// width is the number of transition columns in a row
	width int

	// alphabet maps an input byte to its transition column. Bytes in no
	// dictionary entry share column 0, which always returns to the root, so
	// the table only pays for the bytes actually used.
	alphabet [maxchar]uint16

	// starts reports whether a byte can begin a dictionary entry, so that
	// the scanners can skip input without walking the state machine
	starts [maxchar]bool

	// counter counts the number of matches done, and is used to
	// prevent output of multiple matches of the same string
	counter int32
}

// setAlphabet gives every byte the dictionary uses its own transition column.
// Unused bytes keep column 0.
func (m *Matcher) setAlphabet(present *[maxchar]bool) {
	width := 1
	for b := 0; b < m.cfg.Limit; b++ {
		if present[b] {
			m.alphabet[b] = uint16(width)
			width++
		}
	}

	if m.cfg.FoldHigh {
		for b := m.cfg.Limit; b < maxchar; b++ {
			m.alphabet[b] = m.alphabet[0]
		}
	}

	m.width = width
}

// countNodesString returns the number of states a sorted dictionary needs:
// its number of distinct prefixes, plus the root. Sorting makes the count
// exact, because in lexicographic order an entry's longest common prefix with
// all earlier entries is its longest common prefix with its predecessor.
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

// buildTrie builds the fundamental trie structure from a set of
// blices.
//
// While building, the transition columns hold plain child links, where 0
// means "no child": the root's row is at offset 0, so it is never a child.
// link then rewrites them into a complete goto table.
func (m *Matcher) buildTrie(dictionary [][]byte) error {
	var present [maxchar]bool
	for _, blice := range dictionary {
		for _, b := range blice {
			if int(b) >= m.cfg.Limit {
				return m.cfg.ErrRange
			}
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

		// cur now points at the state representing a dictionary
		// entry. Empty entries land on the root, which is never reported.
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
			if int(s[i]) >= m.cfg.Limit {
				return m.cfg.ErrRange
			}
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

// link rewrites the child links into a complete goto table and fills in the
// suffix links, in one breadth-first pass.
//
// The pass visits states in order of increasing depth, so the row of a
// state's fail target is already converted when it is needed, and a missing
// child can inherit the fail target's transition.
func (m *Matcher) link() {
	w := m.width
	stride := int32(w + metaRow)
	count := len(m.table) / int(stride)

	// fail maps a state index to the row offset of its fail target. The
	// root's offset is 0, which is also the zero value, so states one byte
	// deep need no initialisation.
	fail := make([]int32, count)
	queue := make([]int32, 0, count)

	// The root's row is already a valid goto row: a missing child reads as
	// 0, which is the root itself.
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

// nextCounter advances the counter. On the practically unreachable wrap it
// clears the per-state markers, so a stale marker cannot suppress a match.
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

// Compile creates a new Matcher over cfg's alphabet using a list of []byte
func (cfg Config) Compile(dictionary [][]byte) (*Matcher, error) {
	m := &Matcher{cfg: cfg}
	if err := m.buildTrie(dictionary); err != nil {
		return nil, err
	}
	return m, nil
}

// CompileString creates a new Matcher over cfg's alphabet using a []string
func (cfg Config) CompileString(dictionary []string) (*Matcher, error) {
	m := &Matcher{cfg: cfg}
	if err := m.buildTrieString(dictionary); err != nil {
		return nil, err
	}
	return m, nil
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
			// Nothing is partially matched, so any byte that cannot
			// begin an entry can be stepped over.
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

// Compile creates a new Matcher using a list of []byte
func Compile(dictionary [][]byte) (*Matcher, error) {
	return fullByte.Compile(dictionary)
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
	return fullByte.CompileString(dictionary)
}

// MustCompileString returns a Matcher or panics
func MustCompileString(dictionary []string) *Matcher {
	m, err := CompileString(dictionary)
	if err != nil {
		panic(err)
	}
	return m
}
