package lysstring

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"time"
	"unicode"
	"unsafe"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// ContainsAny returns true if any string in elements is found in slice
func ContainsAny(slice []string, elements []string) bool {
	for _, v := range elements {
		if slices.Contains(slice, v) {
			return true
		}
	}
	return false
}

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
	for _, v := range in {
		out = append(out, fmt.Sprintf("%s", v))
	}
	return out
}

const (
	numbers  string = "0123456789"
	urlChars string = "-._~:/?#[]@!$&'()*+,;%="
)

// IsAscii returns true if s only contains ASCII chars
// from https://stackoverflow.com/questions/53069040/checking-a-string-contains-only-ascii-characters
func IsAscii(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

const (
	letterBytes   string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits int    = 6                    // 6 bits to represent a letter index
	letterIdxMask int64  = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  int    = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var randSrc = rand.NewSource(time.Now().UnixNano())

// RandString creates a random string
// from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func RandString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, randSrc.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = randSrc.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
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

const (
	deAccents string = "äöüß"
	frAccents string = "éàèùçâêîôûëïü"
)

// ReplaceAccents replaces accent characters such as "ö" with their non-accented equivalents such as "o"
// from https://twin.sh/articles/33/remove-accents-from-characters-in-go
func ReplaceAccents(s string) (res string, err error) {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	res, _, err = transform.String(t, s)
	return res, err
}

// Title returns s in title case using non-specific language rules
func Title(s string) string {
	return cases.Title(language.Und, cases.NoLower).String(s)
}
