package lys

import (
	"bytes"
	"net/http"
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

	rawBody := []byte(`{"a":1,"b":null}`)
	v, err := DecodeJsonBody[value](rawBody)
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
		D lystype.Time `json:"d"`
	}

	_, err := DecodeJsonBody[value](nil)
	assert.EqualValues(t, "body is missing", err.Error(), "body missing")

	_, err = DecodeJsonBody[value]([]byte(""))
	assert.EqualValues(t, "body is missing", err.Error(), "empty body")

	rawBody := []byte(`{"a:1}`)
	_, err = DecodeJsonBody[value](rawBody)
	assert.EqualValues(t, "request body contains badly-formed json", err.Error(), "malformed body")

	rawBody = []byte(`{-
		"a":1
	}`)
	_, err = DecodeJsonBody[value](rawBody)
	assert.EqualValues(t, "json syntax error: line: 1", err.Error(), "syntax error")

	rawBody = []byte(`{
		"a":"1"
	}`)
	_, err = DecodeJsonBody[value](rawBody)
	assert.EqualValues(t, "json type error: line: 2", err.Error(), "type error")

	rawBody = []byte(`{"x":1}`)
	_, err = DecodeJsonBody[value](rawBody)
	assert.EqualValues(t, "unknown field: x", err.Error(), "unknown field")

	rawBody = []byte(`{"c":"2024-01-aa"}`)
	_, err = DecodeJsonBody[value](rawBody)
	assert.EqualValues(t, "failed to parse a date or time: 2024-01-aa", err.Error(), "date parse error")

	rawBody = []byte(`{"c":"2021-06-282"}`)
	_, err = DecodeJsonBody[value](rawBody)
	assert.EqualValues(t, "failed to parse a date or time: parsing time \"2021-06-282\": extra text: \"2\"", err.Error(), "date parse error (extra text)")

	rawBody = []byte(`{"d":"22:61"}`)
	_, err = DecodeJsonBody[value](rawBody)
	assert.EqualValues(t, "failed to parse a date or time: parsing time \"22:61\": minute out of range", err.Error(), "time parse error (invalid minute)")
}

func TestExtractJsonBodySuccess(t *testing.T) {

	rawBody := []byte(`{"a":1,"b":""}`)
	req, err := http.NewRequest("GET", "", bytes.NewReader(rawBody))
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	body, err := ExtractJsonBody(req, 1024*1024)
	if err != nil {
		t.Fatalf("ExtractJsonBody failed: %v", err)
	}

	assert.EqualValues(t, rawBody, body)
}

func TestExtractJsonBodyFailure(t *testing.T) {

	// maxBodySize param missing
	_, err := ExtractJsonBody(&http.Request{}, 0)
	assert.EqualValues(t, "maxBodySize is zero", err.Error())

	// json header not set
	_, err = ExtractJsonBody(&http.Request{}, 1024*1024)
	assert.EqualValues(t, "content type must be application/json", err.Error())

	// body missing
	rawBody := []byte(``)
	req, err := http.NewRequest("GET", "", bytes.NewReader(rawBody))
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = ExtractJsonBody(req, 1024*1024)
	assert.EqualValues(t, "request body missing", err.Error())

	// invalid json
	rawBody = []byte(`"a":"b",`)
	req, err = http.NewRequest("GET", "", bytes.NewReader(rawBody))
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = ExtractJsonBody(req, 1024*1024)
	assert.EqualValues(t, "invalid json", err.Error())
}

func TestFindLineinJson(t *testing.T) {

	body := []byte(`{
		"a":1,
		"b":2
	}`)
	line := findLineinJson(body, 1)
	assert.EqualValues(t, 1, line, "line number should be 1")

	line = findLineinJson(body, 10)
	assert.EqualValues(t, 2, line, "line number should be 2")

	line = findLineinJson(body, 18)
	assert.EqualValues(t, 3, line, "line number should be 3")

	line = findLineinJson(body, len(body))
	assert.EqualValues(t, 4, line, "line number should be 4")
}
