package lys

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/loveyourstack/lys/internal/cmd"
	"github.com/loveyourstack/lys/internal/lyscmd"
	"github.com/loveyourstack/lys/internal/stores/core/coreparamtest"
	"github.com/loveyourstack/lys/internal/stores/core/coresoftdeletetest"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetest"
	"github.com/loveyourstack/lys/internal/stores/core/corevolumetest"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lyspgdb"
	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
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

	endpoint := "/param-test"

	paramTestStore := coreparamtest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get[coreparamtest.Model](apiEnv, paramTestStore)).Methods("GET")

	endpoint = "/soft-delete-test"
	sdTestStore := coresoftdeletetest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get[coresoftdeletetest.Model](apiEnv, sdTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", GetById[coresoftdeletetest.Model](apiEnv, sdTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}/restore", Restore(apiEnv, srvApp.Db, sdTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}/soft", SoftDelete(apiEnv, srvApp.Db, sdTestStore)).Methods("DELETE")

	endpoint = "/type-test"

	typeTestStore := coretypetest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get[coretypetest.Model](apiEnv, typeTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", GetById[coretypetest.Model](apiEnv, typeTestStore)).Methods("GET")
	r.HandleFunc(endpoint, Post[coretypetest.Input, coretypetest.Model](apiEnv, typeTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}", Put[coretypetest.Input](apiEnv, typeTestStore)).Methods("PUT")
	r.HandleFunc(endpoint+"/{id}", Patch(apiEnv, typeTestStore)).Methods("PATCH")
	r.HandleFunc(endpoint+"/{id}", Delete(apiEnv, typeTestStore)).Methods("DELETE")

	endpoint = "/type-test-uuid"

	r.HandleFunc(endpoint+"/{id}", GetByUuid[coretypetest.Model](apiEnv, typeTestStore)).Methods("GET")

	endpoint = "/volume-test"

	volTestStore := corevolumetest.Store{Db: srvApp.Db}
	r.HandleFunc(endpoint, Get[corevolumetest.Model](apiEnv, volTestStore)).Methods("GET")
	r.HandleFunc(endpoint+"/any-10", GetSimple(apiEnv, volTestStore.Select10)).Methods("GET")
	r.HandleFunc(endpoint+"/{id}", GetById[corevolumetest.Model](apiEnv, volTestStore)).Methods("GET")
	r.HandleFunc(endpoint, Post[corevolumetest.Input, corevolumetest.Model](apiEnv, volTestStore)).Methods("POST")
	r.HandleFunc(endpoint+"/{id}", Put[corevolumetest.Input](apiEnv, volTestStore)).Methods("PUT")
	r.HandleFunc(endpoint+"/{id}", Patch(apiEnv, volTestStore)).Methods("PATCH")
	r.HandleFunc(endpoint+"/{id}", Delete(apiEnv, volTestStore)).Methods("DELETE")

	endpoint = "/weekdays"
	r.HandleFunc(endpoint, GetEnumValues(apiEnv, srvApp.Db, "core", "weekday")).Methods("GET")

	return r
}

func mustGetConfig(t testing.TB) lyscmd.Config {
	conf := lyscmd.Config{}
	err := conf.LoadFromFile("lys_config.toml")
	if err != nil {
		t.Fatalf("lys_config.toml not found: %v", err)
	}
	return conf
}

func mustGetSrvApp(t testing.TB, ctx context.Context) *httpServerApplication {

	conf := mustGetConfig(t)

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

func mustParseTime(t testing.TB, layout, value string) time.Time {

	timeV, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("time.Parse failed: %v", err)
	}

	return timeV
}

func testFilledInput(t testing.TB, item coretypetest.Input) {

	// boolean
	assert.EqualValues(t, false, item.CBool, "CBool")
	assert.EqualValues(t, true, *item.CBoolN, "CBoolN")
	expectedCBoolA := []bool{false, true}
	for i := range item.CBoolA {
		assert.EqualValues(t, expectedCBoolA[i], item.CBoolA[i], "CBoolA", i)
	}

	// int
	assert.EqualValues(t, int64(1), item.CInt, "CInt")
	assert.EqualValues(t, int64(2), *item.CIntN, "CIntN")
	expectedCIntA := []int64{1, 2}
	for i := range item.CIntA {
		assert.EqualValues(t, expectedCIntA[i], item.CIntA[i], "CIntA", i)
	}

	// double
	assert.EqualValues(t, float32(1.1), item.CDouble, "CDouble")
	assert.EqualValues(t, float32(2.1), *item.CDoubleN, "CDoubleN")
	expectedCDoubleA := []float32{1.1, 2.1}
	for i := range item.CDoubleA {
		assert.EqualValues(t, expectedCDoubleA[i], item.CDoubleA[i], "CDoubleA", i)
	}

	// numeric
	assert.EqualValues(t, float32(1.11), item.CNumeric, "CNumeric")
	assert.EqualValues(t, float32(2.11), *item.CNumericN, "CNumericN")
	expectedCNumericA := []float32{1.11, 2.11}
	for i := range item.CNumericA {
		assert.EqualValues(t, expectedCNumericA[i], item.CNumericA[i], "CNumericA", i)
	}

	// date
	d1 := mustParseTime(t, lystype.DateFormat, "2001-02-03")
	d2 := mustParseTime(t, lystype.DateFormat, "2002-03-04")
	assert.EqualValues(t, lystype.Date(d1), item.CDate, "CDate")
	assert.EqualValues(t, lystype.Date(d2), *item.CDateN, "CDateN")
	// TODO item.CDateA

	// time
	t1 := mustParseTime(t, lystype.TimeFormat, "12:01")
	t2 := mustParseTime(t, lystype.TimeFormat, "12:02")
	assert.EqualValues(t, lystype.Time(t1), item.CTime, "CTime")
	assert.EqualValues(t, lystype.Time(t2), *item.CTimeN, "CTimeN")
	// TODO item.CTimeA

	// datetime
	dt1 := mustParseTime(t, lystype.DatetimeFormat, "2001-02-03 12:01:00+01")
	dt2 := mustParseTime(t, lystype.DatetimeFormat, "2002-03-04 12:02:00+01")
	assert.EqualValues(t, lystype.Datetime(dt1), item.CDatetime, "CDatetime")
	assert.EqualValues(t, lystype.Datetime(dt2), *item.CDatetimeN, "CDatetimeN")
	// TODO item.CDatetimeA

	// enum
	assert.EqualValues(t, "Monday", item.CEnum, "CEnum")
	assert.EqualValues(t, "Tuesday", *item.CEnumN, "CEnumN")
	expectedCEnumA := []string{"Monday", "Tuesday"}
	for i := range item.CEnumA {
		assert.EqualValues(t, expectedCEnumA[i], item.CEnumA[i], "CEnumA", i)
	}

	// text
	assert.EqualValues(t, "a b", item.CText, "CText")
	assert.EqualValues(t, "b c", *item.CTextN, "CTextN")
	expectedCTextA := []string{"a b", "b c"}
	for i := range item.CTextA {
		assert.EqualValues(t, expectedCTextA[i], item.CTextA[i], "CTextA", i)
	}
}
