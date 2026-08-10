// Package actest holds the behaviour tests shared by the ac and acascii
// packages. The two expose the same API and, for ASCII input, must behave
// identically; keeping one copy of the suite here is what stops them drifting.
//
// Anything that genuinely differs between the packages -- how they treat
// non-ASCII dictionaries and input -- and anything reaching into unexported
// state stays in each package's own test file.
package actest

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

// Matcher is the behaviour both packages' Matcher types provide.
type Matcher interface {
	FindAll(in []byte) [][]byte
	FindAllString(in string) []string
	Match(in []byte) bool
	MatchString(in string) bool
}

// Impl adapts one of the packages for the shared suite.
type Impl struct {
	CompileString     func(dictionary []string) (Matcher, error)
	Compile           func(dictionary [][]byte) (Matcher, error)
	MustCompileString func(dictionary []string) Matcher
	MustCompile       func(dictionary [][]byte) Matcher
}

// Case is one dictionary, one input, and the matches expected from it.
type Case struct {
	Name    string // matches original test name from cloudflare/ahocorasick
	Dict    []string
	Input   string
	Matches []string
}

// Cases are the shared expectations. Every dictionary is pure ASCII, so both
// packages must produce the same answer -- including for the one input that
// carries non-ASCII bytes, since neither package can match on them.
var Cases = []Case{
	{
		"TestNoPatterns",
		[]string{},
		"",
		nil,
	},
	{
		"TestNoData",
		[]string{"foo", "baz", "bar"},
		"",
		nil,
	},
	{
		"TestSuffixes",
		[]string{"Superman", "uperman", "perman", "erman"},
		"The Man Of Steel: Superman",
		[]string{"Superman", "uperman", "perman", "erman"},
	},
	{
		"TestPrefixes",
		[]string{"Superman", "Superma", "Superm", "Super"},
		"The Man Of Steel: Superman",
		[]string{"Super", "Superm", "Superma", "Superman"},
	},
	{
		"TestInterior",
		[]string{"Steel", "tee", "e"},
		"The Man Of Steel: Superman",
		[]string{"e", "tee", "Steel"},
	},
	{
		"TestMatchAtStart",
		[]string{"The", "Th", "he"},
		"The Man Of Steel: Superman",
		[]string{"Th", "The", "he"},
	},
	{
		"TestMatchAtEnd",
		[]string{"teel", "eel", "el"},
		"The Man Of Steel",
		[]string{"teel", "eel", "el"},
	},
	{
		"TestOverlappingPatterns",
		[]string{"Man ", "n Of", "Of S"},
		"The Man Of Steel",
		[]string{"Man ", "n Of", "Of S"},
	},
	{
		"TestMultipleMatches",
		[]string{"The", "Man", "an"},
		"A Man A Plan A Canal: Panama, which Man Planned The Canal",
		[]string{"Man", "an", "The"},
	},
	{
		"TestSingleCharacterMatches",
		[]string{"a", "M", "z"},
		"A Man A Plan A Canal: Panama, which Man Planned The Canal",
		[]string{"M", "a"}},
	{
		"TestNothingMatches",
		[]string{"baz", "bar", "foo"},
		"A Man A Plan A Canal: Panama, which Man Planned The Canal",
		nil,
	},
	{
		"Wikipedia1",
		[]string{"a", "ab", "bc", "bca", "c", "caa"},
		"abccab",
		[]string{"a", "ab", "bc", "c"},
	},
	{
		"Wikipedia2",
		[]string{"a", "ab", "bc", "bca", "c", "caa"},
		"bccab",
		[]string{"bc", "c", "a", "ab"},
	},
	{
		"Wikipedia3",
		[]string{"a", "ab", "bc", "bca", "c", "caa"},
		"bccb",
		[]string{"bc", "c"},
	},
	{
		"Browser1",
		[]string{"Mozilla", "Mac", "Macintosh", "Safari", "Sausage"},
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.101 Safari/537.36",
		[]string{"Mozilla", "Mac", "Macintosh", "Safari"},
	},
	{
		"Browser2",
		[]string{"Mozilla", "Mac", "Macintosh", "Safari", "Sausage"},
		"Mozilla/5.0 (Mac; Intel Mac OS X 10_7_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.101 Safari/537.36",
		[]string{"Mozilla", "Mac", "Safari"},
	},
	{
		"Browser3",
		[]string{"Mozilla", "Mac", "Macintosh", "Safari", "Sausage"},
		"Mozilla/5.0 (Moc; Intel Computer OS X 10_7_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.101 Safari/537.36",
		[]string{"Mozilla", "Safari"},
	},
	{
		"Browser4",
		[]string{"Mozilla", "Mac", "Macintosh", "Safari", "Sausage"},
		"Mozilla/5.0 (Moc; Intel Computer OS X 10_7_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.101 Sofari/537.36",
		[]string{"Mozilla"},
	},
	{
		"Browser5",
		[]string{"Mozilla", "Mac", "Macintosh", "Safari", "Sausage"},
		"Mazilla/5.0 (Moc; Intel Computer OS X 10_7_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.101 Sofari/537.36",
		nil,
	},
	{
		// this is to make sure backtracking works.  We get a partial
		// match of "Superwoman" with "Superman".  Then we need to make
		// sure that we restart the search and find "per".  Some implementations
		// had bugs that didn't backtrack (really start over) and didn't match
		// "per"
		"Backtrack",
		[]string{"Superwoman", "per"},
		"The Man Of Steel: Superman",
		[]string{"per"},
	},
	{
		"NotAsciiInput",
		[]string{"Mozilla", "Mac", "Macintosh", "Safari", "Sausage", "Gecko"},
		"Mazilla/5.0 \u0000 (Moc; Intel Computer OS X 10_7_5) AppleWebKit/537.36 \uFFFF (KHTML, like Gecko) Chrome/30.0.1599.101 Sofari/537.36",
		[]string{"Gecko"},
	},
}

