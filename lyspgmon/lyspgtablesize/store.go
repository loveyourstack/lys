package lyspgtablesize

import (
	"context"
	"log"
	"reflect"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	name           string = "Pg table size"
	schemaName     string = "lyspgmon"
	tableName      string = "v_table_size"
	viewName       string = "v_table_size"
	defaultOrderBy string = "total_bytes DESC"
)

type Model struct {
	IndexBytes     int64   `db:"index_bytes" json:"index_bytes"`
	IndexPretty    string  `db:"index_pretty" json:"index_pretty"`
	RowEstimate    int64   `db:"row_estimate" json:"row_estimate"`
	TableBytes     int64   `db:"table_bytes" json:"table_bytes"`
	TablePretty    string  `db:"table_pretty" json:"table_pretty"`
	TableName      string  `db:"table_name" json:"table_name"`
	TableSchema    string  `db:"table_schema" json:"table_schema"`
	ToastBytes     int64   `db:"toast_bytes" json:"toast_bytes"`
	ToastPretty    string  `db:"toast_pretty" json:"toast_pretty"`
	TotalBytes     int64   `db:"total_bytes" json:"total_bytes"`
	TotalPretty    string  `db:"total_pretty" json:"total_pretty"`
	TotalSizeShare float32 `db:"total_size_share" json:"total_size_share"`
}

var (
	meta lysmeta.Result
)

func init() {
	var err error
	meta, err = lysmeta.AnalyzeStructs(reflect.ValueOf(&Model{}).Elem())
	if err != nil {
		log.Fatalf("lysmeta.AnalyzeStructs failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
}

type Store struct {
	Db *pgxpool.Pool
}

func (s Store) GetMeta() lysmeta.Result {
	return meta
}
func (s Store) GetName() string {
	return name
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, meta.DbTags, params)
}
