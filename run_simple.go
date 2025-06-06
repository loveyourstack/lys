package lys

import (
	"context"
	"fmt"
	"net/http"
)

// RunSimple runs the supplied func
func RunSimple(env Env, runFunc func(context.Context) error) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// run the func
		err := runFunc(r.Context())
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("ProcessSimple: runFunc failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   "ok",
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}
