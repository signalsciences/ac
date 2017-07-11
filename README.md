# ac
Golang implimentation of Aho-Corasick for rapid substring matching on byte
strings.

[![GoDoc](https://godoc.org/github.com/signalsciences/ac?status.svg)](https://godoc.org/github.com/signalsciences/ac) [![Build Status](https://travis-ci.org/signalsciences/ac.svg?branch=master)](https://travis-ci.org/signalsciences/ac)

This is based on the excellent library
[cloudflare/ahocorasick](https://github.com/cloudflare/ahocorasick) (BSD
License).  The fork/changes were needed for a specific application usages
that are incomptabile with the original library.

## NOTES

* This is designed for ASCII pattern matching at the byte level.
  There is no rune support, no UTF-8 support (other than the ASCII subset).
* Similar API to `regexp` package.
* Byte and String-based API work the same.  Again, there is no UTF-8 support.

## IN PROGRESS

* Current API allows for *overlapping* matches.  This is slow
  and potentially exponential and is different than how golang's
  regular expressions work.  This will be changed.
* Support for ASCII case-insensitive matching.
