package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// iPostable is a store that can be used by Post
type iPostable[inputT any, pkT any] interface {
	Insert(ctx context.Context, input inputT) (newId pkT, err error)
	Validate(validate *validator.Validate, input inputT) error
}

// Post handles creating a new item using the supplied store and returning it in the response
func Post[inputT any, pkT any](env Env, store iPostable[inputT, pkT]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Post: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// unmarshal the body
		input, err := DecodeJsonBody[inputT](body)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Post: DecodeJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// validate item
		if err = store.Validate(env.Validate, input); err != nil {
			HandleUserError(http.StatusUnprocessableEntity, err.Error(), w)
			return
		}

		// try to insert the item into db
		newId, err := store.Insert(r.Context(), input)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Post: store.Insert failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   newId,
		}
		JsonResponse(resp, http.StatusCreated, w)
	}
}
