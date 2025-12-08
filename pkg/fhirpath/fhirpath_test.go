package fhirpath

import (
	"testing"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

// Test data
var patientJSON = []byte(`{
	"resourceType": "Patient",
	"id": "123",
	"active": true,
	"name": [
		{
			"use": "official",
			"family": "Doe",
			"given": ["John", "James"]
		},
		{
			"use": "nickname",
			"given": ["Johnny"]
		}
	],
	"birthDate": "1990-01-15",
	"address": [
		{
			"city": "Boston",
			"state": "MA"
		}
	]
}`)

var simpleJSON = []byte(`{
	"value": 42,
	"decimal": 3.14,
	"text": "hello",
	"active": true,
	"items": [1, 2, 3, 4, 5]
}`)

func TestCompile(t *testing.T) {
	t.Run("valid expression", func(t *testing.T) {
		expr, err := Compile("Patient.name.given")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if expr.String() != "Patient.name.given" {
			t.Errorf("expected 'Patient.name.given', got '%s'", expr.String())
		}
	})

	t.Run("empty expression", func(t *testing.T) {
		_, err := Compile("")
		if err == nil {
			t.Error("expected error for empty expression")
		}
	})

	t.Run("invalid syntax", func(t *testing.T) {
		_, err := Compile("Patient.name..")
		if err == nil {
			t.Error("expected error for invalid syntax")
		}
	})
}

func TestLiterals(t *testing.T) {
	t.Run("boolean true", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "true")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("boolean false", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "false")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, false)
	})

	t.Run("integer", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "42")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertIntegerResult(t, result, 42)
	})

	t.Run("decimal", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "3.14")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Empty() || result[0].Type() != "Decimal" {
			t.Errorf("expected Decimal, got %v", result)
		}
	})

	t.Run("string", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "'hello world'")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertStringResult(t, result, "hello world")
	})

	t.Run("empty collection", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "{}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Empty() {
			t.Errorf("expected empty collection, got %v", result)
		}
	})
}

func TestNavigation(t *testing.T) {
	t.Run("simple path", func(t *testing.T) {
		result, err := Evaluate(patientJSON, "Patient.id")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertStringResult(t, result, "123")
	})

	t.Run("boolean field", func(t *testing.T) {
		result, err := Evaluate(patientJSON, "Patient.active")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("nested path", func(t *testing.T) {
		result, err := Evaluate(patientJSON, "Patient.name.family")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should return all family names
		if result.Count() != 1 {
			t.Errorf("expected 1 family name, got %d", result.Count())
		}
		assertStringResult(t, result, "Doe")
	})

	t.Run("array navigation", func(t *testing.T) {
		result, err := Evaluate(patientJSON, "Patient.name.given")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should return all given names from all name entries
		if result.Count() != 3 {
			t.Errorf("expected 3 given names, got %d: %v", result.Count(), result)
		}
	})

	t.Run("non-existent path", func(t *testing.T) {
		result, err := Evaluate(patientJSON, "Patient.nonexistent")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Empty() {
			t.Errorf("expected empty collection for non-existent path, got %v", result)
		}
	})
}

func TestArithmeticOperators(t *testing.T) {
	t.Run("addition integers", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "2 + 3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertIntegerResult(t, result, 5)
	})

	t.Run("subtraction", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "10 - 3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertIntegerResult(t, result, 7)
	})

	t.Run("multiplication", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "4 * 5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertIntegerResult(t, result, 20)
	})

	t.Run("division", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "10 / 4")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Division returns Decimal
		if result.Empty() || result[0].Type() != "Decimal" {
			t.Errorf("expected Decimal, got %v", result)
		}
	})

	t.Run("integer division", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "10 div 3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertIntegerResult(t, result, 3)
	})

	t.Run("modulo", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "10 mod 3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertIntegerResult(t, result, 1)
	})

	t.Run("negation", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "-5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertIntegerResult(t, result, -5)
	})

	t.Run("string concatenation with +", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "'hello' + ' world'")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertStringResult(t, result, "hello world")
	})

	t.Run("string concatenation with &", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "'hello' & ' world'")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertStringResult(t, result, "hello world")
	})
}

func TestComparisonOperators(t *testing.T) {
	t.Run("less than true", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "5 < 10")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("less than false", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "10 < 5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, false)
	})

	t.Run("greater than", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "10 > 5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("less or equal", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "5 <= 5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("greater or equal", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "5 >= 10")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, false)
	})
}