// Run exercises every behaviour the two packages share.
func Run(t *testing.T, impl Impl) {
	t.Run("Cases", func(t *testing.T) { runCases(t, impl) })
	t.Run("Random", func(t *testing.T) { runRandom(t, impl) })
	t.Run("Must", func(t *testing.T) { runMust(t, impl) })
	t.Run("CounterWrap", func(t *testing.T) { runCounterWrap(t, impl) })
	t.Run("ByteSweep", func(t *testing.T) { runByteSweep(t, impl) })
}

// runByteSweep pushes every possible byte value through all four entry
// points. None may panic, and the string and []byte forms must agree.
//
// This is here because the pre-rewrite ASCII matcher clamped the input byte
// with > rather than >= in one of its four entry points, so the single byte
// 0x80 indexed one past the end of a 128-entry array. Three of the four
// entry points were correct, and the fixed test inputs happened to use other
// high bytes, so nothing caught it.
func runByteSweep(t *testing.T, impl Impl) {
	m := impl.MustCompileString([]string{"abc", "a", "\x00z"})
	mb := impl.MustCompile(ToBytes([]string{"abc", "a", "\x00z"}))

	for b := 0; b < 256; b++ {
		in := []byte{'a', byte(b), 'c', byte(b)}
		s := string(in)

		if got, want := m.MatchString(s), mb.Match(in); got != want {
			t.Errorf("byte %#02x: MatchString = %v, Match = %v", b, got, want)
		}

		hs := m.FindAllString(s)
		hb := mb.FindAll(in)
		if len(hs) != len(hb) {
			t.Errorf("byte %#02x: FindAllString = %q, FindAll = %q", b, hs, hb)
			continue
		}
		for i := range hs {
			if hs[i] != string(hb[i]) {
				t.Errorf("byte %#02x: hit %d, FindAllString = %q, FindAll = %q", b, i, hs[i], hb[i])
			}
		}
	}
}

// runCounterWrap checks that the duplicate-suppression markers from one call
// do not leak into the next.
func runCounterWrap(t *testing.T, impl Impl) {
	m := impl.MustCompileString([]string{"ab", "b", "abc"})
	want := m.FindAllString("xabcx")
	if len(want) == 0 {
		t.Fatal("expected matches")
	}

	// repeated calls must keep giving the same answer
	for i := 0; i < 3; i++ {
		if got := m.FindAllString("xabcx"); !reflect.DeepEqual(got, want) {
			t.Errorf("call %d: FindAllString = %q, want %q", i+2, got, want)
		}
	}
}

