// Package main demonstrates FHIRPath expression evaluation using gofhir.
// This example covers various FHIRPath functions including existence, filtering,
// string manipulation, math operations, and FHIR-specific functions.
//
// Examples are organized by category:
// - navigation.go: Basic path navigation
// - existence.go: exists(), empty(), count()
// - filtering.go: where(), select()
// - strings.go: String manipulation functions
// - math.go: Math functions and operators
// - collections.go: Collection functions
// - types.go: Type conversion and checking
// - constraints.go: Constraint-style expressions
//
//nolint:errcheck // Example code intentionally ignores errors for brevity
package main

import (
	"encoding/json"
	"fmt"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath"
)

const separator = "=============================================="

func main() {
	fmt.Println("=== FHIRPath Examples ===")
	fmt.Println("Demonstrating FHIRPath expression evaluation")

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

	// Run all FHIRPath examples organized by category
	RunNavigationExamples(patient, observation)
	RunExistenceExamples(patient)
	RunFilteringExamples(patient, observation)
	RunStringExamples(patient)
	RunMathExamples(observation)
	RunCollectionExamples(patient)
	RunTypeExamples(patient)
	RunConstraintExamples(patient, observation)

	// Additional: Compiled expressions
	fmt.Println("\n" + separator)
	fmt.Println("COMPILED EXPRESSIONS")
	fmt.Println(separator)
	demonstrateCompiledExpressions(patient)

	fmt.Println("\n" + separator)
	fmt.Println("All FHIRPath examples completed!")
	fmt.Println(separator)
}

// demonstrateCompiledExpressions shows expression reuse
func demonstrateCompiledExpressions(patient []byte) {
	fmt.Println("\n--- Compile Once, Evaluate Multiple Times ---")

	// Compile once
	expr, err := fhirpath.Compile("name.family")
	if err != nil {
		fmt.Printf("Compile error: %v\n", err)
		return
	}

	// Evaluate same expression on different resources
	patients := [][]byte{
		[]byte(`{"resourceType": "Patient", "name": [{"family": "Smith"}]}`),
		[]byte(`{"resourceType": "Patient", "name": [{"family": "Johnson"}]}`),
		[]byte(`{"resourceType": "Patient", "name": [{"family": "Williams"}]}`),
	}

	fmt.Println("  Compiled expression 'name.family' evaluated on multiple patients:")
	for i, p := range patients {
		result, err := expr.Evaluate(p)
		if err != nil {
			fmt.Printf("    Patient %d: ERROR - %v\n", i+1, err)
			continue
		}
		fmt.Printf("    Patient %d: %v\n", i+1, formatResult(result))
	}

	// Using MustCompile for expressions you know are valid
	fmt.Println("\n--- MustCompile for Known-Valid Expressions ---")
	nameExpr := fhirpath.MustCompile("name.given.first()")
	result, _ := nameExpr.Evaluate(patient)
	fmt.Printf("  First given name: %v\n", formatResult(result))
}

// evaluate runs a FHIRPath expression and prints the result
func evaluate(resource []byte, expr string) {
	result, err := fhirpath.Evaluate(resource, expr)
	if err != nil {
		fmt.Printf("  %s => ERROR: %v\n", expr, err)
		return
	}
	fmt.Printf("  %s => %v\n", expr, formatResult(result))
}

// formatResult formats a FHIRPath result for display
func formatResult(result interface{}) string {
	switch v := result.(type) {
	case fhirpath.Collection:
		return formatCollection(v)
	case []interface{}:
		if len(v) == 0 {
			return "{ }"
		}
		formatted := make([]string, 0, len(v))
		for _, item := range v {
			formatted = append(formatted, formatValue(item))
		}
		if len(formatted) == 1 {
			return formatted[0]
		}
		return "{ " + joinStrings(formatted, ", ") + " }"
	default:
		return formatValue(v)
	}
}

// formatCollection formats a FHIRPath collection
func formatCollection(c fhirpath.Collection) string {
	if len(c) == 0 {
		return "{ }"
	}
	formatted := make([]string, 0, len(c))
	for _, item := range c {
		formatted = append(formatted, formatValue(item))
	}
	if len(formatted) == 1 {
		return formatted[0]
	}
	return "{ " + joinStrings(formatted, ", ") + " }"
}

