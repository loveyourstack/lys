package coreimporttest

import (
	"context"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	name       string = "Import test"
	schemaName string = "core"
	tableName  string = "import_test"
)

// Input and Model are in a separate package, coretypetestm, so that they can be used for testing in lyspg

var (
	plan, inputPlan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.Analyze(coretypetestm.Model{})
	if err != nil {
		log.Fatalf("lysmeta.Analyze failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
	inputPlan, _ = lysmeta.Analyze(coretypetestm.Input{})
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

func (s Store) GetName() string {
	return name
}
func (s Store) GetPlan() lysmeta.Plan {
	return plan
}

func (s Store) InsertTx(ctx context.Context, tx pgx.Tx, input coretypetestm.Input) (newId int64, err error) {
	return lyspg.Insert[coretypetestm.Input, int64](ctx, tx, schemaName, tableName, "id", input)
}

func (s Store) Validate(input coretypetestm.Input) error {

	// add dummy failure condition for tests
	if input.CText == "fail" {
		return fmt.Errorf("CText is invalid")
	}

	return lysmeta.Validate(s.Validator, input)
}
