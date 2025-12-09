// extension_validation.go demonstrates FHIR extension validation.
// This includes validation of:
// - Simple extensions with value[x]
// - Complex extensions with nested extensions
// - Modifier extensions
// - HL7 standard extensions
// - Extension structure rules
package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/robertoaraneda/gofhir/pkg/validator"
)

// RunExtensionValidationExamples demonstrates extension validation features
func RunExtensionValidationExamples(ctx context.Context) {
	fmt.Println("\n" + separator)
	fmt.Println("EXTENSION VALIDATION EXAMPLES")
	fmt.Println(separator)

	v, err := createExtensionValidator()
	if err != nil {
		log.Printf("Failed to create validator: %v", err)
		return
	}

	// Example 1: Valid simple extension
	fmt.Println("\n--- 1. Valid Simple Extension ---")
	validateSimpleExtension(ctx, v)

	// Example 2: Valid complex extension
	fmt.Println("\n--- 2. Valid Complex Extension (nested) ---")
	validateComplexExtension(ctx, v)

	// Example 3: Invalid extension - missing URL
	fmt.Println("\n--- 3. Invalid Extension (missing URL) ---")
	validateMissingURLExtension(ctx, v)

	// Example 4: Invalid extension - missing value
	fmt.Println("\n--- 4. Invalid Extension (missing value) ---")
	validateMissingValueExtension(ctx, v)

	// Example 5: Invalid extension - both value and nested
	fmt.Println("\n--- 5. Invalid Extension (both value and nested) ---")
	validateBothValueAndNestedExtension(ctx, v)

	// Example 6: Valid modifier extension
	fmt.Println("\n--- 6. Valid Modifier Extension ---")
	validateModifierExtension(ctx, v)

	// Example 7: Extension on nested element
	fmt.Println("\n--- 7. Extension on Nested Element ---")
	validateNestedElementExtension(ctx, v)

	// Example 8: Multiple extensions
	fmt.Println("\n--- 8. Multiple Extensions ---")
	validateMultipleExtensions(ctx, v)

	// Example 9: HL7 standard extension
	fmt.Println("\n--- 9. HL7 Standard Extension ---")
	validateHL7StandardExtension(ctx, v)

	// Example 10: HL7 extension with wrong type
	fmt.Println("\n--- 10. HL7 Extension with Wrong Type ---")
	validateHL7ExtensionWrongType(ctx, v)
}

func createExtensionValidator() (*validator.Validator, error) {
	registry := validator.NewRegistry(validator.FHIRVersionR4)
	specsDir := filepath.Join("..", "..", "specs", "r4")
	// Load all specs including extension definitions
	_, err := registry.LoadR4Specs(specsDir)
	if err != nil {
		return nil, err
	}

	opts := validator.ValidatorOptions{
		ValidateExtensions:  true, // Enable extension validation
		ValidateReferences:  false,
		ValidateConstraints: false,
		StrictMode:          true, // Enable strict mode to see warnings for unknown extensions
	}
	return validator.NewValidator(registry, opts), nil
}

func validateSimpleExtension(ctx context.Context, v *validator.Validator) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "pat-simple-ext",
		"extension": [{
			"url": "http://example.org/fhir/StructureDefinition/patient-importance",
			"valueCode": "VIP"
		}],
		"name": [{"family": "Smith"}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Simple Extension", result)
	fmt.Println("  -> Extension has URL and value[x] (valueCode)")
}

func validateComplexExtension(ctx context.Context, v *validator.Validator) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "pat-complex-ext",
		"extension": [{
			"url": "http://example.org/fhir/StructureDefinition/patient-geolocation",
			"extension": [
				{
					"url": "latitude",
					"valueDecimal": 40.7128
				},
				{
					"url": "longitude",
					"valueDecimal": -74.0060
				}
			]
		}],
		"name": [{"family": "Johnson"}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Complex Extension", result)
	fmt.Println("  -> Extension has URL and nested extensions (no direct value)")
}

func validateMissingURLExtension(ctx context.Context, v *validator.Validator) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "pat-missing-url",
		"extension": [{
			"valueString": "some value without URL"
		}],
		"name": [{"family": "Williams"}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Missing URL Extension", result)
	for _, issue := range result.Issues {
		if issue.Severity == validator.SeverityError {
			fmt.Printf("  Error: %s\n", issue.Diagnostics)
		}
	}
	fmt.Println("\n  -> Extension MUST have a 'url' field")
}

