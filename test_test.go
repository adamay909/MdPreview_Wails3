package main

import (
	"testing"
)

func TestCJK(t *testing.T) {
	tests := []struct {
		char     rune
		response bool
	}{
		{'あ', true},
		{'イ', true},
		{'兎', true},
		{'、', true},
		{'。', true},
		{'Ａ', true},
		{'Ｂ', true},
		{'Ｃ', true},
		{'１', true},
		{'２', true},
		{'３', true},
		{'a', false},
		{'B', false},
		{'c', false},
		{'1', false},
		{'2', false},
		{'3', false},
		{',', false},
	}
	for i := range tests {
		if got := isCJK(tests[i].char); got != tests[i].response {
			t.Errorf("want isCJK(%q)=%v, got %v", tests[i].char, tests[i].response, got)
		}
	}
}
