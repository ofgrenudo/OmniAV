package models

import (
	"fmt"
	"strings"
	"time"
)

type DaysOfWeekT int16

const (
	Monday DaysOfWeekT = iota
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
	Sunday
)

var daysOfWeekNames = [...]string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}

func (d DaysOfWeekT) String() string {
	if d < Monday || d > Sunday {
		return "Unknown"
	}
	return daysOfWeekNames[d]
}

// ParseDaysOfWeek parses a weekday name (e.g. "Monday", case-insensitive) into a DaysOfWeekT.
func ParseDaysOfWeek(s string) (DaysOfWeekT, error) {
	for i, name := range daysOfWeekNames {
		if strings.EqualFold(name, s) {
			return DaysOfWeekT(i), nil
		}
	}
	return 0, fmt.Errorf("invalid day of week %q", s)
}

// FromGoWeekday converts Go's time.Weekday (Sunday=0) to DaysOfWeekT (Monday=0), so a Request's
// DaysOfWeek can be derived from an actual calendar date instead of trusting client-supplied
// input to agree with it.
func FromGoWeekday(w time.Weekday) DaysOfWeekT {
	switch w {
	case time.Monday:
		return Monday
	case time.Tuesday:
		return Tuesday
	case time.Wednesday:
		return Wednesday
	case time.Thursday:
		return Thursday
	case time.Friday:
		return Friday
	case time.Saturday:
		return Saturday
	default:
		return Sunday
	}
}

func (d DaysOfWeekT) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

func (d *DaysOfWeekT) UnmarshalJSON(data []byte) error {
	parsed, err := ParseDaysOfWeek(strings.Trim(string(data), `"`))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}
