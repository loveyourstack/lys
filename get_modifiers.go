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

// GetReqModifiers contains data from a GET request's Url params which is used to modify a database SELECT statement
type GetReqModifiers struct {
	Format     string
	Fields     []string
	Conditions []lyspg.Condition
	Page       int
	PerPage    int
	Sorts      []string
}

// ExtractGetRequestModifiers reads the Url params of the supplied GET request and converts them into a GetReqModifiers
func ExtractGetRequestModifiers(r *http.Request, validJsonFields []string, getOptions GetOptions) (getReqModifiers GetReqModifiers, err error) {

	// format
	getReqModifiers.Format, err = ExtractFormat(r, getOptions.FormatParamName)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractFormat failed: %w", err)
	}

	// filters (become WHERE clause conditions)
	getReqModifiers.Conditions, err = ExtractFilters(r.URL.Query(), validJsonFields, getOptions)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractFilters failed: %w", err)
	}

	// sorts (become ORDER BY)
	getReqModifiers.Sorts, err = ExtractSorts(r, validJsonFields, getOptions.SortParamName)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractSorts failed: %w", err)
	}

	// skip fields and paging if outputting to file
	if getReqModifiers.Format != FormatJson {
		return getReqModifiers, nil
	}

	// json output

	// fields (become column selection)
	getReqModifiers.Fields, err = ExtractFields(r, validJsonFields, getOptions.FieldsParamName)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractFields failed: %w", err)
	}

	// paging (become LIMIT and OFFSET)
	getReqModifiers.Page, getReqModifiers.PerPage, err = ExtractPaging(r, getOptions.PageParamName, getOptions.PerPageParamName, getOptions.DefaultPerPage, getOptions.MaxPerPage)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractPaging failed: %w", err)
	}

	return getReqModifiers, nil
}

// ExtractFormat returns
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
func ExtractFields(r *http.Request, validJsonFields []string, fieldsReqParamName string) (fields []string, err error) {

	// fieldsReqParamName: e.g. "xfields"
	// example: &xfields=id,name

	// see if fields GET param exists, and if so, ensure correct format and valid fields
	fieldsRaw := r.FormValue(fieldsReqParamName)
	if fieldsRaw != "" {

		// split by comma
		fieldVals := strings.Split(fieldsRaw, ",")

		// for each field val
		for _, v := range fieldVals {

			// ensure value is a valid json field
			if !slices.Contains(validJsonFields, v) {
				return nil, lyserr.User{
					Message: fieldsReqParamName + " param value is invalid: " + v}
			}

			// add field
			fields = append(fields, v)
		}
	} else {
		// if no fields param defined, return all
		fields = validJsonFields
	}

	return fields, nil
}

// ExtractFilters returns a slice of conditions parsed from the request's params
// to get urlValues from a request: r.Url.Query()
func ExtractFilters(urlValues url.Values, validJsonFields []string, getOptions GetOptions) (conds []lyspg.Condition, err error) {

	// for each Url value
	for key, vals := range urlValues {

		// skip if this is a Url key assigned to one of the other purposes (.e.g paging, sorting)
		specialParams := []string{getOptions.FormatParamName, getOptions.FieldsParamName, getOptions.PageParamName, getOptions.PerPageParamName, getOptions.SortParamName}
		if slices.Contains(specialParams, key) {
			continue
		}

		// if same param is sent more than once, treat each one as a separate filter param
		for _, val := range vals {

			// create condition from this filter
			cond, err := processFilterParam(key, val, validJsonFields, getOptions)
			if err != nil {
				// don't wrap err: only user errors are returned
				return nil, err
			}
			conds = append(conds, cond)
		}
	}

	return conds, nil
}

// processFilterParam returns a condition parsed from a single Url filter param
func processFilterParam(field, rawValue string, validJsonFields []string, getOptions GetOptions) (cond lyspg.Condition, err error) {

	// ensure field is valid
	if !slices.Contains(validJsonFields, field) {
		return lyspg.Condition{}, lyserr.User{
			Message: "invalid filter field: " + field}
	}

	// check for empty value, e.g. "&x" or "&x="
	if rawValue == "" {
		return lyspg.Condition{}, lyserr.User{
			Message: "empty value in filter field: " + field}
	}

	cond.Field = field

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

	return cond, nil
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
func ExtractSorts(r *http.Request, validJsonFields []string, sortReqParamName string) (sortCols []string, err error) {

	// sortReqParamName: e.g. "xsort"

	// format: sort=entry_by,-name (use "-" (minus) for a DESC sort)

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

			// ensure field is valid
			if !slices.Contains(validJsonFields, fieldName) {
				return nil, lyserr.User{
					Message: "invalid sort field: " + fieldName}
			}

			// add sortCol
			sortCols = append(sortCols, fieldName+sortDesc)
		}
	}

	return sortCols, nil
}
