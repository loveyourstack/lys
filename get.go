package lys

import (
	"context"
	"fmt"
	"mime"
	"net/http"

	"github.com/loveyourstack/lys/lyscsv"
	"github.com/loveyourstack/lys/lysexcel"
	"github.com/loveyourstack/lys/lysmeta"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lysset"
	"github.com/loveyourstack/lys/lysstring"
	"github.com/loveyourstack/lys/lystype"
)

// iGetable is a store that can be used by Get
type iGetable[T any] interface {
	GetName() string // file output: for setting filename
	GetPlan() lysmeta.Plan
	Select(ctx context.Context, params lyspg.SelectParams) (items []T, unpagedCount lyspg.TotalCount, err error)
}

type GetOpts[T any] struct {

	// AdditionalFilterParamNames are param names that are not in the store's db tags, but should be allowed anyway. Must be handled by the store's Select func.
	AdditionalFilterParamNames lysset.Set[string]

	// GetLastSyncAt gets the last synced timestamp for external data and returns it as a response header.
	GetLastSyncAt func(ctx context.Context) (lastSyncAt lystype.Datetime, err error)

	// SelectFunc, if passed, overrides the default store Select() func.
	SelectFunc func(ctx context.Context, params lyspg.SelectParams) (items []T, unpagedCount lyspg.TotalCount, err error)

	// SetFuncUrlParamNames are used if selecting from a setFunc rather than a view. They are the names of the url params that will be passed, in order, to the setFunc.
	// Don't use a Set: order must be preserved.
	SetFuncUrlParamNames []string
}

// Get handles retrieval of multiple items from the supplied store
func Get[T any](env Env, store iGetable[T], opts *GetOpts[T]) http.HandlerFunc {

	// preprocess options and get store vars before returning handler func to avoid doing this on every request

	// set option defaults
	additionalFilterParamNames := lysset.New[string]()
	var getLastSyncAt func(ctx context.Context) (lastSyncAt lystype.Datetime, err error) = nil
	storeSelectFunc := store.Select
	setFuncUrlParamNames := []string{}

	// override defaults with any supplied options
	if opts != nil {
		if opts.AdditionalFilterParamNames.Len() > 0 {
			additionalFilterParamNames = opts.AdditionalFilterParamNames
		}
		if opts.GetLastSyncAt != nil {
			getLastSyncAt = opts.GetLastSyncAt
		}
		if opts.SelectFunc != nil {
			storeSelectFunc = opts.SelectFunc
		}
		if len(opts.SetFuncUrlParamNames) > 0 {
			setFuncUrlParamNames = opts.SetFuncUrlParamNames
		}
	}

	// get store vars
	plan := store.GetPlan()
	storeName := store.GetName()

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get request modifiers from url params
		getReqModifiers, err := ExtractGetRequestModifiers(r,
			ExtractGetRequestModifierParams{
				AdditionalFilterParamNames: additionalFilterParamNames,
				DbNames:                    lysset.FromSlice(plan.DbNames()),
				GetOptions:                 env.GetOptions,
				JsonKeyDbNameMap:           plan.JsonKeyDbNameMap(),
				SetFuncUrlParamNames:       setFuncUrlParamNames,
			})
		if err != nil {
			HandleError(ctx, fmt.Errorf("Get: ExtractGetRequestModifiers failed: %w", err), env.Logger, w)
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

		// select items from db
		items, unpagedCount, err := storeSelectFunc(ctx, selectParams)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Get: storeSelectFunc failed: %w", err), env.Logger, w)
			return
		}

		// if GetLastSyncAt func was passed, call it and add timestamp to resp headers
		if getLastSyncAt != nil {
			lastSyncAt, err := getLastSyncAt(ctx)
			if err != nil {
				// log error but don't fail the request
				env.Logger.Error("Get: getLastSyncAt failed", "error", err)
			} else {
				w.Header().Set("LastSyncAt", lastSyncAt.Format(lystype.DatetimeFormat))
			}
		}

		// output the required format
		switch getReqModifiers.Format {

		case FormatCsv:

			// set file download headers
			w.Header().Set("Content-Type", "text/csv")
			w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
				"filename": lysstring.SafeFileName(storeName, ".csv"),
			}))

			// stream csv to response writer
			err = lyscsv.WriteItems(items, plan.JsonKeyTypeMap(), env.GetOptions.CsvDelimiter, w)
			if err != nil {
				HandleInternalError(ctx, fmt.Errorf("Get: lyscsv.WriteItems failed: %w", err), env.Logger, w)
				return
			}

		case FormatExcel:

			// set file download headers
			w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
			w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
				"filename": lysstring.SafeFileName(storeName, ".xlsx"),
			}))

			// stream Excel to response writer
			err = lysexcel.WriteItems(items, plan.JsonKeyTypeMap(), "", w)
			if err != nil {
				HandleInternalError(ctx, fmt.Errorf("Get: lysexcel.WriteItems failed: %w", err), env.Logger, w)
				return
			}

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
			JsonResponse(resp, http.StatusOK, w)

		default:
			// should never happen assuming format param gets checked
			HandleInternalError(ctx, fmt.Errorf("Get: unknown format: '%s'", getReqModifiers.Format), env.Logger, w)
		}
	}
}

// iGetableWithLastSync is a store that can be used by GetWithLastSync.
type iGetableWithLastSync[T any] interface {
	iGetable[T]
	GetLastSyncAt(ctx context.Context) (lastSyncAt lystype.Datetime, err error)
}

// GetWithLastSync is a wrapper for Get which adds the lastSyncAt timestamp from the supplied func to the response headers.
func GetWithLastSync[T any](env Env, store iGetableWithLastSync[T]) http.HandlerFunc {
	return Get(env, store, &GetOpts[T]{GetLastSyncAt: store.GetLastSyncAt})
}

// GetFunc is a wrapper for Get which allows passing an alternative Select func with the same signature.
func GetFunc[T any](env Env, store iGetable[T], selectFunc func(ctx context.Context, params lyspg.SelectParams) (items []T, unpagedCount lyspg.TotalCount, err error)) http.HandlerFunc {
	return Get(env, store, &GetOpts[T]{SelectFunc: selectFunc})
}
