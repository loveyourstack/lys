package lysstring

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	deAccents string = "äöüß"
	frAccents string = "éàèùçâêîôûëïü"
)

// ReplaceAccents replaces accent characters such as "é" with their non-accented equivalents such as "e"
// from https://twin.sh/articles/33/remove-accents-from-characters-in-go
// contains special handling for German accents
func ReplaceAccents(s string) (res string, err error) {

	var deReplacer = strings.NewReplacer(
		"ä", "ae",
		"Ä", "AE",
		"ö", "oe",
		"Ö", "OE",
		"ü", "ue", // also used in frAccents. Maybe pass lang iso as optional param?
		"Ü", "UE",
		"ß", "ss",
	)
	res = deReplacer.Replace(s)

	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	res, _, err = transform.String(t, res)
	return res, err
}
