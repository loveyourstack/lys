package lystype

import (
	"fmt"
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

// Scan implements the database/sql Scanner interface
// lib/pq doesn't need this, but pgx does
func (t *Time) Scan(src any) error {

	if src == nil {
		*t = Time(time.Time{})
		return nil
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
		*t = Time(time.Time{})
		return nil
	}
}

// MarshalJSON converts the receiver to json
func (t Time) MarshalJSON() ([]byte, error) {

	// marshal using db format
	stamp := fmt.Sprintf("\"%s\"", t.Format(TimeFormat))
	return []byte(stamp), nil
}

// Format is a wrapper for the same function on the underlying time.Time variable
func (t Time) Format(layout string) string {
	return time.Time(t).Format(layout)
}

// IsZero is a wrapper for the same function on the underlying time.Time variable
func (t Time) IsZero() bool {
	return time.Time(t).IsZero()
}
