package lys

import (
	"context"
	"testing"

	"github.com/loveyourstack/lys/internal/stores/core/coretagtest"
	"github.com/loveyourstack/lys/internal/stores/core/coretypetestm"
	"github.com/loveyourstack/lys/lysclient"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetByIdSuccess(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/type-test/1"
	item := lysclient.MustDoToValue[coretypetestm.Model](ctx, t, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, true, item.CBool)
	assert.EqualValues(t, "a b", item.CText)
}

func TestGetByIdIrregularTags(t *testing.T) {

	// CExtra is a field in the Model which is populated in app code, not in db
	// CHidden column is hidden to API (no json tag)
	// CObscured column is obscured in API (different json tag)

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	targetUrl := "/tag-test/1"
	item := lysclient.MustDoToValue[coretagtest.Model](ctx, t, srvApp.getRouter(), "GET", targetUrl)
	assert.Equal(t, "e1", item.CEditable, "CEditable")
	assert.Equal(t, "extra", item.CExtra, "CExtra")
	assert.Equal(t, "", item.CHidden, "CHidden via API") // hidden field should be empty in API result
	assert.Equal(t, "o1", item.CObscured, "CObscured")

	// check hidden value with a select
	item, err := lyspg.SelectUnique[coretagtest.Model](ctx, srvApp.Db, "core", "tag_test", "id", 1)
	require.NoError(t, err, "lyspg.SelectUnique failed")
	assert.Equal(t, "h1", item.CHidden, "CHidden via DB")
}

func TestGetByIdFailure(t *testing.T) {

	ctx := context.Background()
	srvApp := mustGetSrvApp(ctx, t)
	defer srvApp.Db.Close()

	// id wrong type
	targetUrl := "/type-test/a"
	_, err := lysclient.DoToValueTester[coretypetestm.Model](ctx, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, "id not an integer", err.Error())

	// id doesn't exist
	targetUrl = "/type-test/100000"
	_, err = lysclient.DoToValueTester[coretypetestm.Model](ctx, srvApp.getRouter(), "GET", targetUrl)
	assert.EqualValues(t, "row(s) not found", err.Error())
}
