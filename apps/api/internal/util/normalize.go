package util

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func Normalize(s string) string {
	s = strings.ToLower(s)
	s = removeAccents(s)
	s = replacePunctuation(s, " ")
	s = strings.Join(strings.Fields(s), " ")
	return s
}

func removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

func replacePunctuation(s string, repl string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) || unicode.IsSymbol(r) {
			return []rune(repl)[0]
		}
		return r
	}, s)
}
