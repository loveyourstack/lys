package lys

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/loveyourstack/lys/lyserr"
)

// iPostable is a store that can be used by Post
type iPostable[inputT any, itemT any] interface {
	Insert(ctx context.Context, input inputT) (item itemT, stmt string, err error)
	Validate(validate *validator.Validate, input inputT) error
}

// Post handles creating a new item using the supplied store and returning it in the response
func Post[inputT any, itemT any](env Env, store iPostable[inputT, itemT]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			if userErr, ok := err.(lyserr.User); ok {
				HandleUserError(http.StatusBadRequest, userErr.Message, w)
			} else {
				HandleInternalError(r.Context(), fmt.Errorf("Post: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			}
			return
		}

		// unmarshal the body
		input, err := DecodeJsonBody[inputT](body)
		if err != nil {
			if userErr, ok := err.(lyserr.User); ok {
				HandleUserError(http.StatusBadRequest, userErr.Message, w)
			} else {
				HandleInternalError(r.Context(), fmt.Errorf("Post: DecodeJsonBody failed: %w", err), env.ErrorLog, w)
			}
			return
		}

		// validate item
		if err = store.Validate(env.Validate, input); err != nil {
			HandleUserError(http.StatusUnprocessableEntity, err.Error(), w)
			return
		}

		// try to insert the item into db
		newItem, stmt, err := store.Insert(r.Context(), input)
		if err != nil {

			// expected error: request canceled
			if errors.Is(err, context.Canceled) {
				return
			}

			// handle errors from postgres
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				HandlePostgresError(r.Context(), stmt, "Post: store.Insert", pgErr, env.ErrorLog, w)
				return
			}

			// unknown db error
			HandleDbError(r.Context(), stmt, fmt.Errorf("Post: store.Insert failed: %w", err), env.ErrorLog, w)
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
