package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

// iPostable is a store that can be used by Post
type iPostable[inputT any, itemT any] interface {
	Insert(ctx context.Context, input inputT) (item itemT, err error)
	Validate(validate *validator.Validate, input inputT) error
}

// Post handles creating a new item using the supplied store and returning it in the response
func Post[inputT any, itemT any](env Env, store iPostable[inputT, itemT]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			/*var userErr lyserr.User
			if errors.As(err, &userErr) {
				HandleUserError(http.StatusBadRequest, userErr.Message, w)
			} else {
				HandleInternalError(r.Context(), fmt.Errorf("Post: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			}*/
			HandleError(r.Context(), fmt.Errorf("Post: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// unmarshal the body
		input, err := DecodeJsonBody[inputT](body)
		if err != nil {
			/*var userErr lyserr.User
			if errors.As(err, &userErr) {
				HandleUserError(http.StatusBadRequest, userErr.Message, w)
			} else {
				HandleInternalError(r.Context(), fmt.Errorf("Post: DecodeJsonBody failed: %w", err), env.ErrorLog, w)
			}*/
			HandleError(r.Context(), fmt.Errorf("Post: DecodeJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// validate item
		if err = store.Validate(env.Validate, input); err != nil {
			HandleUserError(http.StatusUnprocessableEntity, err.Error(), w)
			return
		}

		// try to insert the item into db
		newItem, err := store.Insert(r.Context(), input)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Post: store.Insert failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   newItem,
		}
		JsonResponse(resp, http.StatusCreated, w)
	}
}
