package lys

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lysexcel"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
)

// iGetable is a store that can be used by Get
type iGetable[T any] interface {
	GetMeta() lysmeta.Result
	GetName() string // file output: for setting filename
	Select(ctx context.Context, params lyspg.SelectParams) (items []T, unpagedCount lyspg.TotalCount, stmt string, err error)
}

// Get handles retrieval of multiple items from the supplied store
func Get[T any](env Env, store iGetable[T]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get request modifiers from url params
		getReqModifiers, err := ExtractGetRequestModifiers(r, store.GetMeta().JsonTags, env.GetOptions)
		if err != nil {
			if userErr, ok := err.(lyserr.User); ok {
				HandleUserError(http.StatusBadRequest, userErr.Message, w)
			} else {
				HandleInternalError(r.Context(), fmt.Errorf("Get: ExtractGetRequestModifiers failed: %w", err), env.ErrorLog, w)
			}
			return
		}

		// define params for store select func
		selectParams := lyspg.SelectParams{
			Conditions: getReqModifiers.Conditions,
			Sorts:      getReqModifiers.Sorts,
		}

		// if returning json, use fields and paging
		if getReqModifiers.Format == FormatJson {

			selectParams.Fields = getReqModifiers.Fields

			// get offset from paging params (starts at 0, not 1)
			offset := getReqModifiers.PerPage * (getReqModifiers.Page - 1)

			selectParams.Limit = getReqModifiers.PerPage
			selectParams.Offset = offset
			selectParams.GetUnpagedCount = true

		} else {
			// returning file: set max number of records
			selectParams.Limit = env.GetOptions.DefaultMaxFileRecs
		}

		// select items from db
		items, unpagedCount, stmt, err := store.Select(r.Context(), selectParams)
		if err != nil {

			// expected error: request canceled
			if errors.Is(err, context.Canceled) {
				return
			}

			// unknown db error
			HandleDbError(r.Context(), stmt, fmt.Errorf("Get: store.Select failed: %w", err), env.ErrorLog, w)
			return
		}

		// output the required format
		switch getReqModifiers.Format {

		case FormatExcel:

			// create unique temp file
			f, err := os.CreateTemp("", store.GetName()+".*.xlsx")
			if err != nil {
				HandleInternalError(r.Context(), fmt.Errorf("Get: os.CreateTemp failed: %w", err), env.ErrorLog, w)
				return
			}
			f.Close()

			// write items to temp file as Excel workbook
			err = lysexcel.WriteItemsToFile(items, store.GetMeta().JsonTagTypeMap, f.Name(), "")
			if err != nil {
				HandleInternalError(r.Context(), fmt.Errorf("Get: lysexcel.WriteItemsToFile failed: %w", err), env.ErrorLog, w)
				return
			}

			// copy file into response then remove it
			FileResponse(f.Name(), store.GetName()+".xlsx", true, w)

		case FormatJson:

			// add unpagedCount as header
			headers := []RespHeader{}
			headers = append(headers, RespHeader{Key: "X-Total-Count", Value: strconv.FormatInt(unpagedCount.Value, 10)})
			headers = append(headers, RespHeader{Key: "X-Total-Count-Estimated", Value: strconv.FormatBool(unpagedCount.IsEstimated)})

			// marshal items to json response
			resp := StdResponse{
				Status: ReqSucceeded,
				Data:   items,
			}
			JsonResponse(resp, http.StatusOK, headers, w)

		default:
			// should never happen assuming format param gets checked
			HandleInternalError(r.Context(), fmt.Errorf("Get: unknown format: '%s'", getReqModifiers.Format), env.ErrorLog, w)
		}
	}
}
