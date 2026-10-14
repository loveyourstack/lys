package lys

import (
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspg"
)

// Put handles changing an item using the supplied store.
func Put[idT lyspg.PrimaryKeyType, inputT any](env Env, store lyspg.Updater[idT, inputT]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get the id param and parse it into an idT
		id, err := getIdFromReq[idT](r)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Put: getIdFromReq failed: %w", err), env.Logger, w)
			return
		}

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Put: ExtractJsonBody failed: %w", err), env.Logger, w)
			return
		}

		// unmarshal the body
		input, err := DecodeJsonBody[inputT](body)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Put: DecodeJsonBody failed: %w", err), env.Logger, w)
			return
		}

		// validate item
		if err = store.Validate(env.Validate, input); err != nil {
			HandleUserError(lyserr.User{Message: err.Error(), StatusCode: http.StatusUnprocessableEntity}, w)
			return
		}

		// try to update the item in db
		err = store.Update(ctx, input, id)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Put: store.Update failed: %w", err), env.Logger, w)
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
