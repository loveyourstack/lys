package lyspg

import (
	"context"
	"fmt"
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

// Update changes a single record with the values contained in input
// T must be a struct with "db" tags
func Update[T any, pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, input T, pkVal pkT) error {

	// get input values by reflecting input T
	meta, err := lysmeta.AnalyzeT(input, true)
	if err != nil {
		return fmt.Errorf("lysmeta.AnalyzeT failed: %w", err)
	}

	dbNames, inputVals, err := meta.DbValues()
	if err != nil {
		return fmt.Errorf("meta.DbValues failed: %w", err)
	}

	stmt := getUpdateStmt(schemaName, tableName, pkColName, dbNames)
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

// UpdateWithLastUserBy is a wrapper for Update that adds a last_user_update_by field to the input struct and sets it to the supplied lastUserUpdateBy value
func UpdateWithLastUserBy[T any, pkT PrimaryKeyType](ctx context.Context, db PoolOrTx, schemaName, tableName, pkColName string, input T, pkVal pkT, lastUserUpdateBy string) error {

	type inputWithLastUserBy struct {
		Input            T
		LastUserUpdateBy string `db:"last_user_update_by"`
	}
	var inputLub inputWithLastUserBy
	inputLub.Input = input
	inputLub.LastUserUpdateBy = lastUserUpdateBy
	return Update(ctx, db, schemaName, tableName, pkColName, inputLub, pkVal)
}
