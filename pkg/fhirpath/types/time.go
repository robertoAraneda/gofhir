package types

import (
	"fmt"
	"regexp"
	"strconv"
	gotime "time"
)

// Time represents a FHIRPath time value.
type Time struct {
	hour      int
	minute    int
	second    int
	millis    int
	precision TimePrecision
}

// TimePrecision indicates the precision of a time.
type TimePrecision int

const (
	HourPrecision TimePrecision = iota
	MinutePrecision
	SecondPrecision
	MillisPrecision
)

// Time regex pattern
var timePattern = regexp.MustCompile(
	`^T?(\d{2})(?::(\d{2})(?::(\d{2})(?:\.(\d+))?)?)?$`,
)

// NewTime creates a Time from a string.
func NewTime(s string) (Time, error) {
	matches := timePattern.FindStringSubmatch(s)
	if matches == nil {
		return Time{}, fmt.Errorf("invalid time format: %s", s)
	}

	t := Time{}
	precision := HourPrecision

	// Hour (required)
	hour, err := strconv.Atoi(matches[1])
	if err != nil {
		return Time{}, fmt.Errorf("invalid hour in time: %s", s)
	}
	t.hour = hour

	// Minute
	if matches[2] != "" {
		minute, err := strconv.Atoi(matches[2])
		if err != nil {
			return Time{}, fmt.Errorf("invalid minute in time: %s", s)
		}
		t.minute = minute
		precision = MinutePrecision
	}

	// Second
	if matches[3] != "" {
		second, err := strconv.Atoi(matches[3])
		if err != nil {
			return Time{}, fmt.Errorf("invalid second in time: %s", s)
		}
		t.second = second
		precision = SecondPrecision
	}

	// Milliseconds
	if matches[4] != "" {
		ms := matches[4]
		for len(ms) < 3 {
			ms += "0"
		}
		if len(ms) > 3 {
			ms = ms[:3]
		}
		millis, err := strconv.Atoi(ms)
		if err != nil {
			return Time{}, fmt.Errorf("invalid milliseconds in time: %s", s)
		}
		t.millis = millis
		precision = MillisPrecision
	}

	t.precision = precision
	return t, nil
}

// NewTimeFromGoTime creates a Time from time.Time.
func NewTimeFromGoTime(t gotime.Time) Time {
	return Time{
		hour:      t.Hour(),
		minute:    t.Minute(),
		second:    t.Second(),
		millis:    t.Nanosecond() / 1000000,
		precision: MillisPrecision,
	}
}

// Type returns the type name.
func (t Time) Type() string {
	return "Time"
}

// Equal checks equality with another value.
func (t Time) Equal(other Value) bool {
	if o, ok := other.(Time); ok {
		if t.precision != o.precision {
			return false
		}
		if t.hour != o.hour {
			return false
		}
		if t.precision >= MinutePrecision && t.minute != o.minute {
			return false
		}
		if t.precision >= SecondPrecision && t.second != o.second {
			return false
		}
		if t.precision >= MillisPrecision && t.millis != o.millis {
			return false
		}
		return true
	}
	return false
}

// Equivalent checks equivalence with another value.
func (t Time) Equivalent(other Value) bool {
	return t.Equal(other)
}

// String returns the string representation.
func (t Time) String() string {
	result := fmt.Sprintf("%02d", t.hour)

	if t.precision >= MinutePrecision {
		result += fmt.Sprintf(":%02d", t.minute)
	}
	if t.precision >= SecondPrecision {
		result += fmt.Sprintf(":%02d", t.second)
	}
	if t.precision >= MillisPrecision {
		result += fmt.Sprintf(".%03d", t.millis)
	}

	return result
}

// IsEmpty returns false for Time.
func (t Time) IsEmpty() bool {
	return false
}

// Accessors
func (t Time) Hour() int        { return t.hour }
func (t Time) Minute() int      { return t.minute }
func (t Time) Second() int      { return t.second }
func (t Time) Millisecond() int { return t.millis }

// Compare compares two times. Returns -1, 0, or 1.
// Implements the Comparable interface.
// Returns error if precisions differ and comparison is ambiguous.
func (t Time) Compare(other Value) (int, error) {
	otherTime, ok := other.(Time)
	if !ok {
		return 0, fmt.Errorf("cannot compare Time with %s", other.Type())
	}

	// Check for ambiguous comparison due to different precisions
	if t.precision != otherTime.precision {
		// Compare at the lowest common precision
		minPrecision := t.precision
		if otherTime.precision < minPrecision {
			minPrecision = otherTime.precision
		}

		// Compare hour
		if t.hour != otherTime.hour {
			if t.hour < otherTime.hour {
				return -1, nil
			}
			return 1, nil
		}

		// Compare minute if both have at least minute precision
		if minPrecision >= MinutePrecision {
			if t.minute != otherTime.minute {
				if t.minute < otherTime.minute {
					return -1, nil
				}
				return 1, nil
			}
		} else {
			return 0, fmt.Errorf("ambiguous comparison between times with different precisions")
		}

		// Compare second if both have at least second precision
		if minPrecision >= SecondPrecision {
			if t.second != otherTime.second {
				if t.second < otherTime.second {
					return -1, nil
				}
				return 1, nil
			}
		} else {
			return 0, fmt.Errorf("ambiguous comparison between times with different precisions")
		}

		// If we get here, comparison is ambiguous at milliseconds level
		return 0, fmt.Errorf("ambiguous comparison between times with different precisions")
	}

	// Same precision - direct comparison
	if t.hour < otherTime.hour {
		return -1, nil
	}
	if t.hour > otherTime.hour {
		return 1, nil
	}

	if t.precision >= MinutePrecision {
		if t.minute < otherTime.minute {
			return -1, nil
		}
		if t.minute > otherTime.minute {
			return 1, nil
		}
	}

	if t.precision >= SecondPrecision {
		if t.second < otherTime.second {
			return -1, nil
		}
		if t.second > otherTime.second {
			return 1, nil
		}
	}

	if t.precision >= MillisPrecision {
		if t.millis < otherTime.millis {
			return -1, nil
		}
		if t.millis > otherTime.millis {
			return 1, nil
		}
	}

	return 0, nil
}
