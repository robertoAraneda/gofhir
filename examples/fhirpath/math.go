// math.go demonstrates FHIRPath math functions and operations.
// This includes:
// - Arithmetic operators (+, -, *, /)
// - Comparison operators (<, <=, >, >=)
// - Math functions (abs, ceiling, floor, round)
package main

import "fmt"

// RunMathExamples demonstrates math functions
func RunMathExamples(observation []byte) {
	fmt.Println("\n" + separator)
	fmt.Println("MATH FUNCTIONS")
	fmt.Println(separator)

	fmt.Println("\n--- Extracting Numeric Values ---")
	numericExpressions := []string{
		"component.valueQuantity.value",         // All component values
		"component.valueQuantity.value.first()", // First value (systolic)
		"component.valueQuantity.value.last()",  // Last value (diastolic)
		"component.count()",                     // Number of components
	}
	for _, expr := range numericExpressions {
		evaluate(observation, expr)
	}

	fmt.Println("\n--- Arithmetic Operators ---")
	arithmeticExpressions := []string{
		"1 + 2",          // Addition
		"10 - 3",         // Subtraction
		"5 * 4",          // Multiplication
		"20 / 4",         // Division
		"17 mod 5",       // Modulo
		"17 div 5",       // Integer division
		"(120 + 80) / 2", // Mean arterial calculation
	}
	for _, expr := range arithmeticExpressions {
		evaluate(observation, expr)
	}

	fmt.Println("\n--- Comparison Operators ---")
	comparisonExpressions := []string{
		"120 > 100",  // Greater than
		"80 < 90",    // Less than
		"120 >= 120", // Greater or equal
		"80 <= 80",   // Less or equal
		"120 = 120",  // Equal
		"120 != 80",  // Not equal
	}
	for _, expr := range comparisonExpressions {
		evaluate(observation, expr)
	}

	fmt.Println("\n--- Blood Pressure Analysis ---")
	fmt.Println("Using math on actual Observation values")
	bpExpressions := []string{
		// Get systolic value
		"component.where(code.coding.code = '8480-6').valueQuantity.value",
		// Get diastolic value
		"component.where(code.coding.code = '8462-4').valueQuantity.value",
		// Check if systolic is high (>140)
		"component.where(code.coding.code = '8480-6').valueQuantity.value > 140",
		// Check if systolic is normal (<140)
		"component.where(code.coding.code = '8480-6').valueQuantity.value < 140",
		// Check if diastolic is normal (<90)
		"component.where(code.coding.code = '8462-4').valueQuantity.value < 90",
	}
	for _, expr := range bpExpressions {
		evaluate(observation, expr)
	}

	fmt.Println("\n--- Math Functions ---")
	mathFuncExpressions := []string{
		"(-5).abs()",       // Absolute value
		"3.7.ceiling()",    // Ceiling (round up)
		"3.7.floor()",      // Floor (round down)
		"3.5.round()",      // Round
		"3.14159.round(2)", // Round to 2 decimals
		"16.sqrt()",        // Square root
		"2.power(3)",       // Power (2^3)
		"10.ln()",          // Natural logarithm
		"100.log(10)",      // Log base 10
		"(-1).abs() + 5",   // Combined operations
	}
	for _, expr := range mathFuncExpressions {
		evaluate(observation, expr)
	}

	fmt.Println("\n--- Aggregate Functions ---")
	aggregateExpressions := []string{
		"component.valueQuantity.value.sum()", // Sum of all values
		"component.valueQuantity.value.min()", // Minimum value
		"component.valueQuantity.value.max()", // Maximum value
		"component.valueQuantity.value.avg()", // Average value
	}
	for _, expr := range aggregateExpressions {
		evaluate(observation, expr)
	}
}
