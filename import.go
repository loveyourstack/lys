package lys

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyserr"
)

// iImportable is a store that can be used by Import
type iImportable[T any] interface {
	BulkInsert(ctx context.Context, inputs []T) (rowsAffected int64, err error)
	Validate(validate *validator.Validate, input T) error
}

/*
ImportValueRepl is a struct that can be used when the input contains a foreign key.
It allows the referenced table's string representation to be passed by the user.
The string attribute gets replaced with the int64 attribute, and all the values are mapped using the supplied map. For example:

	StringJsonName: "car_manufacturer"
	Int64JsonName:  "car_manufacturer_fk"
	MapFunc:        returns Ford = 1, Nissan = 2, etc
*/
type ImportValueRepl struct {
	StringJsonName string
	Int64JsonName  string
	MapFunc        func(context.Context, *pgxpool.Pool) (map[string]int64, error)
}

// Import handles creating multiple new items using the supplied store and returning the number of rows inserted
// the supplied db is used for the MapFunc in valRepls
func Import[T any](env Env, db *pgxpool.Pool, store iImportable[T], valRepls ...ImportValueRepl) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Import: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// replace string values with int64 if needed
		if len(valRepls) > 0 {
			body, err = importReplaceValues(r.Context(), db, body, valRepls...)
			if err != nil {
				HandleError(r.Context(), fmt.Errorf("Import: importReplaceValues failed: %w", err), env.ErrorLog, w)
				return
			}
		}

		// unmarshal the body into a slice of inputs
		inputs, err := DecodeJsonBody[[]T](body)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Import: DecodeJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// check for empty input slice
		if len(inputs) == 0 {
			HandleUserError(lyserr.User{Message: "no inputs found", StatusCode: http.StatusUnprocessableEntity}, w)
			return
		}

		// validate each item
		for i, input := range inputs {
			if err = store.Validate(env.Validate, input); err != nil {
				HandleUserError(lyserr.User{Message: fmt.Sprintf("line %v: %s", i+1, err.Error()), StatusCode: http.StatusUnprocessableEntity}, w)
				return
			}
		}

		// bulk insert the items into db
		rowsAffected, err := store.BulkInsert(r.Context(), inputs)
		if err != nil {
			HandleError(r.Context(), fmt.Errorf("Import: store.BulkInsert failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   rowsAffected,
		}
		JsonResponse(resp, http.StatusCreated, w)
	}
}

func importReplaceValues(ctx context.Context, db *pgxpool.Pool, inBody []byte, valRepls ...ImportValueRepl) (outBody []byte, err error) {

	// unmarshal to []map[string]any so that keys can be processed
	dataA := []map[string]any{}
	err = json.Unmarshal(inBody, &dataA)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal failed: %w", err)
	}

	// for each value replacement
	for _, valRepl := range valRepls {

		// get map
		valMap, err := valRepl.MapFunc(ctx, db)
		if err != nil {
			return nil, fmt.Errorf("valRepl.MapFunc failed: %w", err)
		}

		// for each input
		for _, d := range dataA {

			// try to find the StringJsonName key: skip if not found
			strValAny, ok := d[valRepl.StringJsonName]
			if !ok {
				//fmt.Println(valRepl.StringJsonName, "not found")
				continue
			}

			// if found, it needs to be a string
			strVal, ok := strValAny.(string)
			if !ok {
				return nil, fmt.Errorf("key '%s': string assertion of value '%v' failed", valRepl.StringJsonName, strValAny)
			}

			// get the mapped int64
			int64Val, ok := valMap[strVal]
			if !ok {
				return nil, lyserr.User{Message: fmt.Sprintf("key '%s': no mapped value for '%s'", valRepl.StringJsonName, strVal)}
			}

			// delete the string key
			delete(d, valRepl.StringJsonName)

			// add the int64 key
			d[valRepl.Int64JsonName] = int64Val
		}
	}

	// marshal
	outBody, err = json.Marshal(dataA)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal failed: %w", err)
	}

	return outBody, nil
}
