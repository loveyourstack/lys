package lyspg

import (
	"context"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
)

// SelectEnum returns all values of the supplied enum type
// includeVals = enum values to include
// excludeVals = enum values to exclude
// sortVal = either "" (no sort), "val" or "-val"
func SelectEnum(ctx context.Context, db PoolOrTx, enumName string, includeVals, excludeVals []string, sortVal string) (vals []string, stmt string, err error) {

	// get selection stmt, casting enum to text if sorting is needed
	castStr := ""
	if sortVal != "" {
		castStr = "::text"
	}
	stmt = fmt.Sprintf("WITH vals AS (SELECT unnest(enum_range(NULL::%s)) val) SELECT val%s FROM vals WHERE 1=1", enumName, castStr)

	inputVals := []any{}
	paramNum := 0

	// process filters
	if len(includeVals) > 0 {
		paramNum++
		stmt += fmt.Sprintf(" AND val = ANY($%v)", paramNum)
		inputVals = append(inputVals, includeVals)
	}

	if len(excludeVals) > 0 {
		paramNum++
		stmt += fmt.Sprintf(" AND NOT (val = ANY($%v))", paramNum)
		inputVals = append(inputVals, excludeVals)
	}

	// process sorts
	switch sortVal {
	case "val":
		stmt += " ORDER BY val"
	case "-val":
		stmt += " ORDER BY val DESC"
	}

	rows, _ := db.Query(ctx, stmt, inputVals...)
	vals, err = pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, stmt, fmt.Errorf("pgx.CollectRows failed: %w", err)
	}

	// success
	return vals, "", nil
}

// CheckEnumValue returns an error if the suppied testVal is not found in the supplied pg enum type
func CheckEnumValue(ctx context.Context, db PoolOrTx, dbEnum, testVal, enumName string) (err error) {

	vals, _, err := SelectEnum(ctx, db, dbEnum, nil, nil, "")
	if err != nil {
		return fmt.Errorf("SelectEnum failed: %w", err)
	}

	if !slices.Contains(vals, testVal) {
		return fmt.Errorf("value %s not found in enum %s", testVal, enumName)
	}

	// success
	return nil
}
