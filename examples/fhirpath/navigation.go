// navigation.go demonstrates basic FHIRPath navigation expressions.
// This includes:
// - Path navigation (Resource.field.subfield)
// - Array indexing (field[0])
// - Nested field access
package main

import "fmt"

// RunNavigationExamples demonstrates basic path navigation
func RunNavigationExamples(patient, observation []byte) {
	fmt.Println("\n" + separator)
	fmt.Println("BASIC PATH NAVIGATION")
	fmt.Println(separator)

	fmt.Println("\n--- Simple Field Access ---")
	expressions := []string{
		"Patient.id",   // Get resource ID with type prefix
		"id",           // Get resource ID (implicit)
		"resourceType", // Get resource type
		"active",       // Get boolean field
		"gender",       // Get code field
		"birthDate",    // Get date field
	}
	for _, expr := range expressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Array/Collection Access ---")
	arrayExpressions := []string{
		"name",    // Get all names (array)
		"name[0]", // Get first name
		"name[1]", // Get second name
		"telecom", // Get all telecom entries
		"address", // Get all addresses
	}
	for _, expr := range arrayExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Nested Field Access ---")
	nestedExpressions := []string{
		"name.family",      // Get all family names
		"name.given",       // Get all given names (flattened from arrays)
		"name[0].family",   // Get first name's family
		"name[0].given",    // Get first name's given names
		"name[0].given[0]", // Get first given name of first name
		"address.city",     // Get all cities
		"address.line",     // Get all address lines (flattened)
		"telecom.value",    // Get all telecom values
	}
	for _, expr := range nestedExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Deep Navigation ---")
	deepExpressions := []string{
		"contact.name.family",   // Contact's family name
		"contact.telecom.value", // Contact's telecom values
		"identifier.system",     // Identifier systems
		"identifier.value",      // Identifier values
	}
	for _, expr := range deepExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Observation Navigation ---")
	obsExpressions := []string{
		"status",                        // Observation status
		"code.coding",                   // Code codings
		"code.coding.code",              // LOINC codes
		"code.coding.display",           // Display names
		"subject.reference",             // Subject reference
		"effectiveDateTime",             // Effective date/time
		"component",                     // All components
		"component.code.coding.code",    // Component codes
		"component.valueQuantity.value", // Component values
		"component.valueQuantity.unit",  // Component units
	}
	for _, expr := range obsExpressions {
		evaluate(observation, expr)
	}
}
