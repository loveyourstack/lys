package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspg"
)

// iPutable is a store that can be used by Put.
type iPutable[idT lyspg.PrimaryKeyType, inputT any] interface {
	Update(ctx context.Context, input inputT, id idT) error
	Validate(input inputT) error
}

// Put handles changing an item using the supplied store.
func Put[idT lyspg.PrimaryKeyType, inputT any](env Env, store iPutable[idT, inputT]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

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
			HandleError(ctx, fmt.Errorf("Put: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// unmarshal the body
		input, err := DecodeJsonBody[inputT](body)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Put: DecodeJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// validate item
		if err = store.Validate(input); err != nil {
			HandleUserError(lyserr.User{Message: err.Error(), StatusCode: http.StatusUnprocessableEntity}, w)
			return
		}

		// try to update the item in db
		err = store.Update(ctx, input, id)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Put: store.Update failed: %w", err), env.ErrorLog, w)
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