func validateMissingValueExtension(ctx context.Context, v *validator.Validator) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "pat-missing-value",
		"extension": [{
			"url": "http://example.org/fhir/StructureDefinition/empty-extension"
		}],
		"name": [{"family": "Brown"}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Missing Value Extension", result)
	for _, issue := range result.Issues {
		if issue.Severity == validator.SeverityError {
			fmt.Printf("  Error: %s\n", issue.Diagnostics)
		}
	}
	fmt.Println("\n  -> Extension MUST have either value[x] OR nested extensions")
}

func validateBothValueAndNestedExtension(ctx context.Context, v *validator.Validator) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "pat-both-value-nested",
		"extension": [{
			"url": "http://example.org/fhir/StructureDefinition/invalid-structure",
			"valueString": "a value",
			"extension": [{
				"url": "nested",
				"valueCode": "should not have both"
			}]
		}],
		"name": [{"family": "Davis"}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Both Value and Nested Extension", result)
	for _, issue := range result.Issues {
		if issue.Severity == validator.SeverityError {
			fmt.Printf("  Error: %s\n", issue.Diagnostics)
		}
	}
	fmt.Println("\n  -> Extension MUST NOT have both value[x] AND nested extensions")
}

func validateModifierExtension(ctx context.Context, v *validator.Validator) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "pat-modifier-ext",
		"modifierExtension": [{
			"url": "http://example.org/fhir/StructureDefinition/patient-confidential",
			"valueBoolean": true
		}],
		"name": [{"family": "Miller"}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Modifier Extension", result)
	fmt.Println("  -> modifierExtension changes resource meaning")
	fmt.Println("  -> Consumers MUST understand it before processing")
}

func validateNestedElementExtension(ctx context.Context, v *validator.Validator) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "pat-nested-ext",
		"name": [{
			"family": "Wilson",
			"extension": [{
				"url": "http://example.org/fhir/StructureDefinition/name-pronunciation",
				"valueString": "WIL-son"
			}]
		}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Nested Element Extension", result)
	fmt.Println("  -> Extensions can appear on any FHIR element")
	fmt.Println("  -> This extension is on Patient.name (HumanName)")
}

func validateMultipleExtensions(ctx context.Context, v *validator.Validator) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "pat-multiple-ext",
		"extension": [
			{
				"url": "http://example.org/fhir/StructureDefinition/patient-importance",
				"valueCode": "VIP"
			},
			{
				"url": "http://example.org/fhir/StructureDefinition/patient-nationality",
				"valueCodeableConcept": {
					"coding": [{
						"system": "urn:iso:std:iso:3166",
						"code": "US",
						"display": "United States"
					}]
				}
			}
		],
		"name": [{"family": "Anderson"}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Multiple Extensions", result)
	fmt.Println("  -> Resources can have multiple extensions")
	fmt.Println("  -> Each extension is validated independently")
}

func validateHL7StandardExtension(ctx context.Context, v *validator.Validator) {
	// Using patient-birthPlace which is a standard HL7 extension
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "pat-hl7-ext",
		"extension": [{
			"url": "http://hl7.org/fhir/StructureDefinition/patient-birthPlace",
			"valueAddress": {
				"city": "New York",
				"state": "NY",
				"country": "USA"
			}
		}],
		"name": [{"family": "Thompson"}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("HL7 Standard Extension", result)
	// Show any issues (should be minimal for valid extension)
	for _, issue := range result.Issues {
		fmt.Printf("  [%s] %s: %s\n", issue.Severity, issue.Code, issue.Diagnostics)
	}
	fmt.Println("\n  -> HL7 extensions are validated against their StructureDefinition")
	fmt.Println("  -> patient-birthPlace expects valueAddress")
}

func validateHL7ExtensionWrongType(ctx context.Context, v *validator.Validator) {
	// Using wrong value type for patient-birthPlace (should be Address, not String)
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "pat-hl7-wrong-type",
		"extension": [{
			"url": "http://hl7.org/fhir/StructureDefinition/patient-birthPlace",
			"valueString": "New York"
		}],
		"name": [{"family": "Garcia"}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("HL7 Extension Wrong Type", result)
	for _, issue := range result.Issues {
		if issue.Severity == validator.SeverityError {
			fmt.Printf("  Error: %s\n", issue.Diagnostics)
		}
	}
	fmt.Println("\n  -> patient-birthPlace requires valueAddress, not valueString")
}
