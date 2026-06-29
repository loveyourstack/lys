package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys/lyspg"
)

// iGetableById is a store that can be used by GetById.
type iGetableById[idT lyspg.PrimaryKeyType, outT any] interface {
	SelectById(ctx context.Context, id idT) (item outT, err error)
}

// GetById handles retrieval of a single item from the supplied store.
func GetById[idT lyspg.PrimaryKeyType, outT any](env Env, store iGetableById[idT, outT]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get the id param and parse it into an idT
		id, err := getIdFromReq[idT](r)
		if err != nil {
			HandleError(ctx, fmt.Errorf("GetById: getIdFromReq failed: %w", err), env.Logger, w)
			return
		}

		// select item from Db
		item, err := store.SelectById(ctx, id)
		if err != nil {
			HandleError(ctx, fmt.Errorf("GetById: SelectById failed: %w", err), env.Logger, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   item,
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}
