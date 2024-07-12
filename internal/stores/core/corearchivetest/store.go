package corearchivetest

import (
	"context"
	"log"
	"reflect"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	name           string = "Archive test"
	schemaName     string = "core"
	tableName      string = "archive_test"
	viewName       string = "archive_test"
	pkColName      string = "id"
	defaultOrderBy string = "id"
)

type Input struct {
	CInt  *int64  `db:"c_int" json:"c_int,omitempty"`
	CText *string `db:"c_text" json:"c_text,omitempty"`
}

type Model struct {
	Id int64 `db:"id" json:"id,omitempty"`
	Input
}

var (
	meta lysmeta.Result
)

func init() {
	var err error
	meta, err = lysmeta.AnalyzeStructs(reflect.ValueOf(&Input{}).Elem(), reflect.ValueOf(&Model{}).Elem())
	if err != nil {
		log.Fatalf("lysmeta.AnalyzeStructs failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
}

type Store struct {
	Db *pgxpool.Pool
}

func (s Store) Archive(ctx context.Context, tx pgx.Tx, id int64) (stmt string, err error) {
	return lyspg.Archive(ctx, tx, schemaName, tableName, pkColName, id, false)
}

func (s Store) GetMeta() lysmeta.Result {
	return meta
}
func (s Store) GetName() string {
	return name
}

func (s Store) Restore(ctx context.Context, tx pgx.Tx, id int64) (stmt string, err error) {
	return lyspg.Restore(ctx, tx, schemaName, tableName, pkColName, id, false)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, stmt string, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, meta.DbTags, params)
}

func (s Store) SelectById(ctx context.Context, fields []string, id int64) (item Model, stmt string, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, fields, meta.DbTags, id)
}
