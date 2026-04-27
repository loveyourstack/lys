package corebulkdeletetest

import (
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
)

const (
	name       string = "Bulk delete test"
	schemaName string = "core"
	tableName  string = "bulk_delete_test"
)

type Input struct {
	CText string `db:"c_text" json:"c_text,omitempty"`
}

type Model struct {
	Id int64 `db:"id" json:"id,omitempty"`
	Input
}

var (
	plan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.AnalyzeAndCheckT(Model{})
	if err != nil {
		log.Fatalf("lysmeta.AnalyzeAndCheckT failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
}

type Store struct {
	Db *pgxpool.Pool
}

func (s Store) GetName() string {
	return name
}
func (s Store) GetPlan() lysmeta.Plan {
	return plan
}
