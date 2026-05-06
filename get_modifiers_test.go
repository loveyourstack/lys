package lys

import (
	"net/url"
	"testing"

	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lysset"
	"github.com/stretchr/testify/assert"
)

func TestExtractFieldsSuccess(t *testing.T) {

	dbNames := lysset.New("a_db", "b_db", "c_db")
	jsonKeyDbNameMap := map[string]string{
		"a": "a_db",
		"b": "b_db",
		"c": "c_db",
	}
	fieldsParamName := "xfields"

	// with inclusion fields param
	t.Run("inclusion", func(t *testing.T) {
		fields, err := ExtractFields(fieldsParamName, "a,b", dbNames, jsonKeyDbNameMap)
		if err != nil {
			t.Errorf("ExtractFields failed: %v", err)
		}
		assert.EqualValues(t, []string{"a_db", "b_db"}, fields)
	})

	// with exclusion fields param
	t.Run("exclusion", func(t *testing.T) {
		fields, err := ExtractFields(fieldsParamName, "-a,b", dbNames, jsonKeyDbNameMap)
		if err != nil {
			t.Errorf("ExtractFields failed: %v", err)
		}
		assert.EqualValues(t, []string{"c_db"}, fields)
	})

	// without fields param (use default)
	t.Run("default", func(t *testing.T) {
		fields, err := ExtractFields(fieldsParamName, "", dbNames, jsonKeyDbNameMap)
		if err != nil {
			t.Errorf("ExtractFields failed: %v", err)
		}
		assert.EqualValues(t, []string{"a_db", "b_db", "c_db"}, fields)
	})
}

func TestExtractFieldsFailure(t *testing.T) {

	dbNames := lysset.New("a", "b")
	jsonKeyDbNameMap := map[string]string{
		"a": "a_db",
		"b": "b_db",
	}
	fieldsParamName := "xfields"

	// inclusion: invalid param value
	_, err := ExtractFields(fieldsParamName, "c", dbNames, jsonKeyDbNameMap)
	assert.EqualValues(t, "xfields param value is invalid: c", err.Error(), "inclusion: invalid param value")

	// exclusion: invalid param value
	_, err = ExtractFields(fieldsParamName, "-c", dbNames, jsonKeyDbNameMap)
	assert.EqualValues(t, "xfields param value is invalid: c", err.Error(), "exclusion: invalid param value")

	// exclusion: wrong usage
	_, err = ExtractFields(fieldsParamName, "-a,-b", dbNames, jsonKeyDbNameMap)
	assert.EqualValues(t, "xfields param value is invalid: -b", err.Error(), "exclusion: wrong usage")
}

