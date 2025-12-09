// Package main demonstrates FHIRPath expression evaluation using gofhir.
// This example covers various FHIRPath functions including existence, filtering,
// string manipulation, math operations, and FHIR-specific functions.
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath"
)

func main() {
	fmt.Println("=== FHIRPath Examples ===")

	// Sample Patient resource for demonstrations
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "example",
		"active": true,
		"name": [
			{
				"use": "official",
				"family": "Doe",
				"given": ["John", "James"]
			},
			{
				"use": "nickname",
				"given": ["Johnny"]
			}
		],
		"telecom": [
			{"system": "phone", "value": "+1-555-0100", "use": "home"},
			{"system": "phone", "value": "+1-555-0101", "use": "work"},
			{"system": "email", "value": "john.doe@example.com"}
		],
		"gender": "male",
		"birthDate": "1990-05-15",
		"address": [
			{
				"use": "home",
				"line": ["123 Main St", "Apt 4B"],
				"city": "Springfield",
				"state": "IL",
				"postalCode": "62701"
			}
		],
		"contact": [
			{
				"relationship": [{"coding": [{"code": "E"}]}],
				"name": {"family": "Doe", "given": ["Jane"]},
				"telecom": [{"system": "phone", "value": "+1-555-0200"}]
			}
		],
		"identifier": [
			{"system": "http://hospital.example.org/patients", "value": "12345"},
			{"system": "http://national.example.org/id", "value": "98765"}
		]
	}`)

	// Sample Observation for more examples
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "blood-pressure",
		"status": "final",
		"code": {
			"coding": [
				{"system": "http://loinc.org", "code": "85354-9", "display": "Blood pressure panel"}
			]
		},
		"subject": {"reference": "Patient/example"},
		"effectiveDateTime": "2024-01-15T10:30:00Z",
		"component": [
			{
				"code": {"coding": [{"system": "http://loinc.org", "code": "8480-6", "display": "Systolic"}]},
				"valueQuantity": {"value": 120, "unit": "mmHg"}
			},
			{
				"code": {"coding": [{"system": "http://loinc.org", "code": "8462-4", "display": "Diastolic"}]},
				"valueQuantity": {"value": 80, "unit": "mmHg"}
			}
		]
	}`)

	// 1. Basic Path Navigation
	fmt.Println("\n--- 1. Basic Path Navigation ---")
	demonstrateBasicNavigation(patient)

	// 2. Existence Functions
	fmt.Println("\n--- 2. Existence Functions ---")
	demonstrateExistenceFunctions(patient)

	// 3. Filtering Functions
	fmt.Println("\n--- 3. Filtering Functions ---")
	demonstrateFilteringFunctions(patient)

	// 4. String Functions
	fmt.Println("\n--- 4. String Functions ---")
	demonstrateStringFunctions(patient)

	// 5. Math Functions
	fmt.Println("\n--- 5. Math Functions ---")
	demonstrateMathFunctions(observation)

	// 6. Collection Functions
	fmt.Println("\n--- 6. Collection Functions ---")
	demonstrateCollectionFunctions(patient)

	// 7. Type Functions
	fmt.Println("\n--- 7. Type Functions ---")
	demonstrateTypeFunctions(patient)

	// 8. Combining Functions
	fmt.Println("\n--- 8. Combining Functions (Complex Queries) ---")
	demonstrateComplexQueries(patient, observation)

	// 9. Compiled Expression Reuse
	fmt.Println("\n--- 9. Compiled Expression Reuse ---")
	demonstrateCompiledExpressions(patient)

	// 10. Constraint-style Expressions
	fmt.Println("\n--- 10. Constraint-style Expressions ---")
	demonstrateConstraintExpressions(patient)
}

// demonstrateBasicNavigation shows simple path navigation
func demonstrateBasicNavigation(patient []byte) {
	expressions := []string{
		"Patient.id",                   // Get resource ID
		"name",                         // Get all names
		"name.family",                  // Get all family names
		"name.given",                   // Get all given names (flattened)
		"name[0].given",                // Get first name's given names
		"telecom.value",                // Get all telecom values
		"address.line",                 // Get all address lines
		"birthDate",                    // Get birth date
		"gender",                       // Get gender
	}

	for _, expr := range expressions {
		evaluate(patient, expr)
	}
}