// runMust checks the panicking constructors agree with the others. The panic
// itself needs a dictionary the package rejects, so it is tested per package.
func runMust(t *testing.T, impl Impl) {
	dict := []string{"Superman", "uperman", "perman", "erman"}
	const input = "The Man Of Steel: Superman"
	want := NaiveFindAllString(dict, input)

	if got := impl.MustCompileString(dict).FindAllString(input); !reflect.DeepEqual(got, want) {
		t.Errorf("MustCompileString: FindAllString = %q, want %q", got, want)
	}

	var wantB [][]byte
	for _, s := range want {
		wantB = append(wantB, []byte(s))
	}
	if got := impl.MustCompile(ToBytes(dict)).FindAll([]byte(input)); !reflect.DeepEqual(got, wantB) {
		t.Errorf("MustCompile: FindAll = %q, want %q", got, wantB)
	}
}

// runCases checks the hand-written expectations, and checks the oracle
// against them too so that runRandom is standing on something trustworthy.
func runCases(t *testing.T, impl Impl) {
	for _, tt := range Cases {
		t.Run(tt.Name, func(t *testing.T) {
			if got := NaiveFindAllString(tt.Dict, tt.Input); !reflect.DeepEqual(got, tt.Matches) {
				t.Errorf("oracle FindAllString = %q, want %q", got, tt.Matches)
			}
			checkBoth(t, impl, tt.Dict, tt.Input, tt.Matches)
		})
	}
}

// runRandom hammers both entry points with small random alphabets, so that
// overlaps, nested suffixes and repeated matches all occur frequently, and
// compares against the oracle.
func runRandom(t *testing.T, impl Impl) {
	alphabets := []string{"ab", "abc", "abcd"}
	r := rand.New(rand.NewSource(1))

	for iter := 0; iter < 3000; iter++ {
		alpha := alphabets[r.Intn(len(alphabets))]

		dict := make([]string, 1+r.Intn(8))
		for i := range dict {
			dict[i] = randString(r, alpha, r.Intn(5)) // deliberately includes length 0
		}
		input := randString(r, alpha, r.Intn(40))

		checkBoth(t, impl, dict, input, NaiveFindAllString(dict, input))
		if t.Failed() {
			return
		}
	}
}

func randString(r *rand.Rand, alpha string, n int) string {
	var sb strings.Builder
	for i := 0; i < n; i++ {
		sb.WriteByte(alpha[r.Intn(len(alpha))])
	}
	return sb.String()
}

// checkBoth asserts that all four entry points agree with want.
func checkBoth(t *testing.T, impl Impl, dict []string, input string, want []string) {
	t.Helper()

	ms, err := impl.CompileString(dict)
	if err != nil {
		t.Fatalf("CompileString(%q): %s", dict, err)
	}
	mb, err := impl.Compile(ToBytes(dict))
	if err != nil {
		t.Fatalf("Compile(%q): %s", dict, err)
	}

	wantMatch := len(want) != 0

	if got := ms.FindAllString(input); !reflect.DeepEqual(got, want) {
		t.Errorf("dict %q input %q: FindAllString = %q, want %q", dict, input, got, want)
	}
	if got := ms.MatchString(input); got != wantMatch {
		t.Errorf("dict %q input %q: MatchString = %v, want %v", dict, input, got, wantMatch)
	}

	var wantB [][]byte
	for _, s := range want {
		wantB = append(wantB, []byte(s))
	}
	if got := mb.FindAll([]byte(input)); !reflect.DeepEqual(got, wantB) {
		t.Errorf("dict %q input %q: FindAll = %q, want %q", dict, input, got, wantB)
	}
	if got := mb.Match([]byte(input)); got != wantMatch {
		t.Errorf("dict %q input %q: Match = %v, want %v", dict, input, got, wantMatch)
	}
}

// NaiveFindAllString reproduces Matcher.FindAllString using an O(n*m) scan,
// sharing no machinery with the automaton it is used to check.
//
// A matcher reports each dictionary entry at most once per call. At every
// input position it walks the entries ending there from longest to shortest,
// stopping at the first one already reported -- except that an entry
// occupying the full current trie depth never stops the walk.
func NaiveFindAllString(dict []string, in string) []string {
	// stateDepth is the length of the longest suffix of in[:end] that is a
	// prefix of some dictionary entry, i.e. the depth of the trie node the
	// matcher sits on after consuming end bytes.
	stateDepth := func(end int) int {
		for l := end; l > 0; l-- {
			s := in[end-l : end]
			for _, d := range dict {
				if len(d) >= l && d[:l] == s {
					return l
				}
			}
		}
		return 0
	}

	reported := make(map[string]bool)
	var hits []string
	for end := 1; end <= len(in); end++ {
		depth := stateDepth(end)

		// entries ending at this position, longest first
		var ending []string
		for l := depth; l > 0; l-- {
			p := in[end-l : end]
			for _, d := range dict {
				if d == p {
					ending = append(ending, p)
					break
				}
			}
		}

		for i, p := range ending {
			if i == 0 && len(p) == depth {
				if !reported[p] {
					reported[p] = true
					hits = append(hits, p)
				}
				continue
			}
			if reported[p] {
				break
			}
			reported[p] = true
			hits = append(hits, p)
		}
	}
	return hits
}

