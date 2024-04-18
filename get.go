package lys

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspg"
)

// iGetable is a store that can be used by Get
type iGetable[T any] interface {
	GetJsonFields() []string
	Select(ctx context.Context, params lyspg.SelectParams) (items []T, unpagedCount lyspg.TotalCount, stmt string, err error)
}

// Get handles retrieval of multiple items from the supplied store
func Get[T any](env Env, store iGetable[T]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get request modifiers from Url params
		getReqModifiers, err := ExtractGetRequestModifiers(r, store.GetJsonFields(), env.GetOptions)
		if err != nil {
			if userErr, ok := err.(lyserr.User); ok {
				HandleUserError(http.StatusBadRequest, userErr.Message, w)
			} else {
				HandleInternalError(r.Context(), fmt.Errorf("Get: ExtractGetRequestModifiers failed: %w", err), env.ErrorLog, w)
			}
			return
		}

		// get offset from paging params (starts at 0, not 1)
		offset := getReqModifiers.PerPage * (getReqModifiers.Page - 1)

		// define params for store select func
		selectParams := lyspg.SelectParams{
			Fields:          getReqModifiers.Fields,
			Conditions:      getReqModifiers.Conditions,
			Sorts:           getReqModifiers.Sorts,
			Limit:           getReqModifiers.PerPage,
			Offset:          offset,
			GetUnpagedCount: true,
		}

		// select items from Db
		items, unpagedCount, stmt, err := store.Select(r.Context(), selectParams)
		if err != nil {

			// expected error: request canceled
			if errors.Is(err, context.Canceled) {
				return
			}

			// unknown db error
			HandleDbError(r.Context(), stmt, fmt.Errorf("Get: store.Select failed: %w", err), env.ErrorLog, w)
			return
		}

		// add unpagedCount as header
		headers := []RespHeader{}
		headers = append(headers, RespHeader{Key: "X-Total-Count", Value: strconv.FormatInt(unpagedCount.Value, 10)})
		headers = append(headers, RespHeader{Key: "X-Total-Count-Estimated", Value: strconv.FormatBool(unpagedCount.IsEstimated)})

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   items,
		}
		JsonResponse(resp, http.StatusOK, headers, w)
	}
}
