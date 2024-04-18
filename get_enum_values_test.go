package lys

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestGetEnumValuesSuccess(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// all values
	targetUrl := "/weekdays"
	vals := lysclient.MustGetArray[string](t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "None", vals[0])
	assert.EqualValues(t, 8, len(vals))

	// with include filter
	targetUrl = "/weekdays?vals=Sunday,Monday"
	vals = lysclient.MustGetArray[string](t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "Sunday", vals[0])
	assert.EqualValues(t, 2, len(vals))

	// with exclude filter
	targetUrl = "/weekdays?vals=!None,Sunday"
	vals = lysclient.MustGetArray[string](t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "Monday", vals[0])
	assert.EqualValues(t, 6, len(vals))

	// with ASC sort
	targetUrl = "/weekdays?xsort=val"
	vals = lysclient.MustGetArray[string](t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "Friday", vals[0])

	// with DESC sort
	targetUrl = "/weekdays?xsort=-val"
	vals = lysclient.MustGetArray[string](t, srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "Wednesday", vals[0])
}

func TestGetEnumValuesFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// invalid param key
	targetUrl := "/weekdays?a=1"
	_, err := lysclient.GetArrayTester[string](srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, "invalid enum filter field: a", err.Error())

	// invalid param value
	targetUrl = "/weekdays?vals=x"
	_, err = lysclient.GetArrayTester[string](srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, `invalid text: invalid input value for enum core.weekday: "x"`, err.Error())

	// invalid sort value
	targetUrl = "/weekdays?xsort=x"
	_, err = lysclient.GetArrayTester[string](srvApp.getRouter(), targetUrl)
	assert.EqualValues(t, `unknown enum sort value: x`, err.Error())
}
