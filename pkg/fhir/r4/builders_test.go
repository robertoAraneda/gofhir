package r4

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Functional Options Pattern Tests
// =============================================================================

func TestPatientFunctionalOptions(t *testing.T) {
	t.Run("create patient with options", func(t *testing.T) {
		patient := NewPatient(
			WithPatientId("patient-123"),
			WithPatientActive(true),
			WithPatientGender(AdministrativeGenderMale),
			WithPatientBirthDate("1990-01-15"),
		)

		require.NotNil(t, patient)
		assert.Equal(t, "patient-123", *patient.Id)
		assert.True(t, *patient.Active)
		assert.Equal(t, AdministrativeGenderMale, *patient.Gender)
		assert.Equal(t, "1990-01-15", *patient.BirthDate)
	})

	t.Run("add multiple names", func(t *testing.T) {
		family := "Smith"
		use := NameUseOfficial

		patient := NewPatient(
			WithPatientId("patient-456"),
			WithPatientName(HumanName{
				Use:    &use,
				Family: &family,
				Given:  []string{"John"},
			}),
			WithPatientName(HumanName{
				Family: &family,
				Given:  []string{"Johnny"},
			}),
		)

		require.NotNil(t, patient)
		require.Len(t, patient.Name, 2)
		assert.Equal(t, "Smith", *patient.Name[0].Family)
		assert.Equal(t, NameUseOfficial, *patient.Name[0].Use)
	})

	t.Run("add identifiers", func(t *testing.T) {
		system := "http://hospital.example.org/mrn"
		value := "12345"

		patient := NewPatient(
			WithPatientIdentifier(Identifier{
				System: &system,
				Value:  &value,
			}),
		)

		require.NotNil(t, patient)
		require.Len(t, patient.Identifier, 1)
		assert.Equal(t, "http://hospital.example.org/mrn", *patient.Identifier[0].System)
		assert.Equal(t, "12345", *patient.Identifier[0].Value)
	})

	t.Run("empty patient", func(t *testing.T) {
		patient := NewPatient()

		require.NotNil(t, patient)
		assert.Nil(t, patient.Id)
		assert.Nil(t, patient.Active)
		assert.Empty(t, patient.Name)
	})
}

func TestObservationFunctionalOptions(t *testing.T) {
	t.Run("create observation with options", func(t *testing.T) {
		codeSystem := "http://loinc.org"
		codeCode := "8867-4"
		codeDisplay := "Heart rate"

		obs := NewObservation(
			WithObservationId("obs-123"),
			WithObservationStatus(ObservationStatusFinal),
			WithObservationCode(CodeableConcept{
				Coding: []Coding{
					{System: &codeSystem, Code: &codeCode, Display: &codeDisplay},
				},
			}),
			WithObservationEffectiveDateTime("2024-01-15T10:30:00Z"),
		)

		require.NotNil(t, obs)
		assert.Equal(t, "obs-123", *obs.Id)
		assert.Equal(t, ObservationStatusFinal, *obs.Status)
		assert.Equal(t, "2024-01-15T10:30:00Z", *obs.EffectiveDateTime)
		require.Len(t, obs.Code.Coding, 1)
		assert.Equal(t, "http://loinc.org", *obs.Code.Coding[0].System)
	})

	t.Run("observation with value quantity", func(t *testing.T) {
		value := 72.0
		unit := "bpm"
		system := "http://unitsofmeasure.org"
		code := "/min"

		obs := NewObservation(
			WithObservationId("obs-456"),
			WithObservationStatus(ObservationStatusFinal),
			WithObservationValueQuantity(Quantity{
				Value:  &value,
				Unit:   &unit,
				System: &system,
				Code:   &code,
			}),
		)

		require.NotNil(t, obs)
		require.NotNil(t, obs.ValueQuantity)
		assert.Equal(t, 72.0, *obs.ValueQuantity.Value)
		assert.Equal(t, "bpm", *obs.ValueQuantity.Unit)
	})
}

// =============================================================================
// Fluent Builder Pattern Tests
// =============================================================================

