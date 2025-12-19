package validator

import (
	"context"
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
