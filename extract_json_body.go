package lys

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/loveyourstack/lys/lyserr"
)

// ExtractJsonBody reads and validates the body of the supplied request
func ExtractJsonBody(r *http.Request, maxBodySize int64) (body []byte, err error) {

	// check param
	if maxBodySize == 0 {
		return nil, fmt.Errorf("maxBodySize is zero")
	}

	// make sure Content-Type header is json
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, lyserr.User{
			Message: ErrDescInvalidContentType}
	}

	// read req body
	body, err = io.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll failed: %w", err)
	}

	// ensure there's a body
	if len(body) == 0 {
		return nil, lyserr.User{
			Message: ErrDescBodyMissing}
	}

	// ensure body is valid JSON
	if !json.Valid(body) {
		return nil, lyserr.User{
			Message: ErrDescInvalidJson}
	}

	return body, nil
}
