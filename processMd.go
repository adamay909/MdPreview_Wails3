package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// remove line breaks (\n) between CJK multibyte chars
// to prevent Safari and the like from inserting spurious
// spaces
func removeWraps(in string) string {
	b := new(strings.Builder)

	var prevChar rune

	for i, c := range in {
		if c != '\n' {
			b.WriteRune(c)
			prevChar = c
			continue
		}
		if !isCJK(prevChar) {
			b.WriteRune(c)
			prevChar = c
			continue
		}
		j := i + utf8.RuneLen(c)
		if j >= len(in) {
			continue
		}
		if n, _ := utf8.DecodeRuneInString(in[j:]); !isCJK(n) {
			b.WriteRune(c)
			prevChar = c
			continue
		}
	}
	return b.String()
}

func isCJK(r rune) bool {
	// unicode.Is(unicode.Han, r) returns true for CJK unified ideographs
	switch {
	case unicode.Is(unicode.Han, r):
		return true
	case unicode.Is(unicode.Hiragana, r):
		return true
	case unicode.Is(unicode.Katakana, r):
		return true
	default:
		return false
	}
}
