// types.go demonstrates FHIRPath type functions.
// This includes:
// - Type conversion (toInteger, toString, toDecimal, toBoolean)
// - Type checking (is, as)
// - ofType() for filtering by type
package main

import "fmt"

// RunTypeExamples demonstrates type functions
func RunTypeExamples(patient []byte) {
	fmt.Println("\n" + separator)
	fmt.Println("TYPE FUNCTIONS")
	fmt.Println(separator)

	fmt.Println("\n--- Type Conversion: toInteger() ---")
	fmt.Println("Convert to integer")
	toIntExpressions := []string{
		"'123'.toInteger()",     // String to integer
		"'456'.toInteger() + 1", // Convert and use in math
		"true.toInteger()",      // Boolean true = 1
		"false.toInteger()",     // Boolean false = 0
	}
	for _, expr := range toIntExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Type Conversion: toString() ---")
	fmt.Println("Convert to string")
	toStringExpressions := []string{
		"123.toString()",         // Integer to string
		"3.14.toString()",        // Decimal to string
		"true.toString()",        // Boolean to string
		"false.toString()",       // Boolean to string
		"@2024-01-15.toString()", // Date to string
	}
	for _, expr := range toStringExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Type Conversion: toDecimal() ---")
	fmt.Println("Convert to decimal")
	toDecimalExpressions := []string{
		"'3.14'.toDecimal()", // String to decimal
		"'100'.toDecimal()",  // String integer to decimal
		"5.toDecimal()",      // Integer to decimal
	}
	for _, expr := range toDecimalExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Type Conversion: toBoolean() ---")
	fmt.Println("Convert to boolean")
	toBoolExpressions := []string{
		"'true'.toBoolean()",  // String to boolean
		"'false'.toBoolean()", // String to boolean
		"1.toBoolean()",       // Integer 1 = true
		"0.toBoolean()",       // Integer 0 = false
	}
	for _, expr := range toBoolExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Type Checking: is() ---")
	fmt.Println("Check if value is of specific type")
	isExpressions := []string{
		"active.is(Boolean)", // Is active a boolean?
		"gender.is(String)",  // Is gender a string?
		"name.is(HumanName)", // Is name a HumanName?
		"birthDate.is(Date)", // Is birthDate a date?
	}
	for _, expr := range isExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Type Casting: as() ---")
	fmt.Println("Cast value to specific type")
	asExpressions := []string{
		"active.as(Boolean)", // Cast to boolean
		"gender.as(String)",  // Cast to string
	}
	for _, expr := range asExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- Literal Types ---")
	fmt.Println("Creating typed literals")
	literalExpressions := []string{
		"@2024-01-15",           // Date literal
		"@2024-01-15T10:30:00Z", // DateTime literal
		"@T10:30:00",            // Time literal
		"10 'kg'",               // Quantity literal
		"100 'cm'",              // Quantity literal
	}
	for _, expr := range literalExpressions {
		evaluate(patient, expr)
	}

	fmt.Println("\n--- iif() Function ---")
	fmt.Println("Conditional: iif(condition, true-result, false-result)")
	iifExpressions := []string{
		"iif(active, 'Active', 'Inactive')",           // If active
		"iif(gender = 'male', 'M', 'F')",              // Gender short form
		"iif(name.count() > 1, 'Multiple', 'Single')", // Name count check
		"iif(deceased.exists(), 'Deceased', 'Alive')", // Deceased check
	}
	for _, expr := range iifExpressions {
		evaluate(patient, expr)
	}
}
