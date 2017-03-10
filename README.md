# ac
temp repo for fooling around with Aho-Corasick

This was based from the excellent library cloudflare/ahocorasick (BSD
License).  The fork/changes were needed for a specific application usages
that are incomptabile with the original library.

## NOTES

* This is designed for ASCII pattern matching at the byte level.
  There is no rune support, no UTF-8 support (other than ASCII).  
* Similar API to `regexp` package.
* Byte and String-based API work the same.  Again, there is no UTF-8 support.

## IN PROGRESS

* Current API allows for *overlapping* matches.  This is slow
  and potentially exponential and is different than how golang's
  regular expressions work.  This will be changed.
* Support for ASCII case-insensitive matching.

