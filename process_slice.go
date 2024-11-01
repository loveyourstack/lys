package lys

import (
	"context"
	"fmt"
	"net/http"
)

// ProcessSlice extracts a slice from the req body and passes it into the supplied processFunc
func ProcessSlice[T any](env Env, processFunc func(context.Context, []T) (int64, error)) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("ProcessSlice: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// unmarshal the body
		vals, err := DecodeJsonBody[[]T](body)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("ProcessSlice: DecodeJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// run the process func
		respVal, err := processFunc(r.Context(), vals)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("ProcessSlice: processFunc failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   respVal,
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}
