package lys

import (
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys/lyspg"
)

// GetById handles retrieval of a single item from the supplied store.
func GetById[idT lyspg.PrimaryKeyType, outT any](env Env, store lyspg.SelecterById[idT, outT]) http.HandlerFunc {

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
