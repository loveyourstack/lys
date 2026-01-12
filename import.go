package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/loveyourstack/lys/lyserr"
)

// iImportable is a store that can be used by Import
type iImportable[T any] interface {
	BulkInsert(ctx context.Context, inputs []T) (rowsAffected int64, err error)
	Validate(validate *validator.Validate, input T) error
}

// Import handles creating multiple new items using the supplied store and returning the number of rows inserted
func Import[T any](env Env, store iImportable[T]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Import: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// unmarshal the body into a slice of inputs
		inputs, err := DecodeJsonBody[[]T](body)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Import: DecodeJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// check for empty input slice
		if len(inputs) == 0 {
			HandleUserError(lyserr.User{Message: "no inputs found", StatusCode: http.StatusUnprocessableEntity}, w)
			return
		}

		// validate each item
		for i, input := range inputs {
			if err = store.Validate(env.Validate, input); err != nil {
				HandleUserError(lyserr.User{Message: fmt.Sprintf("line %v: %s", i+1, err.Error()), StatusCode: http.StatusUnprocessableEntity}, w)
				return
			}
		}

		// bulk insert the items into db
		rowsAffected, err := store.BulkInsert(r.Context(), inputs)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Import: store.BulkInsert failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   rowsAffected,
		}
		JsonResponse(resp, http.StatusCreated, w)
	}
}
