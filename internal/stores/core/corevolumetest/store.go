package corevolumetest

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	schemaName     string = "core"
	tableName      string = "volume_test"
	viewName       string = "volume_test"
	pkColName      string = "id"
	defaultOrderBy string = "id"
)

type Input struct {
	CRnd int64 `db:"c_rnd" json:"c_rnd"`
	CInt int64 `db:"c_int" json:"c_int"`
}

type Model struct {
	Id int64 `db:"id" json:"id"`
	Input
}

var (
	gDbTags      []string
	gJsonTags    []string
	gInputDbTags []string
)

func init() {
	var err error
	gDbTags, gJsonTags, err = lysmeta.GetStructTags(reflect.ValueOf(&Input{}).Elem(), reflect.ValueOf(&Model{}).Elem())
	if err != nil {
		log.Fatalf("lysmeta.GetStructTags failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
	gInputDbTags, _, _ = lysmeta.GetStructTags(reflect.ValueOf(&Input{}).Elem())
}

type Store struct {
	Db *pgxpool.Pool
}

func (s Store) Delete(ctx context.Context, id int64) (stmt string, err error) {
	return lyspg.DeleteUnique(ctx, s.Db, schemaName, tableName, pkColName, id)
}

func (s Store) GetJsonFields() []string {
	return gJsonTags
}

func (s Store) Insert(ctx context.Context, input Input) (newItem Model, stmt string, err error) {
	return lyspg.Insert[Input, Model](ctx, s.Db, schemaName, tableName, viewName, pkColName, gDbTags, input)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, stmt string, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, gDbTags, params)
}

func (s Store) Select10(ctx context.Context) (vals []int, stmt string, err error) {

	stmt = "SELECT c_int FROM core.volume_test ORDER BY c_int LIMIT 10;"
	rows, _ := s.Db.Query(ctx, stmt)
	vals, err = pgx.CollectRows(rows, pgx.RowTo[int])
	if err != nil {
		return nil, stmt, fmt.Errorf("pgx.CollectRows failed: %w", err)
	}

	return vals, "", nil
}

func (s Store) SelectById(ctx context.Context, fields []string, id int64) (item Model, stmt string, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, fields, gDbTags, id)
}

func (s Store) Update(ctx context.Context, input Input, id int64) (stmt string, err error) {
	return lyspg.Update(ctx, s.Db, schemaName, tableName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (stmt string, err error) {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, gInputDbTags, assignmentsMap, id)
}

func (s Store) Validate(validate *validator.Validate, input Input) error {
	return lysmeta.Validate[Input](validate, input)
}
