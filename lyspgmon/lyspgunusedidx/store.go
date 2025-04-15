package lyspgunusedidx

import (
	"context"
	"log"
	"reflect"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Pg unused indexes"
	schemaName     string = "lyspgmon"
	tableName      string = "v_unused_indexes"
	viewName       string = "v_unused_indexes"
	defaultOrderBy string = "index_size DESC"
)

type Model struct {
	IndexName       string           `db:"index_name" json:"index_name"`
	IndexScans      int64            `db:"index_scans" json:"index_scans"`
	IndexSize       int64            `db:"index_size" json:"index_size"`
	IndexSizePretty string           `db:"index_size_pretty" json:"index_size_pretty"`
	LastIdxScan     lystype.Datetime `db:"last_idx_scan" json:"last_idx_scan"`
	TableName       string           `db:"table_name" json:"table_name"`
	TableSchema     string           `db:"table_schema" json:"table_schema"`
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
