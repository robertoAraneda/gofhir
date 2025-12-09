// structural_validation.go demonstrates FHIR structural validation.
// This includes validation of:
// - Required fields
// - Cardinality (min/max)
// - Unknown elements (in strict mode)
// - Primitive data types
//
//nolint:errcheck // Example code intentionally ignores errors for brevity
package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/robertoaraneda/gofhir/pkg/validator"
)

// RunStructuralValidationExamples demonstrates structural validation features
func RunStructuralValidationExamples(ctx context.Context) {
	fmt.Println("\n" + separator)
	fmt.Println("STRUCTURAL VALIDATION EXAMPLES")
	fmt.Println(separator)

	v, err := createStructuralValidator()
	if err != nil {
		log.Printf("Failed to create validator: %v", err)
		return
	}

	// Example 1: Valid Patient with all required structure
	fmt.Println("\n--- 1. Valid Patient (complete structure) ---")
	validateValidPatientStructure(ctx, v)

	// Example 2: Patient with unknown field
	fmt.Println("\n--- 2. Patient with Unknown Field ---")
	validateUnknownField(ctx, v)

	// Example 3: Patient with invalid primitive types
	fmt.Println("\n--- 3. Invalid Primitive Types ---")
	validateInvalidPrimitiveTypes(ctx, v)

	// Example 4: Cardinality validation
	fmt.Println("\n--- 4. Cardinality Validation ---")
	validateCardinality(ctx, v)

	// Example 5: Complex resource structure (Observation)
	fmt.Println("\n--- 5. Complex Resource Structure ---")
	validateComplexStructure(ctx, v)

	// Example 6: Strict mode vs non-strict mode
	fmt.Println("\n--- 6. Strict Mode Comparison ---")
	validateStrictMode(ctx)
}

func createStructuralValidator() (*validator.Validator, error) {
	registry := validator.NewRegistry(validator.FHIRVersionR4)
	specsDir := filepath.Join("..", "..", "specs", "r4")
	_, err := registry.LoadR4Specs(specsDir)
	if err != nil {
		return nil, err
	}

	opts := validator.ValidatorOptions{
		ValidateConstraints: false, // Disable constraints for structural-only validation
		ValidateExtensions:  false,
		ValidateReferences:  false,
		StrictMode:          false,
	}
	return validator.NewValidator(registry, opts), nil
}

func validateValidPatientStructure(ctx context.Context, v *validator.Validator) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "example-valid",
		"active": true,
		"name": [{
			"use": "official",
			"family": "Smith",
			"given": ["John", "James"]
		}],
		"gender": "male",
		"birthDate": "1990-05-15",
		"telecom": [{
			"system": "phone",
			"value": "+1-555-0100",
			"use": "home"
		}],
		"address": [{
			"use": "home",
			"line": ["123 Main St"],
			"city": "Springfield",
			"state": "IL",
			"postalCode": "62701"
		}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Valid Patient Structure", result)
	fmt.Println("  -> All structural elements are valid")
}

