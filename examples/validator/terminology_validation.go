// terminology_validation.go demonstrates terminology binding validation.
// This includes validation of:
// - Code elements against ValueSets (required/extensible bindings)
// - Coding elements with system validation
// - CodeableConcept elements with multiple codings
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

// RunTerminologyValidationExamples demonstrates terminology binding validation features
func RunTerminologyValidationExamples(ctx context.Context) {
	fmt.Println("\n" + separator)
	fmt.Println("TERMINOLOGY VALIDATION EXAMPLES")
	fmt.Println(separator)

	v, termService, err := createTerminologyValidator()
	if err != nil {
		log.Printf("Failed to create validator: %v", err)
		return
	}

	// Show terminology service stats
	codeSystems, valueSets, totalCodes := termService.Stats()
	fmt.Printf("\nTerminology service loaded: %d CodeSystems, %d ValueSets, %d total codes\n",
		codeSystems, valueSets, totalCodes)

	// Example 1: Valid Patient with correct gender code
	fmt.Println("\n--- 1. Valid Patient (correct gender code) ---")
	validateValidGender(ctx, v)

	// Example 2: Invalid Patient with wrong gender code
	fmt.Println("\n--- 2. Invalid Patient (invalid gender code) ---")
	validateInvalidGender(ctx, v)

	// Example 3: Valid Observation with correct status
	fmt.Println("\n--- 3. Valid Observation (correct status code) ---")
	validateValidObservationStatus(ctx, v)

	// Example 4: Invalid Observation with wrong status
	fmt.Println("\n--- 4. Invalid Observation (invalid status code) ---")
	validateInvalidObservationStatus(ctx, v)

	// Example 5: CodeableConcept validation
	fmt.Println("\n--- 5. CodeableConcept Validation ---")
	validateCodeableConcept(ctx, v)

	// Example 6: Terminology validation disabled vs enabled
	fmt.Println("\n--- 6. Terminology Validation Toggle ---")
	validateTerminologyToggle(ctx)
}

func createTerminologyValidator() (*validator.Validator, *validator.LocalTerminologyService, error) {
	// Load StructureDefinitions
	registry := validator.NewRegistry(validator.FHIRVersionR4)
	specsDir := filepath.Join("..", "..", "specs", "r4")
	_, err := registry.LoadR4Specs(specsDir)
	if err != nil {
		return nil, nil, err
	}

	// Load terminology (ValueSets and CodeSystems)
	termService := validator.NewLocalTerminologyService()
	valueSetsPath := filepath.Join(specsDir, "valuesets.json")
	if err := termService.LoadFromFile(valueSetsPath); err != nil {
		return nil, nil, fmt.Errorf("failed to load valuesets: %w", err)
	}

	opts := validator.ValidatorOptions{
		ValidateConstraints: false, // Focus on terminology
		ValidateTerminology: true,  // Enable terminology validation
		ValidateExtensions:  false,
		ValidateReferences:  false,
		StrictMode:          false,
	}

	v := validator.NewValidator(registry, opts).WithTerminologyService(termService)
	return v, termService, nil
}

