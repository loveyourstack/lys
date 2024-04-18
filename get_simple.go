package lys

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

// GetSimple handles retrieval of all items returned by selectFunc, which may only take ctx as param
func GetSimple[T any](env Env, selectFunc func(ctx context.Context) (items []T, stmt string, err error)) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// select items from Db
		items, stmt, err := selectFunc(r.Context())
		if err != nil {

			// expected error: request canceled
			if errors.Is(err, context.Canceled) {
				return
			}

			// unknown db error
			HandleDbError(r.Context(), stmt, fmt.Errorf("GetSimple: selectFunc failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   items,
		}
		JsonResponse(resp, http.StatusOK, nil, w)
	}
}
