package funcs

import (
	"testing"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func TestExistenceFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("empty", func(t *testing.T) {
		fn, _ := Get("empty")

		// Empty collection
		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for empty collection")
		}

		// Non-empty collection
		result, err = fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false for non-empty collection")
		}
	})

	t.Run("exists", func(t *testing.T) {
		fn, _ := Get("exists")

		// Empty collection
		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false for empty collection")
		}

		// Non-empty collection
		result, err = fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for non-empty collection")
		}
	})

	t.Run("count", func(t *testing.T) {
		fn, _ := Get("count")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 3 {
			t.Errorf("expected 3, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("distinct", func(t *testing.T) {
		fn, _ := Get("distinct")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(1),
			types.NewInteger(3),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 3 {
			t.Errorf("expected 3 distinct, got %d", result.Count())
		}
	})

	t.Run("isDistinct", func(t *testing.T) {
		fn, _ := Get("isDistinct")

		// Distinct collection
		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for distinct collection")
		}

		// Non-distinct collection
		result, err = fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(1),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false for non-distinct collection")
		}
	})

	t.Run("allTrue", func(t *testing.T) {
		fn, _ := Get("allTrue")

		// All true
		result, err := fn.Fn(ctx, types.Collection{
			types.NewBoolean(true),
			types.NewBoolean(true),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true")
		}

		// Not all true
		result, err = fn.Fn(ctx, types.Collection{
			types.NewBoolean(true),
			types.NewBoolean(false),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false")
		}
	})

	t.Run("anyTrue", func(t *testing.T) {
		fn, _ := Get("anyTrue")

		// Some true
		result, err := fn.Fn(ctx, types.Collection{
			types.NewBoolean(false),
			types.NewBoolean(true),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true")
		}

		// None true
		result, err = fn.Fn(ctx, types.Collection{
			types.NewBoolean(false),
			types.NewBoolean(false),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false")
		}
	})

	t.Run("allFalse", func(t *testing.T) {
		fn, _ := Get("allFalse")

		// All false
		result, err := fn.Fn(ctx, types.Collection{
			types.NewBoolean(false),
			types.NewBoolean(false),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for all false")
		}

		// Not all false
		result, err = fn.Fn(ctx, types.Collection{
			types.NewBoolean(false),
			types.NewBoolean(true),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false for not all false")
		}

		// Empty collection
		result, err = fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for empty collection")
		}
	})

	t.Run("anyFalse", func(t *testing.T) {
		fn, _ := Get("anyFalse")

		// Some false
		result, err := fn.Fn(ctx, types.Collection{
			types.NewBoolean(true),
			types.NewBoolean(false),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for some false")
		}

		// None false
		result, err = fn.Fn(ctx, types.Collection{
			types.NewBoolean(true),
			types.NewBoolean(true),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false for none false")
		}

		// Empty collection
		result, err = fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false for empty collection")
		}
	})

	t.Run("all", func(t *testing.T) {
		fn, _ := Get("all")

		// Empty collection - vacuous truth
		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for empty collection (vacuous truth)")
		}
	})

	t.Run("subsetOf", func(t *testing.T) {
		fn, _ := Get("subsetOf")

		// Subset
		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
		}, []interface{}{types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
		}})
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for subset")
		}

		// Not subset
		result, err = fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(5),
		}, []interface{}{types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
		}})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false for not subset")
		}

		// Empty collection is subset of anything
		result, err = fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{
			types.NewInteger(1),
		}})
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for empty subset")
		}
	})

	t.Run("supersetOf", func(t *testing.T) {
		fn, _ := Get("supersetOf")

		// Superset
		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
		}, []interface{}{types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
		}})
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for superset")
		}

		// Not superset
		result, err = fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
		}, []interface{}{types.Collection{
			types.NewInteger(1),
			types.NewInteger(5),
		}})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false for not superset")
		}

		// Superset of empty is always true
		result, err = fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
		}, []interface{}{types.Collection{}})
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for superset of empty")
		}
	})
}

func TestSubsettingFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("first", func(t *testing.T) {
		fn, _ := Get("first")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 1 {
			t.Errorf("expected 1, got %d", result[0].(types.Integer).Value())
		}

		// Empty collection
		result, err = fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for empty input")
		}
	})

	t.Run("last", func(t *testing.T) {
		fn, _ := Get("last")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 3 {
			t.Errorf("expected 3, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("tail", func(t *testing.T) {
		fn, _ := Get("tail")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2, got %d", result.Count())
		}
		if result[0].(types.Integer).Value() != 2 {
			t.Errorf("expected first element 2, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("skip", func(t *testing.T) {
		fn, _ := Get("skip")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
			types.NewInteger(4),
			types.NewInteger(5),
		}, []interface{}{types.Collection{types.NewInteger(2)}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 3 {
			t.Errorf("expected 3, got %d", result.Count())
		}
	})

	t.Run("take", func(t *testing.T) {
		fn, _ := Get("take")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
			types.NewInteger(4),
			types.NewInteger(5),
		}, []interface{}{types.Collection{types.NewInteger(3)}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 3 {
			t.Errorf("expected 3, got %d", result.Count())
		}
	})

	t.Run("single", func(t *testing.T) {
		fn, _ := Get("single")

		// Single element
		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(42)}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 42 {
			t.Errorf("expected 42, got %d", result[0].(types.Integer).Value())
		}

		// Multiple elements - should error
		_, err = fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
		}, nil)
		if err == nil {
			t.Error("expected error for multiple elements")
		}
	})
}

func TestRegistryFunctions(t *testing.T) {
	t.Run("list functions", func(t *testing.T) {
		names := List()
		if len(names) == 0 {
			t.Error("expected registered functions")
		}

		// Check some expected functions
		expected := []string{"empty", "exists", "count", "first", "last"}
		for _, name := range expected {
			if !Has(name) {
				t.Errorf("expected function '%s' to be registered", name)
			}
		}
	})

	t.Run("get registry", func(t *testing.T) {
		registry := GetRegistry()
		if registry == nil {
			t.Error("expected registry to not be nil")
		}
		// Verify it has functions
		if !registry.Has("empty") {
			t.Error("expected registry to have 'empty' function")
		}
	})
}

func TestFilteringFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("where with boolean collection", func(t *testing.T) {
		fn, _ := Get("where")

		// Filter using pre-evaluated boolean collection
		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
		}, []interface{}{types.Collection{
			types.NewBoolean(true),
			types.NewBoolean(false),
			types.NewBoolean(true),
		}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2 elements, got %d", result.Count())
		}
		if result[0].(types.Integer).Value() != 1 {
			t.Errorf("expected first element 1, got %d", result[0].(types.Integer).Value())
		}
		if result[1].(types.Integer).Value() != 3 {
			t.Errorf("expected second element 3, got %d", result[1].(types.Integer).Value())
		}
	})

	t.Run("where no args", func(t *testing.T) {
		fn, _ := Get("where")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err == nil {
			t.Error("expected error for where without arguments")
		}
	})

	t.Run("where with empty collection", func(t *testing.T) {
		fn, _ := Get("where")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty result")
		}
	})

	t.Run("select with collection", func(t *testing.T) {
		fn, _ := Get("select")

		// Select returns the provided collection
		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
		}, []interface{}{types.Collection{
			types.NewString("a"),
			types.NewString("b"),
		}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2 elements, got %d", result.Count())
		}
		if result[0].(types.String).Value() != "a" {
			t.Errorf("expected 'a', got %s", result[0].(types.String).Value())
		}
	})

	t.Run("select no args", func(t *testing.T) {
		fn, _ := Get("select")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err == nil {
			t.Error("expected error for select without arguments")
		}
	})

	t.Run("select with non-collection", func(t *testing.T) {
		fn, _ := Get("select")

		result, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, []interface{}{"not a collection"})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty result for non-collection argument")
		}
	})

	t.Run("repeat no args", func(t *testing.T) {
		fn, _ := Get("repeat")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err == nil {
			t.Error("expected error for repeat without arguments")
		}
	})

	t.Run("repeat returns input", func(t *testing.T) {
		fn, _ := Get("repeat")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
		}, []interface{}{types.Collection{}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2 elements, got %d", result.Count())
		}
	})

	t.Run("ofType filters by type", func(t *testing.T) {
		fn, _ := Get("ofType")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewString("hello"),
			types.NewInteger(2),
			types.NewBoolean(true),
		}, []interface{}{types.Collection{types.NewString("Integer")}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2 integers, got %d", result.Count())
		}
	})

	t.Run("ofType no args", func(t *testing.T) {
		fn, _ := Get("ofType")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err == nil {
			t.Error("expected error for ofType without arguments")
		}
	})

	t.Run("ofType with string type name", func(t *testing.T) {
		fn, _ := Get("ofType")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewString("hello"),
		}, []interface{}{"String"})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 1 {
			t.Errorf("expected 1 string, got %d", result.Count())
		}
	})

	t.Run("ofType with empty type name", func(t *testing.T) {
		fn, _ := Get("ofType")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
		}, []interface{}{types.Collection{}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty result for empty type name")
		}
	})
}

func TestAdditionalSubsettingFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("intersect", func(t *testing.T) {
		fn, _ := Get("intersect")

		// Common elements
		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
		}, []interface{}{types.Collection{
			types.NewInteger(2),
			types.NewInteger(3),
			types.NewInteger(4),
		}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2 elements, got %d", result.Count())
		}
	})

	t.Run("intersect no args", func(t *testing.T) {
		fn, _ := Get("intersect")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err == nil {
			t.Error("expected error for intersect without arguments")
		}
	})

	t.Run("intersect invalid type", func(t *testing.T) {
		fn, _ := Get("intersect")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, []interface{}{"not a collection"})
		if err == nil {
			t.Error("expected error for invalid argument type")
		}
	})

	t.Run("intersect empty", func(t *testing.T) {
		fn, _ := Get("intersect")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{types.NewInteger(1)}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty result")
		}
	})

	t.Run("exclude", func(t *testing.T) {
		fn, _ := Get("exclude")

		// Exclude elements
		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
			types.NewInteger(3),
		}, []interface{}{types.Collection{
			types.NewInteger(2),
		}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2 elements, got %d", result.Count())
		}
	})

	t.Run("exclude no args", func(t *testing.T) {
		fn, _ := Get("exclude")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err == nil {
			t.Error("expected error for exclude without arguments")
		}
	})

	t.Run("exclude invalid type", func(t *testing.T) {
		fn, _ := Get("exclude")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, []interface{}{"not a collection"})
		if err == nil {
			t.Error("expected error for invalid argument type")
		}
	})

	t.Run("exclude all", func(t *testing.T) {
		fn, _ := Get("exclude")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
		}, []interface{}{types.Collection{
			types.NewInteger(1),
			types.NewInteger(2),
		}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty result when excluding all")
		}
	})

	t.Run("skip no args", func(t *testing.T) {
		fn, _ := Get("skip")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err == nil {
			t.Error("expected error for skip without arguments")
		}
	})

	t.Run("skip invalid type", func(t *testing.T) {
		fn, _ := Get("skip")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, []interface{}{"not integer"})
		if err == nil {
			t.Error("expected error for invalid argument type")
		}
	})

	t.Run("take no args", func(t *testing.T) {
		fn, _ := Get("take")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, nil)
		if err == nil {
			t.Error("expected error for take without arguments")
		}
	})

	t.Run("take invalid type", func(t *testing.T) {
		fn, _ := Get("take")

		_, err := fn.Fn(ctx, types.Collection{types.NewInteger(1)}, []interface{}{"not integer"})
		if err == nil {
			t.Error("expected error for invalid argument type")
		}
	})

	t.Run("single empty", func(t *testing.T) {
		fn, _ := Get("single")

		_, err := fn.Fn(ctx, types.Collection{}, nil)
		if err == nil {
			t.Error("expected error for empty collection")
		}
	})

	t.Run("last empty", func(t *testing.T) {
		fn, _ := Get("last")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty result")
		}
	})

	t.Run("tail empty", func(t *testing.T) {
		fn, _ := Get("tail")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty result")
		}
	})
}
