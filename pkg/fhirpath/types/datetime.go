package types

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// DateTime represents a FHIRPath datetime value.
type DateTime struct {
	year      int
	month     int
	day       int
	hour      int
	minute    int
	second    int
	millis    int
	tzOffset  int  // timezone offset in minutes
	hasTZ     bool // whether timezone is specified
	precision DateTimePrecision
}

// DateTimePrecision indicates the precision of a datetime.
type DateTimePrecision int

const (
	DTYearPrecision DateTimePrecision = iota
	DTMonthPrecision
	DTDayPrecision
	DTHourPrecision
	DTMinutePrecision
	DTSecondPrecision
	DTMillisPrecision
)

// DateTime regex pattern
var dateTimePattern = regexp.MustCompile(
	`^(\d{4})(?:-(\d{2})(?:-(\d{2})(?:T(\d{2})(?::(\d{2})(?::(\d{2})(?:\.(\d+))?)?)?)?)?)?(Z|[+-]\d{2}:\d{2})?$`,
)

// NewDateTime creates a DateTime from a string.
func NewDateTime(s string) (DateTime, error) {
	matches := dateTimePattern.FindStringSubmatch(s)
	if matches == nil {
		return DateTime{}, fmt.Errorf("invalid datetime format: %s", s)
	}

	dt := DateTime{}
	precision := DTYearPrecision

	// Year (required)
	year, err := strconv.Atoi(matches[1])
	if err != nil {
		return DateTime{}, fmt.Errorf("invalid year in datetime: %s", s)
	}
	dt.year = year

	// Month
	if matches[2] != "" {
		month, err := strconv.Atoi(matches[2])
		if err != nil {
			return DateTime{}, fmt.Errorf("invalid month in datetime: %s", s)
		}
		dt.month = month
		precision = DTMonthPrecision
	}

	// Day
	if matches[3] != "" {
		day, err := strconv.Atoi(matches[3])
		if err != nil {
			return DateTime{}, fmt.Errorf("invalid day in datetime: %s", s)
		}
		dt.day = day
		precision = DTDayPrecision
	}

	// Hour
	if matches[4] != "" {
		hour, err := strconv.Atoi(matches[4])
		if err != nil {
			return DateTime{}, fmt.Errorf("invalid hour in datetime: %s", s)
		}
		dt.hour = hour
		precision = DTHourPrecision
	}

	// Minute
	if matches[5] != "" {
		minute, err := strconv.Atoi(matches[5])
		if err != nil {
			return DateTime{}, fmt.Errorf("invalid minute in datetime: %s", s)
		}
		dt.minute = minute
		precision = DTMinutePrecision
	}

	// Second
	if matches[6] != "" {
		second, err := strconv.Atoi(matches[6])
		if err != nil {
			return DateTime{}, fmt.Errorf("invalid second in datetime: %s", s)
		}
		dt.second = second
		precision = DTSecondPrecision
	}

	// Milliseconds
	if matches[7] != "" {
		// Pad or truncate to 3 digits
		ms := matches[7]
		for len(ms) < 3 {
			ms += "0"
		}
		if len(ms) > 3 {
			ms = ms[:3]
		}
		millis, err := strconv.Atoi(ms)
		if err != nil {
			return DateTime{}, fmt.Errorf("invalid milliseconds in datetime: %s", s)
		}
		dt.millis = millis
		precision = DTMillisPrecision
	}

	// Timezone
	if matches[8] != "" {
		dt.hasTZ = true
		if matches[8] == "Z" {
			dt.tzOffset = 0
		} else {
			// Parse timezone offset
			sign := 1
			if matches[8][0] == '-' {
				sign = -1
			}
			hours, err := strconv.Atoi(matches[8][1:3])
			if err != nil {
				return DateTime{}, fmt.Errorf("invalid timezone hours in datetime: %s", s)
			}
			mins, err := strconv.Atoi(matches[8][4:6])
			if err != nil {
				return DateTime{}, fmt.Errorf("invalid timezone minutes in datetime: %s", s)
			}
			dt.tzOffset = sign * (hours*60 + mins)
		}
	}

	dt.precision = precision
	return dt, nil
}

