// Package fhirpath provides a FHIRPath expression evaluator.
//
// FHIRPath is a path-based navigation and extraction language for FHIR resources.
// This implementation supports the full FHIRPath specification including:
//   - Path navigation
//   - Filtering and projection
//   - Boolean logic
//   - String manipulation
//   - Math operations
//   - Date/time operations
//   - Type operations
//   - FHIR-specific functions
//
// Usage:
//
//	result, err := fhirpath.Evaluate("name.given.first()", patient)
//	exists, err := fhirpath.EvaluateToBoolean("active.exists()", patient)
package fhirpath
