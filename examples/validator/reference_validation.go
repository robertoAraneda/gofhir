// reference_validation.go demonstrates FHIR reference validation.
// This includes validation of:
// - Relative references (Patient/123)
// - Contained references (#contained-id)
// - Absolute URL references (https://example.org/fhir/Patient/123)
// - URN references (urn:uuid:...)
// - Logical references (identifier or display only)
package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/robertoaraneda/gofhir/pkg/validator"
)

// RunReferenceValidationExamples demonstrates reference validation features
func RunReferenceValidationExamples(ctx context.Context) {
	fmt.Println("\n" + separator)
	fmt.Println("REFERENCE VALIDATION EXAMPLES")
	fmt.Println(separator)

	v, err := createReferenceValidator()
	if err != nil {
		log.Printf("Failed to create validator: %v", err)
		return
	}

	// Example 1: Valid relative reference
	fmt.Println("\n--- 1. Valid Relative Reference ---")
	validateRelativeReference(ctx, v)

	// Example 2: Valid contained reference
	fmt.Println("\n--- 2. Valid Contained Reference ---")
	validateContainedReference(ctx, v)

	// Example 3: Invalid contained reference
	fmt.Println("\n--- 3. Invalid Contained Reference ---")
	validateInvalidContainedReference(ctx, v)

	// Example 4: Valid absolute URL reference
	fmt.Println("\n--- 4. Valid Absolute URL Reference ---")
	validateAbsoluteURLReference(ctx, v)

	// Example 5: Valid URN reference
	fmt.Println("\n--- 5. Valid URN Reference ---")
	validateURNReference(ctx, v)

	// Example 6: Logical reference with display
	fmt.Println("\n--- 6. Logical Reference (display only) ---")
	validateLogicalReferenceDisplay(ctx, v)

	// Example 7: Logical reference with identifier
	fmt.Println("\n--- 7. Logical Reference (identifier) ---")
	validateLogicalReferenceIdentifier(ctx, v)
}

func createReferenceValidator() (*validator.Validator, error) {
	registry := validator.NewRegistry(validator.FHIRVersionR4)
	specsDir := filepath.Join("..", "..", "specs", "r4")
	_, err := registry.LoadR4Specs(specsDir)
	if err != nil {
		return nil, err
	}

	opts := validator.ValidatorOptions{
		ValidateReferences:  true, // Enable reference validation
		ValidateExtensions:  false,
		ValidateConstraints: false,
		StrictMode:          false,
	}
	return validator.NewValidator(registry, opts), nil
}

func validateRelativeReference(ctx context.Context, v *validator.Validator) {
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "obs-relative",
		"status": "final",
		"code": {"coding": [{"code": "test"}]},
		"subject": {
			"reference": "Patient/123"
		}
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Relative Reference", result)
	fmt.Println("  -> Format: ResourceType/id (e.g., Patient/123)")
}

func validateContainedReference(ctx context.Context, v *validator.Validator) {
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "obs-contained",
		"status": "final",
		"code": {"coding": [{"code": "test"}]},
		"contained": [{
			"resourceType": "Patient",
			"id": "pat1",
			"name": [{"family": "Smith"}]
		}],
		"subject": {
			"reference": "#pat1"
		}
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Contained Reference", result)
	fmt.Println("  -> Format: #contained-id")
	fmt.Println("  -> Referenced resource must exist in 'contained' array")
}

func validateInvalidContainedReference(ctx context.Context, v *validator.Validator) {
	// References a contained resource that doesn't exist
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "obs-invalid-contained",
		"status": "final",
		"code": {"coding": [{"code": "test"}]},
		"subject": {
			"reference": "#nonexistent"
		}
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Invalid Contained Reference", result)
	for _, issue := range result.Issues {
		if issue.Severity == validator.SeverityError {
			fmt.Printf("  Error: %s\n", issue.Diagnostics)
		}
	}
}

func validateAbsoluteURLReference(ctx context.Context, v *validator.Validator) {
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "obs-absolute",
		"status": "final",
		"code": {"coding": [{"code": "test"}]},
		"subject": {
			"reference": "https://example.org/fhir/Patient/456"
		}
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Absolute URL Reference", result)
	fmt.Println("  -> Format: https://server/fhir/ResourceType/id")
	fmt.Println("  -> Validator checks format but doesn't resolve external URLs")
}

func validateURNReference(ctx context.Context, v *validator.Validator) {
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "obs-urn",
		"status": "final",
		"code": {"coding": [{"code": "test"}]},
		"subject": {
			"reference": "urn:uuid:550e8400-e29b-41d4-a716-446655440000"
		}
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("URN Reference", result)
	fmt.Println("  -> Formats: urn:uuid:... or urn:oid:...")
	fmt.Println("  -> Common in Bundles for temporary references")
}

func validateLogicalReferenceDisplay(ctx context.Context, v *validator.Validator) {
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "obs-logical-display",
		"status": "final",
		"code": {"coding": [{"code": "test"}]},
		"subject": {
			"display": "Patient John Doe"
		}
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Logical Reference (display)", result)
	fmt.Println("  -> Only contains display text, no actual reference")
	fmt.Println("  -> Valid but cannot be resolved")
}

func validateLogicalReferenceIdentifier(ctx context.Context, v *validator.Validator) {
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "obs-logical-identifier",
		"status": "final",
		"code": {"coding": [{"code": "test"}]},
		"subject": {
			"type": "Patient",
			"identifier": {
				"system": "http://hospital.example.org/patients",
				"value": "12345"
			}
		}
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Logical Reference (identifier)", result)
	fmt.Println("  -> Contains type and identifier instead of reference URL")
	fmt.Println("  -> Can be resolved using search: GET /Patient?identifier=...")
}
