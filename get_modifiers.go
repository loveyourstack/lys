package lys

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/loveyourstack/lys/lyserr"
	"github.com/loveyourstack/lys/lyspg"
)

// output format consts
const (
	FormatCsv   string = "csv"
	FormatExcel string = "excel"
	FormatJson  string = "json"
)

var (
	ValidFormats = [...]string{FormatCsv, FormatExcel, FormatJson}
)

type ExtractGetRequestModifierParams struct {
	DbNames                    []string
	JsonKeyDbNameMap           map[string]string
	SetFuncUrlParamNames       []string
	AdditionalFilterParamNames []string
	GetOptions                 GetOptions
}

// GetReqModifiers contains data from a GET request's Url params which is used to modify a database SELECT statement
type GetReqModifiers struct {
	Format             string
	Fields             []string
	Conditions         []lyspg.Condition
	Page               int
	PerPage            int
	Sorts              []string
	SetFuncParamValues []any
}

// ExtractGetRequestModifiers reads the Url params of the supplied GET request and converts them into a GetReqModifiers
func ExtractGetRequestModifiers(r *http.Request, params ExtractGetRequestModifierParams) (getReqModifiers GetReqModifiers, err error) {

	// format (output format of GET req)
	getReqModifiers.Format, err = ExtractFormat(r, params.GetOptions.FormatParamName)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractFormat failed: %w", err)
	}

	// filters (become WHERE clause conditions)
	getReqModifiers.Conditions, err = ExtractFilters(r.URL.Query(), params.JsonKeyDbNameMap, params.AdditionalFilterParamNames, params.SetFuncUrlParamNames, params.GetOptions)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractFilters failed: %w", err)
	}

	// setFunc params (if setFunc is used, are passed as param values)
	getReqModifiers.SetFuncParamValues, err = ExtractSetFuncParamValues(r, params.SetFuncUrlParamNames)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractSetFuncParamValues failed: %w", err)
	}

	// sorts (become ORDER BY)
	getReqModifiers.Sorts, err = ExtractSorts(r, params.JsonKeyDbNameMap, params.GetOptions.SortParamName)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractSorts failed: %w", err)
	}

	// skip fields and paging if outputting to file
	if getReqModifiers.Format != FormatJson {
		return getReqModifiers, nil
	}

	// json output

	// fields (become column selection)
	getReqModifiers.Fields, err = ExtractFields(r, params.DbNames, params.JsonKeyDbNameMap, params.GetOptions.FieldsParamName)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractFields failed: %w", err)
	}

	// paging (become LIMIT and OFFSET)
	getReqModifiers.Page, getReqModifiers.PerPage, err = ExtractPaging(r, params.GetOptions.PageParamName, params.GetOptions.PerPageParamName,
		params.GetOptions.DefaultPerPage, params.GetOptions.MaxPerPage)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractPaging failed: %w", err)
	}

	return getReqModifiers, nil
}

// ExtractFormat returns the output format for the GET req
func ExtractFormat(r *http.Request, formatReqParamName string) (format string, err error) {

	// formatReqParamName: e.g. "xformat"
	// example: &xformat=excel

	formatVal := r.FormValue(formatReqParamName)

	// if no format param defined, default to json
	if formatVal == "" {
		return FormatJson, nil
	}

	// ensure value is valid
	if !slices.Contains(ValidFormats[:], formatVal) {
		return "", lyserr.User{
			Message: formatReqParamName + " param value is invalid: " + formatVal}
	}

	return formatVal, nil
}

// ExtractFields returns a slice of strings parsed from the request's fields param
func ExtractFields(r *http.Request, dbNames []string, jsonKeyDbNameMap map[string]string, fieldsReqParamName string) (fields []string, err error) {

	/*
	  fieldsReqParamName: e.g. "xfields"
	  inclusion example: &xfields=id,name
	  exclusion example: &xfields=-id,name
	*/

	// see if fields GET param exists
	fieldsRaw := r.FormValue(fieldsReqParamName)

	// if no fields param defined, return all db names
	if fieldsRaw == "" {
		return dbNames, nil
	}

	// check for inclusion or exclusion
	inclusion := true
	if fieldsRaw[:1] == "-" {
		inclusion = false
		fieldsRaw = fieldsRaw[1:]
	}

	// ensure correct format and valid fields

	// split by comma
	jsonVals := strings.Split(fieldsRaw, ",")
	dbVals := []string{}

	// for each field val
	for _, v := range jsonVals {

		// get corresponding db name
		dbVal, ok := jsonKeyDbNameMap[v]
		if !ok {
			return nil, lyserr.User{
				Message: fieldsReqParamName + " param value is invalid: " + v}
		}
		dbVals = append(dbVals, dbVal)
	}

	// if inclusion, fields are the dbVals
	if inclusion {
		return dbVals, nil
	}

	// exclusion: fields are all dbNames that are not in dbVals
	for _, f := range dbNames {
		if !slices.Contains(dbVals, f) {
			fields = append(fields, f)
		}
	}

	return fields, nil
}

