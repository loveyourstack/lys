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
	"github.com/loveyourstack/lys/lysset"
)

// output format consts
const (
	FormatCsv   string = "csv"
	FormatExcel string = "excel"
	FormatJson  string = "json"
)

var (
	ValidFormats = []string{FormatCsv, FormatExcel, FormatJson}
)

type ExtractGetRequestModifierParams struct {
	AdditionalFilterParamNames lysset.Set[string]
	DbNames                    lysset.Set[string]
	GetOptions                 GetOptions
	JsonKeyDbNameMap           map[string]string
	SetFuncUrlParamNames       []string
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
	getReqModifiers.Format, err = ExtractFormat(params.GetOptions.FormatParamName, r.FormValue(params.GetOptions.FormatParamName))
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
	getReqModifiers.Sorts, err = ExtractSorts(params.GetOptions.SortParamName, r.FormValue(params.GetOptions.SortParamName), params.JsonKeyDbNameMap)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractSorts failed: %w", err)
	}

	// skip fields and paging if outputting to file
	if getReqModifiers.Format != FormatJson {
		return getReqModifiers, nil
	}

	// json output

	// fields (become column selection)
	getReqModifiers.Fields, err = ExtractFields(params.GetOptions.FieldsParamName, r.FormValue(params.GetOptions.FieldsParamName), params.DbNames, params.JsonKeyDbNameMap)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractFields failed: %w", err)
	}

	// paging (become LIMIT and OFFSET)
	getReqModifiers.Page, getReqModifiers.PerPage, err = ExtractPaging(params.GetOptions.PageParamName, r.FormValue(params.GetOptions.PageParamName),
		params.GetOptions.PerPageParamName, r.FormValue(params.GetOptions.PerPageParamName), params.GetOptions.DefaultPerPage, params.GetOptions.MaxPerPage)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractPaging failed: %w", err)
	}

	return getReqModifiers, nil
}

// ExtractFormat returns the output format for the GET req
func ExtractFormat(formatParamName, formatVal string) (format string, err error) {

	// formatParamName: e.g. "xformat"
	// example: &xformat=excel

	// if no format param defined, default to json
	if formatVal == "" {
		return FormatJson, nil
	}

	// ensure value is valid
	if !slices.Contains(ValidFormats, formatVal) {
		return "", lyserr.User{
			Message: formatParamName + " param value is invalid: " + formatVal}
	}

	return formatVal, nil
}

// ExtractFields returns a slice of strings parsed from the request's fields param
func ExtractFields(fieldsParamName, fieldsVal string, dbNames lysset.Set[string], jsonKeyDbNameMap map[string]string) (fields []string, err error) {

	/*
	  fieldsParamName: e.g. "xfields"
	  inclusion example: &xfields=id,name
	  exclusion example: &xfields=-id,name
	*/

	// if no fields param defined, return sorted db names
	if fieldsVal == "" {
		fields = dbNames.Values()
		slices.Sort(fields)
		return fields, nil
	}

	// check for inclusion or exclusion
	inclusion := true
	if fieldsVal[:1] == "-" {
		inclusion = false
		fieldsVal = fieldsVal[1:]
	}

	// ensure correct format and valid fields

	// split by comma
	jsonVals := strings.Split(fieldsVal, ",")
	dbVals := lysset.New[string]()

	// for each field val
	for _, v := range jsonVals {

		// get corresponding db name
		dbVal, ok := jsonKeyDbNameMap[v]
		if !ok {
			return nil, lyserr.User{
				Message: fieldsParamName + " param value is invalid: " + v}
		}
		dbVals.Add(dbVal)
	}

	// if inclusion, fields are the dbVals
	if inclusion {
		fields = dbVals.Values()
		slices.Sort(fields)
		return fields, nil
	}

	// exclusion: fields are all dbNames that are not in dbVals
	for f := range dbNames {
		if !dbVals.Contains(f) {
			fields = append(fields, f)
		}
	}

	slices.Sort(fields)
	return fields, nil
}

