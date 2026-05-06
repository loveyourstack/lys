package lys

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/internal/stores/core/coretagtest"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSuccess(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// basic test only: sorting, filtering etc is tested below
	targetUrl := "/param-test"
	resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, false, resp.GetMetadata.TotalCountIsEstimated)
	assert.EqualValues(t, 2, resp.GetMetadata.TotalCount)
	assert.EqualValues(t, 2, resp.GetMetadata.Count)
	assert.EqualValues(t, 2, len(resp.Data))
}

func TestGetFailure(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// invalid url
	targetUrl := "/xx"
	_, err := lysclient.GetItemRespTester(ctx, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "route not found", err.Error())

	// invalid filter field
	targetUrl = "/param-test?a=b"
	_, err = lysclient.GetItemRespTester(ctx, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "invalid filter field: a", err.Error())

	// TODO: further param tests on those functions directly
}

func TestGetSuccessFields(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// id
	targetUrl := "/param-test?xfields=id"
	resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	item := resp.Data[0]

	// present
	_, ok := item["id"]
	if !ok {
		t.Errorf("id: expected in response, but missing")
	}

	// omitempty
	_, ok = item["c_int"]
	if ok {
		t.Errorf("c_int: not expected in response, but present")
	}

	// omitzero
	_, ok = item["c_date"]
	if ok {
		t.Errorf("c_date: not expected in response, but present")
	}
	_, ok = item["c_time"]
	if ok {
		t.Errorf("c_time: not expected in response, but present")
	}
	_, ok = item["c_datetime"]
	if ok {
		t.Errorf("c_datetime: not expected in response, but present")
	}
}

func TestGetSuccessFilters(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// each filter string should return 1 item
	filterStrA := []string{

		// equals
		"?c_bool=true",
		"?c_int=1",
		"?c_double=1.1",
		"?c_date=2001-01-01",
		"?c_time=12:01",
		"?c_datetime=2001-01-01%2012:01:00%2B01",
		"?c_enum=Monday",
		"?c_text=a",
		"?c_textn={empty}",
		"?c_booln={null}",

		// not equals
		"?c_int=!1",
		"?c_double=!1.1",
		"?c_date=!2001-01-01",
		"?c_time=!12:01",
		"?c_datetime=!2001-01-01%2012:01:00%2B01",
		"?c_enum=!Monday",
		"?c_text=!a",
		"?c_textn={!empty}",
		"?c_booln={!null}",

		// greater than
		"?c_int=>1",
		"?c_double=>1.1",
		"?c_date=>2001-01-01",
		"?c_time=>12:01",
		"?c_datetime=>2001-01-01%2012:01:00%2B01",

		// less than
		"?c_int=<2",
		"?c_double=<2.1",
		"?c_date=<2002-01-01",
		"?c_time=<12:02",
		"?c_datetime=<2002-01-01%2012:01:00%2B01",

		// starts with
		"?c_date=2001~",
		"?c_time=12:01~",
		"?c_datetime=2001~",
		"?c_textn=a~",

		// ends with
		"?c_date=~01-01-01",
		"?c_time=~01:00",
		"?c_datetime=~01-01-01%2012:01:00%2B01",
		"?c_textn=~c",

		// contains
		"?c_date=~001~",
		"?c_time=~2:01~",
		"?c_datetime=~001~",
		"?c_textn=~b~",

		// not contains
		"?c_textn=!~b~",

		// in
		"?c_textn=abc|d",

		// not in
		"?c_textn=!abc|d",

		// contains any
		"?c_textn=~[b|d]~",
	}

	for _, filterStr := range filterStrA {
		targetUrl := "/param-test" + filterStr
		resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
		assert.EqualValues(t, 1, len(resp.Data), filterStr)
	}
}

func TestGetSuccessFormat(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// json
	targetUrl := "/param-test?id=1&xformat=json"
	resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 1, len(resp.Data), "format json")

	// csv
	targetUrl = "/param-test?id=1&xformat=csv"
	lysclient.MustGetFile(ctx, t, srvApp.getRouter(), targetUrl)

	// excel
	targetUrl = "/param-test?id=1&xformat=excel"
	lysclient.MustGetFile(ctx, t, srvApp.getRouter(), targetUrl)
}

func TestGetSuccessPaging(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// page 1, limit 1

	pagingStr := "?xsort=c_text&xpage=1&xper_page=1"
	targetUrl := "/param-test" + pagingStr
	resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "a", resp.Data[0]["c_text"].(string), "page 1 limit 1")

	// page 2, limit 1

	pagingStr = "?xsort=c_text&xpage=2&xper_page=1"
	targetUrl = "/param-test" + pagingStr
	resp = lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "b", resp.Data[0]["c_text"].(string), "page 2 limit 1")
}

