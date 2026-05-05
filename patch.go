package lys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lyspg"
)

// iPatchableById is a store that can be used by PatchById.
type iPatchableById interface {
	UpdatePartialById(ctx context.Context, assignmentsMap map[string]any, id int64) error
}

// iPatchableByUuid is a store that can be used by PatchByUuid.
type iPatchableByUuid interface {
	UpdatePartialByUuid(ctx context.Context, assignmentsMap map[string]any, id uuid.UUID) error
}

// PatchById handles changing some of an item's fields using the supplied store and an int64 id.
func PatchById(env Env, store iPatchableById) http.HandlerFunc {
	return patch(env, store.UpdatePartialById, parseIdFunc, "PatchById")
}

// PatchByUuid handles changing some of an item's fields using the supplied store and a UUID.
func PatchByUuid(env Env, store iPatchableByUuid) http.HandlerFunc {
	return patch(env, store.UpdatePartialByUuid, parseUuidFunc, "PatchByUuid")
}

// patch handles changing some of an item's fields using the supplied store.
func patch[idT lyspg.PrimaryKeyType](env Env, patchFunc func(context.Context, map[string]any, idT) error,
	parseIdFunc func(string) (idT, error), callingFunc string) http.HandlerFunc {

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

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("%s: ExtractJsonBody failed: %w", callingFunc, err), env.ErrorLog, w)
			return
		}

		// unmarshal the body
		assignmentsMap := make(map[string]any)
		err = json.Unmarshal(body, &assignmentsMap)
		if err != nil {
			HandleInternalError(r.Context(), fmt.Errorf("%s: json.Unmarshal failed: %w", callingFunc, err), env.ErrorLog, w)
			return
		}
		if len(assignmentsMap) == 0 {
			HandleUserError(ErrNoAssignments, w)
			return
		}

		// try to update the item in db
		err = patchFunc(r.Context(), assignmentsMap, id)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("%s: patchFunc failed: %w", callingFunc, err), env.ErrorLog, w)
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
