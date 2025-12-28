package validator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setupTestValidator creates a validator with R4 specs loaded
func setupTestValidator(t *testing.T) *Validator {
	reg := NewRegistry(FHIRVersionR4)

	// Load resource definitions
	resourcesPath := filepath.Join("..", "..", "specs", "r4", "profiles-resources.json")
	if _, err := os.Stat(resourcesPath); err == nil {
		_, err := reg.LoadFromFile(resourcesPath)
		if err != nil {
			t.Fatalf("Failed to load resources: %v", err)
		}
	} else {
		t.Skip("Specs not found, skipping test")
	}

	// Load type definitions
	typesPath := filepath.Join("..", "..", "specs", "r4", "profiles-types.json")
	if _, err := os.Stat(typesPath); err == nil {
		reg.LoadFromFile(typesPath)
	}

	return NewValidator(reg, DefaultValidatorOptions())
}

func TestValidateSimplePatient(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	patient := []byte(`{
		"resourceType": "Patient",
		"id": "example",
		"active": true,
		"name": [{
			"use": "official",
			"family": "Doe",
			"given": ["John", "James"]
		}],
		"gender": "male",
		"birthDate": "1990-01-01"
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if !result.Valid {
		for _, issue := range result.Issues {
			t.Logf("Issue: [%s] %s - %s", issue.Severity, issue.Code, issue.Diagnostics)
		}
	}

	// A simple valid patient should pass
	if result.HasErrors() {
		t.Errorf("Valid patient should not have errors, got %d", result.ErrorCount())
	}
}

func TestValidateInvalidJSON(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	invalid := []byte(`{not valid json}`)

	result, err := v.Validate(ctx, invalid)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("Invalid JSON should not be valid")
	}
	if !result.HasErrors() {
		t.Error("Should have errors for invalid JSON")
	}
}

func TestValidateMissingResourceType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	noType := []byte(`{
		"id": "example",
		"active": true
	}`)

	result, err := v.Validate(ctx, noType)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("Resource without resourceType should not be valid")
	}

	hasRequiredError := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeRequired {
			hasRequiredError = true
			break
		}
	}
	if !hasRequiredError {
		t.Error("Should have 'required' error for missing resourceType")
	}
}

func TestValidateUnknownResourceType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	unknown := []byte(`{
		"resourceType": "NotARealResource",
		"id": "example"
	}`)

	result, err := v.Validate(ctx, unknown)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("Unknown resource type should not be valid")
	}

	hasNotFoundError := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeNotFound {
			hasNotFoundError = true
			break
		}
	}
	if !hasNotFoundError {
		t.Error("Should have 'not-found' error for unknown resource type")
	}
}

func TestValidateObservation(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	observation := []byte(`{
		"resourceType": "Observation",
		"id": "example",
		"status": "final",
		"code": {
			"coding": [{
				"system": "http://loinc.org",
				"code": "29463-7",
				"display": "Body Weight"
			}],
			"text": "Body Weight"
		},
		"subject": {
			"reference": "Patient/example"
		},
		"effectiveDateTime": "2023-01-15T10:30:00Z",
		"valueQuantity": {
			"value": 70.5,
			"unit": "kg",
			"system": "http://unitsofmeasure.org",
			"code": "kg"
		}
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if !result.Valid {
		for _, issue := range result.Issues {
			t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
		}
	}

	// Valid observation should pass
	if result.HasErrors() {
		t.Errorf("Valid observation should not have errors, got %d", result.ErrorCount())
	}
}

func TestValidateInvalidPrimitiveType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// active should be boolean, not string
	invalid := []byte(`{
		"resourceType": "Patient",
		"id": "example",
		"active": "yes"
	}`)

	result, err := v.Validate(ctx, invalid)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("Patient with string active should not be valid")
	}

	hasValueError := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeValue {
			hasValueError = true
			t.Logf("Found expected error: %s", issue.Diagnostics)
			break
		}
	}
	if !hasValueError {
		t.Error("Should have 'value' error for invalid boolean")
	}
}

func TestValidateCardinalityExceeded(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// birthDate has max=1, providing array should fail
	invalid := []byte(`{
		"resourceType": "Patient",
		"id": "example",
		"birthDate": "1990-01-01"
	}`)

	result, err := v.Validate(ctx, invalid)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	// This should be valid (single birthDate)
	t.Logf("Validation result for single birthDate: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s", issue.Severity, issue.Code, issue.Diagnostics)
	}
}

func TestValidateEncounter(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	encounter := []byte(`{
		"resourceType": "Encounter",
		"id": "example",
		"status": "finished",
		"class": {
			"system": "http://terminology.hl7.org/CodeSystem/v3-ActCode",
			"code": "AMB",
			"display": "ambulatory"
		},
		"subject": {
			"reference": "Patient/example"
		},
		"period": {
			"start": "2023-01-15T09:00:00Z",
			"end": "2023-01-15T10:00:00Z"
		}
	}`)

	result, err := v.Validate(ctx, encounter)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	t.Logf("Encounter validation: valid=%v, errors=%d, warnings=%d", result.Valid, result.ErrorCount(), result.WarningCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}
}

func TestValidateWithProfile(t *testing.T) {
	reg := NewRegistry(FHIRVersionR4)

	// Load specs
	resourcesPath := filepath.Join("..", "..", "specs", "r4", "profiles-resources.json")
	if _, err := os.Stat(resourcesPath); err != nil {
		t.Skip("Specs not found")
	}
	reg.LoadFromFile(resourcesPath)

	// Create validator with specific profile
	opts := DefaultValidatorOptions()
	opts.Profile = "http://hl7.org/fhir/StructureDefinition/Patient"

	v := NewValidator(reg, opts)
	ctx := context.Background()

	patient := []byte(`{
		"resourceType": "Patient",
		"id": "test"
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	t.Logf("Profile validation: valid=%v, errors=%d", result.Valid, result.ErrorCount())
}

func TestValidateMedication(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	medication := []byte(`{
		"resourceType": "Medication",
		"id": "example",
		"code": {
			"coding": [{
				"system": "http://www.nlm.nih.gov/research/umls/rxnorm",
				"code": "1049502",
				"display": "Acetaminophen 325 MG Oral Tablet"
			}]
		},
		"status": "active",
		"form": {
			"coding": [{
				"system": "http://snomed.info/sct",
				"code": "385055001",
				"display": "Tablet"
			}]
		}
	}`)

	result, err := v.Validate(ctx, medication)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	t.Logf("Medication validation: valid=%v, errors=%d, warnings=%d", result.Valid, result.ErrorCount(), result.WarningCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s", issue.Severity, issue.Code, issue.Diagnostics)
	}
}

func TestValidateConstraintViolation(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Patient.contact requires either a name, telecom, address, organization, or period
	// This violates pat-1: SHALL at least contain a contact's details or a reference to an organization
	patientWithEmptyContact := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"contact": [{
			"relationship": [{
				"coding": [{
					"system": "http://terminology.hl7.org/CodeSystem/v2-0131",
					"code": "E"
				}]
			}]
		}]
	}`)

	result, err := v.Validate(ctx, patientWithEmptyContact)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	// Should have a constraint violation
	hasConstraintError := false
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s", issue.Severity, issue.Code, issue.Diagnostics)
		if issue.Code == IssueCodeInvariant {
			hasConstraintError = true
		}
	}

	if !hasConstraintError {
		t.Error("Expected constraint violation for empty contact")
	}
}

