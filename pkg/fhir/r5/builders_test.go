package r5

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		)

		require.NotNil(t, patient)
		require.Len(t, patient.Name, 1)
		assert.Equal(t, "Smith", *patient.Name[0].Family)
		assert.Equal(t, NameUseOfficial, *patient.Name[0].Use)
	})
}

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

	t.Run("JSON round trip", func(t *testing.T) {
		family := "Johnson"

		original := NewPatientBuilder().
			SetId("pt-json").
			SetActive(true).
			SetGender(AdministrativeGenderMale).
			AddName(HumanName{Family: &family, Given: []string{"Robert"}}).
			Build()

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded Patient
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, *original.Id, *decoded.Id)
		assert.Equal(t, *original.Active, *decoded.Active)
		assert.Equal(t, *original.Gender, *decoded.Gender)
	})
}

func TestObservationBuilder(t *testing.T) {
	t.Run("build observation with fluent API", func(t *testing.T) {
		value := 120.0
		unit := "mmHg"

		obs := NewObservationBuilder().
			SetId("obs-bp").
			SetStatus(ObservationStatusFinal).
			SetCode(CodeableConcept{}).
			SetValueQuantity(Quantity{
				Value: &value,
				Unit:  &unit,
			}).
			SetEffectiveDateTime("2024-06-15T14:30:00Z").
			Build()

		require.NotNil(t, obs)
		assert.Equal(t, "obs-bp", *obs.Id)
		assert.Equal(t, ObservationStatusFinal, *obs.Status)
		require.NotNil(t, obs.ValueQuantity)
		assert.Equal(t, 120.0, *obs.ValueQuantity.Value)
		assert.Equal(t, "mmHg", *obs.ValueQuantity.Unit)
	})
}

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
