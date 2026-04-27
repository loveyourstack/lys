package coretypetest

import (
	"context"
	"log"

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
	plan, inputPlan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.AnalyzeAndCheckT(coretypetestm.Model{})
	if err != nil {
		log.Fatalf("lysmeta.AnalyzeAndCheckT failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
	inputPlan, _ = lysmeta.AnalyzeAndCheckT(coretypetestm.Input{})
}

type Store struct {
	Db *pgxpool.Pool
}

func (s Store) Delete(ctx context.Context, id int64) error {
	return lyspg.DeleteUnique(ctx, s.Db, schemaName, tableName, pkColName, id)
}

func (s Store) GetName() string {
	return name
}
func (s Store) GetPlan() lysmeta.Plan {
	return plan
}

func (s Store) Insert(ctx context.Context, input coretypetestm.Input) (newId int64, err error) {
	return lyspg.Insert[coretypetestm.Input, int64](ctx, s.Db, schemaName, tableName, pkColName, input)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []coretypetestm.Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[coretypetestm.Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

func (s Store) SelectById(ctx context.Context, id int64) (item coretypetestm.Model, err error) {
	return lyspg.SelectUnique[coretypetestm.Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}

func (s Store) SelectByUuid(ctx context.Context, id uuid.UUID) (item coretypetestm.Model, err error) {
	return lyspg.SelectUnique[coretypetestm.Model](ctx, s.Db, schemaName, viewName, "id_uu", id)
}

func (s Store) Update(ctx context.Context, input coretypetestm.Input, id int64) error {
	return lyspg.Update(ctx, s.Db, schemaName, tableName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) error {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, inputPlan.DbNames(), assignmentsMap, id)
}

func (s Store) Validate(validate *validator.Validate, input coretypetestm.Input) error {
	return lysmeta.Validate(validate, input)
}
