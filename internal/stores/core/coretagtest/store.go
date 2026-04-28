package coretagtest

import (
	"context"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	name           string = "Tag test"
	schemaName     string = "core"
	tableName      string = "tag_test"
	viewName       string = "v_tag_test"
	pkColName      string = "id"
	defaultOrderBy string = "id"
)

type Input struct {
	CEditable string `db:"c_editable" json:"c_editable"`
	CHidden   string `db:"c_hidden" json:"-"` // in db, but hidden to API. Insert / update value must be added in app code
}

type Model struct {
	Id     int64  `db:"id" json:"id"`
	CExtra string `json:"c_extra"` // no db tag: column not in db. Select value must be added in app code
	Input
}

var (
	plan, inputPlan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.AnalyzeAndCheckT(Model{})
	if err != nil {
		log.Fatalf("lysmeta.AnalyzeAndCheckT failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
	inputPlan, _ = lysmeta.AnalyzeAndCheckT(Input{})
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

func (s Store) Insert(ctx context.Context, input Input) (newId int64, err error) {

	input.CHidden = "y" // required field in db, but no value incoming from API, so must be set here

	return lyspg.Insert[Input, int64](ctx, s.Db, schemaName, tableName, pkColName, input)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	items, unpagedCount, err = lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
	if err != nil {
		return nil, lyspg.TotalCount{}, fmt.Errorf("lyspg.Select failed: %w", err)
	}

	for i := range items {
		items[i].CExtra = "extra" // field not in db, added in app code
	}

	return items, unpagedCount, nil
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	item, err = lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
	if err != nil {
		return item, fmt.Errorf("lyspg.SelectUnique failed: %w", err)
	}

	item.CExtra = "extra" // field not in db, added in app code

	return item, nil
}

func (s Store) Update(ctx context.Context, input Input, id int64) error {

	input.CHidden = "d1" // required field in db, but no value incoming from API, so must be set here

	return lyspg.Update(ctx, s.Db, schemaName, tableName, pkColName, input, id)
}

func (s Store) UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) error {
	return lyspg.UpdatePartial(ctx, s.Db, schemaName, tableName, pkColName, inputPlan.DbNames(), assignmentsMap, id)
}

func (s Store) Validate(validate *validator.Validate, input Input) error {
	return lysmeta.Validate(validate, input)
}
