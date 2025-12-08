package funcs

import (
	"testing"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func TestAggregateFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{"name": "test", "child": {"value": 1}}`))

	t.Run("combine", func(t *testing.T) {
		fn, _ := Get("combine")

		c1 := types.Collection{types.NewInteger(1), types.NewInteger(2)}
		c2 := types.Collection{types.NewInteger(2), types.NewInteger(3)}

		result, err := fn.Fn(ctx, c1, []interface{}{c2})
		if err != nil {
			t.Fatal(err)
		}
		// Combine keeps duplicates
		if result.Count() != 4 {
			t.Errorf("expected 4 elements, got %d", result.Count())
		}
	})

	t.Run("children", func(t *testing.T) {
		fn, _ := Get("children")

		result, err := fn.Fn(ctx, ctx.Root(), nil)
		if err != nil {
			t.Fatal(err)
		}
		// Should return all direct children
		if result.Empty() {
			t.Error("expected non-empty children")
		}
	})

	t.Run("descendants", func(t *testing.T) {
		fn, _ := Get("descendants")

		result, err := fn.Fn(ctx, ctx.Root(), nil)
		if err != nil {
			t.Fatal(err)
		}
		// Should return all descendants
		if result.Empty() {
			t.Error("expected non-empty descendants")
		}
	})

	t.Run("hasValue", func(t *testing.T) {
		fn, _ := Get("hasValue")

		// Primitive type has value
		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected integer to have value")
		}

		// Empty collection
		result, err = fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected empty collection to not have value")
		}
	})

	t.Run("getValue", func(t *testing.T) {
		fn, _ := Get("getValue")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(42)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 42 {
			t.Errorf("expected 42, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("trace", func(t *testing.T) {
		fn, _ := Get("trace")

		input := types.Collection{types.NewString("test")}
		result, err := fn.Fn(ctx, input, []interface{}{types.Collection{types.NewString("label")}})
		if err != nil {
			t.Fatal(err)
		}
		// Trace returns the input unchanged
		if result[0].(types.String).Value() != "test" {
			t.Errorf("expected 'test', got %s", result[0].(types.String).Value())
		}
	})
}

func TestTypeFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("ofType", func(t *testing.T) {
		fn, _ := Get("ofType")

		input := types.Collection{
			types.NewInteger(1),
			types.NewString("test"),
			types.NewInteger(2),
		}

		result, err := fn.Fn(ctx, input, []interface{}{types.Collection{types.NewString("Integer")}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2 integers, got %d", result.Count())
		}
	})

	t.Run("as", func(t *testing.T) {
		fn, _ := Get("as")

		// Integer as Integer
		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)},
			[]interface{}{types.Collection{types.NewString("Integer")}})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 1 {
			t.Errorf("expected 1, got %d", result[0].(types.Integer).Value())
		}

		// String as Integer returns empty
		result, err = fn.Fn(ctx, types.Collection{types.NewString("test")},
			[]interface{}{types.Collection{types.NewString("Integer")}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for string as Integer")
		}
	})

	t.Run("not", func(t *testing.T) {
		fn, _ := Get("not")

		// not true = false
		result, err := fn.Fn(ctx, types.Collection{types.NewBoolean(true)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected not true = false")
		}

		// not false = true
		result, err = fn.Fn(ctx, types.Collection{types.NewBoolean(false)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected not false = true")
		}

		// not empty = empty
		result, err = fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected not empty = empty")
		}
	})

	t.Run("aggregate", func(t *testing.T) {
		// Note: aggregate needs special handling with expression evaluation
		// This test just exercises the basic function registration
		_, ok := Get("aggregate")
		if !ok {
			t.Error("expected aggregate function to be registered")
		}
	})
}

func TestUnionFunction(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("union removes duplicates", func(t *testing.T) {
		fn, _ := Get("union")

		c1 := types.Collection{types.NewInteger(1), types.NewInteger(2)}
		c2 := types.Collection{types.NewInteger(2), types.NewInteger(3)}

		result, err := fn.Fn(ctx, c1, []interface{}{c2})
		if err != nil {
			t.Fatal(err)
		}
		// Union removes duplicates (2 appears in both)
		if result.Count() != 3 {
			t.Errorf("expected 3 elements, got %d", result.Count())
		}
	})

	t.Run("union with empty collection", func(t *testing.T) {
		fn, _ := Get("union")

		c1 := types.Collection{types.NewInteger(1), types.NewInteger(2)}

		result, err := fn.Fn(ctx, c1, []interface{}{types.Collection{}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2 elements, got %d", result.Count())
		}
	})

	t.Run("union empty with non-empty", func(t *testing.T) {
		fn, _ := Get("union")

		c2 := types.Collection{types.NewInteger(1), types.NewInteger(2)}

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{c2})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2 elements, got %d", result.Count())
		}
	})

	t.Run("union both empty", func(t *testing.T) {
		fn, _ := Get("union")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for union of empty collections")
		}
	})

	t.Run("union with all duplicates", func(t *testing.T) {
		fn, _ := Get("union")

		c1 := types.Collection{types.NewInteger(1), types.NewInteger(2)}
		c2 := types.Collection{types.NewInteger(1), types.NewInteger(2)}

		result, err := fn.Fn(ctx, c1, []interface{}{c2})
		if err != nil {
			t.Fatal(err)
		}
		// All duplicates removed
		if result.Count() != 2 {
			t.Errorf("expected 2 elements, got %d", result.Count())
		}
	})

	t.Run("union with strings", func(t *testing.T) {
		fn, _ := Get("union")

		c1 := types.Collection{types.NewString("a"), types.NewString("b")}
		c2 := types.Collection{types.NewString("b"), types.NewString("c")}

		result, err := fn.Fn(ctx, c1, []interface{}{c2})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 3 {
			t.Errorf("expected 3 elements, got %d", result.Count())
		}
	})

	t.Run("union with mixed types", func(t *testing.T) {
		fn, _ := Get("union")

		c1 := types.Collection{types.NewInteger(1), types.NewString("a")}
		c2 := types.Collection{types.NewInteger(1), types.NewString("b")}

		result, err := fn.Fn(ctx, c1, []interface{}{c2})
		if err != nil {
			t.Fatal(err)
		}
		// Integer 1 appears twice, should be deduplicated
		if result.Count() != 3 {
			t.Errorf("expected 3 elements, got %d", result.Count())
		}
	})

	t.Run("union no arguments error", func(t *testing.T) {
		fn, _ := Get("union")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, []interface{}{})
		if err == nil {
			t.Error("expected error for union without arguments")
		}
	})

	t.Run("union with non-collection argument", func(t *testing.T) {
		fn, _ := Get("union")

		c1 := types.Collection{types.NewInteger(1), types.NewInteger(2)}

		// When argument is not a Collection, returns input unchanged
		result, err := fn.Fn(ctx, c1, []interface{}{"not a collection"})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2 elements, got %d", result.Count())
		}
	})
}

func TestAdditionalAggregateFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("combine empty collections", func(t *testing.T) {
		fn, _ := Get("combine")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for combining empty collections")
		}
	})

	t.Run("hasValue with multiple values", func(t *testing.T) {
		fn, _ := Get("hasValue")

		// Multiple primitive values - should return true (has at least one primitive)
		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for multiple primitive values")
		}
	})

	t.Run("getValue with empty", func(t *testing.T) {
		fn, _ := Get("getValue")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for getValue of empty")
		}
	})

	t.Run("getValue with multiple values", func(t *testing.T) {
		fn, _ := Get("getValue")

		// getValue returns all primitive values
		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2 values, got %d", result.Count())
		}
	})
}
