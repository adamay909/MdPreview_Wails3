package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// remove line breaks (\n) between CJK multibyte chars
// to prevent Webkit from inserting spurious
// spaces if lines are hardwrapped.
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
	case (0x2FFF < r) && (r < 0x3040): //日本語句読点
		return true
	case (0xFF00 < r) && (r < 0xFF5B): //全角英数字
		return true
	case (0xFF60 < r) && (r < 0xFFA0): //半角カナ
		return true
	default:
		return false
	}
}
