package lyspg

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
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
func Update[T any, pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, input T, pkVal pkT, options ...UpdateOption) error {

	// get columns to update by reflecting input T
	inputReflVals := reflect.ValueOf(input)
	meta, err := lysmeta.AnalyzeStruct(inputReflVals)
	if err != nil {
		return fmt.Errorf("lysmeta.AnalyzeStruct failed: %w", err)
	}

	// get fields to omit from the update, if any
	omitFields := getOmitFields(options...)

	// get updateFields (dbTags with omitted fields removed)
	updateFields := getUpdateFields(meta.DbTags, omitFields)

	// get input values by reflecting input T
	inputVals := getInputValsFromStruct(inputReflVals, omitFields)

	stmt := getUpdateStmt(schemaName, tableName, pkColName, updateFields)
	inputVals = append(inputVals, pkVal)

	cmdTag, err := db.Exec(ctx, stmt, inputVals...)
	if err != nil {
		return lyserr.Db{Err: fmt.Errorf(ErrDescUpdateExecFailed+": %w", err), Stmt: stmt}
	}

	if cmdTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	// success
	return nil
}

func getOmitFields(options ...UpdateOption) (omitFields []string) {

	for _, option := range options {
		if len(option.OmitFields) == 0 {
			continue
		}
		omitFields = append(omitFields, option.OmitFields...)
	}

	return omitFields
}

func getUpdateFields(dbTags, omitFields []string) (updateFields []string) {

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

	return updateFields
}

// UpdateWithLastUserBy is a wrapper for Update that adds a last_user_update_by field to the input struct and sets it to the supplied lastUserUpdateBy value
func UpdateWithLastUserBy[T any, pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, input T, pkVal pkT,
	lastUserUpdateBy string, options ...UpdateOption) error {

	type inputWithLastUserBy struct {
		Input            T
		LastUserUpdateBy string `db:"last_user_update_by"`
	}
	var inputLub inputWithLastUserBy
	inputLub.Input = input
	inputLub.LastUserUpdateBy = lastUserUpdateBy
	return Update(ctx, db, schemaName, tableName, pkColName, inputLub, pkVal, options...)
}