func TestEqualityOperators(t *testing.T) {
	t.Run("equal integers", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "5 = 5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("not equal", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "5 != 10")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("string equality", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "'hello' = 'hello'")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("equivalence case insensitive", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "'HELLO' ~ 'hello'")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("not equivalent", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "'hello' !~ 'world'")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})
}

func TestBooleanOperators(t *testing.T) {
	t.Run("and true", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "true and true")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("and false", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "true and false")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, false)
	})

	t.Run("or true", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "false or true")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("or false", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "false or false")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, false)
	})

	t.Run("xor", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "true xor false")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("implies true", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "false implies true")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("implies false", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "true implies false")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, false)
	})
}

func TestCollectionOperators(t *testing.T) {
	t.Run("union", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "(1 | 2) | (2 | 3)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Union removes duplicates
		if result.Count() != 3 {
			t.Errorf("expected 3 elements, got %d: %v", result.Count(), result)
		}
	})

	t.Run("in membership", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "2 in (1 | 2 | 3)")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("contains", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "(1 | 2 | 3) contains 2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})
}

func TestIndexer(t *testing.T) {
	t.Run("array index", func(t *testing.T) {
		result, err := Evaluate(patientJSON, "Patient.name[0].family")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertStringResult(t, result, "Doe")
	})

	t.Run("index out of bounds", func(t *testing.T) {
		result, err := Evaluate(patientJSON, "Patient.name[10]")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Empty() {
			t.Errorf("expected empty for out of bounds, got %v", result)
		}
	})
}

func TestTypeOperators(t *testing.T) {
	t.Run("is type check", func(t *testing.T) {
		result, err := Evaluate(patientJSON, "Patient.active is Boolean")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})

	t.Run("as type cast", func(t *testing.T) {
		result, err := Evaluate(patientJSON, "Patient.active as Boolean")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertBooleanResult(t, result, true)
	})
}

func TestEmptyPropagation(t *testing.T) {
	t.Run("empty + value", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "{} + 5")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Empty() {
			t.Errorf("expected empty, got %v", result)
		}
	})

	t.Run("empty and true", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "{} and true")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// In three-valued logic, empty and true = empty
		if !result.Empty() {
			t.Errorf("expected empty, got %v", result)
		}
	})

	t.Run("empty and false", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "{} and false")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// In three-valued logic, empty and false = false
		assertBooleanResult(t, result, false)
	})
}

func TestParentheses(t *testing.T) {
	t.Run("precedence override", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "(2 + 3) * 4")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertIntegerResult(t, result, 20)
	})

	t.Run("default precedence", func(t *testing.T) {
		result, err := Evaluate(simpleJSON, "2 + 3 * 4")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertIntegerResult(t, result, 14)
	})
}

// Helper functions

func assertBooleanResult(t *testing.T, result types.Collection, expected bool) {
	t.Helper()
	if result.Empty() {
		t.Fatalf("expected boolean %v, got empty collection", expected)
	}
	if len(result) != 1 {
		t.Fatalf("expected single value, got %d: %v", len(result), result)
	}
	b, ok := result[0].(types.Boolean)
	if !ok {
		t.Fatalf("expected Boolean, got %s: %v", result[0].Type(), result[0])
	}
	if b.Bool() != expected {
		t.Errorf("expected %v, got %v", expected, b.Bool())
	}
}

func assertIntegerResult(t *testing.T, result types.Collection, expected int64) {
	t.Helper()
	if result.Empty() {
		t.Fatalf("expected integer %d, got empty collection", expected)
	}
	if len(result) != 1 {
		t.Fatalf("expected single value, got %d: %v", len(result), result)
	}
	i, ok := result[0].(types.Integer)
	if !ok {
		t.Fatalf("expected Integer, got %s: %v", result[0].Type(), result[0])
	}
	if i.Value() != expected {
		t.Errorf("expected %d, got %d", expected, i.Value())
	}
}

func assertStringResult(t *testing.T, result types.Collection, expected string) {
	t.Helper()
	if result.Empty() {
		t.Fatalf("expected string '%s', got empty collection", expected)
	}
	if len(result) != 1 {
		t.Fatalf("expected single value, got %d: %v", len(result), result)
	}
	s, ok := result[0].(types.String)
	if !ok {
		t.Fatalf("expected String, got %s: %v", result[0].Type(), result[0])
	}
	if s.Value() != expected {
		t.Errorf("expected '%s', got '%s'", expected, s.Value())
	}
}
