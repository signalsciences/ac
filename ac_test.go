package ac

import (
	"fmt"
	"reflect"
	"testing"
)

var cases = []struct {
	name    string // matches original test name from cloudflare/ahocorasick
	dict    []string
	input   string
	matches []string
}{
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
}

func TestAC(t *testing.T) {
	for _, tt := range cases {
		m, err := CompileString(tt.dict)
		if err != nil {
			t.Fatalf("%s:unable to compile %s", tt.name, err)
		}

		//
		matches := m.FindAllString(tt.input)
		if !reflect.DeepEqual(matches, tt.matches) {
			t.Errorf("%s: FindAllString want %v, got %v", tt.name, tt.matches, matches)
		}

		//
		contains := m.MatchString(tt.input)
		if contains {
			if len(tt.matches) == 0 {
				t.Errorf("%s: MatchString want false, but got true", tt.name)
			}
		} else {
			// does not contain, but got matches
			if len(tt.matches) != 0 {
				t.Errorf("%s: MatchString want true, but got false", tt.name)
			}
		}
	}
}

func TestACBlices(t *testing.T) {
	for _, tt := range cases {
		var dict [][]byte
		for _, d := range tt.dict {
			dict = append(dict, []byte(d))
		}
		m := MustCompile(dict)

		matches := m.FindAll([]byte(tt.input))
		var mb [][]byte
		for _, m := range matches {
			mb = append(mb, []byte(m))
		}
		if !reflect.DeepEqual(matches, mb) {
			t.Errorf("%s: FindAll = %v, want %v", tt.name, mb, matches)
		}

		contains := m.Match([]byte(tt.input))
		if contains {
			if len(tt.matches) == 0 {
				t.Errorf("%s: MatchString = true, want false", tt.name)
			}
		} else {
			// does not contain, but got matches
			if len(tt.matches) != 0 {
				t.Errorf("%s: Match = false, want true", tt.name)
			}
		}

	}
}

var (
	source1  = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.101 Safari/537.36"
	source1b = []byte(source1)
	dict1    = []string{"Mozilla", "Mac", "Macintosh", "Safari", "Sausage"}
	dict2    = []string{"Googlebot", "bingbot", "msnbot", "Yandex", "Baiduspider"}
	re1      = MustCompileString(dict1)
	re2      = MustCompileString(dict2)
)

// this is to prevent optimizer tricks
var result1 bool

func BenchmarkAC1(b *testing.B) {
	var result bool
	for i := 0; i < b.N; i++ {
		result = re1.MatchString(source1)
	}
	result1 = result
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

func BenchmarkAC2(b *testing.B) {
	var result bool
	for i := 0; i < b.N; i++ {
		result = re2.MatchString(source1)
	}
	result1 = result
}
func BenchmarkAC2Byte(b *testing.B) {
	var result bool
	for i := 0; i < b.N; i++ {
		result = re2.Match(source1b)
	}
	result1 = result
}
