package fhirpath

import (
	"testing"
)

var patient = []byte(`{
	"resourceType": "Patient",
	"id": "example",
	"active": true,
	"name": [
		{
			"use": "official",
			"family": "Chalmers",
			"given": ["Peter", "James"]
		},
		{
			"use": "usual",
			"given": ["Jim"]
		}
	],
	"telecom": [
		{
			"system": "phone",
			"value": "(03) 5555 6473"
		}
	],
	"gender": "male",
	"birthDate": "1974-12-25"
}`)

func BenchmarkCompile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Compile("Patient.name.given")
	}
}

func BenchmarkEvaluateSimple(b *testing.B) {
	expr := MustCompile("Patient.id")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(patient)
	}
}

func BenchmarkEvaluateNested(b *testing.B) {
	expr := MustCompile("Patient.name.given")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(patient)
	}
}

func BenchmarkEvaluateWithFunction(b *testing.B) {
	expr := MustCompile("Patient.name.given.count()")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(patient)
	}
}

func BenchmarkEvaluateComplex(b *testing.B) {
	expr := MustCompile("Patient.name.first().given.join(', ')")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(patient)
	}
}

func BenchmarkEvaluateArithmetic(b *testing.B) {
	expr := MustCompile("2 + 3 * 4 - 1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(patient)
	}
}

func BenchmarkEvaluateString(b *testing.B) {
	expr := MustCompile("'Hello'.lower().startsWith('hel')")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(patient)
	}
}

func BenchmarkEvaluateMath(b *testing.B) {
	expr := MustCompile("16.sqrt().power(2)")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(patient)
	}
}

func BenchmarkDirectEvaluate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Evaluate(patient, "Patient.name.given")
	}
}

func BenchmarkEvaluateBoolean(b *testing.B) {
	expr := MustCompile("true and false or true")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(patient)
	}
}

func BenchmarkEvaluateComparison(b *testing.B) {
	expr := MustCompile("5 < 10 and 10 > 5")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(patient)
	}
}

func BenchmarkEvaluateExists(b *testing.B) {
	expr := MustCompile("Patient.name.exists()")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(patient)
	}
}

func BenchmarkEvaluateEmpty(b *testing.B) {
	expr := MustCompile("Patient.name.empty()")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = expr.Evaluate(patient)
	}
}
