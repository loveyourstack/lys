package lys

import (
	"net/url"
	"testing"

	"github.com/loveyourstack/lys/lyspg"
	"github.com/stretchr/testify/assert"
)

func TestExtractFieldsSuccess(t *testing.T) {

	validJsonFields := []string{"a", "b", "c"}
	fieldsReqParamName := "xfields"

	// with fields param
	r := mustCreateGetReq(t, "/test?xfields=a,b")
	fields, err := ExtractFields(r, validJsonFields, fieldsReqParamName)
	if err != nil {
		t.Errorf("ExtractFields failed: %v", err)
	}
	assert.EqualValues(t, []string{"a", "b"}, fields)

	// without fields param (use default)
	r = mustCreateGetReq(t, "/test")
	fields, err = ExtractFields(r, validJsonFields, fieldsReqParamName)
	if err != nil {
		t.Errorf("ExtractFields failed: %v", err)
	}
	assert.EqualValues(t, validJsonFields, fields)
}

func TestExtractFieldsFailure(t *testing.T) {

	validJsonFields := []string{"a", "b"}
	fieldsReqParamName := "xfields"

	// invalid param value
	r := mustCreateGetReq(t, "/test?xfields=c")
	_, err := ExtractFields(r, validJsonFields, fieldsReqParamName)
	assert.EqualValues(t, "xfields param value is invalid: c", err.Error())
}

func TestExtractFiltersOptionsSuccess(t *testing.T) {

	validJsonFields := []string{"a", "b", "c"}
	getOptions := FillGetOptions(GetOptions{})

	// equals
	urlValues := url.Values{}
	urlValues.Add("a", "1")
	conds := mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond := lyspg.Condition{Field: "a", Operator: lyspg.OpEquals, Value: "1"}
	assert.EqualValues(t, cond, conds[0], "equals")

	// not equals
	urlValues = url.Values{}
	urlValues.Add("a", "!1")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpNotEquals, Value: "1"}
	assert.EqualValues(t, cond, conds[0], "not equals")

	// greater than or equals
	urlValues = url.Values{}
	urlValues.Add("a", ">eq1")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpGreaterThanEquals, Value: "1"}
	assert.EqualValues(t, cond, conds[0], "greater than or equals")

	// greater than
	urlValues = url.Values{}
	urlValues.Add("a", ">1")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpGreaterThan, Value: "1"}
	assert.EqualValues(t, cond, conds[0], "greater than")

	// less than or equals
	urlValues = url.Values{}
	urlValues.Add("a", "<eq1")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpLessThanEquals, Value: "1"}
	assert.EqualValues(t, cond, conds[0], "less than or equals")

	// less than
	urlValues = url.Values{}
	urlValues.Add("a", "<1")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpLessThan, Value: "1"}
	assert.EqualValues(t, cond, conds[0], "less than")

	// starts with
	urlValues = url.Values{}
	urlValues.Add("a", "b~")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpStartsWith, Value: "b"}
	assert.EqualValues(t, cond, conds[0], "starts with")

	// ends with
	urlValues = url.Values{}
	urlValues.Add("a", "~b")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpEndsWith, Value: "b"}
	assert.EqualValues(t, cond, conds[0], "ends with")

	// contains
	urlValues = url.Values{}
	urlValues.Add("a", "~b~")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpContains, Value: "b"}
	assert.EqualValues(t, cond, conds[0], "contains")

	// not contains
	urlValues = url.Values{}
	urlValues.Add("a", "!~b~")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpNotContains, Value: "b"}
	assert.EqualValues(t, cond, conds[0], "not contains")

	// empty
	urlValues = url.Values{}
	urlValues.Add("a", "{empty}")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpEmpty, Value: "0"}
	assert.EqualValues(t, cond, conds[0], "empty")

	// not empty
	urlValues = url.Values{}
	urlValues.Add("a", "{!empty}")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpNotEmpty, Value: "0"}
	assert.EqualValues(t, cond, conds[0], "not empty")

	// null
	urlValues = url.Values{}
	urlValues.Add("a", "{null}")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpNull, Value: ""}
	assert.EqualValues(t, cond, conds[0], "null")

	// not null
	urlValues = url.Values{}
	urlValues.Add("a", "{!null}")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpNotNull, Value: ""}
	assert.EqualValues(t, cond, conds[0], "not null")

	// in
	urlValues = url.Values{}
	urlValues.Add("a", "b|c")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpIn, Value: "", InValues: []string{"b", "c"}}
	assert.EqualValues(t, cond, conds[0], "in")

	// not in
	urlValues = url.Values{}
	urlValues.Add("a", "!b|c")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpNotIn, Value: "", InValues: []string{"b", "c"}}
	assert.EqualValues(t, cond, conds[0], "not in")

	// contains any
	urlValues = url.Values{}
	urlValues.Add("a", "~[b|c]~")
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond = lyspg.Condition{Field: "a", Operator: lyspg.OpContainsAny, Value: "", InValues: []string{"b", "c"}}
	assert.EqualValues(t, cond, conds[0], "contains any")
}

