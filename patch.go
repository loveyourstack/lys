package lys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lyspg"
)

// iPatchable is a store that can be used by Patch.
type iPatchable[idT lyspg.PrimaryKeyType] interface {
	UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id idT) error
}

// Patch handles changing some of an item's fields using the supplied store.
func Patch[idT lyspg.PrimaryKeyType](env Env, store iPatchable[idT]) http.HandlerFunc {

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

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Patch: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
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
			HandleUserError(ErrNoAssignments, w)
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
