package validator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRegistryBasicOperations(t *testing.T) {
	reg := NewRegistry(FHIRVersionR4)

	// Register a simple StructureDef
	sd := &StructureDef{
		URL:  "http://example.org/fhir/StructureDefinition/TestResource",
		Name: "TestResource",
		Type: "TestResource",
		Kind: "resource",
	}

	err := reg.Register(sd)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Test Get
	ctx := context.Background()
	retrieved, err := reg.Get(ctx, sd.URL)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if retrieved.Name != sd.Name {
		t.Errorf("Expected name %s, got %s", sd.Name, retrieved.Name)
	}

	// Test GetByType
	retrieved, err = reg.GetByType(ctx, "TestResource")
	if err != nil {
		t.Fatalf("GetByType failed: %v", err)
	}
	if retrieved.URL != sd.URL {
		t.Errorf("Expected URL %s, got %s", sd.URL, retrieved.URL)
	}

	// Test List
	urls, err := reg.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(urls) != 1 {
		t.Errorf("Expected 1 URL, got %d", len(urls))
	}

	// Test Size
	if reg.Size() != 1 {
		t.Errorf("Expected size 1, got %d", reg.Size())
	}
}

func TestRegistryNotFound(t *testing.T) {
	reg := NewRegistry(FHIRVersionR4)
	ctx := context.Background()

	_, err := reg.Get(ctx, "http://nonexistent.org/sd")
	if err == nil {
		t.Error("Expected error for non-existent URL")
	}

	_, err = reg.GetByType(ctx, "NonExistentType")
	if err == nil {
		t.Error("Expected error for non-existent type")
	}
}

func TestParseStructureDefinition(t *testing.T) {
	json := `{
		"resourceType": "StructureDefinition",
		"url": "http://hl7.org/fhir/StructureDefinition/Patient",
		"name": "Patient",
		"type": "Patient",
		"kind": "resource",
		"abstract": false,
		"baseDefinition": "http://hl7.org/fhir/StructureDefinition/DomainResource",
		"snapshot": {
			"element": [
				{
					"id": "Patient",
					"path": "Patient",
					"min": 0,
					"max": "*"
				},
				{
					"id": "Patient.id",
					"path": "Patient.id",
					"min": 0,
					"max": "1",
					"type": [{"code": "id"}]
				},
				{
					"id": "Patient.name",
					"path": "Patient.name",
					"min": 0,
					"max": "*",
					"type": [{"code": "HumanName"}],
					"constraint": [
						{
							"key": "ele-1",
							"severity": "error",
							"human": "All FHIR elements must have a @value or children",
							"expression": "hasValue() or (children().count() > id.count())"
						}
					]
				},
				{
					"id": "Patient.gender",
					"path": "Patient.gender",
					"min": 0,
					"max": "1",
					"type": [{"code": "code"}],
					"binding": {
						"strength": "required",
						"valueSet": "http://hl7.org/fhir/ValueSet/administrative-gender"
					}
				}
			]
		}
	}`

	sd, err := ParseStructureDefinition([]byte(json))
	if err != nil {
		t.Fatalf("ParseStructureDefinition failed: %v", err)
	}

	if sd.URL != "http://hl7.org/fhir/StructureDefinition/Patient" {
		t.Errorf("Wrong URL: %s", sd.URL)
	}
	if sd.Name != "Patient" {
		t.Errorf("Wrong Name: %s", sd.Name)
	}
	if sd.Type != "Patient" {
		t.Errorf("Wrong Type: %s", sd.Type)
	}
	if sd.Kind != "resource" {
		t.Errorf("Wrong Kind: %s", sd.Kind)
	}
	if sd.Abstract {
		t.Error("Should not be abstract")
	}

	// Check snapshot elements
	if len(sd.Snapshot) != 4 {
		t.Errorf("Expected 4 snapshot elements, got %d", len(sd.Snapshot))
	}

	// Check Patient.name element
	var nameElem *ElementDef
	for i := range sd.Snapshot {
		if sd.Snapshot[i].Path == "Patient.name" {
			nameElem = &sd.Snapshot[i]
			break
		}
	}
	if nameElem == nil {
		t.Fatal("Patient.name element not found")
	}
	if nameElem.Max != "*" {
		t.Errorf("Expected max *, got %s", nameElem.Max)
	}
	if len(nameElem.Constraints) != 1 {
		t.Errorf("Expected 1 constraint, got %d", len(nameElem.Constraints))
	}
	if nameElem.Constraints[0].Key != "ele-1" {
		t.Errorf("Wrong constraint key: %s", nameElem.Constraints[0].Key)
	}

	// Check Patient.gender element with binding
	var genderElem *ElementDef
	for i := range sd.Snapshot {
		if sd.Snapshot[i].Path == "Patient.gender" {
			genderElem = &sd.Snapshot[i]
			break
		}
	}
	if genderElem == nil {
		t.Fatal("Patient.gender element not found")
	}
	if genderElem.Binding == nil {
		t.Fatal("Patient.gender should have binding")
	}
	if genderElem.Binding.Strength != "required" {
		t.Errorf("Wrong binding strength: %s", genderElem.Binding.Strength)
	}
}

