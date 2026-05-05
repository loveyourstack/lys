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
	"github.com/loveyourstack/lys/internal/myapp"
	"github.com/loveyourstack/lys/internal/stores/core/corearchivetest"
	"github.com/loveyourstack/lys/internal/stores/core/corearchivetestuuid"
	"github.com/loveyourstack/lys/internal/stores/core/coreimporttest"
	"github.com/loveyourstack/lys/internal/stores/core/coreparamtest"
	"github.com/loveyourstack/lys/internal/stores/core/coresetfunc"
	"github.com/loveyourstack/lys/internal/stores/core/coretagtest"
	"github.com/loveyourstack/lys/internal/stores/core/coretrackingtest"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetest"
	"github.com/loveyourstack/lys/internal/stores/core/coreuuidtest"
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
	r.HandleFunc(endpoint, Get(apiEnv, archiveTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", GetById(apiEnv, archiveTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}/restore", Restore(apiEnv, srvApp.Db, archiveTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}/archive", Archive(apiEnv, srvApp.Db, archiveTestStore)).Methods("DELETE")

	endpoint = "/archive-test-uuid"

	archiveTestUuidStore := corearchivetestuuid.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint+"/{id}", GetById(apiEnv, archiveTestUuidStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}/restore", Restore(apiEnv, srvApp.Db, archiveTestUuidStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}/archive", Archive(apiEnv, srvApp.Db, archiveTestUuidStore)).Methods("DELETE")

	endpoint = "/import-test"

	importTestStore := coreimporttest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint+"/import", Import(apiEnv, srvApp.Db, importTestStore)).Methods("POST")

	endpoint = "/param-test"

	paramTestStore := coreparamtest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get(apiEnv, paramTestStore)).Methods("GET")

	endpoint = "/process-slice-test"

	processSliceFunc := func(ctx context.Context, vals []int) (int64, error) {
		return int64(len(vals)), nil
	}
	r.HandleFunc(endpoint, ProcessSlice(apiEnv, processSliceFunc)).Methods("POST")

	endpoint = "/setfunc-test"

	setfuncTestStore := coresetfunc.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get(apiEnv, setfuncTestStore, GetOption[coresetfunc.Model]{
		SetFuncUrlParamNames: setfuncTestStore.GetSetFuncUrlParamNames(),
	})).Methods("GET")

	endpoint = "/tag-test"

	tagTestStore := coretagtest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get(apiEnv, tagTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", GetById(apiEnv, tagTestStore)).Methods("GET")
	r.HandleFunc(endpoint, Post(apiEnv, tagTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}", Put(apiEnv, tagTestStore)).Methods("PUT")
	r.HandleFunc(endpoint+"/{id}", Patch(apiEnv, tagTestStore)).Methods("PATCH")
	r.HandleFunc(endpoint+"/{id}", Delete(apiEnv, tagTestStore)).Methods("DELETE")

	endpoint = "/tracking-test"

	trackingTestStore := coretrackingtest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get(apiEnv, trackingTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", GetById(apiEnv, trackingTestStore)).Methods("GET")
	r.HandleFunc(endpoint, Post(apiEnv, trackingTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}", Put(apiEnv, trackingTestStore)).Methods("PUT")
	r.HandleFunc(endpoint+"/{id}", Patch(apiEnv, trackingTestStore)).Methods("PATCH")
	r.HandleFunc(endpoint+"/{id}", Delete(apiEnv, trackingTestStore)).Methods("DELETE")

	endpoint = "/type-test"

	typeTestStore := coretypetest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get(apiEnv, typeTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", GetById(apiEnv, typeTestStore)).Methods("GET")
	r.HandleFunc(endpoint, Post(apiEnv, typeTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}", Put(apiEnv, typeTestStore)).Methods("PUT")
	r.HandleFunc(endpoint+"/{id}", Patch(apiEnv, typeTestStore)).Methods("PATCH")
	r.HandleFunc(endpoint+"/{id}", Delete(apiEnv, typeTestStore)).Methods("DELETE")

	endpoint = "/uuid-test"

	uuidTestStore := coreuuidtest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get(apiEnv, uuidTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", GetById(apiEnv, uuidTestStore)).Methods("GET")
	r.HandleFunc(endpoint, Post(apiEnv, uuidTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}", Put(apiEnv, uuidTestStore)).Methods("PUT")
	r.HandleFunc(endpoint+"/{id}", Patch(apiEnv, uuidTestStore)).Methods("PATCH")
	r.HandleFunc(endpoint+"/{id}", Delete(apiEnv, uuidTestStore)).Methods("DELETE")

	endpoint = "/volume-test"

	volTestStore := corevolumetest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get(apiEnv, volTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/any-10", GetSimple(apiEnv, volTestStore.Select10)).Methods("GET")
	r.HandleFunc(endpoint+"/int-1", GetValue[int](apiEnv, srvApp.Db, "SELECT 1;")).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", GetById(apiEnv, volTestStore)).Methods("GET")
	r.HandleFunc(endpoint, Post(apiEnv, volTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}", Put(apiEnv, volTestStore)).Methods("PUT")
	r.HandleFunc(endpoint+"/{id}", Patch(apiEnv, volTestStore)).Methods("PATCH")
	r.HandleFunc(endpoint+"/{id}", Delete(apiEnv, volTestStore)).Methods("DELETE")

	endpoint = "/weekdays"
	r.HandleFunc(endpoint, GetEnumValues(apiEnv, srvApp.Db, "core", "weekday")).Methods("GET")

	return r
}

func mustGetSrvApp(ctx context.Context, t testing.TB) *httpServerApplication {

	conf := myapp.MustGetConfig(t)

	app := &cmd.Application{
		Config:   &conf,
		InfoLog:  slog.New(slog.NewTextHandler(os.Stdout, nil)),
		ErrorLog: slog.New(slog.NewTextHandler(os.Stderr, nil)),
		Validate: validator.New(validator.WithRequiredStructEnabled()),
	}

	// create http server app
	srvApp := &httpServerApplication{
		Application: app,
		PostOptions: FillPostOptions(PostOptions{}), // use defaults
	}

	var err error
	srvApp.GetOptions, err = FillGetOptions(GetOptions{})
	if err != nil {
		t.Fatalf("FillGetOptions failed: %v", err)
	}

	// register core.weekday type in any conn added to the pool so that Patch of type_test core.weekday[] works. If don't do this: "encode plan not found"
	dataTypeNames := []string{
		"core.weekday",
		"core.weekday[]",
	}
	srvApp.Db, err = lyspgdb.GetPoolWithTypes(ctx, conf.Db, conf.DbOwnerUser, "test", dataTypeNames)
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

func mustFillGetOptions(t testing.TB, input GetOptions) GetOptions {

	getOptions, err := FillGetOptions(input)
	if err != nil {
		t.Fatalf("FillGetOptions failed: %v", err)
	}

	return getOptions
}

func mustExtractFilters(t testing.TB, urlValues url.Values, jsonKeyDbNameMap map[string]string, additionalFilterParamNames, setFuncUrlParamNames []string, getOptions GetOptions) []lyspg.Condition {

	conds, err := ExtractFilters(urlValues, jsonKeyDbNameMap, additionalFilterParamNames, setFuncUrlParamNames, getOptions)
	if err != nil {
		t.Fatalf("ExtractFilters failed: %v", err)
	}

	return conds
}
