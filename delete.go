package lys

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// iDeletable is a store that can be used by Delete
type iDeletable interface {
	Delete(ctx context.Context, id int64) error
}

// Delete handles deletion of a single item using the supplied store
func Delete(env Env, store iDeletable) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the Id and ensure it is an int
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			HandleUserError(ErrIdNotAnInteger, w)
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
