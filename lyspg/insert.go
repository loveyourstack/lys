package lyspg

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/loveyourstack/lys/lysmeta"
)

// getInsertStmt returns an INSERT statement using the supplied params
func getInsertStmt(schemaName, tableName, pkColName string, inputFields []string) string {

	var paramPlaceholders []string
	for k := range inputFields {
		paramPlaceholders = append(paramPlaceholders, "$"+strconv.Itoa(k+1))
	}

	return fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s) RETURNING %s;",
		schemaName, tableName, strings.Join(inputFields, ", "), strings.Join(paramPlaceholders, ", "), pkColName)
}

// Insert inserts a single record and then returns it
// inputT must be a struct with "db" tags
func Insert[inputT any, itemT any](ctx context.Context, db PoolOrTx, schemaName, tableName, viewName, pkColName string, allFields []string,
	input inputT) (newItem itemT, stmt string, err error) {

	// get input db struct tags
	inputReflVals := reflect.ValueOf(input)
	meta, err := lysmeta.AnalyzeStructs(inputReflVals)
	if err != nil {
		return newItem, "", fmt.Errorf("lysmeta.AnalyzeStructs failed: %w", err)
	}

	if len(meta.DbTags) == 0 {
		return newItem, "", fmt.Errorf("input type does not have db tags")
	}

	// get the input values via reflection
	inputVals := getInputValsFromStruct(inputReflVals, nil)

	stmt = getInsertStmt(schemaName, tableName, pkColName, meta.DbTags)

	var newPk any
	if err = db.QueryRow(ctx, stmt, inputVals...).Scan(&newPk); err != nil {
		return newItem, stmt, fmt.Errorf(ErrDescInsertScanFailed+": %w", err)
	}

	return SelectUnique[itemT](ctx, db, schemaName, viewName, pkColName, nil, allFields, newPk)
}
