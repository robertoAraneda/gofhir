// constraint_validation.go demonstrates FHIRPath constraint validation.
// This includes validation of:
// - FHIRPath constraints defined in StructureDefinitions
// - Constraint violation detection
// - Constraint evaluation options
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

// RunConstraintValidationExamples demonstrates FHIRPath constraint validation features
func RunConstraintValidationExamples(ctx context.Context) {
	fmt.Println("\n" + separator)
	fmt.Println("CONSTRAINT VALIDATION EXAMPLES")
	fmt.Println(separator)

	v, err := createConstraintValidator()
	if err != nil {
		log.Printf("Failed to create validator: %v", err)
		return
	}

	// Example 1: Valid Patient satisfying all constraints
	fmt.Println("\n--- 1. Valid Patient (satisfies all constraints) ---")
	validateValidPatientConstraints(ctx, v)

	// Example 2: Patient violating pat-1 constraint
	fmt.Println("\n--- 2. Constraint Violation (pat-1) ---")
	validatePat1Violation(ctx, v)

	// Example 3: ele-1 constraint (all elements must have value or children)
	fmt.Println("\n--- 3. Element Constraint (ele-1) ---")
	validateEle1Constraint(ctx, v)

	// Example 4: Observation constraint validation
	fmt.Println("\n--- 4. Observation Constraint Validation ---")
	validateObservationConstraints(ctx, v)

	// Example 5: Constraint validation disabled vs enabled
	fmt.Println("\n--- 5. Constraint Validation Toggle ---")
	validateConstraintToggle(ctx)
}

func createConstraintValidator() (*validator.Validator, error) {
	registry := validator.NewRegistry(validator.FHIRVersionR4)
	specsDir := filepath.Join("..", "..", "specs", "r4")
	_, err := registry.LoadR4Specs(specsDir)
	if err != nil {
		return nil, err
	}

	opts := validator.ValidatorOptions{
		ValidateConstraints: true, // Enable constraint validation
		ValidateExtensions:  false,
		ValidateReferences:  false,
		StrictMode:          false,
	}
	return validator.NewValidator(registry, opts), nil
}

