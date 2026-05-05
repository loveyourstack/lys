package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lyspg"
)

// iGetableById is a store that can be used by GetById
type iGetableById[T any] interface {
	SelectById(ctx context.Context, id int64) (item T, err error)
}

// iGetableByUuid is a store that can be used by GetByUuid
type iGetableByUuid[T any] interface {
	SelectByUuid(ctx context.Context, id uuid.UUID) (item T, err error)
}

// GetById handles retrieval of a single item from the supplied store using an integer id
func GetById[T any](env Env, store iGetableById[T]) http.HandlerFunc {
	return getSingle(env, store.SelectById, parseIdFunc, "GetById")
}

// GetByUuid handles retrieval of a single item from the supplied store using a text id
func GetByUuid[T any](env Env, store iGetableByUuid[T]) http.HandlerFunc {
	return getSingle(env, store.SelectByUuid, parseUuidFunc, "GetByUuid")
}

func getSingle[idT lyspg.PrimaryKeyType, outT any](env Env, selectByIdFunc func(context.Context, idT) (outT, error),
	parseIdFunc func(string) (idT, error), callingFunc string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the id param
		idStr := mux.Vars(r)["id"]
		if idStr == "" {
			HandleUserError(ErrIdMissing, w)
			return
		}

		// parse the id param into a T
		id, err := parseIdFunc(idStr)
		if err != nil {
			HandleUserError(ErrIdParseError, w)
			return
		}

		// select item from Db
		item, err := selectByIdFunc(r.Context(), id)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("%s: selectByIdFunc failed: %w", callingFunc, err), env.ErrorLog, w)
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
