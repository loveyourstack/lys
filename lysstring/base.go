package lysstring

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const (
	numbers  string = "0123456789"
	urlChars string = "-._~:/?#[]@!$&'()*+,;%="
)

// Convert changes value separation of s, e.g. from CSV (,) to TSV (|)
func Convert(s, inSep, outSep string, f func(string) string) (res string) {

	// split string by inSep
	sA := strings.Split(s, inSep)

	// if f is supplied, apply f() to all elements
	if f != nil {
		for i := range sA {
			sA[i] = f(sA[i])
		}
	}

	// join by outSep
	return strings.Join(sA, outSep)
}

// DeAlias converts []T to []string, where T is an alias of string
func DeAlias[T ~string](in []T) (out []string) {
	out = make([]string, len(in))
	for i, v := range in {
		out[i] = fmt.Sprintf("%s", v)
	}
	return out
}

// IsAscii returns true if s only contains ASCII chars
// from https://stackoverflow.com/questions/53069040/checking-a-string-contains-only-ascii-characters
func IsAscii(s string) bool {
	for i := range len(s) {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// RemoveCharacters returns input with the chars in charsToRemove removed
// from https://socketloop.com/tutorials/golang-remove-characters-from-string-example
func RemoveCharacters(input string, charsToRemove string) string {
	filter := func(r rune) rune {
		if !strings.ContainsRune(charsToRemove, r) {
			return r
		}
		return -1
	}

	return strings.Map(filter, input)
}

// SingleLines removes tabs and excess line breaks from s
// "a\n\n\tb\n" -> "a\nb"
// from https://stackoverflow.com/questions/35360080/golang-idiomatic-way-to-remove-a-blank-line-from-a-multi-line-string
func SingleLine(s string) string {
	return regexp.MustCompile(`[\t\r\n]+`).ReplaceAllString(strings.TrimSpace(s), "\n")
}

// SingleSpace removes excess tabs and spaces from s
// " a   b  " -> "a b"
// from https://stackoverflow.com/questions/37290693/how-to-remove-redundant-spaces-whitespace-from-a-string-in-golang
func SingleSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
