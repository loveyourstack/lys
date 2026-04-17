package lys

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/loveyourstack/lys/lyspg"
)

// Message returns the supplied msg in the Data field
func Message(msg string) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   msg,
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}

// NotFound provides a response informing the user that the requested route was not found
func NotFound() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		HandleUserError(ErrRouteNotFound, w)
	}
}

// PgSleep creates an artifical longrunning query in the db which can be viewed using pg_stat_activity.
// Pass cancelAfterSecs as 0 to not cancel the request.
// Used for testing context cancelation
func PgSleep(db lyspg.PoolOrTx, errorLog *slog.Logger, sleepSecs, cancelAfterSecs int) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		if cancelAfterSecs > 0 && cancelAfterSecs < sleepSecs {
			cancelCtx, cancel := context.WithTimeout(ctx, time.Duration(cancelAfterSecs)*time.Second)
			ctx = cancelCtx
			defer cancel()

			go func() {
				<-ctx.Done()
				errorLog.Info(fmt.Sprintf("canceling sleep after %d seconds", cancelAfterSecs))
			}()
		}

		err := lyspg.Sleep(ctx, db, sleepSecs)
		if err != nil {
			HandleError(ctx, err, errorLog, w)
			return
		}

		resp := StdResponse{
			Status: ReqSucceeded,
			Data:   "done",
		}
		JsonResponse(resp, http.StatusOK, w)
	}
}
