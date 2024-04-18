package lys

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/loveyourstack/lys/lyserr"
)

// findLine returns the line number in a json body corresponding to a given offset
func findLine(body []byte, offset int) (line int) {

	js := string(body)
	start := strings.LastIndex(js[:offset], "\n") + 1
	return strings.Count(js[:start], "\n")
}

// DecodeJsonBody decodes the supplied json body into dest and checks for a variety of error conditions
// adapted from https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
func DecodeJsonBody[T any](body []byte) (dest T, err error) {

	if body == nil || string(body) == "" {
		return dest, lyserr.User{Message: "body is missing"}
	}

	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()

	if err = dec.Decode(&dest); err != nil {

		var syntaxErr *json.SyntaxError
		var unmarshalTypeErr *json.UnmarshalTypeError
		var timeParseErr *time.ParseError

		//fmt.Println("DecodeJsonBody: " + err.Error())

		switch {
		case errors.Is(err, io.ErrUnexpectedEOF):
			return dest, lyserr.User{Message: "request body contains badly-formed json"}

		case errors.As(err, &syntaxErr):
			line := findLine(body, int(syntaxErr.Offset))
			return dest, lyserr.User{Message: "json syntax error: line: " + strconv.Itoa(line)}

		case errors.As(err, &unmarshalTypeErr):
			line := findLine(body, int(unmarshalTypeErr.Offset))
			return dest, lyserr.User{Message: "json type error: line: " + strconv.Itoa(line)}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return dest, lyserr.User{Message: "unknown field: " + strings.Trim(fieldName, `"`)}

		// TODO - passing a date value such as "2021-06-282" in body causes a slice out of bounds panic, find out how to catch it
		case errors.As(err, &timeParseErr):
			asLoc := strings.Index(timeParseErr.Error(), " as ")
			if asLoc > 13 && len(timeParseErr.Error()) > 14 {
				return dest, lyserr.User{Message: "failed to parse a date or time: " + strings.Trim(timeParseErr.Error()[13:asLoc], "\\\"")}
			}
			return dest, lyserr.User{Message: "failed to parse a date or time: " + timeParseErr.Error()}

		default:
			return dest, fmt.Errorf("dec.Decode failed: %w", err)
		}
	}

	return dest, nil
}
