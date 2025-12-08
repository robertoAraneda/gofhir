package funcs

import (
	"time"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func init() {
	// Register temporal component functions
	Register(FuncDef{
		Name:    "year",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnYear,
	})

	Register(FuncDef{
		Name:    "month",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnMonth,
	})

	Register(FuncDef{
		Name:    "day",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnDay,
	})

	Register(FuncDef{
		Name:    "hour",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnHour,
	})

	Register(FuncDef{
		Name:    "minute",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnMinute,
	})

	Register(FuncDef{
		Name:    "second",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnSecond,
	})

	Register(FuncDef{
		Name:    "millisecond",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnMillisecond,
	})

	// Override the placeholder functions with real implementations
	Register(FuncDef{
		Name:    "now",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnNowReal,
	})

	Register(FuncDef{
		Name:    "today",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnTodayReal,
	})

	Register(FuncDef{
		Name:    "timeOfDay",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnTimeOfDayReal,
	})
}

// fnYear returns the year component.
func fnYear(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	switch v := input[0].(type) {
	case types.Date:
		return types.Collection{types.NewInteger(int64(v.Year()))}, nil
	case types.DateTime:
		return types.Collection{types.NewInteger(int64(v.Year()))}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnMonth returns the month component.
func fnMonth(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	switch v := input[0].(type) {
	case types.Date:
		if v.Month() == 0 {
			return types.Collection{}, nil
		}
		return types.Collection{types.NewInteger(int64(v.Month()))}, nil
	case types.DateTime:
		if v.Month() == 0 {
			return types.Collection{}, nil
		}
		return types.Collection{types.NewInteger(int64(v.Month()))}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnDay returns the day component.
func fnDay(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	switch v := input[0].(type) {
	case types.Date:
		if v.Day() == 0 {
			return types.Collection{}, nil
		}
		return types.Collection{types.NewInteger(int64(v.Day()))}, nil
	case types.DateTime:
		if v.Day() == 0 {
			return types.Collection{}, nil
		}
		return types.Collection{types.NewInteger(int64(v.Day()))}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnHour returns the hour component.
func fnHour(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	switch v := input[0].(type) {
	case types.DateTime:
		return types.Collection{types.NewInteger(int64(v.Hour()))}, nil
	case types.Time:
		return types.Collection{types.NewInteger(int64(v.Hour()))}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnMinute returns the minute component.
func fnMinute(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	switch v := input[0].(type) {
	case types.DateTime:
		return types.Collection{types.NewInteger(int64(v.Minute()))}, nil
	case types.Time:
		return types.Collection{types.NewInteger(int64(v.Minute()))}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnSecond returns the second component.
func fnSecond(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	switch v := input[0].(type) {
	case types.DateTime:
		return types.Collection{types.NewInteger(int64(v.Second()))}, nil
	case types.Time:
		return types.Collection{types.NewInteger(int64(v.Second()))}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnMillisecond returns the millisecond component.
func fnMillisecond(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	switch v := input[0].(type) {
	case types.DateTime:
		return types.Collection{types.NewInteger(int64(v.Millisecond()))}, nil
	case types.Time:
		return types.Collection{types.NewInteger(int64(v.Millisecond()))}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnNowReal returns the current datetime.
func fnNowReal(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	return types.Collection{types.NewDateTimeFromTime(time.Now())}, nil
}

// fnTodayReal returns the current date.
func fnTodayReal(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	return types.Collection{types.NewDateFromTime(time.Now())}, nil
}

// fnTimeOfDayReal returns the current time.
func fnTimeOfDayReal(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	return types.Collection{types.NewTimeFromGoTime(time.Now())}, nil
}
