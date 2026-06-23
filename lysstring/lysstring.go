package lysstring

import (
	"regexp"
	"strings"
	"unicode"
)

// Convert changes value separation of s, e.g. from CSV (,) to TSV (|).
// If inSep is empty, s is split by character.
func Convert(s, inSep, outSep string, f func(string) string) (res string) {

	if inSep == outSep && f == nil {
		return s
	}

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

// DeAlias converts []T to []string, where T is an alias of string.
func DeAlias[T ~string](in []T) (out []string) {
	out = make([]string, len(in))
	for i, v := range in {
		out[i] = string(v)
	}
	return out
}

// IsAscii returns true if s only contains ASCII chars.
// From https://stackoverflow.com/questions/53069040/checking-a-string-contains-only-ascii-characters
func IsAscii(s string) bool {
	for i := range len(s) {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// RemoveCharacters returns input with the chars in charsToRemove removed.
func RemoveCharacters(input string, charsToRemove string) string {

	if input == "" || charsToRemove == "" {
		return input
	}

	removeSet := make(map[rune]bool, len(charsToRemove))
	for _, r := range charsToRemove {
		removeSet[r] = true
	}

	return strings.Map(func(r rune) rune {
		if _, ok := removeSet[r]; ok {
			return -1
		}
		return r
	}, input)
}

// SafeFileName returns a version of name that is safe to use as a file name, with ext appended. It replaces chars that are not letters, digits, hyphens, underscores, or periods with underscores, and trims leading and trailing periods and underscores.
// If the resulting name is empty, it returns "file" + ext.
func SafeFileName(name, ext string) string {

	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == ' ' || r == '.' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	n := strings.Trim(b.String(), "._")
	if n == "" {
		n = "file"
	}
	return n + ext
}

var singleLineRegexp = regexp.MustCompile(`[\t\r\n]+`)

// SingleLine removes tabs and excess line breaks from s.
// "a\n\n\tb\n" -> "a\nb".
// From https://stackoverflow.com/questions/35360080/golang-idiomatic-way-to-remove-a-blank-line-from-a-multi-line-string.
func SingleLine(s string) string {
	return singleLineRegexp.ReplaceAllString(strings.TrimSpace(s), "\n")
}

// SingleSpace removes excess tabs and spaces from s.
// " a   b  " -> "a b".
// From https://stackoverflow.com/questions/37290693/how-to-remove-redundant-spaces-whitespace-from-a-string-in-golang.
func SingleSpace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// StripPunct removes Unicode punctuation characters from s.
func StripPunct(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) {
			return -1
		}
		return r
	}, s)
}
