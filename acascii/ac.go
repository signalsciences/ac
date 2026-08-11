// Package ac provides an implementation of the Aho-Corasick string matching
// algorithm. Throughout this code []byte is referred to
// as a blice.
//
// http://en.wikipedia.org/wiki/Aho%E2%80%93Corasick_string_matching_algorithm
//
// Copyright (c) 2013 CloudFlare, Inc.
//
// Originally from https://github.com/cloudflare/ahocorasick
package acascii

import (
	"errors"

	"github.com/signalsciences/ac"
)

// ErrNotASCII is returned when the dictionary input is not ASCII
var ErrNotASCII = errors.New("non-ASCII input")

// ErrTooLarge is returned when the dictionary is too large to compile
var ErrTooLarge = ac.ErrTooLarge

// ascii rejects a dictionary holding a byte outside ASCII, and folds such a
// byte onto byte 0 when it appears in the input.
var ascii = ac.Config{
	Limit:    128,
	ErrRange: ErrNotASCII,
}

// Matcher contains a list of blices to match against
type Matcher = ac.Matcher

// Compile creates a new Matcher using a list of []byte
func Compile(dictionary [][]byte) (*Matcher, error) {
	return ascii.Compile(dictionary)
}

// MustCompile returns a Matcher or panics
func MustCompile(dictionary [][]byte) *Matcher {
	m, err := Compile(dictionary)
	if err != nil {
		panic(err)
	}
	return m
}

// CompileString creates a new Matcher used to match against a set
// of strings (this is a helper to make initialization easy)
func CompileString(dictionary []string) (*Matcher, error) {
	return ascii.CompileString(dictionary)
}

// MustCompileString returns a Matcher or panics
func MustCompileString(dictionary []string) *Matcher {
	m, err := CompileString(dictionary)
	if err != nil {
		panic(err)
	}
	return m
}
