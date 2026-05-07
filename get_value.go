package lys

import (
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetValue handles retrieval of a single value returned by stmt
func GetValue[T any](env Env, db *pgxpool.Pool, stmt string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// select value from db
		rows, _ := db.Query(ctx, stmt)
		res, err := pgx.CollectExactlyOneRow(rows, pgx.RowTo[T])
		if err != nil {
			HandleError(ctx, fmt.Errorf("GetValue: pgx.CollectExactlyOneRow failed: %w", err), env.ErrorLog, w)
			return
		}

		// success
		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   res,
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}
