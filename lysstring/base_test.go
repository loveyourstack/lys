package lysstring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAsciiSuccess(t *testing.T) {

	res := IsAscii(letterBytes)
	assert.Equal(t, true, res, "en alphabet")

	res = IsAscii(numbers)
	assert.Equal(t, true, res, "numbers")

	res = IsAscii(urlChars)
	assert.Equal(t, true, res, "url chars")
}

func TestIsAsciiFailure(t *testing.T) {

	res := IsAscii(deAccents)
	assert.Equal(t, false, res, "de accents")

	res = IsAscii(frAccents)
	assert.Equal(t, false, res, "fr accents")
}

func TestReplaceAccentsSuccess(t *testing.T) {

	res, err := ReplaceAccents(letterBytes)
	assert.Equal(t, letterBytes, res, "en alphabet")
	assert.Equal(t, nil, err, "en alphabet")

	res, err = ReplaceAccents(deAccents)
	assert.Equal(t, "aouß", res, "de accents")
	assert.Equal(t, nil, err, "de accents")

	res, err = ReplaceAccents(frAccents)
	assert.Equal(t, "eaeucaeioueiu", res, "fr accents")
	assert.Equal(t, nil, err, "fr accents")
}
