package lys

import (
	"fmt"
	"net/http"

	"github.com/loveyourstack/lys/lysapi"
	"github.com/loveyourstack/lys/lyspg"
	"github.com/loveyourstack/lys/lysset"
)

type ExtractGetRequestModifierParams struct {
	AdditionalFilterParamNames lysset.Set[string]
	DbNames                    lysset.Set[string]
	GetOptions                 lysapi.GetOptions
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
	getReqModifiers.Format, err = lysapi.ExtractFormat(params.GetOptions.FormatParamName, r.FormValue(params.GetOptions.FormatParamName))
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractFormat failed: %w", err)
	}

	// filters (become WHERE clause conditions)
	getReqModifiers.Conditions, err = lysapi.ExtractFilters(r.URL.Query(), params.JsonKeyDbNameMap, params.AdditionalFilterParamNames, params.SetFuncUrlParamNames, params.GetOptions)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractFilters failed: %w", err)
	}

	// setFunc params (if setFunc is used, are passed as param values)
	getReqModifiers.SetFuncParamValues, err = lysapi.ExtractSetFuncParamValues(r, params.SetFuncUrlParamNames)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractSetFuncParamValues failed: %w", err)
	}

	// sorts (become ORDER BY)
	getReqModifiers.Sorts, err = lysapi.ExtractSorts(params.GetOptions.SortParamName, r.FormValue(params.GetOptions.SortParamName), params.JsonKeyDbNameMap)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractSorts failed: %w", err)
	}

	// skip fields and paging if outputting to file
	if getReqModifiers.Format != lysapi.FormatJson {
		return getReqModifiers, nil
	}

	// json output

	// fields (become column selection)
	getReqModifiers.Fields, err = lysapi.ExtractFields(params.GetOptions.FieldsParamName, r.FormValue(params.GetOptions.FieldsParamName), params.DbNames, params.JsonKeyDbNameMap)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractFields failed: %w", err)
	}

	// paging (become LIMIT and OFFSET)
	getReqModifiers.Page, getReqModifiers.PerPage, err = lysapi.ExtractPaging(params.GetOptions.PageParamName, r.FormValue(params.GetOptions.PageParamName),
		params.GetOptions.PerPageParamName, r.FormValue(params.GetOptions.PerPageParamName), params.GetOptions.DefaultPerPage, params.GetOptions.MaxPerPage)
	if err != nil {
		return GetReqModifiers{}, fmt.Errorf("ExtractPaging failed: %w", err)
	}

	return getReqModifiers, nil
}
