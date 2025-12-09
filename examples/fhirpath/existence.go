// existence.go demonstrates FHIRPath existence functions.
// This includes:
// - exists() - Check if collection has elements
// - empty() - Check if collection is empty
// - count() - Count elements in collection
// - Boolean comparisons
package main

import "fmt"

// RunExistenceExamples demonstrates existence-related functions
func RunExistenceExamples(patient []byte) {
	fmt.Println("\n" + separator)
	fmt.Println("EXISTENCE FUNCTIONS")
	fmt.Println(separator)

	fmt.Println("\n--- exists() Function ---")
	fmt.Println("Returns true if collection has at least one element")
	existsExpressions := []string{
		"name.exists()",            // Check if name exists
		"active.exists()",          // Check if active exists
		"deceased.exists()",        // Check optional field (should be false)
		"deceasedBoolean.exists()", // Check specific deceased type
		"identifier.exists()",      // Check if identifiers exist
		"photo.exists()",           // Check optional field
	}
	for _, expr := range existsExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- empty() Function ---")
	fmt.Println("Returns true if collection has no elements")
	emptyExpressions := []string{
		"name.empty()",          // Check if name is empty (false)
		"photo.empty()",         // Check if photo is empty (true)
		"deceased.empty()",      // Check if deceased is empty
		"maritalStatus.empty()", // Check optional field
	}
	for _, expr := range emptyExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- count() Function ---")
	fmt.Println("Returns the number of elements in collection")
	countExpressions := []string{
		"name.count()",          // Count names
		"telecom.count()",       // Count telecom entries
		"address.count()",       // Count addresses
		"identifier.count()",    // Count identifiers
		"contact.count()",       // Count contacts
		"name.given.count()",    // Count all given names (flattened)
		"name[0].given.count()", // Count first name's given names
	}
	for _, expr := range countExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Boolean Comparisons ---")
	fmt.Println("Comparing values with equality and inequality")
	boolExpressions := []string{
		"active = true",          // Check if active is true
		"active != false",        // Check if active is not false
		"gender = 'male'",        // Check gender
		"gender != 'female'",     // Check not female
		"name.count() > 0",       // Has at least one name
		"name.count() >= 2",      // Has at least two names
		"identifier.count() = 2", // Has exactly two identifiers
		"telecom.count() < 10",   // Has fewer than 10 telecoms
	}
	for _, expr := range boolExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Conditional Existence ---")
	fmt.Println("Combining where() with exists()")
	conditionalExpressions := []string{
		"name.where(use = 'official').exists()",    // Has official name
		"name.where(use = 'nickname').exists()",    // Has nickname
		"telecom.where(system = 'phone').exists()", // Has phone
		"telecom.where(system = 'email').exists()", // Has email
		"telecom.where(use = 'home').exists()",     // Has home contact
		"address.where(use = 'home').exists()",     // Has home address
	}
	for _, expr := range conditionalExpressions {
		evaluate(patient, expr)
	}
}
