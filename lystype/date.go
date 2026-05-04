package lystype

import (
	"fmt"
	"strconv"
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

// Scan implements the database/sql Scanner interface.
func (t *Date) Scan(src any) error {

	if src == nil {
		return fmt.Errorf("unsupported Scan source type: %T", src)
	}

	switch src := src.(type) {
	case time.Time:
		*t = Date(src)
		return nil
	default:
		return fmt.Errorf("unsupported Scan source type: %T", src)
	}
}

// MarshalJSON converts the receiver to json
func (t Date) MarshalJSON() ([]byte, error) {
	return strconv.AppendQuote(nil, t.Format(DateFormat)), nil
}

// Format is a wrapper for the same function on the underlying time.Time variable
func (t Date) Format(layout string) string {
	return time.Time(t).Format(layout)
}

// IsZero is a wrapper for the same function on the underlying time.Time variable
func (t Date) IsZero() bool {
	return time.Time(t).IsZero()
}

// String returns the Date as a string in the DateFormat layout.
func (t Date) String() string {
	return t.Format(DateFormat)
}

// ToTime converts the Date to a time.Time.
func (t Date) ToTime() time.Time {
	return time.Time(t)
}
