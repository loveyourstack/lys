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

var accentsDeReplacer = strings.NewReplacer(
	"ä", "ae",
	"Ä", "AE",
	"ö", "oe",
	"Ö", "OE",
	"ü", "ue", // also used in frAccents. Maybe pass lang iso as optional param?
	"Ü", "UE",
	"ß", "ss",
)

var accentsTransChain = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

// ReplaceAccents replaces accent characters such as "é" with their non-accented equivalents such as "e".
// Contains special handling for German accents.
// From https://twin.sh/articles/33/remove-accents-from-characters-in-go.
func ReplaceAccents(s string) (res string, err error) {

	res = accentsDeReplacer.Replace(s)

	res, _, err = transform.String(accentsTransChain, res)
	return res, err
}