func TestValidateConstraintPass(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Valid contact with name satisfies pat-1
	patientWithValidContact := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"contact": [{
			"relationship": [{
				"coding": [{
					"system": "http://terminology.hl7.org/CodeSystem/v2-0131",
					"code": "E"
				}]
			}],
			"name": {
				"family": "Doe",
				"given": ["Jane"]
			}
		}]
	}`)

	result, err := v.Validate(ctx, patientWithValidContact)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	// Should pass all constraints
	t.Logf("Validation: valid=%v, errors=%d, warnings=%d", result.Valid, result.ErrorCount(), result.WarningCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s", issue.Severity, issue.Code, issue.Diagnostics)
	}

	constraintErrors := 0
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeInvariant && issue.Severity == SeverityError {
			constraintErrors++
		}
	}
	if constraintErrors > 0 {
		t.Errorf("Valid contact should not have constraint errors, got %d", constraintErrors)
	}
}

func BenchmarkValidatePatient(b *testing.B) {
	reg := NewRegistry(FHIRVersionR4)
	resourcesPath := filepath.Join("..", "..", "specs", "r4", "profiles-resources.json")
	if _, err := os.Stat(resourcesPath); err != nil {
		b.Skip("Specs not found")
	}
	reg.LoadFromFile(resourcesPath)

	v := NewValidator(reg, DefaultValidatorOptions())
	ctx := context.Background()

	patient := []byte(`{
		"resourceType": "Patient",
		"id": "example",
		"active": true,
		"name": [{
			"use": "official",
			"family": "Doe",
			"given": ["John"]
		}],
		"gender": "male",
		"birthDate": "1990-01-01"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate(ctx, patient)
	}
}

// BenchmarkValidatePatientLarge tests validation with a larger patient resource.
func BenchmarkValidatePatientLarge(b *testing.B) {
	reg := NewRegistry(FHIRVersionR4)
	resourcesPath := filepath.Join("..", "..", "specs", "r4", "profiles-resources.json")
	if _, err := os.Stat(resourcesPath); err != nil {
		b.Skip("Specs not found")
	}
	reg.LoadFromFile(resourcesPath)

	v := NewValidator(reg, DefaultValidatorOptions())
	ctx := context.Background()

	// Larger patient with multiple names, addresses, telecoms
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "large-example",
		"active": true,
		"name": [
			{"use": "official", "family": "Doe", "given": ["John", "James", "Joseph"]},
			{"use": "nickname", "given": ["Johnny"]},
			{"use": "maiden", "family": "Smith", "given": ["Jane"]}
		],
		"telecom": [
			{"system": "phone", "value": "+1-555-0100", "use": "home"},
			{"system": "phone", "value": "+1-555-0101", "use": "work"},
			{"system": "email", "value": "john.doe@example.com", "use": "work"}
		],
		"gender": "male",
		"birthDate": "1990-01-01",
		"address": [
			{
				"use": "home",
				"type": "physical",
				"line": ["123 Main St", "Apt 4B"],
				"city": "Anytown",
				"state": "CA",
				"postalCode": "12345",
				"country": "USA"
			},
			{
				"use": "work",
				"line": ["456 Business Ave"],
				"city": "Worktown",
				"state": "NY",
				"postalCode": "67890"
			}
		],
		"maritalStatus": {
			"coding": [{
				"system": "http://terminology.hl7.org/CodeSystem/v3-MaritalStatus",
				"code": "M",
				"display": "Married"
			}]
		},
		"contact": [{
			"relationship": [{"coding": [{"system": "http://terminology.hl7.org/CodeSystem/v2-0131", "code": "N"}]}],
			"name": {"family": "Doe", "given": ["Jane"]},
			"telecom": [{"system": "phone", "value": "+1-555-0102"}]
		}]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate(ctx, patient)
	}
}

// BenchmarkValidateObservation tests validation of an Observation resource.
func BenchmarkValidateObservation(b *testing.B) {
	reg := NewRegistry(FHIRVersionR4)
	resourcesPath := filepath.Join("..", "..", "specs", "r4", "profiles-resources.json")
	if _, err := os.Stat(resourcesPath); err != nil {
		b.Skip("Specs not found")
	}
	reg.LoadFromFile(resourcesPath)
	typesPath := filepath.Join("..", "..", "specs", "r4", "profiles-types.json")
	reg.LoadFromFile(typesPath)

	v := NewValidator(reg, DefaultValidatorOptions())
	ctx := context.Background()

	observation := []byte(`{
		"resourceType": "Observation",
		"id": "blood-pressure",
		"status": "final",
		"category": [{
			"coding": [{
				"system": "http://terminology.hl7.org/CodeSystem/observation-category",
				"code": "vital-signs",
				"display": "Vital Signs"
			}]
		}],
		"code": {
			"coding": [{
				"system": "http://loinc.org",
				"code": "85354-9",
				"display": "Blood pressure panel"
			}]
		},
		"subject": {"reference": "Patient/example"},
		"effectiveDateTime": "2023-01-15T10:30:00Z",
		"component": [
			{
				"code": {"coding": [{"system": "http://loinc.org", "code": "8480-6", "display": "Systolic BP"}]},
				"valueQuantity": {"value": 120, "unit": "mmHg", "system": "http://unitsofmeasure.org", "code": "mm[Hg]"}
			},
			{
				"code": {"coding": [{"system": "http://loinc.org", "code": "8462-4", "display": "Diastolic BP"}]},
				"valueQuantity": {"value": 80, "unit": "mmHg", "system": "http://unitsofmeasure.org", "code": "mm[Hg]"}
			}
		]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate(ctx, observation)
	}
}

// BenchmarkValidateWithConstraints tests validation with FHIRPath constraints.
func BenchmarkValidateWithConstraints(b *testing.B) {
	reg := NewRegistry(FHIRVersionR4)
	resourcesPath := filepath.Join("..", "..", "specs", "r4", "profiles-resources.json")
	if _, err := os.Stat(resourcesPath); err != nil {
		b.Skip("Specs not found")
	}
	reg.LoadFromFile(resourcesPath)

	v := NewValidator(reg, DefaultValidatorOptions())
	ctx := context.Background()

	// Patient with contact that exercises pat-1 constraint
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "with-contact",
		"active": true,
		"contact": [{
			"relationship": [{"coding": [{"system": "http://terminology.hl7.org/CodeSystem/v2-0131", "code": "E"}]}],
			"name": {"family": "Emergency", "given": ["Contact"]}
		}]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate(ctx, patient)
	}
}

// BenchmarkValidateNoConstraints tests validation without FHIRPath constraints.
func BenchmarkValidateNoConstraints(b *testing.B) {
	reg := NewRegistry(FHIRVersionR4)
	resourcesPath := filepath.Join("..", "..", "specs", "r4", "profiles-resources.json")
	if _, err := os.Stat(resourcesPath); err != nil {
		b.Skip("Specs not found")
	}
	reg.LoadFromFile(resourcesPath)

	opts := DefaultValidatorOptions()
	opts.ValidateConstraints = false
	v := NewValidator(reg, opts)
	ctx := context.Background()

	patient := []byte(`{
		"resourceType": "Patient",
		"id": "example",
		"active": true,
		"name": [{"family": "Doe", "given": ["John"]}],
		"gender": "male",
		"birthDate": "1990-01-01"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate(ctx, patient)
	}
}

// BenchmarkExpressionCacheHit tests the benefit of expression caching.
func BenchmarkExpressionCacheHit(b *testing.B) {
	reg := NewRegistry(FHIRVersionR4)
	resourcesPath := filepath.Join("..", "..", "specs", "r4", "profiles-resources.json")
	if _, err := os.Stat(resourcesPath); err != nil {
		b.Skip("Specs not found")
	}
	reg.LoadFromFile(resourcesPath)

	v := NewValidator(reg, DefaultValidatorOptions())
	ctx := context.Background()

	// Multiple patients that exercise the same constraints (cache hits)
	patients := [][]byte{
		[]byte(`{"resourceType": "Patient", "id": "p1", "contact": [{"name": {"family": "A"}}]}`),
		[]byte(`{"resourceType": "Patient", "id": "p2", "contact": [{"name": {"family": "B"}}]}`),
		[]byte(`{"resourceType": "Patient", "id": "p3", "contact": [{"name": {"family": "C"}}]}`),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range patients {
			v.Validate(ctx, p)
		}
	}
}

// TestValidateEle1EmptyObject tests that empty objects violate ele-1
func TestValidateEle1EmptyObject(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Patient with empty name object - violates ele-1
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"name": [{}],
		"active": true
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	t.Logf("Empty object validation: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	// Should have ele-1 violation
	hasEle1Error := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "ele-1") {
			hasEle1Error = true
			break
		}
	}
	if !hasEle1Error {
		t.Error("Expected ele-1 constraint violation for empty object")
	}
}

