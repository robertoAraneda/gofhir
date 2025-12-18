# FHIR

Strongly-typed FHIR resource definitions for Go.

## Overview

This package provides strongly-typed FHIR resource definitions with support for multiple FHIR versions (R4, R4B, R5). It includes auto-generated resource types, fluent builders, and functional options for creating FHIR resources.

## Installation

```go
import "github.com/robertoaraneda/gofhir/pkg/fhir/r4"
import "github.com/robertoaraneda/gofhir/pkg/fhir/r4b"
import "github.com/robertoaraneda/gofhir/pkg/fhir/r5"
```

## Quick Start

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/robertoaraneda/gofhir/pkg/fhir/r4"
)

func main() {
    // Using Builder Pattern
    patient := r4.NewPatientBuilder().
        SetId("patient-123").
        SetActive(true).
        AddName(r4.HumanName{
            Use:    ptr("official"),
            Family: ptr("Doe"),
            Given:  []string{"John", "Michael"},
        }).
        AddTelecom(r4.ContactPoint{
            System: ptr("email"),
            Value:  ptr("john.doe@example.com"),
        }).
        Build()

    // Serialize to JSON
    data, _ := json.MarshalIndent(patient, "", "  ")
    fmt.Println(string(data))
}

func ptr[T any](v T) *T { return &v }
```

## Supported Versions

| Version | Package | FHIR Version |
|---------|---------|--------------|
| R4 | `pkg/fhir/r4` | 4.0.1 |
| R4B | `pkg/fhir/r4b` | 4.3.0 |
| R5 | `pkg/fhir/r5` | 5.0.0 |

## Resource Types

### R4 Resources (150+)

#### Clinical Resources

| Category | Resources |
|----------|-----------|
| **Patient** | Patient, RelatedPerson, Person, Group |
| **Practitioner** | Practitioner, PractitionerRole, Organization |
| **Care Team** | CareTeam, CarePlan, Goal, ServiceRequest |
| **Encounter** | Encounter, EpisodeOfCare, Flag |
| **Condition** | Condition, Procedure, FamilyMemberHistory |
| **Observation** | Observation, DiagnosticReport, Specimen |
| **Allergy** | AllergyIntolerance, AdverseEvent |

#### Medication Resources

| Category | Resources |
|----------|-----------|
| **Medications** | Medication, MedicationRequest, MedicationAdministration |
| **Dispensing** | MedicationDispense, MedicationStatement |
| **Immunization** | Immunization, ImmunizationRecommendation |

#### Financial Resources

| Category | Resources |
|----------|-----------|
| **Billing** | Claim, ClaimResponse, Invoice |
| **Coverage** | Coverage, CoverageEligibilityRequest/Response |
| **Payment** | PaymentNotice, PaymentReconciliation |

#### Infrastructure Resources

| Category | Resources |
|----------|-----------|
| **Bundle** | Bundle, Binary, Parameters |
| **Definitions** | StructureDefinition, ValueSet, CodeSystem |
| **Operations** | OperationDefinition, OperationOutcome |

## Creating Resources

### Builder Pattern

Every resource has a fluent builder:

```go
// Patient with full details
patient := r4.NewPatientBuilder().
    SetId("123").
    SetActive(true).
    SetGender("male").
    SetBirthDate("1990-05-15").
    AddIdentifier(r4.Identifier{
        System: ptr("http://hospital.org/mrn"),
        Value:  ptr("MRN-12345"),
    }).
    AddName(r4.HumanName{
        Use:    ptr("official"),
        Family: ptr("Doe"),
        Given:  []string{"John"},
    }).
    AddAddress(r4.Address{
        Use:        ptr("home"),
        Line:       []string{"123 Main St"},
        City:       ptr("Springfield"),
        State:      ptr("IL"),
        PostalCode: ptr("62701"),
    }).
    Build()
```

### Functional Options Pattern

```go
// Using functional options
patient := r4.NewPatient(
    r4.WithPatientId("123"),
    r4.WithPatientActive(true),
    r4.WithPatientGender("male"),
    r4.WithPatientBirthDate("1990-05-15"),
)
```

### Direct Instantiation

```go
// Direct struct creation
patient := &r4.Patient{
    ResourceType: "Patient",
    Id:           ptr("123"),
    Active:       ptr(true),
    Gender:       ptr("male"),
    Name: []r4.HumanName{
        {
            Family: ptr("Doe"),
            Given:  []string{"John"},
        },
    },
}
```

## Data Types

### Primitive Types

| Type | Go Type | Example |
|------|---------|---------|
| `string` | `*string` | `ptr("value")` |
| `boolean` | `*bool` | `ptr(true)` |
| `integer` | `*int` | `ptr(42)` |
| `decimal` | `*float64` | `ptr(3.14)` |
| `date` | `*string` | `ptr("2024-01-15")` |
| `dateTime` | `*string` | `ptr("2024-01-15T10:30:00Z")` |
| `instant` | `*string` | `ptr("2024-01-15T10:30:00.000Z")` |
| `uri` | `*string` | `ptr("http://example.com")` |
| `code` | `*string` | `ptr("active")` |

### Complex Types

```go
// CodeableConcept
codeableConcept := r4.CodeableConcept{
    Coding: []r4.Coding{
        {
            System:  ptr("http://loinc.org"),
            Code:    ptr("8867-4"),
            Display: ptr("Heart rate"),
        },
    },
    Text: ptr("Heart rate"),
}