// demonstrateExistenceFunctions shows existence-related functions
func demonstrateExistenceFunctions(patient []byte) {
	expressions := []string{
		"name.exists()",                              // Check if name exists
		"name.empty()",                               // Check if name is empty
		"deceasedBoolean.exists()",                   // Check optional field
		"name.count()",                               // Count names
		"telecom.count()",                            // Count telecom entries
		"name.where(use = 'official').exists()",     // Conditional existence
		"identifier.count() > 1",                    // Numeric comparison
		"active = true",                              // Boolean comparison
	}

	for _, expr := range expressions {
		evaluate(patient, expr)
	}
}

// demonstrateFilteringFunctions shows where() and other filtering
func demonstrateFilteringFunctions(patient []byte) {
	expressions := []string{
		"name.where(use = 'official')",               // Filter by use
		"name.where(use = 'official').family",        // Filter then select
		"telecom.where(system = 'phone')",            // Filter telecoms
		"telecom.where(system = 'phone').value",      // Get phone numbers
		"telecom.where(use = 'home').value",          // Get home contact
		"identifier.where(system.contains('hospital'))", // Complex filter
	}

	for _, expr := range expressions {
		evaluate(patient, expr)
	}
}

// demonstrateStringFunctions shows string manipulation
func demonstrateStringFunctions(patient []byte) {
	expressions := []string{
		"name.family.lower()",                        // Lowercase
		"name.family.upper()",                        // Uppercase
		"name.given.first().length()",                // String length
		"gender.startsWith('m')",                     // Starts with
		"address.city.contains('field')",             // Contains
		"telecom.value.where($this.startsWith('+1'))", // Filter by prefix
		"name.family.substring(0, 2)",                // Substring
		"name.given.first() + ' ' + name.family",     // String concatenation
	}

	for _, expr := range expressions {
		evaluate(patient, expr)
	}
}

// demonstrateMathFunctions shows math operations
func demonstrateMathFunctions(observation []byte) {
	expressions := []string{
		"component.valueQuantity.value",              // Get values
		"component.valueQuantity.value.first()",      // Get first value
		"component.count()",                          // Count components
		"(120 + 80) / 2",                             // Arithmetic
		"component.valueQuantity.value.first() > 100", // Comparison
	}

	for _, expr := range expressions {
		evaluate(observation, expr)
	}
}

// demonstrateCollectionFunctions shows collection operations
func demonstrateCollectionFunctions(patient []byte) {
	expressions := []string{
		"name.first()",                               // First element
		"name.last()",                                // Last element
		"name.tail()",                                // All but first
		"name.given.distinct()",                      // Unique values
		"telecom.skip(1)",                            // Skip first
		"telecom.take(2)",                            // Take first two
		"name | address",                             // Union
		"name.single()",                              // Single (fails if > 1) - wrapped in try
	}

	for _, expr := range expressions {
		evaluate(patient, expr)
	}
}

// demonstrateTypeFunctions shows type conversion functions
func demonstrateTypeFunctions(patient []byte) {
	expressions := []string{
		"(active).toInteger()",                       // Boolean to integer (true=1)
		"'123'.toInteger()",                          // String to integer
		"123.toString()",                             // Integer to string
		"'3.14'.toDecimal()",                         // String to decimal
		"true.toString()",                            // Boolean to string
	}

	for _, expr := range expressions {
		evaluate(patient, expr)
	}
}

// demonstrateComplexQueries shows combinations of functions
func demonstrateComplexQueries(patient, observation []byte) {
	fmt.Println("Patient queries:")
	patientQueries := []string{
		// Get official full name
		"name.where(use = 'official').given.first() + ' ' + name.where(use = 'official').family",

		// Check if patient has home phone
		"telecom.where(system = 'phone' and use = 'home').exists()",

		// Get all contact methods
		"telecom.value | contact.telecom.value",

		// Check multiple identifiers exist
		"identifier.count() >= 2",

		// Complex address query
		"address.where(use = 'home').select(line.first() + ', ' + city + ', ' + state)",
	}

	for _, expr := range patientQueries {
		evaluate(patient, expr)
	}

	fmt.Println("\nObservation queries:")
	observationQueries := []string{
		// Get systolic value
		"component.where(code.coding.code = '8480-6').valueQuantity.value",

		// Check if blood pressure is normal
		"component.where(code.coding.code = '8480-6').valueQuantity.value < 140",

		// Get all component codes
		"component.code.coding.display",
	}

	for _, expr := range observationQueries {
		evaluate(observation, expr)
	}
}