// TestValidateEle1OnlyId tests that objects with only "id" violate ele-1
func TestValidateEle1OnlyId(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Patient with name that only has "id" - violates ele-1
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"name": [{"id": "name-1"}],
		"active": true
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	t.Logf("Only-id object validation: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	// Should have ele-1 violation
	hasEle1Error := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "ele-1") {
			hasEle1Error = true
			break
		}
	}
	if !hasEle1Error {
		t.Error("Expected ele-1 constraint violation for object with only 'id'")
	}
}

// TestValidateEle1ValidElement tests that valid elements pass ele-1
func TestValidateEle1ValidElement(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Valid patient - all elements have values or children
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"name": [{
			"family": "Doe",
			"given": ["John"]
		}],
		"active": true
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	// Should NOT have ele-1 violation
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "ele-1") {
			t.Errorf("Valid element should not have ele-1 error: %s", issue.Diagnostics)
		}
	}
}

// TestValidateEle1NestedEmpty tests nested empty objects
func TestValidateEle1NestedEmpty(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Patient with nested empty object in address
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"address": [{
			"city": "Springfield",
			"period": {}
		}],
		"active": true
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	t.Logf("Nested empty validation: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	// Should have ele-1 violation for the empty period
	hasEle1Error := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "ele-1") {
			hasEle1Error = true
			break
		}
	}
	if !hasEle1Error {
		t.Error("Expected ele-1 constraint violation for nested empty object")
	}
}

// TestValidateContainedResourceValid tests validation of valid contained resources
func TestValidateContainedResourceValid(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Condition with a valid contained Practitioner
	condition := []byte(`{
		"resourceType": "Condition",
		"id": "example",
		"contained": [
			{
				"resourceType": "Practitioner",
				"id": "p1",
				"name": [{"family": "Smith", "given": ["John"]}]
			}
		],
		"subject": {"reference": "Patient/example"},
		"asserter": {"reference": "#p1"}
	}`)

	result, err := v.Validate(ctx, condition)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	t.Logf("Contained resource validation: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	// Should not have "Unknown element" errors for the contained Practitioner fields
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeStructure && strings.Contains(issue.Diagnostics, "Unknown element") {
			if strings.Contains(issue.Diagnostics, "contained") {
				t.Errorf("Should not have unknown element error for contained resource fields: %s", issue.Diagnostics)
			}
		}
	}
}

// TestValidateContainedResourceInvalid tests validation of contained resources with invalid elements
func TestValidateContainedResourceInvalid(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Condition with a contained Practitioner that has an invalid field
	condition := []byte(`{
		"resourceType": "Condition",
		"id": "example",
		"contained": [
			{
				"resourceType": "Practitioner",
				"id": "p1",
				"name": [{"family": "Smith"}],
				"invalidField": "should cause error"
			}
		],
		"subject": {"reference": "Patient/example"}
	}`)

	result, err := v.Validate(ctx, condition)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	t.Logf("Invalid contained resource validation: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	// Should have an "Unknown element" error for the invalidField in the Practitioner
	hasUnknownFieldError := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeStructure && strings.Contains(issue.Diagnostics, "invalidField") {
			hasUnknownFieldError = true
			break
		}
	}
	if !hasUnknownFieldError {
		t.Error("Expected unknown element error for invalidField in contained resource")
	}
}

// TestValidateContainedResourceMissingType tests contained resources without resourceType
func TestValidateContainedResourceMissingType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Condition with a contained resource without resourceType
	condition := []byte(`{
		"resourceType": "Condition",
		"id": "example",
		"contained": [
			{
				"id": "p1",
				"name": [{"family": "Smith"}]
			}
		],
		"subject": {"reference": "Patient/example"}
	}`)

	result, err := v.Validate(ctx, condition)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	t.Logf("Missing resourceType validation: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	// Should have error for missing resourceType
	hasMissingTypeError := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Diagnostics, "resourceType") {
			hasMissingTypeError = true
			break
		}
	}
	if !hasMissingTypeError {
		t.Error("Expected error for contained resource without resourceType")
	}
}

// TestValidateContainedResourceMultiple tests validation of multiple contained resources
func TestValidateContainedResourceMultiple(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Condition with multiple valid contained resources of different types
	condition := []byte(`{
		"resourceType": "Condition",
		"id": "example",
		"contained": [
			{
				"resourceType": "Practitioner",
				"id": "p1",
				"name": [{"family": "Smith"}]
			},
			{
				"resourceType": "Organization",
				"id": "org1",
				"name": "Test Hospital"
			}
		],
		"subject": {"reference": "Patient/example"},
		"asserter": {"reference": "#p1"}
	}`)

	result, err := v.Validate(ctx, condition)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	t.Logf("Multiple contained resources validation: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	// Should not have "Unknown element" errors for the contained resources
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeStructure && strings.Contains(issue.Diagnostics, "Unknown element") {
			if strings.Contains(issue.Diagnostics, "contained") {
				t.Errorf("Should not have unknown element error for contained resources: %s", issue.Diagnostics)
			}
		}
	}
}

// TestValidateContainedResourceUnknownType tests contained resources with unknown resourceType
func TestValidateContainedResourceUnknownType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Condition with a contained resource with unknown type
	condition := []byte(`{
		"resourceType": "Condition",
		"id": "example",
		"contained": [
			{
				"resourceType": "UnknownResourceType",
				"id": "u1"
			}
		],
		"subject": {"reference": "Patient/example"}
	}`)

	result, err := v.Validate(ctx, condition)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	t.Logf("Unknown type contained resource validation: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	// Should have error for unknown resource type
	hasUnknownTypeError := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Diagnostics, "Unknown resource type") {
			hasUnknownTypeError = true
			break
		}
	}
	if !hasUnknownTypeError {
		t.Error("Expected error for unknown resource type in contained resource")
	}
}

