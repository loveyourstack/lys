package lys

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/loveyourstack/lys/lyserr"
)

// iImportable is a store that can be used by Import
type iImportable[inputT any] interface {
	InsertTx(ctx context.Context, tx pgx.Tx, input inputT) (newId int64, err error)
	Validate(validate *validator.Validate, input inputT) error
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
func Import[inputT any](env Env, db *pgxpool.Pool, store iImportable[inputT], valRepls ...ImportValueRepl) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// get req body
		body, err := ExtractJsonBody(r, env.PostOptions.MaxBodySize)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Import: ExtractJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// replace string values with int64 if needed
		if len(valRepls) > 0 {
			body, err = importReplaceValues(ctx, db, body, valRepls...)
			if err != nil {
				HandleError(ctx, fmt.Errorf("Import: importReplaceValues failed: %w", err), env.ErrorLog, w)
				return
			}
		}

		// unmarshal the body into a slice of inputs
		inputs, err := DecodeJsonBody[[]inputT](body)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Import: DecodeJsonBody failed: %w", err), env.ErrorLog, w)
			return
		}

		// check for empty input slice
		if len(inputs) == 0 {
			HandleUserError(lyserr.User{Message: "no inputs found", StatusCode: http.StatusUnprocessableEntity}, w)
			return
		}

		if len(inputs) > env.PostOptions.MaxImportRecs {
			HandleUserError(lyserr.User{Message: fmt.Sprintf("found %v records; max allowed is %v", len(inputs), env.PostOptions.MaxImportRecs), StatusCode: http.StatusUnprocessableEntity}, w)
			return
		}

		// validate each item
		for i, input := range inputs {
			if err = store.Validate(env.Validate, input); err != nil {
				HandleUserError(lyserr.User{Message: fmt.Sprintf("line %v: %s", i+1, err.Error()), StatusCode: http.StatusUnprocessableEntity}, w)
				return
			}
		}

		// begin tx
		tx, err := db.Begin(ctx)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Import: db.Begin failed: %w", err), env.ErrorLog, w)
			return
		}
		defer tx.Rollback(ctx)

		// insert items as a tx
		for i, input := range inputs {
			if _, err := store.InsertTx(ctx, tx, input); err != nil {

				// if it was user-fixable db error, e.g. a unique constraint violation, show the line number to user
				dbErr := lyserr.Db{}
				if errors.As(err, &dbErr) {
					HandleDbError(ctx, i+1, dbErr.Stmt, fmt.Errorf("Import: store.InsertTx failed: %w", dbErr.Err), env.ErrorLog, w)
					return
				}

				HandleError(ctx, fmt.Errorf("Import: store.InsertTx failed on line %v: %w", i+1, err), env.ErrorLog, w)
				return
			}
		}

		// success: commit tx
		err = tx.Commit(ctx)
		if err != nil {
			HandleError(ctx, fmt.Errorf("Import: tx.Commit failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   len(inputs),
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
		for i, d := range dataA {

			// try to find the StringJsonName key: skip if not found
			strValAny, ok := d[valRepl.StringJsonName]
			if !ok {
				//fmt.Println(valRepl.StringJsonName, "not found")
				continue
			}

			// if found, it needs to be a string
			strVal, ok := strValAny.(string)
			if !ok {
				return nil, fmt.Errorf("line %v: key '%s': string assertion of value '%v' failed", i+1, valRepl.StringJsonName, strValAny)
			}

			// get the mapped int64
			int64Val, ok := valMap[strVal]
			if !ok {
				return nil, lyserr.User{Message: fmt.Sprintf("line %v: key '%s': no mapped value for '%s'", i+1, valRepl.StringJsonName, strVal)}
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
