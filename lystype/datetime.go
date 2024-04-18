package lystype

import (
	"fmt"
	"strings"
	"time"
)

const (
	// DatetimeFormat is the format into which Datetime values are marshalled
	DatetimeFormat string = "2006-01-02 15:04:05-07"
)

// Datetime is an implementation of time.Time which represents datetimes exchanged via json
type Datetime time.Time

// UnmarshalJSON converts the supplied json to a Datetime and writes the result to the receiver
func (t *Datetime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		*t = Datetime(time.Time{})
		return nil
	}
	ti, err := time.Parse(DatetimeFormat, s)
	if err != nil {
		return fmt.Errorf("time.Parse failed: %w", err)
	}
	*t = Datetime(ti)
	return nil
}

// Scan implements the database/sql Scanner interface
// lib/pq doesn't need this, but pgx does
func (t *Datetime) Scan(src any) error {

	if src == nil {
		*t = Datetime(time.Time{})
		return nil
	}

	switch src := src.(type) {
	case time.Time:
		*t = Datetime(src)
		return nil
	default:
		*t = Datetime(time.Time{})
		return nil
	}
}

// MarshalJSON converts the receiver to json
func (t Datetime) MarshalJSON() ([]byte, error) {

	// marshal using db format
	stamp := fmt.Sprintf("\"%s\"", t.Format(DatetimeFormat))
	return []byte(stamp), nil
}

// Format is a wrapper for the same function on the underlying time.Time variable
func (t Datetime) Format(layout string) string {
	return time.Time(t).Format(layout)
}

// IsZero is a wrapper for the same function on the underlying time.Time variable
func (t Datetime) IsZero() bool {
	return time.Time(t).IsZero()
}