func TestExtractFiltersOtherSuccess(t *testing.T) {

	validJsonFields := []string{"a", "b", "c"}
	getOptions := FillGetOptions(GetOptions{})

	// multiple filters, different keys
	urlValues := url.Values{}
	urlValues.Add("a", "1")
	urlValues.Add("b", "c")
	conds := mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond1 := lyspg.Condition{Field: "a", Operator: lyspg.OpEquals, Value: "1"}
	cond2 := lyspg.Condition{Field: "b", Operator: lyspg.OpEquals, Value: "c"}

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
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	cond1 = lyspg.Condition{Field: "a", Operator: lyspg.OpEquals, Value: "1"}
	cond2 = lyspg.Condition{Field: "a", Operator: lyspg.OpEquals, Value: "2"}

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
	conds = mustExtractFilters(t, urlValues, validJsonFields, nil, getOptions)
	assert.EqualValues(t, 1, len(conds), "ignore special params")

	// ignore params used as setFunc params
	setFuncParamNames := []string{"x"}

	urlValues = url.Values{}
	urlValues.Add("a", "1")
	urlValues.Add("x", "1")
	conds = mustExtractFilters(t, urlValues, validJsonFields, setFuncParamNames, getOptions)
	assert.EqualValues(t, 1, len(conds), "ignore setFuncParamNames")
}

func TestExtractFiltersFailure(t *testing.T) {

	validJsonFields := []string{"a", "b", "c"}
	getOptions := FillGetOptions(GetOptions{})

	// invalid param key
	urlValues := url.Values{}
	urlValues.Add("d", "1")
	_, err := ExtractFilters(urlValues, validJsonFields, nil, getOptions)
	assert.EqualValues(t, "invalid filter field: d", err.Error())

	// empty param value
	urlValues = url.Values{}
	urlValues.Add("a", "")
	_, err = ExtractFilters(urlValues, validJsonFields, nil, getOptions)
	assert.EqualValues(t, "empty value in filter field: a", err.Error())
}

func TestExtractFormatSuccess(t *testing.T) {

	formatReqParamName := "xformat"

	// with fields param
	r := mustCreateGetReq(t, "/test?xformat=excel")
	format, err := ExtractFormat(r, formatReqParamName)
	if err != nil {
		t.Errorf("ExtractFormat failed: %v", err)
	}
	assert.EqualValues(t, "excel", format)

	// without format param (use json default)
	r = mustCreateGetReq(t, "/test")
	format, err = ExtractFormat(r, formatReqParamName)
	if err != nil {
		t.Errorf("ExtractFormat failed: %v", err)
	}
	assert.EqualValues(t, "json", format)
}

func TestExtractFormatFailure(t *testing.T) {

	formatReqParamName := "xformat"

	// invalid value
	r := mustCreateGetReq(t, "/test?xformat=a")
	_, err := ExtractFormat(r, formatReqParamName)
	assert.EqualValues(t, "xformat param value is invalid: a", err.Error())
}

func TestExtractPagingSuccess(t *testing.T) {

	pageReqParamName := "xpage"
	perPageReqParamName := "xper_page"
	defaultPerPage := 20
	maxPerPage := 500

	// with paging params
	r := mustCreateGetReq(t, "/test?xpage=2&xper_page=50")
	page, perPage, err := ExtractPaging(r, pageReqParamName, perPageReqParamName, defaultPerPage, maxPerPage)
	if err != nil {
		t.Errorf("ExtractPaging failed: %v", err)
	}
	assert.EqualValues(t, 2, page)
	assert.EqualValues(t, 50, perPage)

	// without paging params (use default)
	r = mustCreateGetReq(t, "/test")
	page, perPage, err = ExtractPaging(r, pageReqParamName, perPageReqParamName, defaultPerPage, maxPerPage)
	if err != nil {
		t.Errorf("ExtractPaging failed: %v", err)
	}
	assert.EqualValues(t, 1, page)
	assert.EqualValues(t, defaultPerPage, perPage)

	// enforcement of maxPerPage
	r = mustCreateGetReq(t, "/test?xper_page=1000000")
	_, perPage, err = ExtractPaging(r, pageReqParamName, perPageReqParamName, defaultPerPage, maxPerPage)
	if err != nil {
		t.Errorf("ExtractPaging failed: %v", err)
	}
	assert.EqualValues(t, maxPerPage, perPage)
}

