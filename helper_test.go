package lys

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/internal/cmd"
	"github.com/loveyourstack/lys/internal/lyscmd"
	"github.com/loveyourstack/lys/internal/stores/core/corearchivetest"
	"github.com/loveyourstack/lys/internal/stores/core/coreparamtest"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetest"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/internal/stores/core/corevolumetest"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lyspgdb"
)

type httpServerApplication struct {
	*cmd.Application
	GetOptions  GetOptions
	PostOptions PostOptions
}

// getRouter returns a mux providing the HTTP server's routes
func (srvApp *httpServerApplication) getRouter() http.Handler {

	// define env struct needed for lys route handlers
	apiEnv := Env{
		ErrorLog:    srvApp.ErrorLog,
		Validate:    srvApp.Validate,
		GetOptions:  srvApp.GetOptions,
		PostOptions: srvApp.PostOptions,
	}

	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(NotFound())

	endpoint := "/archive-test"

	archiveTestStore := corearchivetest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get[corearchivetest.Model](apiEnv, archiveTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", GetById[corearchivetest.Model](apiEnv, archiveTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}/restore", RestoreById(apiEnv, srvApp.Db, archiveTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}/archive", ArchiveById(apiEnv, srvApp.Db, archiveTestStore)).Methods("DELETE")

	endpoint = "/archive-test-uuid"

	r.HandleFunc(endpoint+"/{id}", GetByUuid[corearchivetest.Model](apiEnv, archiveTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}/restore", RestoreByUuid(apiEnv, srvApp.Db, archiveTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}/archive", ArchiveByUuid(apiEnv, srvApp.Db, archiveTestStore)).Methods("DELETE")

	endpoint = "/param-test"

	paramTestStore := coreparamtest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get[coreparamtest.Model](apiEnv, paramTestStore)).Methods("GET")

	endpoint = "/process-slice-test"

	processSliceFunc := func(ctx context.Context, vals []int) (int64, error) {
		return int64(len(vals)), nil
	}
	r.HandleFunc(endpoint, ProcessSlice(apiEnv, processSliceFunc)).Methods("POST")

	endpoint = "/type-test"

	typeTestStore := coretypetest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get[coretypetestm.Model](apiEnv, typeTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", GetById[coretypetestm.Model](apiEnv, typeTestStore)).Methods("GET")
	r.HandleFunc(endpoint, Post[coretypetestm.Input, int64](apiEnv, typeTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}", Put[coretypetestm.Input](apiEnv, typeTestStore)).Methods("PUT")
	r.HandleFunc(endpoint+"/{id}", Patch(apiEnv, typeTestStore)).Methods("PATCH")
	r.HandleFunc(endpoint+"/{id}", Delete(apiEnv, typeTestStore)).Methods("DELETE")

	endpoint = "/type-test-uuid"

	r.HandleFunc(endpoint+"/{id}", GetByUuid[coretypetestm.Model](apiEnv, typeTestStore)).Methods("GET")

	endpoint = "/volume-test"

	volTestStore := corevolumetest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get[corevolumetest.Model](apiEnv, volTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/any-10", GetSimple(apiEnv, volTestStore.Select10)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", GetById[corevolumetest.Model](apiEnv, volTestStore)).Methods("GET")
	r.HandleFunc(endpoint, Post[corevolumetest.Input, int64](apiEnv, volTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}", Put[corevolumetest.Input](apiEnv, volTestStore)).Methods("PUT")
	r.HandleFunc(endpoint+"/{id}", Patch(apiEnv, volTestStore)).Methods("PATCH")
	r.HandleFunc(endpoint+"/{id}", Delete(apiEnv, volTestStore)).Methods("DELETE")

	endpoint = "/weekdays"
	r.HandleFunc(endpoint, GetEnumValues(apiEnv, srvApp.Db, "core", "weekday")).Methods("GET")

	return r
}

func mustGetSrvApp(t testing.TB, ctx context.Context) *httpServerApplication {

	conf := lyscmd.MustGetConfig(t)

	app := &cmd.Application{
		Config:   &conf,
		InfoLog:  slog.New(slog.NewTextHandler(os.Stdout, nil)),
		ErrorLog: slog.New(slog.NewTextHandler(os.Stderr, nil)),
		Validate: validator.New(validator.WithRequiredStructEnabled()),
	}

	// create http server app
	srvApp := &httpServerApplication{
		Application: app,
		GetOptions:  FillGetOptions(GetOptions{}),   // use defaults
		PostOptions: FillPostOptions(PostOptions{}), // use defaults
	}

	var err error
	// register core.weekday type in any conn added to the pool so that Patch of type_test core.weekday[] works. If don't do this: "encode plan not found"
	dataTypeNames := []string{
		"core.weekday",
		"core.weekday[]",
	}
	srvApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbOwnerUser, dataTypeNames)
	if err != nil {
		t.Fatalf("lyspgdb.GetPoolWithTypes failed: %v", err)
	}

	return srvApp
}

func mustCreateGetReq(t testing.TB, targetUrl string) *http.Request {

	req, err := http.NewRequest("GET", targetUrl, nil)
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return req
}

func mustExtractFilters(t testing.TB, urlValues url.Values, validJsonFields []string, getOptions GetOptions) []lyspg.Condition {

	conds, err := ExtractFilters(urlValues, validJsonFields, getOptions)
	if err != nil {
		t.Fatalf("ExtractFilters failed: %v", err)
	}

	return conds
}