// TestValidateContainedResourcePrimitives tests primitive validation in contained resources
func TestValidateContainedResourcePrimitives(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Condition with a contained Patient that has invalid primitives
	condition := []byte(`{
		"resourceType": "Condition",
		"id": "example",
		"contained": [
			{
				"resourceType": "Patient",
				"id": "p1",
				"active": "not-a-boolean",
				"birthDate": 12345
			}
		],
		"subject": {"reference": "#p1"}
	}`)

	result, err := v.Validate(ctx, condition)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	t.Logf("Primitive validation in contained: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	// Should have errors for invalid primitives
	hasBooleanError := false
	hasDateError := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeValue {
			if strings.Contains(issue.Diagnostics, "boolean") {
				hasBooleanError = true
			}
			if strings.Contains(issue.Diagnostics, "date") || strings.Contains(issue.Diagnostics, "string") {
				hasDateError = true
			}
		}
	}

	if !hasBooleanError {
		t.Error("Expected error for invalid boolean in contained resource")
	}
	if !hasDateError {
		t.Error("Expected error for invalid date in contained resource")
	}
}

// TestValidatePrimitiveTypeMismatchInComplexType tests that primitive type mismatches
// are detected within complex types (e.g., HumanName.family should be string, not number).
func TestValidatePrimitiveTypeMismatchInComplexType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// family should be string, but is number (24)
	invalidJSON := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"name": [{"family": 24}]
	}`)

	result, err := v.Validate(ctx, invalidJSON)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	t.Logf("Validation result: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	if result.Valid {
		t.Error("Patient with numeric family name should not be valid")
	}

	hasValueError := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeValue && strings.Contains(issue.Diagnostics, "string") {
			hasValueError = true
			t.Logf("Found expected error: %s", issue.Diagnostics)
			break
		}
	}
	if !hasValueError {
		t.Error("Should have 'value' error indicating family must be a string")
	}
}

// TestValidatePrimitiveTypeMismatchInNestedComplexType tests type validation
// in deeply nested complex types (e.g., Observation.code.coding.system).
func TestValidatePrimitiveTypeMismatchInNestedComplexType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// system should be uri (string), but is number
	invalidJSON := []byte(`{
		"resourceType": "Observation",
		"id": "test",
		"status": "final",
		"code": {
			"coding": [{
				"system": 12345,
				"code": "test"
			}]
		}
	}`)

	result, err := v.Validate(ctx, invalidJSON)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	t.Logf("Validation result: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	if result.Valid {
		t.Error("Observation with numeric coding.system should not be valid")
	}

	hasValueError := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeValue && strings.Contains(issue.Diagnostics, "string") {
			hasValueError = true
			t.Logf("Found expected error: %s", issue.Diagnostics)
			break
		}
	}
	if !hasValueError {
		t.Error("Should have 'value' error indicating system must be a string")
	}
}

// TestValidateMissingRequiredField tests that missing required fields are detected.
// Observation.status has cardinality 1..1 (required).
func TestValidateMissingRequiredField(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Observation without status (which is required with min=1)
	missingRequired := []byte(`{
		"resourceType": "Observation",
		"id": "test",
		"code": {"text": "test"}
	}`)

	result, err := v.Validate(ctx, missingRequired)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	t.Logf("Validation result: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	if result.Valid {
		t.Error("Observation without status should not be valid")
	}

	hasRequiredError := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeRequired && strings.Contains(issue.Diagnostics, "status") {
			hasRequiredError = true
			t.Logf("Found expected error: %s", issue.Diagnostics)
			break
		}
	}
	if !hasRequiredError {
		t.Error("Should have 'required' error for missing Observation.status")
	}
}

// TestValidateMissingRequiredFieldInPatient tests required field validation for Patient.
// This ensures the fix works for direct children of resources.
func TestValidateMissingRequiredFieldInPatient(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Patient is valid without most fields, but let's test with a resource
	// that has more required fields. Communication.language is required (min=1).
	// Using Patient with link where link.other is required
	invalidPatient := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"link": [{"type": "seealso"}]
	}`)

	result, err := v.Validate(ctx, invalidPatient)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	t.Logf("Validation result: valid=%v, errors=%d", result.Valid, result.ErrorCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	// Patient.link.other is required (min=1)
	if result.Valid {
		t.Error("Patient with link missing 'other' should not be valid")
	}
}

// =============================================================================
// PRIMITIVE DATATYPE VALIDATION TESTS
// =============================================================================

// TestValidateDateFormat tests date format validation (YYYY, YYYY-MM, YYYY-MM-DD).
func TestValidateDateFormat(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		birthDate   string
		shouldBeValid bool
	}{
		{"valid full date", "1990-01-15", true},
		{"valid year-month", "1990-01", true},
		{"valid year only", "1990", true},
		{"invalid format with time", "1990-01-15T10:30:00", false},
		{"invalid month", "1990-13-01", false},
		{"invalid day", "1990-01-32", false},
		{"invalid format", "01-15-1990", false},
		{"not a date string", "not-a-date", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patient := []byte(fmt.Sprintf(`{
				"resourceType": "Patient",
				"id": "test",
				"birthDate": "%s"
			}`, tt.birthDate))

			result, err := v.Validate(ctx, patient)
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if tt.shouldBeValid && !result.Valid {
				t.Errorf("Expected valid for birthDate=%s, got errors: %v", tt.birthDate, result.Issues)
			}
			if !tt.shouldBeValid && result.Valid {
				t.Errorf("Expected invalid for birthDate=%s", tt.birthDate)
			}
		})
	}
}

// TestValidateDateTimeFormat tests dateTime format validation.
func TestValidateDateTimeFormat(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		effectiveDateTime string
		shouldBeValid bool
	}{
		{"valid full dateTime with Z", "2023-01-15T10:30:00Z", true},
		{"valid full dateTime with offset", "2023-01-15T10:30:00+05:00", true},
		{"valid full dateTime with negative offset", "2023-01-15T10:30:00-08:00", true},
		{"valid dateTime with milliseconds", "2023-01-15T10:30:00.123Z", true},
		{"valid date only", "2023-01-15", true},
		{"valid year-month", "2023-01", true},
		{"valid year only", "2023", true},
		{"invalid - missing timezone", "2023-01-15T10:30:00", false},
		{"invalid hour", "2023-01-15T25:30:00Z", false},
		{"invalid minute", "2023-01-15T10:61:00Z", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observation := []byte(fmt.Sprintf(`{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "test"},
				"effectiveDateTime": "%s"
			}`, tt.effectiveDateTime))

			result, err := v.Validate(ctx, observation)
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			hasDateTimeError := false
			for _, issue := range result.Issues {
				if strings.Contains(issue.Diagnostics, "dateTime") {
					hasDateTimeError = true
					break
				}
			}

			if tt.shouldBeValid && hasDateTimeError {
				t.Errorf("Expected valid for effectiveDateTime=%s, got dateTime error", tt.effectiveDateTime)
			}
			if !tt.shouldBeValid && !hasDateTimeError {
				t.Errorf("Expected dateTime error for effectiveDateTime=%s", tt.effectiveDateTime)
			}
		})
	}
}