func TestPatientBuilder(t *testing.T) {
	t.Run("build patient with fluent API", func(t *testing.T) {
		family := "Garcia"
		use := NameUseOfficial

		patient := NewPatientBuilder().
			SetId("patient-789").
			SetActive(true).
			SetGender(AdministrativeGenderFemale).
			SetBirthDate("1985-06-20").
			AddName(HumanName{
				Use:    &use,
				Family: &family,
				Given:  []string{"Maria"},
			}).
			Build()

		require.NotNil(t, patient)
		assert.Equal(t, "patient-789", *patient.Id)
		assert.True(t, *patient.Active)
		assert.Equal(t, AdministrativeGenderFemale, *patient.Gender)
		assert.Equal(t, "1985-06-20", *patient.BirthDate)
		require.Len(t, patient.Name, 1)
		assert.Equal(t, "Garcia", *patient.Name[0].Family)
	})

	t.Run("add multiple elements", func(t *testing.T) {
		system := "http://hospital.example.org/mrn"
		value1 := "MRN-001"
		value2 := "MRN-002"

		patient := NewPatientBuilder().
			SetId("patient-multi").
			AddIdentifier(Identifier{System: &system, Value: &value1}).
			AddIdentifier(Identifier{System: &system, Value: &value2}).
			Build()

		require.NotNil(t, patient)
		require.Len(t, patient.Identifier, 2)
		assert.Equal(t, "MRN-001", *patient.Identifier[0].Value)
		assert.Equal(t, "MRN-002", *patient.Identifier[1].Value)
	})

	t.Run("JSON round trip", func(t *testing.T) {
		family := "Johnson"
		city := "Boston"
		use := AddressUseHome

		original := NewPatientBuilder().
			SetId("pt-json").
			SetActive(true).
			SetGender(AdministrativeGenderMale).
			AddName(HumanName{Family: &family, Given: []string{"Robert"}}).
			AddAddress(Address{Use: &use, City: &city}).
			Build()

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded Patient
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, *original.Id, *decoded.Id)
		assert.Equal(t, *original.Active, *decoded.Active)
		assert.Equal(t, *original.Gender, *decoded.Gender)
		require.Len(t, decoded.Name, 1)
		assert.Equal(t, *original.Name[0].Family, *decoded.Name[0].Family)
	})
}

func TestObservationBuilder(t *testing.T) {
	t.Run("build observation with fluent API", func(t *testing.T) {
		codeSystem := "http://loinc.org"
		codeCode := "8480-6"
		codeDisplay := "Systolic blood pressure"
		value := 120.0
		unit := "mmHg"
		unitSystem := "http://unitsofmeasure.org"
		unitCode := "mm[Hg]"

		obs := NewObservationBuilder().
			SetId("obs-bp").
			SetStatus(ObservationStatusFinal).
			SetCode(CodeableConcept{
				Coding: []Coding{
					{System: &codeSystem, Code: &codeCode, Display: &codeDisplay},
				},
			}).
			SetValueQuantity(Quantity{
				Value:  &value,
				Unit:   &unit,
				System: &unitSystem,
				Code:   &unitCode,
			}).
			SetEffectiveDateTime("2024-06-15T14:30:00Z").
			Build()

		require.NotNil(t, obs)
		assert.Equal(t, "obs-bp", *obs.Id)
		assert.Equal(t, ObservationStatusFinal, *obs.Status)
		require.NotNil(t, obs.ValueQuantity)
		assert.Equal(t, 120.0, *obs.ValueQuantity.Value)
		assert.Equal(t, "mmHg", *obs.ValueQuantity.Unit)
		assert.Equal(t, "2024-06-15T14:30:00Z", *obs.EffectiveDateTime)
	})

	t.Run("add categories and performers", func(t *testing.T) {
		catSystem := "http://terminology.hl7.org/CodeSystem/observation-category"
		catCode := "vital-signs"
		ref := "Practitioner/123"

		obs := NewObservationBuilder().
			SetId("obs-cat").
			SetStatus(ObservationStatusFinal).
			SetCode(CodeableConcept{}).
			AddCategory(CodeableConcept{
				Coding: []Coding{{System: &catSystem, Code: &catCode}},
			}).
			AddPerformer(Reference{Reference: &ref}).
			Build()

		require.NotNil(t, obs)
		require.Len(t, obs.Category, 1)
		require.Len(t, obs.Performer, 1)
		assert.Equal(t, "vital-signs", *obs.Category[0].Coding[0].Code)
		assert.Equal(t, "Practitioner/123", *obs.Performer[0].Reference)
	})
}

