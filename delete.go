package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lyspg"
)

// iDeletableById is a store that can be used by DeleteById.
type iDeletableById interface {
	DeleteById(ctx context.Context, id int64) error
}

// iDeletableByUuid is a store that can be used by DeleteByUuid.
type iDeletableByUuid interface {
	DeleteByUuid(ctx context.Context, id uuid.UUID) error
}

// DeleteById handles deletion of a single item using the supplied store and an int64 id.
func DeleteById(env Env, store iDeletableById) http.HandlerFunc {
	return deleteCall(env, store.DeleteById, parseIdFunc, "DeleteById")
}

// DeleteByUuid handles deletion of a single item using the supplied store and a UUID.
func DeleteByUuid(env Env, store iDeletableByUuid) http.HandlerFunc {
	return deleteCall(env, store.DeleteByUuid, parseUuidFunc, "DeleteByUuid")
}

// deleteCall handles deletion of a single item using the supplied store. Not named "delete" to avoid confusion with the built in delete function.
func deleteCall[idT lyspg.PrimaryKeyType](env Env, deleteFunc func(context.Context, idT) error, parseIdFunc func(string) (idT, error),
	callingFunc string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the id param
		idStr := mux.Vars(r)["id"]
		if idStr == "" {
			HandleUserError(ErrIdMissing, w)
			return
		}

		// parse the id param into a idT
		id, err := parseIdFunc(idStr)
		if err != nil {
			HandleUserError(ErrIdParseError, w)
			return
		}

		// delete item from db
		err = deleteFunc(r.Context(), id)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("%s: deleteFunc failed: %w", callingFunc, err), env.ErrorLog, w)
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
