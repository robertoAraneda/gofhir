# FHIR

Version-agnostic FHIR resource definitions and factories for Go.

## Overview

This package provides strongly-typed FHIR resource definitions with support for multiple FHIR versions (R4, R4B, R5). It includes auto-generated resource types, fluent builders, and functional options for creating FHIR resources.

## Installation

```go
import "github.com/robertoaraneda/gofhir/pkg/fhir"
```

For version-specific types:

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

    "github.com/robertoaraneda/gofhir/pkg/fhir"
    "github.com/robertoaraneda/gofhir/pkg/fhir/r4"
    "github.com/robertoaraneda/gofhir/pkg/fhir/common"
)

func main() {
    // Using Builder Pattern
    patient := r4.NewPatientBuilder().
        SetId("patient-123").
        SetActive(true).
        AddName(r4.HumanName{
            Use:    common.String("official"),
            Family: common.String("Doe"),
            Given:  []string{"John", "Michael"},
        }).
        AddTelecom(r4.ContactPoint{
            System: common.String("email"),
            Value:  common.String("john.doe@example.com"),
        }).
        Build()

    // Serialize to JSON
    data, _ := json.MarshalIndent(patient, "", "  ")
    fmt.Println(string(data))
}
```

## Version-Agnostic API

### Factory Pattern

```go
// Register factory (done automatically by version packages)
fhir.RegisterFactory(r4.NewFactory())

// Get factory for specific version
factory, err := fhir.GetFactory(fhir.R4)
if err != nil {
    panic(err)
}

// Create resource dynamically
resource, err := factory.NewResource("Patient")
patient := resource.(*r4.Patient)

// Unmarshal from JSON
resource, err = factory.UnmarshalResource(jsonData)
```

### Supported Versions

| Version | Constant | FHIR Version |
|---------|----------|--------------|
| R4 | `fhir.R4` | 4.0.1 |
| R4B | `fhir.R4B` | 4.3.0 |
| R5 | `fhir.R5` | 5.0.0 |

```go
// Check version support
if fhir.IsVersionSupported(fhir.R4) {
    // R4 is available
}

// List all supported versions
versions := fhir.SupportedVersions() // [R4, R4B, R5]
```

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
        System: common.String("http://hospital.org/mrn"),
        Value:  common.String("MRN-12345"),
    }).
    AddName(r4.HumanName{
        Use:    common.String("official"),
        Family: common.String("Doe"),
        Given:  []string{"John"},
    }).
    AddAddress(r4.Address{
        Use:        common.String("home"),
        Line:       []string{"123 Main St"},
        City:       common.String("Springfield"),
        State:      common.String("IL"),
        PostalCode: common.String("62701"),
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
    Id:           common.String("123"),
    Active:       common.Bool(true),
    Gender:       common.String("male"),
    Name: []r4.HumanName{
        {
            Family: common.String("Doe"),
            Given:  []string{"John"},
        },
    },
}
```

## Data Types

### Primitive Types

| Type | Go Type | Example |
|------|---------|---------|
| `string` | `*string` | `common.String("value")` |
| `boolean` | `*bool` | `common.Bool(true)` |
| `integer` | `*int` | `common.Int(42)` |
| `decimal` | `*float64` | `common.Float64(3.14)` |
| `date` | `*string` | `common.String("2024-01-15")` |
| `dateTime` | `*string` | `common.String("2024-01-15T10:30:00Z")` |
| `instant` | `*string` | `common.String("2024-01-15T10:30:00.000Z")` |
| `uri` | `*string` | `common.String("http://example.com")` |
| `code` | `*string` | `common.String("active")` |

### Complex Types

