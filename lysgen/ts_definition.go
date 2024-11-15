package lysgen

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lysstring"
)

// TsDefinition generates the TS type definition from the supplied db table
// only handles 1 level of join
func TsDefinition(ctx context.Context, db *pgxpool.Pool, schema, table string) (res string, err error) {

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

	// convert snake case to pascal case
	tsTableName := lysstring.Convert(table, "_", "", lysstring.Title)

	// build result
	resA := []string{}

	// input
	inputResA, err := getTsInput(tsTableName, cols)
	if err != nil {
		return "", fmt.Errorf("getTsInput failed: %w", err)
	}
	resA = append(resA, inputResA...)

	// def
	modelResA, err := getTsDefinition(tsTableName, cols, parentCols, childFks)
	if err != nil {
		return "", fmt.Errorf("getTsDefinition failed: %w", err)
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

func getTsInput(tsTableName string, cols []lyspg.Column) (resA []string, err error) {

	resA = append(resA, fmt.Sprintf("export interface %sInput {", tsTableName))

	var colVals []string

	// for each column in main table
	for _, col := range cols {

		// skip db-assigned cols
		if col.IsIdentity || col.IsGenerated || col.IsTracking {
			continue
		}

		// get Ts data type
		tsDataType, err := GetTsDataTypeFromPg(col.DataType)
		if err != nil {
			return nil, fmt.Errorf("GetTsDataTypeFromPg failed: %w", err)
		}

		// add line for column
		colVal := fmt.Sprintf("  %s: %s", col.Name, tsDataType)

		colVals = append(colVals, colVal)
	}

	resA = append(resA, colVals...)
	resA = append(resA, "}")

	return resA, nil
}

func getTsDefinition(tsTableName string, cols []lyspg.Column, parentCols []lyspg.Column, childFks []lyspg.ForeignKey) (resA []string, err error) {

	resA = append(resA, fmt.Sprintf("export interface %s extends %sInput {", tsTableName, tsTableName))

	colVals := []string{}

	// for each column in main table
	for _, col := range cols {

		// skip user inputted cols
		if !(col.IsIdentity || col.IsGenerated || col.IsTracking) {
			continue
		}

		// get Ts data type
		tsDataType, err := GetTsDataTypeFromPg(col.DataType)
		if err != nil {
			return nil, fmt.Errorf("GetTsDataTypeFromPg failed: %w", err)
		}

		// add line for column
		colVal := fmt.Sprintf("  %s: %s", col.Name, tsDataType)
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

		// get Ts data type
		tsDataType, err := GetTsDataTypeFromPg(parCol.DataType)
		if err != nil {
			return nil, fmt.Errorf("GetTsDataTypeFromPg failed: %w", err)
		}

		// add line for column
		colVal := fmt.Sprintf("  %s: %s", prefixedColName, tsDataType)
		colVals = append(colVals, colVal)
	}

	// for each child FK
	for _, fk := range childFks {

		// add count line
		colVal := fmt.Sprintf("  %s_count: number", fk.ChildTable)
		colVals = append(colVals, colVal)
	}

	resA = append(resA, colVals...)
	resA = append(resA, "}")

	return resA, nil
}
