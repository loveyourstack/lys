package lys

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspg"
)

// iPutableById is a store that can be used by PutById.
type iPutableById[T any] interface {
	UpdateById(ctx context.Context, item T, id int64) error
	Validate(validate *validator.Validate, item T) error
}

// iPutableByUuid is a store that can be used by PutByUuid.
type iPutableByUuid[T any] interface {
	UpdateByUuid(ctx context.Context, item T, id uuid.UUID) error
	Validate(validate *validator.Validate, item T) error
}

// PutById handles changing an item using the supplied store and an int64 id.
func PutById[T any](env Env, store iPutableById[T]) http.HandlerFunc {
	return put(env, store.UpdateById, store.Validate, parseIdFunc, "PutById")
}

// PutByUuid handles changing an item using the supplied store and a UUID.
func PutByUuid[T any](env Env, store iPutableByUuid[T]) http.HandlerFunc {
	return put(env, store.UpdateByUuid, store.Validate, parseUuidFunc, "PutByUuid")
}

// put handles changing an item using the supplied store.
func put[idT lyspg.PrimaryKeyType, outT any](env Env, putFunc func(context.Context, outT, idT) error,
	validateFunc func(validate *validator.Validate, item outT) error, parseIdFunc func(string) (idT, error),
	callingFunc string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the id param
		idStr := mux.Vars(r)["id"]
		if idStr == "" {
			HandleUserError(ErrIdMissing, w)
			return
		}

		// parse the id param into a idT
		id, err := parseIdFunc(idStr)
		if err != nil {
			HandleUserError(ErrIdParseError, w)
			return
		}

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("%s: ExtractJsonBody failed: %w", callingFunc, err), env.ErrorLog, w)
			return
		}

		// unmarshal the body
		input, err := DecodeJsonBody[outT](body)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("%s: DecodeJsonBody failed: %w", callingFunc, err), env.ErrorLog, w)
			return
		}

		// validate item
		if err = validateFunc(env.Validate, input); err != nil {
			HandleUserError(lyserr.User{Message: err.Error(), StatusCode: http.StatusUnprocessableEntity}, w)
			return
		}

		// try to update the item in db
		err = putFunc(r.Context(), input, id)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("%s: putFunc failed: %w", callingFunc, err), env.ErrorLog, w)
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
