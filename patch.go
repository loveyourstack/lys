package lys

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/loveyourstack/lys/lyserr"
)

// iPatchable is a store that can be used by Patch
type iPatchable interface {
	UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) (stmt string, err error)
}

// Patch handles changing some of an item's fields using the supplied store
func Patch(env Env, store iPatchable) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the Id and ensure it is an int
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			HandleUserError(http.StatusBadRequest, ErrDescIdNotAnInteger, w)
			return
		}

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			var userErr lyserr.User
			if errors.As(err, &userErr) {
				HandleUserError(http.StatusBadRequest, userErr.Message, w)
			} else {
				HandleInternalError(r.Context(), fmt.Errorf("Patch: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			}
			return
		}

		// unmarshal the body
		assignmentsMap := make(map[string]any)
		err = json.Unmarshal(body, &assignmentsMap)
		if err != nil {
			HandleInternalError(r.Context(), fmt.Errorf("Patch: json.Unmarshal failed: %w", err), env.ErrorLog, w)
			return
		}
		if len(assignmentsMap) == 0 {
			HandleUserError(http.StatusBadRequest, "no assignments found", w)
			return
		}

		// try to update the item in db
		stmt, err := store.UpdatePartial(r.Context(), assignmentsMap, id)
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

			// expected error: invalid field
			if strings.Contains(err.Error(), "invalid field") {
				HandleUserError(http.StatusBadRequest, err.Error(), w)
				return
			}

			// handle errors from postgres
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				HandlePostgresError(r.Context(), stmt, "Patch: store.UpdatePartial", pgErr, env.ErrorLog, w)
				return
			}

			// unknown db error
			HandleDbError(r.Context(), stmt, fmt.Errorf("Patch: store.UpdatePartial failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   DataUpdated,
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}
