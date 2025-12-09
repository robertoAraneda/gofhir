// Package main demonstrates how to create FHIR resources using gofhir.
// This example shows three different approaches: direct struct construction,
// fluent builders, and functional options - for R4, R4B, and R5 versions.
package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/robertoaraneda/gofhir/pkg/fhir/r4"
	"github.com/robertoaraneda/gofhir/pkg/fhir/r4b"
	"github.com/robertoaraneda/gofhir/pkg/fhir/r5"
)

func main() {
	fmt.Println("=== FHIR R4 Examples ===")
	demonstrateR4()

	fmt.Println("\n=== FHIR R4B Examples ===")
	demonstrateR4B()

	fmt.Println("\n=== FHIR R5 Examples ===")
	demonstrateR5()
}

// =============================================================================
// FHIR R4 Examples
// =============================================================================

func demonstrateR4() {
	// 1. Direct struct construction
	fmt.Println("\n--- 1. Direct Struct Construction ---")
	patientDirect := createR4PatientDirect()
	printJSON("Patient (direct)", patientDirect)

	// 2. Fluent Builder pattern
	fmt.Println("\n--- 2. Fluent Builder Pattern ---")
	patientFluent := createR4PatientFluent()
	printJSON("Patient (fluent)", patientFluent)

	// 3. Functional Options pattern
	fmt.Println("\n--- 3. Functional Options Pattern ---")
	patientOptions := createR4PatientOptions()
	printJSON("Patient (options)", patientOptions)

	// 4. Complex resource: Observation
	fmt.Println("\n--- 4. Complex Resource: Observation ---")
	observation := createR4Observation()
	printJSON("Observation", observation)
}

// createR4PatientDirect creates a Patient using direct struct initialization
func createR4PatientDirect() *r4.Patient {
	active := true
	male := r4.AdministrativeGenderMale
	birthDate := "1990-05-15"

	return &r4.Patient{
		Id:        ptr("example-r4"),
		Active:    &active,
		Gender:    &male,
		BirthDate: &birthDate,
		Name: []r4.HumanName{
			{
				Use:    ptr(r4.NameUseOfficial),
				Family: ptr("Doe"),
				Given:  []string{"John", "James"},
			},
			{
				Use:   ptr(r4.NameUseNickname),
				Given: []string{"Johnny"},
			},
		},
		Telecom: []r4.ContactPoint{
			{
				System: ptr(r4.ContactPointSystemPhone),
				Value:  ptr("+1-555-0100"),
				Use:    ptr(r4.ContactPointUseHome),
			},
			{
				System: ptr(r4.ContactPointSystemEmail),
				Value:  ptr("john.doe@example.com"),
				Use:    ptr(r4.ContactPointUseWork),
			},
		},
		Address: []r4.Address{
			{
				Use:        ptr(r4.AddressUseHome),
				Type:       ptr(r4.AddressTypePhysical),
				Line:       []string{"123 Main Street", "Apt 4B"},
				City:       ptr("Springfield"),
				State:      ptr("IL"),
				PostalCode: ptr("62701"),
				Country:    ptr("USA"),
			},
		},
		Identifier: []r4.Identifier{
			{
				System: ptr("http://hospital.example.org/patients"),
				Value:  ptr("12345"),
			},
		},
	}
}

// createR4PatientFluent creates a Patient using the fluent builder pattern
func createR4PatientFluent() *r4.Patient {
	return r4.NewPatientBuilder().
		SetId("example-fluent-r4").
		SetActive(true).
		SetGender(r4.AdministrativeGenderFemale).
		SetBirthDate("1985-08-22").
		AddName(r4.HumanName{
			Use:    ptr(r4.NameUseOfficial),
			Family: ptr("Smith"),
			Given:  []string{"Jane", "Marie"},
		}).
		AddTelecom(r4.ContactPoint{
			System: ptr(r4.ContactPointSystemPhone),
			Value:  ptr("+1-555-0200"),
			Use:    ptr(r4.ContactPointUseMobile),
		}).
		AddAddress(r4.Address{
			Use:        ptr(r4.AddressUseHome),
			City:       ptr("Chicago"),
			State:      ptr("IL"),
			PostalCode: ptr("60601"),
		}).
		AddIdentifier(r4.Identifier{
			System: ptr("http://hospital.example.org/patients"),
			Value:  ptr("67890"),
		}).
		Build()
}

// createR4PatientOptions creates a Patient using functional options
func createR4PatientOptions() *r4.Patient {
	return r4.NewPatient(
		r4.WithPatientId("example-options-r4"),
		r4.WithPatientActive(true),
		r4.WithPatientGender(r4.AdministrativeGenderOther),
		r4.WithPatientBirthDate("2000-12-01"),
		r4.WithPatientName(r4.HumanName{
			Use:    ptr(r4.NameUseOfficial),
			Family: ptr("Garcia"),
			Given:  []string{"Alex"},
		}),
		r4.WithPatientTelecom(r4.ContactPoint{
			System: ptr(r4.ContactPointSystemEmail),
			Value:  ptr("alex.garcia@example.com"),
		}),
		r4.WithPatientIdentifier(r4.Identifier{
			System: ptr("http://hospital.example.org/patients"),
			Value:  ptr("11111"),
		}),
	)
}

