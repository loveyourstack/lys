package lys

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/loveyourstack/lys/lyserr"
)

// DecodeJsonBody decodes the supplied json body into dest and checks for a variety of error conditions.
// Caller should check that body is valid JSON and should enforce a maximum body size (usually done in ExtractJsonBody).
// Adapted from https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body.
func DecodeJsonBody[T any](body []byte) (dest T, err error) {

	if len(body) == 0 {
		return dest, lyserr.User{Message: "body is missing"}
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()

	if err = dec.Decode(&dest); err != nil {

		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError
		var timeParseErr *time.ParseError

		switch {
		case errors.Is(err, io.ErrUnexpectedEOF):
			return dest, lyserr.User{Message: "request body contains badly-formed json"}

		case errors.As(err, &syntaxErr):
			line := findLineinJson(body, int(syntaxErr.Offset))
			return dest, lyserr.User{Message: "json syntax error: line: " + strconv.Itoa(line)}

		case errors.As(err, &unmarshalTypeErr):
			line := findLineinJson(body, int(unmarshalTypeErr.Offset))
			return dest, lyserr.User{Message: "json type error: line: " + strconv.Itoa(line)}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return dest, lyserr.User{Message: "unknown field: " + strings.Trim(fieldName, `"`)}

		case strings.HasSuffix(err.Error(), "unable to parse IP"):
			return dest, lyserr.User{Message: "failed to parse IP address"}

		case errors.As(err, &timeParseErr):
			msg := timeParseErr.Error()
			asLoc := strings.Index(msg, " as ")
			const longMsgLen = 13
			if asLoc > longMsgLen && len(msg) > longMsgLen+1 {
				return dest, lyserr.User{Message: "failed to parse a date or time: " + strings.Trim(msg[longMsgLen:asLoc], "\\\"")}
			}
			return dest, lyserr.User{Message: "failed to parse a date or time: " + msg}

		default:
			return dest, fmt.Errorf("dec.Decode failed: %w", err)
		}
	}

	return dest, nil
}

// ExtractJsonBody reads and validates the body of the supplied request.
func ExtractJsonBody(r *http.Request, maxBodySize int64) (body []byte, err error) {

	// check param
	if maxBodySize == 0 {
		return nil, fmt.Errorf("maxBodySize is zero")
	}

	// make sure Content-Type header is json
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return nil, ErrInvalidContentType
	}

	defer r.Body.Close()

	// read req body
	body, err = io.ReadAll(io.LimitReader(r.Body, maxBodySize))
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll failed: %w", err)
	}

	// ensure there's a body
	if len(body) == 0 {
		return nil, ErrBodyMissing
	}

	// ensure body is valid JSON
	if !json.Valid(body) {
		return nil, ErrInvalidJson
	}

	return body, nil
}

// findLineinJson returns the line number in a json body corresponding to a given offset. This is used to provide more helpful error messages when json decoding fails.
func findLineinJson(body []byte, offset int) (line int) {
	return bytes.Count(body[:offset], []byte("\n")) + 1
}
