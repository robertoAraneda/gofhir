package funcs

import (
	"testing"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func TestTemporalFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("now", func(t *testing.T) {
		fn, _ := Get("now")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].Type() != "DateTime" {
			t.Errorf("expected DateTime, got %s", result[0].Type())
		}
	})

	t.Run("today", func(t *testing.T) {
		fn, _ := Get("today")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].Type() != "Date" {
			t.Errorf("expected Date, got %s", result[0].Type())
		}
	})

	t.Run("timeOfDay", func(t *testing.T) {
		fn, _ := Get("timeOfDay")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].Type() != "Time" {
			t.Errorf("expected Time, got %s", result[0].Type())
		}
	})

	t.Run("year", func(t *testing.T) {
		fn, _ := Get("year")

		date, _ := types.NewDate("2023-12-25")
		result, err := fn.Fn(ctx, types.Collection{date}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 2023 {
			t.Errorf("expected 2023, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("month", func(t *testing.T) {
		fn, _ := Get("month")

		date, _ := types.NewDate("2023-12-25")
		result, err := fn.Fn(ctx, types.Collection{date}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 12 {
			t.Errorf("expected 12, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("day", func(t *testing.T) {
		fn, _ := Get("day")

		date, _ := types.NewDate("2023-12-25")
		result, err := fn.Fn(ctx, types.Collection{date}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 25 {
			t.Errorf("expected 25, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("hour", func(t *testing.T) {
		fn, _ := Get("hour")

		dt, _ := types.NewDateTime("2023-12-25T10:30:45")
		result, err := fn.Fn(ctx, types.Collection{dt}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 10 {
			t.Errorf("expected 10, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("minute", func(t *testing.T) {
		fn, _ := Get("minute")

		dt, _ := types.NewDateTime("2023-12-25T10:30:45")
		result, err := fn.Fn(ctx, types.Collection{dt}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 30 {
			t.Errorf("expected 30, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("second", func(t *testing.T) {
		fn, _ := Get("second")

		dt, _ := types.NewDateTime("2023-12-25T10:30:45")
		result, err := fn.Fn(ctx, types.Collection{dt}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 45 {
			t.Errorf("expected 45, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("millisecond", func(t *testing.T) {
		fn, _ := Get("millisecond")

		dt, _ := types.NewDateTime("2023-12-25T10:30:45.123")
		result, err := fn.Fn(ctx, types.Collection{dt}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 123 {
			t.Errorf("expected 123, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("time components", func(t *testing.T) {
		time, _ := types.NewTime("10:30:45")

		// hour
		fn, _ := Get("hour")
		result, err := fn.Fn(ctx, types.Collection{time}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 10 {
			t.Errorf("hour: expected 10, got %d", result[0].(types.Integer).Value())
		}

		// minute
		fn, _ = Get("minute")
		result, err = fn.Fn(ctx, types.Collection{time}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 30 {
			t.Errorf("minute: expected 30, got %d", result[0].(types.Integer).Value())
		}

		// second
		fn, _ = Get("second")
		result, err = fn.Fn(ctx, types.Collection{time}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 45 {
			t.Errorf("second: expected 45, got %d", result[0].(types.Integer).Value())
		}
	})
}