// demonstrateCompiledExpressions shows expression reuse
func demonstrateCompiledExpressions(patient []byte) {
	// Compile once, evaluate multiple times
	expr, err := fhirpath.Compile("name.family")
	if err != nil {
		log.Printf("Compile error: %v", err)
		return
	}

	// Evaluate same expression on different resources
	patients := [][]byte{
		[]byte(`{"resourceType": "Patient", "name": [{"family": "Smith"}]}`),
		[]byte(`{"resourceType": "Patient", "name": [{"family": "Johnson"}]}`),
		[]byte(`{"resourceType": "Patient", "name": [{"family": "Williams"}]}`),
	}

	fmt.Println("Compiled expression 'name.family' evaluated on multiple patients:")
	for i, p := range patients {
		result, err := expr.Evaluate(p)
		if err != nil {
			log.Printf("Patient %d error: %v", i+1, err)
			continue
		}
		fmt.Printf("  Patient %d: %v\n", i+1, formatResult(result))
	}

	// Using MustCompile for expressions you know are valid
	nameExpr := fhirpath.MustCompile("name.given.first()")
	result, _ := nameExpr.Evaluate(patient)
	fmt.Printf("First given name: %v\n", formatResult(result))
}

// demonstrateConstraintExpressions shows FHIR constraint-style expressions
func demonstrateConstraintExpressions(patient []byte) {
	// These are typical FHIRPath expressions used in FHIR constraints
	constraints := []struct {
		name       string
		expression string
	}{
		{
			name:       "pat-1 (contact must have details)",
			expression: "contact.all(name.exists() or telecom.exists() or address.exists() or organization.exists())",
		},
		{
			name:       "has-name",
			expression: "name.count() >= 1",
		},
		{
			name:       "has-identifier",
			expression: "identifier.exists()",
		},
		{
			name:       "active-patient",
			expression: "active.exists() and active = true",
		},
		{
			name:       "valid-gender",
			expression: "gender.exists() implies gender.matches('male|female|other|unknown')",
		},
	}

	for _, c := range constraints {
		result, err := fhirpath.Evaluate(patient, c.expression)
		if err != nil {
			fmt.Printf("Constraint '%s' error: %v\n", c.name, err)
			continue
		}

		passed := isTruthy(result)
		status := "PASS"
		if !passed {
			status = "FAIL"
		}
		fmt.Printf("[%s] %s: %s\n", status, c.name, c.expression)
	}
}

// isTruthy checks if a FHIRPath result is truthy
func isTruthy(result interface{}) bool {
	switch v := result.(type) {
	case []interface{}:
		if len(v) == 0 {
			return false
		}
		if len(v) == 1 {
			return isTruthy(v[0])
		}
		return true // Non-empty collection is truthy
	case bool:
		return v
	case nil:
		return false
	default:
		return true // Any non-nil value is truthy
	}
}

// Helper functions

func evaluate(resource []byte, expr string) {
	result, err := fhirpath.Evaluate(resource, expr)
	if err != nil {
		fmt.Printf("  %s => ERROR: %v\n", expr, err)
		return
	}
	fmt.Printf("  %s => %v\n", expr, formatResult(result))
}

func formatResult(result interface{}) string {
	switch v := result.(type) {
	case fhirpath.Collection:
		return formatCollection(v)
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		formatted := make([]string, 0, len(v))
		for _, item := range v {
			formatted = append(formatted, formatValue(item))
		}
		if len(formatted) == 1 {
			return formatted[0]
		}
		return "[" + joinStrings(formatted, ", ") + "]"
	default:
		return formatValue(v)
	}
}

func formatCollection(c fhirpath.Collection) string {
	if len(c) == 0 {
		return "{ }" // FHIRPath empty collection notation
	}
	formatted := make([]string, 0, len(c))
	for _, item := range c {
		formatted = append(formatted, formatValue(item))
	}
	if len(formatted) == 1 {
		return formatted[0]
	}
	// FHIRPath uses pipe notation for collections, but for readability we use { }
	return "{ " + joinStrings(formatted, ", ") + " }"
}

