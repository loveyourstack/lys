package lys

// GetOptions contains the options used when processing GET requests, such as paging param names and default values
// Since the json field names are used as filters, param names should be chosen which will never appear as json field names
type GetOptions struct {
	FormatParamName        string // param name to determine the output format of a GET request, e.g. "xformat"
	FieldsParamName        string // param name to limit the fields returned by a GET request, e.g. "xfields"
	PageParamName          string // param name to define the page offset returned by a paged GET request, e.g. "xpage"
	PerPageParamName       string // param name to define the number of records returned by a paged GET request, e.g. "xper_page"
	SortParamName          string // param name to identify the sort param used by a GET request, e.g. "xsort"
	MultipleValueSeparator string // the string used by a GET request to separate values in a filter where each value should be returned, e.g. "|", usage: "name=Bill|Sam"
	MetadataSeparator      string // the string used to separate any extra data appended to a GET request query filter, e.g. "^", usage: "sales=>100^Last 7 days"
	DefaultPerPage         int    // default number of results returned by a paged GET request, e.g. 20
	MaxPerPage             int    // max number of results returned per paged GET request, regardless of what the caller enters in the "PerPageParamName" param, e.g. 500
	MaxFileRecs            int    // max number of records contained in a file output
	CsvDelimiter           rune   // delimiter between values in CSV file output
}

// FillGetOptions returns input GetOptions if they are passed, and sets any unset fields to a sensible default value
func FillGetOptions(input GetOptions) (ret GetOptions) {
	ret = input

	if ret.FormatParamName == "" {
		ret.FormatParamName = "xformat"
	}
	if ret.FieldsParamName == "" {
		ret.FieldsParamName = "xfields"
	}
	if ret.PageParamName == "" {
		ret.PageParamName = "xpage"
	}
	if ret.PerPageParamName == "" {
		ret.PerPageParamName = "xper_page"
	}
	if ret.SortParamName == "" {
		ret.SortParamName = "xsort"
	}
	if ret.MultipleValueSeparator == "" {
		ret.MultipleValueSeparator = "|"
	}
	if ret.MetadataSeparator == "" {
		ret.MetadataSeparator = "^"
	}
	if ret.DefaultPerPage == 0 {
		ret.DefaultPerPage = 20
	}
	if ret.MaxPerPage == 0 {
		ret.MaxPerPage = 500
	}
	if ret.MaxFileRecs == 0 {
		ret.MaxFileRecs = 10000
	}
	if ret.CsvDelimiter == 0 {
		ret.CsvDelimiter = ','
	}

	return ret
}

// PostOptions contains the options used when processing POST or PUT requests
type PostOptions struct {
	MaxBodySize int64 // bytes
}

// FillPostOptions returns input PostOptions if they are passed, and sets any unset fields to a sensible default value
func FillPostOptions(input PostOptions) (ret PostOptions) {
	ret = input

	if ret.MaxBodySize == 0 {
		ret.MaxBodySize = 1024 * 1024 // 1 Mb
	}

	return ret
}
