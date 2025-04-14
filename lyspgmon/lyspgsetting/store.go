package lyspgsetting

import (
	"context"
	"log"
	"reflect"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	name           string = "Pg settings"
	schemaName     string = "lyspgmon"
	tableName      string = "v_settings"
	viewName       string = "v_settings"
	defaultOrderBy string = "name"
)

type Model struct {
	Name      string `db:"name" json:"name"`
	Setting   string `db:"setting" json:"setting"`
	BootVal   string `db:"boot_val" json:"boot_val"`
	Unit      string `db:"unit" json:"unit"`
	Context   string `db:"context" json:"context"`
	ShortDesc string `db:"short_desc" json:"short_desc"`
	ExtraDesc string `db:"extra_desc" json:"extra_desc"`
	Changed   bool   `db:"changed" json:"changed"`
}

var (
	meta lysmeta.Result
)

func init() {
	var err error
	meta, err = lysmeta.AnalyzeStructs(reflect.ValueOf(&Model{}).Elem())
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

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, meta.DbTags, params)
}
