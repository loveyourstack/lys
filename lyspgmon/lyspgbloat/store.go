package lyspgbloat

import (
	"context"
	"log"
	"reflect"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	name           string = "Pg table and index bloat"
	schemaName     string = "lyspgmon"
	tableName      string = "v_bloat"
	viewName       string = "v_bloat"
	defaultOrderBy string = "table_waste DESC"
)

type Model struct {
	IndexBloat       float32 `db:"index_bloat" json:"index_bloat"`
	IndexName        string  `db:"index_name" json:"index_name"`
	IndexWaste       int64   `db:"index_waste" json:"index_waste"`
	IndexWastePretty string  `db:"index_waste_pretty" json:"index_waste_pretty"`
	TableBloat       float32 `db:"table_bloat" json:"table_bloat"`
	TableName        string  `db:"table_name" json:"table_name"`
	TableSchema      string  `db:"table_schema" json:"table_schema"`
	TableWaste       int64   `db:"table_waste" json:"table_waste"`
	TableWastePretty string  `db:"table_waste_pretty" json:"table_waste_pretty"`
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
