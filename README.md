# GoFHIR

[![CI](https://github.com/robertoaraneda/gofhir/actions/workflows/ci.yml/badge.svg)](https://github.com/robertoaraneda/gofhir/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/robertoaraneda/gofhir.svg)](https://pkg.go.dev/github.com/robertoaraneda/gofhir)
[![codecov](https://codecov.io/gh/robertoaraneda/gofhir/branch/main/graph/badge.svg)](https://codecov.io/gh/robertoaraneda/gofhir)

Production-grade FHIR toolkit for Go.

## Features

- **Strongly-typed resources**: All FHIR R4, R4B, and R5 resources as Go structs
- **Multi-version abstraction**: Common interfaces for version-agnostic code
- **Fluent builders**: Construct resources with a fluent API
- **FHIRPath engine**: Complete FHIRPath 2.0 implementation with UCUM support
- **Validation**: Validate resources against StructureDefinitions with FHIRPath constraints
- **UCUM normalization**: Unit conversion for quantity comparisons
- **Full extension support**: Primitive elements support extensions via `_field` pattern
- **JSON field ordering**: Typed structs guarantee FHIR-compliant field order

## Installation

```bash
go get github.com/robertoaraneda/gofhir
```

## Quick Start

### Creating Resources

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/robertoaraneda/gofhir/pkg/fhir/common"
    "github.com/robertoaraneda/gofhir/pkg/fhir/r4"
)

func main() {
    // Create a patient using the builder
    patient := r4.NewPatientBuilder().
        SetId("example-001").
        SetActive(true).
        AddName(r4.HumanName{
            Family: common.String("Garcia"),
            Given:  []string{"Maria"},
        }).
        SetGender("female").
        SetBirthDate("1985-03-15").
        Build()

    data, _ := json.MarshalIndent(patient, "", "  ")
    fmt.Println(string(data))
}
```

### Evaluating FHIRPath Expressions

```go
package main

import (
    "fmt"
    "github.com/robertoaraneda/gofhir/pkg/fhirpath"
)

func main() {
    patient := []byte(`{
        "resourceType": "Patient",
        "name": [{"family": "Doe", "given": ["John"]}],
        "birthDate": "1990-05-15"
    }`)

    // Simple evaluation
    result, _ := fhirpath.Evaluate(patient, "Patient.name.family")
    fmt.Println(result) // ["Doe"]

    // Compile once, evaluate many times
    expr := fhirpath.MustCompile("birthDate > @1980-01-01")
    result = expr.Evaluate(patient)
    fmt.Println(result) // [true]

    // UCUM unit comparison
    result, _ = fhirpath.Evaluate(patient, "1000 'mg' = 1 'g'")
    fmt.Println(result) // [true]
}
```

### Validating Resources

```go
package main

import (
    "context"
    "fmt"
    "github.com/robertoaraneda/gofhir/pkg/validator"
)

func main() {
    // Create validator with embedded R4 definitions
    registry := validator.NewEmbeddedRegistry("R4")
    opts := validator.DefaultValidatorOptions()
    v := validator.NewValidator(registry, opts)

    patient := []byte(`{
        "resourceType": "Patient",
        "id": "123",
        "gender": "invalid-gender"
    }`)

    result, _ := v.Validate(context.Background(), patient)

    if !result.Valid {
        for _, issue := range result.Issues {
            fmt.Printf("[%s] %s: %s\n",
                issue.Severity, issue.Code, issue.Diagnostics)
        }
    }
}
```

## Packages

| Package | Description | Documentation |
|---------|-------------|---------------|
| [`pkg/fhir`](pkg/fhir/) | Multi-version abstraction with ResourceFactory, Resource, Meta interfaces | [README](pkg/fhir/README.md) |
| [`pkg/fhir/r4`](pkg/fhir/r4/) | FHIR R4 (4.0.1) types, builders, and helpers | [README](pkg/fhir/README.md) |
| [`pkg/fhir/r4b`](pkg/fhir/r4b/) | FHIR R4B (4.3.0) types, builders, and helpers | [README](pkg/fhir/README.md) |
| [`pkg/fhir/r5`](pkg/fhir/r5/) | FHIR R5 (5.0.0) types, builders, and helpers | [README](pkg/fhir/README.md) |
| [`pkg/fhirpath`](pkg/fhirpath/) | FHIRPath 2.0 expression evaluator | [README](pkg/fhirpath/README.md) |
| [`pkg/validator`](pkg/validator/) | Resource validation against StructureDefinitions | [README](pkg/validator/README.md) |
| [`pkg/ucum`](pkg/ucum/) | UCUM unit normalization | - |
| [`pkg/common`](pkg/common/) | Shared utilities (pointer helpers, cloning) | - |

## FHIRPath Engine

Complete implementation of the [FHIRPath Normative Release 2.0.0](https://hl7.org/fhirpath/).

### Highlights

- **40+ functions**: String, math, existence, filtering, conversion, temporal, and more
- **All operators**: Arithmetic, comparison, boolean, collection, and type operators
- **UCUM normalization**: Compare quantities with different units (`1 'kg' = 1000 'g'`)
- **Lazy evaluation**: `iif()` only evaluates the matching branch
- **Polymorphic elements**: Automatic resolution of `value[x]` patterns
- **Environment variables**: `%resource`, `%context` support

```go
// Examples of supported expressions
fhirpath.Evaluate(obs, "Observation.value.ofType(Quantity)")
fhirpath.Evaluate(patient, "Patient.name.where(use='official').given.first()")
fhirpath.Evaluate(resource, "iif(active, 'Active', 'Inactive')")
fhirpath.Evaluate(resource, "100 'cm' = 1 'm'")  // true
fhirpath.Evaluate(resource, "'Hello' ~ 'hello'") // true (case-insensitive)
```

See [FHIRPath README](pkg/fhirpath/README.md) for complete documentation.

## Validation

Comprehensive validation against StructureDefinitions:

- **Structural validation**: Element presence and cardinality
- **Type validation**: Field type conformance
- **Primitive validation**: Date, dateTime, instant, time formats
- **FHIRPath constraints**: Evaluate invariants from StructureDefinitions
- **Terminology validation**: Check codes against ValueSets
- **Reference validation**: Resolve and validate references
- **Extension validation**: Validate extension structure

```go
opts := validator.ValidatorOptions{
    ValidateConstraints: true,
    ValidateTerminology: true,
    TerminologyService:  validator.TerminologyEmbeddedR4,
    ValidateReferences:  true,
    ValidateExtensions:  true,
}

v := validator.NewValidator(registry, opts)
result, _ := v.Validate(ctx, resource)
```

See [Validator README](pkg/validator/README.md) for complete documentation.

## UCUM Unit Support

The toolkit includes UCUM (Unified Code for Units of Measure) normalization for quantity comparisons:

| Dimension | Canonical | Supported Units |
|-----------|-----------|-----------------|
| Mass | `g` | kg, g, mg, ug, ng, pg, lb, oz |
| Length | `m` | km, m, dm, cm, mm, um, nm, in, ft |
| Volume | `L` | L, dL, cL, mL, uL, gal, qt, pt |
| Time | `s` | a, mo, wk, d, h, min, s, ms, us, ns |
| Concentration | `g/L` | g/L, mg/L, ug/L, g/dL, mg/dL, g/mL |
| Molar | `mol/L` | mol/L, mmol/L, umol/L, nmol/L |
| Pressure | `Pa` | Pa, kPa, mm[Hg], psi |

```go
// FHIRPath uses UCUM normalization automatically
fhirpath.Evaluate(resource, "1000 'mg' = 1 'g'")   // true
fhirpath.Evaluate(resource, "60 'min' = 1 'h'")    // true
fhirpath.Evaluate(resource, "100 'cm' < 2 'm'")    // true

// Direct UCUM normalization
normalized := ucum.Normalize(500, "mg")
// normalized.Value = 0.5, normalized.Code = "g"
```

## FHIR Resource Types

### R4 Resources (150+)

| Category | Examples |
|----------|----------|
| **Clinical** | Patient, Practitioner, Organization, Encounter, Condition |
| **Diagnostic** | Observation, DiagnosticReport, Specimen, ImagingStudy |
| **Medication** | Medication, MedicationRequest, MedicationAdministration, Immunization |
| **Financial** | Claim, ClaimResponse, Coverage, Invoice |
| **Workflow** | Task, Appointment, Schedule, ServiceRequest |
| **Infrastructure** | Bundle, OperationOutcome, StructureDefinition, ValueSet |

### Building Resources

```go
// Builder pattern
patient := r4.NewPatientBuilder().
    SetId("123").
    SetActive(true).
    AddIdentifier(r4.Identifier{
        System: common.String("http://hospital.org/mrn"),
        Value:  common.String("MRN-12345"),
    }).
    Build()

// Functional options
patient := r4.NewPatient(
    r4.WithPatientId("123"),
    r4.WithPatientActive(true),
)

// Direct struct
patient := &r4.Patient{
    ResourceType: "Patient",
    Id:           common.String("123"),
    Active:       common.Bool(true),
}
```

### Helper Functions

```go
import "github.com/robertoaraneda/gofhir/pkg/fhir/r4/helpers"

// LOINC codes for vital signs
heartRateCode := helpers.HeartRate()       // 8867-4
bodyTempCode := helpers.BodyTemperature()  // 8310-5
bloodPressure := helpers.BloodPressure()   // 85354-9

// UCUM quantities
weight := helpers.QuantityKg(72.5)         // 72.5 kg
height := helpers.QuantityCm(175)          // 175 cm
temp := helpers.QuantityCelsius(37.2)      // 37.2 Cel
```

## CLI

```bash
# Validate a resource
gofhir validate patient.json

# Evaluate FHIRPath
gofhir fhirpath "name.given.first()" patient.json

# Generate types from specs
gofhir generate --specs ./specs/r4 --output ./pkg/fhir/r4
```

## Development

### Prerequisites

- Go 1.23+
- golangci-lint

### Setup

```bash
# Clone the repository
git clone https://github.com/robertoaraneda/gofhir.git
cd gofhir

# Download dependencies
make deps

# Run tests
make test

# Run linter
make lint
```

### Project Structure

```text
gofhir/
├── cmd/gofhir/          # CLI application
├── pkg/
│   ├── fhir/            # FHIR types and multi-version abstraction
│   │   ├── fhir.go      # Common interfaces (Resource, Meta, ResourceFactory)
│   │   ├── r4/          # R4 types, builders, helpers
│   │   ├── r4b/         # R4B types, builders, helpers
│   │   ├── r5/          # R5 types, builders, helpers
│   │   └── common/      # Shared utilities
│   ├── fhirpath/        # FHIRPath 2.0 engine
│   │   ├── eval/        # Expression evaluator
│   │   ├── funcs/       # 40+ function implementations
│   │   ├── types/       # Type system
│   │   └── parser/      # ANTLR4-generated parser
│   ├── validator/       # Resource validation
│   │   ├── validator.go # Main validator
│   │   └── terminology*.go # Embedded terminology
│   ├── ucum/            # Unit normalization
│   └── common/          # Shared utilities
├── internal/
│   └── codegen/         # Code generation tools
└── specs/               # FHIR specifications
```

## Specification Compliance

### FHIRPath

Implements **FHIRPath Normative Release 2.0.0**:

- [x] Full type system (Boolean, Integer, Decimal, String, Date, DateTime, Time, Quantity)
- [x] All operators (arithmetic, comparison, boolean, collection, type)
- [x] All standard functions (40+ functions)
- [x] UCUM unit normalization for Quantity comparisons
- [x] Three-valued logic (empty propagation)
- [x] Lazy evaluation for `iif()`
- [x] Polymorphic element resolution (value[x])
- [x] Environment variables (%resource, %context)
- [x] Delimited identifiers (backticks)

### FHIR

Supports **FHIR R4 (4.0.1)**, **R4B (4.3.0)**, and **R5 (5.0.0)**:

- [x] All resource types (150+ per version)
- [x] All data types (50+ complex types)
- [x] Extensions on primitive elements
- [x] Contained resources
- [x] Bundle handling

### Validator

- [x] Structural validation (cardinality, types)
- [x] FHIRPath constraint evaluation
- [x] Terminology binding validation (required, extensible, preferred)
- [x] Reference validation
- [x] Extension validation
- [x] Embedded terminology for R4, R4B, R5

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