// Normalize adjusts a fuzzer-generated dictionary and input into a form the
// package accepts and the oracle can predict. The full byte package needs no
// adjustment; the ASCII-only one clears the high bit of dictionary bytes so
// they compile, and folds high input bytes the way the matcher itself does,
// so that the oracle and the matcher are asked the same question.
type Normalize func(dict []string, input string) ([]string, string)

// decodeDict turns fuzzer bytes into a dictionary. Entries are length
// prefixed so that any byte, including NUL, can appear inside one.
func decodeDict(b []byte) []string {
	var out []string
	for len(b) > 0 && len(out) < 32 {
		n := int(b[0])
		b = b[1:]
		if n > len(b) {
			n = len(b)
		}
		out = append(out, string(b[:n]))
		b = b[n:]
	}
	return out
}

// Fuzz checks the matcher against the oracle on arbitrary dictionaries. The
// seeds cover the structures that are easy to get wrong -- shared prefixes,
// nested suffixes, repeats, entries that are suffixes of other entries --
// rather than any particular real dictionary.
func Fuzz(f *testing.F, impl Impl, norm Normalize) {
	seeds := [][]string{
		{"a"},
		{"a", "ab", "abc"},         // nested prefixes
		{"abc", "bc", "c"},         // nested suffixes
		{"ab", "ba"},               // overlapping
		{"aa"},                     // self overlapping
		{"a", ""},                  // empty entry
		{"a", "a"},                 // duplicate
		{"Superwoman", "per"},      // needs backtracking
		{"\x00", "\x00\x00"},       // NUL
		{"\xff\xfe", "\xfe"},       // high bytes
		{strings.Repeat("ab", 40)}, // long entry
	}
	inputs := []string{"", "a", "aa", "abcabc", "xabax", "\x00\x00", "\xff\xfe\xff", strings.Repeat("ab", 50)}

	for _, dict := range seeds {
		var enc []byte
		for _, e := range dict {
			enc = append(enc, byte(len(e)))
			enc = append(enc, e...)
		}
		for _, in := range inputs {
			f.Add(enc, in)
		}
	}

	f.Fuzz(func(t *testing.T, data []byte, input string) {
		dict := decodeDict(data)
		if len(dict) == 0 {
			return
		}
		dict, input = norm(dict, input)

		m, err := impl.CompileString(dict)
		if err != nil {
			t.Fatalf("CompileString(%q): %s", dict, err)
		}
		mb, err := impl.Compile(ToBytes(dict))
		if err != nil {
			t.Fatalf("Compile(%q): %s", dict, err)
		}

		want := NaiveFindAllString(dict, input)
		wantMatch := len(want) != 0

		if got := m.FindAllString(input); !reflect.DeepEqual(got, want) {
			t.Fatalf("dict %q input %q: FindAllString = %q, want %q", dict, input, got, want)
		}
		if got := m.MatchString(input); got != wantMatch {
			t.Fatalf("dict %q input %q: MatchString = %v, want %v", dict, input, got, wantMatch)
		}

		var wantB [][]byte
		for _, s := range want {
			wantB = append(wantB, []byte(s))
		}
		if got := mb.FindAll([]byte(input)); !reflect.DeepEqual(got, wantB) {
			t.Fatalf("dict %q input %q: FindAll = %q, want %q", dict, input, got, wantB)
		}
		if got := mb.Match([]byte(input)); got != wantMatch {
			t.Fatalf("dict %q input %q: Match = %v, want %v", dict, input, got, wantMatch)
		}

		// running twice must give the same answer: the duplicate-suppression
		// counters are matcher state that persists across calls
		if got := m.FindAllString(input); !reflect.DeepEqual(got, want) {
			t.Fatalf("dict %q input %q: second FindAllString = %q, want %q", dict, input, got, want)
		}
	})
}