func TestGetSuccessSorts(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// ascending: each sort string should return the first record
	sortStrA := []string{
		"?xsort=c_text",
		"?xsort=c_intn,c_text",
	}
	for _, sortStr := range sortStrA {
		targetUrl := "/param-test" + sortStr
		resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
		assert.EqualValues(t, "a", resp.Data[0]["c_text"].(string), sortStr)
	}

	// descending: each sort string should return the second record
	sortStrA = []string{
		"?xsort=-c_text",
		"?xsort=c_intn,-c_text",
	}
	for _, sortStr := range sortStrA {
		targetUrl := "/param-test" + sortStr
		resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
		assert.EqualValues(t, "b", resp.Data[0]["c_text"].(string), sortStr)
	}
}

func TestGetSetfuncSuccess(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// basic selection passing setfunc params only
	targetUrl := "/setfunc-test?p_text=a&p_int=1&p_inta=1,2"
	resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 2, len(resp.Data), "basic: len")
	assert.EqualValues(t, "a", resp.Data[0]["text_val"], "basic: text val")
	assert.EqualValues(t, 2, resp.Data[0]["int_val"], "basic: int val 1")
	assert.EqualValues(t, 3, resp.Data[1]["int_val"], "basic: int val 2")

	// with filter param
	targetUrl = "/setfunc-test?p_text=a&p_int=1&p_inta=1,2&int_val=2"
	resp = lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 1, len(resp.Data), "filtered: len")
	assert.EqualValues(t, 2, resp.Data[0]["int_val"], "filtered: int val")

	// with sort param
	targetUrl = "/setfunc-test?p_text=a&p_int=1&p_inta=1,2&xsort=-int_val"
	resp = lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 2, len(resp.Data), "sorted: len")
	assert.EqualValues(t, 3, resp.Data[0]["int_val"], "sorted: int val")
}

func TestGetSetfuncFailure(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	//targetUrl := "/setfunc-test?p_text=a&p_int=1&p_inta=1,2"

	// param missing
	targetUrl := "/setfunc-test?p_text=a&p_int=1"
	_, err := lysclient.GetItemRespTester(ctx, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "setFunc param name 'p_inta' is missing", err.Error())

	// param is wrong type
	// TODO - not working - GetItemsTester does not return an error
	// add test on db level
	/*targetUrl = "/setfunc-test?p_text=1&p_int=1&p_inta=1,2"
	_, err = lysclient.GetItemsTester(srvApp.getRouter(), targetUrl)
	assert.Error(t, err)*/
}

func TestGetVolumeSuccess(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// unfiltered total row count is estimated, but close to actual row count
	targetUrl := "/volume-test"
	resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, true, resp.GetMetadata.TotalCountIsEstimated)
	assert.InDelta(t, 1_000_000, resp.GetMetadata.TotalCount, 10_000, "unfiltered total count")

	// filtered total row count is estimated, but close to actual row count
	targetUrl = "/volume-test?c_int=5"
	resp = lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, true, resp.GetMetadata.TotalCountIsEstimated)
	assert.InDelta(t, 100_000, resp.GetMetadata.TotalCount, 5_000, "filtered total count")
}

func TestGetIrregularTags(t *testing.T) {

	// CExtra is a field in the Model which is populated in app code, not in db
	// CHidden column is hidden to API (no json tag)
	// CObscured column is obscured in API (different json tag)

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/tag-test"
	resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)

	cEditable, ok := resp.Data[0]["c_editable"]
	if !ok {
		t.Errorf("c_editable: expected in response, but missing")
	} else {
		assert.Equal(t, "e1", cEditable, "c_editable value")
	}

	cExtra, ok := resp.Data[0]["c_extra"]
	if !ok {
		t.Errorf("c_extra: expected in response, but missing")
	} else {
		assert.Equal(t, "extra", cExtra, "c_extra value")
	}

	_, ok = resp.Data[0]["c_hidden"]
	if ok {
		t.Errorf("c_hidden: not expected in response, but present")
	}

	// note that json name is used for output, not db name
	cObscured, ok := resp.Data[0]["c_obscured_json"]
	if !ok {
		t.Errorf("c_obscured_json: expected in response, but missing")
	} else {
		assert.Equal(t, "o1", cObscured, "c_obscured_json value")
	}
	_, ok = resp.Data[0]["c_obscured"]
	if ok {
		t.Errorf("c_obscured: not expected in response, but present")
	}

	// check hidden value with a select
	items, _, err := lyspg.Select[coretagtest.Model](ctx, srvApp.Db, "core", "tag_test", "v_tag_test", "id", []string{"id", "c_hidden"}, lyspg.SelectParams{})
	require.NoError(t, err, "lyspg.Select failed")
	assert.Equal(t, "h1", items[0].CHidden, "CHidden via DB")
}

// TestGetOptNilIsValid verifies that passing nil GetOpt works correctly
func TestGetOptNilIsValid(t *testing.T) {
	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// nil GetOpt should use all defaults
	targetUrl := "/param-test"
	resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 2, len(resp.Data), "nil GetOpt returns results")
}

