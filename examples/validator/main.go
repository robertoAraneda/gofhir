// Package main demonstrates FHIR resource validation using gofhir.
// This example shows how to validate resources against StructureDefinitions,
// including structural validation, constraint validation, and custom profiles.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/robertoaraneda/gofhir/pkg/validator"
)

func main() {
	fmt.Println("=== FHIR Validator Examples ===")

	// Setup validator with R4 specs
	v, err := setupValidator()
	if err != nil {
		log.Fatalf("Failed to setup validator: %v", err)
	}

	ctx := context.Background()

	// 1. Validate a simple valid Patient
	fmt.Println("\n--- 1. Valid Patient ---")
	validateValidPatient(ctx, v)

	// 2. Validate Patient with missing required fields
	fmt.Println("\n--- 2. Patient with Invalid Structure ---")
	validateInvalidStructure(ctx, v)

	// 3. Validate Patient with invalid primitive types
	fmt.Println("\n--- 3. Patient with Invalid Primitive Types ---")
	validateInvalidPrimitives(ctx, v)

	// 4. Validate Patient that violates constraints
	fmt.Println("\n--- 4. Patient Violating Constraints ---")
	validateConstraintViolation(ctx, v)

	// 5. Validate complex Observation
	fmt.Println("\n--- 5. Valid Observation ---")
	validateObservation(ctx, v)

	// 6. Validate with different options
	fmt.Println("\n--- 6. Validation with Different Options ---")
	validateWithOptions(ctx, v)

	// 7. Batch validation
	fmt.Println("\n--- 7. Batch Validation ---")
	batchValidation(ctx, v)

	// 8. Show validation result analysis
	fmt.Println("\n--- 8. Analyzing Validation Results ---")
	analyzeValidationResults(ctx, v)
}

// setupValidator creates a validator with R4 StructureDefinitions
func setupValidator() (*validator.Validator, error) {
	// Create registry for R4
	registry := validator.NewRegistry(validator.FHIRVersionR4)

	// Try to load from specs directory (adjust path as needed)
	specsDir := filepath.Join("..", "..", "specs", "r4")

	// Load resource definitions
	resourcesPath := filepath.Join(specsDir, "profiles-resources.json")
	if _, err := os.Stat(resourcesPath); err == nil {
		count, err := registry.LoadFromFile(resourcesPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load resources: %w", err)
		}
		fmt.Printf("Loaded %d resource definitions\n", count)
	} else {
		fmt.Println("Warning: specs/r4/profiles-resources.json not found")
		fmt.Println("Download FHIR R4 specs from https://hl7.org/fhir/R4/downloads.html")
		return nil, fmt.Errorf("specs not found")
	}

	// Load type definitions
	typesPath := filepath.Join(specsDir, "profiles-types.json")
	if _, err := os.Stat(typesPath); err == nil {
		count, err := registry.LoadFromFile(typesPath)
		if err != nil {
			fmt.Printf("Warning: failed to load types: %v\n", err)
		} else {
			fmt.Printf("Loaded %d type definitions\n", count)
		}
	}

	// Create validator with default options
	opts := validator.DefaultValidatorOptions()
	return validator.NewValidator(registry, opts), nil
}

