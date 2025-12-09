package fhirpath_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/robertoaraneda/gofhir/pkg/fhir/r4"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath"
)

// Test evaluating FHIRPath against JSON bytes
func TestEvaluateJSON(t *testing.T) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "example",
		"active": true,
		"name": [
			{
				"use": "official",
				"family": "Smith",
				"given": ["John", "James"]
			}
		],
		"birthDate": "1990-01-15"
	}`)

	tests := []struct {
		name      string
		expr      string
		wantCount int
		wantFirst string
		wantBool  *bool
	}{
		{
			name:      "simple path",
			expr:      "Patient.id",
			wantCount: 1,
			wantFirst: "example",
		},
		{
			name:      "nested path",
			expr:      "Patient.name.family",
			wantCount: 1,
			wantFirst: "Smith",
		},
		{
			name:      "array access",
			expr:      "Patient.name.given",
			wantCount: 2,
			wantFirst: "John",
		},
		{
			name:      "first function",
			expr:      "Patient.name.given.first()",
			wantCount: 1,
			wantFirst: "John",
		},
		{
			name:      "count function",
			expr:      "Patient.name.given.count()",
			wantCount: 1,
			wantFirst: "2",
		},
		{
			name:      "exists function",
			expr:      "Patient.name.exists()",
			wantCount: 1,
			wantBool:  boolPtr(true),
		},
		{
			name:      "empty check",
			expr:      "Patient.telecom.empty()",
			wantCount: 1,
			wantBool:  boolPtr(true),
		},
		{
			name:      "where filter",
			expr:      "Patient.name.where(use = 'official').family",
			wantCount: 1,
			wantFirst: "Smith",
		},
		{
			name:      "boolean field",
			expr:      "Patient.active",
			wantCount: 1,
			wantBool:  boolPtr(true),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fhirpath.Evaluate(patient, tt.expr)
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}

			if len(result) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(result), tt.wantCount)
			}

			if tt.wantFirst != "" && len(result) > 0 {
				if got := result[0].String(); got != tt.wantFirst {
					t.Errorf("first result = %q, want %q", got, tt.wantFirst)
				}
			}

			if tt.wantBool != nil && len(result) > 0 {
				got, err := result.ToBoolean()
				if err != nil {
					t.Errorf("ToBoolean() error = %v", err)
				} else if got != *tt.wantBool {
					t.Errorf("boolean result = %v, want %v", got, *tt.wantBool)
				}
			}
		})
	}
}

// Test evaluating against Go structs using EvaluateResource
func TestEvaluateResource(t *testing.T) {
	// Create patient using fluent builder - resourceType is set automatically
	patient := r4.NewPatientBuilder().
		SetId("test-patient").
		SetActive(true).
		AddName(r4.HumanName{
			Use:    ptrTo(r4.NameUse("official")),
			Family: strPtr("Doe"),
			Given:  []string{"Jane", "Marie"},
		}).
		SetGender(r4.AdministrativeGenderFemale).
		SetBirthDate("1985-06-20").
		Build()

	// Test with EvaluateResource - should work now that MarshalJSON includes resourceType
	result, err := fhirpath.EvaluateResource(patient, "Patient.name.given.first()")
	if err != nil {
		t.Fatalf("EvaluateResource() error = %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("got %d results, want 1", len(result))
	}

	if got := result[0].String(); got != "Jane" {
		t.Errorf("got %q, want %q", got, "Jane")
	}

	// Verify JSON serialization includes resourceType
	jsonBytes, err := json.Marshal(patient)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &data); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if data["resourceType"] != "Patient" {
		t.Errorf("resourceType = %v, want Patient", data["resourceType"])
	}
}

func ptrTo[T any](v T) *T {
	return &v
}

// Test ResourceJSON wrapper for efficient repeated evaluation
func TestResourceJSON(t *testing.T) {
	// Use fluent builder to create patient with resourceType
	patient := r4.NewPatientBuilder().
		SetId("cached-patient").
		AddName(r4.HumanName{Family: strPtr("Cached")}).
		Build()

	rj, err := fhirpath.NewResourceJSON(patient)
	if err != nil {
		t.Fatalf("NewResourceJSON() error = %v", err)
	}

	// Evaluate multiple expressions efficiently
	expressions := []string{
		"Patient.id",
		"Patient.name.family",
		"Patient.name.exists()",
	}

	for _, expr := range expressions {
		result, err := rj.EvaluateCached(expr)
		if err != nil {
			t.Errorf("EvaluateCached(%q) error = %v", expr, err)
		}
		if result.Empty() {
			t.Errorf("EvaluateCached(%q) returned empty", expr)
		}
	}
}

// Test expression caching
func TestExpressionCache(t *testing.T) {
	cache := fhirpath.NewExpressionCache(100)

	patient := []byte(`{"resourceType": "Patient", "id": "test"}`)

	// First call compiles and caches
	expr1, err := cache.Get("Patient.id")
	if err != nil {
		t.Fatalf("cache.Get() error = %v", err)
	}

	// Second call should return cached
	expr2, err := cache.Get("Patient.id")
	if err != nil {
		t.Fatalf("cache.Get() second call error = %v", err)
	}

	// Should be the same pointer
	if expr1 != expr2 {
		t.Error("cache should return same expression instance")
	}

	// Verify it works
	result, err := expr1.Evaluate(patient)
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if result[0].String() != "test" {
		t.Errorf("got %q, want %q", result[0].String(), "test")
	}

	// Check cache size
	if cache.Size() != 1 {
		t.Errorf("cache size = %d, want 1", cache.Size())
	}
}

// Test evaluation with options
func TestEvaluateWithOptions(t *testing.T) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "options-test",
		"name": [{"family": "Test"}]
	}`)

	expr := fhirpath.MustCompile("Patient.id")

	// Test with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	result, err := expr.EvaluateWithOptions(patient,
		fhirpath.WithContext(ctx),
		fhirpath.WithTimeout(1*time.Second),
		fhirpath.WithMaxDepth(50),
	)

	if err != nil {
		t.Fatalf("EvaluateWithOptions() error = %v", err)
	}

	if result[0].String() != "options-test" {
		t.Errorf("got %q, want %q", result[0].String(), "options-test")
	}
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "helper-test",
		"active": true,
		"name": [{"family": "Helper"}, {"family": "Test"}]
	}`)

	t.Run("EvaluateToBoolean", func(t *testing.T) {
		result, err := fhirpath.EvaluateToBoolean(patient, "Patient.active")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if !result {
			t.Error("expected true")
		}
	})

	t.Run("EvaluateToString", func(t *testing.T) {
		result, err := fhirpath.EvaluateToString(patient, "Patient.id")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if result != "helper-test" {
			t.Errorf("got %q, want %q", result, "helper-test")
		}
	})

	t.Run("EvaluateToStrings", func(t *testing.T) {
		result, err := fhirpath.EvaluateToStrings(patient, "Patient.name.family")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if len(result) != 2 {
			t.Errorf("got %d results, want 2", len(result))
		}
	})

	t.Run("Exists", func(t *testing.T) {
		result, err := fhirpath.Exists(patient, "Patient.name")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if !result {
			t.Error("expected true")
		}
	})

	t.Run("Count", func(t *testing.T) {
		result, err := fhirpath.Count(patient, "Patient.name")
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if result != 2 {
			t.Errorf("got %d, want 2", result)
		}
	})
}

// Test FHIR-specific functions
func TestFHIRFunctions(t *testing.T) {
	t.Run("extension", func(t *testing.T) {
		patient := []byte(`{
			"resourceType": "Patient",
			"id": "ext-test",
			"extension": [
				{
					"url": "http://example.org/birthPlace",
					"valueString": "Boston"
				},
				{
					"url": "http://example.org/race",
					"valueCode": "white"
				}
			]
		}`)

		result, err := fhirpath.Evaluate(patient, "Patient.extension('http://example.org/birthPlace')")
		if err != nil {
			t.Fatalf("error = %v", err)
		}

		if result.Empty() {
			t.Error("expected extension to be found")
		}
	})

	t.Run("hasExtension", func(t *testing.T) {
		patient := []byte(`{
			"resourceType": "Patient",
			"extension": [{"url": "http://example.org/test", "valueBoolean": true}]
		}`)

		result, err := fhirpath.EvaluateToBoolean(patient, "Patient.hasExtension('http://example.org/test')")
		if err != nil {
			t.Fatalf("error = %v", err)
		}

		if !result {
			t.Error("expected hasExtension to return true")
		}
	})
}

// Test arithmetic operators
func TestArithmetic(t *testing.T) {
	patient := []byte(`{"resourceType": "Patient"}`)

	tests := []struct {
		expr string
		want string
	}{
		{"2 + 3", "5"},
		{"10 - 4", "6"},
		{"3 * 4", "12"},
		{"15 / 3", "5"},
		{"17 div 5", "3"},
		{"17 mod 5", "2"},
		{"-5", "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := fhirpath.Evaluate(patient, tt.expr)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := result[0].String(); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

// Test comparison operators
func TestComparison(t *testing.T) {
	patient := []byte(`{"resourceType": "Patient"}`)

	tests := []struct {
		expr string
		want bool
	}{
		{"5 < 10", true},
		{"5 > 10", false},
		{"5 <= 5", true},
		{"5 >= 5", true},
		{"5 = 5", true},
		{"5 != 10", true},
		{"'abc' = 'abc'", true},
		{"'ABC' ~ 'abc'", true}, // equivalence is case-insensitive
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := fhirpath.EvaluateToBoolean(patient, tt.expr)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

// Test boolean logic
func TestBooleanLogic(t *testing.T) {
	patient := []byte(`{"resourceType": "Patient"}`)

	tests := []struct {
		expr string
		want bool
	}{
		{"true and true", true},
		{"true and false", false},
		{"true or false", true},
		{"false or false", false},
		{"true xor false", true},
		{"true xor true", false},
		{"false implies true", true},
		{"true implies false", false},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := fhirpath.EvaluateToBoolean(patient, tt.expr)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if result != tt.want {
				t.Errorf("got %v, want %v", result, tt.want)
			}
		})
	}
}

// Test string functions
func TestStringFunctions(t *testing.T) {
	patient := []byte(`{"resourceType": "Patient"}`)

	tests := []struct {
		expr string
		want string
	}{
		{"'Hello'.lower()", "hello"},
		{"'hello'.upper()", "HELLO"},
		{"'hello world'.startsWith('hello')", "true"},
		{"'hello world'.endsWith('world')", "true"},
		{"'hello world'.contains('lo wo')", "true"},
		{"'hello'.length()", "5"},
		{"'hello world'.replace('world', 'there')", "hello there"},
		{"'a,b,c'.split(',').count()", "3"},
	}

	for _, tt := range tests {
		t.Run(tt.expr, func(t *testing.T) {
			result, err := fhirpath.Evaluate(patient, tt.expr)
			if err != nil {
				t.Fatalf("error = %v", err)
			}
			if got := result[0].String(); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

// Benchmark compilation
func BenchmarkCompile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = fhirpath.Compile("Patient.name.given.first()")
	}
}

// Benchmark cached compilation
func BenchmarkCompileCached(b *testing.B) {
	cache := fhirpath.NewExpressionCache(100)
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("Patient.name.given.first()")
	}
}

// Benchmark evaluation
func BenchmarkEvaluate(b *testing.B) {
	patient := []byte(`{
		"resourceType": "Patient",
		"name": [{"given": ["John", "James"]}]
	}`)
	expr := fhirpath.MustCompile("Patient.name.given.first()")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(patient)
	}
}

// Benchmark struct evaluation
func BenchmarkEvaluateResource(b *testing.B) {
	patient := &r4.Patient{
		Id: strPtr("bench"),
		Name: []r4.HumanName{
			{Given: []string{"John", "James"}},
		},
	}

	// Pre-serialize for fair comparison
	jsonBytes, _ := json.Marshal(patient)
	expr := fhirpath.MustCompile("Patient.name.given.first()")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(jsonBytes)
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
