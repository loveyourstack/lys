package lys

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFillGetOptionsDefaults(t *testing.T) {
	opts, err := FillGetOptions(GetOptions{})
	assert.NoError(t, err)

	assert.Equal(t, defaultFormatParamName, opts.FormatParamName)
	assert.Equal(t, defaultFieldsParamName, opts.FieldsParamName)
	assert.Equal(t, defaultPageParamName, opts.PageParamName)
	assert.Equal(t, defaultPerPageParamName, opts.PerPageParamName)
	assert.Equal(t, defaultSortParamName, opts.SortParamName)
	assert.Equal(t, defaultMultipleValueSeparator, opts.MultipleValueSeparator)
	assert.Equal(t, defaultMetadataSeparator, opts.MetadataSeparator)
	assert.Equal(t, defaultPerPage, opts.DefaultPerPage)
	assert.Equal(t, defaultMaxPerPage, opts.MaxPerPage)
	assert.Equal(t, defaultMaxFileRecs, opts.MaxFileRecs)
	assert.Equal(t, defaultCsvDelimiter, opts.CsvDelimiter)
}

func TestFillGetOptionsCustomValues(t *testing.T) {
	input := GetOptions{
		FormatParamName:        "fmt",
		FieldsParamName:        "flds",
		PageParamName:          "pg",
		PerPageParamName:       "pp",
		SortParamName:          "srt",
		MultipleValueSeparator: ",",
		MetadataSeparator:      "~",
		DefaultPerPage:         10,
		MaxPerPage:             100,
		MaxFileRecs:            500,
		CsvDelimiter:           ';',
	}

	opts, err := FillGetOptions(input)
	assert.NoError(t, err)

	// all custom values should be preserved unchanged
	assert.Equal(t, input.FormatParamName, opts.FormatParamName)
	assert.Equal(t, input.FieldsParamName, opts.FieldsParamName)
	assert.Equal(t, input.PageParamName, opts.PageParamName)
	assert.Equal(t, input.PerPageParamName, opts.PerPageParamName)
	assert.Equal(t, input.SortParamName, opts.SortParamName)
	assert.Equal(t, input.MultipleValueSeparator, opts.MultipleValueSeparator)
	assert.Equal(t, input.MetadataSeparator, opts.MetadataSeparator)
	assert.Equal(t, input.DefaultPerPage, opts.DefaultPerPage)
	assert.Equal(t, input.MaxPerPage, opts.MaxPerPage)
	assert.Equal(t, input.MaxFileRecs, opts.MaxFileRecs)
	assert.Equal(t, input.CsvDelimiter, opts.CsvDelimiter)
}

func TestFillGetOptionsDuplicateParamNames(t *testing.T) {
	// two param names the same
	_, err := FillGetOptions(GetOptions{
		FormatParamName: "xpage",
		PageParamName:   "xpage",
	})
	assert.Error(t, err)
}

func TestFillGetOptionsDuplicateSeparator(t *testing.T) {
	// separator collides with a param name
	_, err := FillGetOptions(GetOptions{
		FormatParamName:        "xformat",
		MultipleValueSeparator: "xformat",
	})
	assert.Error(t, err)
}

func TestFillGetOptionsPartialDefaults(t *testing.T) {
	// only some fields set: unset fields should get defaults, set fields should be preserved
	opts, err := FillGetOptions(GetOptions{
		DefaultPerPage: 5,
		MaxPerPage:     50,
	})
	assert.NoError(t, err)

	assert.Equal(t, 5, opts.DefaultPerPage)
	assert.Equal(t, 50, opts.MaxPerPage)
	assert.Equal(t, defaultFormatParamName, opts.FormatParamName)
	assert.Equal(t, defaultMaxFileRecs, opts.MaxFileRecs)
}
