package lysgen

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lysstring"
)

// InputModel generates the store Input and Model structs from the supplied db table
// only handles 1 level of join, and does not coalesce nulls
func InputModel(ctx context.Context, db *pgxpool.Pool, schema, table string, withValidation bool) (res string, err error) {

	// get table columns
	cols, err := lyspg.GetTableColumns(ctx, db, schema, table)
	if err != nil {
		return "", fmt.Errorf("lyspg.GetTableColumns failed: %w", err)
	}

	// get table parent FKs
	parentFks, err := lyspg.GetForeignKeys(ctx, db, schema, table)
	if err != nil {
		return "", fmt.Errorf("lyspg.GetForeignKeys failed: %w", err)
	}

	parentCols := []lyspg.Column{}

	// for each parent FK
	for _, fk := range parentFks {

		// get parent cols
		fkCols, err := lyspg.GetTableColumns(ctx, db, fk.ParentSchema, fk.ParentTable)
		if err != nil {
			return "", fmt.Errorf("lyspg.GetTableColumns failed for table: %s.%s: %w", fk.ParentSchema, fk.ParentTable, err)
		}
		parentCols = append(parentCols, fkCols...)
	}

	// get table child FKs
	childFks, err := lyspg.GetChildForeignKeys(ctx, db, schema, table)
	if err != nil {
		return "", fmt.Errorf("lyspg.GetChildForeignKeys failed: %w", err)
	}

	// build result
	resA := []string{}

	// input
	inputResA, err := getInput(cols, withValidation)
	if err != nil {
		return "", fmt.Errorf("getInput failed: %w", err)
	}
	resA = append(resA, inputResA...)

	resA = append(resA, "")

	// model
	modelResA, err := getModel(cols, parentCols, childFks)
	if err != nil {
		return "", fmt.Errorf("getModel failed: %w", err)
	}
	resA = append(resA, modelResA...)

	res = strings.Join(resA, "\n")

	// write to clipboard for convenience
	err = WriteToClipboard(res)
	if err != nil {
		return "", fmt.Errorf("WriteToClipboard failed: %w", err)
	}

	return "\n" + res + "\n", nil
}

func getInput(cols []lyspg.Column, withValidation bool) (resA []string, err error) {

	resA = append(resA, "type Input struct {")

	var colVals []string

	// for each column in main table
	for _, col := range cols {

		// skip db-assigned cols
		if col.IsIdentity || col.IsGenerated || col.IsTracking {
			continue
		}

		// convert snake case to pascal case
		goName := lysstring.Convert(col.Name, "_", "", lysstring.Title)

		// get Go data type
		goDataType, err := GetGoDataTypeFromPg(col.DataType)
		if err != nil {
			return nil, fmt.Errorf("GetGoDataTypeFromPg failed: %w", err)
		}

		// add line for column
		colVal := fmt.Sprintf("    %s  %s  `db:\"%s\" json:\"%s\"", goName, goDataType, col.Name, col.Name)
		if withValidation {
			colVal += " validate:\"required\""
		}
		colVal += "`"

		colVals = append(colVals, colVal)
	}

	resA = append(resA, colVals...)
	resA = append(resA, "}")

	return resA, nil
}

func getModel(cols []lyspg.Column, parentCols []lyspg.Column, childFks []lyspg.ForeignKey) (resA []string, err error) {

	resA = append(resA, "type Model struct {")

	colVals := []string{}

	// for each column in main table
	for _, col := range cols {

		// skip user inputted cols
		if !(col.IsIdentity || col.IsGenerated || col.IsTracking) {
			continue
		}

		// convert snake case to pascal case
		goName := lysstring.Convert(col.Name, "_", "", lysstring.Title)

		// get Go data type
		goDataType, err := GetGoDataTypeFromPg(col.DataType)
		if err != nil {
			return nil, fmt.Errorf("GetGoDataTypeFromPg failed: %w", err)
		}

		// add line for column
		colVal := fmt.Sprintf("    %s  %s  `db:\"%s\" json:\"%s\"`", goName, goDataType, col.Name, col.Name)
		colVals = append(colVals, colVal)
	}

	// for each parent col
	for _, parCol := range parentCols {

		// skip identity and tracking cols
		if parCol.IsIdentity || parCol.IsTracking {
			continue
		}

		// prefix table name to col
		prefixedColName := parCol.TableName + "_" + parCol.Name

		// convert snake case to pascal case
		goName := lysstring.Convert(prefixedColName, "_", "", lysstring.Title)

		// get Go data type
		goDataType, err := GetGoDataTypeFromPg(parCol.DataType)
		if err != nil {
			return nil, fmt.Errorf("GetGoDataTypeFromPg failed: %w", err)
		}

		// add line for column
		colVal := fmt.Sprintf("    %s  %s  `db:\"%s\" json:\"%s\"`", goName, goDataType, prefixedColName, prefixedColName)
		colVals = append(colVals, colVal)
	}

	// for each child FK
	for _, fk := range childFks {

		// convert table name from snake case to pascal case
		goTableName := lysstring.Convert(fk.ChildTable, "_", "", lysstring.Title)

		// add count line
		colVal := fmt.Sprintf("    %sCount  int  `db:\"%s_count\" json:\"%s_count\"`", goTableName, fk.ChildTable, fk.ChildTable)
		colVals = append(colVals, colVal)
	}

	resA = append(resA, colVals...)
	resA = append(resA, "    Input")
	resA = append(resA, "}")

	return resA, nil
}
