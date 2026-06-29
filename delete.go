package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys/lyspg"
)

// iDeletable is a store that can be used by Delete.
type iDeletable[idT lyspg.PrimaryKeyType] interface {
	Delete(ctx context.Context, id idT) error
}

// Delete handles deletion of a single item using the supplied store.
func Delete[idT lyspg.PrimaryKeyType](env Env, store iDeletable[idT]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get the id param and parse it into an idT
		id, err := getIdFromReq[idT](r)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Delete: getIdFromReq failed: %w", err), env.Logger, w)
			return
		}

		// delete item from db
		err = store.Delete(ctx, id)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Delete: store.Delete failed: %w", err), env.Logger, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   DataDeleted,
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}
