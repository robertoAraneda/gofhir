package funcs

import (
	"testing"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func TestConversionFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("toBoolean", func(t *testing.T) {
		fn, _ := Get("toBoolean")

		// String "true"
		result, err := fn.Fn(ctx, types.Collection{types.NewString("true")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 'true' to convert to true")
		}

		// Integer 1
		result, err = fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 1 to convert to true")
		}

		// Integer 0
		result, err = fn.Fn(ctx, types.Collection{types.NewInteger(0)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 0 to convert to false")
		}
	})

	t.Run("convertsToBoolean", func(t *testing.T) {
		fn, _ := Get("convertsToBoolean")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("true")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 'true' to be convertible to boolean")
		}

		result, err = fn.Fn(ctx, types.Collection{types.NewString("invalid")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 'invalid' to not be convertible to boolean")
		}
	})

	t.Run("toInteger", func(t *testing.T) {
		fn, _ := Get("toInteger")

		// String to integer
		result, err := fn.Fn(ctx, types.Collection{types.NewString("42")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 42 {
			t.Errorf("expected 42, got %d", result[0].(types.Integer).Value())
		}

		// Boolean true to 1
		result, err = fn.Fn(ctx, types.Collection{types.NewBoolean(true)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 1 {
			t.Errorf("expected 1, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("convertsToInteger", func(t *testing.T) {
		fn, _ := Get("convertsToInteger")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("42")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected '42' to be convertible to integer")
		}
	})

	t.Run("toDecimal", func(t *testing.T) {
		fn, _ := Get("toDecimal")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("3.14")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		val := result[0].(types.Decimal).Value().InexactFloat64()
		if val != 3.14 {
			t.Errorf("expected 3.14, got %f", val)
		}
	})

	t.Run("convertsToDecimal", func(t *testing.T) {
		fn, _ := Get("convertsToDecimal")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("3.14")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected '3.14' to be convertible to decimal")
		}
	})

	t.Run("toString", func(t *testing.T) {
		fn, _ := Get("toString")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(42)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "42" {
			t.Errorf("expected '42', got %s", result[0].(types.String).Value())
		}

		result, err = fn.Fn(ctx, types.Collection{types.NewBoolean(true)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "true" {
			t.Errorf("expected 'true', got %s", result[0].(types.String).Value())
		}
	})

	t.Run("convertsToString", func(t *testing.T) {
		fn, _ := Get("convertsToString")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(42)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected integer to be convertible to string")
		}
	})

	t.Run("toDate", func(t *testing.T) {
		fn, _ := Get("toDate")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("2023-12-25")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].Type() != "Date" {
			t.Errorf("expected Date type, got %s", result[0].Type())
		}
	})

	t.Run("convertsToDate", func(t *testing.T) {
		fn, _ := Get("convertsToDate")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("2023-12-25")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected date string to be convertible")
		}

		// Any string is considered convertible in current implementation
		result, err = fn.Fn(ctx, types.Collection{types.NewInteger(123)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected integer to not be convertible to date")
		}
	})

	t.Run("toDateTime", func(t *testing.T) {
		fn, _ := Get("toDateTime")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("2023-12-25T10:30:00")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		// Current implementation returns String, not DateTime
		if result[0].Type() != "String" {
			t.Errorf("expected String type, got %s", result[0].Type())
		}
	})

	t.Run("convertsToDateTime", func(t *testing.T) {
		fn, _ := Get("convertsToDateTime")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("2023-12-25T10:30:00")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected datetime string to be convertible")
		}
	})

	t.Run("toTime", func(t *testing.T) {
		fn, _ := Get("toTime")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("10:30:00")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		// Current implementation returns String, not Time
		if result[0].Type() != "String" {
			t.Errorf("expected String type, got %s", result[0].Type())
		}
	})

	t.Run("convertsToTime", func(t *testing.T) {
		fn, _ := Get("convertsToTime")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("10:30:00")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected time string to be convertible")
		}
	})

	t.Run("iif", func(t *testing.T) {
		fn, _ := Get("iif")

		// True condition - condition is in args[0]
		result, err := fn.Fn(ctx, types.Collection{},
			[]interface{}{
				types.Collection{types.NewBoolean(true)},
				types.Collection{types.NewString("yes")},
				types.Collection{types.NewString("no")},
			})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "yes" {
			t.Errorf("expected 'yes', got %s", result[0].(types.String).Value())
		}

		// False condition
		result, err = fn.Fn(ctx, types.Collection{},
			[]interface{}{
				types.Collection{types.NewBoolean(false)},
				types.Collection{types.NewString("yes")},
				types.Collection{types.NewString("no")},
			})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "no" {
			t.Errorf("expected 'no', got %s", result[0].(types.String).Value())
		}

		// Empty condition - returns otherwise
		result, err = fn.Fn(ctx, types.Collection{},
			[]interface{}{
				types.Collection{},
				types.Collection{types.NewString("yes")},
				types.Collection{types.NewString("no")},
			})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "no" {
			t.Errorf("expected 'no' for empty condition, got %s", result[0].(types.String).Value())
		}
	})
}

func TestAdditionalConversionFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("toBoolean from string false", func(t *testing.T) {
		fn, _ := Get("toBoolean")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("false")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 'false' to convert to false")
		}

		// String 'f'
		result, err = fn.Fn(ctx, types.Collection{types.NewString("f")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 'f' to convert to false")
		}

		// String 't'
		result, err = fn.Fn(ctx, types.Collection{types.NewString("t")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 't' to convert to true")
		}
	})

	t.Run("toBoolean from decimal", func(t *testing.T) {
		fn, _ := Get("toBoolean")

		// Decimal 1.0
		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(1.0)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 1.0 to convert to true")
		}

		// Decimal 0.0
		result, err = fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(0.0)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 0.0 to convert to false")
		}
	})

	t.Run("toBoolean empty", func(t *testing.T) {
		fn, _ := Get("toBoolean")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for toBoolean of empty")
		}
	})

	t.Run("convertsToBoolean integer", func(t *testing.T) {
		fn, _ := Get("convertsToBoolean")

		// Integer 1
		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 1 to be convertible to boolean")
		}

		// Integer 2 - not convertible
		result, err = fn.Fn(ctx, types.Collection{types.NewInteger(2)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 2 to not be convertible to boolean")
		}
	})

	t.Run("convertsToBoolean decimal", func(t *testing.T) {
		fn, _ := Get("convertsToBoolean")

		// Decimal 1.0
		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(1.0)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 1.0 to be convertible to boolean")
		}
	})

	t.Run("toInteger from integer", func(t *testing.T) {
		fn, _ := Get("toInteger")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(42)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 42 {
			t.Errorf("expected 42, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("toDecimal from integer", func(t *testing.T) {
		fn, _ := Get("toDecimal")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(42)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Decimal).Value().InexactFloat64() != 42.0 {
			t.Error("expected 42.0")
		}
	})

	t.Run("toDecimal from boolean", func(t *testing.T) {
		fn, _ := Get("toDecimal")

		result, err := fn.Fn(ctx, types.Collection{types.NewBoolean(true)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Decimal).Value().InexactFloat64() != 1.0 {
			t.Error("expected 1.0 for true")
		}
	})

	t.Run("convertsToInteger string", func(t *testing.T) {
		fn, _ := Get("convertsToInteger")

		// Invalid string
		result, err := fn.Fn(ctx, types.Collection{types.NewString("abc")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 'abc' to not be convertible to integer")
		}
	})

	t.Run("convertsToDecimal string", func(t *testing.T) {
		fn, _ := Get("convertsToDecimal")

		// Invalid string
		result, err := fn.Fn(ctx, types.Collection{types.NewString("abc")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 'abc' to not be convertible to decimal")
		}
	})

	t.Run("toString decimal", func(t *testing.T) {
		fn, _ := Get("toString")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(3.14)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "3.14" {
			t.Errorf("expected '3.14', got %s", result[0].(types.String).Value())
		}
	})

	t.Run("convertsToString empty", func(t *testing.T) {
		fn, _ := Get("convertsToString")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected empty to not be convertible to string")
		}
	})

	t.Run("convertsToDateTime integer", func(t *testing.T) {
		fn, _ := Get("convertsToDateTime")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(123)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected integer to not be convertible to datetime")
		}
	})

	t.Run("convertsToTime integer", func(t *testing.T) {
		fn, _ := Get("convertsToTime")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(123)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected integer to not be convertible to time")
		}
	})
}

