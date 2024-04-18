package lys

// GetOptions contains the options used when processing GET requests, such as paging param names and default values
// Since the json field names are used as filters, param names should be chosen which will never appear as json field names
type GetOptions struct {
	FieldsParamName        string // param name to limit the fields returned by a GET request, e.g. "xfields"
	PageParamName          string // param name to define the page offset returned by a GET request, e.g. "xpage"
	PerPageParamName       string // param name to define the number of records returned by a GET request, e.g. "xper_page"
	SortParamName          string // param name to identify the sort param used by a GET request, e.g. "xsort"
	MultipleValueSeparator string // the string used by a GET request to separate values in a filter where each value should be returned, e.g. "|", usage: "name=Bill|Sam"
	DefaultPerPage         int    // default number of results returned by a GET request, e.g. 20
	DefaultMaxPerPage      int    // default max number of results returned per GET request, regardless of what the caller enters in the "PerPageParamName" param, e.g. 500
}

// FillGetOptions returns input GetOptions if they are passed, and sets any unset fields to a sensible default value
func FillGetOptions(input GetOptions) (ret GetOptions) {
	if input.FieldsParamName == "" {
		ret.FieldsParamName = "xfields"
	}
	if input.PageParamName == "" {
		ret.PageParamName = "xpage"
	}
	if input.PerPageParamName == "" {
		ret.PerPageParamName = "xper_page"
	}
	if input.SortParamName == "" {
		ret.SortParamName = "xsort"
	}
	if input.MultipleValueSeparator == "" {
		ret.MultipleValueSeparator = "|"
	}
	if input.DefaultPerPage == 0 {
		ret.DefaultPerPage = 20
	}
	if input.DefaultMaxPerPage == 0 {
		ret.DefaultMaxPerPage = 500
	}

	return ret
}

// PostOptions contains the options used when processing POST or PUT requests
type PostOptions struct {
	MaxBodySize int64 // bytes
}

// FillPostOptions returns input PostOptions if they are passed, and sets any unset fields to a sensible default value
func FillPostOptions(input PostOptions) (ret PostOptions) {
	if input.MaxBodySize == 0 {
		ret.MaxBodySize = 1024 * 1024 // 1 Mb
	}

	return ret
}
