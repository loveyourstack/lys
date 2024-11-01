package lys

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspg"
)

// GetEnumValues returns enum values from the supplied schema and enum type name
func GetEnumValues(env Env, db *pgxpool.Pool, schema, enum string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get include or exclude filters, if any
		includeVals, excludeVals, err := extractEnumFilters(r.URL.Query(), env.GetOptions.SortParamName)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("GetEnumValues: extractEnumFilters failed: %w", err), env.ErrorLog, w)
			return
		}

		// get sort instruction, if any
		sortVal, err := extractEnumSort(r.URL.Query(), env.GetOptions.SortParamName)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("GetEnumValues: extractEnumSort failed: %w", err), env.ErrorLog, w)
			return
		}

		// select enum from db
		vals, err := lyspg.SelectEnum(r.Context(), db, schema+"."+enum, includeVals, excludeVals, sortVal)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("GetEnumValues: lyspg.SelectEnum failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   vals,
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}

// extractEnumFilters returns strings from the supplied Url values that should be included or excluded when selecting the enum
// inclusion: vals=x,y
// exclusion: vals=!x,y
func extractEnumFilters(urlValues url.Values, sortParamName string) (includeVals, excludeVals []string, err error) {

	// for each Url key
	for key, vals := range urlValues {

		// ignore sort param
		if key == sortParamName {
			continue
		}

		// the only key allowed is "vals"
		if key != "vals" {
			return nil, nil, lyserr.User{
				Message: "invalid enum filter field: " + key}
		}

		// for each Url value
		for _, csvVals := range vals {

			// if 1st char is "!", treat as exclusion
			if csvVals[:1] == "!" {
				excludeVals = append(excludeVals, strings.Split(csvVals[1:], ",")...)
			} else {
				includeVals = append(includeVals, strings.Split(csvVals, ",")...)
			}
		}
	}

	return includeVals, excludeVals, nil
}

// extractEnumSort returns an optional instruction about whether or not to sort the enum
// asc: xsort=val
// desc: xsort=-val
func extractEnumSort(urlValues url.Values, sortParamName string) (sortVal string, err error) {

	// for each Url key
	for key, vals := range urlValues {

		// ignore anything other than sort param
		if key != sortParamName {
			continue
		}

		// only consider first key
		sortVal = vals[0]

		switch sortVal {
		case "", "val", "-val": // legitimate values
			return sortVal, nil
		default:
			return "", lyserr.User{Message: "unknown enum sort value: " + sortVal}
		}
	}

	return "", nil
}