func validateUnknownField(ctx context.Context, v *validator.Validator) {
	// Patient with an unknown field "customField"
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "unknown-field-example",
		"active": true,
		"customField": "this field does not exist in FHIR",
		"name": [{"family": "Doe"}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Unknown Field", result)

	// In non-strict mode, unknown fields may just be warnings
	// Create strict validator to show the difference
	registry := validator.NewRegistry(validator.FHIRVersionR4)
	specsDir := filepath.Join("..", "..", "specs", "r4")
	registry.LoadR4Specs(specsDir)

	strictOpts := validator.ValidatorOptions{
		ValidateConstraints: false,
		StrictMode:          true,
	}
	strictValidator := validator.NewValidator(registry, strictOpts)
	strictResult, _ := strictValidator.Validate(ctx, patient)

	fmt.Println("\n  With StrictMode enabled:")
	printResult("Unknown Field (Strict)", strictResult)
}

func validateInvalidPrimitiveTypes(ctx context.Context, v *validator.Validator) {
	// Patient with wrong primitive types
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "invalid-primitives",
		"active": "yes",
		"birthDate": "not-a-valid-date",
		"multipleBirthInteger": "three"
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Invalid Primitives", result)
	fmt.Println("\n  Expected errors:")
	fmt.Println("  - 'active' should be boolean, not string")
	fmt.Println("  - 'birthDate' should match date pattern")
	fmt.Println("  - 'multipleBirthInteger' should be integer, not string")
}

func validateCardinality(ctx context.Context, v *validator.Validator) {
	// Observation requires 'status' and 'code' (min=1)
	// Missing required fields
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "missing-required"
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Missing Required Fields", result)
	fmt.Println("\n  Observation requires:")
	fmt.Println("  - status (required, min=1)")
	fmt.Println("  - code (required, min=1)")

	// Valid Observation with required fields
	fmt.Println("\n  Valid Observation with required fields:")
	validObs := []byte(`{
		"resourceType": "Observation",
		"id": "valid-obs",
		"status": "final",
		"code": {
			"coding": [{
				"system": "http://loinc.org",
				"code": "12345-6"
			}]
		}
	}`)

	validResult, _ := v.Validate(ctx, validObs)
	printResult("Valid Observation", validResult)
}

func validateComplexStructure(ctx context.Context, v *validator.Validator) {
	// Complex Observation with components
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "blood-pressure",
		"status": "final",
		"category": [{
			"coding": [{
				"system": "http://terminology.hl7.org/CodeSystem/observation-category",
				"code": "vital-signs",
				"display": "Vital Signs"
			}]
		}],
		"code": {
			"coding": [{
				"system": "http://loinc.org",
				"code": "85354-9",
				"display": "Blood pressure panel"
			}]
		},
		"subject": {
			"reference": "Patient/example"
		},
		"effectiveDateTime": "2024-01-15T10:30:00Z",
		"component": [
			{
				"code": {
					"coding": [{
						"system": "http://loinc.org",
						"code": "8480-6",
						"display": "Systolic blood pressure"
					}]
				},
				"valueQuantity": {
					"value": 120,
					"unit": "mmHg",
					"system": "http://unitsofmeasure.org",
					"code": "mm[Hg]"
				}
			},
			{
				"code": {
					"coding": [{
						"system": "http://loinc.org",
						"code": "8462-4",
						"display": "Diastolic blood pressure"
					}]
				},
				"valueQuantity": {
					"value": 80,
					"unit": "mmHg",
					"system": "http://unitsofmeasure.org",
					"code": "mm[Hg]"
				}
			}
		]
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Complex Observation", result)
	fmt.Println("  -> Validates nested structures: category, code, component, valueQuantity")
}

func validateStrictMode(ctx context.Context) {
	registry := validator.NewRegistry(validator.FHIRVersionR4)
	specsDir := filepath.Join("..", "..", "specs", "r4")
	registry.LoadR4Specs(specsDir)

	patient := []byte(`{
		"resourceType": "Patient",
		"id": "strict-test",
		"active": true,
		"unknownElement": "test",
		"name": [{"family": "Test"}]
	}`)

	// Non-strict mode
	nonStrictOpts := validator.ValidatorOptions{
		ValidateConstraints: false,
		StrictMode:          false,
	}
	nonStrictValidator := validator.NewValidator(registry, nonStrictOpts)
	nonStrictResult, _ := nonStrictValidator.Validate(ctx, patient)

	fmt.Println("  Non-Strict Mode:")
	printResult("Patient with unknown element", nonStrictResult)

	// Strict mode
	strictOpts := validator.ValidatorOptions{
		ValidateConstraints: false,
		StrictMode:          true,
	}
	strictValidator := validator.NewValidator(registry, strictOpts)
	strictResult, _ := strictValidator.Validate(ctx, patient)

	fmt.Println("\n  Strict Mode:")
	printResult("Patient with unknown element", strictResult)

	fmt.Println("\n  -> In strict mode, unknown elements are errors")
	fmt.Println("  -> In non-strict mode, unknown elements may be warnings or ignored")
}
