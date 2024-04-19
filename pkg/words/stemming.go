package words

import (
	"slices"
	"strings"
	"unicode"

	"github.com/kljensen/snowball/english"
	"golang.org/x/exp/maps"
)

func Stem(phrase string) []string {
	// split by spaces and unknown junk
	words := strings.FieldsFunc(phrase, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	// delete short or frequent words
	words = slices.DeleteFunc(words, func(s string) bool {
		return len(s) < 3 || english.IsStopWord(s)
	})
	// get unique set of stems
	stems := make(map[string]struct{})
	for _, word := range words {
		stems[english.Stem(word, false)] = struct{}{}
	}
	return maps.Keys(stems)
}