// NewDateTimeFromTime creates a DateTime from time.Time.
func NewDateTimeFromTime(t time.Time) DateTime {
	_, offset := t.Zone()
	return DateTime{
		year:      t.Year(),
		month:     int(t.Month()),
		day:       t.Day(),
		hour:      t.Hour(),
		minute:    t.Minute(),
		second:    t.Second(),
		millis:    t.Nanosecond() / 1000000,
		tzOffset:  offset / 60,
		hasTZ:     true,
		precision: DTMillisPrecision,
	}
}

// Type returns the type name.
func (dt DateTime) Type() string {
	return "DateTime"
}

// Equal checks equality with another value.
func (dt DateTime) Equal(other Value) bool {
	if o, ok := other.(DateTime); ok {
		return dt.ToTime().Equal(o.ToTime())
	}
	return false
}

// Equivalent checks equivalence with another value.
func (dt DateTime) Equivalent(other Value) bool {
	return dt.Equal(other)
}

// String returns the string representation.
func (dt DateTime) String() string {
	result := fmt.Sprintf("%04d", dt.year)

	if dt.precision >= DTMonthPrecision {
		result += fmt.Sprintf("-%02d", dt.month)
	}
	if dt.precision >= DTDayPrecision {
		result += fmt.Sprintf("-%02d", dt.day)
	}
	if dt.precision >= DTHourPrecision {
		result += fmt.Sprintf("T%02d", dt.hour)
	}
	if dt.precision >= DTMinutePrecision {
		result += fmt.Sprintf(":%02d", dt.minute)
	}
	if dt.precision >= DTSecondPrecision {
		result += fmt.Sprintf(":%02d", dt.second)
	}
	if dt.precision >= DTMillisPrecision {
		result += fmt.Sprintf(".%03d", dt.millis)
	}

	if dt.hasTZ {
		if dt.tzOffset == 0 {
			result += "Z"
		} else {
			sign := "+"
			offset := dt.tzOffset
			if offset < 0 {
				sign = "-"
				offset = -offset
			}
			result += fmt.Sprintf("%s%02d:%02d", sign, offset/60, offset%60)
		}
	}

	return result
}

// IsEmpty returns false for DateTime.
func (dt DateTime) IsEmpty() bool {
	return false
}

// ToTime converts to time.Time.
func (dt DateTime) ToTime() time.Time {
	month := dt.month
	if month == 0 {
		month = 1
	}
	day := dt.day
	if day == 0 {
		day = 1
	}

	var loc *time.Location
	if dt.hasTZ {
		loc = time.FixedZone("", dt.tzOffset*60)
	} else {
		loc = time.UTC
	}

	return time.Date(dt.year, time.Month(month), day, dt.hour, dt.minute, dt.second, dt.millis*1000000, loc)
}

// Accessors
func (dt DateTime) Year() int        { return dt.year }
func (dt DateTime) Month() int       { return dt.month }
func (dt DateTime) Day() int         { return dt.day }
func (dt DateTime) Hour() int        { return dt.hour }
func (dt DateTime) Minute() int      { return dt.minute }
func (dt DateTime) Second() int      { return dt.second }
func (dt DateTime) Millisecond() int { return dt.millis }

