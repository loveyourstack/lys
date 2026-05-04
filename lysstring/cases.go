package lysstring

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Camel removes punctuation, accents, and spaces in s and converts it to camel case
// My string -> myString
func Camel(s string) (res string) {
	return FirstLower(Pascal(s))
}

// FirstLower changes the first character of s to lower case.
// From https://stackoverflow.com/questions/75988064/make-first-letter-of-string-lower-case-in-golang.
func FirstLower(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToLower(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}

// Joined removes punctuation, accents, and spaces in s and converts it to lower case
// My string -> mystring
func Joined(s string) (res string) {

	res, _ = ReplaceAccents(s)
	res = StripPunct(res)

	// lower case, no separation
	sFields := strings.Fields(res)
	return Convert(strings.Join(sFields, " "), " ", "", strings.ToLower)
}

// Kebab removes punctuation and accents in s and changes it to lower case with hyphen separation
// My string -> my-string
func Kebab(name string) (res string) {

	res, _ = ReplaceAccents(name)

	// replace hyphens with spaces so they don't get stripped by StripPunct and lost as separators
	res = strings.ReplaceAll(res, "-", " ")

	res = StripPunct(res)

	// lower case, hyphen separation
	sFields := strings.Fields(res)
	return Convert(strings.Join(sFields, " "), " ", "-", strings.ToLower)
}

// Pascal removes punctuation, accents, and spaces in s and changes it to Pascal case
// My string -> MyString
func Pascal(s string) (res string) {

	res, _ = ReplaceAccents(s)
	res = StripPunct(res)

	// title case, no separation
	sFields := strings.Fields(res)
	return Convert(strings.Join(sFields, " "), " ", "", Title)
}

// Snake removes punctuation and accents in s and changes it to lower case with underscore separation
// My string -> my_string
func Snake(s string) (res string) {

	res, _ = ReplaceAccents(s)
	res = StripPunct(res)

	// lower case, snake separation
	sFields := strings.Fields(res)
	return Convert(strings.Join(sFields, " "), " ", "_", strings.ToLower)
}

var titleCaser = cases.Title(language.Und, cases.NoLower)

// Title returns s in title case using non-specific language rules
func Title(s string) string {
	return titleCaser.String(s)
}
