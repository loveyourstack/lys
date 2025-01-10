package lyspg

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
)

// TotalCount contains the total number of table records. If IsEstimated is true, the Value was estimated using db statistics rather than calculated using a record count
type TotalCount struct {
	Value       int64 `db:"rowcount"`
	IsEstimated bool  `db:"estimated"`
}

// fastRowCount returns a fast rowcount of the specified db table
// the result may be exact or inexact, depending on the size of the table and whether or not any query conditions are present
// query must be select stmt with placeholders, without order by, limit, or offset
func fastRowCount(ctx context.Context, db PoolOrTx, schemaName, tableName string, conds []Condition, orCondSets [][]Condition, query string) (totalCount TotalCount, err error) {

	var threshold int64 = 10000

	// get rowcount from info schema stats
	statsRowCount, err := GetStatsRowCount(ctx, db, schemaName, tableName)
	if err != nil {
		return TotalCount{}, fmt.Errorf("GetStatsRowCount failed: %w", err)
	}

	// if table is relatively small (rowcount below threshold), get the exact count
	if statsRowCount < threshold {

		// if no conds, straight rowcount
		if len(conds) == 0 && len(orCondSets) == 0 {
			rowCount, err := GetRowCount(ctx, db, schemaName, tableName)
			if err != nil {
				return TotalCount{}, fmt.Errorf("GetRowCount failed: %w", err)
			}
			return TotalCount{Value: rowCount, IsEstimated: false}, nil
		}

		// has conds: get count from unpaged query result
		paramValues := GetSelectParamValues(conds, orCondSets, false, 0, 0)
		rowCount, err := GetRowCountPlaceholderQry(ctx, db, query, paramValues)
		if err != nil {
			return TotalCount{}, fmt.Errorf("GetRowCountPlaceholderQry failed: %w", err)
		}
		return TotalCount{Value: rowCount, IsEstimated: false}, nil
	}

	// large table (rowcount above threshold)

	// if no conditions, just return est rowcount from stats
	if len(conds) == 0 && len(orCondSets) == 0 {
		return TotalCount{Value: statsRowCount, IsEstimated: true}, nil
	}

	// has conditions: get est rowcount from query plan
	// from https://www.cybertec-postgresql.com/en/postgresql-count-made-fast/
	paramValues := GetSelectParamValues(conds, orCondSets, false, 0, 0)
	rowCount, err := GetRowCountExplain(ctx, db, query, paramValues)
	if err != nil {
		return TotalCount{}, fmt.Errorf("GetRowCountExplain failed: %w", err)
	}

	return TotalCount{Value: rowCount, IsEstimated: true}, nil
}

// GetRowCount returns a straight rowcount of the supplied table
func GetRowCount(ctx context.Context, db PoolOrTx, schemaName, tableName string) (rowCount int64, err error) {

	stmt := fmt.Sprintf("SELECT count(*) FROM %s.%s;", schemaName, tableName)
	rows, _ := db.Query(ctx, stmt)
	rowCount, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[int64])
	if err != nil {
		return 0, lyserr.Db{Err: fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err), Stmt: stmt}
	}

	return rowCount, nil
}

// GetRowCountPlaceholderQry returns the exact rowcount of the supplied qry
// qry must use placeholders for params, and paramValues must be supplied
func GetRowCountPlaceholderQry(ctx context.Context, db PoolOrTx, qry string, paramValues []any) (rowCount int64, err error) {

	stmt := fmt.Sprintf("SELECT count(*) FROM (%s) res;", qry)
	rows, _ := db.Query(ctx, stmt, paramValues...)
	rowCount, err = pgx.CollectExactlyOneRow(rows, pgx.RowTo[int64])
	if err != nil {
		return 0, lyserr.Db{Err: fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err), Stmt: stmt}
	}

	return rowCount, nil
}

type explainResp []struct {
	Plan struct {
		NodeType      string  `json:"Node Type"`
		ParallelAware bool    `json:"Parallel Aware"`
		AsyncCapable  bool    `json:"Async Capable"`
		RelationName  string  `json:"Relation Name"`
		Alias         string  `json:"Alias"`
		StartupCost   float64 `json:"Startup Cost"`
		TotalCost     float64 `json:"Total Cost"`
		PlanRows      int64   `json:"Plan Rows"`
		PlanWidth     int64   `json:"Plan Width"`
	} `json:"Plan"`
}

// GetRowCountExplain returns the estimated rowcount of the supplied qry using the query planner EXPLAIN output
// qry must use placeholders for params, and paramValues must be supplied
func GetRowCountExplain(ctx context.Context, db PoolOrTx, qry string, paramValues []any) (rowCount int64, err error) {

	// get plan in json format
	stmt := fmt.Sprintf("EXPLAIN (FORMAT JSON) %s;", qry)
	rows, _ := db.Query(ctx, stmt, paramValues...)
	plan, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[string])
	if err != nil {
		return 0, lyserr.Db{Err: fmt.Errorf("pgx.CollectExactlyOneRow failed: %w", err), Stmt: stmt}
	}

	// unmarshal into ExplainResp
	expl := explainResp{}
	err = json.Unmarshal([]byte(plan), &expl)
	if err != nil {
		return 0, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	rowCount = expl[0].Plan.PlanRows

	return rowCount, nil
}
