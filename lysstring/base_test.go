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

	res = Convert(s, ",", ",", nil)
	assert.Equal(t, "A,B", res, "CSV to CSV no change")

	res = Convert(s, "", "", nil)
	assert.Equal(t, "A,B", res, "empty separators no change")

	res = Convert(s, "", " ", nil)
	assert.Equal(t, "A , B", res, "empty inSep, space outSep")
}

func TestDeAliasSuccess(t *testing.T) {
	type MyString string
	s := []MyString{"a", "b", "世界"}
	res := DeAlias(s)
	assert.Equal(t, []string{"a", "b", "世界"}, res)
}

func TestIsAsciiSuccess(t *testing.T) {

	res := IsAscii(letterBytes)
	assert.Equal(t, true, res, "en alphabet")

	numbers := "0123456789"
	res = IsAscii(numbers)
	assert.Equal(t, true, res, "numbers")

	urlChars := "-._~:/?#[]@!$&'()*+,;%="
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
	s := "-a,,b-c+"
	res := RemoveCharacters(s, ",-+")
	assert.Equal(t, "abc", res, "basic")

	resDupChars := RemoveCharacters(s, ",-++")
	assert.Equal(t, "abc", resDupChars, "duplicate chars")

	unicodeChars := RemoveCharacters("a世界b", "界")
	assert.Equal(t, "a世b", unicodeChars, "unicode")

	emptyInput := RemoveCharacters("", ",-+")
	assert.Equal(t, "", emptyInput, "empty input")

	emptyChars := RemoveCharacters(s, "")
	assert.Equal(t, s, emptyChars, "empty chars")
}

func TestSingleLine(t *testing.T) {
	type comboS struct {
		Name     string
		Input    string
		Expected string
	}

	combos := []comboS{
		{
			Name: "tabs and multiple blank lines",
			Input: `
	a
	
	
	b
	
	`,
			Expected: "a\nb",
		},
		{Name: "empty string", Input: "", Expected: ""},
		{Name: "all whitespace", Input: "\t\t\n\n  \r\n", Expected: ""},
		{Name: "no line breaks", Input: "hello world", Expected: "hello world"},
		{Name: "single newline", Input: "a\nb", Expected: "a\nb"},
		{Name: "Windows CRLF", Input: "a\r\nb\r\nc", Expected: "a\nb\nc"},
		{Name: "leading and trailing newlines trimmed", Input: "\n\na\nb\n\n", Expected: "a\nb"},
		{Name: "mixed tabs and newlines", Input: "a\t\tb\n\tc", Expected: "a\nb\nc"},
		{Name: "Unicode content preserved", Input: "\n世界\n\n宇宙\n", Expected: "世界\n宇宙"},
	}

	for _, c := range combos {
		res := SingleLine(c.Input)
		assert.Equal(t, c.Expected, res, c.Name)
	}
}

func TestSingleSpace(t *testing.T) {
	type comboS struct {
		Name     string
		Input    string
		Expected string
	}

	combos := []comboS{
		{Name: "tabs and spaces", Input: "		a	   b 		 ", Expected: "a b"},
		{Name: "empty string", Input: "", Expected: ""},
		{Name: "all whitespace", Input: "   \t  \t  ", Expected: ""},
		{Name: "no extra spaces", Input: "a b c", Expected: "a b c"},
		{Name: "leading and trailing spaces", Input: "  a b  ", Expected: "a b"},
		{Name: "single word", Input: "  hello  ", Expected: "hello"},
		{Name: "newlines collapsed to space", Input: "a\nb\nc", Expected: "a b c"},
		{Name: "Unicode content preserved", Input: "  世界  宇宙  ", Expected: "世界 宇宙"},
	}

	for _, c := range combos {
		res := SingleSpace(c.Input)
		assert.Equal(t, c.Expected, res, c.Name)
	}
}

func TestStripPunct(t *testing.T) {
	s := "a, b! c? d. e-"
	res := StripPunct(s)
	assert.Equal(t, "a b c d e", res)
}
