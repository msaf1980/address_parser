package main

import (
	"strings"
	"unicode/utf8"
)

var (
	SpaceTrimmed = " \t"
)

func CutFunc(r rune) bool {
	return r < utf8.RuneSelf && strings.ContainsRune(SpaceTrimmed, r)
}

func TrimLeftAnyByte(s string, trim []byte) string {
	var trimPos int

	for i := range s {
		trimmed := false
		for _, t := range trim {
			if s[i] == t {
				trimmed = true
				trimPos = i + 1
				break
			}
		}
		if !trimmed {
			break
		}
	}

	return s[trimPos:]
}
