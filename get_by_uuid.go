package lys

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysmeta"
)

// iGetableByUuid is a store that can be used by GetByUuid
type iGetableByUuid[T any] interface {
	GetMeta() lysmeta.Result
	SelectByUuid(ctx context.Context, fields []string, id uuid.UUID) (item T, stmt string, err error)
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

		// get fields param if present (e.g. &xfields=)
		fields, err := ExtractFields(r, store.GetMeta().JsonTags, env.GetOptions.FieldsParamName)
		if err != nil {
			if userErr, ok := err.(lyserr.User); ok {
				HandleUserError(http.StatusBadRequest, userErr.Message, w)
			} else {
				HandleInternalError(r.Context(), fmt.Errorf("GetByUuid: ExtractFields failed: %w", err), env.ErrorLog, w)
			}
			return
		}

		// select item from Db
		item, stmt, err := store.SelectByUuid(r.Context(), fields, idUu)
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
			HandleDbError(r.Context(), stmt, fmt.Errorf("GetByUuid: store.SelectByUuid failed: %w", err), env.ErrorLog, w)
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
