package lys

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lyserr"
)

// iPatchable is a store that can be used by Patch
type iPatchable interface {
	UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id int64) error
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
		err = store.UpdatePartial(r.Context(), assignmentsMap, id)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Patch: store.UpdatePartial failed: %w", err), env.ErrorLog, w)
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
