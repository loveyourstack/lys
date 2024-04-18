package lyspgdb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lystype"
)

type StatActivity struct {
	ApplicationName string           `db:"application_name" json:"application_name"`
	ClientAddr      string           `db:"client_addr" json:"client_addr"`
	Pid             int              `db:"pid" json:"pid"`
	Query           string           `db:"query" json:"query"`
	QueryStart      lystype.Datetime `db:"query_start" json:"query_start"`
	State           string           `db:"state" json:"state"`
	UseName         string           `db:"usename" json:"usename"`
}

// https://www.postgresql.org/docs/current/monitoring-stats.html#MONITORING-PG-STAT-ACTIVITY-VIEW
func GetStatActivity(ctx context.Context, db *pgxpool.Pool, dbName string) (items []StatActivity, stmt string, err error) {

	fields := []string{
		"application_name",
		"client_addr::text",
		"pid",
		"query",
		"query_start",
		"state",
		"usename",
	}

	stmt = "SELECT " + strings.Join(fields, ",") + " FROM pg_stat_activity WHERE state IS NOT NULL AND query NOT LIKE '%pg_stat_activity%' AND datname = $1 ORDER BY query_start DESC;"
	rows, _ := db.Query(ctx, stmt, dbName)
	items, err = pgx.CollectRows(rows, pgx.RowToStructByNameLax[StatActivity])
	if err != nil {
		return nil, stmt, fmt.Errorf("pgx.CollectRows failed: %w", err)
	}

	return items, "", nil
}
