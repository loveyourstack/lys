package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lyspg"
)

// iGetableById is a store that can be used by GetById.
type iGetableById[idT lyspg.PrimaryKeyType, outT any] interface {
	SelectById(ctx context.Context, id idT) (item outT, err error)
}

// GetById handles retrieval of a single item from the supplied store.
func GetById[idT lyspg.PrimaryKeyType, outT any](env Env, store iGetableById[idT, outT]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the id param
		idStr := mux.Vars(r)["id"]
		if idStr == "" {
			HandleUserError(ErrIdMissing, w)
			return
		}

		// parse the id param into an idT
		id, err := parseIdByType[idT](idStr)
		if err != nil {
			HandleUserError(ErrIdParseError, w)
			return
		}

		// select item from Db
		item, err := store.SelectById(r.Context(), id)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("GetById: SelectById failed: %w", err), env.ErrorLog, w)
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