// ExtractFilters returns a slice of conditions parsed from the request's params
// to get urlValues from a request: r.Url.Query()
func ExtractFilters(urlValues url.Values, jsonKeyDbNameMap map[string]string, additionalFilterParamNames, setFuncUrlParamNames []string, getOptions GetOptions) (conds []lyspg.Condition, err error) {

	// for each Url value
	for key, vals := range urlValues {

		// skip if this is a Url key assigned to one of the other purposes (.e.g paging, sorting) or is expected as a setFunc url param
		specialParams := []string{getOptions.FormatParamName, getOptions.FieldsParamName, getOptions.PageParamName, getOptions.PerPageParamName, getOptions.SortParamName}
		specialParams = append(specialParams, setFuncUrlParamNames...)
		if slices.Contains(specialParams, key) {
			continue
		}

		dbName := ""

		// if this is one of the additionalFilterParamNames, allow it even though it's not in the jsonKeyDbNameMap
		if slices.Contains(additionalFilterParamNames, key) {
			dbName = key
		} else {

			// get db name for this key
			dbVal, ok := jsonKeyDbNameMap[key]
			if !ok {
				return nil, lyserr.User{Message: "invalid filter field: " + key}
			}
			dbName = dbVal
		}

		// if same param is sent more than once, treat each one as a separate filter param
		for _, val := range vals {

			// check for empty value, e.g. "&x" or "&x="
			if val == "" {
				return nil, lyserr.User{Message: "empty value in filter field: " + key}
			}

			// create condition from this filter
			cond := processFilterParam(dbName, val, getOptions)
			conds = append(conds, cond)
		}
	}

	return conds, nil
}

// processFilterParam returns a condition parsed from a single Url filter param
func processFilterParam(dbName, rawValue string, getOptions GetOptions) (cond lyspg.Condition) {

	cond.Field = dbName

	// check for presence of appended metadata in rawValue
	if strings.Contains(rawValue, getOptions.MetadataSeparator) {

		rawValueA := strings.Split(rawValue, getOptions.MetadataSeparator)
		rawValue = rawValueA[0]

		// if metadata is not empty, add it to cond
		if len(rawValueA) > 1 {
			cond.Metadata = strings.Join(rawValueA[1:], "")
		}
	}

	// extract the operator and target value. Note: <= and >= are not allowed, since "=" is reserved for key/value separation
	switch {

	// greater than or equals (>eq at start)
	case len(rawValue) > 3 && rawValue[:3] == ">eq":
		cond.Operator = lyspg.OpGreaterThanEquals
		cond.Value = rawValue[3:]

	// greater than (> at start)
	case len(rawValue) > 1 && rawValue[:1] == ">":
		cond.Operator = lyspg.OpGreaterThan
		cond.Value = rawValue[1:]

	// less than or equals (<eq at start)
	case len(rawValue) > 3 && rawValue[:3] == "<eq":
		cond.Operator = lyspg.OpLessThanEquals
		cond.Value = rawValue[3:]

	// less than (< at start)
	case len(rawValue) > 1 && rawValue[:1] == "<":
		cond.Operator = lyspg.OpLessThan
		cond.Value = rawValue[1:]

	// containsAny (~[ at start, ]~ at end, values separated by MultipleValueSeparator)
	case len(rawValue) > 4 && (rawValue[:2] == "~[" && rawValue[len(rawValue)-2:] == "]~"):
		cond.Operator = lyspg.OpContainsAny
		cond.InValues = strings.Split(rawValue[2:len(rawValue)-2], getOptions.MultipleValueSeparator)

	// not contains (!~ at start, ~ at end)
	case len(rawValue) > 3 && (rawValue[:2] == "!~" && rawValue[len(rawValue)-1:] == "~"):
		cond.Operator = lyspg.OpNotContains
		cond.Value = strings.Replace(rawValue[1:], "~", "", -1)

	// contains (~ at start and end)
	case len(rawValue) > 2 && (rawValue[:1] == "~" && rawValue[len(rawValue)-1:] == "~"):
		cond.Operator = lyspg.OpContains
		cond.Value = strings.Replace(rawValue, "~", "", -1)

	// starts with (~ at end)
	case len(rawValue) > 1 && rawValue[len(rawValue)-1:] == "~":
		cond.Operator = lyspg.OpStartsWith
		cond.Value = strings.Replace(rawValue, "~", "", -1)

	// end with (~ at start)
	case len(rawValue) > 1 && rawValue[:1] == "~":
		cond.Operator = lyspg.OpEndsWith
		cond.Value = strings.Replace(rawValue, "~", "", -1)

	// empty (={empty})
	case len(rawValue) > 1 && rawValue == "{empty}":
		cond.Operator = lyspg.OpEmpty
		cond.Value = "0"

	// not empty (={!empty})
	case len(rawValue) > 1 && rawValue == "{!empty}":
		cond.Operator = lyspg.OpNotEmpty
		cond.Value = "0"

	// is null (={null})
	case len(rawValue) > 1 && rawValue == "{null}":
		cond.Operator = lyspg.OpNull
		cond.Value = ""

	// is not null (={!null})
	case len(rawValue) > 1 && rawValue == "{!null}":
		cond.Operator = lyspg.OpNotNull
		cond.Value = ""

	// not equals (! at start, no MultipleValueSeparator)
	case len(rawValue) > 1 && rawValue[:1] == "!" && !strings.ContainsAny(rawValue, getOptions.MultipleValueSeparator):
		cond.Operator = lyspg.OpNotEquals
		cond.Value = rawValue[1:]

	// not in (! at start, contains the MultipleValueSeparator)
	case len(rawValue) > 2 && rawValue[:1] == "!" && strings.ContainsAny(rawValue, getOptions.MultipleValueSeparator):
		cond.Operator = lyspg.OpNotIn
		cond.InValues = strings.Split(rawValue[1:], getOptions.MultipleValueSeparator)

	// in (contains the MultipleValueSeparator - treat each MultipleValueSeparator-separated value as part of the IN clause)
	case len(rawValue) > 1 && strings.ContainsAny(rawValue, getOptions.MultipleValueSeparator):
		cond.Operator = lyspg.OpIn
		cond.InValues = strings.Split(rawValue, getOptions.MultipleValueSeparator)

	// assume "="
	default:
		cond.Operator = lyspg.OpEquals
		cond.Value = rawValue
	}

	return cond
}

