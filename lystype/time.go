package lystype

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	// TimeFormat is the format into which Time values are marshalled
	TimeFormat   string = "15:04"    // use when marshalling to json
	TimeFormatDb string = "15:04:05" // use when inserting into db
)

// Time is an implementation of time.Time which represents times exchanged via json
type Time time.Time

// UnmarshalJSON converts the supplied json to a Time and writes the result to the receiver
func (t *Time) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		*t = Time(time.Time{})
		return nil
	}
	ti, err := time.Parse(TimeFormat, s)
	if err != nil {
		return fmt.Errorf("time.Parse failed: %w", err)
	}
	*t = Time(ti)
	return nil
}

// Scan implements the database/sql Scanner interface.
func (t *Time) Scan(src any) error {

	if src == nil {
		return fmt.Errorf("unsupported Scan source type: %T", src)
	}

	switch src := src.(type) {
	case string:
		expectedFormat := TimeFormatDb
		ti, err := time.Parse(expectedFormat, src)
		if err != nil {
			return fmt.Errorf("time.Parse failed: %w", err)
		}
		*t = Time(ti)
		return nil
	default:
		return fmt.Errorf("unsupported Scan source type: %T", src)
	}
}

// MarshalJSON converts the receiver to json
func (t Time) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuote(nil, t.Format(TimeFormat)), nil
}

// Format is a wrapper for the same function on the underlying time.Time variable
func (t Time) Format(layout string) string {
	return time.Time(t).Format(layout)
}

// IsZero is a wrapper for the same function on the underlying time.Time variable
func (t Time) IsZero() bool {
	return time.Time(t).IsZero()
}

// String returns the Time as a string in the TimeFormat layout.
func (t Time) String() string {
	return t.Format(TimeFormat)
}