// TestGetCsvContentType verifies CSV response has correct Content-Type header
func TestGetCsvContentType(t *testing.T) {
	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/param-test?id=1&xformat=csv"
	_, respHeaders := lysclient.MustGetFileWithHeaders(ctx, t, srvApp.getRouter(), targetUrl)
	assert.Equal(t, "text/csv", respHeaders.Get("Content-Type"), "CSV Content-Type")
}

// TestGetExcelContentType verifies Excel response has correct Content-Type header
func TestGetExcelContentType(t *testing.T) {
	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/param-test?id=1&xformat=excel"
	_, respHeaders := lysclient.MustGetFileWithHeaders(ctx, t, srvApp.getRouter(), targetUrl)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", respHeaders.Get("Content-Type"), "Excel Content-Type")
}

// TestGetCsvContentDisposition verifies CSV Content-Disposition header includes safe filename
func TestGetCsvContentDisposition(t *testing.T) {
	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/param-test?id=1&xformat=csv"
	_, respHeaders := lysclient.MustGetFileWithHeaders(ctx, t, srvApp.getRouter(), targetUrl)
	disposition := respHeaders.Get("Content-Disposition")
	assert.Contains(t, disposition, "attachment;", "CSV Content-Disposition has attachment")
	assert.Contains(t, disposition, "filename=", "CSV Content-Disposition has filename")
	assert.Contains(t, disposition, ".csv", "CSV Content-Disposition filename ends with .csv")
}

// TestGetExcelContentDisposition verifies Excel Content-Disposition header includes safe filename
func TestGetExcelContentDisposition(t *testing.T) {
	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/param-test?id=1&xformat=excel"
	_, respHeaders := lysclient.MustGetFileWithHeaders(ctx, t, srvApp.getRouter(), targetUrl)
	disposition := respHeaders.Get("Content-Disposition")
	assert.Contains(t, disposition, "attachment;", "Excel Content-Disposition has attachment")
	assert.Contains(t, disposition, "filename=", "Excel Content-Disposition has filename")
	assert.Contains(t, disposition, ".xlsx", "Excel Content-Disposition filename ends with .xlsx")
}

// TestGetCsvAlwaysOutputsFile verifies CSV format always outputs file, even with no results
func TestGetCsvAlwaysOutputsFile(t *testing.T) {
	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// Filter that returns no results
	targetUrl := "/param-test?c_int=99999&xformat=csv"
	_, respHeaders := lysclient.MustGetFileWithHeaders(ctx, t, srvApp.getRouter(), targetUrl)
	assert.Equal(t, "text/csv", respHeaders.Get("Content-Type"), "CSV format with no results still outputs file")
	assert.Contains(t, respHeaders.Get("Content-Disposition"), "attachment;", "CSV file headers present even with empty results")
}

// TestGetExcelAlwaysOutputsFile verifies Excel format always outputs file, even with no results
func TestGetExcelAlwaysOutputsFile(t *testing.T) {
	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// Filter that returns no results
	targetUrl := "/param-test?c_int=99999&xformat=excel"
	_, respHeaders := lysclient.MustGetFileWithHeaders(ctx, t, srvApp.getRouter(), targetUrl)
	assert.Equal(t, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", respHeaders.Get("Content-Type"), "Excel format with no results still outputs file")
	assert.Contains(t, respHeaders.Get("Content-Disposition"), "attachment;", "Excel file headers present even with empty results")
}

// TestGetSetFuncUrlParamNamesInheritedFromStore verifies SetFuncUrlParamNames from store
func TestGetSetFuncUrlParamNamesInheritedFromStore(t *testing.T) {
	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// setfunc-test route already passes SetFuncUrlParamNames from store
	targetUrl := "/setfunc-test?p_text=a&p_int=1&p_inta=1,2"
	resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 2, len(resp.Data), "setfunc with params works correctly")
}

// TestGetJsonMetadataIncluded verifies JSON response includes metadata
func TestGetJsonMetadataIncluded(t *testing.T) {
	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/param-test?xformat=json"
	resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.NotNil(t, resp.GetMetadata, "JSON response includes metadata")
	assert.Greater(t, resp.GetMetadata.Count, 0, "metadata Count is set")
	assert.Greater(t, resp.GetMetadata.TotalCount, int64(0), "metadata TotalCount is set")
}

// TestGetJsonDefaultFormat verifies JSON is default format when xformat not specified
func TestGetJsonDefaultFormat(t *testing.T) {
	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/param-test?id=1"
	resp := lysclient.MustGetItemResp(ctx, t, srvApp.getRouter(), targetUrl)
	assert.NotNil(t, resp.GetMetadata, "default format is JSON with metadata")
	assert.EqualValues(t, 1, len(resp.Data), "default format returns items")
}