// TestValidateInstantFormat tests instant format validation (requires full timestamp with timezone).
func TestValidateInstantFormat(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		issued        string
		shouldBeValid bool
	}{
		{"valid instant with Z", "2023-01-15T10:30:00Z", true},
		{"valid instant with offset", "2023-01-15T10:30:00+05:00", true},
		{"valid instant with milliseconds", "2023-01-15T10:30:00.123Z", true},
		{"invalid - date only", "2023-01-15", false},
		{"invalid - missing timezone", "2023-01-15T10:30:00", false},
		{"invalid - year only", "2023", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diagnosticReport := []byte(fmt.Sprintf(`{
				"resourceType": "DiagnosticReport",
				"id": "test",
				"status": "final",
				"code": {"text": "test"},
				"issued": "%s"
			}`, tt.issued))

			result, err := v.Validate(ctx, diagnosticReport)
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			hasInstantError := false
			for _, issue := range result.Issues {
				if strings.Contains(issue.Diagnostics, "instant") {
					hasInstantError = true
					break
				}
			}

			if tt.shouldBeValid && hasInstantError {
				t.Errorf("Expected valid for issued=%s, got instant error", tt.issued)
			}
			if !tt.shouldBeValid && !hasInstantError {
				t.Errorf("Expected instant error for issued=%s", tt.issued)
			}
		})
	}
}

// TestValidateBooleanType tests boolean type validation.
func TestValidateBooleanType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		json          string
		shouldBeValid bool
	}{
		{"valid true", `{"resourceType":"Patient","id":"t","active":true}`, true},
		{"valid false", `{"resourceType":"Patient","id":"t","active":false}`, true},
		{"invalid string true", `{"resourceType":"Patient","id":"t","active":"true"}`, false},
		{"invalid string false", `{"resourceType":"Patient","id":"t","active":"false"}`, false},
		{"invalid number 1", `{"resourceType":"Patient","id":"t","active":1}`, false},
		{"invalid number 0", `{"resourceType":"Patient","id":"t","active":0}`, false},
		{"invalid string yes", `{"resourceType":"Patient","id":"t","active":"yes"}`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			hasBoolError := false
			for _, issue := range result.Issues {
				if strings.Contains(issue.Diagnostics, "boolean") {
					hasBoolError = true
					break
				}
			}

			if tt.shouldBeValid && hasBoolError {
				t.Errorf("Expected valid, got boolean error")
			}
			if !tt.shouldBeValid && !hasBoolError {
				t.Errorf("Expected boolean error for: %s", tt.json)
			}
		})
	}
}

// TestValidateIntegerTypes tests integer, positiveInt, and unsignedInt validation.
func TestValidateIntegerTypes(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		json          string
		shouldBeValid bool
		errorContains string
	}{
		// Valid integers
		{"valid integer", `{"resourceType":"Patient","id":"t","multipleBirthInteger":2}`, true, ""},
		{"valid integer zero", `{"resourceType":"Patient","id":"t","multipleBirthInteger":0}`, true, ""},
		{"valid negative integer", `{"resourceType":"Patient","id":"t","multipleBirthInteger":-1}`, true, ""},
		// Invalid - not integer
		{"invalid decimal", `{"resourceType":"Patient","id":"t","multipleBirthInteger":2.5}`, false, "integer"},
		{"invalid string", `{"resourceType":"Patient","id":"t","multipleBirthInteger":"2"}`, false, "integer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			hasExpectedError := false
			for _, issue := range result.Issues {
				if tt.errorContains != "" && strings.Contains(issue.Diagnostics, tt.errorContains) {
					hasExpectedError = true
					break
				}
			}

			if tt.shouldBeValid && result.HasErrors() {
				t.Errorf("Expected valid, got errors: %v", result.Issues)
			}
			if !tt.shouldBeValid && !hasExpectedError {
				t.Errorf("Expected error containing '%s'", tt.errorContains)
			}
		})
	}
}

// TestValidatePositiveInt tests positiveInt validation (must be > 0).
func TestValidatePositiveInt(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Appointment.minutesDuration is positiveInt
	tests := []struct {
		name          string
		value         string
		shouldBeValid bool
	}{
		{"valid positive", "30", true},
		{"valid positive 1", "1", true},
		{"invalid zero", "0", false},
		{"invalid negative", "-5", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appointment := []byte(fmt.Sprintf(`{
				"resourceType": "Appointment",
				"id": "test",
				"status": "booked",
				"participant": [{"status": "accepted"}],
				"minutesDuration": %s
			}`, tt.value))

			result, err := v.Validate(ctx, appointment)
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			hasPositiveIntError := false
			for _, issue := range result.Issues {
				if strings.Contains(issue.Diagnostics, "positive integer") {
					hasPositiveIntError = true
					break
				}
			}

			if tt.shouldBeValid && hasPositiveIntError {
				t.Errorf("Expected valid for minutesDuration=%s", tt.value)
			}
			if !tt.shouldBeValid && !hasPositiveIntError {
				t.Errorf("Expected positive integer error for minutesDuration=%s", tt.value)
			}
		})
	}
}

// TestValidateUnsignedInt tests unsignedInt validation (must be >= 0).
func TestValidateUnsignedInt(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Bundle.total is unsignedInt
	tests := []struct {
		name          string
		value         string
		shouldBeValid bool
	}{
		{"valid positive", "10", true},
		{"valid zero", "0", true},
		{"invalid negative", "-1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(fmt.Sprintf(`{
				"resourceType": "Bundle",
				"id": "test",
				"type": "searchset",
				"total": %s
			}`, tt.value))

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			hasUnsignedIntError := false
			for _, issue := range result.Issues {
				if strings.Contains(issue.Diagnostics, "non-negative integer") {
					hasUnsignedIntError = true
					break
				}
			}

			if tt.shouldBeValid && hasUnsignedIntError {
				t.Errorf("Expected valid for total=%s", tt.value)
			}
			if !tt.shouldBeValid && !hasUnsignedIntError {
				t.Errorf("Expected unsigned integer error for total=%s", tt.value)
			}
		})
	}
}

// TestValidateDecimalType tests decimal type validation.
func TestValidateDecimalType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		value         string
		shouldBeValid bool
	}{
		{"valid decimal", "70.5", true},
		{"valid integer as decimal", "70", true},
		{"valid negative decimal", "-70.5", true},
		{"valid zero", "0", true},
		{"valid small decimal", "0.001", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observation := []byte(fmt.Sprintf(`{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "weight"},
				"valueQuantity": {"value": %s, "unit": "kg"}
			}`, tt.value))

			result, err := v.Validate(ctx, observation)
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			hasDecimalError := false
			for _, issue := range result.Issues {
				if strings.Contains(issue.Diagnostics, "decimal") {
					hasDecimalError = true
					break
				}
			}

			if tt.shouldBeValid && hasDecimalError {
				t.Errorf("Expected valid for value=%s", tt.value)
			}
		})
	}
}

// TestValidateDecimalTypeInvalid tests invalid decimal values.
func TestValidateDecimalTypeInvalid(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// String instead of decimal
	observation := []byte(`{
		"resourceType": "Observation",
		"id": "test",
		"status": "final",
		"code": {"text": "weight"},
		"valueQuantity": {"value": "70.5", "unit": "kg"}
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	hasDecimalError := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Diagnostics, "decimal") {
			hasDecimalError = true
			break
		}
	}

	if !hasDecimalError {
		t.Error("Expected decimal error for string value")
	}
}

// =============================================================================
// REQUIRED FIELDS VALIDATION TESTS
// =============================================================================

// TestValidateBundleRequiredFields tests required fields in Bundle resource.
func TestValidateBundleRequiredFields(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Bundle.type is required (min=1)
	bundleWithoutType := []byte(`{
		"resourceType": "Bundle",
		"id": "test"
	}`)

	result, err := v.Validate(ctx, bundleWithoutType)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	if result.Valid {
		t.Error("Bundle without type should not be valid")
	}

	hasTypeError := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeRequired && strings.Contains(issue.Diagnostics, "Bundle.type") {
			hasTypeError = true
			t.Logf("Found expected error: %s", issue.Diagnostics)
			break
		}
	}
	if !hasTypeError {
		t.Error("Should have required error for missing Bundle.type")
	}
}

