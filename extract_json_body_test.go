package lys

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractJsonBodySuccess(t *testing.T) {

	rawBody := `{"a":1,"b":""}`
	req, err := http.NewRequest("GET", "", bytes.NewReader([]byte(rawBody)))
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	body, err := ExtractJsonBody(req, 1024*1024)
	if err != nil {
		t.Fatalf("ExtractJsonBody failed: %v", err)
	}

	assert.EqualValues(t, `{"a":1,"b":null}`, string(body))
}

func TestExtractJsonBodyFailure(t *testing.T) {

	// maxBodySize param missing
	_, err := ExtractJsonBody(&http.Request{}, 0)
	assert.EqualValues(t, "maxBodySize is zero", err.Error())

	// json header not set
	_, err = ExtractJsonBody(&http.Request{}, 1024*1024)
	assert.EqualValues(t, "content type must be application/json", err.Error())

	// body missing
	rawBody := ``
	req, err := http.NewRequest("GET", "", bytes.NewReader([]byte(rawBody)))
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = ExtractJsonBody(req, 1024*1024)
	assert.EqualValues(t, "request body missing", err.Error())

	// invalid json
	rawBody = `"a":"b",`
	req, err = http.NewRequest("GET", "", bytes.NewReader([]byte(rawBody)))
	if err != nil {
		t.Fatalf("http.NewRequest failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	_, err = ExtractJsonBody(req, 1024*1024)
	assert.EqualValues(t, "invalid json", err.Error())
}
