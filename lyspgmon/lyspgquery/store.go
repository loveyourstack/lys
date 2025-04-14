package lyspgquery

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
	name           string = "Pg queries"
	schemaName     string = "lyspgmon"
	tableName      string = "v_queries"
	viewName       string = "v_queries"
	defaultOrderBy string = "query_start"
)

type Model struct {
	ApplicationName string           `db:"application_name" json:"application_name"`
	ClientAddr      string           `db:"client_addr" json:"client_addr"`
	Pid             int              `db:"pid" json:"pid"`
	Query           string           `db:"query" json:"query"`
	QueryStart      lystype.Datetime `db:"query_start" json:"query_start"`
	State           string           `db:"state" json:"state"`
	UseName         string           `db:"usename" json:"usename"`
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
