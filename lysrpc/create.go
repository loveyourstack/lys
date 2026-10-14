package lysrpc

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/loveyourstack/lys/lyspg"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Create validates the provided gRPC input, inserts the record, and returns the corresponding gRPC model.
func Create[inputT any, modelT any, pbInputT any, pbModelT any](
	ctx context.Context,
	pbInput pbInputT,
	store lyspg.InsertSelecter[inputT, modelT],
	pbToInput func(pbInputT) inputT,
	toPb func(modelT) *pbModelT,
) (*pbModelT, error) {

	input := pbToInput(pbInput)

	if err := store.Validate(validator.New(), input); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	item, err := store.InsertSelect(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("store.InsertSelect failed: %w", err)
	}

	return toPb(item), nil
}
