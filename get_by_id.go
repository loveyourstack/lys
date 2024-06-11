package lys

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
)

// iGetablebyId is a store that can be used by GetById
type iGetablebyId[T any] interface {
	GetMeta() lysmeta.Result
	SelectById(ctx context.Context, fields []string, id int64) (item T, stmt string, err error)
}

// GetById handles retrieval of a single item from the supplied store using an integer id
func GetById[T any](env Env, store iGetablebyId[T]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the Id and ensure it is an int
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			HandleUserError(http.StatusBadRequest, ErrDescIdNotAnInteger, w)
			return
		}

		// get fields param if present (e.g. &xfields=)
		fields, err := ExtractFields(r, store.GetMeta().JsonTags, env.GetOptions.FieldsParamName)
		if err != nil {
			if userErr, ok := err.(lyserr.User); ok {
				HandleUserError(http.StatusBadRequest, userErr.Message, w)
			} else {
				HandleInternalError(r.Context(), fmt.Errorf("GetById: ExtractFields failed: %w", err), env.ErrorLog, w)
			}
			return
		}

		// select item from Db
		item, stmt, err := store.SelectById(r.Context(), fields, id)
		if err != nil {

			// expected error: request canceled
			if errors.Is(err, context.Canceled) {
				return
			}

			// expected error: Id does not exist
			if errors.Is(err, pgx.ErrNoRows) {
				HandleUserError(http.StatusBadRequest, ErrDescInvalidId, w)
				return
			}

			// unknown db error
			HandleDbError(r.Context(), stmt, fmt.Errorf("GetById: store.SelectById failed: %w", err), env.ErrorLog, w)
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
