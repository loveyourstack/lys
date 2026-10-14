package lysrpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyspg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Delete handles deletion of a single item using the supplied store.
func Delete[idT lyspg.PrimaryKeyType](ctx context.Context, id idT, store lyspg.Deleter[idT]) (*emptypb.Empty, error) {

	err := store.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "item not found")
		}
		return nil, fmt.Errorf("store.Delete failed: %w", err)
	}

	return &emptypb.Empty{}, nil
}