// formatValue formats a single value
func formatValue(v interface{}) string {
	// Check for ObjectValue (complex FHIR types)
	if obj, ok := v.(interface{ Data() []byte }); ok {
		return formatObjectValue(obj.Data())
	}

	// Check for String() method
	if stringer, ok := v.(interface{ String() string }); ok {
		s := stringer.String()
		if s == "true" || s == "false" {
			return s
		}
		if s != "" && s[0] != '{' {
			if _, isStr := v.(interface{ Value() string }); isStr {
				return "'" + s + "'"
			}
			return s
		}
	}

	switch val := v.(type) {
	case bool:
		return fmt.Sprintf("%t", val)
	case string:
		return fmt.Sprintf("'%s'", val)
	case int, int64, int32:
		return fmt.Sprintf("%d", val)
	case float64, float32:
		return fmt.Sprintf("%v", val)
	case map[string]interface{}:
		return formatComplexType(val)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		s := string(data)
		if len(s) > 80 {
			return s[:77] + "..."
		}
		return s
	}
}

// formatObjectValue formats a FHIRPath ObjectValue from JSON
func formatObjectValue(data []byte) string {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return "{...}"
	}
	return formatComplexType(obj)
}

// formatComplexType formats FHIR complex datatypes for display
func formatComplexType(val map[string]interface{}) string {
	if s := formatHumanName(val); s != "" {
		return s
	}
	if s := formatContactPointOrIdentifier(val); s != "" {
		return s
	}
	if s := formatAddress(val); s != "" {
		return s
	}
	if s := formatCoding(val); s != "" {
		return s
	}
	if s := formatCodeableConcept(val); s != "" {
		return s
	}
	if s := formatQuantity(val); s != "" {
		return s
	}
	if s := formatReference(val); s != "" {
		return s
	}
	return formatGeneric(val)
}

func formatHumanName(val map[string]interface{}) string {
	family, ok := val["family"].(string)
	if !ok {
		return ""
	}
	if given, ok := val["given"].([]interface{}); ok && len(given) > 0 {
		return fmt.Sprintf("HumanName('%s, %v')", family, given[0])
	}
	return fmt.Sprintf("HumanName('%s')", family)
}

func formatContactPointOrIdentifier(val map[string]interface{}) string {
	system, ok := val["system"].(string)
	if !ok {
		return ""
	}
	value, ok := val["value"].(string)
	if !ok {
		return ""
	}
	contactSystems := map[string]bool{
		"phone": true, "fax": true, "email": true,
		"pager": true, "url": true, "sms": true, "other": true,
	}
	if contactSystems[system] {
		return fmt.Sprintf("ContactPoint(%s: '%s')", system, value)
	}
	return fmt.Sprintf("Identifier('%s')", value)
}

func formatAddress(val map[string]interface{}) string {
	city, ok := val["city"].(string)
	if !ok {
		return ""
	}
	if state, ok := val["state"].(string); ok {
		return fmt.Sprintf("Address('%s, %s')", city, state)
	}
	return fmt.Sprintf("Address('%s')", city)
}

func formatCoding(val map[string]interface{}) string {
	code, ok := val["code"].(string)
	if !ok {
		return ""
	}
	if display, ok := val["display"].(string); ok {
		return fmt.Sprintf("Coding('%s': '%s')", code, display)
	}
	return fmt.Sprintf("Coding('%s')", code)
}

func formatCodeableConcept(val map[string]interface{}) string {
	coding, ok := val["coding"].([]interface{})
	if !ok || len(coding) == 0 {
		return ""
	}
	c, ok := coding[0].(map[string]interface{})
	if !ok {
		return ""
	}
	code, ok := c["code"].(string)
	if !ok {
		return ""
	}
	return fmt.Sprintf("CodeableConcept('%s')", code)
}

func formatQuantity(val map[string]interface{}) string {
	value, ok := val["value"].(float64)
	if !ok {
		return ""
	}
	if unit, ok := val["unit"].(string); ok {
		return fmt.Sprintf("Quantity(%v '%s')", value, unit)
	}
	return fmt.Sprintf("Quantity(%v)", value)
}

func formatReference(val map[string]interface{}) string {
	ref, ok := val["reference"].(string)
	if !ok {
		return ""
	}
	return fmt.Sprintf("Reference('%s')", ref)
}

func formatGeneric(val map[string]interface{}) string {
	keys := make([]string, 0, 3)
	for k := range val {
		keys = append(keys, k)
		if len(keys) >= 3 {
			break
		}
	}
	return fmt.Sprintf("{%s...}", joinStrings(keys, ", "))
}

// joinStrings joins strings with separator
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
