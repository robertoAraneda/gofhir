// Package main demonstrates FHIR resource validation using gofhir.
// This example shows how to validate resources against StructureDefinitions.
//
// Examples are organized by validation type:
// - structural_validation.go: Structure, cardinality, primitive types
// - constraint_validation.go: FHIRPath constraint validation
// - reference_validation.go: Reference validation (contained, relative, absolute, URN)
// - extension_validation.go: Extension validation (simple, complex, HL7 standard)
// - terminology_validation.go: Terminology binding validation (code, Coding, CodeableConcept)
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/robertoaraneda/gofhir/pkg/validator"
)

// separator is used for visual separation in output
const separator = "=============================================="

func main() {
	fmt.Println("=== FHIR Validator Examples ===")
	fmt.Println("Demonstrating validation of FHIR resources using gofhir")

	// Check if specs are available
	specsDir := filepath.Join("..", "..", "specs", "r4")
	resourcesPath := filepath.Join(specsDir, "profiles-resources.json")
	if _, err := os.Stat(resourcesPath); err != nil {
		log.Println("Warning: specs/r4/profiles-resources.json not found")
		log.Println("Download FHIR R4 specs from https://hl7.org/fhir/R4/downloads.html")
		log.Fatal("Cannot run examples without FHIR specifications")
	}

	ctx := context.Background()

	// Run all validation examples organized by type
	RunStructuralValidationExamples(ctx)
	RunConstraintValidationExamples(ctx)
	RunReferenceValidationExamples(ctx)
	RunExtensionValidationExamples(ctx)
	RunTerminologyValidationExamples(ctx)

	fmt.Println("\n" + separator)
	fmt.Println("All examples completed!")
	fmt.Println(separator)
}

// printResult prints validation result in a standard format
func printResult(name string, result *validator.ValidationResult) {
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