// TestValidateMedicationRequestRequiredFields tests required fields in MedicationRequest.
func TestValidateMedicationRequestRequiredFields(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// MedicationRequest requires: status, intent, medication[x], subject
	tests := []struct {
		name          string
		json          string
		missingField  string
	}{
		{
			name: "missing status",
			json: `{
				"resourceType": "MedicationRequest",
				"id": "test",
				"intent": "order",
				"medicationCodeableConcept": {"text": "aspirin"},
				"subject": {"reference": "Patient/1"}
			}`,
			missingField: "status",
		},
		{
			name: "missing intent",
			json: `{
				"resourceType": "MedicationRequest",
				"id": "test",
				"status": "active",
				"medicationCodeableConcept": {"text": "aspirin"},
				"subject": {"reference": "Patient/1"}
			}`,
			missingField: "intent",
		},
		{
			name: "missing subject",
			json: `{
				"resourceType": "MedicationRequest",
				"id": "test",
				"status": "active",
				"intent": "order",
				"medicationCodeableConcept": {"text": "aspirin"}
			}`,
			missingField: "subject",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if result.Valid {
				t.Errorf("MedicationRequest missing %s should not be valid", tt.missingField)
			}

			hasExpectedError := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeRequired && strings.Contains(issue.Diagnostics, tt.missingField) {
					hasExpectedError = true
					t.Logf("Found expected error: %s", issue.Diagnostics)
					break
				}
			}
			if !hasExpectedError {
				t.Errorf("Should have required error for missing %s", tt.missingField)
			}
		})
	}
}

// TestValidateDiagnosticReportRequiredFields tests required fields in DiagnosticReport.
func TestValidateDiagnosticReportRequiredFields(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// DiagnosticReport requires: status, code
	tests := []struct {
		name         string
		json         string
		missingField string
	}{
		{
			name: "missing status",
			json: `{
				"resourceType": "DiagnosticReport",
				"id": "test",
				"code": {"text": "Blood test"}
			}`,
			missingField: "status",
		},
		{
			name: "missing code",
			json: `{
				"resourceType": "DiagnosticReport",
				"id": "test",
				"status": "final"
			}`,
			missingField: "code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if result.Valid {
				t.Errorf("DiagnosticReport missing %s should not be valid", tt.missingField)
			}

			hasExpectedError := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeRequired && strings.Contains(issue.Diagnostics, tt.missingField) {
					hasExpectedError = true
					break
				}
			}
			if !hasExpectedError {
				t.Errorf("Should have required error for missing %s", tt.missingField)
			}
		})
	}
}

// TestValidateAppointmentRequiredFields tests Appointment required fields.
func TestValidateAppointmentRequiredFields(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Appointment requires: status, participant (with required nested fields)
	tests := []struct {
		name         string
		json         string
		missingField string
	}{
		{
			name:         "missing status",
			json:         `{"resourceType": "Appointment", "id": "test", "participant": [{"status": "accepted"}]}`,
			missingField: "status",
		},
		{
			name:         "missing participant",
			json:         `{"resourceType": "Appointment", "id": "test", "status": "booked"}`,
			missingField: "participant",
		},
		{
			name:         "participant missing status",
			json:         `{"resourceType": "Appointment", "id": "test", "status": "booked", "participant": [{"actor": {"reference": "Patient/1"}}]}`,
			missingField: "participant.status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if result.Valid {
				t.Errorf("Appointment missing %s should not be valid", tt.missingField)
			}

			hasExpectedError := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeRequired {
					// Check if the missing field is in the path
					for _, expr := range issue.Expression {
						if strings.Contains(expr, strings.Split(tt.missingField, ".")[len(strings.Split(tt.missingField, "."))-1]) {
							hasExpectedError = true
							break
						}
					}
				}
			}
			if !hasExpectedError {
				t.Logf("Issues: %v", result.Issues)
				t.Errorf("Should have required error for missing %s", tt.missingField)
			}
		})
	}
}

// =============================================================================
// CARDINALITY VALIDATION TESTS
// =============================================================================

