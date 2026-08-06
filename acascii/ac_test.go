package acascii

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
	States: func(m actest.Matcher) int {
		mm := m.(*Matcher)
		return len(mm.table) / (mm.width + metaRow)
	},
	ExhaustCounter: func(m actest.Matcher) { m.(*Matcher).counter = math.MaxInt32 },
}

// TestShared runs the behaviour this package has in common with ac.
func TestShared(t *testing.T) {
	actest.Run(t, impl)
}

// TestNonASCIIDictionary rejects dictionaries with bytes this package cannot
// index. The ac package accepts them.
func TestNonASCIIDictionary(t *testing.T) {
	if _, err := CompileString([]string{"hello world", "こんにちは世界"}); err != ErrNotASCII {
		t.Errorf("CompileString err = %v, want %v", err, ErrNotASCII)
	}
	if _, err := Compile([][]byte{[]byte("ok"), {0x80}}); err != ErrNotASCII {
		t.Errorf("Compile err = %v, want %v", err, ErrNotASCII)
	}
}

// TestNonASCIIInput pins the long-standing behaviour that an input byte
// outside ASCII is folded onto byte 0 rather than rejected.
func TestNonASCIIInput(t *testing.T) {
	m := MustCompileString([]string{"a\x00b", "cd"})
	tests := []struct {
		in   string
		want []string
	}{
		{"a\x00b", []string{"a\x00b"}},
		// hits are subslices of the input, so the folded byte is echoed back
		{"a\xffb", []string{"a\xffb"}},
		{"a\xc3\xa9b", nil}, // two folded bytes, so no match
		{"x\xffcd", []string{"cd"}},
	}
	for _, tt := range tests {
		if got := m.FindAllString(tt.in); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("FindAllString(%q) = %q, want %q", tt.in, got, tt.want)
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

// TestMustCompilePanics covers the panic path of the Must constructors, which
// only this package can reach with a dictionary it rejects.
func TestMustCompilePanics(t *testing.T) {
	for _, tt := range []struct {
		name string
		fn   func()
	}{
		{"MustCompileString", func() { MustCompileString([]string{"héllo"}) }},
		{"MustCompile", func() { MustCompile([][]byte{{0x80}}) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != ErrNotASCII {
					t.Errorf("recover() = %v, want %v", r, ErrNotASCII)
				}
			}()
			tt.fn()
			t.Error("did not panic")
		})
	}
}
