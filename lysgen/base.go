package lysgen

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetGoDataTypeFromPg returns a Go data type from a PostgreSQL data type
func GetGoDataTypeFromPg(pgType string) (goType, omitStr string, err error) {

	switch pgType {
	case "ARRAY":
		return "[]string", "omitempty", nil // defaulting to string, change type manually as needed
	case "bigint", "bigserial":
		return "int64", "omitempty", nil
	case "bit", "boolean":
		return "bool", "omitempty", nil
	case "character", "character varying", "text", "USER-DEFINED": // "USER-DEFINED" is enum or domain. Change as needed
		return "string", "omitempty", nil
	case "date":
		return "lystype.Date", "omitzero", nil
	case "double precision", "money", "numeric", "real":
		return "float64", "omitempty", nil
	case "integer", "serial", "smallint", "smallserial":
		return "int", "omitempty", nil
	case "time", "time without time zone":
		return "lystype.Time", "omitzero", nil
	case "timestamp", "timestamp with time zone":
		return "lystype.Datetime", "omitzero", nil
	default:
		return "", "", fmt.Errorf("no go type found for pgType: %s", pgType)
	}
}

// GetTsDataTypeFromPg returns a Typescript data type from a PostgreSQL data type
func GetTsDataTypeFromPg(pgType string) (tsType string, err error) {

	switch pgType {
	case "ARRAY":
		return "string[]", nil // defaulting to string, change type manually as needed
	case "bigint", "bigserial", "double precision", "integer", "money", "numeric", "real", "serial", "smallint", "smallserial":
		return "number", nil
	case "bit", "boolean":
		return "boolean", nil
	case "character", "character varying", "date", "text", "time", "time without time zone", "USER-DEFINED": // "USER-DEFINED" is enum or domain
		return "string", nil
	case "timestamp", "timestamp with time zone":
		return "Date", nil
	default:
		return "", fmt.Errorf("no Typescript type found for pgType: %s", pgType)
	}
}

// currently only tested on WSL2
func WriteToClipboard(s string) error {
	cmd := exec.Command("clip.exe")
	cmd.Stdin = strings.NewReader(s)
	return cmd.Run()
}
