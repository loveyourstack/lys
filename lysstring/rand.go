package lysstring

import "math/rand/v2"

const (
	letterBytes   string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits int    = 6                    // 6 bits to represent a letter index
	letterIdxMask uint64 = 1<<letterIdxBits - 1 // 6-bit mask
	letterIdxMax  int    = 64 / int(letterIdxBits)
)

// Rand creates a random string of length n.
// From https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go, but uses rand/v2 patterns.
func Rand(n int) string {
	if n <= 0 {
		return ""
	}

	b := make([]byte, n)

	for i, cache, remain := n-1, rand.Uint64(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Uint64(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
