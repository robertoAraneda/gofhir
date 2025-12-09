package types

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Date represents a FHIRPath date value.
// Supports partial dates: year, year-month, year-month-day.
type Date struct {
	year      int
	month     int // 0 if not specified
	day       int // 0 if not specified
	precision DatePrecision
}

// DatePrecision indicates the precision of a date.
type DatePrecision int

const (
	YearPrecision DatePrecision = iota
	MonthPrecision
	DayPrecision
)

// Date regex patterns
var (
	dateYearPattern  = regexp.MustCompile(`^(\d{4})$`)
	dateMonthPattern = regexp.MustCompile(`^(\d{4})-(\d{2})$`)
	dateDayPattern   = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)
)

// NewDate creates a Date from a string.
func NewDate(s string) (Date, error) {
	// Try full date first
	if matches := dateDayPattern.FindStringSubmatch(s); matches != nil {
		year, err := strconv.Atoi(matches[1])
		if err != nil {
			return Date{}, fmt.Errorf("invalid year in date: %s", s)
		}
		month, err := strconv.Atoi(matches[2])
		if err != nil {
			return Date{}, fmt.Errorf("invalid month in date: %s", s)
		}
		day, err := strconv.Atoi(matches[3])
		if err != nil {
			return Date{}, fmt.Errorf("invalid day in date: %s", s)
		}
		return Date{year: year, month: month, day: day, precision: DayPrecision}, nil
	}

	// Try year-month
	if matches := dateMonthPattern.FindStringSubmatch(s); matches != nil {
		year, err := strconv.Atoi(matches[1])
		if err != nil {
			return Date{}, fmt.Errorf("invalid year in date: %s", s)
		}
		month, err := strconv.Atoi(matches[2])
		if err != nil {
			return Date{}, fmt.Errorf("invalid month in date: %s", s)
		}
		return Date{year: year, month: month, precision: MonthPrecision}, nil
	}

	// Try year only
	if matches := dateYearPattern.FindStringSubmatch(s); matches != nil {
		year, err := strconv.Atoi(matches[1])
		if err != nil {
			return Date{}, fmt.Errorf("invalid year in date: %s", s)
		}
		return Date{year: year, precision: YearPrecision}, nil
	}

	return Date{}, fmt.Errorf("invalid date format: %s", s)
}

// NewDateFromTime creates a Date from a time.Time.
func NewDateFromTime(t time.Time) Date {
	return Date{
		year:      t.Year(),
		month:     int(t.Month()),
		day:       t.Day(),
		precision: DayPrecision,
	}
}

// Type returns the type name.
func (d Date) Type() string {
	return "Date"
}

// Equal checks equality with another value.
func (d Date) Equal(other Value) bool {
	if o, ok := other.(Date); ok {
		if d.precision != o.precision {
			return false
		}
		if d.year != o.year {
			return false
		}
		if d.precision >= MonthPrecision && d.month != o.month {
			return false
		}
		if d.precision >= DayPrecision && d.day != o.day {
			return false
		}
		return true
	}
	return false
}

// Equivalent checks equivalence with another value.
func (d Date) Equivalent(other Value) bool {
	return d.Equal(other)
}

// String returns the string representation.
func (d Date) String() string {
	switch d.precision {
	case YearPrecision:
		return fmt.Sprintf("%04d", d.year)
	case MonthPrecision:
		return fmt.Sprintf("%04d-%02d", d.year, d.month)
	default:
		return fmt.Sprintf("%04d-%02d-%02d", d.year, d.month, d.day)
	}
}

// IsEmpty returns false for Date.
func (d Date) IsEmpty() bool {
	return false
}

// Year returns the year component.
func (d Date) Year() int {
	return d.year
}

// Month returns the month component (0 if not specified).
func (d Date) Month() int {
	return d.month
}

// Day returns the day component (0 if not specified).
func (d Date) Day() int {
	return d.day
}

// Precision returns the date precision.
func (d Date) Precision() DatePrecision {
	return d.precision
}

// ToTime converts to time.Time (uses defaults for missing components).
func (d Date) ToTime() time.Time {
	month := d.month
	if month == 0 {
		month = 1
	}
	day := d.day
	if day == 0 {
		day = 1
	}
	return time.Date(d.year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// Compare compares two dates. Returns -1, 0, or 1.
// Implements the Comparable interface.
// Returns empty (error) if precisions differ and comparison is ambiguous.
func (d Date) Compare(other Value) (int, error) {
	otherDate, ok := other.(Date)
	if !ok {
		return 0, fmt.Errorf("cannot compare Date with %s", other.Type())
	}

	// Check for ambiguous comparison due to different precisions
	// According to FHIRPath spec, comparing dates with different precisions
	// where the more precise date falls within the less precise date's range
	// should return empty (represented as error here)
	if d.precision != otherDate.precision {
		// If years are different, we can still compare
		if d.year != otherDate.year {
			if d.year < otherDate.year {
				return -1, nil
			}
			return 1, nil
		}

		// Years are equal but precisions differ
		minPrecision := d.precision
		if otherDate.precision < minPrecision {
			minPrecision = otherDate.precision
		}

		// If one has only year precision, comparison is ambiguous
		if minPrecision == YearPrecision {
			return 0, fmt.Errorf("ambiguous comparison between dates with different precisions")
		}

		// Check months if both have at least month precision
		if d.precision >= MonthPrecision && otherDate.precision >= MonthPrecision {
			if d.month != otherDate.month {
				if d.month < otherDate.month {
					return -1, nil
				}
				return 1, nil
			}
		}

		// If we get here, comparison is ambiguous
		return 0, fmt.Errorf("ambiguous comparison between dates with different precisions")
	}

	// Same precision - direct comparison
	if d.year < otherDate.year {
		return -1, nil
	}
	if d.year > otherDate.year {
		return 1, nil
	}

	// Compare months if both have at least month precision
	if d.precision >= MonthPrecision {
		if d.month < otherDate.month {
			return -1, nil
		}
		if d.month > otherDate.month {
			return 1, nil
		}
	}

	// Compare days if both have day precision
	if d.precision >= DayPrecision {
		if d.day < otherDate.day {
			return -1, nil
		}
		if d.day > otherDate.day {
			return 1, nil
		}
	}

	return 0, nil
}

// AddDuration adds a duration (as Quantity with temporal unit) to the date.
// Supported units: year(s), month(s), week(s), day(s)
func (d Date) AddDuration(value int, unit string) Date {
	t := d.ToTime()

	switch unit {
	case "year", "years", "'year'", "'years'":
		t = t.AddDate(value, 0, 0)
	case "month", "months", "'month'", "'months'":
		t = t.AddDate(0, value, 0)
	case "week", "weeks", "'week'", "'weeks'":
		t = t.AddDate(0, 0, value*7)
	case "day", "days", "'day'", "'days'":
		t = t.AddDate(0, 0, value)
	default:
		// For unsupported units, return unchanged
		return d
	}

	result := Date{
		year:      t.Year(),
		month:     int(t.Month()),
		day:       t.Day(),
		precision: d.precision,
	}

	// Adjust precision
	if d.precision < MonthPrecision {
		result.month = 0
	}
	if d.precision < DayPrecision {
		result.day = 0
	}

	return result
}

// SubtractDuration subtracts a duration from the date.
func (d Date) SubtractDuration(value int, unit string) Date {
	return d.AddDuration(-value, unit)
}