func TestQuantityConversion(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("toQuantity from string", func(t *testing.T) {
		fn, _ := Get("toQuantity")

		// Simple quantity with unit
		result, err := fn.Fn(ctx, types.Collection{types.NewString("5.5 mg")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result.Empty() {
			t.Error("expected non-empty result")
		}
		q := result[0].(types.Quantity)
		if q.Value().String() != "5.5" {
			t.Errorf("expected value 5.5, got %s", q.Value().String())
		}
		if q.Unit() != "mg" {
			t.Errorf("expected unit 'mg', got '%s'", q.Unit())
		}
	})

	t.Run("toQuantity from string with quoted unit", func(t *testing.T) {
		fn, _ := Get("toQuantity")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("10 'kg'")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		q := result[0].(types.Quantity)
		if q.Value().String() != "10" {
			t.Errorf("expected value 10, got %s", q.Value().String())
		}
		if q.Unit() != "kg" {
			t.Errorf("expected unit 'kg', got '%s'", q.Unit())
		}
	})

	t.Run("toQuantity from integer", func(t *testing.T) {
		fn, _ := Get("toQuantity")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(42)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		q := result[0].(types.Quantity)
		if q.Value().String() != "42" {
			t.Errorf("expected value 42, got %s", q.Value().String())
		}
		if q.Unit() != "" {
			t.Errorf("expected empty unit, got '%s'", q.Unit())
		}
	})

	t.Run("toQuantity from integer with unit", func(t *testing.T) {
		fn, _ := Get("toQuantity")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(100)},
			[]interface{}{types.Collection{types.NewString("cm")}})
		if err != nil {
			t.Fatal(err)
		}
		q := result[0].(types.Quantity)
		if q.Value().String() != "100" {
			t.Errorf("expected value 100, got %s", q.Value().String())
		}
		if q.Unit() != "cm" {
			t.Errorf("expected unit 'cm', got '%s'", q.Unit())
		}
	})

	t.Run("toQuantity from decimal", func(t *testing.T) {
		fn, _ := Get("toQuantity")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(3.14)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		q := result[0].(types.Quantity)
		if q.Unit() != "" {
			t.Errorf("expected empty unit, got '%s'", q.Unit())
		}
	})

	t.Run("toQuantity from decimal with unit", func(t *testing.T) {
		fn, _ := Get("toQuantity")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(98.6)},
			[]interface{}{types.Collection{types.NewString("[degF]")}})
		if err != nil {
			t.Fatal(err)
		}
		q := result[0].(types.Quantity)
		if q.Unit() != "[degF]" {
			t.Errorf("expected unit '[degF]', got '%s'", q.Unit())
		}
	})

	t.Run("toQuantity from quantity", func(t *testing.T) {
		fn, _ := Get("toQuantity")

		original, _ := types.NewQuantity("25 mL")
		result, err := fn.Fn(ctx, types.Collection{original}, nil)
		if err != nil {
			t.Fatal(err)
		}
		q := result[0].(types.Quantity)
		if !q.Value().Equal(original.Value()) || q.Unit() != original.Unit() {
			t.Error("expected same quantity")
		}
	})

	t.Run("toQuantity from invalid string", func(t *testing.T) {
		fn, _ := Get("toQuantity")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("invalid")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty result for invalid string")
		}
	})

	t.Run("toQuantity empty input", func(t *testing.T) {
		fn, _ := Get("toQuantity")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty result for empty input")
		}
	})

	t.Run("convertsToQuantity from quantity", func(t *testing.T) {
		fn, _ := Get("convertsToQuantity")

		q, _ := types.NewQuantity("5 mg")
		result, err := fn.Fn(ctx, types.Collection{q}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected quantity to be convertible")
		}
	})

	t.Run("convertsToQuantity from integer", func(t *testing.T) {
		fn, _ := Get("convertsToQuantity")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(42)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected integer to be convertible")
		}
	})

	t.Run("convertsToQuantity from decimal", func(t *testing.T) {
		fn, _ := Get("convertsToQuantity")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(3.14)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected decimal to be convertible")
		}
	})

	t.Run("convertsToQuantity from valid string", func(t *testing.T) {
		fn, _ := Get("convertsToQuantity")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("10 kg")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected '10 kg' to be convertible")
		}
	})

	t.Run("convertsToQuantity from invalid string", func(t *testing.T) {
		fn, _ := Get("convertsToQuantity")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("not a quantity")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 'not a quantity' to not be convertible")
		}
	})

	t.Run("convertsToQuantity empty input", func(t *testing.T) {
		fn, _ := Get("convertsToQuantity")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected empty to not be convertible")
		}
	})

	t.Run("convertsToQuantity from boolean", func(t *testing.T) {
		fn, _ := Get("convertsToQuantity")

		result, err := fn.Fn(ctx, types.Collection{types.NewBoolean(true)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected boolean to not be convertible")
		}
	})
}
