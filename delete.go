package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lyspg"
)

// iDeletable is a store that can be used by Delete.
type iDeletable[idT lyspg.PrimaryKeyType] interface {
	Delete(ctx context.Context, id idT) error
}

// Delete handles deletion of a single item using the supplied store.
func Delete[idT lyspg.PrimaryKeyType](env Env, store iDeletable[idT]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the id param
		idStr := mux.Vars(r)["id"]
		if idStr == "" {
			HandleUserError(ErrIdMissing, w)
			return
		}

		// parse the id param into a idT
		id, err := parseIdByType[idT](idStr)
		if err != nil {
			HandleUserError(ErrIdParseError, w)
			return
		}

		// delete item from db
		err = store.Delete(r.Context(), id)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Delete: store.Delete failed: %w", err), env.ErrorLog, w)
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
