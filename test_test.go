package main

import (
	"fmt"
	"testing"
)

func TestCJK(t *testing.T) {

	s := "あイ兎、。"

	for _, c := range s {
		fmt.Println(c, isCJK(c))
	}

}

func TestRemoveWraps(t *testing.T) {
	s := `わたしは
猫ではない。`

	fmt.Println(s)

	fmt.Println()
	fmt.Println(removeWraps(s))
}
