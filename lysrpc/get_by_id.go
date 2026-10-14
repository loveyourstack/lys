package lysrpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyspg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetById retrieves a record by its primary key and returns the corresponding gRPC model.
func GetById[idT lyspg.PrimaryKeyType, modelT any, pbModelT any](
	ctx context.Context,
	id idT,
	store lyspg.SelecterById[idT, modelT],
	toPb func(modelT) *pbModelT,
) (*pbModelT, error) {

	item, err := store.SelectById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		return nil, fmt.Errorf("store.SelectById failed: %w", err)
	}

	return toPb(item), nil
}
