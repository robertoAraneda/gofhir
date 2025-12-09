// constraints.go demonstrates FHIRPath constraint expressions.
// These are the type of expressions used in FHIR StructureDefinition constraints.
// This includes:
// - Common FHIR constraints (pat-1, obs-3, etc.)
// - Boolean logic (and, or, implies)
// - all() and exists() patterns
package main

import (
	"fmt"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath"
)

// RunConstraintExamples demonstrates constraint-style expressions
func RunConstraintExamples(patient, observation []byte) {
	fmt.Println("\n" + separator)
	fmt.Println("CONSTRAINT EXPRESSIONS")
	fmt.Println(separator)

	fmt.Println("\n--- Patient Constraints ---")
	patientConstraints := []struct {
		name       string
		expression string
	}{
		{
			name:       "pat-1: contact must have details",
			expression: "contact.all(name.exists() or telecom.exists() or address.exists() or organization.exists())",
		},
		{
			name:       "has-name: patient must have name",
			expression: "name.count() >= 1",
		},
		{
			name:       "has-identifier",
			expression: "identifier.exists()",
		},
		{
			name:       "active-patient",
			expression: "active.exists() and active = true",
		},
		{
			name:       "valid-gender",
			expression: "gender.exists() implies gender.matches('male|female|other|unknown')",
		},
		{
			name:       "birthdate-past",
			expression: "birthDate.exists() implies birthDate <= today()",
		},
	}

	for _, c := range patientConstraints {
		evaluateConstraint(patient, c.name, c.expression)
	}

	fmt.Println("\n--- Observation Constraints ---")
	observationConstraints := []struct {
		name       string
		expression string
	}{
		{
			name:       "obs-3: value or dataAbsentReason",
			expression: "value.exists() or dataAbsentReason.exists()",
		},
		{
			name:       "obs-6: exclusive dataAbsentReason",
			expression: "dataAbsentReason.empty() or value.empty()",
		},
		{
			name:       "has-status",
			expression: "status.exists()",
		},
		{
			name:       "has-code",
			expression: "code.exists()",
		},
		{
			name: "valid-status",
			//nolint:misspell // "cancelled" is the official FHIR R4 ObservationStatus value
			expression: "status.matches('registered|preliminary|final|amended|corrected|cancelled|entered-in-error|unknown')",
		},
	}

	for _, c := range observationConstraints {
		evaluateConstraint(observation, c.name, c.expression)
	}

	fmt.Println("\n--- Boolean Logic ---")
	fmt.Println("Using 'and', 'or', 'not', 'implies', 'xor'")
	logicExpressions := []struct {
		name       string
		expression string
	}{
		{"and operator", "active = true and gender = 'male'"},
		{"or operator", "gender = 'male' or gender = 'female'"},
		{"not operator", "deceased.exists().not()"},
		{"implies operator", "active.exists() implies active = true"},
		{"xor operator", "true xor false"},
		{"combined logic", "(active = true) and (gender = 'male' or gender = 'female')"},
	}

	for _, e := range logicExpressions {
		evaluateConstraint(patient, e.name, e.expression)
	}

	fmt.Println("\n--- Collection Patterns ---")
	fmt.Println("Common patterns for constraint validation")
	patternExpressions := []struct {
		name       string
		expression string
	}{
		{"all() with condition", "name.all(family.exists())"},
		{"any/exists with condition", "telecom.where(system = 'email').exists()"},
		{"count constraint", "identifier.count() >= 1"},
		{"empty check", "deceased.empty()"},
		{"not empty check", "name.empty().not()"},
		{"nested all()", "contact.all(telecom.all(value.exists()))"},
	}

	for _, e := range patternExpressions {
		evaluateConstraint(patient, e.name, e.expression)
	}

	fmt.Println("\n--- Complex Validation Scenarios ---")
	complexConstraints := []struct {
		name       string
		expression string
		resource   []byte
	}{
		{
			name:       "BP systolic < 200",
			expression: "component.where(code.coding.code = '8480-6').valueQuantity.value < 200",
			resource:   observation,
		},
		{
			name:       "BP diastolic < 120",
			expression: "component.where(code.coding.code = '8462-4').valueQuantity.value < 120",
			resource:   observation,
		},
		{
			name:       "has home contact",
			expression: "telecom.where(use = 'home').exists() or address.where(use = 'home').exists()",
			resource:   patient,
		},
		{
			name:       "official name present",
			expression: "name.where(use = 'official').exists()",
			resource:   patient,
		},
	}

	for _, c := range complexConstraints {
		evaluateConstraint(c.resource, c.name, c.expression)
	}
}

// evaluateConstraint evaluates a constraint and shows PASS/FAIL status
func evaluateConstraint(resource []byte, name, expression string) {
	result, err := fhirpath.Evaluate(resource, expression)
	if err != nil {
		fmt.Printf("  [ERROR] %s\n", name)
		fmt.Printf("          Expression: %s\n", expression)
		fmt.Printf("          Error: %v\n", err)
		return
	}

	passed := isTruthy(result)
	status := "PASS"
	if !passed {
		status = "FAIL"
	}
	fmt.Printf("  [%s] %s\n", status, name)
	fmt.Printf("         %s\n", expression)
}

// isTruthy checks if a FHIRPath result is truthy
func isTruthy(result interface{}) bool {
	switch v := result.(type) {
	case []interface{}:
		if len(v) == 0 {
			return false
		}
		if len(v) == 1 {
			return isTruthy(v[0])
		}
		return true // Non-empty collection is truthy
	case fhirpath.Collection:
		if len(v) == 0 {
			return false
		}
		if len(v) == 1 {
			return isTruthy(v[0])
		}
		return true
	case bool:
		return v
	case nil:
		return false
	default:
		return true // Any non-nil value is truthy
	}
}
