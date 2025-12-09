// strings.go demonstrates FHIRPath string functions.
// This includes:
// - Case conversion (lower, upper)
// - String inspection (length, startsWith, endsWith, contains)
// - String manipulation (substring, replace)
// - String concatenation
package main

import "fmt"

// RunStringExamples demonstrates string manipulation functions
func RunStringExamples(patient []byte) {
	fmt.Println("\n" + separator)
	fmt.Println("STRING FUNCTIONS")
	fmt.Println(separator)

	fmt.Println("\n--- Case Conversion ---")
	caseExpressions := []string{
		"name.family.lower()", // Lowercase family names
		"name.family.upper()", // Uppercase family names
		"gender.lower()",      // Lowercase gender
		"gender.upper()",      // Uppercase gender
	}
	for _, expr := range caseExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- length() Function ---")
	fmt.Println("Returns the length of string")
	lengthExpressions := []string{
		"name.family.length()",        // Length of family names
		"name.given.first().length()", // Length of first given name
		"id.length()",                 // Length of ID
		"gender.length()",             // Length of gender
	}
	for _, expr := range lengthExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- startsWith() and endsWith() ---")
	fmt.Println("Check string prefix/suffix")
	startsEndsExpressions := []string{
		"name.family.startsWith('D')",    // Family name starts with D
		"name.family.startsWith('S')",    // Family name starts with S
		"birthDate.startsWith('1990')",   // Birth year
		"gender.endsWith('ale')",         // Gender ends with 'ale'
		"telecom.value.startsWith('+1')", // US phone numbers
	}
	for _, expr := range startsEndsExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- contains() Function ---")
	fmt.Println("Check if string contains substring")
	containsExpressions := []string{
		"name.family.contains('oe')",             // Family contains 'oe'
		"address.city.contains('field')",         // City contains 'field'
		"telecom.value.contains('555')",          // Phone contains 555
		"identifier.system.contains('hospital')", // System contains 'hospital'
	}
	for _, expr := range containsExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- substring() Function ---")
	fmt.Println("Extract part of string: substring(start, length)")
	substringExpressions := []string{
		"name.family.substring(0, 2)", // First 2 chars of family
		"birthDate.substring(0, 4)",   // Year from birthDate
		"birthDate.substring(5, 2)",   // Month from birthDate
		"birthDate.substring(8, 2)",   // Day from birthDate
		"id.substring(0, 3)",          // First 3 chars of ID
	}
	for _, expr := range substringExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- matches() Function ---")
	fmt.Println("Check if string matches regex pattern")
	matchesExpressions := []string{
		"gender.matches('male|female')",                           // Valid gender
		"birthDate.matches('[0-9]{4}-[0-9]{2}-[0-9]{2}')",         // Date format
		"telecom.value.where($this.matches('\\\\+1.*')).exists()", // Has US phone
	}
	for _, expr := range matchesExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- String Concatenation ---")
	fmt.Println("Combining strings with + operator")
	concatExpressions := []string{
		"name.given.first() + ' ' + name.family",                            // Full name
		"'Patient: ' + name.family",                                         // Prefix
		"address.city + ', ' + address.state",                               // City, State
		"address.line.first() + ', ' + address.city + ', ' + address.state", // Full address
	}
	for _, expr := range concatExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- indexOf() Function ---")
	fmt.Println("Find position of substring")
	indexOfExpressions := []string{
		"name.family.indexOf('o')", // Position of 'o' in family name
		"birthDate.indexOf('-')",   // Position of first dash
	}
	for _, expr := range indexOfExpressions {
		evaluate(patient, expr)
	}
}
