//nolint:misspell // FHIR R4 uses British spelling "cancelled" for ObservationStatus
package validator

import (
	"context"
	"testing"
)

// TestLocalTerminologyService tests the local terminology service.
func TestLocalTerminologyService(t *testing.T) {
	// Create a simple bundle with CodeSystem and ValueSet
	bundle := []byte(`{
		"resourceType": "Bundle",
		"entry": [
			{
				"resource": {
					"resourceType": "CodeSystem",
					"url": "http://hl7.org/fhir/administrative-gender",
					"name": "AdministrativeGender",
					"status": "active",
					"content": "complete",
					"concept": [
						{"code": "male", "display": "Male"},
						{"code": "female", "display": "Female"},
						{"code": "other", "display": "Other"},
						{"code": "unknown", "display": "Unknown"}
					]
				}
			},
			{
				"resource": {
					"resourceType": "ValueSet",
					"url": "http://hl7.org/fhir/ValueSet/administrative-gender",
					"name": "AdministrativeGender",
					"status": "active",
					"compose": {
						"include": [
							{"system": "http://hl7.org/fhir/administrative-gender"}
						]
					}
				}
			},
			{
				"resource": {
					"resourceType": "CodeSystem",
					"url": "http://hl7.org/fhir/observation-status",
					"name": "ObservationStatus",
					"status": "active",
					"content": "complete",
					"concept": [
						{"code": "registered", "display": "Registered"},
						{"code": "preliminary", "display": "Preliminary"},
						{"code": "final", "display": "Final"},
						{"code": "amended", "display": "Amended"},
						{"code": "corrected", "display": "Corrected"},
						{"code": "cancelled", "display": "Cancelled"},
						{"code": "entered-in-error", "display": "Entered in Error"},
						{"code": "unknown", "display": "Unknown"}
					]
				}
			},
			{
				"resource": {
					"resourceType": "ValueSet",
					"url": "http://hl7.org/fhir/ValueSet/observation-status",
					"name": "ObservationStatus",
					"status": "active",
					"compose": {
						"include": [
							{"system": "http://hl7.org/fhir/observation-status"}
						]
					}
				}
			}
		]
	}`)

	svc := NewLocalTerminologyService()
	if err := svc.LoadFromBundle(bundle); err != nil {
		t.Fatalf("Failed to load bundle: %v", err)
	}

	// Check stats
	codeSystems, valueSets, totalCodes := svc.Stats()
	if codeSystems != 2 {
		t.Errorf("Expected 2 code systems, got %d", codeSystems)
	}
	if valueSets != 2 {
		t.Errorf("Expected 2 value sets, got %d", valueSets)
	}
	if totalCodes != 12 { // 4 gender + 8 observation-status
		t.Errorf("Expected 12 total codes, got %d", totalCodes)
	}

	ctx := context.Background()

	// Test ValidateCode
	tests := []struct {
		name      string
		system    string
		code      string
		valueSet  string
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "valid gender code - male",
			system:    "http://hl7.org/fhir/administrative-gender",
			code:      "male",
			valueSet:  "http://hl7.org/fhir/ValueSet/administrative-gender",
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "valid gender code - female",
			system:    "http://hl7.org/fhir/administrative-gender",
			code:      "female",
			valueSet:  "http://hl7.org/fhir/ValueSet/administrative-gender",
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "invalid gender code",
			system:    "http://hl7.org/fhir/administrative-gender",
			code:      "invalid-code",
			valueSet:  "http://hl7.org/fhir/ValueSet/administrative-gender",
			wantValid: false,
			wantErr:   false,
		},
		{
			name:      "valid code without system",
			system:    "",
			code:      "male",
			valueSet:  "http://hl7.org/fhir/ValueSet/administrative-gender",
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "valid observation status",
			system:    "http://hl7.org/fhir/observation-status",
			code:      "final",
			valueSet:  "http://hl7.org/fhir/ValueSet/observation-status",
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "unknown valueset",
			system:    "http://hl7.org/fhir/administrative-gender",
			code:      "male",
			valueSet:  "http://hl7.org/fhir/ValueSet/unknown",
			wantValid: false,
			wantErr:   true,
		},
		{
			name:      "versioned valueset URL",
			system:    "http://hl7.org/fhir/administrative-gender",
			code:      "male",
			valueSet:  "http://hl7.org/fhir/ValueSet/administrative-gender|4.0.1",
			wantValid: true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := svc.ValidateCode(ctx, tt.system, tt.code, tt.valueSet)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if valid != tt.wantValid {
				t.Errorf("ValidateCode() valid = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

// TestLocalTerminologyServiceExpandValueSet tests ValueSet expansion.
func TestLocalTerminologyServiceExpandValueSet(t *testing.T) {
	bundle := []byte(`{
		"resourceType": "Bundle",
		"entry": [
			{
				"resource": {
					"resourceType": "CodeSystem",
					"url": "http://example.org/status",
					"content": "complete",
					"concept": [
						{"code": "active", "display": "Active"},
						{"code": "inactive", "display": "Inactive"}
					]
				}
			},
			{
				"resource": {
					"resourceType": "ValueSet",
					"url": "http://example.org/ValueSet/status",
					"compose": {
						"include": [{"system": "http://example.org/status"}]
					}
				}
			}
		]
	}`)

	svc := NewLocalTerminologyService()
	if err := svc.LoadFromBundle(bundle); err != nil {
		t.Fatalf("Failed to load bundle: %v", err)
	}

	ctx := context.Background()

	// Test expansion
	codes, err := svc.ExpandValueSet(ctx, "http://example.org/ValueSet/status")
	if err != nil {
		t.Fatalf("ExpandValueSet() error = %v", err)
	}

	if len(codes) != 2 {
		t.Errorf("Expected 2 codes, got %d", len(codes))
	}

	// Verify codes
	foundActive, foundInactive := false, false
	for _, c := range codes {
		if c.Code == "active" {
			foundActive = true
		}
		if c.Code == "inactive" {
			foundInactive = true
		}
	}
	if !foundActive || !foundInactive {
		t.Errorf("Expected active and inactive codes, got %+v", codes)
	}
}

// TestLocalTerminologyServiceLookupCode tests code lookup.
func TestLocalTerminologyServiceLookupCode(t *testing.T) {
	bundle := []byte(`{
		"resourceType": "Bundle",
		"entry": [
			{
				"resource": {
					"resourceType": "CodeSystem",
					"url": "http://example.org/codes",
					"content": "complete",
					"concept": [
						{"code": "A", "display": "Alpha"},
						{"code": "B", "display": "Beta"}
					]
				}
			}
		]
	}`)

	svc := NewLocalTerminologyService()
	if err := svc.LoadFromBundle(bundle); err != nil {
		t.Fatalf("Failed to load bundle: %v", err)
	}

	ctx := context.Background()

	// Test lookup
	info, err := svc.LookupCode(ctx, "http://example.org/codes", "A")
	if err != nil {
		t.Fatalf("LookupCode() error = %v", err)
	}

	if info == nil {
		t.Fatal("LookupCode() returned nil")
	}

	if info.Code != "A" || info.Display != "Alpha" {
		t.Errorf("Expected code A/Alpha, got %s/%s", info.Code, info.Display)
	}

	// Test non-existent code
	info, err = svc.LookupCode(ctx, "http://example.org/codes", "Z")
	if err != nil {
		t.Fatalf("LookupCode() error = %v", err)
	}
	if info != nil {
		t.Errorf("Expected nil for non-existent code, got %+v", info)
	}

	// Test non-existent system
	_, err = svc.LookupCode(ctx, "http://example.org/unknown", "A")
	if err == nil {
		t.Error("Expected error for unknown system")
	}
}

// TestLocalTerminologyServiceNestedConcepts tests hierarchical CodeSystems.
func TestLocalTerminologyServiceNestedConcepts(t *testing.T) {
	bundle := []byte(`{
		"resourceType": "Bundle",
		"entry": [
			{
				"resource": {
					"resourceType": "CodeSystem",
					"url": "http://example.org/hierarchy",
					"content": "complete",
					"concept": [
						{
							"code": "parent",
							"display": "Parent",
							"concept": [
								{"code": "child1", "display": "Child 1"},
								{
									"code": "child2",
									"display": "Child 2",
									"concept": [
										{"code": "grandchild", "display": "Grandchild"}
									]
								}
							]
						}
					]
				}
			},
			{
				"resource": {
					"resourceType": "ValueSet",
					"url": "http://example.org/ValueSet/hierarchy",
					"compose": {
						"include": [{"system": "http://example.org/hierarchy"}]
					}
				}
			}
		]
	}`)

	svc := NewLocalTerminologyService()
	if err := svc.LoadFromBundle(bundle); err != nil {
		t.Fatalf("Failed to load bundle: %v", err)
	}

	ctx := context.Background()

	// All codes should be flattened
	codes, err := svc.ExpandValueSet(ctx, "http://example.org/ValueSet/hierarchy")
	if err != nil {
		t.Fatalf("ExpandValueSet() error = %v", err)
	}

	if len(codes) != 4 {
		t.Errorf("Expected 4 codes (parent + 2 children + grandchild), got %d", len(codes))
	}

	// Verify grandchild is included
	valid, _ := svc.ValidateCode(ctx, "http://example.org/hierarchy", "grandchild", "http://example.org/ValueSet/hierarchy")
	if !valid {
		t.Error("Expected grandchild code to be valid")
	}
}

// TestLocalTerminologyServiceExplicitConcepts tests ValueSets with explicit concepts.
func TestLocalTerminologyServiceExplicitConcepts(t *testing.T) {
	bundle := []byte(`{
		"resourceType": "Bundle",
		"entry": [
			{
				"resource": {
					"resourceType": "ValueSet",
					"url": "http://example.org/ValueSet/explicit",
					"compose": {
						"include": [
							{
								"system": "http://example.org/codes",
								"concept": [
									{"code": "X", "display": "X-Ray"},
									{"code": "Y", "display": "Yankee"}
								]
							}
						]
					}
				}
			}
		]
	}`)

	svc := NewLocalTerminologyService()
	if err := svc.LoadFromBundle(bundle); err != nil {
		t.Fatalf("Failed to load bundle: %v", err)
	}

	ctx := context.Background()

	// Only explicitly listed codes should be valid
	valid, _ := svc.ValidateCode(ctx, "http://example.org/codes", "X", "http://example.org/ValueSet/explicit")
	if !valid {
		t.Error("Expected X to be valid")
	}

	valid, _ = svc.ValidateCode(ctx, "http://example.org/codes", "Y", "http://example.org/ValueSet/explicit")
	if !valid {
		t.Error("Expected Y to be valid")
	}

	valid, _ = svc.ValidateCode(ctx, "http://example.org/codes", "Z", "http://example.org/ValueSet/explicit")
	if valid {
		t.Error("Expected Z to be invalid (not in explicit list)")
	}
}

// TestLocalTerminologyServiceHasValueSet tests HasValueSet method.
func TestLocalTerminologyServiceHasValueSet(t *testing.T) {
	bundle := []byte(`{
		"resourceType": "Bundle",
		"entry": [
			{
				"resource": {
					"resourceType": "ValueSet",
					"url": "http://example.org/ValueSet/test",
					"compose": {
						"include": [
							{
								"system": "http://example.org/codes",
								"concept": [{"code": "A"}]
							}
						]
					}
				}
			}
		]
	}`)

	svc := NewLocalTerminologyService()
	if err := svc.LoadFromBundle(bundle); err != nil {
		t.Fatalf("Failed to load bundle: %v", err)
	}

	if !svc.HasValueSet("http://example.org/ValueSet/test") {
		t.Error("Expected HasValueSet to return true")
	}

	if !svc.HasValueSet("http://example.org/ValueSet/test|1.0.0") {
		t.Error("Expected HasValueSet to return true for versioned URL")
	}

	if svc.HasValueSet("http://example.org/ValueSet/unknown") {
		t.Error("Expected HasValueSet to return false for unknown ValueSet")
	}
}

// TestTerminologyValidationIntegration tests terminology validation in the validator.
func TestTerminologyValidationIntegration(t *testing.T) {
	// Create a minimal StructureDefinition with binding
	sd := &StructureDef{
		URL:  "http://hl7.org/fhir/StructureDefinition/Patient",
		Name: "Patient",
		Type: "Patient",
		Kind: "resource",
		Snapshot: []ElementDef{
			{Path: "Patient", Min: 0, Max: "*"},
			{Path: "Patient.id", Min: 0, Max: "1"},
			{
				Path:  "Patient.gender",
				Min:   0,
				Max:   "1",
				Types: []TypeRef{{Code: "code"}},
				Binding: &ElementBinding{
					Strength: "required",
					ValueSet: "http://hl7.org/fhir/ValueSet/administrative-gender",
				},
			},
		},
	}

	// Create mock registry
	registry := &mockRegistry{sds: map[string]*StructureDef{
		"Patient": sd,
	}}

	// Create terminology service
	termBundle := []byte(`{
		"resourceType": "Bundle",
		"entry": [
			{
				"resource": {
					"resourceType": "CodeSystem",
					"url": "http://hl7.org/fhir/administrative-gender",
					"content": "complete",
					"concept": [
						{"code": "male"},
						{"code": "female"},
						{"code": "other"},
						{"code": "unknown"}
					]
				}
			},
			{
				"resource": {
					"resourceType": "ValueSet",
					"url": "http://hl7.org/fhir/ValueSet/administrative-gender",
					"compose": {
						"include": [{"system": "http://hl7.org/fhir/administrative-gender"}]
					}
				}
			}
		]
	}`)

	termService := NewLocalTerminologyService()
	if err := termService.LoadFromBundle(termBundle); err != nil {
		t.Fatalf("Failed to load terminology: %v", err)
	}

	// Create validator with terminology enabled
	opts := ValidatorOptions{
		ValidateTerminology: true,
	}
	validator := NewValidator(registry, opts).WithTerminologyService(termService)

	ctx := context.Background()

	// Test valid patient
	validPatient := []byte(`{
		"resourceType": "Patient",
		"gender": "male"
	}`)

	result, err := validator.Validate(ctx, validPatient)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid patient, got issues: %+v", result.Issues)
	}

	// Test invalid patient (wrong gender code)
	invalidPatient := []byte(`{
		"resourceType": "Patient",
		"gender": "invalid-gender"
	}`)

	result, err = validator.Validate(ctx, invalidPatient)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid patient (wrong gender code)")
	}

	// Check for terminology error
	foundTermError := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeCodeInvalid {
			foundTermError = true
			break
		}
	}
	if !foundTermError {
		t.Error("Expected terminology validation error")
	}
}

// mockRegistry is a simple mock for testing.
type mockRegistry struct {
	sds map[string]*StructureDef
}

func (m *mockRegistry) Get(_ context.Context, url string) (*StructureDef, error) {
	if sd, ok := m.sds[url]; ok {
		return sd, nil
	}
	return nil, nil
}

func (m *mockRegistry) GetByType(_ context.Context, resourceType string) (*StructureDef, error) {
	if sd, ok := m.sds[resourceType]; ok {
		return sd, nil
	}
	return nil, nil
}

func (m *mockRegistry) List(_ context.Context) ([]string, error) {
	urls := make([]string, 0, len(m.sds))
	for url := range m.sds {
		urls = append(urls, url)
	}
	return urls, nil
}

// TestEmbeddedTerminologyService tests the embedded terminology service.
func TestEmbeddedTerminologyService(t *testing.T) {
	svc := NewEmbeddedTerminologyServiceR4()
	ctx := context.Background()

	// Test that administrative-gender is available (most common required binding)
	if !svc.HasValueSet("http://hl7.org/fhir/ValueSet/administrative-gender") {
		t.Error("Expected administrative-gender ValueSet to be embedded")
	}

	// Test validation of valid code
	valid, err := svc.ValidateCode(ctx, "", "male", "http://hl7.org/fhir/ValueSet/administrative-gender")
	if err != nil {
		t.Errorf("ValidateCode() error = %v", err)
	}
	if !valid {
		t.Error("Expected 'male' to be valid in administrative-gender")
	}

	// Test validation of invalid code
	valid, err = svc.ValidateCode(ctx, "", "invalid-code", "http://hl7.org/fhir/ValueSet/administrative-gender")
	if err != nil {
		t.Errorf("ValidateCode() error = %v", err)
	}
	if valid {
		t.Error("Expected 'invalid-code' to be invalid in administrative-gender")
	}

	// Test versioned URL
	valid, err = svc.ValidateCode(ctx, "", "female", "http://hl7.org/fhir/ValueSet/administrative-gender|4.0.1")
	if err != nil {
		t.Errorf("ValidateCode() error = %v", err)
	}
	if !valid {
		t.Error("Expected 'female' to be valid with versioned URL")
	}

	// Test observation-status
	valid, err = svc.ValidateCode(ctx, "", "final", "http://hl7.org/fhir/ValueSet/observation-status")
	if err != nil {
		t.Errorf("ValidateCode() error = %v", err)
	}
	if !valid {
		t.Error("Expected 'final' to be valid in observation-status")
	}

	// Test ExpandValueSet
	codes, err := svc.ExpandValueSet(ctx, "http://hl7.org/fhir/ValueSet/administrative-gender")
	if err != nil {
		t.Errorf("ExpandValueSet() error = %v", err)
	}
	if len(codes) != 4 { // male, female, other, unknown
		t.Errorf("Expected 4 codes in administrative-gender, got %d", len(codes))
	}

	// Test Stats
	valueSets, totalCodes := svc.Stats()
	if valueSets == 0 {
		t.Error("Expected at least one embedded ValueSet")
	}
	if totalCodes == 0 {
		t.Error("Expected at least one embedded code")
	}
	t.Logf("Embedded terminology: %d ValueSets, %d total codes", valueSets, totalCodes)
}

// TestEmbeddedTerminologyServiceVersions tests all FHIR versions.
func TestEmbeddedTerminologyServiceVersions(t *testing.T) {
	tests := []struct {
		name        string
		constructor func() *EmbeddedTerminologyService
		version     string
	}{
		{"R4", NewEmbeddedTerminologyServiceR4, "4.0.1"},
		{"R4B", NewEmbeddedTerminologyServiceR4B, "4.3.0"},
		{"R5", NewEmbeddedTerminologyServiceR5, "5.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.constructor()

			if svc.FHIRVersion() != tt.version {
				t.Errorf("Expected FHIR version %s, got %s", tt.version, svc.FHIRVersion())
			}

			// administrative-gender should be available in all versions
			if !svc.HasValueSet("http://hl7.org/fhir/ValueSet/administrative-gender") {
				t.Error("Expected administrative-gender ValueSet to be embedded")
			}

			valueSets, totalCodes := svc.Stats()
			if valueSets == 0 {
				t.Error("Expected at least one embedded ValueSet")
			}
			t.Logf("%s: %d ValueSets, %d total codes", tt.name, valueSets, totalCodes)
		})
	}
}

