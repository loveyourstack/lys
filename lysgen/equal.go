package lysgen

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lysstring"
)

// Equal generates the Go store Equal function from the supplied db table
func Equal(ctx context.Context, db *pgxpool.Pool, schema, table string) (res string, err error) {

	// get table columns
	cols, err := lyspg.GetTableColumns(ctx, db, schema, table)
	if err != nil {
		return "", fmt.Errorf("lyspg.GetTableColumns failed: %w", err)
	}

	// build result
	resA := []string{}

	inputResA, err := getEqual(cols)
	if err != nil {
		return "", fmt.Errorf("getEqual failed: %w", err)
	}
	resA = append(resA, inputResA...)
	res = strings.Join(resA, "\n")

	// write to clipboard for convenience
	err = WriteToClipboard(res)
	if err != nil {
		return "", fmt.Errorf("WriteToClipboard failed: %w", err)
	}

	return "\n" + res + "\n", nil
}

func getEqual(cols []lyspg.Column) (resA []string, err error) {

	resA = append(resA, "func (s Store) Equal(a, b Model) bool {")

	var colVals []string

	// for each column in main table
	line := 0
	for _, col := range cols {

		// skip db-assigned cols
		if col.IsIdentity || col.IsGenerated || col.IsTracking {
			continue
		}

		line++

		// convert snake case to pascal case
		goName := lysstring.Convert(col.Name, "_", "", lysstring.Title)
		colVal := "    "

		if line == 1 {
			colVal += "return "
		} else {
			colVal += "    "
		}

		switch col.DataType {
		case "ARRAY":
			colVal += fmt.Sprintf("lysslices.EqualUnordered(a.%s, b.%s)", goName, goName)
		case "bigint", "bigserial", "integer", "serial", "smallint", "smallserial":
			colVal += fmt.Sprintf("a.%s == b.%s", goName, goName)
		case "bit", "boolean":
			colVal += fmt.Sprintf("a.%s == b.%s", goName, goName)
		case "character", "character varying", "text", "USER-DEFINED": // "USER-DEFINED" is enum
			colVal += fmt.Sprintf("a.%s == b.%s", goName, goName)
		case "date":
			colVal += fmt.Sprintf("a.%s.Format(lystype.DateFormat) == b.%s.Format(lystype.DateFormat)", goName, goName)
		case "double precision", "money", "numeric", "real":
			colVal += fmt.Sprintf("fmt.Sprintf(\"%%.4f\", a.%s) == fmt.Sprintf(\"%%.4f\", b.%s)", goName, goName)
		case "time":
			colVal += fmt.Sprintf("a.%s.Format(lystype.TimeFormat) == b.%s.Format(lystype.TimeFormat)", goName, goName)
		case "timestamp", "timestamp with time zone":
			colVal += fmt.Sprintf("a.%s.Format(lystype.DatetimeFormat) == b.%s.Format(lystype.DatetimeFormat)", goName, goName)
		default:
			return nil, fmt.Errorf("no rule type found for DataType: %s", col.DataType)
		}
		colVal += " &&"

		colVals = append(colVals, colVal)
	}

	// remove && in final line
	if len(colVals) > 0 {
		colVals[len(colVals)-1] = strings.Replace(colVals[len(colVals)-1], " &&", "", 1)
	}

	resA = append(resA, colVals...)
	resA = append(resA, "}")

	return resA, nil
}
