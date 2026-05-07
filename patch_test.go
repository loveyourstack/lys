package lys

import (
	"context"
	"strconv"
	"testing"

	"github.com/loveyourstack/lys/internal/stores/core/coretrackingtest"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/internal/stores/core/coreuuidtest"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/stretchr/testify/assert"
)

func TestPatchSuccess(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// create a record with minimal values
	minInput := coretypetestm.GetEmptyInput()
	newId := lysclient.MustPostToValue[coretypetestm.Input, int64](ctx, t, srvApp.getRouter(), "POST", "/type-test", minInput)

	targetUrl := "/type-test/" + strconv.FormatInt(newId, 10)

	// get filled type test input
	filledInput, err := coretypetestm.GetFilledInput()
	if err != nil {
		t.Fatalf("coretypetestm.GetFilledInput failed: %v", err)
	}

	// PATCH the filled input to the minimal record
	_ = lysclient.MustPostToValue[coretypetestm.Input, string](ctx, t, srvApp.getRouter(), "PATCH", targetUrl, filledInput)

	// get changed record
	filledItem := lysclient.MustDoToValue[coretypetestm.Model](ctx, t, srvApp.getRouter(), "GET", targetUrl)

	// check changed record
	coretypetestm.TestFilledInput(t, filledItem.Input)
}

func TestPatchIrregularTags(t *testing.T) {

	// CHidden column is hidden to API (no json tag)
	// CObscured column is obscured in API (different json tag)

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/tag-test/3"

	t.Run("field with no json tag", func(t *testing.T) {

		hiddenMap := make(map[string]any)
		hiddenMap["c_hidden"] = "h33"

		_, err := lysclient.PostToValueTester[map[string]any, string](ctx, srvApp.getRouter(), "PATCH", targetUrl, hiddenMap)
		assert.EqualValues(t, "invalid field: c_hidden", err.Error())
	})

	t.Run("trying to use db name when json tag should be used", func(t *testing.T) {
		obscuredMap := make(map[string]any)
		obscuredMap["c_obscured"] = "o33"

		_, err := lysclient.PostToValueTester[map[string]any, string](ctx, srvApp.getRouter(), "PATCH", targetUrl, obscuredMap)
		assert.EqualValues(t, "invalid field: c_obscured", err.Error())
	})

	t.Run("using correct json tag for obscured field", func(t *testing.T) {
		obscuredJsonMap := make(map[string]any)
		obscuredJsonMap["c_obscured_json"] = "o33"

		_, err := lysclient.PostToValueTester[map[string]any, string](ctx, srvApp.getRouter(), "PATCH", targetUrl, obscuredJsonMap)
		assert.NoError(t, err)
	})
}

func TestPatchUuid(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/uuid-test/0707b293-b39c-4ae6-8ab9-a92b592e9568"

	assignmentsMap := make(map[string]any)
	assignmentsMap["c_int"] = int64(4)

	_ = lysclient.MustPostToValue[map[string]any, string](ctx, t, srvApp.getRouter(), "PATCH", targetUrl, assignmentsMap)
	item := lysclient.MustDoToValue[coreuuidtest.Model](ctx, t, srvApp.getRouter(), "GET", targetUrl)

	assert.Equal(t, int64(4), item.CInt, "CInt")
}

func TestPatchWithExtras(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/tracking-test/2"

	assignmentsMap := make(map[string]any)
	assignmentsMap["c_editable"] = "e22"

	_ = lysclient.MustPostToValue[map[string]any, string](ctx, t, srvApp.getRouter(), "PATCH", targetUrl, assignmentsMap)
	item := lysclient.MustDoToValue[coretrackingtest.Model](ctx, t, srvApp.getRouter(), "GET", targetUrl)

	assert.Equal(t, "e22", item.CEditable, "CEditable")
	assert.Equal(t, "insert", item.CreatedBy, "CreatedBy")
	assert.Equal(t, "update partial", item.LastUserUpdateBy, "LastUserUpdateBy")
}

func TestPatchFailure(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	minInput := coretypetestm.GetEmptyInput()
	newId := lysclient.MustPostToValue[coretypetestm.Input, int64](ctx, t, srvApp.getRouter(), "POST", "/type-test", minInput)

	targetUrl := "/type-test/" + strconv.FormatInt(newId, 10)

	t.Run("json not parseable to map[string]any", func(t *testing.T) {
		_, err := lysclient.PostToValueTester[string, string](ctx, srvApp.getRouter(), "PATCH", targetUrl, "not a map")
		assert.EqualValues(t, "json body could not be parsed into a map of field names to values", err.Error())
	})

	t.Run("invalid field", func(t *testing.T) {

		type testS struct {
			Val string
		}
		inputTestS := testS{
			Val: "a",
		}
		_, err := lysclient.PostToValueTester[testS, string](ctx, srvApp.getRouter(), "PATCH", targetUrl, inputTestS)
		assert.EqualValues(t, "invalid field: Val", err.Error(), "invalid field")
	})

	t.Run("no assignments", func(t *testing.T) {
		_, err := lysclient.PostToValueTester[any, string](ctx, srvApp.getRouter(), "PATCH", targetUrl, nil)
		assert.EqualValues(t, "no assignments found", err.Error())
	})

	t.Run("empty struct", func(t *testing.T) {
		inputTT := coretypetestm.Input{}
		_, err := lysclient.PostToValueTester[coretypetestm.Input, string](ctx, srvApp.getRouter(), "PATCH", targetUrl, inputTT)
		assert.EqualValues(t, "invalid text: invalid input value for enum core.weekday: \"\"", err.Error())
	})

	t.Run("id wrong type", func(t *testing.T) {
		_, err := lysclient.PostToValueTester[coretypetestm.Input, string](ctx, srvApp.getRouter(), "PATCH", "/type-test/a", minInput)
		assert.EqualValues(t, "id could not be parsed", err.Error())
	})

	t.Run("invalid id", func(t *testing.T) {
		_, err := lysclient.PostToValueTester[coretypetestm.Input, string](ctx, srvApp.getRouter(), "PATCH", "/type-test/100000", minInput)
		assert.EqualValues(t, "row(s) not found", err.Error())
	})
}
