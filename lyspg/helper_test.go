package lyspg

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/internal/lyscmd"
	"github.com/loveyourstack/lys/lyspgdb"
)

func mustGetDb(t testing.TB, ctx context.Context) *pgxpool.Pool {

	conf := lyscmd.MustGetConfig(t)

	var err error
	// register core.weekday type in any conn added to the pool so that Patch of type_test core.weekday[] works. If don't do this: "encode plan not found"
	dataTypeNames := []string{
		"core.weekday",
		"core.weekday[]",
	}
	db, err := lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbOwnerUser, "test", dataTypeNames)
	if err != nil {
		t.Fatalf("lyspgdb.GetPoolWithTypes failed: %v", err)
	}

	return db
}
