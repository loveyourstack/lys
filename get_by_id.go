package lys

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/lysmeta"
)

// iGetableById is a store that can be used by GetById
type iGetableById[T any] interface {
	GetMeta() lysmeta.Result
	SelectById(ctx context.Context, fields []string, id int64) (item T, err error)
}

// GetById handles retrieval of a single item from the supplied store using an integer id
func GetById[T any](env Env, store iGetableById[T]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get the Id and ensure it is an int
		vars := mux.Vars(r)
		id, err := strconv.ParseInt(vars["id"], 10, 64)
		if err != nil {
			HandleUserError(http.StatusBadRequest, ErrDescIdNotAnInteger, w)
			return
		}

		// get fields param if present (e.g. &xfields=)
		fields, err := ExtractFields(r, store.GetMeta().JsonTags, env.GetOptions.FieldsParamName)
		if err != nil {
			/*var userErr lyserr.User
			if errors.As(err, &userErr) {
				HandleUserError(http.StatusBadRequest, userErr.Message, w)
			} else {
				HandleInternalError(r.Context(), fmt.Errorf("GetById: ExtractFields failed: %w", err), env.ErrorLog, w)
			}*/
			HandleError(r.Context(), fmt.Errorf("GetById: ExtractFields failed: %w", err), env.ErrorLog, w)
			return
		}

		// select item from Db
		item, err := store.SelectById(r.Context(), fields, id)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("GetById: store.SelectById failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   item,
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}
