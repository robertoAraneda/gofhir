package funcs

import (
	"testing"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func TestStringFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("startsWith", func(t *testing.T) {
		fn, _ := Get("startsWith")

		// True case
		result, err := fn.Fn(ctx, types.Collection{types.NewString("Hello")},
			[]interface{}{types.Collection{types.NewString("Hel")}})
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true")
		}

		// False case
		result, err = fn.Fn(ctx, types.Collection{types.NewString("Hello")},
			[]interface{}{types.Collection{types.NewString("llo")}})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false")
		}
	})

	t.Run("endsWith", func(t *testing.T) {
		fn, _ := Get("endsWith")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("Hello")},
			[]interface{}{types.Collection{types.NewString("llo")}})
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true")
		}
	})

	t.Run("contains", func(t *testing.T) {
		fn, _ := Get("contains")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("Hello World")},
			[]interface{}{types.Collection{types.NewString("lo Wo")}})
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true")
		}
	})

	t.Run("replace", func(t *testing.T) {
		fn, _ := Get("replace")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("Hello World")},
			[]interface{}{
				types.Collection{types.NewString("World")},
				types.Collection{types.NewString("FHIRPath")},
			})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "Hello FHIRPath" {
			t.Errorf("expected 'Hello FHIRPath', got '%s'", result[0].(types.String).Value())
		}
	})

	t.Run("indexOf", func(t *testing.T) {
		fn, _ := Get("indexOf")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("Hello")},
			[]interface{}{types.Collection{types.NewString("l")}})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 2 {
			t.Errorf("expected 2, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("substring", func(t *testing.T) {
		fn, _ := Get("substring")

		// With length
		result, err := fn.Fn(ctx, types.Collection{types.NewString("Hello")},
			[]interface{}{
				types.Collection{types.NewInteger(1)},
				types.Collection{types.NewInteger(3)},
			})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "ell" {
			t.Errorf("expected 'ell', got '%s'", result[0].(types.String).Value())
		}

		// Without length
		result, err = fn.Fn(ctx, types.Collection{types.NewString("Hello")},
			[]interface{}{types.Collection{types.NewInteger(2)}})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "llo" {
			t.Errorf("expected 'llo', got '%s'", result[0].(types.String).Value())
		}
	})

	t.Run("lower", func(t *testing.T) {
		fn, _ := Get("lower")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("HELLO")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "hello" {
			t.Errorf("expected 'hello', got '%s'", result[0].(types.String).Value())
		}
	})

	t.Run("upper", func(t *testing.T) {
		fn, _ := Get("upper")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("hello")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "HELLO" {
			t.Errorf("expected 'HELLO', got '%s'", result[0].(types.String).Value())
		}
	})

	t.Run("length", func(t *testing.T) {
		fn, _ := Get("length")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("Hello")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != 5 {
			t.Errorf("expected 5, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("toChars", func(t *testing.T) {
		fn, _ := Get("toChars")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("Hi")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 2 {
			t.Errorf("expected 2, got %d", result.Count())
		}
		if result[0].(types.String).Value() != "H" {
			t.Errorf("expected 'H', got '%s'", result[0].(types.String).Value())
		}
	})

	t.Run("split", func(t *testing.T) {
		fn, _ := Get("split")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("a,b,c")},
			[]interface{}{types.Collection{types.NewString(",")}})
		if err != nil {
			t.Fatal(err)
		}
		if result.Count() != 3 {
			t.Errorf("expected 3, got %d", result.Count())
		}
	})

	t.Run("join", func(t *testing.T) {
		fn, _ := Get("join")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewString("a"),
			types.NewString("b"),
			types.NewString("c"),
		}, []interface{}{types.Collection{types.NewString("-")}})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "a-b-c" {
			t.Errorf("expected 'a-b-c', got '%s'", result[0].(types.String).Value())
		}
	})

	t.Run("trim", func(t *testing.T) {
		fn, _ := Get("trim")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("  hello  ")}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "hello" {
			t.Errorf("expected 'hello', got '%s'", result[0].(types.String).Value())
		}
	})

	t.Run("matches", func(t *testing.T) {
		fn, _ := Get("matches")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("test123")},
			[]interface{}{types.Collection{types.NewString("[a-z]+[0-9]+")}})
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true")
		}
	})

	t.Run("replaceMatches", func(t *testing.T) {
		fn, _ := Get("replaceMatches")

		// Replace digits with X
		result, err := fn.Fn(ctx, types.Collection{types.NewString("test123")},
			[]interface{}{
				types.Collection{types.NewString("[0-9]")},
				types.Collection{types.NewString("X")},
			})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "testXXX" {
			t.Errorf("expected 'testXXX', got '%s'", result[0].(types.String).Value())
		}
	})
}

