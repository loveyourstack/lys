package corebulkdeletetest

import (
	"log"
	"reflect"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
)

const (
	name           string = "Bulk delete test"
	schemaName     string = "core"
	tableName      string = "bulk_delete_test"
	viewName       string = "bulk_delete_test"
	pkColName      string = "id"
	defaultOrderBy string = "id"
)

type Input struct {
	CText string `db:"c_text" json:"c_text,omitempty"`
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

func (s Store) GetMeta() lysmeta.Result {
	return meta
}
func (s Store) GetName() string {
	return name
}