func TestLoadFromBundle(t *testing.T) {
	bundle := `{
		"resourceType": "Bundle",
		"entry": [
			{
				"resource": {
					"resourceType": "StructureDefinition",
					"url": "http://example.org/sd/Resource1",
					"name": "Resource1",
					"type": "Resource1",
					"kind": "resource"
				}
			},
			{
				"resource": {
					"resourceType": "StructureDefinition",
					"url": "http://example.org/sd/Resource2",
					"name": "Resource2",
					"type": "Resource2",
					"kind": "resource"
				}
			},
			{
				"resource": {
					"resourceType": "ValueSet",
					"url": "http://example.org/vs/SomeValueSet"
				}
			}
		]
	}`

	reg := NewRegistry(FHIRVersionR4)
	count, err := reg.LoadFromBundle([]byte(bundle))
	if err != nil {
		t.Fatalf("LoadFromBundle failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 StructureDefinitions loaded, got %d", count)
	}

	if reg.Size() != 2 {
		t.Errorf("Expected registry size 2, got %d", reg.Size())
	}
}

func TestLoadFromSpecsR4(t *testing.T) {
	// Find the specs directory
	specsPath := filepath.Join("..", "..", "specs", "r4", "profiles-resources.json")

	// Check if file exists
	if _, err := os.Stat(specsPath); os.IsNotExist(err) {
		t.Skip("Specs file not found, skipping integration test")
	}

	reg := NewRegistry(FHIRVersionR4)
	count, err := reg.LoadFromFile(specsPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	t.Logf("Loaded %d StructureDefinitions from profiles-resources.json", count)

	if count == 0 {
		t.Error("Expected to load some StructureDefinitions")
	}

	// Test that we can get Patient
	ctx := context.Background()
	patient, err := reg.GetByType(ctx, "Patient")
	if err != nil {
		t.Fatalf("Failed to get Patient: %v", err)
	}

	if patient.Name != "Patient" {
		t.Errorf("Wrong name: %s", patient.Name)
	}

	// Check Patient has snapshot elements
	if len(patient.Snapshot) == 0 {
		t.Error("Patient should have snapshot elements")
	}

	t.Logf("Patient has %d snapshot elements", len(patient.Snapshot))

	// Verify some known Patient elements exist
	elements := make(map[string]bool)
	for _, elem := range patient.Snapshot {
		elements[elem.Path] = true
	}

	expectedPaths := []string{
		"Patient",
		"Patient.id",
		"Patient.name",
		"Patient.gender",
		"Patient.birthDate",
	}

	for _, path := range expectedPaths {
		if !elements[path] {
			t.Errorf("Missing expected element: %s", path)
		}
	}
}

func TestLoadFromSpecsTypes(t *testing.T) {
	specsPath := filepath.Join("..", "..", "specs", "r4", "profiles-types.json")

	if _, err := os.Stat(specsPath); os.IsNotExist(err) {
		t.Skip("Specs file not found, skipping integration test")
	}

	reg := NewRegistry(FHIRVersionR4)
	count, err := reg.LoadFromFile(specsPath)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	t.Logf("Loaded %d StructureDefinitions from profiles-types.json", count)

	if count == 0 {
		t.Error("Expected to load some StructureDefinitions")
	}

	// Test that we can get primitive types
	ctx := context.Background()
	stringSD, err := reg.Get(ctx, "http://hl7.org/fhir/StructureDefinition/string")
	if err != nil {
		t.Fatalf("Failed to get string type: %v", err)
	}

	if stringSD.Kind != "primitive-type" {
		t.Errorf("Expected primitive-type, got %s", stringSD.Kind)
	}
}

func TestValidationResult(t *testing.T) {
	result := NewValidationResult()

	if !result.Valid {
		t.Error("New result should be valid")
	}
	if result.HasErrors() {
		t.Error("New result should have no errors")
	}

	// Add warning
	result.AddIssue(ValidationIssue{
		Severity:    SeverityWarning,
		Code:        IssueCodeValue,
		Diagnostics: "This is a warning",
	})

	if !result.Valid {
		t.Error("Result with only warnings should still be valid")
	}
	if result.HasErrors() {
		t.Error("Should not have errors")
	}
	if !result.HasWarnings() {
		t.Error("Should have warnings")
	}
	if result.WarningCount() != 1 {
		t.Errorf("Expected 1 warning, got %d", result.WarningCount())
	}

	// Add error
	result.AddIssue(ValidationIssue{
		Severity:    SeverityError,
		Code:        IssueCodeRequired,
		Diagnostics: "Required field missing",
		Expression:  []string{"Patient.name"},
	})

	if result.Valid {
		t.Error("Result with error should not be valid")
	}
	if !result.HasErrors() {
		t.Error("Should have errors")
	}
	if result.ErrorCount() != 1 {
		t.Errorf("Expected 1 error, got %d", result.ErrorCount())
	}
}

func TestValidationResultMerge(t *testing.T) {
	r1 := NewValidationResult()
	r1.AddIssue(ValidationIssue{
		Severity: SeverityWarning,
		Code:     IssueCodeValue,
	})

	r2 := NewValidationResult()
	r2.AddIssue(ValidationIssue{
		Severity: SeverityError,
		Code:     IssueCodeRequired,
	})

	r1.Merge(r2)

	if len(r1.Issues) != 2 {
		t.Errorf("Expected 2 issues after merge, got %d", len(r1.Issues))
	}
	if r1.Valid {
		t.Error("Merged result should not be valid (has error)")
	}
}