// createR4Observation creates a blood pressure Observation
func createR4Observation() *r4.Observation {
	return r4.NewObservationBuilder().
		SetId("blood-pressure-r4").
		SetStatus(r4.ObservationStatusFinal).
		SetCode(r4.CodeableConcept{
			Coding: []r4.Coding{
				{
					System:  ptr("http://loinc.org"),
					Code:    ptr("85354-9"),
					Display: ptr("Blood pressure panel"),
				},
			},
			Text: ptr("Blood Pressure"),
		}).
		SetSubject(r4.Reference{
			Reference: ptr("Patient/example-r4"),
			Display:   ptr("John Doe"),
		}).
		SetEffectiveDateTime("2024-01-15T10:30:00Z").
		AddCategory(r4.CodeableConcept{
			Coding: []r4.Coding{
				{
					System:  ptr("http://terminology.hl7.org/CodeSystem/observation-category"),
					Code:    ptr("vital-signs"),
					Display: ptr("Vital Signs"),
				},
			},
		}).
		AddComponent(r4.ObservationComponent{
			Code: r4.CodeableConcept{
				Coding: []r4.Coding{
					{
						System:  ptr("http://loinc.org"),
						Code:    ptr("8480-6"),
						Display: ptr("Systolic blood pressure"),
					},
				},
			},
			ValueQuantity: &r4.Quantity{
				Value:  ptr(120.0),
				Unit:   ptr("mmHg"),
				System: ptr("http://unitsofmeasure.org"),
				Code:   ptr("mm[Hg]"),
			},
		}).
		AddComponent(r4.ObservationComponent{
			Code: r4.CodeableConcept{
				Coding: []r4.Coding{
					{
						System:  ptr("http://loinc.org"),
						Code:    ptr("8462-4"),
						Display: ptr("Diastolic blood pressure"),
					},
				},
			},
			ValueQuantity: &r4.Quantity{
				Value:  ptr(80.0),
				Unit:   ptr("mmHg"),
				System: ptr("http://unitsofmeasure.org"),
				Code:   ptr("mm[Hg]"),
			},
		}).
		Build()
}

// =============================================================================
// FHIR R4B Examples
// =============================================================================

func demonstrateR4B() {
	// Fluent Builder for R4B
	patient := r4b.NewPatientBuilder().
		SetId("example-r4b").
		SetActive(true).
		SetGender(r4b.AdministrativeGenderMale).
		SetBirthDate("1992-03-10").
		AddName(r4b.HumanName{
			Use:    ptr(r4b.NameUseOfficial),
			Family: ptr("Wilson"),
			Given:  []string{"Robert"},
		}).
		AddTelecom(r4b.ContactPoint{
			System: ptr(r4b.ContactPointSystemPhone),
			Value:  ptr("+1-555-0300"),
		}).
		Build()

	printJSON("Patient R4B", patient)

	// Medication example (shows R4B specific features)
	medication := r4b.NewMedicationBuilder().
		SetId("med-example-r4b").
		SetCode(r4b.CodeableConcept{
			Coding: []r4b.Coding{
				{
					System:  ptr("http://www.nlm.nih.gov/research/umls/rxnorm"),
					Code:    ptr("1049502"),
					Display: ptr("Acetaminophen 325 MG Oral Tablet"),
				},
			},
		}).
		SetStatus(r4b.MedicationStatusCodesActive).
		Build()

	printJSON("Medication R4B", medication)
}

// =============================================================================
// FHIR R5 Examples
// =============================================================================

func demonstrateR5() {
	// R5 Patient with new features
	patient := r5.NewPatientBuilder().
		SetId("example-r5").
		SetActive(true).
		SetGender(r5.AdministrativeGenderFemale).
		SetBirthDate("1988-11-25").
		AddName(r5.HumanName{
			Use:    ptr(r5.NameUseOfficial),
			Family: ptr("Johnson"),
			Given:  []string{"Emily", "Rose"},
		}).
		AddTelecom(r5.ContactPoint{
			System: ptr(r5.ContactPointSystemPhone),
			Value:  ptr("+1-555-0400"),
			Use:    ptr(r5.ContactPointUseMobile),
		}).
		AddAddress(r5.Address{
			Use:        ptr(r5.AddressUseHome),
			City:       ptr("Boston"),
			State:      ptr("MA"),
			PostalCode: ptr("02101"),
			Country:    ptr("USA"),
		}).
		Build()

	printJSON("Patient R5", patient)

	// R5 Observation with enhanced features
	observation := r5.NewObservationBuilder().
		SetId("vitals-r5").
		SetStatus(r5.ObservationStatusFinal).
		SetCode(r5.CodeableConcept{
			Coding: []r5.Coding{
				{
					System:  ptr("http://loinc.org"),
					Code:    ptr("8867-4"),
					Display: ptr("Heart rate"),
				},
			},
		}).
		SetSubject(r5.Reference{
			Reference: ptr("Patient/example-r5"),
		}).
		SetEffectiveDateTime("2024-01-15T14:00:00Z").
		SetValueQuantity(r5.Quantity{
			Value:  ptr(72.0),
			Unit:   ptr("beats/minute"),
			System: ptr("http://unitsofmeasure.org"),
			Code:   ptr("/min"),
		}).
		Build()

	printJSON("Observation R5", observation)
}

// =============================================================================
// Helper Functions
// =============================================================================

// ptr is a helper to create pointers to values
func ptr[T any](v T) *T {
	return &v
}

// printJSON prints a resource as formatted JSON
func printJSON(name string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error marshaling %s: %v", name, err)
		return
	}
	fmt.Printf("%s:\n%s\n", name, string(data))
}
