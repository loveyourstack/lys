package corearchivetest

import (
	"context"
	"log"
	"reflect"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/internal/stores/core/corearchivetestm"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

const (
	name           string = "Archive test"
	schemaName     string = "core"
	tableName      string = "archive_test"
	viewName       string = "archive_test"
	pkColName      string = "id"
	defaultOrderBy string = "id"
)

// Input and Model are in a separate package, corearchivetestm, so that they can be used for testing in lyspg

var (
	meta lysmeta.Result
)

func init() {
	var err error
	meta, err = lysmeta.AnalyzeStructs(reflect.ValueOf(&corearchivetestm.Input{}).Elem(), reflect.ValueOf(&corearchivetestm.Model{}).Elem())
	if err != nil {
		log.Fatalf("lysmeta.AnalyzeStructs failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
}

type Store struct {
	Db *pgxpool.Pool
}

func (s Store) ArchiveById(ctx context.Context, tx pgx.Tx, id int64) error {
	return lyspg.Archive(ctx, tx, schemaName, tableName, pkColName, id, false)
}

func (s Store) ArchiveByUuid(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	return lyspg.Archive(ctx, tx, schemaName, tableName, "id_uu", id, false)
}

func (s Store) GetMeta() lysmeta.Result {
	return meta
}
func (s Store) GetName() string {
	return name
}

func (s Store) RestoreById(ctx context.Context, tx pgx.Tx, id int64) error {
	return lyspg.Restore(ctx, tx, schemaName, tableName, pkColName, id, false)
}

func (s Store) RestoreByUuid(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	return lyspg.Restore(ctx, tx, schemaName, tableName, "id_uu", id, false)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []corearchivetestm.Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[corearchivetestm.Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, meta.DbTags, params)
}

func (s Store) SelectById(ctx context.Context, id int64) (item corearchivetestm.Model, err error) {
	return lyspg.SelectUnique[corearchivetestm.Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}

func (s Store) SelectByUuid(ctx context.Context, id uuid.UUID) (item corearchivetestm.Model, err error) {
	return lyspg.SelectUnique[corearchivetestm.Model](ctx, s.Db, schemaName, viewName, "id_uu", id)
}
