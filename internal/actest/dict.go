package actest

import (
	"math/rand"
	"strconv"
)

// ToBytes converts a dictionary for the []byte entry points.
func ToBytes(dict []string) [][]byte {
	out := make([][]byte, len(dict))
	for i, s := range dict {
		out[i] = []byte(s)
	}
	return out
}

// syllables are combined to build dictionary entries that share prefixes
// heavily, which is the interesting case for a trie.
var syllables = []string{
	"con", "tra", "ver", "sion", "pre", "fix", "suf", "ing", "ed", "log",
	"in", "out", "get", "set", "user", "admin", "data", "base", "sql", "select",
	"union", "script", "alert", "eval", "exec", "proc", "sys", "etc", "passwd", "bin",
}

// GenWords builds n dictionary entries out of syllables, so entries share
// prefixes heavily.
func GenWords(n, seed int) []string {
	r := rand.New(rand.NewSource(int64(seed)))
	seen := make(map[string]bool, n)
	out := make([]string, 0, n)
	for len(out) < n {
		w := ""
		for j := 0; j < 2+r.Intn(3); j++ {
			w += syllables[r.Intn(len(syllables))]
		}
		if seen[w] {
			w += strconv.Itoa(r.Intn(1000))
		}
		if seen[w] {
			continue
		}
		seen[w] = true
		out = append(out, w)
	}
	return out
}

// GenRandom builds n dictionary entries with almost no shared prefixes, the
// worst case for trie sizing.
func GenRandom(n, seed int) []string {
	const alpha = "abcdefghijklmnopqrstuvwxyz0123456789-_./"
	r := rand.New(rand.NewSource(int64(seed)))
	out := make([]string, n)
	buf := make([]byte, 24)
	for i := range out {
		l := 8 + r.Intn(16)
		for j := 0; j < l; j++ {
			buf[j] = alpha[r.Intn(len(alpha))]
		}
		out[i] = string(buf[:l])
	}
	return out
}

// pathWords are combined into URL-path-like entries, a common shape for
// dictionary entries.
var pathWords = []string{
	"admin", "login", "index", "panel", "config", "console", "manager",
	"account", "user", "system", "backup", "server", "control", "webmaster",
	"cgi-bin", "include", "upload", "private", "session", "database",
}

var pathExts = []string{"", "/", ".php", ".asp", ".aspx", ".html", ".jsp", ".cgi"}

// GenPaths builds n distinct URL-path-like entries.
func GenPaths(n, seed int) []string {
	r := rand.New(rand.NewSource(int64(seed)))
	seen := make(map[string]bool, n)
	out := make([]string, 0, n)
	for len(out) < n {
		p := ""
		for j := 0; j < 1+r.Intn(3); j++ {
			p += "/" + pathWords[r.Intn(len(pathWords))]
		}
		p += pathExts[r.Intn(len(pathExts))]
		if seen[p] {
			continue
		}
		seen[p] = true
		out = append(out, p)
	}
	return out
}

// CaseVariants returns every combination of upper and lower case for each
// word. It builds many entries over a small alphabet, which is a shape a
// case-sensitive matcher gets used for.
func CaseVariants(words ...string) []string {
	var out []string
	for _, w := range words {
		var letters []int
		for i := 0; i < len(w); i++ {
			if c := w[i]; c >= 'a' && c <= 'z' {
				letters = append(letters, i)
			}
		}
		for mask := 0; mask < 1<<uint(len(letters)); mask++ {
			b := []byte(w)
			for k, i := range letters {
				if mask&(1<<uint(k)) != 0 {
					b[i] -= 'a' - 'A'
				}
			}
			out = append(out, string(b))
		}
	}
	return out
}

// Dict is one benchmark dictionary, in both forms the API accepts.
type Dict struct {
	Name string
	Dict []string
	Blob [][]byte
}

// Dict1, Dict2 and Source1 back the original benchmarks carried over from
// cloudflare/ahocorasick.
var (
	Dict1   = []string{"Mozilla", "Mac", "Macintosh", "Safari", "Sausage"}
	Dict2   = []string{"Googlebot", "bingbot", "msnbot", "Yandex", "Baiduspider"}
	Source1 = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_7_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/30.0.1599.101 Safari/537.36"
)

// Dicts returns the dictionaries the compile and match benchmarks run over.
//
// The first five are the sizes and shapes seen in practice: a single entry to
// a few hundred, each 3-25 bytes, over an alphabet of 6-45 distinct bytes.
// The last two are headroom, not typical.
func Dicts() []Dict {
	sets := []struct {
		name string
		dict []string
	}{
		{"One1", []string{"file:///"}},
		{"Tiny5", Dict1},
		{"Dirs13", []string{"/.chef/", "/.copy", "/.git/", "/.ssh/", "/.svn/", "/.trash/", "/WEB-INF/", `\global.asa`, `\ProgramFiles\`, "/usr/local/", "/opt/", "/private/", "/var/www/"}},
		{"Case48", CaseVariants("http://", "https://")},
		{"Paths384", GenPaths(384, 4)},
		{"Words10k", GenWords(10000, 2)},
		{"Random5k", GenRandom(5000, 3)},
	}

	out := make([]Dict, len(sets))
	for i, s := range sets {
		out[i] = Dict{Name: s.name, Dict: s.dict, Blob: ToBytes(s.dict)}
	}
	return out
}

// Haystacks returns the inputs the match benchmarks scan, from a short
// string up to 64k.
func Haystacks() []struct {
	Name string
	In   string
} {
	r := rand.New(rand.NewSource(42))
	const alpha = "abcdefghijklmnopqrstuvwxyz /?=&."
	buf := make([]byte, 64*1024)
	for i := range buf {
		buf[i] = alpha[r.Intn(len(alpha))]
	}

	return []struct {
		Name string
		In   string
	}{
		{"Short", Source1},
		{"Req", "GET /store/cart/checkout?id=123&ref=%2Fhome&next=/account/login HTTP/1.1"},
		{"Long64k", string(buf)},
	}
}