func validateValidGender(ctx context.Context, v *validator.Validator) {
	// Patient.gender has required binding to administrative-gender ValueSet
	// Valid codes: male, female, other, unknown
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "valid-gender",
		"gender": "male",
		"name": [{"family": "Doe", "given": ["John"]}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Patient with valid gender", result)
	fmt.Println("  -> 'male' is valid in http://hl7.org/fhir/ValueSet/administrative-gender")
}

func validateInvalidGender(ctx context.Context, v *validator.Validator) {
	// Invalid gender code - not in administrative-gender ValueSet
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "invalid-gender",
		"gender": "not-a-valid-gender",
		"name": [{"family": "Doe", "given": ["Jane"]}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Patient with invalid gender", result)
	fmt.Println("\n  Patient.gender binding:")
	fmt.Println("    Strength: required")
	fmt.Println("    ValueSet: http://hl7.org/fhir/ValueSet/administrative-gender")
	fmt.Println("    Valid codes: male, female, other, unknown")
}

func validateValidObservationStatus(ctx context.Context, v *validator.Validator) {
	// Observation.status has required binding to observation-status ValueSet
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "valid-status",
		"status": "final",
		"code": {
			"coding": [{
				"system": "http://loinc.org",
				"code": "15074-8",
				"display": "Glucose [Moles/volume] in Blood"
			}]
		}
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Observation with valid status", result)
	fmt.Println("  -> 'final' is valid in http://hl7.org/fhir/ValueSet/observation-status")
}

func validateInvalidObservationStatus(ctx context.Context, v *validator.Validator) {
	// Invalid status code
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "invalid-status",
		"status": "completed",
		"code": {
			"coding": [{
				"system": "http://loinc.org",
				"code": "15074-8"
			}]
		}
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Observation with invalid status", result)
	fmt.Println("\n  Observation.status binding:")
	fmt.Println("    Strength: required")
	fmt.Println("    ValueSet: http://hl7.org/fhir/ValueSet/observation-status")
	fmt.Println("    Valid codes: registered, preliminary, final, amended, corrected,")
	//nolint:misspell // FHIR R4 uses British spelling "cancelled"
	fmt.Println("                 cancelled, entered-in-error, unknown")
	fmt.Println("    Note: 'completed' is NOT a valid ObservationStatus code")
}

func validateCodeableConcept(ctx context.Context, v *validator.Validator) {
	fmt.Println("  CodeableConcept elements can have multiple codings.")
	fmt.Println("  Each coding is validated against the bound ValueSet.")
	fmt.Println()

	// Valid CodeableConcept with valid coding
	validObs := []byte(`{
		"resourceType": "Observation",
		"id": "valid-concept",
		"status": "final",
		"code": {
			"coding": [{
				"system": "http://loinc.org",
				"code": "85354-9",
				"display": "Blood pressure panel"
			}],
			"text": "Blood Pressure"
		}
	}`)

	result, err := v.Validate(ctx, validObs)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Observation with CodeableConcept", result)
	fmt.Println("  -> CodeableConcept.coding validated against Observation.code binding")
}

func validateTerminologyToggle(ctx context.Context) {
	registry := validator.NewRegistry(validator.FHIRVersionR4)
	specsDir := filepath.Join("..", "..", "specs", "r4")
	registry.LoadR4Specs(specsDir)

	// Load terminology service
	termService := validator.NewLocalTerminologyService()
	valueSetsPath := filepath.Join(specsDir, "valuesets.json")
	termService.LoadFromFile(valueSetsPath)

	// Patient with invalid gender (to show the difference)
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "toggle-test",
		"gender": "invalid-code"
	}`)

	// With terminology enabled
	fmt.Println("  With ValidateTerminology = true:")
	enabledOpts := validator.ValidatorOptions{
		ValidateTerminology: true,
	}
	enabledValidator := validator.NewValidator(registry, enabledOpts).WithTerminologyService(termService)
	enabledResult, _ := enabledValidator.Validate(ctx, patient)
	printResult("Terminology Enabled", enabledResult)

	// With terminology disabled
	fmt.Println("\n  With ValidateTerminology = false:")
	disabledOpts := validator.ValidatorOptions{
		ValidateTerminology: false,
	}
	disabledValidator := validator.NewValidator(registry, disabledOpts)
	disabledResult, _ := disabledValidator.Validate(ctx, patient)
	printResult("Terminology Disabled", disabledResult)

	fmt.Println("\n  -> Disabling terminology validation skips ValueSet code checking")
	fmt.Println("  -> Structural validation still runs (the gender field is still valid JSON)")

	// Example 7: Using embedded terminology (no file I/O)
	fmt.Println("\n--- 7. Embedded Terminology Service ---")
	validateWithEmbeddedTerminology(ctx, registry)
}

func validateWithEmbeddedTerminology(ctx context.Context, registry *validator.Registry) {
	fmt.Println("  Using pre-compiled ValueSets (no file I/O required):")

	patient := []byte(`{
		"resourceType": "Patient",
		"id": "embedded-test",
		"gender": "male"
	}`)

	// Method 1: Simplest - just set ValidateTerminology=true (uses R4 by default)
	fmt.Println("\n  Method 1: Auto-configured (defaults to R4)")
	opts1 := validator.ValidatorOptions{
		ValidateTerminology: true, // Automatically uses embedded R4
	}
	v1 := validator.NewValidator(registry, opts1)
	result1, _ := v1.Validate(ctx, patient)
	printResult("Auto R4", result1)

	// Method 2: Explicit version via options
	fmt.Println("\n  Method 2: Explicit version via ValidatorOptions")
	opts2 := validator.ValidatorOptions{
		ValidateTerminology: true,
		TerminologyService:  validator.TerminologyEmbeddedR4B, // Use R4B
	}
	v2 := validator.NewValidator(registry, opts2)
	result2, _ := v2.Validate(ctx, patient)
	printResult("Explicit R4B", result2)

	// Method 3: Manual service injection (for custom services)
	fmt.Println("\n  Method 3: Manual service injection")
	embeddedService := validator.NewEmbeddedTerminologyServiceR5()
	valueSets, totalCodes := embeddedService.Stats()
	fmt.Printf("  R5 service: %d ValueSets, %d codes\n", valueSets, totalCodes)

	opts3 := validator.ValidatorOptions{
		ValidateTerminology: true,
	}
	v3 := validator.NewValidator(registry, opts3).WithTerminologyService(embeddedService)
	result3, _ := v3.Validate(ctx, patient)
	printResult("Manual R5", result3)

	// Show available options
	fmt.Println("\n  Available TerminologyService options:")
	fmt.Println("    - TerminologyNone (default when ValidateTerminology=false)")
	fmt.Println("    - TerminologyEmbeddedR4 (default when ValidateTerminology=true)")
	fmt.Println("    - TerminologyEmbeddedR4B")
	fmt.Println("    - TerminologyEmbeddedR5")
}