func TestExtractPagingFailure(t *testing.T) {

	pageReqParamName := "xpage"
	perPageReqParamName := "xper_page"
	defaultPerPage := 20
	maxPerPage := 500

	// page wrong type
	r := mustCreateGetReq(t, "/test?xpage=a")
	_, _, err := ExtractPaging(r, pageReqParamName, perPageReqParamName, defaultPerPage, maxPerPage)
	assert.EqualValues(t, "xpage param not an integer", err.Error())

	// page invalid
	r = mustCreateGetReq(t, "/test?xpage=0")
	_, _, err = ExtractPaging(r, pageReqParamName, perPageReqParamName, defaultPerPage, maxPerPage)
	assert.EqualValues(t, "invalid xpage param", err.Error())

	// perPage wrong type
	r = mustCreateGetReq(t, "/test?xper_page=a")
	_, _, err = ExtractPaging(r, pageReqParamName, perPageReqParamName, defaultPerPage, maxPerPage)
	assert.EqualValues(t, "xper_page param not an integer", err.Error())

	// perPage invalid
	r = mustCreateGetReq(t, "/test?xper_page=0")
	_, _, err = ExtractPaging(r, pageReqParamName, perPageReqParamName, defaultPerPage, maxPerPage)
	assert.EqualValues(t, "invalid xper_page param", err.Error())
}

func TestExtractSetFuncParamValuesSuccess(t *testing.T) {

	setFuncParamNames := []string{"x", "y"}

	r := mustCreateGetReq(t, "/test?x=1&y=2")
	setFuncParamValues, err := ExtractSetFuncParamValues(r, setFuncParamNames)
	if err != nil {
		t.Errorf("ExtractSetFuncParamValues failed: %v", err)
	}
	assert.EqualValues(t, []string{"1", "2"}, setFuncParamValues)
}

func TestExtractSetFuncParamValuesFailure(t *testing.T) {

	setFuncParamNames := []string{"x", "y"}

	// missing a value
	r := mustCreateGetReq(t, "/test?x=1")
	_, err := ExtractSetFuncParamValues(r, setFuncParamNames)
	assert.EqualValues(t, "setFunc param name y is missing", err.Error())
}

func TestExtractSortsSuccess(t *testing.T) {

	validJsonFields := []string{"a", "b", "c"}
	sortReqParamName := "xsort"

	// with single ASC sort param
	r := mustCreateGetReq(t, "/test?xsort=a")
	sortCols, err := ExtractSorts(r, validJsonFields, sortReqParamName)
	if err != nil {
		t.Errorf("ExtractSorts failed: %v", err)
	}
	assert.EqualValues(t, []string{"a"}, sortCols)

	// with single DESC sort param
	r = mustCreateGetReq(t, "/test?xsort=-a")
	sortCols, err = ExtractSorts(r, validJsonFields, sortReqParamName)
	if err != nil {
		t.Errorf("ExtractSorts failed: %v", err)
	}
	assert.EqualValues(t, []string{"a DESC"}, sortCols)

	// with mixed sort params
	r = mustCreateGetReq(t, "/test?xsort=a,-b")
	sortCols, err = ExtractSorts(r, validJsonFields, sortReqParamName)
	if err != nil {
		t.Errorf("ExtractSorts failed: %v", err)
	}
	assert.EqualValues(t, []string{"a", "b DESC"}, sortCols)

	// without sort param (no default)
	r = mustCreateGetReq(t, "/test")
	sortCols, err = ExtractSorts(r, validJsonFields, sortReqParamName)
	if err != nil {
		t.Errorf("ExtractSorts failed: %v", err)
	}
	assert.EqualValues(t, []string(nil), sortCols)
}

func TestExtractSortsFailure(t *testing.T) {

	validJsonFields := []string{"a", "b"}
	sortReqParamName := "xsort"

	// invalid param value
	r := mustCreateGetReq(t, "/test?xsort=c")
	_, err := ExtractSorts(r, validJsonFields, sortReqParamName)
	assert.EqualValues(t, "invalid sort field: c", err.Error())
}
