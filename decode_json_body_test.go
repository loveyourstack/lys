package lys

import (
	"testing"

	"github.com/loveyourstack/lys/lystype"
	"github.com/stretchr/testify/assert"
)

func TestDecodeJsonBodySuccess(t *testing.T) {

	// Note: more complete type testing is done as part of post_test.go

	type value struct {
		A int     `json:"a"`
		B *string `json:"b"`
	}

	rawBody := `{"a":1,"b":null}`
	v, err := DecodeJsonBody[value]([]byte(rawBody))
	if err != nil {
		t.Fatalf("DecodeJsonBody failed: %v", err)
	}

	assert.EqualValues(t, 1, v.A)
	assert.Nil(t, v.B)
}

func TestDecodeJsonBodyFailure(t *testing.T) {

	type value struct {
		A int          `json:"a"`
		B *string      `json:"b"`
		C lystype.Date `json:"c"`
	}

	// body missing
	_, err := DecodeJsonBody[value](nil)
	assert.EqualValues(t, "body is missing", err.Error(), "body missing")

	// empty body
	_, err = DecodeJsonBody[value]([]byte(""))
	assert.EqualValues(t, "body is missing", err.Error(), "empty body")

	// malformed body
	rawBody := `{"a:1}`
	_, err = DecodeJsonBody[value]([]byte(rawBody))
	assert.EqualValues(t, "request body contains badly-formed json", err.Error(), "malformed body")

	// syntax error
	rawBody = `{"a":-}`
	_, err = DecodeJsonBody[value]([]byte(rawBody))
	assert.EqualValues(t, "json syntax error: line: 0", err.Error(), "syntax error")

	// type error
	rawBody = `{"a":"1"}`
	_, err = DecodeJsonBody[value]([]byte(rawBody))
	assert.EqualValues(t, "json type error: line: 0", err.Error(), "type error")

	// unknown field
	rawBody = `{"d":1}`
	_, err = DecodeJsonBody[value]([]byte(rawBody))
	assert.EqualValues(t, "unknown field: d", err.Error(), "unknown field")

	// time parse error
	rawBody = `{"c":"2024-01-aa"}`
	_, err = DecodeJsonBody[value]([]byte(rawBody))
	assert.EqualValues(t, "failed to parse a date or time: 2024-01-aa", err.Error(), "time parse error")
}