func validateValidPatientConstraints(ctx context.Context, v *validator.Validator) {
	// Patient with valid contact (has name, satisfies pat-1)
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "valid-constraints",
		"active": true,
		"name": [{"family": "Doe", "given": ["John"]}],
		"contact": [{
			"relationship": [{
				"coding": [{
					"system": "http://terminology.hl7.org/CodeSystem/v2-0131",
					"code": "E"
				}]
			}],
			"name": {"family": "Doe", "given": ["Jane"]}
		}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Valid Patient Constraints", result)
	fmt.Println("  -> Patient.contact has name, satisfying pat-1 constraint")
}

func validateEle1Constraint(ctx context.Context, v *validator.Validator) {
	// ele-1: All FHIR elements must have a @value or children
	// Expression: hasValue() or (children().count() > id.count())
	// This constraint ensures that elements are not empty

	fmt.Println("  ele-1 constraint:")
	fmt.Println("  'All FHIR elements must have a @value or children'")
	fmt.Println("  Expression: hasValue() or (children().count() > id.count())")
	fmt.Println("\n  NOTE: ele-1 is a universal constraint that applies to ALL FHIR elements.")
	fmt.Println("  Full validation of ele-1 on empty objects {} is not yet implemented.")
	fmt.Println("  Currently, specific validators (like extension validation) catch these cases.")

	// Valid case: Element with value
	fmt.Println("\n  1. Valid: Element with primitive value")
	validPatient := []byte(`{
		"resourceType": "Patient",
		"id": "valid-ele1",
		"active": true,
		"gender": "male",
		"birthDate": "1990-05-15"
	}`)

	validResult, err := v.Validate(ctx, validPatient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}
	printResult("Patient with primitive values", validResult)

	// Valid case: Element with children (complex type)
	fmt.Println("\n  2. Valid: Element with children (complex type)")
	complexPatient := []byte(`{
		"resourceType": "Patient",
		"id": "valid-complex",
		"name": [{
			"family": "Smith",
			"given": ["Jane", "Marie"]
		}],
		"address": [{
			"city": "Springfield",
			"state": "IL"
		}]
	}`)

	complexResult, err := v.Validate(ctx, complexPatient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}
	printResult("Patient with complex types", complexResult)

	// Invalid case: Extension without value - detected by extension validator
	fmt.Println("\n  3. Invalid: Extension without value[x]")
	fmt.Println("     (Detected by extension validator implementing ele-1/ext-1 rules)")

	registry := validator.NewRegistry(validator.FHIRVersionR4)
	specsDir := filepath.Join("..", "..", "specs", "r4")
	registry.LoadR4Specs(specsDir)

	extValidatorOpts := validator.ValidatorOptions{
		ValidateConstraints: true,
		ValidateExtensions:  true,
	}
	extValidator := validator.NewValidator(registry, extValidatorOpts)

	noValueExtPatient := []byte(`{
		"resourceType": "Patient",
		"id": "invalid-ext",
		"extension": [{
			"url": "http://example.org/fhir/StructureDefinition/importance"
		}],
		"name": [{"family": "Test"}]
	}`)

	noValueExtResult, err := extValidator.Validate(ctx, noValueExtPatient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}
	printResult("Extension without value", noValueExtResult)

	// Show what ele-1 means conceptually
	fmt.Println("\n  ele-1 ensures:")
	fmt.Println("  - Primitive elements have a value (e.g., active: true, gender: 'male')")
	fmt.Println("  - Complex elements have child elements (e.g., name: {family: 'Doe'})")
	fmt.Println("  - Empty objects {} are NOT valid FHIR elements")
}

func validatePat1Violation(ctx context.Context, v *validator.Validator) {
	// Patient.contact requires name OR telecom OR address OR organization
	// This contact only has relationship, violating pat-1 constraint
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "constraint-violation",
		"active": true,
		"contact": [{
			"relationship": [{
				"coding": [{
					"system": "http://terminology.hl7.org/CodeSystem/v2-0131",
					"code": "E"
				}]
			}]
		}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("pat-1 Constraint Violation", result)
	fmt.Println("\n  pat-1 constraint:")
	fmt.Println("  'SHALL at least contain a contact's details or a reference to an organization'")
	fmt.Println("  Expression: name.exists() or telecom.exists() or address.exists() or organization.exists()")
}

func validateObservationConstraints(ctx context.Context, v *validator.Validator) {
	// Observation with components - obs-6 constraint applies
	// obs-6: dataAbsentReason SHALL only be present if Observation.value[x] is not present
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "obs-constraints",
		"status": "final",
		"code": {
			"coding": [{
				"system": "http://loinc.org",
				"code": "85354-9",
				"display": "Blood pressure panel"
			}]
		},
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
			}
		]
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Observation Constraints", result)
	fmt.Println("\n  Observation constraints evaluated:")
	fmt.Println("  - obs-3: Must have at least a value[x] or a data absent reason")
	fmt.Println("  - obs-6: dataAbsentReason SHALL only be present if value is not present")
	fmt.Println("  - obs-7: If component code and value code are same, should have same value")
}

func validateConstraintToggle(ctx context.Context) {
	registry := validator.NewRegistry(validator.FHIRVersionR4)
	specsDir := filepath.Join("..", "..", "specs", "r4")
	registry.LoadR4Specs(specsDir)

	// Patient with constraint violation (pat-1)
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "toggle-test",
		"contact": [{
			"relationship": [{"coding": [{"code": "E"}]}]
		}]
	}`)

	// With constraints enabled
	fmt.Println("  With ValidateConstraints = true:")
	enabledOpts := validator.ValidatorOptions{
		ValidateConstraints: true,
	}
	enabledValidator := validator.NewValidator(registry, enabledOpts)
	enabledResult, _ := enabledValidator.Validate(ctx, patient)
	printResult("Constraints Enabled", enabledResult)

	// With constraints disabled
	fmt.Println("\n  With ValidateConstraints = false:")
	disabledOpts := validator.ValidatorOptions{
		ValidateConstraints: false,
	}
	disabledValidator := validator.NewValidator(registry, disabledOpts)
	disabledResult, _ := disabledValidator.Validate(ctx, patient)
	printResult("Constraints Disabled", disabledResult)

	fmt.Println("\n  -> Disabling constraint validation skips FHIRPath evaluation")
	fmt.Println("  -> Structural validation still runs (cardinality, types, etc.)")
}