```go
// CodeableConcept
codeableConcept := r4.CodeableConcept{
    Coding: []r4.Coding{
        {
            System:  common.String("http://loinc.org"),
            Code:    common.String("8867-4"),
            Display: common.String("Heart rate"),
        },
    },
    Text: common.String("Heart rate"),
}

// Quantity
quantity := r4.Quantity{
    Value:  common.Float64(72),
    Unit:   common.String("beats/min"),
    System: common.String("http://unitsofmeasure.org"),
    Code:   common.String("/min"),
}

// Reference
reference := r4.Reference{
    Reference: common.String("Patient/123"),
    Display:   common.String("John Doe"),
}

// Period
period := r4.Period{
    Start: common.String("2024-01-01T00:00:00Z"),
    End:   common.String("2024-12-31T23:59:59Z"),
}

// Identifier
identifier := r4.Identifier{
    Use:    common.String("official"),
    System: common.String("http://hospital.org/mrn"),
    Value:  common.String("MRN-12345"),
}
```

## Helper Functions

### Common Pointer Helpers

```go
import "github.com/robertoaraneda/gofhir/pkg/fhir/common"

// Create pointers
strPtr := common.String("value")
boolPtr := common.Bool(true)
intPtr := common.Int(42)
floatPtr := common.Float64(3.14)

// Get values (with defaults)
str := common.StringVal(strPtr)       // "value"
str := common.StringVal(nil)          // ""
b := common.BoolVal(boolPtr)          // true
b := common.BoolVal(nil)              // false
```

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

### Creating a Search Bundle

```go
factory, _ := fhir.GetFactory(fhir.R4)

bundle, _ := factory.BuildSearchBundle(fhir.SearchBundleConfig{
    Total:  100,
    Count:  10,
    Offset: 0,
    SelfURL: "http://example.com/Patient?_count=10",
    NextURL: "http://example.com/Patient?_count=10&_offset=10",
    Entries: patients, // []Resource
})
```

### Creating an OperationOutcome

```go
outcome, _ := factory.BuildOperationOutcome(fhir.OutcomeConfig{
    Severity:    "error",
    Code:        "invalid",
    Diagnostics: "Patient.birthDate is required",
    Expression:  []string{"Patient.birthDate"},
})
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

// Or use factory for dynamic unmarshaling
factory, _ := fhir.GetFactory(fhir.R4)
resource, err := factory.UnmarshalResource(data)
switch r := resource.(type) {
case *r4.Patient:
    fmt.Printf("Patient: %s\n", *r.Id)
case *r4.Observation:
    fmt.Printf("Observation: %s\n", *r.Id)
}
```

## Resource Interfaces

```go
// All resources implement these interfaces
type Resource interface {
    GetResourceType() string
    GetID() *string
    SetID(string)
    GetMeta() Meta
    SetMeta(Meta)
}

// Usage
func processResource(r fhir.Resource) {
    fmt.Printf("Type: %s, ID: %s\n",
        r.GetResourceType(),
        common.StringVal(r.GetID()))
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
        Reference: common.String("Patient/123"),
    }).
    SetEffectiveDateTime("2024-01-15T10:30:00Z").
    SetValueQuantity(r4.Quantity{
        Value:  common.Float64(72),
        Unit:   common.String("beats/min"),
        System: common.String("http://unitsofmeasure.org"),
        Code:   common.String("/min"),
    }).
    Build()
```

### Creating a Bundle with Multiple Resources

```go
bundle := r4.NewBundleBuilder().
    SetId("bundle-123").
    SetType("collection").
    AddEntry(r4.BundleEntry{
        FullUrl:  common.String("urn:uuid:patient-1"),
        Resource: patient,
    }).
    AddEntry(r4.BundleEntry{
        FullUrl:  common.String("urn:uuid:obs-1"),
        Resource: observation,
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
                System:  common.String("http://www.nlm.nih.gov/research/umls/rxnorm"),
                Code:    common.String("1049502"),
                Display: common.String("Acetaminophen 325 MG Oral Tablet"),
            },
        },
    }).
    SetSubject(r4.Reference{
        Reference: common.String("Patient/123"),
    }).
    Build()
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
