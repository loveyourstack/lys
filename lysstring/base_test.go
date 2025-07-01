package lysstring

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertSuccess(t *testing.T) {
	s := "A,B"
	res := Convert(s, ",", "|", nil)
	assert.Equal(t, "A|B", res, "CSV to TSV")

	res = Convert(s, ",", "|", strings.ToLower)
	assert.Equal(t, "a|b", res, "CSV to TSV, ToLower")
}

func TestDeAliasSuccess(t *testing.T) {
	type MyString string
	s := []MyString{"a", "b"}
	res := DeAlias(s)
	assert.Equal(t, []string{"a", "b"}, res)
}

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

func TestRemoveCharactersSuccess(t *testing.T) {
	s := "a,b-c+"
	res := RemoveCharacters(s, ",-+")
	assert.Equal(t, "abc", res)
}

func TestStdLinesSuccess(t *testing.T) {
	s := `
	a
	
	
	b
	
	`
	res := StdLines(s)
	assert.Equal(t, "a\nb", res)
}

func TestStdWhitespaceSuccess(t *testing.T) {
	s := "		a   b 		 "
	res := StdWhitespace(s)
	assert.Equal(t, "a b", res)
}