// ExtractFilters returns a slice of conditions parsed from the request's params
// to get urlValues from a request: r.Url.Query()
func ExtractFilters(urlValues url.Values, jsonKeyDbNameMap map[string]string, additionalFilterParamNames lysset.Set[string], setFuncUrlParamNames []string, getOptions GetOptions) (conds []lyspg.Condition, err error) {

	// define special param names which have another purpose and may not be used as filter keys
	specialParams := lysset.New(getOptions.FormatParamName, getOptions.FieldsParamName, getOptions.PageParamName, getOptions.PerPageParamName, getOptions.SortParamName)
	specialParams.AddAll(setFuncUrlParamNames...)

	// for each Url value
	for key, vals := range urlValues {

		// skip if this is a special param
		if specialParams.Contains(key) {
			continue
		}

		dbName := ""

		// if this is one of the additionalFilterParamNames, allow it even though it's not in the jsonKeyDbNameMap
		if additionalFilterParamNames.Contains(key) {
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

	// contains any (~[ at start, ]~ at end, values separated by MultipleValueSeparator)
	case len(rawValue) > 4 && (rawValue[:2] == "~[" && rawValue[len(rawValue)-2:] == "]~"):
		cond.Operator = lyspg.OpContainsAny
		cond.InValues = strings.Split(rawValue[2:len(rawValue)-2], getOptions.MultipleValueSeparator)

	// not contains (!~ at start, ~ at end)
	case len(rawValue) > 3 && (rawValue[:2] == "!~" && rawValue[len(rawValue)-1:] == "~"):
		cond.Operator = lyspg.OpNotContains
		cond.Value = rawValue[2 : len(rawValue)-1]

	// contains (~ at start and end)
	case len(rawValue) > 2 && (rawValue[:1] == "~" && rawValue[len(rawValue)-1:] == "~"):
		cond.Operator = lyspg.OpContains
		cond.Value = rawValue[1 : len(rawValue)-1]

	// starts with (~ at end)
	case len(rawValue) > 1 && rawValue[len(rawValue)-1:] == "~":
		cond.Operator = lyspg.OpStartsWith
		cond.Value = rawValue[:len(rawValue)-1]

	// end with (~ at start)
	case len(rawValue) > 1 && rawValue[:1] == "~":
		cond.Operator = lyspg.OpEndsWith
		cond.Value = rawValue[1:]

	// empty (={empty})
	case rawValue == "{empty}":
		cond.Operator = lyspg.OpEmpty
		cond.Value = "0"

	// not empty (={!empty})
	case rawValue == "{!empty}":
		cond.Operator = lyspg.OpNotEmpty
		cond.Value = "0"

	// is null (={null})
	case rawValue == "{null}":
		cond.Operator = lyspg.OpNull
		cond.Value = ""

	// is not null (={!null})
	case rawValue == "{!null}":
		cond.Operator = lyspg.OpNotNull
		cond.Value = ""

	// not equals (! at start, no MultipleValueSeparator)
	case len(rawValue) > 1 && rawValue[:1] == "!" && !strings.Contains(rawValue, getOptions.MultipleValueSeparator):
		cond.Operator = lyspg.OpNotEquals
		cond.Value = rawValue[1:]

	// not in (! at start, contains the MultipleValueSeparator)
	case len(rawValue) > 2 && rawValue[:1] == "!" && strings.Contains(rawValue, getOptions.MultipleValueSeparator):
		cond.Operator = lyspg.OpNotIn
		cond.InValues = strings.Split(rawValue[1:], getOptions.MultipleValueSeparator)

	// in (contains the MultipleValueSeparator)
	case len(rawValue) > 1 && strings.Contains(rawValue, getOptions.MultipleValueSeparator):
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
func ExtractPaging(pageParamName, pageVal, perPageParamName, perPageVal string, defaultPerPage, maxPerPage int) (page int, perPage int, err error) {

	// pageParamName: e.g. "xpage"
	// perPageParamName: e.g. "xper_page"

	if maxPerPage < 1 {
		return 0, 0, fmt.Errorf("maxPerPage must be >= 1")
	}

	// see if page GET param exists, and if so, ensure it is an integer >= 1
	page = 1
	if pageVal != "" {
		page, err = strconv.Atoi(pageVal)
		if err != nil {
			return 0, 0, lyserr.User{
				Message: pageParamName + " param value is not an integer"}
		}
		if page < 1 {
			return 0, 0, lyserr.User{
				Message: pageParamName + " param value must be >= 1"}
		}
	}

	// see if per_page GET param exists, and if so, ensure it is an integer >= 1
	perPage = defaultPerPage
	if perPageVal != "" {
		perPage, err = strconv.Atoi(perPageVal)
		if err != nil {
			return 0, 0, lyserr.User{
				Message: perPageParamName + " param value is not an integer"}
		}
		if perPage < 1 {
			return 0, 0, lyserr.User{
				Message: perPageParamName + " param value must be >= 1"}
		}
	}

	// don't allow perPage to exceed the max
	if perPage > maxPerPage {
		perPage = maxPerPage
	}

	return page, perPage, nil
}

// ExtractSorts returns an array of SQL sorting statements parsed from the request's sort param
func ExtractSorts(sortParamName, sortVal string, jsonKeyDbNameMap map[string]string) (sortCols []string, err error) {

	// sortParamName: e.g. "xsort"
	// format: xsort=entry_by,-name (use "-" (minus) for a DESC sort)

	// see if sort GET param exists, and if so, ensure correct format and valid fields
	if sortVal != "" {

		// split by comma into array
		sortVals := strings.Split(sortVal, ",")

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
				return nil, lyserr.User{Message: sortParamName + " has invalid field: " + fieldName}
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
			return nil, lyserr.User{Message: fmt.Sprintf("setFunc param name '%s' is missing", paramName)}
		}

		setFuncUrlParamValues = append(setFuncUrlParamValues, paramValue)
	}

	return setFuncUrlParamValues, nil
}
