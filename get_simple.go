package lys

import (
	"context"
	"fmt"
	"net/http"
)

// GetSimple handles retrieval of all items returned by selectFunc, which may only take ctx as param
func GetSimple[T any](env Env, selectFunc func(ctx context.Context) (items []T, err error)) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// select items from Db
		items, err := selectFunc(r.Context())
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("GetSimple: selectFunc failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   items,
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}
