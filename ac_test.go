package ac

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/signalsciences/ac/internal/actest"
)

// impl adapts this package for the shared behaviour suite.
var impl = actest.Impl{
	CompileString: func(dictionary []string) (actest.Matcher, error) {
		m, err := CompileString(dictionary)
		if err != nil {
			return nil, err
		}
		return m, nil
	},
	Compile: func(dictionary [][]byte) (actest.Matcher, error) {
		m, err := Compile(dictionary)
		if err != nil {
			return nil, err
		}
		return m, nil
	},
	MustCompileString: func(dictionary []string) actest.Matcher { return MustCompileString(dictionary) },
	MustCompile:       func(dictionary [][]byte) actest.Matcher { return MustCompile(dictionary) },
}

// TestShared runs the behaviour this package has in common with acascii.
func TestShared(t *testing.T) {
	actest.Run(t, impl)
}

// TestNonASCII covers the full byte range, which this package indexes and
// acascii does not.
func TestNonASCII(t *testing.T) {
	dict := []string{"héllo", "日本", "\x00\xff", "lo\xff", "本語"}
	inputs := []string{
		"say héllo to 日本語 and \x00\xff bytes",
		"héllo\xffhéllo",
		"日本語日本",
		"",
	}
	for _, in := range inputs {
		m, err := CompileString(dict)
		if err != nil {
			t.Fatalf("CompileString: %s", err)
		}
		want := actest.NaiveFindAllString(dict, in)
		if got := m.FindAllString(in); !reflect.DeepEqual(got, want) {
			t.Errorf("input %q: FindAllString = %q, want %q", in, got, want)
		}
	}
}

func ExampleMatcher_FindAllString() {
	m := MustCompileString([]string{"Superman", "uperman", "perman", "erman"})
	matches := m.FindAllString("The Man Of Steel: Superman")
	fmt.Println(matches)
	// Output: [Superman uperman perman erman]
}

func ExampleMatcher_MatchString() {
	m := MustCompileString([]string{"Superman", "uperman", "perman", "erman"})
	contains := m.MatchString("The Man Of Steel: Superman")
	fmt.Println(contains)
	// Output: true
}

// FuzzMatcher checks the matcher against the brute-force oracle. This package
// indexes every byte, so nothing needs adjusting.
func FuzzMatcher(f *testing.F) {
	actest.Fuzz(f, impl, func(dict []string, input string) ([]string, string) {
		return dict, input
	})
}

// TestExactSizing checks the transition table is sized to the states the
// dictionary needs, with no slack. Sizing it to the sum of the entry lengths
// instead was the original source of the compile-time blowup, so this counts
// the distinct prefixes independently of countNodesString.
func TestExactSizing(t *testing.T) {
	dicts := [][]string{
		{},
		{""},
		{"a"},
		{"a", "a"},
		{"abc", "abd", "abe"},
		{"Mozilla", "Mac", "Macintosh", "Safari", "Sausage"},
		actest.GenWords(500, 7),
	}
	for _, dict := range dicts {
		m := MustCompileString(dict)

		seen := make(map[string]bool)
		for _, d := range dict {
			for i := 1; i <= len(d); i++ {
				seen[d[:i]] = true
			}
		}
		want := len(seen) + 1 // plus the root

		if got := len(m.table) / (m.width + metaRow); got != want {
			t.Errorf("dict of %d entries: %d states, want %d", len(dict), got, want)
		}
	}
}

// TestCounterWrap exercises the reset performed when the counter overflows,
// which is unreachable in practice but would silently drop matches.
func TestCounterWrap(t *testing.T) {
	m := MustCompileString([]string{"ab", "b", "abc"})
	want := m.FindAllString("xabcx")
	if len(want) == 0 {
		t.Fatal("expected matches")
	}

	m.counter = math.MaxInt32
	for i := 0; i < 2; i++ {
		if got := m.FindAllString("xabcx"); !reflect.DeepEqual(got, want) {
			t.Errorf("call %d after wrap: FindAllString = %q, want %q", i+1, got, want)
		}
	}
	if m.counter != 2 {
		t.Errorf("counter = %d, want 2", m.counter)
	}
}