func TestExtractFiltersOptionsSuccess(t *testing.T) {

	jsonKeyDbNameMap := map[string]string{
		"a": "a_db",
		"b": "b_db",
		"c": "c_db",
	}
	getOptions := mustFillGetOptions(t, GetOptions{})

	// equals
	urlValues := url.Values{}
	urlValues.Add("a", "1")
	conds := mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond := lyspg.Condition{Field: "a_db", Operator: lyspg.OpEquals, Value: "1"}
	assert.EqualValues(t, cond, conds[0], "equals")

	// not equals
	urlValues = url.Values{}
	urlValues.Add("a", "!1")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpNotEquals, Value: "1"}
	assert.EqualValues(t, cond, conds[0], "not equals")

	// greater than or equals
	urlValues = url.Values{}
	urlValues.Add("a", ">eq1")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpGreaterThanEquals, Value: "1"}
	assert.EqualValues(t, cond, conds[0], "greater than or equals")

	// greater than
	urlValues = url.Values{}
	urlValues.Add("a", ">1")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpGreaterThan, Value: "1"}
	assert.EqualValues(t, cond, conds[0], "greater than")

	// less than or equals
	urlValues = url.Values{}
	urlValues.Add("a", "<eq1")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpLessThanEquals, Value: "1"}
	assert.EqualValues(t, cond, conds[0], "less than or equals")

	// less than
	urlValues = url.Values{}
	urlValues.Add("a", "<1")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpLessThan, Value: "1"}
	assert.EqualValues(t, cond, conds[0], "less than")

	// starts with
	urlValues = url.Values{}
	urlValues.Add("a", "b~")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpStartsWith, Value: "b"}
	assert.EqualValues(t, cond, conds[0], "starts with")

	// ends with
	urlValues = url.Values{}
	urlValues.Add("a", "~b")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpEndsWith, Value: "b"}
	assert.EqualValues(t, cond, conds[0], "ends with")

	// contains
	urlValues = url.Values{}
	urlValues.Add("a", "~b~")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpContains, Value: "b"}
	assert.EqualValues(t, cond, conds[0], "contains")

	// not contains
	urlValues = url.Values{}
	urlValues.Add("a", "!~b~")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpNotContains, Value: "b"}
	assert.EqualValues(t, cond, conds[0], "not contains")

	// empty
	urlValues = url.Values{}
	urlValues.Add("a", "{empty}")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpEmpty, Value: "0"}
	assert.EqualValues(t, cond, conds[0], "empty")

	// not empty
	urlValues = url.Values{}
	urlValues.Add("a", "{!empty}")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpNotEmpty, Value: "0"}
	assert.EqualValues(t, cond, conds[0], "not empty")

	// null
	urlValues = url.Values{}
	urlValues.Add("a", "{null}")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpNull, Value: ""}
	assert.EqualValues(t, cond, conds[0], "null")

	// not null
	urlValues = url.Values{}
	urlValues.Add("a", "{!null}")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpNotNull, Value: ""}
	assert.EqualValues(t, cond, conds[0], "not null")

	// in
	urlValues = url.Values{}
	urlValues.Add("a", "b|c")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpIn, Value: "", InValues: []string{"b", "c"}}
	assert.EqualValues(t, cond, conds[0], "in")

	// not in
	urlValues = url.Values{}
	urlValues.Add("a", "!b|c")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpNotIn, Value: "", InValues: []string{"b", "c"}}
	assert.EqualValues(t, cond, conds[0], "not in")

	// contains any
	urlValues = url.Values{}
	urlValues.Add("a", "~[b|c]~")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpContainsAny, Value: "", InValues: []string{"b", "c"}}
	assert.EqualValues(t, cond, conds[0], "contains any")
}

func TestExtractFiltersOtherSuccess(t *testing.T) {

	jsonKeyDbNameMap := map[string]string{
		"a": "a_db",
		"b": "b_db",
		"c": "c_db",
	}
	getOptions := mustFillGetOptions(t, GetOptions{})

	// multiple filters, different keys
	urlValues := url.Values{}
	urlValues.Add("a", "1")
	urlValues.Add("b", "c")
	conds := mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond1 := lyspg.Condition{Field: "a_db", Operator: lyspg.OpEquals, Value: "1"}
	cond2 := lyspg.Condition{Field: "b_db", Operator: lyspg.OpEquals, Value: "c"}

	// map order not guaranteed: ensure both conds are present in any order
	var cond1Found, cond2Found bool
	for _, cond := range conds {
		if cond.Field == cond1.Field && cond.Value == cond1.Value {
			cond1Found = true
		}
		if cond.Field == cond2.Field && cond.Value == cond2.Value {
			cond2Found = true
		}
	}
	assert.EqualValues(t, true, cond1Found, "multiple, diff keys, cond1")
	assert.EqualValues(t, true, cond2Found, "multiple, diff keys, cond2")

	// multiple filters, same keys
	urlValues = url.Values{}
	urlValues.Add("a", "1")
	urlValues.Add("a", "2")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond1 = lyspg.Condition{Field: "a_db", Operator: lyspg.OpEquals, Value: "1"}
	cond2 = lyspg.Condition{Field: "a_db", Operator: lyspg.OpEquals, Value: "2"}

	cond1Found = false
	cond2Found = false
	for _, cond := range conds {
		if cond.Field == cond1.Field && cond.Value == cond1.Value {
			cond1Found = true
		}
		if cond.Field == cond2.Field && cond.Value == cond2.Value {
			cond2Found = true
		}
	}
	assert.EqualValues(t, true, cond1Found, "multiple, same keys, cond1")
	assert.EqualValues(t, true, cond2Found, "multiple, same keys, cond2")

	//------------------

	// ignore params used for fields, paging, sorting
	urlValues = url.Values{}
	urlValues.Add("a", "1")
	urlValues.Add(getOptions.FieldsParamName, "1")
	urlValues.Add(getOptions.PageParamName, "1")
	urlValues.Add(getOptions.PerPageParamName, "1")
	urlValues.Add(getOptions.SortParamName, "1")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	assert.EqualValues(t, 1, len(conds), "ignore special params")

	// ignore params used as setFunc params
	setFuncParamNames := []string{"x"}

	urlValues = url.Values{}
	urlValues.Add("a", "1")
	urlValues.Add("x", "1")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, setFuncParamNames, getOptions)
	assert.EqualValues(t, 1, len(conds), "ignore setFuncParamNames")

	//------------------

	// empty metadata
	urlValues = url.Values{}
	urlValues.Add("a", "1^")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond := lyspg.Condition{Field: "a_db", Operator: lyspg.OpEquals, Value: "1", Metadata: ""}
	assert.EqualValues(t, cond, conds[0], "empty metadata")

	// metadata
	urlValues = url.Values{}
	urlValues.Add("a", "1^b")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpEquals, Value: "1", Metadata: "b"}
	assert.EqualValues(t, cond, conds[0], "metadata")

	// metadata separator repeated
	urlValues = url.Values{}
	urlValues.Add("a", "1^b^c")
	conds = mustExtractFilters(t, urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	cond = lyspg.Condition{Field: "a_db", Operator: lyspg.OpEquals, Value: "1", Metadata: "bc"}
	assert.EqualValues(t, cond, conds[0], "metadata separator repeated")
}

