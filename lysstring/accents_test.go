package lysstring

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceAccentsSuccess(t *testing.T) {

	res, err := ReplaceAccents(letterBytes)
	assert.Equal(t, letterBytes, res, "en alphabet - no replacements")
	assert.Equal(t, nil, err, "en alphabet - err")

	res, err = ReplaceAccents(deAccents)
	assert.Equal(t, "aeoeuess", res, "de accents replaced")
	assert.Equal(t, nil, err, "de accents - err")

	res, err = ReplaceAccents(frAccents)
	assert.Equal(t, "eaeucaeioueiue", res, "fr accents replaced")
	assert.Equal(t, nil, err, "fr accents - err")
}
