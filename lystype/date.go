package lystype

import (
	"fmt"
	"strings"
	"time"
)

const (
	// DateFormat is the format into which Date values are marshalled
	DateFormat string = "2006-01-02"
)

// Date is an implementation of time.Time which represents dates exchanged via json
type Date time.Time

// UnmarshalJSON converts the supplied json to a Date and writes the result to the receiver
func (t *Date) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		*t = Date(time.Time{})
		return nil
	}
	ti, err := time.Parse(DateFormat, s)
	if err != nil {
		return fmt.Errorf("time.Parse failed: %w", err)
	}
	*t = Date(ti)
	return nil
}

// Scan implements the database/sql Scanner interface
// lib/pq doesn't need this, but pgx does
func (t *Date) Scan(src any) error {

	if src == nil {
		*t = Date(time.Time{})
		return nil
	}

	switch src := src.(type) {
	case time.Time:
		*t = Date(src)
		return nil
	default:
		*t = Date(time.Time{})
		return nil
	}
}

// MarshalJSON converts the receiver to json
func (t Date) MarshalJSON() ([]byte, error) {

	// marshal using db format
	stamp := fmt.Sprintf("\"%s\"", t.Format(DateFormat))
	return []byte(stamp), nil
}

// Format is a wrapper for the same function on the underlying time.Time variable
func (t Date) Format(layout string) string {
	return time.Time(t).Format(layout)
}

// IsZero is a wrapper for the same function on the underlying time.Time variable
func (t Date) IsZero() bool {
	return time.Time(t).IsZero()
}
