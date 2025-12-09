// filtering.go demonstrates FHIRPath filtering functions.
// This includes:
// - where() - Filter collection by condition
// - select() - Transform collection elements
// - Chaining filters
package main

import "fmt"

// RunFilteringExamples demonstrates filtering functions
func RunFilteringExamples(patient, observation []byte) {
	fmt.Println("\n" + separator)
	fmt.Println("FILTERING FUNCTIONS")
	fmt.Println(separator)

	fmt.Println("\n--- where() Function ---")
	fmt.Println("Filters collection by a boolean condition")
	whereExpressions := []string{
		"name.where(use = 'official')",    // Filter by use
		"name.where(use = 'nickname')",    // Get nicknames
		"telecom.where(system = 'phone')", // Get phones only
		"telecom.where(system = 'email')", // Get emails only
		"telecom.where(use = 'home')",     // Get home contacts
		"telecom.where(use = 'work')",     // Get work contacts
	}
	for _, expr := range whereExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- where() then Field Selection ---")
	fmt.Println("Filter first, then select specific fields")
	chainedExpressions := []string{
		"name.where(use = 'official').family",   // Official family name
		"name.where(use = 'official').given",    // Official given names
		"telecom.where(system = 'phone').value", // Phone numbers only
		"telecom.where(system = 'email').value", // Email addresses
		"telecom.where(use = 'home').value",     // Home contact values
	}
	for _, expr := range chainedExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Complex where() Conditions ---")
	fmt.Println("Using 'and', 'or' in filter conditions")
	complexExpressions := []string{
		"telecom.where(system = 'phone' and use = 'home')",    // Home phone
		"telecom.where(system = 'phone' and use = 'work')",    // Work phone
		"telecom.where(system = 'phone' or system = 'email')", // Phone or email
		"name.where(use = 'official' or use = 'usual')",       // Official or usual
	}
	for _, expr := range complexExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- where() with String Functions ---")
	fmt.Println("Combining where() with string operations")
	stringFilterExpressions := []string{
		"identifier.where(system.contains('hospital'))", // Hospital identifier
		"identifier.where(system.contains('national'))", // National identifier
		"telecom.where(value.startsWith('+1'))",         // US phone numbers
		"name.where(family.startsWith('D'))",            // Family names starting with D
	}
	for _, expr := range stringFilterExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- select() Function ---")
	fmt.Println("Transform/project collection elements")
	selectExpressions := []string{
		"name.select(family)",                   // Just family names
		"name.select(given.first())",            // First given name of each
		"telecom.select(system + ': ' + value)", // Format as 'system: value'
		"address.select(city + ', ' + state)",   // Format address
	}
	for _, expr := range selectExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Observation Filtering ---")
	fmt.Println("Filtering Observation components")
	obsFilterExpressions := []string{
		"component.where(code.coding.code = '8480-6')",                     // Systolic component
		"component.where(code.coding.code = '8462-4')",                     // Diastolic component
		"component.where(code.coding.code = '8480-6').valueQuantity",       // Systolic value
		"component.where(code.coding.code = '8480-6').valueQuantity.value", // Systolic number
	}
	for _, expr := range obsFilterExpressions {
		evaluate(observation, expr)
	}
}
