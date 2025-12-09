package validator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateExtensions_ValidSimpleExtension(t *testing.T) {
	registry := NewRegistry(FHIRVersionR4)

	err := loadTestStructureDefinitions(registry)
	if err != nil {
		t.Skipf("Skipping test - could not load specs: %v", err)
	}

	opts := ValidatorOptions{
		ValidateConstraints: false,
		ValidateExtensions:  true,
	}
	v := NewValidator(registry, opts)

	// Patient with a valid simple extension
	resource := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"extension": [
			{
				"url": "http://example.org/fhir/StructureDefinition/patient-importance",
				"valueCode": "VIP"
			}
		]
	}`)

	result, err := v.Validate(context.Background(), resource)
	require.NoError(t, err)

	// Count extension-related errors
	extErrors := countExtensionErrors(result)
	assert.Equal(t, 0, extErrors, "Should not have extension errors. Issues: %v", result.Issues)
}

func TestValidateExtensions_MissingURL(t *testing.T) {
	registry := NewRegistry(FHIRVersionR4)

	err := loadTestStructureDefinitions(registry)
	if err != nil {
		t.Skipf("Skipping test - could not load specs: %v", err)
	}

	opts := ValidatorOptions{
		ValidateConstraints: false,
		ValidateExtensions:  true,
	}
	v := NewValidator(registry, opts)

	// Extension without URL
	resource := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"extension": [
			{
				"valueString": "test value"
			}
		]
	}`)

	result, err := v.Validate(context.Background(), resource)
	require.NoError(t, err)

	// Should have error for missing URL
	extErrors := countExtensionErrors(result)
	assert.GreaterOrEqual(t, extErrors, 1, "Should have at least one extension error for missing URL")
}

func TestValidateExtensions_MissingValue(t *testing.T) {
	registry := NewRegistry(FHIRVersionR4)

	err := loadTestStructureDefinitions(registry)
	if err != nil {
		t.Skipf("Skipping test - could not load specs: %v", err)
	}

	opts := ValidatorOptions{
		ValidateConstraints: false,
		ValidateExtensions:  true,
	}
	v := NewValidator(registry, opts)

	// Extension without value and without nested extensions
	resource := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"extension": [
			{
				"url": "http://example.org/fhir/StructureDefinition/some-extension"
			}
		]
	}`)

	result, err := v.Validate(context.Background(), resource)
	require.NoError(t, err)

	// Should have error for missing value
	extErrors := countExtensionErrors(result)
	assert.GreaterOrEqual(t, extErrors, 1, "Should have extension error for missing value")
}

func TestValidateExtensions_ComplexExtension(t *testing.T) {
	registry := NewRegistry(FHIRVersionR4)

	err := loadTestStructureDefinitions(registry)
	if err != nil {
		t.Skipf("Skipping test - could not load specs: %v", err)
	}

	opts := ValidatorOptions{
		ValidateConstraints: false,
		ValidateExtensions:  true,
	}
	v := NewValidator(registry, opts)

	// Complex extension with nested extensions
	resource := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"extension": [
			{
				"url": "http://example.org/fhir/StructureDefinition/patient-address-details",
				"extension": [
					{
						"url": "latitude",
						"valueDecimal": 40.7128
					},
					{
						"url": "longitude",
						"valueDecimal": -74.0060
					}
				]
			}
		]
	}`)

	result, err := v.Validate(context.Background(), resource)
	require.NoError(t, err)

	// Should not have extension errors for valid complex extension
	extErrors := countExtensionErrors(result)
	assert.Equal(t, 0, extErrors, "Should not have extension errors. Issues: %v", result.Issues)
}

func TestValidateExtensions_ValueAndNestedExtensions(t *testing.T) {
	registry := NewRegistry(FHIRVersionR4)

	err := loadTestStructureDefinitions(registry)
	if err != nil {
		t.Skipf("Skipping test - could not load specs: %v", err)
	}

	opts := ValidatorOptions{
		ValidateConstraints: false,
		ValidateExtensions:  true,
	}
	v := NewValidator(registry, opts)

	// Invalid: has both value and nested extensions
	resource := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"extension": [
			{
				"url": "http://example.org/fhir/StructureDefinition/invalid-extension",
				"valueString": "some value",
				"extension": [
					{
						"url": "nested",
						"valueCode": "test"
					}
				]
			}
		]
	}`)

	result, err := v.Validate(context.Background(), resource)
	require.NoError(t, err)

	// Should have error for having both value and nested extensions
	extErrors := countExtensionErrors(result)
	assert.GreaterOrEqual(t, extErrors, 1, "Should have extension error for both value and nested extensions")
}

func TestValidateExtensions_ModifierExtension(t *testing.T) {
	registry := NewRegistry(FHIRVersionR4)

	err := loadTestStructureDefinitions(registry)
	if err != nil {
		t.Skipf("Skipping test - could not load specs: %v", err)
	}

	opts := ValidatorOptions{
		ValidateConstraints: false,
		ValidateExtensions:  true,
	}
	v := NewValidator(registry, opts)

	// Valid modifier extension
	resource := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"modifierExtension": [
			{
				"url": "http://example.org/fhir/StructureDefinition/patient-confidential",
				"valueBoolean": true
			}
		]
	}`)

	result, err := v.Validate(context.Background(), resource)
	require.NoError(t, err)

	// Should not have extension errors
	extErrors := countExtensionErrors(result)
	assert.Equal(t, 0, extErrors, "Should not have extension errors. Issues: %v", result.Issues)
}

