package validator

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseReference(t *testing.T) {
	tests := []struct {
		name        string
		ref         string
		wantValid   bool
		wantType    string
		wantResType string
		wantID      string
		wantVersion string
	}{
		// Valid relative references
		{
			name:        "relative reference",
			ref:         "Patient/123",
			wantValid:   true,
			wantType:    RefTypeRelative,
			wantResType: "Patient",
			wantID:      "123",
		},
		{
			name:        "relative reference with dashes",
			ref:         "Observation/obs-123-abc",
			wantValid:   true,
			wantType:    RefTypeRelative,
			wantResType: "Observation",
			wantID:      "obs-123-abc",
		},
		{
			name:        "relative reference with dots",
			ref:         "Patient/123.456",
			wantValid:   true,
			wantType:    RefTypeRelative,
			wantResType: "Patient",
			wantID:      "123.456",
		},

		// Valid contained references
		{
			name:      "contained reference",
			ref:       "#med1",
			wantValid: true,
			wantType:  RefTypeContained,
			wantID:    "med1",
		},
		{
			name:      "contained reference with dashes",
			ref:       "#medication-123",
			wantValid: true,
			wantType:  RefTypeContained,
			wantID:    "medication-123",
		},

		// Valid absolute references
		{
			name:        "absolute reference http",
			ref:         "http://example.org/fhir/Patient/123",
			wantValid:   true,
			wantType:    RefTypeAbsolute,
			wantResType: "Patient",
			wantID:      "123",
		},
		{
			name:        "absolute reference https",
			ref:         "https://example.org/fhir/r4/Observation/obs-456",
			wantValid:   true,
			wantType:    RefTypeAbsolute,
			wantResType: "Observation",
			wantID:      "obs-456",
		},

		// Valid URN references
		{
			name:      "urn:uuid reference",
			ref:       "urn:uuid:550e8400-e29b-41d4-a716-446655440000",
			wantValid: true,
			wantType:  RefTypeUrnUUID,
			wantID:    "550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:      "urn:oid reference",
			ref:       "urn:oid:2.16.840.1.113883.4.642.1.1",
			wantValid: true,
			wantType:  RefTypeUrnOID,
			wantID:    "2.16.840.1.113883.4.642.1.1",
		},

		// Valid canonical references
		// Note: StructureDefinition URLs that end with ResourceType/name
		// are parsed as absolute references (which is correct - they can be both)
		{
			name:        "canonical reference matching absolute pattern",
			ref:         "http://hl7.org/fhir/StructureDefinition/Patient",
			wantValid:   true,
			wantType:    RefTypeAbsolute, // Matches absolute pattern first
			wantResType: "StructureDefinition",
			wantID:      "Patient",
		},
		{
			name:        "canonical reference with version",
			ref:         "http://hl7.org/fhir/StructureDefinition/Patient|4.0.1",
			wantValid:   true,
			wantType:    RefTypeCanonical, // Version suffix makes it canonical
			wantVersion: "4.0.1",
		},
		{
			name:      "canonical reference - ValueSet URL",
			ref:       "http://hl7.org/fhir/ValueSet/administrative-gender",
			wantValid: true,
			wantType:  RefTypeAbsolute, // Also matches absolute pattern
		},
		{
			name:      "canonical reference - no resource pattern",
			ref:       "http://example.org/custom/profile",
			wantValid: true,
			wantType:  RefTypeCanonical, // Does not match absolute pattern
		},

		// Invalid references
		{
			name:      "empty reference",
			ref:       "",
			wantValid: false,
			wantType:  RefTypeUnknown,
		},
		{
			name:      "invalid format - just text",
			ref:       "invalid",
			wantValid: false,
			wantType:  RefTypeUnknown,
		},
		{
			name:      "invalid urn:uuid - wrong format",
			ref:       "urn:uuid:invalid-uuid",
			wantValid: false,
			wantType:  RefTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseReference(tt.ref)

			assert.Equal(t, tt.wantValid, result.Valid, "Valid mismatch")
			assert.Equal(t, tt.wantType, result.Type, "Type mismatch")
			assert.Equal(t, tt.ref, result.Raw, "Raw mismatch")

			if tt.wantResType != "" {
				assert.Equal(t, tt.wantResType, result.ResourceType, "ResourceType mismatch")
			}
			if tt.wantID != "" {
				assert.Equal(t, tt.wantID, result.ID, "ID mismatch")
			}
			if tt.wantVersion != "" {
				assert.Equal(t, tt.wantVersion, result.Version, "Version mismatch")
			}
		})
	}
}

