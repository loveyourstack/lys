package lysgen

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lysstring"
)

// TsTypes generates the TS type definitions from the supplied db table
// only handles 1 level of join
func TsTypes(ctx context.Context, db *pgxpool.Pool, schema, table string) (res string, err error) {

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

	// input interface
	inputResA, err := getTsInputIFace(tsTableName, cols)
	if err != nil {
		return "", fmt.Errorf("getTsInputIFace failed: %w", err)
	}
	resA = append(resA, inputResA...)

	// interface including input
	modelResA, err := getTsIFace(tsTableName, cols, parentCols, childFks)
	if err != nil {
		return "", fmt.Errorf("getTsIFace failed: %w", err)
	}
	resA = append(resA, modelResA...)

	// constructor
	constructorResA, err := getTsConstructor(tsTableName, cols, parentCols, childFks)
	if err != nil {
		return "", fmt.Errorf("getTsConstructor failed: %w", err)
	}
	resA = append(resA, constructorResA...)

	// input from item func
	inputFromItemResA, err := getTsInputFromItemFunc(tsTableName, cols)
	if err != nil {
		return "", fmt.Errorf("getTsInputFromItemFunc failed: %w", err)
	}
	resA = append(resA, inputFromItemResA...)

	res = strings.Join(resA, "\n")

	// write to clipboard for convenience
	err = WriteToClipboard(res)
	if err != nil {
		return "", fmt.Errorf("WriteToClipboard failed: %w", err)
	}

	return "\n" + res + "\n", nil
}

func getTsInputIFace(tsTableName string, cols []lyspg.Column) (resA []string, err error) {

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
		colVal := ""
		switch col.DataType {
		case "boolean":
			colVal = fmt.Sprintf("  %s: %s", col.Name, tsDataType)
		default:
			colVal = fmt.Sprintf("  %s: %s | undefined", col.Name, tsDataType)
		}

		colVals = append(colVals, colVal)
	}

	resA = append(resA, colVals...)
	resA = append(resA, "}")

	return resA, nil
}

func getTsIFace(tsTableName string, cols []lyspg.Column, parentCols []lyspg.Column, childFks []lyspg.ForeignKey) (resA []string, err error) {

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

func getTsConstructor(tsTableName string, cols []lyspg.Column, parentCols []lyspg.Column, childFks []lyspg.ForeignKey) (resA []string, err error) {

	resA = append(resA, fmt.Sprintf("export function New%s(): %s {", tsTableName, tsTableName))
	resA = append(resA, "  return {")

	var inputCols, generatedCols []lyspg.Column
	var colVals []string

	for _, col := range cols {
		if col.IsIdentity || col.IsGenerated || col.IsTracking {
			generatedCols = append(generatedCols, col)
			continue
		}
		inputCols = append(inputCols, col)
	}

	// for each input column
	for _, col := range inputCols {

		// add line for column
		colVal := ""
		switch col.DataType {
		case "boolean":
			colVal = fmt.Sprintf("    %s: false,", col.Name)
		default:
			colVal = fmt.Sprintf("    %s: undefined,", col.Name)
		}

		colVals = append(colVals, colVal)
	}

	colVals = append(colVals, "")

	// for each generated column
	for _, col := range generatedCols {

		// get Ts data type
		tsDataType, err := GetTsDataTypeFromPg(col.DataType)
		if err != nil {
			return nil, fmt.Errorf("GetTsDataTypeFromPg failed: %w", err)
		}

		// add line for column
		colVal := fmt.Sprintf("    %s: %s,", col.Name, getTsInitialValue(tsDataType))
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
		colVal := fmt.Sprintf("    %s: %s,", prefixedColName, getTsInitialValue(tsDataType))
		colVals = append(colVals, colVal)
	}

	// for each child FK
	for _, fk := range childFks {

		// add count line
		colVal := fmt.Sprintf("    %s_count: 0,", fk.ChildTable)
		colVals = append(colVals, colVal)
	}

	resA = append(resA, colVals...)

	resA = append(resA, "  }")
	resA = append(resA, "}")

	return resA, nil
}

func getTsInputFromItemFunc(tsTableName string, cols []lyspg.Column) (resA []string, err error) {

	resA = append(resA, fmt.Sprintf("export function Get%sInputFromItem(item: %s): %sInput {", tsTableName, tsTableName, tsTableName))
	resA = append(resA, "  return {")

	var colVals []string

	// for each column in main table
	for _, col := range cols {

		// skip db-assigned cols
		if col.IsIdentity || col.IsGenerated || col.IsTracking {
			continue
		}

		colVals = append(colVals, fmt.Sprintf("    %s: item.%s,", col.Name, col.Name))
	}

	resA = append(resA, colVals...)

	resA = append(resA, "  }")
	resA = append(resA, "}")

	return resA, nil
}

func getTsInitialValue(tsDataType string) string {

	switch tsDataType {
	case "boolean":
		return "false"
	case "Date":
		return "new Date()"
	case "number":
		return "0"
	case "string":
		return "''"
	default:
		return "undefined"
	}
}
