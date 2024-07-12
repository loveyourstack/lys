package coreparamtest

import (
	"context"
	"log"
	"reflect"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Param test"
	schemaName     string = "core"
	tableName      string = "param_test"
	viewName       string = "param_test"
	pkColName      string = "id"
	defaultOrderBy string = "id"
)

type Input struct {
	CBool      bool              `db:"c_bool" json:"c_bool"`
	CBoolN     *bool             `db:"c_booln" json:"c_booln"`
	CInt       int64             `db:"c_int" json:"c_int,omitempty"`
	CIntN      *int64            `db:"c_intn" json:"c_intn,omitempty"`
	CDouble    float32           `db:"c_double" json:"c_double,omitempty"`
	CDoubleN   *float32          `db:"c_doublen" json:"c_doublen,omitempty"`
	CDate      lystype.Date      `db:"c_date" json:"c_date,omitempty"`
	CDateN     *lystype.Date     `db:"c_daten" json:"c_daten,omitempty"`
	CTime      lystype.Time      `db:"c_time" json:"c_time,omitempty"`
	CTimeN     *lystype.Time     `db:"c_timen" json:"c_timen,omitempty"`
	CDatetime  lystype.Datetime  `db:"c_datetime" json:"c_datetime,omitempty"`
	CDatetimeN *lystype.Datetime `db:"c_datetimen" json:"c_datetimen,omitempty"`
	CEnum      string            `db:"c_enum" json:"c_enum,omitempty"`
	CEnumN     *string           `db:"c_enumn" json:"c_enumn,omitempty"`
	CText      string            `db:"c_text" json:"c_text,omitempty"`
	CTextN     *string           `db:"c_textn" json:"c_textn,omitempty"`
}

type Model struct {
	Id   int64     `db:"id" json:"id,omitempty"`
	Iduu uuid.UUID `db:"id_uu" json:"id_uu,omitempty"`
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

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, stmt string, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, meta.DbTags, params)
}

func (s Store) SelectById(ctx context.Context, fields []string, id int64) (item Model, stmt string, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, fields, meta.DbTags, id)
}