// AddDuration adds a duration (as Quantity with temporal unit) to the datetime.
// Supported units: year(s), month(s), week(s), day(s), hour(s), minute(s), second(s), millisecond(s)
func (dt DateTime) AddDuration(value int, unit string) DateTime {
	t := dt.ToTime()

	switch unit {
	case "year", "years", "'year'", "'years'":
		t = t.AddDate(value, 0, 0)
	case "month", "months", "'month'", "'months'":
		t = t.AddDate(0, value, 0)
	case "week", "weeks", "'week'", "'weeks'":
		t = t.AddDate(0, 0, value*7)
	case "day", "days", "'day'", "'days'":
		t = t.AddDate(0, 0, value)
	case "hour", "hours", "'hour'", "'hours'":
		t = t.Add(time.Duration(value) * time.Hour)
	case "minute", "minutes", "'minute'", "'minutes'":
		t = t.Add(time.Duration(value) * time.Minute)
	case "second", "seconds", "'second'", "'seconds'":
		t = t.Add(time.Duration(value) * time.Second)
	case "millisecond", "milliseconds", "'millisecond'", "'milliseconds'", "ms":
		t = t.Add(time.Duration(value) * time.Millisecond)
	default:
		// For unsupported units, return unchanged
		return dt
	}

	result := DateTime{
		year:      t.Year(),
		month:     int(t.Month()),
		day:       t.Day(),
		hour:      t.Hour(),
		minute:    t.Minute(),
		second:    t.Second(),
		millis:    t.Nanosecond() / 1000000,
		tzOffset:  dt.tzOffset,
		hasTZ:     dt.hasTZ,
		precision: dt.precision,
	}

	// Adjust precision - zero out components beyond precision
	if dt.precision < DTMonthPrecision {
		result.month = 0
	}
	if dt.precision < DTDayPrecision {
		result.day = 0
	}
	if dt.precision < DTHourPrecision {
		result.hour = 0
	}
	if dt.precision < DTMinutePrecision {
		result.minute = 0
	}
	if dt.precision < DTSecondPrecision {
		result.second = 0
	}
	if dt.precision < DTMillisPrecision {
		result.millis = 0
	}

	return result
}

// SubtractDuration subtracts a duration from the datetime.
func (dt DateTime) SubtractDuration(value int, unit string) DateTime {
	return dt.AddDuration(-value, unit)
}

// Compare compares two datetimes. Returns -1, 0, or 1.
// Implements the Comparable interface.
// Returns error if precisions differ and comparison is ambiguous.
func (dt DateTime) Compare(other Value) (int, error) {
	otherDT, ok := other.(DateTime)
	if !ok {
		return 0, fmt.Errorf("cannot compare DateTime with %s", other.Type())
	}

	// Check for ambiguous comparison due to different precisions
	if dt.precision != otherDT.precision {
		// Compare at the lowest common precision
		minPrecision := dt.precision
		if otherDT.precision < minPrecision {
			minPrecision = otherDT.precision
		}

		// Compare year
		if dt.year != otherDT.year {
			if dt.year < otherDT.year {
				return -1, nil
			}
			return 1, nil
		}

		// Compare month if both have at least month precision
		if minPrecision >= DTMonthPrecision {
			if dt.month != otherDT.month {
				if dt.month < otherDT.month {
					return -1, nil
				}
				return 1, nil
			}
		} else {
			return 0, fmt.Errorf("ambiguous comparison between datetimes with different precisions")
		}

		// Compare day if both have at least day precision
		if minPrecision >= DTDayPrecision {
			if dt.day != otherDT.day {
				if dt.day < otherDT.day {
					return -1, nil
				}
				return 1, nil
			}
		} else {
			return 0, fmt.Errorf("ambiguous comparison between datetimes with different precisions")
		}

		// Compare hour if both have at least hour precision
		if minPrecision >= DTHourPrecision {
			if dt.hour != otherDT.hour {
				if dt.hour < otherDT.hour {
					return -1, nil
				}
				return 1, nil
			}
		} else {
			return 0, fmt.Errorf("ambiguous comparison between datetimes with different precisions")
		}

		// Compare minute if both have at least minute precision
		if minPrecision >= DTMinutePrecision {
			if dt.minute != otherDT.minute {
				if dt.minute < otherDT.minute {
					return -1, nil
				}
				return 1, nil
			}
		} else {
			return 0, fmt.Errorf("ambiguous comparison between datetimes with different precisions")
		}

		// Compare second if both have at least second precision
		if minPrecision >= DTSecondPrecision {
			if dt.second != otherDT.second {
				if dt.second < otherDT.second {
					return -1, nil
				}
				return 1, nil
			}
		} else {
			return 0, fmt.Errorf("ambiguous comparison between datetimes with different precisions")
		}

		// If we get here, comparison is ambiguous at milliseconds level
		return 0, fmt.Errorf("ambiguous comparison between datetimes with different precisions")
	}

	// Same precision - convert to time.Time and compare
	t1 := dt.ToTime()
	t2 := otherDT.ToTime()

	if t1.Before(t2) {
		return -1, nil
	}
	if t1.After(t2) {
		return 1, nil
	}
	return 0, nil
}
