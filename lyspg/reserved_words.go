package lyspg

func init() {
	ReservedWords["case"] = true
	ReservedWords["cast"] = true
	ReservedWords["check"] = true
	ReservedWords["column"] = true
	ReservedWords["constraint"] = true
	ReservedWords["default"] = true
	ReservedWords["distinct"] = true
	ReservedWords["offset"] = true
	ReservedWords["references"] = true
	ReservedWords["table"] = true
	ReservedWords["user"] = true
}

// pg reserved words: need to be escaped
var ReservedWords = make(map[string]bool)

// EscapeReserved escapes elements of s with double quotes if the element is a PG reserved word
// intended for field lists when selecting from views
func EscapeReserved(s []string) {
	for i := range s {
		if ReservedWords[s[i]] {
			s[i] = `"` + s[i] + `"`
		}
	}
}
