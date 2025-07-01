package lysstring

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Camel removes accents and spaces in s and converts it to camel case
// My string -> myString
func Camel(s string) (res string) {
	return FirstLower(Pascal(s))
}

// FirstLower changes the first character of s to lower case
// from https://stackoverflow.com/questions/75988064/make-first-letter-of-string-lower-case-in-golang
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

// Joined removes accents and spaces in s and converts it to lower case
// My string -> mystring
func Joined(s string) (res string) {

	var err error
	res, err = ReplaceAccents(s)
	if err != nil {
		// should never happen
		panic("ReplaceAccents failed:" + err.Error())
	}

	// lower case, no separation
	return Convert(res, " ", "", strings.ToLower)
}

// Kebab removes accents in s and changes it to lower case with hyphen separation
// My string -> my-string
func Kebab(name string) (res string) {

	var err error
	res, err = ReplaceAccents(name)
	if err != nil {
		// should never happen
		panic("ReplaceAccents failed:" + err.Error())
	}

	// lower case, hyphen separation
	return Convert(res, " ", "-", strings.ToLower)
}

// Pascal removes accents and spaces in s and changes it to Pascal case
// My string -> MyString
func Pascal(s string) (res string) {

	var err error
	res, err = ReplaceAccents(s)
	if err != nil {
		// should never happen
		panic("ReplaceAccents failed:" + err.Error())
	}

	// title case, no separation
	return Convert(res, " ", "", Title)
}

// Snake removes accents in s and changes it to lower case with underscore separation
// My string -> my_string
func Snake(s string) (res string) {

	var err error
	res, err = ReplaceAccents(s)
	if err != nil {
		// should never happen
		panic("ReplaceAccents failed:" + err.Error())
	}

	// lower case, snake separation
	return Convert(res, " ", "_", strings.ToLower)
}

// Title returns s in title case using non-specific language rules
func Title(s string) string {
	return cases.Title(language.Und, cases.NoLower).String(s)
}