func formatValue(v interface{}) string {
	// Handle FHIRPath types by checking for common methods
	// FHIRPath types typically implement String() for their representation

	// First, check for ObjectValue (complex FHIR types) - they have Data() method
	if obj, ok := v.(interface{ Data() []byte }); ok {
		return formatObjectValue(obj.Data())
	}

	// Check for String() method (most FHIRPath types have this)
	if stringer, ok := v.(interface{ String() string }); ok {
		s := stringer.String()
		// Boolean values: true, false (no quotes per FHIRPath spec)
		if s == "true" || s == "false" {
			return s
		}
		// If it's not an object representation, return the string value
		// FHIRPath strings should be quoted with single quotes
		if s != "" && s[0] != '{' {
			// Check if this is a string type that should be quoted
			if _, isStr := v.(interface{ Value() string }); isStr {
				return "'" + s + "'" // FHIRPath uses single quotes for strings
			}
			return s
		}
	}

	switch val := v.(type) {
	case interface{ Bool() bool }:
		// Boolean: true or false (no quotes)
		return fmt.Sprintf("%t", val.Bool())
	case interface{ Value() string }:
		// String: single-quoted per FHIRPath spec
		return fmt.Sprintf("'%s'", val.Value())
	case interface{ Value() int64 }:
		// Integer: plain number
		return fmt.Sprintf("%d", val.Value())
	case bool:
		return fmt.Sprintf("%t", val)
	case string:
		// String literals use single quotes in FHIRPath
		return fmt.Sprintf("'%s'", val)
	case int, int64, int32:
		return fmt.Sprintf("%d", val)
	case float64, float32:
		return fmt.Sprintf("%v", val)
	case map[string]interface{}:
		// Complex types - show meaningful representation
		// Check for resourceType (FHIR resource)
		if rt, ok := val["resourceType"].(string); ok {
			if id, ok := val["id"].(string); ok {
				return fmt.Sprintf("%s/%s", rt, id)
			}
			return fmt.Sprintf("%s", rt)
		}
		// For complex datatypes like HumanName, show key fields
		return formatComplexType(val)
	default:
		// Try JSON as fallback
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		s := string(data)
		// Truncate long values
		if len(s) > 80 {
			return s[:77] + "..."
		}
		return s
	}
}

// formatObjectValue formats a FHIRPath ObjectValue from its JSON data
func formatObjectValue(data []byte) string {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return "{...}"
	}
	return formatComplexType(obj)
}

// formatComplexType formats FHIR complex datatypes for display
func formatComplexType(val map[string]interface{}) string {
	// HumanName - has family and/or given
	if family, ok := val["family"].(string); ok {
		if given, ok := val["given"].([]interface{}); ok && len(given) > 0 {
			return fmt.Sprintf("HumanName('%s, %v')", family, given[0])
		}
		return fmt.Sprintf("HumanName('%s')", family)
	}

	// Both Identifier and ContactPoint have system+value, distinguish by system format
	if system, ok := val["system"].(string); ok {
		if value, ok := val["value"].(string); ok {
			// ContactPoint systems are: phone, fax, email, pager, url, sms, other
			contactSystems := map[string]bool{
				"phone": true, "fax": true, "email": true,
				"pager": true, "url": true, "sms": true, "other": true,
			}
			if contactSystems[system] {
				return fmt.Sprintf("ContactPoint(%s: '%s')", system, value)
			}
			// Identifier systems are typically URIs (http://, urn:, etc.)
			return fmt.Sprintf("Identifier('%s')", value)
		}
	}

	// Address - has city and/or line
	if city, ok := val["city"].(string); ok {
		if state, ok := val["state"].(string); ok {
			return fmt.Sprintf("Address('%s, %s')", city, state)
		}
		return fmt.Sprintf("Address('%s')", city)
	}
	// Coding
	if code, ok := val["code"].(string); ok {
		if display, ok := val["display"].(string); ok {
			return fmt.Sprintf("Coding('%s': '%s')", code, display)
		}
		return fmt.Sprintf("Coding('%s')", code)
	}
	// CodeableConcept
	if coding, ok := val["coding"].([]interface{}); ok && len(coding) > 0 {
		if c, ok := coding[0].(map[string]interface{}); ok {
			if code, ok := c["code"].(string); ok {
				return fmt.Sprintf("CodeableConcept('%s')", code)
			}
		}
	}
	// Quantity
	if value, ok := val["value"].(float64); ok {
		if unit, ok := val["unit"].(string); ok {
			return fmt.Sprintf("Quantity(%v '%s')", value, unit)
		}
		return fmt.Sprintf("Quantity(%v)", value)
	}
	// Reference
	if ref, ok := val["reference"].(string); ok {
		return fmt.Sprintf("Reference('%s')", ref)
	}
	// Generic: show first few keys
	keys := make([]string, 0)
	for k := range val {
		keys = append(keys, k)
		if len(keys) >= 3 {
			break
		}
	}
	return fmt.Sprintf("{%s...}", joinStrings(keys, ", "))
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