func TestValidateReferences_ContainedResources(t *testing.T) {
	registry := NewRegistry(FHIRVersionR4)

	// Load minimal StructureDefinitions for testing
	err := loadTestStructureDefinitions(registry)
	if err != nil {
		t.Skipf("Skipping test - could not load specs: %v", err)
	}

	opts := ValidatorOptions{
		ValidateConstraints: false,
		ValidateReferences:  true,
	}
	v := NewValidator(registry, opts)

	tests := []struct {
		name          string
		resource      []byte
		wantRefErrors int
	}{
		{
			name: "valid contained reference",
			resource: []byte(`{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "test"},
				"contained": [
					{"resourceType": "Patient", "id": "pat1"}
				],
				"subject": {"reference": "#pat1"}
			}`),
			wantRefErrors: 0,
		},
		{
			name: "invalid contained reference - not found",
			resource: []byte(`{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "test"},
				"contained": [
					{"resourceType": "Patient", "id": "pat1"}
				],
				"subject": {"reference": "#patXXX"}
			}`),
			wantRefErrors: 1,
		},
		{
			name: "invalid reference format",
			resource: []byte(`{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "test"},
				"subject": {"reference": "invalid-format"}
			}`),
			wantRefErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(context.Background(), tt.resource)
			require.NoError(t, err)

			// Count reference-related errors (errors with "reference" in expression)
			refErrors := 0
			for _, issue := range result.Issues {
				if issue.Severity == SeverityError {
					if len(issue.Expression) > 0 && containsString(issue.Expression[0], "reference") {
						refErrors++
					}
				}
			}
			assert.Equal(t, tt.wantRefErrors, refErrors, "Reference error count mismatch. Issues: %v", result.Issues)
		})
	}
}

func TestValidateReferences_RelativeReferences(t *testing.T) {
	registry := NewRegistry(FHIRVersionR4)

	err := loadTestStructureDefinitions(registry)
	if err != nil {
		t.Skipf("Skipping test - could not load specs: %v", err)
	}

	opts := ValidatorOptions{
		ValidateConstraints: false,
		ValidateReferences:  true,
	}
	v := NewValidator(registry, opts)

	tests := []struct {
		name      string
		resource  []byte
		wantValid bool
	}{
		{
			name: "valid relative reference",
			resource: []byte(`{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "test"},
				"subject": {"reference": "Patient/123"}
			}`),
			wantValid: true,
		},
		{
			name: "valid absolute reference",
			resource: []byte(`{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "test"},
				"subject": {"reference": "https://example.org/fhir/Patient/123"}
			}`),
			wantValid: true,
		},
		{
			name: "valid urn:uuid reference",
			resource: []byte(`{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "test"},
				"subject": {"reference": "urn:uuid:550e8400-e29b-41d4-a716-446655440000"}
			}`),
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(context.Background(), tt.resource)
			require.NoError(t, err)

			if tt.wantValid {
				// Check that there are no reference-related errors
				for _, issue := range result.Issues {
					if len(issue.Expression) > 0 && containsString(issue.Expression[0], "reference") {
						t.Errorf("Unexpected reference error: %v", issue)
					}
				}
			}
		})
	}
}

func TestExtractResourceTypeFromProfile(t *testing.T) {
	tests := []struct {
		profile  string
		expected string
	}{
		{
			profile:  "http://hl7.org/fhir/StructureDefinition/Patient",
			expected: "Patient",
		},
		{
			profile:  "http://hl7.org/fhir/StructureDefinition/Observation|4.0.1",
			expected: "Observation",
		},
		{
			profile:  "https://example.org/fhir/StructureDefinition/MyProfile",
			expected: "MyProfile",
		},
		{
			profile:  "Patient",
			expected: "Patient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.profile, func(t *testing.T) {
			result := extractResourceTypeFromProfile(tt.profile)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPathWithoutArrayIndices(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Patient.contact[0].reference",
			expected: "Patient.contact.reference",
		},
		{
			input:    "Bundle.entry[5].resource.subject[0].reference",
			expected: "Bundle.entry.resource.subject.reference",
		},
		{
			input:    "Patient.name",
			expected: "Patient.name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := pathWithoutArrayIndices(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to load test StructureDefinitions
func loadTestStructureDefinitions(registry *Registry) error {
	// Try to load from specs directory
	_, err := registry.LoadFromFile("../../specs/r4/profiles-resources.json")
	if err != nil {
		return err
	}
	// Also load types (Address, Identifier, etc.)
	_, err = registry.LoadFromFile("../../specs/r4/profiles-types.json")
	return err
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