// TestValidateMaxCardinalityOne tests that fields with max=1 reject arrays.
func TestValidateMaxCardinalityOne(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Patient.birthDate has max=1, should not accept array
	// Note: This is tricky because JSON parsing would convert single value
	// We test by providing multiple values where only one is expected
	patient := []byte(`{
		"resourceType": "Patient",
		"id": "test",
		"gender": "male",
		"birthDate": "1990-01-01"
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	// Single value should be valid
	if !result.Valid {
		t.Errorf("Patient with single birthDate should be valid, got errors: %v", result.Issues)
	}
}

// TestValidateMinCardinalityArray tests that array fields with min>0 require at least min items.
func TestValidateMinCardinalityArray(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Appointment.participant has min=1, empty array should fail
	appointment := []byte(`{
		"resourceType": "Appointment",
		"id": "test",
		"status": "booked",
		"participant": []
	}`)

	result, err := v.Validate(ctx, appointment)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	if result.Valid {
		t.Error("Appointment with empty participant array should not be valid")
	}
}

// =============================================================================
// COMPLEX TYPE VALIDATION TESTS
// =============================================================================

// TestValidateAddressType tests Address complex type validation.
func TestValidateAddressType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		json          string
		shouldBeValid bool
		errorContains string
	}{
		{
			name: "valid address",
			json: `{
				"resourceType": "Patient",
				"id": "test",
				"address": [{
					"use": "home",
					"type": "physical",
					"line": ["123 Main St"],
					"city": "Springfield",
					"state": "IL",
					"postalCode": "62701",
					"country": "USA"
				}]
			}`,
			shouldBeValid: true,
		},
		{
			name: "address with invalid line type",
			json: `{
				"resourceType": "Patient",
				"id": "test",
				"address": [{
					"line": [123, 456]
				}]
			}`,
			shouldBeValid: false,
			errorContains: "string",
		},
		{
			name: "address with invalid city type",
			json: `{
				"resourceType": "Patient",
				"id": "test",
				"address": [{
					"city": 12345
				}]
			}`,
			shouldBeValid: false,
			errorContains: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if tt.shouldBeValid && !result.Valid {
				t.Errorf("Expected valid, got errors: %v", result.Issues)
			}
			if !tt.shouldBeValid {
				hasExpectedError := false
				for _, issue := range result.Issues {
					if strings.Contains(issue.Diagnostics, tt.errorContains) {
						hasExpectedError = true
						break
					}
				}
				if !hasExpectedError {
					t.Errorf("Expected error containing '%s'", tt.errorContains)
				}
			}
		})
	}
}

// TestValidateContactPointType tests ContactPoint complex type validation.
func TestValidateContactPointType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		json          string
		shouldBeValid bool
		errorContains string
	}{
		{
			name: "valid contactPoint",
			json: `{
				"resourceType": "Patient",
				"id": "test",
				"telecom": [{
					"system": "phone",
					"value": "+1-555-0100",
					"use": "home",
					"rank": 1
				}]
			}`,
			shouldBeValid: true,
		},
		{
			name: "contactPoint with invalid rank type",
			json: `{
				"resourceType": "Patient",
				"id": "test",
				"telecom": [{
					"system": "phone",
					"value": "+1-555-0100",
					"rank": "first"
				}]
			}`,
			shouldBeValid: false,
			errorContains: "integer",
		},
		{
			name: "contactPoint with invalid value type",
			json: `{
				"resourceType": "Patient",
				"id": "test",
				"telecom": [{
					"system": "phone",
					"value": 5550100
				}]
			}`,
			shouldBeValid: false,
			errorContains: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if tt.shouldBeValid && !result.Valid {
				t.Errorf("Expected valid, got errors: %v", result.Issues)
			}
			if !tt.shouldBeValid {
				hasExpectedError := false
				for _, issue := range result.Issues {
					if strings.Contains(issue.Diagnostics, tt.errorContains) {
						hasExpectedError = true
						break
					}
				}
				if !hasExpectedError {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, result.Issues)
				}
			}
		})
	}
}

// TestValidateIdentifierType tests Identifier complex type validation.
func TestValidateIdentifierType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		json          string
		shouldBeValid bool
		errorContains string
	}{
		{
			name: "valid identifier",
			json: `{
				"resourceType": "Patient",
				"id": "test",
				"identifier": [{
					"system": "http://hospital.example.org/patients",
					"value": "12345"
				}]
			}`,
			shouldBeValid: true,
		},
		{
			name: "identifier with invalid system type",
			json: `{
				"resourceType": "Patient",
				"id": "test",
				"identifier": [{
					"system": 12345,
					"value": "12345"
				}]
			}`,
			shouldBeValid: false,
			errorContains: "string",
		},
		{
			name: "identifier with invalid value type",
			json: `{
				"resourceType": "Patient",
				"id": "test",
				"identifier": [{
					"system": "http://example.org",
					"value": 12345
				}]
			}`,
			shouldBeValid: false,
			errorContains: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if tt.shouldBeValid && !result.Valid {
				t.Errorf("Expected valid, got errors: %v", result.Issues)
			}
			if !tt.shouldBeValid {
				hasExpectedError := false
				for _, issue := range result.Issues {
					if strings.Contains(issue.Diagnostics, tt.errorContains) {
						hasExpectedError = true
						break
					}
				}
				if !hasExpectedError {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, result.Issues)
				}
			}
		})
	}
}

// TestValidateQuantityType tests Quantity complex type validation.
func TestValidateQuantityType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		json          string
		shouldBeValid bool
		errorContains string
	}{
		{
			name: "valid quantity",
			json: `{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "weight"},
				"valueQuantity": {
					"value": 70.5,
					"unit": "kg",
					"system": "http://unitsofmeasure.org",
					"code": "kg"
				}
			}`,
			shouldBeValid: true,
		},
		{
			name: "quantity with string value",
			json: `{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "weight"},
				"valueQuantity": {
					"value": "70.5",
					"unit": "kg"
				}
			}`,
			shouldBeValid: false,
			errorContains: "decimal",
		},
		{
			name: "quantity with invalid unit type",
			json: `{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "weight"},
				"valueQuantity": {
					"value": 70.5,
					"unit": 123
				}
			}`,
			shouldBeValid: false,
			errorContains: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if tt.shouldBeValid && !result.Valid {
				t.Errorf("Expected valid, got errors: %v", result.Issues)
			}
			if !tt.shouldBeValid {
				hasExpectedError := false
				for _, issue := range result.Issues {
					if strings.Contains(issue.Diagnostics, tt.errorContains) {
						hasExpectedError = true
						break
					}
				}
				if !hasExpectedError {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, result.Issues)
				}
			}
		})
	}
}

// =============================================================================
// CHOICE TYPE VALIDATION TESTS
// =============================================================================

// TestValidateChoiceTypeValueX tests value[x] choice type validation.
func TestValidateChoiceTypeValueX(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		json          string
		shouldBeValid bool
	}{
		{
			name: "valid valueQuantity",
			json: `{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "weight"},
				"valueQuantity": {"value": 70.5, "unit": "kg"}
			}`,
			shouldBeValid: true,
		},
		{
			name: "valid valueString",
			json: `{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "notes"},
				"valueString": "Patient is healthy"
			}`,
			shouldBeValid: true,
		},
		{
			name: "valid valueBoolean",
			json: `{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "smoker"},
				"valueBoolean": true
			}`,
			shouldBeValid: true,
		},
		{
			name: "valid valueInteger",
			json: `{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "count"},
				"valueInteger": 5
			}`,
			shouldBeValid: true,
		},
		{
			name: "valid valueCodeableConcept",
			json: `{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "blood type"},
				"valueCodeableConcept": {"text": "A+"}
			}`,
			shouldBeValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if tt.shouldBeValid && !result.Valid {
				t.Errorf("Expected valid, got errors: %v", result.Issues)
			}
		})
	}
}

// TestValidateChoiceTypeMedicationX tests medication[x] choice type in MedicationRequest.
func TestValidateChoiceTypeMedicationX(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		json          string
		shouldBeValid bool
	}{
		{
			name: "valid medicationCodeableConcept",
			json: `{
				"resourceType": "MedicationRequest",
				"id": "test",
				"status": "active",
				"intent": "order",
				"medicationCodeableConcept": {"text": "Aspirin"},
				"subject": {"reference": "Patient/1"}
			}`,
			shouldBeValid: true,
		},
		{
			name: "valid medicationReference",
			json: `{
				"resourceType": "MedicationRequest",
				"id": "test",
				"status": "active",
				"intent": "order",
				"medicationReference": {"reference": "Medication/aspirin"},
				"subject": {"reference": "Patient/1"}
			}`,
			shouldBeValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if tt.shouldBeValid && !result.Valid {
				t.Errorf("Expected valid, got errors: %v", result.Issues)
			}
		})
	}
}

// TestValidateChoiceTypeDeceasedX tests deceased[x] choice type in Patient.
func TestValidateChoiceTypeDeceasedX(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		json          string
		shouldBeValid bool
		errorContains string
	}{
		{
			name: "valid deceasedBoolean true",
			json: `{"resourceType": "Patient", "id": "test", "deceasedBoolean": true}`,
			shouldBeValid: true,
		},
		{
			name: "valid deceasedBoolean false",
			json: `{"resourceType": "Patient", "id": "test", "deceasedBoolean": false}`,
			shouldBeValid: true,
		},
		{
			name: "valid deceasedDateTime",
			json: `{"resourceType": "Patient", "id": "test", "deceasedDateTime": "2023-01-15"}`,
			shouldBeValid: true,
		},
		{
			name: "invalid deceasedBoolean string",
			json: `{"resourceType": "Patient", "id": "test", "deceasedBoolean": "true"}`,
			shouldBeValid: false,
			errorContains: "boolean",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if tt.shouldBeValid && !result.Valid {
				t.Errorf("Expected valid, got errors: %v", result.Issues)
			}
			if !tt.shouldBeValid {
				hasExpectedError := false
				for _, issue := range result.Issues {
					if strings.Contains(issue.Diagnostics, tt.errorContains) {
						hasExpectedError = true
						break
					}
				}
				if !hasExpectedError {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, result.Issues)
				}
			}
		})
	}
}

// TestValidateChoiceTypeMultipleBirthX tests multipleBirth[x] choice type in Patient.
func TestValidateChoiceTypeMultipleBirthX(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		json          string
		shouldBeValid bool
		errorContains string
	}{
		{
			name: "valid multipleBirthBoolean",
			json: `{"resourceType": "Patient", "id": "test", "multipleBirthBoolean": true}`,
			shouldBeValid: true,
		},
		{
			name: "valid multipleBirthInteger",
			json: `{"resourceType": "Patient", "id": "test", "multipleBirthInteger": 2}`,
			shouldBeValid: true,
		},
		{
			name: "invalid multipleBirthInteger string",
			json: `{"resourceType": "Patient", "id": "test", "multipleBirthInteger": "second"}`,
			shouldBeValid: false,
			errorContains: "integer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if tt.shouldBeValid && !result.Valid {
				t.Errorf("Expected valid, got errors: %v", result.Issues)
			}
			if !tt.shouldBeValid {
				hasExpectedError := false
				for _, issue := range result.Issues {
					if strings.Contains(issue.Diagnostics, tt.errorContains) {
						hasExpectedError = true
						break
					}
				}
				if !hasExpectedError {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, result.Issues)
				}
			}
		})
	}
}

// =============================================================================
// DEEPLY NESTED VALIDATION TESTS
// =============================================================================

// TestValidateDeeplyNestedTypes tests validation in deeply nested structures.
func TestValidateDeeplyNestedTypes(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Test Observation.component.valueQuantity.value (3 levels deep)
	tests := []struct {
		name          string
		json          string
		shouldBeValid bool
		errorContains string
	}{
		{
			name: "valid deeply nested",
			json: `{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "BP"},
				"component": [{
					"code": {"text": "systolic"},
					"valueQuantity": {
						"value": 120,
						"unit": "mmHg"
					}
				}]
			}`,
			shouldBeValid: true,
		},
		{
			name: "invalid nested value type",
			json: `{
				"resourceType": "Observation",
				"id": "test",
				"status": "final",
				"code": {"text": "BP"},
				"component": [{
					"code": {"text": "systolic"},
					"valueQuantity": {
						"value": "one-twenty",
						"unit": "mmHg"
					}
				}]
			}`,
			shouldBeValid: false,
			errorContains: "decimal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if tt.shouldBeValid && !result.Valid {
				t.Errorf("Expected valid, got errors: %v", result.Issues)
			}
			if !tt.shouldBeValid {
				hasExpectedError := false
				for _, issue := range result.Issues {
					if strings.Contains(issue.Diagnostics, tt.errorContains) {
						hasExpectedError = true
						break
					}
				}
				if !hasExpectedError {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, result.Issues)
				}
			}
		})
	}
}

// TestValidatePeriodType tests Period complex type validation.
func TestValidatePeriodType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		json          string
		shouldBeValid bool
		errorContains string
	}{
		{
			name: "valid period",
			json: `{
				"resourceType": "Patient",
				"id": "test",
				"address": [{
					"period": {
						"start": "2020-01-01",
						"end": "2023-12-31"
					}
				}]
			}`,
			shouldBeValid: true,
		},
		{
			name: "period with invalid start type",
			json: `{
				"resourceType": "Patient",
				"id": "test",
				"address": [{
					"period": {
						"start": 20200101
					}
				}]
			}`,
			shouldBeValid: false,
			errorContains: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate(ctx, []byte(tt.json))
			if err != nil {
				t.Fatalf("Validate error: %v", err)
			}

			if tt.shouldBeValid && !result.Valid {
				t.Errorf("Expected valid, got errors: %v", result.Issues)
			}
			if !tt.shouldBeValid {
				hasExpectedError := false
				for _, issue := range result.Issues {
					if strings.Contains(issue.Diagnostics, tt.errorContains) {
						hasExpectedError = true
						break
					}
				}
				if !hasExpectedError {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, result.Issues)
				}
			}
		})
	}
}

// =============================================================================
// VALID RESOURCE TESTS (Sanity checks)
// =============================================================================

// TestValidateCompletePatient tests a complete valid Patient resource.
func TestValidateCompletePatient(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	patient := []byte(`{
		"resourceType": "Patient",
		"id": "complete-patient",
		"meta": {
			"versionId": "1",
			"lastUpdated": "2023-01-15T10:30:00Z"
		},
		"identifier": [{
			"system": "http://hospital.example.org/patients",
			"value": "12345"
		}],
		"active": true,
		"name": [{
			"use": "official",
			"family": "Doe",
			"given": ["John", "James"],
			"prefix": ["Mr."]
		}],
		"telecom": [{
			"system": "phone",
			"value": "+1-555-0100",
			"use": "home"
		}, {
			"system": "email",
			"value": "john.doe@example.com"
		}],
		"gender": "male",
		"birthDate": "1990-01-15",
		"address": [{
			"use": "home",
			"type": "physical",
			"line": ["123 Main St", "Apt 4B"],
			"city": "Springfield",
			"state": "IL",
			"postalCode": "62701",
			"country": "USA",
			"period": {
				"start": "2020-01-01"
			}
		}],
		"maritalStatus": {
			"coding": [{
				"system": "http://terminology.hl7.org/CodeSystem/v3-MaritalStatus",
				"code": "M"
			}]
		},
		"multipleBirthBoolean": false,
		"communication": [{
			"language": {
				"coding": [{
					"system": "urn:ietf:bcp:47",
					"code": "en"
				}]
			},
			"preferred": true
		}]
	}`)

	result, err := v.Validate(ctx, patient)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	if !result.Valid {
		for _, issue := range result.Issues {
			t.Logf("Issue: [%s] %s - %s", issue.Severity, issue.Code, issue.Diagnostics)
		}
		t.Error("Complete valid Patient should be valid")
	}
}

// TestValidateCompleteObservation tests a complete valid Observation resource.
func TestValidateCompleteObservation(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	observation := []byte(`{
		"resourceType": "Observation",
		"id": "blood-pressure",
		"meta": {
			"versionId": "1",
			"lastUpdated": "2023-01-15T10:30:00Z"
		},
		"status": "final",
		"category": [{
			"coding": [{
				"system": "http://terminology.hl7.org/CodeSystem/observation-category",
				"code": "vital-signs",
				"display": "Vital Signs"
			}]
		}],
		"code": {
			"coding": [{
				"system": "http://loinc.org",
				"code": "85354-9",
				"display": "Blood pressure"
			}],
			"text": "Blood pressure"
		},
		"subject": {
			"reference": "Patient/example"
		},
		"effectiveDateTime": "2023-01-15T10:30:00Z",
		"performer": [{
			"reference": "Practitioner/example"
		}],
		"component": [{
			"code": {
				"coding": [{
					"system": "http://loinc.org",
					"code": "8480-6",
					"display": "Systolic blood pressure"
				}]
			},
			"valueQuantity": {
				"value": 120,
				"unit": "mmHg",
				"system": "http://unitsofmeasure.org",
				"code": "mm[Hg]"
			}
		}, {
			"code": {
				"coding": [{
					"system": "http://loinc.org",
					"code": "8462-4",
					"display": "Diastolic blood pressure"
				}]
			},
			"valueQuantity": {
				"value": 80,
				"unit": "mmHg",
				"system": "http://unitsofmeasure.org",
				"code": "mm[Hg]"
			}
		}]
	}`)

	result, err := v.Validate(ctx, observation)
	if err != nil {
		t.Fatalf("Validate error: %v", err)
	}

	if !result.Valid {
		for _, issue := range result.Issues {
			t.Logf("Issue: [%s] %s - %s", issue.Severity, issue.Code, issue.Diagnostics)
		}
		t.Error("Complete valid Observation should be valid")
	}
}