// TestAvailableEmbeddedVersions tests the version listing function.
func TestAvailableEmbeddedVersions(t *testing.T) {
	versions := AvailableEmbeddedVersions()
	if len(versions) < 3 {
		t.Errorf("Expected at least 3 embedded versions, got %d: %v", len(versions), versions)
	}
	t.Logf("Available embedded versions: %v", versions)
}

// TestValidatorOptionsTerminology tests terminology via ValidatorOptions.
func TestValidatorOptionsTerminology(t *testing.T) {
	ctx := context.Background()

	// Create Patient StructureDef with gender binding
	patientWithBinding := &StructureDef{
		URL:  "http://hl7.org/fhir/StructureDefinition/Patient",
		Name: "Patient",
		Type: "Patient",
		Kind: "resource",
		Snapshot: []ElementDef{
			{Path: "Patient", Min: 0, Max: "*"},
			{Path: "Patient.id", Min: 0, Max: "1"},
			{
				Path:  "Patient.gender",
				Min:   0,
				Max:   "1",
				Types: []TypeRef{{Code: "code"}},
				Binding: &ElementBinding{
					Strength: "required",
					ValueSet: "http://hl7.org/fhir/ValueSet/administrative-gender",
				},
			},
		},
	}

	registry := &mockRegistry{
		sds: map[string]*StructureDef{
			"Patient": patientWithBinding,
		},
	}

	tests := []struct {
		name        string
		opts        ValidatorOptions
		patient     string
		expectValid bool
	}{
		{
			name: "ValidateTerminology=true uses embedded R4 by default",
			opts: ValidatorOptions{
				ValidateTerminology: true,
			},
			patient:     `{"resourceType": "Patient", "gender": "male"}`,
			expectValid: true,
		},
		{
			name: "Invalid code should produce error",
			opts: ValidatorOptions{
				ValidateTerminology: true,
			},
			patient:     `{"resourceType": "Patient", "gender": "invalid-code"}`,
			expectValid: false,
		},
		{
			name: "Explicit R4 terminology service",
			opts: ValidatorOptions{
				ValidateTerminology: true,
				TerminologyService:  TerminologyEmbeddedR4,
			},
			patient:     `{"resourceType": "Patient", "gender": "female"}`,
			expectValid: true,
		},
		{
			name: "Explicit R4B terminology service",
			opts: ValidatorOptions{
				ValidateTerminology: true,
				TerminologyService:  TerminologyEmbeddedR4B,
			},
			patient:     `{"resourceType": "Patient", "gender": "other"}`,
			expectValid: true,
		},
		{
			name: "Explicit R5 terminology service",
			opts: ValidatorOptions{
				ValidateTerminology: true,
				TerminologyService:  TerminologyEmbeddedR5,
			},
			patient:     `{"resourceType": "Patient", "gender": "unknown"}`,
			expectValid: true,
		},
		{
			name: "ValidateTerminology=false skips validation",
			opts: ValidatorOptions{
				ValidateTerminology: false,
			},
			patient:     `{"resourceType": "Patient", "gender": "invalid-code"}`,
			expectValid: true, // No terminology validation, so valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(registry, tt.opts)
			result, err := v.Validate(ctx, []byte(tt.patient))
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}

			if result.Valid != tt.expectValid {
				t.Errorf("Expected Valid=%v, got Valid=%v. Issues: %v",
					tt.expectValid, result.Valid, result.Issues)
			}
		})
	}
}
