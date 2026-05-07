package coretrackingtest

import (
	"context"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	name           string = "Tracking test"
	schemaName     string = "core"
	tableName      string = "tracking_test"
	viewName       string = "tracking_test"
	pkColName      string = "id"
	defaultOrderBy string = "id"
)

type Input struct {
	CEditable string `db:"c_editable" json:"c_editable"`
}

type Model struct {
	Id               int64  `db:"id" json:"id"`
	CreatedBy        string `db:"created_by" json:"created_by,omitempty"`                   // assigned in Insert func
	LastUserUpdateBy string `db:"last_user_update_by" json:"last_user_update_by,omitempty"` // assigned in Update funcs
	Input
}

var (
	plan, inputPlan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.Analyze(Model{})
	if err != nil {
		log.Fatalf("lysmeta.Analyze failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
	inputPlan, _ = lysmeta.Analyze(Input{})
}

type Store struct {
	Db        *pgxpool.Pool
	Validator *validator.Validate
}

func New(db *pgxpool.Pool, validator *validator.Validate) Store {
	return Store{
		Db:        db,
		Validator: validator,
	}
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

func (s Store) Insert(ctx context.Context, input Input) (newId int64, err error) {
	return lyspg.InsertWithExtras[Input, int64](ctx, s.Db, schemaName, tableName, pkColName, input, []string{"created_by"}, []any{"insert"})
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}

func (s Store) Update(ctx context.Context, input Input, id int64) error {
	return lyspg.UpdateWithExtras(ctx, s.Db, schemaName, tableName, pkColName, input, id, []string{"last_user_update_by"}, []any{"update"})
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) error {
	return lyspg.UpdatePartialWithExtras(ctx, s.Db, schemaName, tableName, pkColName, inputPlan.JsonKeyDbNameMap(), assignmentsMap, id, []string{"last_user_update_by"}, []any{"update partial"})
}

func (s Store) Validate(input Input) error {
	return lysmeta.Validate(s.Validator, input)
}
