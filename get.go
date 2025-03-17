package lys

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/loveyourstack/lys/lyscsv"
	"github.com/loveyourstack/lys/lysexcel"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lystype"
)

// iGetable is a store that can be used by Get
type iGetable[T any] interface {
	GetMeta() lysmeta.Result
	GetName() string // file output: for setting filename
	Select(ctx context.Context, params lyspg.SelectParams) (items []T, unpagedCount lyspg.TotalCount, err error)
}

type GetOption[T any] struct {
	AdditionalFilterParamNames []string                                                                                                   // param names that are not part of the store's meta json tags, but should be allowed anyway. Must be handled by store's Select func
	GetLastSyncAt              func(ctx context.Context) (lastSyncAt lystype.Datetime, err error)                                         // for external data: func to get the last synced timestamp
	SelectFunc                 func(ctx context.Context, params lyspg.SelectParams) (items []T, unpagedCount lyspg.TotalCount, err error) // if passed, override the default Select() func
	SetFuncUrlParamNames       []string                                                                                                   // if selecting from a setFunc rather than a view: the names of the url params that must be passed and that will be passed, in order, to the setFunc
}

// Get handles retrieval of multiple items from the supplied store
func Get[T any](env Env, store iGetable[T], options ...GetOption[T]) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get additionalFilterParamNames from options if it was passed
		additionalFilterParamNames := []string{}
		for _, option := range options {
			if len(option.AdditionalFilterParamNames) > 0 {
				additionalFilterParamNames = option.AdditionalFilterParamNames
			}
		}

		// get setFuncUrlParamNames from options if it was passed
		setFuncUrlParamNames := []string{}
		for _, option := range options {
			if len(option.SetFuncUrlParamNames) > 0 {
				setFuncUrlParamNames = option.SetFuncUrlParamNames
			}
		}

		// get request modifiers from url params
		getReqModifiers, err := ExtractGetRequestModifiers(r, store.GetMeta().JsonTags, setFuncUrlParamNames, additionalFilterParamNames, env.GetOptions)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Get: ExtractGetRequestModifiers failed: %w", err), env.ErrorLog, w)
			return
		}

		// define params for store select func
		selectParams := lyspg.SelectParams{
			Conditions:         getReqModifiers.Conditions,
			Sorts:              getReqModifiers.Sorts,
			SetFuncParamValues: getReqModifiers.SetFuncParamValues,
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
			selectParams.Limit = env.GetOptions.MaxFileRecs
		}

		// default select func is Select() in store
		storeSelectFunc := store.Select

		// if a select func with the same signature was sent via options, use it
		for _, option := range options {
			if option.SelectFunc != nil {
				storeSelectFunc = option.SelectFunc
			}
		}

		// select items from db
		items, unpagedCount, err := storeSelectFunc(r.Context(), selectParams)
		if err != nil {

			HandleError(r.Context(), fmt.Errorf("Get: storeSelectFunc failed: %w", err), env.ErrorLog, w)
			return
		}

		// output the required format
		switch getReqModifiers.Format {

		case FormatCsv:

			// if no items, return empty json response
			if len(items) == 0 {
				resp := StdResponse{
					Status: ReqSucceeded,
					Data:   nil,
				}
				JsonResponse(resp, http.StatusOK, w)
				return
			}

			// create unique temp file
			f, err := os.CreateTemp("", store.GetName()+".*.csv")
			if err != nil {
				HandleInternalError(r.Context(), fmt.Errorf("Get: os.CreateTemp failed: %w", err), env.ErrorLog, w)
				return
			}
			f.Close()

			// write items to temp file
			err = lyscsv.WriteItemsToFile(items, store.GetMeta().JsonTagTypeMap, f.Name(), env.GetOptions.CsvDelimiter)
			if err != nil {
				HandleInternalError(r.Context(), fmt.Errorf("Get: lyscsv.WriteItemsToFile failed: %w", err), env.ErrorLog, w)
				return
			}

			// copy file into response then remove it
			FileResponse(f.Name(), store.GetName()+".csv", true, w)

		case FormatExcel:

			// if no items, return empty json response
			if len(items) == 0 {
				resp := StdResponse{
					Status: ReqSucceeded,
					Data:   nil,
				}
				JsonResponse(resp, http.StatusOK, w)
				return
			}

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

			// marshal items to json response
			resp := StdResponse{
				Status: ReqSucceeded,
				Data:   items,
				GetMetadata: &GetMetadata{
					Count:                 len(items),
					TotalCount:            unpagedCount.Value,
					TotalCountIsEstimated: unpagedCount.IsEstimated,
				},
			}

			// if GetLastSyncAt func was passed, call it and add timestamp to resp
			for _, option := range options {
				if option.GetLastSyncAt != nil {
					lastSyncAt, err := option.GetLastSyncAt(r.Context())
					if err != nil {
						HandleError(r.Context(), fmt.Errorf("Get: option.GetLastSyncAt failed: %w", err), env.ErrorLog, w)
						return
					}
					resp.LastSyncAt = &lastSyncAt
				}
			}

			JsonResponse(resp, http.StatusOK, w)

		default:
			// should never happen assuming format param gets checked
			HandleInternalError(r.Context(), fmt.Errorf("Get: unknown format: '%s'", getReqModifiers.Format), env.ErrorLog, w)
		}
	}
}

// iGetableWithLastSync is a store that can be used by GetWithLastSync
type iGetableWithLastSync[T any] interface {
	iGetable[T]
	GetLastSyncAt(ctx context.Context) (lastSyncAt lystype.Datetime, err error)
}

// GetWithLastSync is a wrapper for Get which adds the lastSyncAt timestamp from the supplied func to the JSON response
func GetWithLastSync[T any](env Env, store iGetableWithLastSync[T]) http.HandlerFunc {
	return Get(env, store, GetOption[T]{GetLastSyncAt: store.GetLastSyncAt})
}

// GetFunc is a wrapper for Get which allows passing an alternative Select func with the same signature
func GetFunc[T any](env Env, store iGetable[T], selectFunc func(ctx context.Context, params lyspg.SelectParams) (items []T, unpagedCount lyspg.TotalCount, err error)) http.HandlerFunc {
	return Get(env, store, GetOption[T]{SelectFunc: selectFunc})
}