func TestPractitionerBuilder(t *testing.T) {
	t.Run("build practitioner", func(t *testing.T) {
		family := "Wilson"
		use := NameUseOfficial

		practitioner := NewPractitionerBuilder().
			SetId("prac-001").
			SetActive(true).
			SetGender(AdministrativeGenderFemale).
			AddName(HumanName{
				Use:    &use,
				Family: &family,
				Given:  []string{"Sarah"},
				Prefix: []string{"Dr."},
			}).
			Build()

		require.NotNil(t, practitioner)
		assert.Equal(t, "prac-001", *practitioner.Id)
		assert.True(t, *practitioner.Active)
		assert.Equal(t, AdministrativeGenderFemale, *practitioner.Gender)
		require.Len(t, practitioner.Name, 1)
		assert.Equal(t, "Wilson", *practitioner.Name[0].Family)
	})
}

func TestOrganizationBuilder(t *testing.T) {
	t.Run("build organization", func(t *testing.T) {
		org := NewOrganizationBuilder().
			SetId("org-001").
			SetActive(true).
			SetName("General Hospital").
			Build()

		require.NotNil(t, org)
		assert.Equal(t, "org-001", *org.Id)
		assert.True(t, *org.Active)
		assert.Equal(t, "General Hospital", *org.Name)
	})
}

func TestConditionBuilder(t *testing.T) {
	t.Run("build condition", func(t *testing.T) {
		clinicalSystem := "http://terminology.hl7.org/CodeSystem/condition-clinical"
		clinicalCode := "active"
		ref := "Patient/123"

		condition := NewConditionBuilder().
			SetId("cond-001").
			SetClinicalStatus(CodeableConcept{
				Coding: []Coding{{System: &clinicalSystem, Code: &clinicalCode}},
			}).
			SetSubject(Reference{Reference: &ref}).
			SetOnsetDateTime("2024-01-15").
			Build()

		require.NotNil(t, condition)
		assert.Equal(t, "cond-001", *condition.Id)
		require.NotNil(t, condition.ClinicalStatus)
		assert.Equal(t, "active", *condition.ClinicalStatus.Coding[0].Code)
		assert.Equal(t, "Patient/123", *condition.Subject.Reference)
		assert.Equal(t, "2024-01-15", *condition.OnsetDateTime)
	})
}

func TestBundleBuilder(t *testing.T) {
	t.Run("build transaction bundle", func(t *testing.T) {
		bundle := NewBundleBuilder().
			SetId("bundle-001").
			SetType(BundleTypeTransaction).
			Build()

		require.NotNil(t, bundle)
		assert.Equal(t, "bundle-001", *bundle.Id)
		assert.Equal(t, BundleTypeTransaction, *bundle.Type)
	})
}

// =============================================================================
// Mixed Usage Tests
// =============================================================================

func TestMixedBuilderPatterns(t *testing.T) {
	t.Run("functional options and builder produce same result", func(t *testing.T) {
		family := "Test"

		// Using functional options
		patient1 := NewPatient(
			WithPatientId("test-001"),
			WithPatientActive(true),
			WithPatientGender(AdministrativeGenderMale),
			WithPatientName(HumanName{Family: &family}),
		)

		// Using fluent builder
		patient2 := NewPatientBuilder().
			SetId("test-001").
			SetActive(true).
			SetGender(AdministrativeGenderMale).
			AddName(HumanName{Family: &family}).
			Build()

		// Both should produce equivalent results
		assert.Equal(t, *patient1.Id, *patient2.Id)
		assert.Equal(t, *patient1.Active, *patient2.Active)
		assert.Equal(t, *patient1.Gender, *patient2.Gender)
		require.Len(t, patient1.Name, 1)
		require.Len(t, patient2.Name, 1)
		assert.Equal(t, *patient1.Name[0].Family, *patient2.Name[0].Family)
	})
}