// validateValidPatient shows validation of a complete, valid Patient
func validateValidPatient(ctx context.Context, v *validator.Validator) {
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "valid-example",
		"active": true,
		"name": [{
			"use": "official",
			"family": "Doe",
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
		}],
		"contact": [{
			"relationship": [{"coding": [{"code": "E"}]}],
			"name": {"family": "Doe", "given": ["Jane"]}
		}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printValidationResult("Valid Patient", result)
}

// validateInvalidStructure shows structural validation errors
func validateInvalidStructure(ctx context.Context, v *validator.Validator) {
	// Patient with invalid cardinality (name is array, not single value)
	// and unknown field
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "invalid-structure",
		"unknownField": "this should not be here",
		"active": true
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printValidationResult("Invalid Structure", result)
}

// validateInvalidPrimitives shows primitive type validation
func validateInvalidPrimitives(ctx context.Context, v *validator.Validator) {
	// Patient with wrong types
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "invalid-primitives",
		"active": "yes",
		"birthDate": "not-a-date"
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printValidationResult("Invalid Primitives", result)
}

// validateConstraintViolation shows FHIRPath constraint validation
func validateConstraintViolation(ctx context.Context, v *validator.Validator) {
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

	printValidationResult("Constraint Violation", result)
}

// validateObservation shows validation of a complex resource
func validateObservation(ctx context.Context, v *validator.Validator) {
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

	printValidationResult("Observation", result)
}

// validateWithOptions shows different validation configurations
func validateWithOptions(ctx context.Context, v *validator.Validator) {
	// Create new registry and validator with custom options
	registry := validator.NewRegistry(validator.FHIRVersionR4)
	specsPath := filepath.Join("..", "..", "specs", "r4", "profiles-resources.json")
	registry.LoadFromFile(specsPath)

	patient := []byte(`{
		"resourceType": "Patient",
		"id": "options-test",
		"active": true,
		"contact": [{
			"relationship": [{"coding": [{"code": "E"}]}]
		}]
	}`)

	// Option 1: Strict mode (unknown elements are errors)
	fmt.Println("\n  With StrictMode enabled:")
	strictOpts := validator.ValidatorOptions{
		ValidateConstraints: true,
		StrictMode:          true,
		MaxErrors:           10,
	}
	strictValidator := validator.NewValidator(registry, strictOpts)
	strictResult, _ := strictValidator.Validate(ctx, patient)
	printValidationSummary(strictResult)

	// Option 2: Without constraint validation
	fmt.Println("\n  Without constraint validation:")
	noConstraintOpts := validator.ValidatorOptions{
		ValidateConstraints: false,
		StrictMode:          false,
	}
	noConstraintValidator := validator.NewValidator(registry, noConstraintOpts)
	noConstraintResult, _ := noConstraintValidator.Validate(ctx, patient)
	printValidationSummary(noConstraintResult)

	// Option 3: Max errors limit
	fmt.Println("\n  With MaxErrors = 1:")
	maxErrorOpts := validator.ValidatorOptions{
		ValidateConstraints: true,
		MaxErrors:           1,
	}
	maxErrorValidator := validator.NewValidator(registry, maxErrorOpts)
	maxErrorResult, _ := maxErrorValidator.Validate(ctx, patient)
	printValidationSummary(maxErrorResult)
}

// batchValidation shows validating multiple resources
func batchValidation(ctx context.Context, v *validator.Validator) {
	resources := []struct {
		name     string
		resource []byte
	}{
		{
			name: "Patient 1 (valid)",
			resource: []byte(`{
				"resourceType": "Patient",
				"id": "p1",
				"active": true,
				"name": [{"family": "Smith"}]
			}`),
		},
		{
			name: "Patient 2 (valid)",
			resource: []byte(`{
				"resourceType": "Patient",
				"id": "p2",
				"active": false,
				"name": [{"family": "Johnson"}],
				"gender": "female"
			}`),
		},
		{
			name: "Patient 3 (invalid contact)",
			resource: []byte(`{
				"resourceType": "Patient",
				"id": "p3",
				"contact": [{"relationship": [{}]}]
			}`),
		},
	}

	validCount := 0
	invalidCount := 0

	for _, r := range resources {
		result, err := v.Validate(ctx, r.resource)
		if err != nil {
			fmt.Printf("  %s: ERROR - %v\n", r.name, err)
			continue
		}

		status := "VALID"
		if !result.Valid {
			status = "INVALID"
			invalidCount++
		} else {
			validCount++
		}
		fmt.Printf("  %s: %s (errors: %d, warnings: %d)\n",
			r.name, status, result.ErrorCount(), result.WarningCount())
	}

	fmt.Printf("\nBatch Summary: %d valid, %d invalid\n", validCount, invalidCount)
}

// analyzeValidationResults shows how to work with validation results
func analyzeValidationResults(ctx context.Context, v *validator.Validator) {
	// Resource with multiple issues
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "analysis-example",
		"active": "invalid-boolean",
		"unknownField": true,
		"contact": [{
			"relationship": [{"coding": [{"code": "E"}]}]
		}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	fmt.Printf("Resource valid: %v\n", result.Valid)
	fmt.Printf("Total issues: %d\n", len(result.Issues))
	fmt.Printf("Errors: %d\n", result.ErrorCount())
	fmt.Printf("Warnings: %d\n", result.WarningCount())

	// Group issues by severity
	fmt.Println("\nIssues by Severity:")
	severityGroups := make(map[string][]validator.ValidationIssue)
	for _, issue := range result.Issues {
		severityGroups[issue.Severity] = append(severityGroups[issue.Severity], issue)
	}

	for severity, issues := range severityGroups {
		fmt.Printf("\n  %s (%d):\n", severity, len(issues))
		for _, issue := range issues {
			path := ""
			if len(issue.Expression) > 0 {
				path = issue.Expression[0]
			}
			fmt.Printf("    - [%s] %s: %s\n", issue.Code, path, issue.Diagnostics)
		}
	}

	// Group issues by code
	fmt.Println("\nIssues by Code:")
	codeGroups := make(map[string]int)
	for _, issue := range result.Issues {
		codeGroups[issue.Code]++
	}

	for code, count := range codeGroups {
		fmt.Printf("  %s: %d\n", code, count)
	}

	// Export as JSON
	fmt.Println("\nValidation Result as JSON:")
	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonResult))
}

// Helper functions

func printValidationResult(name string, result *validator.ValidationResult) {
	status := "VALID"
	if !result.Valid {
		status = "INVALID"
	}

	fmt.Printf("%s - %s (errors: %d, warnings: %d)\n",
		name, status, result.ErrorCount(), result.WarningCount())

	if len(result.Issues) > 0 {
		fmt.Println("Issues:")
		for _, issue := range result.Issues {
			path := ""
			if len(issue.Expression) > 0 {
				path = " at " + issue.Expression[0]
			}
			fmt.Printf("  [%s] %s%s: %s\n",
				issue.Severity, issue.Code, path, issue.Diagnostics)
		}
	}
}

func printValidationSummary(result *validator.ValidationResult) {
	status := "VALID"
	if !result.Valid {
		status = "INVALID"
	}
	fmt.Printf("    Result: %s (errors: %d, warnings: %d)\n",
		status, result.ErrorCount(), result.WarningCount())
}
