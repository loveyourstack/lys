package lyspgauditupdate

import (
	"context"
	"log"
	"reflect"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Pg audit update"
	schemaName     string = "lyspgmon"
	tableName      string = "audit_update"
	viewName       string = "audit_update"
	pkColName      string = "id"
	defaultOrderBy string = "affected_at DESC"
)

// No input: records are created by t_audit_update trigger

type Model struct {
	Id                int64            `db:"id" json:"id"`
	AffectedAt        lystype.Datetime `db:"affected_at" json:"affected_at"`
	AffectedBy        string           `db:"affected_by" json:"affected_by"`
	AffectedId        int64            `db:"affected_id" json:"affected_id"`
	AffectedNewValues map[string]any   `db:"affected_new_values" json:"affected_new_values"`
	AffectedOldValues map[string]any   `db:"affected_old_values" json:"affected_old_values"`
	AffectedSchema    string           `db:"affected_schema" json:"affected_schema"`
	AffectedTable     string           `db:"affected_table" json:"affected_table"`
}

var (
	meta lysmeta.Result
)

func init() {
	var err error
	meta, err = lysmeta.AnalyzeStruct(reflect.ValueOf(&Model{}).Elem())
	if err != nil {
		log.Fatalf("lysmeta.AnalyzeStruct failed for %s.%s: %s", schemaName, tableName, err.Error())
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

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, meta.DbTags, params)
}

func (s Store) SelectById(ctx context.Context, id int64) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}
