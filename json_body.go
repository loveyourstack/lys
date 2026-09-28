package lys

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/netip"
	"reflect"
	"strings"
	"time"

	"github.com/loveyourstack/lys/lyserr"
)

// DecodeJsonBody decodes the supplied json body into dest and checks for a variety of error conditions.
// Caller should check that body is valid JSON and should enforce a maximum body size (usually done in ExtractJsonBody).
func DecodeJsonBody[T any](body []byte) (dest T, err error) {

	if len(body) == 0 {
		return dest, lyserr.User{Message: "body is missing"}
	}

	err = json.Unmarshal(body, &dest, json.RejectUnknownMembers(true))
	if err != nil {

		var syntaxErr *jsontext.SyntacticError
		var semanticErr *json.SemanticError
		var timeParseErr *time.ParseError

		// return a useful user error where possible
		switch {
		case errors.Is(err, io.ErrUnexpectedEOF):
			return dest, lyserr.User{Message: "body contains badly-formed json"}

		case errors.As(err, &syntaxErr):
			line := findLineinJson(body, int(syntaxErr.ByteOffset))
			return dest, lyserr.User{Message: fmt.Sprintf("json syntax error on line %d", line)}

		// general semantic error
		case errors.As(err, &semanticErr):
			line := findLineinJson(body, int(semanticErr.ByteOffset))

			// try to narrow it down
			switch {

			// unknown field
			case semanticErr.Err == json.ErrUnknownName:
				return dest, lyserr.User{Message: fmt.Sprintf("unknown field '%s' on line %d", semanticErr.JSONPointer.LastToken(), line)}

			// date/time parse error
			case errors.As(err, &timeParseErr):
				return dest, lyserr.User{Message: fmt.Sprintf("failed to parse date or time '%s' on line %d", timeParseErr.Value, line)}

			// IP address parse error
			case semanticErr.GoType == reflect.TypeFor[netip.Addr]() && semanticErr.JSONKind == jsontext.KindString:
				addr, _ := parseWrappedValue(`ParseAddr("`, `"): `, semanticErr.Err.Error())
				return dest, lyserr.User{Message: fmt.Sprintf("failed to parse IP address '%s' on line %d", addr, line)}

			default: // unknown semantic error: assume type error
				return dest, lyserr.User{Message: fmt.Sprintf("json type error on line %d", line)}
			}

		default: // unknown unmarshal error
			return dest, fmt.Errorf("json.Unmarshal failed: %w", err)
		}
	}

	return dest, nil
}

// ExtractJsonBody reads and validates the body of the supplied request.
func ExtractJsonBody(r *http.Request, maxBodySize int64) (body []byte, err error) {

	// check param
	if maxBodySize <= 0 {
		return nil, fmt.Errorf("maxBodySize must be greater than 0")
	}

	// make sure Content-Type header is json
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return nil, ErrInvalidContentType
	}

	defer r.Body.Close()

	// read req body
	body, err = io.ReadAll(io.LimitReader(r.Body, maxBodySize+1))
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll failed: %w", err)
	}

	// ensure body does not exceed maximum allowed size
	if int64(len(body)) > maxBodySize {
		return nil, fmt.Errorf("request body exceeds maxBodySize")
	}

	// ensure there's a body
	if len(body) == 0 {
		return nil, ErrBodyMissing
	}

	// ensure body is valid JSON
	if !jsontext.Value(body).IsValid() {
		return nil, ErrInvalidJson
	}

	return body, nil
}

// findLineinJson returns the line number in a json body corresponding to a given offset. This is used to provide more helpful error messages when json decoding fails.
func findLineinJson(body []byte, offset int) (line int) {
	return bytes.Count(body[:offset], []byte("\n")) + 1
}

// parseWrappedValue is a helper for extracting a value from an error message that is wrapped in a prefix and suffix.
func parseWrappedValue(prefix, suffix, msg string) (string, bool) {
	_, after, found := strings.Cut(msg, prefix)
	if !found {
		return "", false
	}
	rest := after
	val, _, ok := strings.Cut(rest, suffix)
	return val, ok
}
