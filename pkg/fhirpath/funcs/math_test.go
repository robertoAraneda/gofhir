package funcs

import (
	"math"
	"testing"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func TestMathFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("abs", func(t *testing.T) {
		fn, _ := Get("abs")

		// Negative integer
		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(-5)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 5 {
			t.Errorf("expected 5, got %d", result[0].(types.Integer).Value())
		}

		// Positive integer
		result, err = fn.Fn(ctx, types.Collection{types.NewInteger(5)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 5 {
			t.Errorf("expected 5, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("ceiling", func(t *testing.T) {
		fn, _ := Get("ceiling")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(1.5)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 2 {
			t.Errorf("expected 2, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("floor", func(t *testing.T) {
		fn, _ := Get("floor")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(1.8)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 1 {
			t.Errorf("expected 1, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("sqrt", func(t *testing.T) {
		fn, _ := Get("sqrt")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(16)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		val := result[0].(types.Decimal).Value().InexactFloat64()
		if val != 4.0 {
			t.Errorf("expected 4, got %f", val)
		}

		// Negative number returns empty
		result, err = fn.Fn(ctx, types.Collection{types.NewInteger(-1)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for sqrt of negative")
		}
	})

	t.Run("power", func(t *testing.T) {
		fn, _ := Get("power")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(2)},
			[]interface{}{types.Collection{types.NewInteger(8)}})
		if err != nil {
			t.Fatal(err)
		}
		val := result[0].(types.Decimal).Value().InexactFloat64()
		if val != 256 {
			t.Errorf("expected 256, got %f", val)
		}
	})

	t.Run("ln", func(t *testing.T) {
		fn, _ := Get("ln")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		val := result[0].(types.Decimal).Value().InexactFloat64()
		if val != 0 {
			t.Errorf("expected 0, got %f", val)
		}
	})

	t.Run("exp", func(t *testing.T) {
		fn, _ := Get("exp")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(0)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		val := result[0].(types.Decimal).Value().InexactFloat64()
		if val != 1.0 {
			t.Errorf("expected 1, got %f", val)
		}
	})

	t.Run("log", func(t *testing.T) {
		fn, _ := Get("log")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(100)},
			[]interface{}{types.Collection{types.NewInteger(10)}})
		if err != nil {
			t.Fatal(err)
		}
		val := result[0].(types.Decimal).Value().InexactFloat64()
		if math.Abs(val-2.0) > 0.0001 {
			t.Errorf("expected 2, got %f", val)
		}
	})

	t.Run("round", func(t *testing.T) {
		fn, _ := Get("round")

		// Round to 2 decimal places
		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(3.14159)},
			[]interface{}{types.Collection{types.NewInteger(2)}})
		if err != nil {
			t.Fatal(err)
		}
		val := result[0].(types.Decimal).Value().InexactFloat64()
		if math.Abs(val-3.14) > 0.001 {
			t.Errorf("expected 3.14, got %f", val)
		}
	})

	t.Run("truncate", func(t *testing.T) {
		fn, _ := Get("truncate")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(3.9)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 3 {
			t.Errorf("expected 3, got %d", result[0].(types.Integer).Value())
		}
	})
}

func TestAdditionalMathFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("abs empty", func(t *testing.T) {
		fn, _ := Get("abs")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for abs of empty")
		}
	})

	t.Run("abs decimal", func(t *testing.T) {
		fn, _ := Get("abs")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(-3.14)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Decimal).Value().InexactFloat64() != 3.14 {
			t.Errorf("expected 3.14, got %v", result[0])
		}
	})

	t.Run("ceiling empty", func(t *testing.T) {
		fn, _ := Get("ceiling")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for ceiling of empty")
		}
	})

	t.Run("ceiling integer", func(t *testing.T) {
		fn, _ := Get("ceiling")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(5)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 5 {
			t.Errorf("expected 5, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("floor empty", func(t *testing.T) {
		fn, _ := Get("floor")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for floor of empty")
		}
	})

	t.Run("floor integer", func(t *testing.T) {
		fn, _ := Get("floor")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(5)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 5 {
			t.Errorf("expected 5, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("sqrt empty", func(t *testing.T) {
		fn, _ := Get("sqrt")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for sqrt of empty")
		}
	})

	t.Run("sqrt decimal", func(t *testing.T) {
		fn, _ := Get("sqrt")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(4.0)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Decimal).Value().InexactFloat64() != 2.0 {
			t.Errorf("expected 2.0, got %v", result[0])
		}
	})

	t.Run("power empty", func(t *testing.T) {
		fn, _ := Get("power")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{types.NewInteger(2)}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for power of empty")
		}
	})

	t.Run("power decimal", func(t *testing.T) {
		fn, _ := Get("power")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(2.0)},
			[]interface{}{types.Collection{types.NewDecimalFromFloat(3.0)}})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Decimal).Value().InexactFloat64() != 8.0 {
			t.Errorf("expected 8.0, got %v", result[0])
		}
	})

	t.Run("ln empty", func(t *testing.T) {
		fn, _ := Get("ln")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for ln of empty")
		}
	})

	t.Run("ln of e", func(t *testing.T) {
		fn, _ := Get("ln")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(math.E)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(result[0].(types.Decimal).Value().InexactFloat64()-1.0) > 0.0001 {
			t.Errorf("expected 1.0, got %v", result[0])
		}
	})

	t.Run("exp empty", func(t *testing.T) {
		fn, _ := Get("exp")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for exp of empty")
		}
	})

	t.Run("exp decimal", func(t *testing.T) {
		fn, _ := Get("exp")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(1.0)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(result[0].(types.Decimal).Value().InexactFloat64()-math.E) > 0.0001 {
			t.Errorf("expected e, got %v", result[0])
		}
	})

	t.Run("log empty", func(t *testing.T) {
		fn, _ := Get("log")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{types.NewInteger(10)}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for log of empty")
		}
	})

	t.Run("log decimal", func(t *testing.T) {
		fn, _ := Get("log")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(1000.0)},
			[]interface{}{types.Collection{types.NewDecimalFromFloat(10.0)}})
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(result[0].(types.Decimal).Value().InexactFloat64()-3.0) > 0.0001 {
			t.Errorf("expected 3.0, got %v", result[0])
		}
	})

	t.Run("round empty", func(t *testing.T) {
		fn, _ := Get("round")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for round of empty")
		}
	})

	t.Run("round without precision", func(t *testing.T) {
		fn, _ := Get("round")

		result, err := fn.Fn(ctx, types.Collection{types.NewDecimalFromFloat(3.7)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Decimal).Value().InexactFloat64() != 4.0 {
			t.Errorf("expected 4.0, got %v", result[0])
		}
	})

	t.Run("truncate empty", func(t *testing.T) {
		fn, _ := Get("truncate")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for truncate of empty")
		}
	})

	t.Run("truncate integer", func(t *testing.T) {
		fn, _ := Get("truncate")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(5)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 5 {
			t.Errorf("expected 5, got %d", result[0].(types.Integer).Value())
		}
	})
}
