package acascii

import (
	"runtime"
	"testing"

	"github.com/signalsciences/ac/internal/actest"
)

var benchDicts = actest.Dicts()

var benchHaystacks = actest.Haystacks()

var sinkMatcher *Matcher

func BenchmarkCompileString(b *testing.B) {
	for _, tc := range benchDicts {
		b.Run(tc.Name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m, err := CompileString(tc.Dict)
				if err != nil {
					b.Fatal(err)
				}
				sinkMatcher = m
			}
		})
	}
}

func BenchmarkCompile(b *testing.B) {
	for _, tc := range benchDicts {
		b.Run(tc.Name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m, err := Compile(tc.Blob)
				if err != nil {
					b.Fatal(err)
				}
				sinkMatcher = m
			}
		})
	}
}

// BenchmarkCompileRetained reports the live heap held by a compiled Matcher,
// which is what a long-lived process actually pays for.
func BenchmarkCompileRetained(b *testing.B) {
	for _, tc := range benchDicts {
		b.Run(tc.Name, func(b *testing.B) {
			var before, after runtime.MemStats
			runtime.GC()
			runtime.ReadMemStats(&before)
			m, err := CompileString(tc.Dict)
			if err != nil {
				b.Fatal(err)
			}
			runtime.GC()
			runtime.ReadMemStats(&after)
			b.ReportMetric(float64(after.HeapAlloc-before.HeapAlloc), "retained-B")
			b.ReportMetric(0, "ns/op")
			runtime.KeepAlive(m)
		})
	}
}

// this is to prevent optimizer tricks
var (
	sinkBool    bool
	sinkStrings []string
	sinkBlices  [][]byte
)

func BenchmarkMatchString(b *testing.B) {
	for _, tc := range benchDicts {
		m := MustCompileString(tc.Dict)
		for _, h := range benchHaystacks {
			b.Run(tc.Name+"/"+h.Name, func(b *testing.B) {
				b.SetBytes(int64(len(h.In)))
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					sinkBool = m.MatchString(h.In)
				}
			})
		}
	}
}

func BenchmarkMatch(b *testing.B) {
	for _, tc := range benchDicts {
		m := MustCompile(tc.Blob)
		for _, h := range benchHaystacks {
			in := []byte(h.In)
			b.Run(tc.Name+"/"+h.Name, func(b *testing.B) {
				b.SetBytes(int64(len(in)))
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					sinkBool = m.Match(in)
				}
			})
		}
	}
}

func BenchmarkFindAllString(b *testing.B) {
	for _, tc := range benchDicts {
		m := MustCompileString(tc.Dict)
		for _, h := range benchHaystacks {
			b.Run(tc.Name+"/"+h.Name, func(b *testing.B) {
				b.SetBytes(int64(len(h.In)))
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					sinkStrings = m.FindAllString(h.In)
				}
			})
		}
	}
}

func BenchmarkFindAll(b *testing.B) {
	for _, tc := range benchDicts {
		m := MustCompile(tc.Blob)
		for _, h := range benchHaystacks {
			in := []byte(h.In)
			b.Run(tc.Name+"/"+h.Name, func(b *testing.B) {
				b.SetBytes(int64(len(in)))
				b.ReportAllocs()
				for i := 0; i < b.N; i++ {
					sinkBlices = m.FindAll(in)
				}
			})
		}
	}
}

// The original benchmarks carried over from cloudflare/ahocorasick.
var (
	re1      = MustCompileString(actest.Dict1)
	re2      = MustCompileString(actest.Dict2)
	source1b = []byte(actest.Source1)
)

func BenchmarkAC1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBool = re1.MatchString(actest.Source1)
	}
}

func BenchmarkAC2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBool = re2.MatchString(actest.Source1)
	}
}

func BenchmarkAC2Byte(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBool = re2.Match(source1b)
	}
}
