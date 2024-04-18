package lysgen

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyspg"
)

// View generates the db view from the supplied db table
// only handles 1 level of join, and does not coalesce nulls
func View(ctx context.Context, db *pgxpool.Pool, schema, table string) (res string, stmt string, err error) {

	alias := table

	// if table shortname is set via comment, use it as alias
	shortName, stmt, err := lyspg.GetTableShortName(ctx, db, schema, table)
	if err != nil {
		return "", stmt, fmt.Errorf("lyspg.GetTableShortName failed for table: %s.%s: %w", schema, table, err)
	}
	if shortName != "" {
		alias = shortName
	}

	// get table columns
	cols, stmt, err := lyspg.GetTableColumns(ctx, db, schema, table)
	if err != nil {
		return "", stmt, fmt.Errorf("lyspg.GetTableColumns failed: %w", err)
	}

	// get table parent FKs
	parentFks, stmt, err := lyspg.GetForeignKeys(ctx, db, schema, table)
	if err != nil {
		return "", stmt, fmt.Errorf("lyspg.GetForeignKeys failed: %w", err)
	}

	// get table child FKs
	childFks, stmt, err := lyspg.GetChildForeignKeys(ctx, db, schema, table)
	if err != nil {
		return "", stmt, fmt.Errorf("lyspg.GetChildForeignKeys failed: %w", err)
	}

	// build result
	resA := []string{}
	resA = append(resA, fmt.Sprintf("CREATE OR REPLACE VIEW %s.v_%s AS", schema, table))
	resA = append(resA, "  SELECT")

	var colVals, joins []string

	// for each column in main table
	for _, col := range cols {

		// add line for column
		colVal := fmt.Sprintf("    %s.%s", alias, col.Name)
		colVals = append(colVals, colVal)

		// for each FK
		for _, fk := range parentFks {

			if col.Name != fk.ParentColumn {
				continue
			}

			// this col is FK - get its column values and parentJoin
			parentfkColVals, parentJoin, stmt, err := getViewParentFkInfo(ctx, db, fk, alias)
			if err != nil {
				return "", stmt, fmt.Errorf("getViewParentFkInfo failed for parent table: %s.%s: %w", fk.ParentSchema, fk.ParentTable, err)
			}

			colVals = append(colVals, parentfkColVals...)
			joins = append(joins, parentJoin)
		}
	}

	// append child FK info, if any
	for _, fk := range childFks {
		childfkColVal, childJoin, stmt, err := getViewChildFkInfo(ctx, db, fk, alias)
		if err != nil {
			return "", stmt, fmt.Errorf("getViewChildFkInfo failed for child table: %s.%s: %w", fk.ChildSchema, fk.ChildTable, err)
		}
		colVals = append(colVals, childfkColVal)
		joins = append(joins, childJoin)
	}

	resA = append(resA, strings.Join(colVals, ",\n"))
	resA = append(resA, fmt.Sprintf("  FROM %s.%s %s", schema, table, alias))
	if len(joins) > 0 {
		resA = append(resA, joins...)
	}
	res = strings.Join(resA, "\n") + ";"

	// write to clipboard for convenience
	err = WriteToClipboard(res)
	if err != nil {
		return "", "", fmt.Errorf("WriteToClipboard failed: %w", err)
	}

	return "\n" + res + "\n", "", nil
}

func getViewParentFkInfo(ctx context.Context, db *pgxpool.Pool, fk lyspg.ForeignKey, mainAlias string) (colVals []string, join, stmt string, err error) {

	parentAlias := fk.ParentTable

	// get parent table shortname, if exists
	parentShortName, stmt, err := lyspg.GetTableShortName(ctx, db, fk.ParentSchema, fk.ParentTable)
	if err != nil {
		return nil, "", stmt, fmt.Errorf("lyspg.GetTableShortName failed: %w", err)
	}
	if parentShortName != "" {
		parentAlias = parentShortName
	}

	// add a JOIN to the parent table
	join = fmt.Sprintf("  JOIN %s.%s %s ON %s.%s = %s.%s", fk.ParentSchema, fk.ParentTable, parentAlias, mainAlias, fk.ChildColumn, parentAlias, fk.ParentColumn)

	// get parent cols
	parentCols, stmt, err := lyspg.GetTableColumns(ctx, db, fk.ParentSchema, fk.ParentTable)
	if err != nil {
		return nil, "", stmt, fmt.Errorf("lyspg.GetTableColumns failed for table: %s.%s: %w", fk.ParentSchema, fk.ParentTable, err)
	}

	// define parent colVals
	for _, parCol := range parentCols {

		// skip identity and tracking cols
		if parCol.IsIdentity || parCol.IsTracking {
			continue
		}

		colVal := fmt.Sprintf("    %s.%s AS %s_%s", parentAlias, parCol.Name, fk.ParentTable, parCol.Name)
		colVals = append(colVals, colVal)
	}

	return colVals, join, "", nil
}

func getViewChildFkInfo(ctx context.Context, db *pgxpool.Pool, fk lyspg.ForeignKey, mainAlias string) (colVal string, join, stmt string, err error) {

	childAlias := fk.ChildTable

	// get child table shortname, if exists
	childShortName, stmt, err := lyspg.GetTableShortName(ctx, db, fk.ChildSchema, fk.ChildTable)
	if err != nil {
		return "", "", stmt, fmt.Errorf("lyspg.GetTableShortName failed: %w", err)
	}
	if childShortName != "" {
		childAlias = childShortName
	}

	// define count colVal
	colVal = fmt.Sprintf("    COALESCE(%s.%s_count,0) AS %s_count", childAlias, fk.ChildTable, fk.ChildTable)

	// add a LEFT JOIN aggregate to the parent table
	join = fmt.Sprintf("  LEFT JOIN (SELECT %s, count(*) AS %s_count FROM %s.%s GROUP BY 1) %s ON %s.%s = %s.%s",
		fk.ChildColumn, fk.ChildTable, fk.ChildSchema, fk.ChildTable, childAlias, childAlias, fk.ChildColumn, mainAlias, fk.ParentColumn)

	return colVal, join, "", nil
}
