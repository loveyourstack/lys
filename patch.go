package lys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys/lyspg"
)

// iPatchable is a store that can be used by Patch.
type iPatchable[idT lyspg.PrimaryKeyType] interface {
	UpdatePartial(ctx context.Context, assignmentsMap map[string]any, id idT) error
}

// Patch handles changing some of an item's fields using the supplied store.
func Patch[idT lyspg.PrimaryKeyType](env Env, store iPatchable[idT]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get the id param and parse it into an idT
		id, err := getIdFromReq[idT](r)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Patch: getIdFromReq failed: %w", err), env.ErrorLog, w)
			return
		}

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Patch: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// unmarshal the body
		assignmentsMap := make(map[string]any)
		err = json.Unmarshal(body, &assignmentsMap)
		if err != nil {
			HandleUserError(ErrNotParseableToMap, w)
			return
		}
		if len(assignmentsMap) == 0 {
			HandleUserError(ErrNoAssignments, w)
			return
		}

		// try to update the item in db
		err = store.UpdatePartial(ctx, assignmentsMap, id)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Patch: store.UpdatePartial failed: %w", err), env.ErrorLog, w)
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
