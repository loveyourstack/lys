package lyspg

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lysmeta"
)

// getUpdateStmt returns an UPDATE statement using the supplied params
func getUpdateStmt(schemaName, tableName, pkColName string, inputFields []string) string {

	if len(inputFields) == 0 {
		return ""
	}

	var assignments []string
	for k, field := range inputFields {
		assignment := field + " = $" + strconv.Itoa(k+1)
		assignments = append(assignments, assignment)
	}

	return fmt.Sprintf("UPDATE %s.%s SET %s WHERE %s = $%d;",
		schemaName, tableName, strings.Join(assignments, ", "), pkColName, len(inputFields)+1)
}

type UpdateOption struct {
	OmitFields []string // db columns to exclude from update
}

// Update changes a single record with the values contained in input
// T must be a struct with "db" tags
// pkVal is type "any" so that both int and string PKs can be used
func Update[T any](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, input T, pkVal any, options ...UpdateOption) (stmt string, err error) {

	var updateFields, omitFields []string

	for _, option := range options {
		if len(option.OmitFields) == 0 {
			continue
		}
		omitFields = append(omitFields, option.OmitFields...)
	}

	// get columns to update by reflecting input T
	inputReflVals := reflect.ValueOf(input)
	dbTags, _, err := lysmeta.GetStructTags(inputReflVals)
	if err != nil {
		return "", fmt.Errorf("lysmeta.GetStructTags failed: %w", err)
	}

	// updateFields is dbTags with omitted fields removed
	for _, dbTag := range dbTags {
		found := false
		for _, omitField := range omitFields {
			if dbTag == omitField {
				found = true
				break
			}
		}
		if !found {
			updateFields = append(updateFields, dbTag)
		}
	}

	// get input values by reflecting input T
	inputVals := getInputValsFromStruct(inputReflVals, omitFields)

	stmt = getUpdateStmt(schemaName, tableName, pkColName, updateFields)
	inputVals = append(inputVals, pkVal)

	cmdTag, err := db.Exec(ctx, stmt, inputVals...)
	if err != nil {
		return stmt, fmt.Errorf(ErrDescUpdateExecFailed+": %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return stmt, pgx.ErrNoRows
	}

	// success
	return "", nil
}