// Quantity
quantity := r4.Quantity{
    Value:  ptr(72.0),
    Unit:   ptr("beats/min"),
    System: ptr("http://unitsofmeasure.org"),
    Code:   ptr("/min"),
}

// Reference
reference := r4.Reference{
    Reference: ptr("Patient/123"),
    Display:   ptr("John Doe"),
}

// Period
period := r4.Period{
    Start: ptr("2024-01-01T00:00:00Z"),
    End:   ptr("2024-12-31T23:59:59Z"),
}

// Identifier
identifier := r4.Identifier{
    Use:    ptr("official"),
    System: ptr("http://hospital.org/mrn"),
    Value:  ptr("MRN-12345"),
}
```

## Helper Functions

### LOINC Code Helpers (R4)

```go
import "github.com/robertoaraneda/gofhir/pkg/fhir/r4/helpers"

// Pre-defined LOINC codes for vital signs
heartRateCode := helpers.HeartRate()      // CodeableConcept for 8867-4
bodyTempCode := helpers.BodyTemperature() // CodeableConcept for 8310-5
bloodPressure := helpers.BloodPressure()  // CodeableConcept for 85354-9
bodyWeight := helpers.BodyWeight()        // CodeableConcept for 29463-7
bodyHeight := helpers.BodyHeight()        // CodeableConcept for 8302-2
oxygenSat := helpers.OxygenSaturation()   // CodeableConcept for 2708-6
respRate := helpers.RespiratoryRate()     // CodeableConcept for 9279-1
```

### UCUM Unit Helpers (R4)

```go
import "github.com/robertoaraneda/gofhir/pkg/fhir/r4/helpers"

// Create Quantity with common units
weight := helpers.QuantityKg(72.5)       // 72.5 kg
height := helpers.QuantityCm(175)        // 175 cm
temp := helpers.QuantityCelsius(37.2)    // 37.2 Cel
pressure := helpers.QuantityMmHg(120)    // 120 mm[Hg]
heartRate := helpers.QuantityBpm(72)     // 72 /min
```

## Working with Bundles

### Creating a Bundle

```go
bundle := r4.NewBundle(
    r4.WithBundleType(r4.BundleTypeSearchset),
    r4.WithBundleTotal(100),
    r4.WithBundleLink(r4.BundleLink{
        Relation: ptr("self"),
        Url:      ptr("http://example.com/Patient?_count=10"),
    }),
    r4.WithBundleEntry(r4.BundleEntry{
        FullUrl:  ptr("http://example.com/Patient/123"),
        Resource: patient,
    }),
)
```

### Using the Builder

```go
bundle := r4.NewBundleBuilder().
    SetType(r4.BundleTypeCollection).
    AddEntry(r4.BundleEntry{
        FullUrl:  ptr("urn:uuid:patient-1"),
        Resource: patient,
    }).
    AddEntry(r4.BundleEntry{
        FullUrl:  ptr("urn:uuid:obs-1"),
        Resource: observation,
    }).
    Build()
```

### Creating an OperationOutcome

```go
outcome := r4.NewOperationOutcome(
    r4.WithOperationOutcomeIssue(r4.OperationOutcomeIssue{
        Severity:    ptr(r4.IssueSeverityError),
        Code:        ptr(r4.IssueTypeInvalid),
        Diagnostics: ptr("Patient.birthDate is required"),
    }),
)
```

## JSON Serialization

```go
// Marshal to JSON
patient := r4.NewPatientBuilder().SetId("123").Build()
data, err := json.Marshal(patient)

// Marshal with indentation
data, err := json.MarshalIndent(patient, "", "  ")

// Unmarshal from JSON
var patient r4.Patient
err := json.Unmarshal(data, &patient)

// Dynamic unmarshaling by resource type
resourceType, err := r4.GetResourceType(data)
if err != nil {
    // handle error
}

resource, err := r4.UnmarshalResource(data)
switch r := resource.(type) {
case *r4.Patient:
    fmt.Printf("Patient: %s\n", *r.Id)
case *r4.Observation:
    fmt.Printf("Observation: %s\n", *r.Id)
}
```

## Resource Interfaces

```go
// All resources implement this interface
type Resource interface {
    GetResourceType() string
    GetId() *string
    SetId(string)
    GetMeta() *Meta
    SetMeta(*Meta)
}

// Usage
func processResource(r r4.Resource) {
    fmt.Printf("Type: %s, ID: %s\n",
        r.GetResourceType(),
        deref(r.GetId()))
}