// ExtractPaging returns paging variables parsed from a request's paging params
// page defaults to 1, perPage defaults to defaultPerPage
func ExtractPaging(r *http.Request, pageReqParamName, perPageReqParamName string, defaultPerPage, maxPerPage int) (page int, perPage int, err error) {

	// pageReqParamName: e.g. "xpage"
	// perPageReqParamName: e.g. "xper_page"

	// see if page GET param exists, and if so, ensure it is an integer >= 1
	pageParam := r.FormValue(pageReqParamName)
	page = 1
	if pageParam != "" {
		page, err = strconv.Atoi(pageParam)
		if err != nil {
			return 0, 0, lyserr.User{
				Message: pageReqParamName + " param not an integer"}
		}
		if page < 1 {
			return 0, 0, lyserr.User{
				Message: "invalid " + pageReqParamName + " param"}
		}
	}

	// see if per_page GET param exists, and if so, ensure it is an integer >= 1
	perPageParam := r.FormValue(perPageReqParamName)
	perPage = defaultPerPage
	if perPageParam != "" {
		perPage, err = strconv.Atoi(perPageParam)
		if err != nil {
			return 0, 0, lyserr.User{
				Message: perPageReqParamName + " param not an integer"}
		}
		if perPage < 1 {
			return 0, 0, lyserr.User{
				Message: "invalid " + perPageReqParamName + " param"}
		}
	}

	// don't allow perPage to exceed the max
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return page, perPage, nil
}

// ExtractSorts returns an array of SQL sorting statements parsed from the request's sort param
func ExtractSorts(r *http.Request, jsonKeyDbNameMap map[string]string, sortReqParamName string) (sortCols []string, err error) {

	// sortReqParamName: e.g. "xsort"
	// format: xsort=entry_by,-name (use "-" (minus) for a DESC sort)

	// see if sort GET param exists, and if so, ensure correct format and valid fields
	sortRaw := r.FormValue(sortReqParamName)
	if sortRaw != "" {

		// split by comma into array
		sortVals := strings.Split(sortRaw, ",")

		for _, v := range sortVals {

			fieldName := ""
			sortDesc := ""
			if v[:1] == "-" {
				fieldName = v[1:]
				sortDesc = " DESC"
			} else {
				fieldName = v
			}

			// get corresponding db name for this field
			dbName, ok := jsonKeyDbNameMap[fieldName]
			if !ok {
				return nil, lyserr.User{
					Message: "invalid sort field: " + fieldName}
			}

			// add sortCol
			sortCols = append(sortCols, dbName+sortDesc)
		}
	}

	return sortCols, nil
}

// ExtractSetFuncParamValues returns the values to be passed to the SQL setFunc
// each param is currently treated as mandatory
func ExtractSetFuncParamValues(r *http.Request, setFuncUrlParamNames []string) (setFuncUrlParamValues []any, err error) {

	for _, paramName := range setFuncUrlParamNames {

		paramValue := r.FormValue(paramName)
		if paramValue == "" {
			return nil, lyserr.User{Message: fmt.Sprintf("setFunc param name %s is missing", paramName)}
		}

		setFuncUrlParamValues = append(setFuncUrlParamValues, paramValue)
	}

	return setFuncUrlParamValues, nil
}
