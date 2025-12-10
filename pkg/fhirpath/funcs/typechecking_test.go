package funcs

import (
	"testing"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func TestIsFunction(t *testing.T) {
	fn, ok := Get("is")
	if !ok {
		t.Fatal("is function not registered")
	}

	tests := []struct {
		name     string
		input    types.Collection
		args     []interface{}
		expected bool
		isEmpty  bool
	}{
		{
			name:     "string is String",
			input:    types.Collection{types.NewString("hello")},
			args:     []interface{}{"String"},
			expected: true,
		},
		{
			name:     "string is not Integer",
			input:    types.Collection{types.NewString("hello")},
			args:     []interface{}{"Integer"},
			expected: false,
		},
		{
			name:     "integer is Integer",
			input:    types.Collection{types.NewInteger(42)},
			args:     []interface{}{"Integer"},
			expected: true,
		},
		{
			name:     "boolean is Boolean",
			input:    types.Collection{types.NewBoolean(true)},
			args:     []interface{}{"Boolean"},
			expected: true,
		},
		{
			name:    "empty input returns empty",
			input:   types.Collection{},
			args:    []interface{}{"String"},
			isEmpty: true,
		},
		{
			name:     "case insensitive match",
			input:    types.Collection{types.NewString("hello")},
			args:     []interface{}{"string"},
			expected: true,
		},
	}

	ctx := eval.NewContext([]byte(`{}`))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fn.Fn(ctx, tt.input, tt.args)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.isEmpty {
				if !result.Empty() {
					t.Errorf("expected empty result, got %v", result)
				}
				return
			}

			if result.Empty() {
				t.Fatal("unexpected empty result")
			}

			b, ok := result[0].(types.Boolean)
			if !ok {
				t.Fatalf("expected Boolean, got %T", result[0])
			}

			if b.Bool() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, b.Bool())
			}
		})
	}
}

func TestIsFunctionSingletonError(t *testing.T) {
	fn, _ := Get("is")
	ctx := eval.NewContext([]byte(`{}`))

	// Multiple items should return error
	input := types.Collection{
		types.NewString("a"),
		types.NewString("b"),
	}

	_, err := fn.Fn(ctx, input, []interface{}{"String"})
	if err == nil {
		t.Error("expected singleton error for multiple items")
	}
}

func TestIsFunctionNoArgs(t *testing.T) {
	fn, _ := Get("is")
	ctx := eval.NewContext([]byte(`{}`))

	input := types.Collection{types.NewString("test")}

	_, err := fn.Fn(ctx, input, []interface{}{})
	if err == nil {
		t.Error("expected error for no arguments")
	}
}
