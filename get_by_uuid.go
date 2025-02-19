package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// iGetableByUuid is a store that can be used by GetByUuid
type iGetableByUuid[T any] interface {
	SelectByUuid(ctx context.Context, id uuid.UUID) (item T, err error)
}

// GetByUuid handles retrieval of a single item from the supplied store using a text id
func GetByUuid[T any](env Env, store iGetableByUuid[T]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the uuid and ensure it is valid
		uuidStr := mux.Vars(r)["id"]
		if uuidStr == "" {
			HandleUserError(http.StatusBadRequest, ErrDescIdMissing, w)
			return
		}

		idUu, err := uuid.Parse(uuidStr)
		if err != nil {
			HandleUserError(http.StatusBadRequest, ErrDescIdNotAUuid, w)
			return
		}

		// select item from Db
		item, err := store.SelectByUuid(r.Context(), idUu)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("GetByUuid: store.SelectByUuid failed: %w", err), env.ErrorLog, w)
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
