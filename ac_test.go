package ac

import (
	"reflect"
	"testing"
)

func TestAC(t *testing.T) {
	cases := []struct {
		dict    []string
		input   string
		matches []string
	}{
		//		{[]string{}, "", nil},
		{[]string{"foo", "baz", "bar"}, "", []string{}},
		{[]string{"bar"}, "foobar", []string{"bar"}},
		{[]string{"Superman", "uperman", "perman", "erman"}, "The Man Of Steel: Superman", []string{"Superman"}},
		{[]string{"Superman", "Superma", "Superm", "Super"}, "The Man Of Steel: Superman", []string{"Super", "Superm", "Superma", "Superman"}},
		{[]string{"Steel", "tee", "e"}, "The Man Of Steel: Superman", []string{"e", "Steel", "e"}},
	}

	for pos, tt := range cases {
		m, err := CompileStrings(tt.dict)
		if err != nil {
			t.Fatalf("Unable to compile case %d: %s", pos, err)
		}

		//
		matches := m.FindAllString(tt.input)
		if !reflect.DeepEqual(matches, tt.matches) {
			t.Errorf("Case %d: FindAllString want %v, got %v", pos, tt.matches, matches)
		}

		//
		contains := m.MatchString(tt.input)
		if contains {
			if len(tt.matches) == 0 {
				t.Errorf("Case %d: MatchString want false, but got true")
			}
		} else {
			// does not contain, but got matches
			if len(tt.matches) != 0 {
				t.Errorf("Case %d: MatchString want true, but got false")
			}
		}
	}
}

var source1 = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.101 Safari/537.36"
var dict1 = []string{"Mozilla", "Mac", "Macintosh", "Safari", "Sausage"}
var dict2 = []string{"Googlebot", "bingbot", "msnbot", "Yandex", "Baiduspider"}
var re1 = MustCompileStrings(dict1)
var re2 = MustCompileStrings(dict2)

// this is to prevent optimizer tricks
var result1 bool

func BenchmarkAC1(b *testing.B) {
	var result bool
	for i := 0; i < b.N; i++ {
		result = re1.MatchString(source1)
	}
	result1 = result
}

func BenchmarkAC2(b *testing.B) {
	var result bool
	for i := 0; i < b.N; i++ {
		result = re2.MatchString(source1)
	}
	result1 = result
}
