package lysstring

import (
	"math/rand"
	"slices"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits int   = 6                    // 6 bits to represent a letter index
	letterIdxMask int64 = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  int   = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
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

// SlicesEqualUnordered returns true if the supplied string slices contain the same values in any order
// https://stackoverflow.com/questions/36000487/check-for-equality-on-slices-without-order
func SlicesEqualUnordered(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y] -= 1
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	return len(diff) == 0
}

// Title returns s in title case using non-specific language rules
func Title(s string) string {
	return cases.Title(language.Und, cases.NoLower).String(s)
}
