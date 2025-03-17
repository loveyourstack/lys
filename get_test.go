package lys

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestGetSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// basic test only: sorting, filtering etc is tested below
	targetUrl := "/param-test"
	resp := lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, false, resp.GetMetadata.TotalCountIsEstimated)
	assert.EqualValues(t, 2, resp.GetMetadata.TotalCount)
	assert.EqualValues(t, 2, resp.GetMetadata.Count)
	assert.EqualValues(t, 2, len(resp.Data))
}

func TestGetFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// invalid url
	targetUrl := "/xx"
	_, err := lysclient.GetItemRespTester(srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "route not found", err.Error())

	// invalid filter field
	targetUrl = "/param-test?a=b"
	_, err = lysclient.GetItemRespTester(srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "invalid filter field: a", err.Error())

	// TODO: further param tests on those functions directly
}

func TestGetSuccessFields(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	targetUrl := "/param-test?xfields=id"
	resp := lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
	item := resp.Data[0]
	_, ok := item["id"]
	if !ok {
		t.Errorf("id: expected in response, but missing")
	}
	_, ok = item["c_int"]
	if ok {
		t.Errorf("c_int: not expected in response, but present")
	}
}

func TestGetSuccessFilters(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
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
		resp := lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
		assert.EqualValues(t, 1, len(resp.Data), filterStr)
	}
}

func TestGetSuccessFormat(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// json
	targetUrl := "/param-test?id=1&xformat=json"
	resp := lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 1, len(resp.Data), "format json")

	// excel
	targetUrl = "/param-test?id=1&xformat=excel"
	lysclient.MustGetFile(t, srvApp.getRouter(), targetUrl)
}

func TestGetSuccessPaging(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// page 1, limit 1

	pagingStr := "?xsort=c_text&xpage=1&xper_page=1"
	targetUrl := "/param-test" + pagingStr
	resp := lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "a", resp.Data[0]["c_text"].(string), "page 1 limit 1")

	// page 2, limit 1

	pagingStr = "?xsort=c_text&xpage=2&xper_page=1"
	targetUrl = "/param-test" + pagingStr
	resp = lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "b", resp.Data[0]["c_text"].(string), "page 2 limit 1")
}

func TestGetSuccessSorts(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// ascending: each sort string should return the first record
	sortStrA := []string{
		"?xsort=c_text",
		"?xsort=c_intn,c_text",
	}
	for _, sortStr := range sortStrA {
		targetUrl := "/param-test" + sortStr
		resp := lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
		assert.EqualValues(t, "a", resp.Data[0]["c_text"].(string), sortStr)
	}

	// descending: each sort string should return the second record
	sortStrA = []string{
		"?xsort=-c_text",
		"?xsort=c_intn,-c_text",
	}
	for _, sortStr := range sortStrA {
		targetUrl := "/param-test" + sortStr
		resp := lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
		assert.EqualValues(t, "b", resp.Data[0]["c_text"].(string), sortStr)
	}
}

func TestGetSetfuncSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// basic selection passing setfunc params only
	targetUrl := "/setfunc-test?p_text=a&p_int=1&p_inta=1,2"
	resp := lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 2, len(resp.Data), "basic: len")
	assert.EqualValues(t, "a", resp.Data[0]["text_val"], "basic: text val")
	assert.EqualValues(t, 2, resp.Data[0]["int_val"], "basic: int val 1")
	assert.EqualValues(t, 3, resp.Data[1]["int_val"], "basic: int val 2")

	// with filter param
	targetUrl = "/setfunc-test?p_text=a&p_int=1&p_inta=1,2&int_val=2"
	resp = lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 1, len(resp.Data), "filtered: len")
	assert.EqualValues(t, 2, resp.Data[0]["int_val"], "filtered: int val")

	// with sort param
	targetUrl = "/setfunc-test?p_text=a&p_int=1&p_inta=1,2&xsort=-int_val"
	resp = lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 2, len(resp.Data), "sorted: len")
	assert.EqualValues(t, 3, resp.Data[0]["int_val"], "sorted: int val")
}

func TestGetSetfuncFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	//targetUrl := "/setfunc-test?p_text=a&p_int=1&p_inta=1,2"

	// param missing
	targetUrl := "/setfunc-test?p_text=a&p_int=1"
	_, err := lysclient.GetItemRespTester(srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "setFunc param name p_inta is missing", err.Error())

	// param is wrong type
	// TODO - not working - GetItemsTester does not return an error
	// add test on db level
	/*targetUrl = "/setfunc-test?p_text=1&p_int=1&p_inta=1,2"
	_, err = lysclient.GetItemsTester(srvApp.getRouter(), targetUrl)
	assert.Error(t, err)*/
}

func TestGetVolumeSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// unfiltered total row count is estimated, but close to actual row count
	targetUrl := "/volume-test"
	resp := lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, true, resp.GetMetadata.TotalCountIsEstimated)
	assert.InDelta(t, 1_000_000, resp.GetMetadata.TotalCount, 10_000, "unfiltered total count")

	// filtered total row count is estimated, but close to actual row count
	targetUrl = "/volume-test?c_int=5"
	resp = lysclient.MustGetItemResp(t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, true, resp.GetMetadata.TotalCountIsEstimated)
	assert.InDelta(t, 100_000, resp.GetMetadata.TotalCount, 5_000, "filtered total count")
}