func deref(s *string) string {
    if s == nil { return "" }
    return *s
}
```

## Examples

### Creating an Observation

```go
observation := r4.NewObservationBuilder().
    SetId("obs-123").
    SetStatus("final").
    SetCode(helpers.HeartRate()).
    SetSubject(r4.Reference{
        Reference: ptr("Patient/123"),
    }).
    SetEffectiveDateTime("2024-01-15T10:30:00Z").
    SetValueQuantity(r4.Quantity{
        Value:  ptr(72.0),
        Unit:   ptr("beats/min"),
        System: ptr("http://unitsofmeasure.org"),
        Code:   ptr("/min"),
    }).
    Build()
```

### Creating a MedicationRequest

```go
medRequest := r4.NewMedicationRequestBuilder().
    SetId("medrx-123").
    SetStatus("active").
    SetIntent("order").
    SetMedicationCodeableConcept(r4.CodeableConcept{
        Coding: []r4.Coding{
            {
                System:  ptr("http://www.nlm.nih.gov/research/umls/rxnorm"),
                Code:    ptr("1049502"),
                Display: ptr("Acetaminophen 325 MG Oral Tablet"),
            },
        },
    }).
    SetSubject(r4.Reference{
        Reference: ptr("Patient/123"),
    }).
    Build()
```

### Complete Example: Vital Signs with Helpers

This example shows how to use LOINC and UCUM helpers with builders:

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/robertoaraneda/gofhir/pkg/fhir/r4"
    "github.com/robertoaraneda/gofhir/pkg/fhir/r4/helpers"
)

func ptr[T any](v T) *T { return &v }

func main() {
    // Create a heart rate observation using helpers
    observation := r4.NewObservationBuilder().
        SetId("vitals-hr-001").
        SetStatus("final").
        SetCode(helpers.HeartRate()).              // LOINC 8867-4
        SetSubject(r4.Reference{
            Reference: ptr("Patient/patient-123"),
            Display:   ptr("John Doe"),
        }).
        SetEffectiveDateTime("2024-01-15T10:30:00Z").
        SetValueQuantity(helpers.QuantityBpm(72)). // 72 beats/min
        Build()

    // Print JSON output
    data, _ := json.MarshalIndent(observation, "", "  ")
    fmt.Println(string(data))
}
```

**JSON Output:**

```json
{
  "resourceType": "Observation",
  "id": "vitals-hr-001",
  "status": "final",
  "code": {
    "coding": [
      {
        "system": "http://loinc.org",
        "code": "8867-4",
        "display": "Heart rate"
      }
    ],
    "text": "Heart rate"
  },
  "subject": {
    "reference": "Patient/patient-123",
    "display": "John Doe"
  },
  "effectiveDateTime": "2024-01-15T10:30:00Z",
  "valueQuantity": {
    "value": 72,
    "unit": "beats/min",
    "system": "http://unitsofmeasure.org",
    "code": "/min"
  }
}
```

### Using Functional Options with Helpers

```go
// Blood pressure observation using functional options
observation := r4.NewObservation(
    r4.WithObservationId("vitals-bp-001"),
    r4.WithObservationStatus("final"),
    r4.WithObservationCode(helpers.BloodPressure()), // LOINC 85354-9
    r4.WithObservationSubject(r4.Reference{
        Reference: ptr("Patient/patient-123"),
    }),
    r4.WithObservationEffectiveDateTime("2024-01-15T10:30:00Z"),
    r4.WithObservationComponent(r4.ObservationComponent{
        Code:          helpers.SystolicBP(),         // LOINC 8480-6
        ValueQuantity: ptr(helpers.QuantityMmHg(120)),
    }),
    r4.WithObservationComponent(r4.ObservationComponent{
        Code:          helpers.DiastolicBP(),        // LOINC 8462-4
        ValueQuantity: ptr(helpers.QuantityMmHg(80)),
    }),
)
```

**JSON Output:**

```json
{
  "resourceType": "Observation",
  "id": "vitals-bp-001",
  "status": "final",
  "code": {
    "coding": [
      {
        "system": "http://loinc.org",
        "code": "85354-9",
        "display": "Blood pressure panel"
      }
    ],
    "text": "Blood pressure panel"
  },
  "subject": {
    "reference": "Patient/patient-123"
  },
  "effectiveDateTime": "2024-01-15T10:30:00Z",
  "component": [
    {
      "code": {
        "coding": [
          {
            "system": "http://loinc.org",
            "code": "8480-6",
            "display": "Systolic blood pressure"
          }
        ]
      },
      "valueQuantity": {
        "value": 120,
        "unit": "mmHg",
        "system": "http://unitsofmeasure.org",
        "code": "mm[Hg]"
      }
    },
    {
      "code": {
        "coding": [
          {
            "system": "http://loinc.org",
            "code": "8462-4",
            "display": "Diastolic blood pressure"
          }
        ]
      },
      "valueQuantity": {
        "value": 80,
        "unit": "mmHg",
        "system": "http://unitsofmeasure.org",
        "code": "mm[Hg]"
      }
    }
  ]
}
```

## Code Generation

The R4, R4B, and R5 packages are auto-generated from FHIR StructureDefinitions. To regenerate:

```bash
# Generate R4 resources
go generate ./pkg/fhir/r4/...

# Generate all versions
go generate ./pkg/fhir/...
```

## License

See repository root for license information.