func TestValidateExtensions_NestedInElement(t *testing.T) {
	registry := NewRegistry(FHIRVersionR4)

	err := loadTestStructureDefinitions(registry)
	if err != nil {
		t.Skipf("Skipping test - could not load specs: %v", err)
	}

	opts := ValidatorOptions{
		ValidateConstraints: false,
		ValidateExtensions:  true,
	}
	v := NewValidator(registry, opts)

	// Extension nested inside a complex element
	resource := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"name": [
			{
				"family": "Smith",
				"extension": [
					{
						"url": "http://example.org/fhir/StructureDefinition/name-pronunciation",
						"valueString": "smith"
					}
				]
			}
		]
	}`)

	result, err := v.Validate(context.Background(), resource)
	require.NoError(t, err)

	// Should not have extension errors
	extErrors := countExtensionErrors(result)
	assert.Equal(t, 0, extErrors, "Should not have extension errors. Issues: %v", result.Issues)
}

func TestIsValidExtensionURL(t *testing.T) {
	tests := []struct {
		url   string
		valid bool
	}{
		{"http://example.org/fhir/StructureDefinition/my-extension", true},
		{"https://example.org/fhir/StructureDefinition/my-extension", true},
		{"http://hl7.org/fhir/StructureDefinition/patient-birthPlace", true},
		{"urn:uuid:550e8400-e29b-41d4-a716-446655440000", true},
		{"urn:oid:2.16.840.1.113883.4.642.1.1", true},
		{"latitude", true},     // Valid simple name for nested extensions
		{"my-extension", true}, // Valid simple name with hyphen
		{"extension_1", true},  // Valid simple name with underscore
		{"", false},            // Empty is invalid
		{"has space", false},   // Spaces not allowed
		{"has/slash", false},   // Slashes not allowed in simple names
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isValidExtensionURL(tt.url)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestIsHL7Extension(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"http://hl7.org/fhir/StructureDefinition/patient-birthPlace", true},
		{"http://hl7.org/fhir/StructureDefinition/data-absent-reason", true},
		{"http://example.org/fhir/StructureDefinition/my-extension", false},
		{"https://example.org/extension", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := IsHL7Extension(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractExtensionName(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"http://hl7.org/fhir/StructureDefinition/patient-birthPlace", "patient-birthPlace"},
		{"http://hl7.org/fhir/StructureDefinition/data-absent-reason", "data-absent-reason"},
		{"http://example.org/fhir/StructureDefinition/my-extension", "my-extension"},
		{"https://example.org/custom/profile", "profile"},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := ExtractExtensionName(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasExtensionValue(t *testing.T) {
	tests := []struct {
		name     string
		ext      map[string]interface{}
		expected bool
	}{
		{
			name:     "has valueString",
			ext:      map[string]interface{}{"url": "http://example.org", "valueString": "test"},
			expected: true,
		},
		{
			name:     "has valueCode",
			ext:      map[string]interface{}{"url": "http://example.org", "valueCode": "active"},
			expected: true,
		},
		{
			name:     "has valueBoolean",
			ext:      map[string]interface{}{"url": "http://example.org", "valueBoolean": true},
			expected: true,
		},
		{
			name:     "no value",
			ext:      map[string]interface{}{"url": "http://example.org"},
			expected: false,
		},
		{
			name:     "has nested extensions only",
			ext:      map[string]interface{}{"url": "http://example.org", "extension": []interface{}{}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasExtensionValue(tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetExtensionValueType(t *testing.T) {
	tests := []struct {
		name     string
		ext      map[string]interface{}
		expected string
	}{
		{
			name:     "valueString",
			ext:      map[string]interface{}{"url": "http://example.org", "valueString": "test"},
			expected: "String",
		},
		{
			name:     "valueCode",
			ext:      map[string]interface{}{"url": "http://example.org", "valueCode": "active"},
			expected: "Code",
		},
		{
			name:     "valueBoolean",
			ext:      map[string]interface{}{"url": "http://example.org", "valueBoolean": true},
			expected: "Boolean",
		},
		{
			name:     "valueQuantity",
			ext:      map[string]interface{}{"url": "http://example.org", "valueQuantity": map[string]interface{}{}},
			expected: "Quantity",
		},
		{
			name:     "no value",
			ext:      map[string]interface{}{"url": "http://example.org"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getExtensionValueType(tt.ext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// countExtensionErrors counts extension-related errors in the result
func countExtensionErrors(result *ValidationResult) int {
	count := 0
	for _, issue := range result.Issues {
		if issue.Severity == SeverityError {
			// Check if it's an extension-related error
			if issue.Code == IssueCodeExtension {
				count++
				continue
			}
			// Check if path contains "extension"
			if len(issue.Expression) > 0 && containsString(issue.Expression[0], "extension") {
				count++
			}
		}
	}
	return count
}
