# ac

[![GoDoc](https://godoc.org/github.com/signalsciences/ac?status.svg)](https://godoc.org/github.com/signalsciences/ac) [![Actions Status](https://github.com/signalsciences/ac/workflows/lint/badge.svg)](https://github.com/signalsciences/ac/actions)

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

## ac/acascii for pure ASCII matching

The `ac/acascii` package assumes the dictionary is all ASCII characters (1-127) without NULL bytes.  This results in during setup:

* 50% less memory allocations
* 50% less memory users
* 50% less CPU time

as compared to the plain `ac` package.


## IN PROGRESS

* Support for ASCII case-insensitive matching.
