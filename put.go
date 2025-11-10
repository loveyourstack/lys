package lys

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lyserr"
)

// iPutable is a store that can be used by Put
type iPutable[T any] interface {
	Update(ctx context.Context, item T, id int64) error
	Validate(validate *validator.Validate, item T) error
}

// Put handles changing an item using the supplied store
func Put[T any](env Env, store iPutable[T]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the Id and ensure it is an int
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			HandleUserError(ErrIdNotAnInteger, w)
			return
		}

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Put: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// unmarshal the body
		input, err := DecodeJsonBody[T](body)
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
