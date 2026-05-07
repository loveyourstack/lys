package corearchivetestuuid

import (
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

const (
	name           string = "Archive test (UUID)"
	schemaName     string = "core"
	tableName      string = "archive_test_uuid"
	viewName       string = "archive_test_uuid"
	pkColName      string = "id"
	defaultOrderBy string = "id"
)

type Input struct {
	CInt  *int64  `db:"c_int" json:"c_int,omitempty"`
	CText *string `db:"c_text" json:"c_text,omitempty"`
}

type Model struct {
	Id        uuid.UUID        `db:"id" json:"id,omitempty"`
	CreatedAt lystype.Datetime `db:"created_at" json:"created_at,omitzero"`
	Input
}

var (
	plan lysmeta.Plan
)

func init() {
	var err error
	plan, err = lysmeta.Analyze(Model{})
	if err != nil {
		log.Fatalf("lysmeta.Analyze failed for %s.%s: %s", schemaName, tableName, err.Error())
	}
}

type Store struct {
	Db *pgxpool.Pool
}

func New(db *pgxpool.Pool) Store {
	return Store{
		Db: db,
	}
}

func (s Store) Archive(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	return lyspg.Archive(ctx, tx, schemaName, tableName, pkColName, id, false)
}

func (s Store) GetName() string {
	return name
}
func (s Store) GetPlan() lysmeta.Plan {
	return plan
}

func (s Store) Restore(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	return lyspg.Restore(ctx, tx, schemaName, tableName, pkColName, id, false)
}

func (s Store) Select(ctx context.Context, params lyspg.SelectParams) (items []Model, unpagedCount lyspg.TotalCount, err error) {
	return lyspg.Select[Model](ctx, s.Db, schemaName, tableName, viewName, defaultOrderBy, plan.DbNames(), params)
}

func (s Store) SelectById(ctx context.Context, id uuid.UUID) (item Model, err error) {
	return lyspg.SelectUnique[Model](ctx, s.Db, schemaName, viewName, pkColName, id)
}