func TestExtractFiltersFailure(t *testing.T) {

	jsonKeyDbNameMap := map[string]string{
		"a": "a_db",
		"b": "b_db",
		"c": "c_db",
	}
	getOptions := mustFillGetOptions(t, GetOptions{})

	// invalid param key
	urlValues := url.Values{}
	urlValues.Add("d", "1")
	_, err := ExtractFilters(urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	assert.EqualValues(t, "invalid filter field: d", err.Error())

	// empty param value
	urlValues = url.Values{}
	urlValues.Add("a", "")
	_, err = ExtractFilters(urlValues, jsonKeyDbNameMap, nil, nil, getOptions)
	assert.EqualValues(t, "empty value in filter field: a", err.Error())
}

func TestExtractFormatSuccess(t *testing.T) {

	formatParamName := "xformat"

	// with fields param
	format, err := ExtractFormat(formatParamName, "excel")
	if err != nil {
		t.Errorf("ExtractFormat failed: %v", err)
	}
	assert.EqualValues(t, "excel", format)

	// without format param (use json default)
	format, err = ExtractFormat(formatParamName, "")
	if err != nil {
		t.Errorf("ExtractFormat failed: %v", err)
	}
	assert.EqualValues(t, "json", format)
}

func TestExtractFormatFailure(t *testing.T) {

	formatParamName := "xformat"

	// invalid value
	_, err := ExtractFormat(formatParamName, "a")
	assert.EqualValues(t, "xformat param value is invalid: a", err.Error())
}

func TestExtractPagingSuccess(t *testing.T) {

	pageParamName := "xpage"
	perPageParamName := "xper_page"
	defaultPerPage := 20
	maxPerPage := 500

	// with paging params
	page, perPage, err := ExtractPaging(pageParamName, "2", perPageParamName, "50", defaultPerPage, maxPerPage)
	if err != nil {
		t.Errorf("ExtractPaging failed: %v", err)
	}
	assert.EqualValues(t, 2, page)
	assert.EqualValues(t, 50, perPage)

	// without paging params (use default)
	page, perPage, err = ExtractPaging(pageParamName, "", perPageParamName, "", defaultPerPage, maxPerPage)
	if err != nil {
		t.Errorf("ExtractPaging failed: %v", err)
	}
	assert.EqualValues(t, 1, page)
	assert.EqualValues(t, defaultPerPage, perPage)

	// enforcement of maxPerPage
	_, perPage, err = ExtractPaging(pageParamName, "", perPageParamName, "1000000", defaultPerPage, maxPerPage)
	if err != nil {
		t.Errorf("ExtractPaging failed: %v", err)
	}
	assert.EqualValues(t, maxPerPage, perPage)
}

func TestExtractPagingFailure(t *testing.T) {

	pageParamName := "xpage"
	perPageParamName := "xper_page"
	defaultPerPage := 20
	maxPerPage := 500

	// invalid maxPerPage
	_, _, err := ExtractPaging(pageParamName, "", perPageParamName, "", defaultPerPage, 0)
	assert.EqualValues(t, "maxPerPage must be >= 1", err.Error())

	// page wrong type
	_, _, err = ExtractPaging(pageParamName, "a", perPageParamName, "", defaultPerPage, maxPerPage)
	assert.EqualValues(t, "xpage param value is not an integer", err.Error())

	// page invalid
	_, _, err = ExtractPaging(pageParamName, "0", perPageParamName, "", defaultPerPage, maxPerPage)
	assert.EqualValues(t, "xpage param value must be >= 1", err.Error())

	// perPage wrong type
	_, _, err = ExtractPaging(pageParamName, "", perPageParamName, "a", defaultPerPage, maxPerPage)
	assert.EqualValues(t, "xper_page param value is not an integer", err.Error())

	// perPage invalid
	_, _, err = ExtractPaging(pageParamName, "", perPageParamName, "0", defaultPerPage, maxPerPage)
	assert.EqualValues(t, "xper_page param value must be >= 1", err.Error())
}

func TestExtractSetFuncParamValuesSuccess(t *testing.T) {

	setFuncParamNames := []string{"x", "y"}

	r := mustCreateGetReq(t, "/test?x=1&y=2")
	setFuncParamValues, err := ExtractSetFuncParamValues(r, setFuncParamNames)
	if err != nil {
		t.Errorf("ExtractSetFuncParamValues failed: %v", err)
	}
	assert.EqualValues(t, []any{"1", "2"}, setFuncParamValues)
}

func TestExtractSetFuncParamValuesFailure(t *testing.T) {

	setFuncParamNames := []string{"x", "y"}

	// missing a value
	r := mustCreateGetReq(t, "/test?x=1")
	_, err := ExtractSetFuncParamValues(r, setFuncParamNames)
	assert.EqualValues(t, "setFunc param name 'y' is missing", err.Error())
}

func TestExtractSortsSuccess(t *testing.T) {

	jsonKeyDbNameMap := map[string]string{
		"a": "a_db",
		"b": "b_db",
		"c": "c_db",
	}
	sortParamName := "xsort"

	// with single ASC sort param
	sortCols, err := ExtractSorts(sortParamName, "a", jsonKeyDbNameMap)
	if err != nil {
		t.Errorf("ExtractSorts failed: %v", err)
	}
	assert.EqualValues(t, []string{"a_db"}, sortCols)

	// with single DESC sort param
	sortCols, err = ExtractSorts(sortParamName, "-a", jsonKeyDbNameMap)
	if err != nil {
		t.Errorf("ExtractSorts failed: %v", err)
	}
	assert.EqualValues(t, []string{"a_db DESC"}, sortCols)

	// with mixed sort params
	sortCols, err = ExtractSorts(sortParamName, "a,-b", jsonKeyDbNameMap)
	if err != nil {
		t.Errorf("ExtractSorts failed: %v", err)
	}
	assert.EqualValues(t, []string{"a_db", "b_db DESC"}, sortCols)

	// without sort param (no default)
	sortCols, err = ExtractSorts(sortParamName, "", jsonKeyDbNameMap)
	if err != nil {
		t.Errorf("ExtractSorts failed: %v", err)
	}
	assert.EqualValues(t, []string(nil), sortCols)
}

func TestExtractSortsFailure(t *testing.T) {

	jsonKeyDbNameMap := map[string]string{
		"a": "a_db",
		"b": "b_db",
	}
	sortReqParamName := "xsort"

	// invalid param value
	_, err := ExtractSorts(sortReqParamName, "c", jsonKeyDbNameMap)
	assert.EqualValues(t, "xsort has invalid field: c", err.Error())
}
