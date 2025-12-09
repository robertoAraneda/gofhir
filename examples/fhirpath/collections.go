// collections.go demonstrates FHIRPath collection functions.
// This includes:
// - first(), last(), tail() - Element selection
// - take(), skip() - Pagination
// - distinct() - Unique values
// - union (|), intersect, exclude - Set operations
package main

import "fmt"

// RunCollectionExamples demonstrates collection functions
func RunCollectionExamples(patient []byte) {
	fmt.Println("\n" + separator)
	fmt.Println("COLLECTION FUNCTIONS")
	fmt.Println(separator)

	fmt.Println("\n--- first() and last() ---")
	fmt.Println("Get first or last element of collection")
	firstLastExpressions := []string{
		"name.first()",             // First name entry
		"name.last()",              // Last name entry
		"name.first().family",      // Family of first name
		"name.given.first()",       // First given name overall
		"telecom.first()",          // First telecom
		"telecom.last()",           // Last telecom
		"identifier.first().value", // First identifier value
	}
	for _, expr := range firstLastExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- tail() Function ---")
	fmt.Println("Returns all elements except the first")
	tailExpressions := []string{
		"name.tail()",       // All names except first
		"telecom.tail()",    // All telecoms except first
		"name.given.tail()", // All given names except first
	}
	for _, expr := range tailExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- take() and skip() ---")
	fmt.Println("Pagination: take first N or skip first N elements")
	paginationExpressions := []string{
		"telecom.take(2)",         // First 2 telecoms
		"telecom.take(1)",         // First 1 telecom
		"telecom.skip(1)",         // All but first telecom
		"telecom.skip(2)",         // All but first 2 telecoms
		"telecom.skip(1).take(2)", // Skip 1, then take 2
		"name.given.take(3)",      // First 3 given names
	}
	for _, expr := range paginationExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- distinct() Function ---")
	fmt.Println("Returns unique values only")
	distinctExpressions := []string{
		"name.given.distinct()",     // Unique given names
		"telecom.system.distinct()", // Unique telecom systems
		"telecom.use.distinct()",    // Unique telecom uses
	}
	for _, expr := range distinctExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Union Operator (|) ---")
	fmt.Println("Combines two collections")
	unionExpressions := []string{
		"name.family | name.given",              // All names (family + given)
		"telecom.value | contact.telecom.value", // All contact values
	}
	for _, expr := range unionExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- single() Function ---")
	fmt.Println("Returns element if exactly one, errors otherwise")
	singleExpressions := []string{
		"gender.single()",    // Single gender (should work)
		"birthDate.single()", // Single birth date (should work)
		"active.single()",    // Single active flag (should work)
	}
	for _, expr := range singleExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Collection Boolean Functions ---")
	fmt.Println("Check conditions on collections")
	boolExpressions := []string{
		"name.all(family.exists())",          // All names have family
		"telecom.all(value.exists())",        // All telecoms have value
		"name.given.all($this.length() > 2)", // All given names > 2 chars
		"telecom.any(system = 'email')",      // Any telecom is email
		"name.any(use = 'official')",         // Any name is official
	}
	for _, expr := range boolExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Flattening with children() ---")
	flattenExpressions := []string{
		"name.children()",    // All child elements of name
		"address.children()", // All child elements of address
	}
	for _, expr := range flattenExpressions {
		evaluate(patient, expr)
	}
}
