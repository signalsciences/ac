# ac

[![GoDoc](https://godoc.org/github.com/signalsciences/ac?status.svg)](https://godoc.org/github.com/signalsciences/ac) [![Actions Status](https://github.com/signalsciences/ac/workflows/ci/badge.svg)](https://github.com/signalsciences/ac/actions)

Golang implementation of Aho-Corasick for rapid substring matching on either byte
strings or ASCII strings.

This is based on the excellent library
[cloudflare/ahocorasick](https://github.com/cloudflare/ahocorasick) (BSD
License).  The fork/changes were needed for a specific application usages
that are incomptabile with the original library.  Some other minor optimizations 
around memory and setup were also done.


## Examples

* FindAllString

```
m := ac.MustCompileString([]string{"Superman", "uperman", "perman", "erman"})
matches := m.FindAllString("The Man Of Steel: Superman")
fmt.Println(matches)
```

Output:

```
[Superman uperman perman erman]
```

* MatchString

```
m := ac.MustCompileString([]string{"Superman", "uperman", "perman", "erman"})
contains := m.MatchString("The Man Of Steel: Superman")
fmt.Println(contains)
```

Output:

```
true
```

## Concurrency

`FindAll` and `FindAllString` are not safe to call concurrently on a shared
`Matcher`.  Give each goroutine its own, or serialize the calls.


## ac/acascii for pure ASCII matching

The `ac/acascii` package assumes the dictionary is all ASCII characters (0-127)
and returns `ErrNotASCII` otherwise.  Input bytes outside that range are folded
onto byte 0 rather than matched.

Previously this was about 50% faster and smaller than `ac`, but the two now
perform the same.  What remains is that `acascii` enforces ASCII and folds
higher input bytes, where `ac` matches the full byte range.


## IN PROGRESS

* Support for ASCII case-insensitive matching.
