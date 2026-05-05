package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspg"
)

// iPutable is a store that can be used by Put.
type iPutable[idT lyspg.PrimaryKeyType, outT any] interface {
	Update(ctx context.Context, item outT, id idT) error
	Validate(validate *validator.Validate, item outT) error
}

// Put handles changing an item using the supplied store.
func Put[idT lyspg.PrimaryKeyType, outT any](env Env, store iPutable[idT, outT]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the id param
		idStr := mux.Vars(r)["id"]
		if idStr == "" {
			HandleUserError(ErrIdMissing, w)
			return
		}

		// parse the id param into a idT
		id, err := parseIdByType[idT](idStr)
		if err != nil {
			HandleUserError(ErrIdParseError, w)
			return
		}

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Put: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// unmarshal the body
		input, err := DecodeJsonBody[outT](body)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Put: DecodeJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// validate item
		if err = store.Validate(env.Validate, input); err != nil {
			HandleUserError(lyserr.User{Message: err.Error(), StatusCode: http.StatusUnprocessableEntity}, w)
			return
		}

		// try to update the item in db
		err = store.Update(r.Context(), input, id)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Put: store.Update failed: %w", err), env.ErrorLog, w)
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
