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

	targetUrl := "/param-test"
	items := lysclient.MustGetItems(t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, 2, len(items))
}

func TestGetSuccessFields(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	targetUrl := "/param-test?xfields=id"
	items := lysclient.MustGetItems(t, srvApp.getRouter(), targetUrl)
	item := items[0]
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
		items := lysclient.MustGetItems(t, srvApp.getRouter(), targetUrl)
		assert.EqualValues(t, 1, len(items), filterStr)
	}
}

func TestGetSuccessPaging(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// page 1, limit 1

	pagingStr := "?xsort=c_text&xpage=1&xper_page=1"
	targetUrl := "/param-test" + pagingStr
	items := lysclient.MustGetItems(t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "a", items[0]["c_text"].(string), "page 1 limit 1")

	// page 2, limit 1

	pagingStr = "?xsort=c_text&xpage=2&xper_page=1"
	targetUrl = "/param-test" + pagingStr
	items = lysclient.MustGetItems(t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "b", items[0]["c_text"].(string), "page 2 limit 1")
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
		items := lysclient.MustGetItems(t, srvApp.getRouter(), targetUrl)
		assert.EqualValues(t, "a", items[0]["c_text"].(string), sortStr)
	}

	// descending: each sort string should return the second record
	sortStrA = []string{
		"?xsort=-c_text",
		"?xsort=c_intn,-c_text",
	}
	for _, sortStr := range sortStrA {
		targetUrl := "/param-test" + sortStr
		items := lysclient.MustGetItems(t, srvApp.getRouter(), targetUrl)
		assert.EqualValues(t, "b", items[0]["c_text"].(string), sortStr)
	}
}

func TestGetFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// invalid filter field
	targetUrl := "/param-test?a=b"
	_, err := lysclient.GetItemsTester(srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "invalid filter field: a", err.Error())

	// TODO: further param tests on those functions directly
}
