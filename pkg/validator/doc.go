// Package validator provides FHIR resource validation.
//
// The validator checks resources against:
//   - Structural constraints (cardinality, types)
//   - FHIRPath invariants from StructureDefinitions
//   - Primitive type formats (dates, URIs, etc.)
//   - Terminology bindings (optional)
//   - Reference validity (optional)
//
// Usage:
//
//	v, err := validator.NewValidator(&validator.Options{
//	    FHIRVersion:         "R4",
//	    ValidateConstraints: true,
//	})
//	outcome, err := v.Validate(ctx, patient)
package validator
