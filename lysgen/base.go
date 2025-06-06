package lysgen

import (
	"fmt"
	"os/exec"
)

// GetGoDataTypeFromPg returns a Go data type from a PostgreSQL data type
func GetGoDataTypeFromPg(pgType string) (goType string, err error) {

	switch pgType {
	case "ARRAY":
		return "[]string", nil
	case "bigint", "bigserial":
		return "int64", nil
	case "bit", "boolean":
		return "bool", nil
	case "character", "character varying", "text", "USER-DEFINED": // "USER-DEFINED" is enum
		return "string", nil
	case "date":
		return "lystype.Date", nil
	case "double precision", "real":
		return "float32", nil
	case "integer", "serial", "smallint", "smallserial":
		return "int", nil
	case "money", "numeric":
		return "float64", nil
	case "time", "time without time zone":
		return "lystype.Time", nil
	case "timestamp", "timestamp with time zone":
		return "lystype.Datetime", nil
	default:
		return "", fmt.Errorf("no go type found for pgType: %s", pgType)
	}
}

// GetTsDataTypeFromPg returns a Typescript data type from a PostgreSQL data type
func GetTsDataTypeFromPg(pgType string) (tsType string, err error) {

	switch pgType {
	case "ARRAY":
		return "string[]", nil
	case "bigint", "bigserial", "double precision", "integer", "money", "numeric", "real", "serial", "smallint", "smallserial":
		return "number", nil
	case "bit", "boolean":
		return "boolean", nil
	case "character", "character varying", "text", "USER-DEFINED": // "USER-DEFINED" is enum
		return "string", nil
	case "date", "time", "time without time zone", "timestamp", "timestamp with time zone":
		return "Date", nil
	default:
		return "", fmt.Errorf("no Typescript type found for pgType: %s", pgType)
	}
}

// currently only tested on WSL2
func WriteToClipboard(s string) error {
	cmd := exec.Command("bash", "-c", "echo '"+s+"' | clip.exe")
	return cmd.Run()
}
