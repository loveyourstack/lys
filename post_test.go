package lys

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
)

func TestPostEmptySuccess(t *testing.T) {

	// checks the result of posting a type test with as many fields missing as possible

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	input := coretypetestm.GetEmptyInput()
	newId := lysclient.MustPostToValue[coretypetestm.Input, int64](t, srvApp.getRouter(), "POST", "/type-test", input)
	item := lysclient.MustDoToValue[coretypetestm.Model](t, srvApp.getRouter(), "GET", fmt.Sprintf("/type-test/%v", newId))

	// check nullable fields are nil
	assert.Nil(t, item.CBoolN, "CBoolN")
	assert.Nil(t, item.CIntN, "CIntN")
	assert.Nil(t, item.CDoubleN, "CDoubleN")
	assert.Nil(t, item.CNumericN, "CNumericN")
	assert.Nil(t, item.CEnumN, "CEnumN")
	assert.Nil(t, item.CTextN, "CTextN")

	// default lystype date types should be the zero values, as with other data types

	// date
	expectedCDate := lystype.Date(time.Time{})
	assert.EqualValues(t, expectedCDate, item.CDate, "CDate")

	// time (compare formatted: the check like date above doesn't work)
	expectedCTime := time.Time{}.Format(lystype.TimeFormat)
	cTimeStr := item.CTime.Format(lystype.TimeFormat)
	assert.EqualValues(t, expectedCTime, cTimeStr, "CTime")

	// datetime (compare formatted: the check like date above doesn't work)
	// only checking date, not time: the time doesn't match exactly
	expectedCDatetime := time.Time{}.Format(lystype.DateFormat)
	cDatetimeStr := item.CDatetime.Format(lystype.DateFormat)
	assert.EqualValues(t, expectedCDatetime, cDatetimeStr, "CDatetime")
}

func TestPostFilledSuccess(t *testing.T) {

	// checks the result of posting a type test with all fields entered

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	input, err := coretypetestm.GetFilledInput()
	if err != nil {
		t.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
	}
	newId := lysclient.MustPostToValue[coretypetestm.Input, int64](t, srvApp.getRouter(), "POST", "/type-test", input)
	item := lysclient.MustDoToValue[coretypetestm.Model](t, srvApp.getRouter(), "GET", fmt.Sprintf("/type-test/%v", newId))
	coretypetestm.TestFilledInput(t, item.Input)
}

func TestPostFailure(t *testing.T) {

	srvApp := mustGetSrvApp(t, context.Background())
	defer srvApp.Db.Close()

	// struct with unknown field
	type testS struct {
		Val string
	}
	inputTestS := testS{
		Val: "a",
	}
	_, err := lysclient.PostToValueTester[testS, coretypetestm.Model](srvApp.getRouter(), "POST", "/type-test", inputTestS)
	assert.EqualValues(t, "unknown field: Val", err.Error(), "unknown field")

	// nil input (fails on mandatory enum val)
	_, err = lysclient.PostToValueTester[any, coretypetestm.Model](srvApp.getRouter(), "POST", "/type-test", nil)
	assert.EqualValues(t, `invalid text: invalid input value for enum core.weekday: ""`, err.Error(), "nil")

	// empty struct (fails on mandatory enum val)
	inputTT := coretypetestm.Input{}
	_, err = lysclient.PostToValueTester[coretypetestm.Input, coretypetestm.Model](srvApp.getRouter(), "POST", "/type-test", inputTT)
	assert.EqualValues(t, `invalid text: invalid input value for enum core.weekday: ""`, err.Error(), "empty struct")
}
