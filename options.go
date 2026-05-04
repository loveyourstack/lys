package lys

import (
	"fmt"

	"github.com/loveyourstack/lys/lysslices"
)

const (
	defaultFieldsParamName  string = "xfields"
	defaultFormatParamName  string = "xformat"
	defaultPageParamName    string = "xpage"
	defaultPerPageParamName string = "xper_page"
	defaultSortParamName    string = "xsort"

	defaultMultipleValueSeparator string = "|"
	defaultMetadataSeparator      string = "^"

	defaultPerPage    int = 20
	defaultMaxPerPage int = 500

	defaultMaxFileRecs  int  = 10000
	defaultCsvDelimiter rune = ','

	defaultMaxBodySize int64 = 1024 * 1024 // 1 Mb
)

// GetOptions contains the options used when processing GET requests, such as paging param names and default values.
// Since the json field names are used as filters, param names should be chosen which will never appear as json field names. This is the reason for the "x" prefix in the default param names.
type GetOptions struct {

	// special param names
	FieldsParamName  string // name of the param which limits the fields returned by a GET request, e.g. "xfields=name,age"
	FormatParamName  string // name of the param which determines the output format of a GET request, e.g. "xformat=csv"
	PageParamName    string // name of the param which defines the page offset returned by a paged GET request, e.g. "xpage=1"
	PerPageParamName string // name of the param which defines the number of records returned by a paged GET request, e.g. "xper_page=20"
	SortParamName    string // name of the param which sorts the records returned by a GET request, e.g. "xsort=name,-age"

	// separators
	MultipleValueSeparator string // the string used by a GET request to separate values in a filter where each value should be returned, e.g. "|", usage: "name=Bill|Sam"
	MetadataSeparator      string // the string used to separate any extra data appended to a GET request query filter, e.g. "^", usage: "sales=>100^Last 7 days"

	// paging limits
	DefaultPerPage int // default number of results returned by a paged GET request, e.g. 20
	MaxPerPage     int // max number of results returned per paged GET request, regardless of what the caller enters in the "PerPageParamName" param, e.g. 500

	// output config
	MaxFileRecs  int  // max number of records contained in a file output
	CsvDelimiter rune // delimiter between values in CSV file output. 0 means not set, and the default will be used.
}

// FillGetOptions returns input GetOptions if they are passed, and sets any unset fields to a sensible default value
func FillGetOptions(input GetOptions) (ret GetOptions, err error) {

	ret = input

	if ret.FieldsParamName == "" {
		ret.FieldsParamName = defaultFieldsParamName
	}
	if ret.FormatParamName == "" {
		ret.FormatParamName = defaultFormatParamName
	}
	if ret.PageParamName == "" {
		ret.PageParamName = defaultPageParamName
	}
	if ret.PerPageParamName == "" {
		ret.PerPageParamName = defaultPerPageParamName
	}
	if ret.SortParamName == "" {
		ret.SortParamName = defaultSortParamName
	}
	if ret.MultipleValueSeparator == "" {
		ret.MultipleValueSeparator = defaultMultipleValueSeparator
	}
	if ret.MetadataSeparator == "" {
		ret.MetadataSeparator = defaultMetadataSeparator
	}
	if ret.DefaultPerPage == 0 {
		ret.DefaultPerPage = defaultPerPage
	}
	if ret.MaxPerPage == 0 {
		ret.MaxPerPage = defaultMaxPerPage
	}
	if ret.MaxFileRecs == 0 {
		ret.MaxFileRecs = defaultMaxFileRecs
	}
	if ret.CsvDelimiter == 0 {
		ret.CsvDelimiter = defaultCsvDelimiter
	}

	// param names and separators must be unique
	dups := lysslices.ReportDuplicates([]string{
		ret.FieldsParamName,
		ret.FormatParamName,
		ret.PageParamName,
		ret.PerPageParamName,
		ret.SortParamName,
		ret.MultipleValueSeparator,
		ret.MetadataSeparator,
	})
	if len(dups) > 0 {
		return ret, fmt.Errorf("param names and separators must be unique: duplicate values: %v", dups)
	}

	return ret, nil
}

// PostOptions contains the options used when processing POST or PUT requests
type PostOptions struct {
	MaxBodySize int64 // bytes
}

// FillPostOptions returns input PostOptions if they are passed, and sets any unset fields to a sensible default value
func FillPostOptions(input PostOptions) (ret PostOptions) {
	ret = input

	if ret.MaxBodySize == 0 {
		ret.MaxBodySize = defaultMaxBodySize
	}

	return ret
}
