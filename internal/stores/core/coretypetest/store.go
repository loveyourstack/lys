package coretypetest

import (
	"context"
	"log"
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	name           string = "Type test"
	schemaName     string = "core"
	tableName      string = "type_test"
	viewName       string = "type_test"
	pkColName      string = "id"
	defaultOrderBy string = "id"
)

// Input and Model are in a separate package, coretypetestm, so that they can be used for testing in lyspg

var (
	meta, inputMeta lysmeta.Result
)

func init() {
	var err error
	meta, err = lysmeta.AnalyzeStructs(reflect.ValueOf(&coretypetestm.Input{}).Elem(), reflect.ValueOf(&coretypetestm.Model{}).Elem())
	if err != nil {
		log.Fatalf("lysmeta.AnalyzeStructs failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
	inputMeta, _ = lysmeta.AnalyzeStructs(reflect.ValueOf(&coretypetestm.Input{}).Elem())
}

type Store struct {
	Db *pgxpool.Pool
}

func (s Store) Delete(ctx context.Context, id int64) error {
	return lyspg.DeleteUnique(ctx, s.Db, schemaName, tableName, pkColName, id)
}

func (s Store) GetMeta() lysmeta.Result {
	return meta
}
func (s Store) GetName() string {
	return name
}

func (s Store) Insert(ctx context.Context, input coretypetestm.Input) (newId int64, err error) {
	return lyspg.Insert[coretypetestm.Input, int64](ctx, s.Db, schemaName, tableName, pkColName, input)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []coretypetestm.Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[coretypetestm.Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, meta.DbTags, params)
}

func (s Store) SelectById(ctx context.Context, fields []string, id int64) (item coretypetestm.Model, err error) {
	return lyspg.SelectUnique[coretypetestm.Model](ctx, s.Db, schemaName, viewName, pkColName, fields, meta.DbTags, id)
}

func (s Store) SelectByUuid(ctx context.Context, fields []string, id uuid.UUID) (item coretypetestm.Model, err error) {
	return lyspg.SelectUnique[coretypetestm.Model](ctx, s.Db, schemaName, viewName, "id_uu", fields, meta.DbTags, id)
}

func (s Store) Update(ctx context.Context, input coretypetestm.Input, id int64) error {
	return lyspg.Update(ctx, s.Db, schemaName, tableName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) error {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, inputMeta.DbTags, assignmentsMap, id)
}

func (s Store) Validate(validate *validator.Validate, input coretypetestm.Input) error {
	return lysmeta.Validate[coretypetestm.Input](validate, input)
}