func TestAdditionalStringFunctions(t *testing.T) {
	ctx := eval.NewContext([]byte(`{}`))

	t.Run("startsWith empty", func(t *testing.T) {
		fn, _ := Get("startsWith")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{types.NewString("test")}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for startsWith on empty")
		}
	})

	t.Run("endsWith empty", func(t *testing.T) {
		fn, _ := Get("endsWith")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{types.NewString("test")}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for endsWith on empty")
		}
	})

	t.Run("contains empty", func(t *testing.T) {
		fn, _ := Get("contains")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{types.NewString("test")}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for contains on empty")
		}
	})

	t.Run("replace empty", func(t *testing.T) {
		fn, _ := Get("replace")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{
			types.Collection{types.NewString("a")},
			types.Collection{types.NewString("b")},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for replace on empty")
		}
	})

	t.Run("indexOf not found", func(t *testing.T) {
		fn, _ := Get("indexOf")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("Hello")},
			[]interface{}{types.Collection{types.NewString("xyz")}})
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Integer).Value() != -1 {
			t.Errorf("expected -1 for not found, got %d", result[0].(types.Integer).Value())
		}
	})

	t.Run("indexOf empty", func(t *testing.T) {
		fn, _ := Get("indexOf")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{types.NewString("test")}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for indexOf on empty")
		}
	})

	t.Run("substring negative start", func(t *testing.T) {
		fn, _ := Get("substring")

		result, err := fn.Fn(ctx, types.Collection{types.NewString("Hello")},
			[]interface{}{types.Collection{types.NewInteger(-1)}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for negative start")
		}
	})

	t.Run("substring empty", func(t *testing.T) {
		fn, _ := Get("substring")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{types.NewInteger(0)}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for substring on empty")
		}
	})

	t.Run("lower empty", func(t *testing.T) {
		fn, _ := Get("lower")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for lower on empty")
		}
	})

	t.Run("upper empty", func(t *testing.T) {
		fn, _ := Get("upper")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for upper on empty")
		}
	})

	t.Run("length empty", func(t *testing.T) {
		fn, _ := Get("length")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for length on empty")
		}
	})

	t.Run("toChars empty", func(t *testing.T) {
		fn, _ := Get("toChars")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for toChars on empty")
		}
	})

	t.Run("split empty", func(t *testing.T) {
		fn, _ := Get("split")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{types.NewString(",")}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for split on empty")
		}
	})

	t.Run("join without separator", func(t *testing.T) {
		fn, _ := Get("join")

		result, err := fn.Fn(ctx, types.Collection{
			types.NewString("a"),
			types.NewString("b"),
		}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.String).Value() != "ab" {
			t.Errorf("expected 'ab', got '%s'", result[0].(types.String).Value())
		}
	})

	t.Run("trim empty", func(t *testing.T) {
		fn, _ := Get("trim")

		result, err := fn.Fn(ctx, types.Collection{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for trim on empty")
		}
	})

	t.Run("matches empty", func(t *testing.T) {
		fn, _ := Get("matches")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{types.Collection{types.NewString(".*")}})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for matches on empty")
		}
	})

	t.Run("replaceMatches empty", func(t *testing.T) {
		fn, _ := Get("replaceMatches")

		result, err := fn.Fn(ctx, types.Collection{}, []interface{}{
			types.Collection{types.NewString(".*")},
			types.Collection{types.NewString("X")},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !result.Empty() {
			t.Error("expected empty for replaceMatches on empty")
		}
	})
}
