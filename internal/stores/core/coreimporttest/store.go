package coreimporttest

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/go-playground/validator/v10"
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

func (s Store) BulkInsert(ctx context.Context, inputs []coretypetestm.Input) (rowsAffected int64, err error) {
	return lyspg.BulkInsert(ctx, s.Db, schemaName, tableName, inputs)
}

func (s Store) GetMeta() lysmeta.Result {
	return meta
}
func (s Store) GetName() string {
	return name
}

func (s Store) Validate(validate *validator.Validate, input coretypetestm.Input) error {

	// add dummy failure condition for tests
	if input.CText == "fail" {
		return fmt.Errorf("CText is invalid")
	}

	return lysmeta.Validate(validate, input)
}
