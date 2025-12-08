# Professional Prompt: FHIR Toolkit Implementation in Go with FHIRPath Support

## Executive Summary

You are tasked with building a **production-grade FHIR (Fast Healthcare Interoperability Resources) Toolkit in Go** that provides complete type safety, fluent builders for resource construction, comprehensive validation with FHIRPath support, and multi-version compatibility (R4, R4B, R5). This implementation should mirror the architecture and capabilities of an existing TypeScript FHIR Toolkit while leveraging Go's strengths in performance, concurrency, and strong typing.

---

## Project Overview

### Goals
1. Create a comprehensive Go library for working with FHIR resources
2. Provide strongly-typed structs for all FHIR resources, datatypes, and backbone elements
3. Implement a fluent builder pattern for constructing FHIR resources
4. Build a robust validation system with FHIRPath expression evaluation
5. Support FHIR versions R4 (4.0.1), R4B (4.3.0), and R5 (5.0.0)
6. Implement a code generation pipeline from FHIR StructureDefinitions

### Target Scale
- **147+ FHIR Resources** per version (Patient, Observation, Bundle, etc.)
- **42+ Complex Datatypes** (Coding, CodeableConcept, Period, Quantity, etc.)
- **560+ ValueSet bindings** with type-safe code literals
- **Hundreds of Backbone Elements** for nested structures

---

## Part 1: Project Structure

### Recommended Module Organization

```
fhir-toolkit-go/
├── go.mod
├── go.sum
├── Makefile
├── README.md
│
├── pkg/
│   ├── r4/                          # FHIR R4 (4.0.1)
│   │   ├── types/                   # Go interfaces/structs
│   │   │   ├── base.go              # IElement, IResource, IDomainResource
│   │   │   ├── primitives.go        # Primitive type definitions
│   │   │   ├── datatypes/           # Complex datatypes
│   │   │   │   ├── coding.go
│   │   │   │   ├── codeableconcept.go
│   │   │   │   ├── period.go
│   │   │   │   ├── quantity.go
│   │   │   │   ├── reference.go
│   │   │   │   └── ... (42+ files)
│   │   │   ├── resources/           # Resource interfaces
│   │   │   │   ├── patient.go
│   │   │   │   ├── observation.go
│   │   │   │   ├── bundle.go
│   │   │   │   └── ... (147 files)
│   │   │   ├── backbones/           # Backbone element types
│   │   │   │   ├── patient_contact.go
│   │   │   │   ├── observation_component.go
│   │   │   │   └── ...
│   │   │   └── valuesets/           # ValueSet type literals
│   │   │       ├── administrative_gender.go
│   │   │       ├── observation_status.go
│   │   │       └── ... (560+ files)
│   │   │
│   │   ├── models/                  # Concrete model implementations
│   │   │   ├── base/
│   │   │   │   ├── element.go
│   │   │   │   ├── resource.go
│   │   │   │   └── domain_resource.go
│   │   │   ├── datatypes/
│   │   │   ├── resources/
│   │   │   └── backbones/
│   │   │
│   │   ├── builders/                # Fluent builders
│   │   │   ├── base/
│   │   │   │   ├── element_builder.go
│   │   │   │   ├── resource_builder.go
│   │   │   │   └── domain_resource_builder.go
│   │   │   ├── datatypes/
│   │   │   ├── resources/
│   │   │   └── backbones/
│   │   │
│   │   └── specs/                   # Embedded StructureDefinitions
│   │       ├── resources/
│   │       ├── datatypes/
│   │       └── valuesets/
│   │
│   ├── r4b/                         # FHIR R4B (4.3.0) - same structure
│   ├── r5/                          # FHIR R5 (5.0.0) - same structure
│   │
│   ├── fhirpath/                    # FHIRPath implementation
│   │   ├── parser/                  # FHIRPath grammar parser
│   │   │   ├── lexer.go
│   │   │   ├── parser.go
│   │   │   ├── ast.go
│   │   │   └── grammar.go
│   │   ├── evaluator/               # Expression evaluator
│   │   │   ├── evaluator.go
│   │   │   ├── functions.go         # Built-in FHIRPath functions
│   │   │   ├── operators.go         # Operators (+, -, and, or, etc.)
│   │   │   └── context.go           # Evaluation context
│   │   ├── compiler/                # Expression compiler
│   │   │   ├── compiler.go
│   │   │   └── cache.go             # LRU cache for compiled expressions
│   │   └── fhirpath.go              # Public API
│   │
│   ├── validator/                   # YAFV (Yet Another FHIR Validator)
│   │   ├── validator.go             # Main validator
│   │   ├── options.go               # Validation options
│   │   ├── result.go                # OperationOutcome result
│   │   ├── spec_registry.go         # StructureDefinition registry
│   │   ├── ig_loader.go             # Implementation Guide loader
│   │   ├── validators/
│   │   │   ├── primitive.go         # Primitive type validation
│   │   │   ├── structure.go         # Structural validation
│   │   │   ├── constraint.go        # FHIRPath constraint validation
│   │   │   ├── terminology.go       # ValueSet/CodeSystem validation
│   │   │   ├── extension.go         # Extension validation
│   │   │   ├── reference.go         # Reference validation
│   │   │   ├── slicing.go           # Slice discrimination
│   │   │   ├── bundle.go            # Bundle transaction validation
│   │   │   └── contained.go         # Contained resource validation
│   │   └── cache/
│   │       ├── lru.go               # LRU cache implementation
│   │       └── terminology_cache.go # Terminology lookup cache
│   │
│   └── common/                      # Shared utilities
│       ├── json.go                  # JSON serialization helpers
│       ├── clone.go                 # Deep cloning utilities
│       └── errors.go                # Error types
│
├── cmd/
│   ├── fhir-cli/                    # Command-line interface
│   │   └── main.go
│   └── codegen/                     # Code generation tool
│       └── main.go
│
├── internal/
│   └── codegen/                     # Code generation implementation
│       ├── parser.go                # StructureDefinition parser
│       ├── generator.go             # Main generator
│       ├── type_generator.go        # Type/struct generation
│       ├── model_generator.go       # Model class generation
│       ├── builder_generator.go     # Builder generation
│       └── templates/               # Go templates
│           ├── struct.tmpl
│           ├── model.tmpl
│           ├── builder.tmpl
│           └── valueset.tmpl
│
├── specs/                           # Raw FHIR specifications (JSON)
│   ├── r4/
│   ├── r4b/
│   └── r5/
│
├── examples/
│   ├── basic_usage/
│   ├── builders/
│   ├── validation/
│   └── fhirpath/
│
└── test/
    ├── integration/
    └── fixtures/
```

---

## Part 2: Type System Implementation

### 2.1 Base Types

```go
// pkg/r4/types/base.go

package types

// Element is the base for all FHIR elements
type Element struct {
    ID        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
}

// Resource is the base for all FHIR resources
type Resource struct {
    ResourceType string  `json:"resourceType"`
    ID           *string `json:"id,omitempty"`
    Meta         *Meta   `json:"meta,omitempty"`
    ImplicitRules *string `json:"implicitRules,omitempty"`
    Language      *string `json:"language,omitempty"`
}

// DomainResource extends Resource with narrative and extensions
type DomainResource struct {
    Resource
    Text              *Narrative  `json:"text,omitempty"`
    Contained         []Resource  `json:"contained,omitempty"`
    Extension         []Extension `json:"extension,omitempty"`
    ModifierExtension []Extension `json:"modifierExtension,omitempty"`
}

// BackboneElement is the base for complex nested elements
type BackboneElement struct {
    Element
    ModifierExtension []Extension `json:"modifierExtension,omitempty"`
}
```

### 2.2 Primitive Types with Extensions

```go
// pkg/r4/types/primitives.go

package types

// FHIR primitives support extensions via underscore-prefixed elements
// Example: birthDate has _birthDate for extensions

// StringElement represents a string with optional extension
type StringElement struct {
    Value     *string  `json:"-"`
    Extension *Element `json:"-"`
}

// Custom JSON marshaling to handle the FHIR primitive extension pattern
func (s StringElement) MarshalJSON() ([]byte, error) {
    // Implement FHIR-compliant serialization
}

// DateTimeElement, IntegerElement, BooleanElement, etc.
// Follow the same pattern
```

### 2.3 Choice Type Pattern

```go
// pkg/r4/types/choice.go

package types

// ChoiceType represents FHIR's value[x] pattern
// Use interfaces and type switches for type safety

type ObservationValue interface {
    isObservationValue()
}

// All valid types implement the marker interface
func (Quantity) isObservationValue()         {}
func (CodeableConcept) isObservationValue()  {}
func (String) isObservationValue()           {}
func (Boolean) isObservationValue()          {}
func (Integer) isObservationValue()          {}
func (Range) isObservationValue()            {}
func (Ratio) isObservationValue()            {}
func (SampledData) isObservationValue()      {}
func (Time) isObservationValue()             {}
func (DateTime) isObservationValue()         {}
func (Period) isObservationValue()           {}
```

### 2.4 Reference Type with Generic Constraints

```go
// pkg/r4/types/datatypes/reference.go

package datatypes

// Reference represents a FHIR Reference with type constraints
// Go doesn't have generics like TypeScript, so use runtime validation
// or code generation for type-safe references

type Reference struct {
    Element
    Reference  *string     `json:"reference,omitempty"`
    Type       *string     `json:"type,omitempty"`
    Identifier *Identifier `json:"identifier,omitempty"`
    Display    *string     `json:"display,omitempty"`
}

// For type-safe references, generate specific types:
type PatientReference = Reference
type PractitionerReference = Reference

// Or use a validation approach:
func (r *Reference) ValidateTargetTypes(allowed ...string) error {
    // Validate that r.Type is in allowed list
}
```

### 2.5 ValueSet Type Literals

```go
// pkg/r4/types/valuesets/administrative_gender.go

package valuesets

type AdministrativeGender string

const (
    AdministrativeGenderMale    AdministrativeGender = "male"
    AdministrativeGenderFemale  AdministrativeGender = "female"
    AdministrativeGenderOther   AdministrativeGender = "other"
    AdministrativeGenderUnknown AdministrativeGender = "unknown"
)

func (g AdministrativeGender) IsValid() bool {
    switch g {
    case AdministrativeGenderMale, AdministrativeGenderFemale,
         AdministrativeGenderOther, AdministrativeGenderUnknown:
        return true
    }
    return false
}

// pkg/r4/types/valuesets/observation_status.go

type ObservationStatus string

const (
    ObservationStatusRegistered  ObservationStatus = "registered"
    ObservationStatusPreliminary ObservationStatus = "preliminary"
    ObservationStatusFinal       ObservationStatus = "final"
    ObservationStatusAmended     ObservationStatus = "amended"
    ObservationStatusCorrected   ObservationStatus = "corrected"
    ObservationStatusCancelled   ObservationStatus = "cancelled"
    ObservationStatusEnteredInError ObservationStatus = "entered-in-error"
    ObservationStatusUnknown     ObservationStatus = "unknown"
)
```

---

## Part 3: Model Implementation

### 3.1 Base Model Classes

```go
// pkg/r4/models/base/element.go

package base

import (
    "encoding/json"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"
)

type ElementModel struct {
    data types.Element
}

func (e *ElementModel) ID() *string {
    return e.data.ID
}

func (e *ElementModel) SetID(id string) {
    e.data.ID = &id
}

func (e *ElementModel) Extensions() []types.Extension {
    return e.data.Extension
}

func (e *ElementModel) AddExtension(ext types.Extension) {
    e.data.Extension = append(e.data.Extension, ext)
}

func (e *ElementModel) Clone() *ElementModel {
    clone := &ElementModel{}
    // Deep clone via JSON round-trip
    data, _ := json.Marshal(e.data)
    json.Unmarshal(data, &clone.data)
    return clone
}

func (e *ElementModel) ToJSON() ([]byte, error) {
    return json.Marshal(e.data)
}
```

### 3.2 Resource Model Example

```go
// pkg/r4/models/resources/patient.go

package resources

import (
    "encoding/json"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types/valuesets"
)

// Patient FHIR resource model
type Patient struct {
    data types.Patient
}

// NewPatient creates a new Patient from optional partial data
func NewPatient(data *types.Patient) *Patient {
    p := &Patient{}
    if data != nil {
        p.data = *data
    }
    p.data.ResourceType = "Patient"
    return p
}

// Property accessors
func (p *Patient) ID() *string                        { return p.data.ID }
func (p *Patient) Active() *bool                      { return p.data.Active }
func (p *Patient) Gender() *valuesets.AdministrativeGender { return p.data.Gender }
func (p *Patient) BirthDate() *string                 { return p.data.BirthDate }
func (p *Patient) Names() []types.HumanName           { return p.data.Name }
func (p *Patient) Identifiers() []types.Identifier    { return p.data.Identifier }
func (p *Patient) Telecoms() []types.ContactPoint     { return p.data.Telecom }
func (p *Patient) Addresses() []types.Address         { return p.data.Address }

// Setters
func (p *Patient) SetID(id string)       { p.data.ID = &id }
func (p *Patient) SetActive(active bool) { p.data.Active = &active }
func (p *Patient) SetGender(gender valuesets.AdministrativeGender) {
    p.data.Gender = &gender
}
func (p *Patient) SetBirthDate(date string) { p.data.BirthDate = &date }

func (p *Patient) AddName(name types.HumanName) {
    p.data.Name = append(p.data.Name, name)
}

func (p *Patient) AddIdentifier(id types.Identifier) {
    p.data.Identifier = append(p.data.Identifier, id)
}

// Immutable update - returns new instance
func (p *Patient) With(modifier func(*Patient)) *Patient {
    clone := p.Clone()
    modifier(clone)
    return clone
}

// Deep clone
func (p *Patient) Clone() *Patient {
    data, _ := json.Marshal(p.data)
    var cloned types.Patient
    json.Unmarshal(data, &cloned)
    return &Patient{data: cloned}
}

// Serialization with FHIR property ordering
func (p *Patient) ToJSON() ([]byte, error) {
    return json.Marshal(p.data)
}

// MarshalJSON implements custom ordering
func (p *Patient) MarshalJSON() ([]byte, error) {
    // Implement FHIR-compliant property ordering
    return p.marshalOrdered()
}

// Property order constant (FHIR specification order)
var patientPropertyOrder = []string{
    "resourceType", "id", "meta", "implicitRules", "language",
    "text", "contained", "extension", "modifierExtension",
    "identifier", "active", "name", "telecom", "gender",
    "birthDate", "deceased", "address", "maritalStatus",
    "multipleBirth", "photo", "contact", "communication",
    "generalPractitioner", "managingOrganization", "link",
}

func (p *Patient) marshalOrdered() ([]byte, error) {
    // Serialize properties in FHIR-defined order
    // Implementation omitted for brevity
}

// Validation integration (lazy loading)
func (p *Patient) Validate() (*types.OperationOutcome, error) {
    // Integrate with validator package
}

func (p *Patient) ValidateOrPanic() {
    outcome, err := p.Validate()
    if err != nil {
        panic(err)
    }
    if hasErrors(outcome) {
        panic("validation failed")
    }
}
```

### 3.3 Choice Type Handling in Models

```go
// pkg/r4/models/resources/observation.go

package resources

import "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"

type Observation struct {
    data types.Observation
}

// Value[x] getter - returns the set value with its type
func (o *Observation) Value() (interface{}, string) {
    switch {
    case o.data.ValueQuantity != nil:
        return o.data.ValueQuantity, "Quantity"
    case o.data.ValueCodeableConcept != nil:
        return o.data.ValueCodeableConcept, "CodeableConcept"
    case o.data.ValueString != nil:
        return *o.data.ValueString, "String"
    case o.data.ValueBoolean != nil:
        return *o.data.ValueBoolean, "Boolean"
    case o.data.ValueInteger != nil:
        return *o.data.ValueInteger, "Integer"
    case o.data.ValueRange != nil:
        return o.data.ValueRange, "Range"
    case o.data.ValueRatio != nil:
        return o.data.ValueRatio, "Ratio"
    case o.data.ValueSampledData != nil:
        return o.data.ValueSampledData, "SampledData"
    case o.data.ValueTime != nil:
        return *o.data.ValueTime, "Time"
    case o.data.ValueDateTime != nil:
        return *o.data.ValueDateTime, "DateTime"
    case o.data.ValuePeriod != nil:
        return o.data.ValuePeriod, "Period"
    }
    return nil, ""
}

// Type-safe setters for choice types
func (o *Observation) SetValueQuantity(q *types.Quantity) {
    o.clearValue()
    o.data.ValueQuantity = q
}

func (o *Observation) SetValueCodeableConcept(cc *types.CodeableConcept) {
    o.clearValue()
    o.data.ValueCodeableConcept = cc
}

func (o *Observation) SetValueString(s string) {
    o.clearValue()
    o.data.ValueString = &s
}

// Clear all value[x] fields before setting new one
func (o *Observation) clearValue() {
    o.data.ValueQuantity = nil
    o.data.ValueCodeableConcept = nil
    o.data.ValueString = nil
    o.data.ValueBoolean = nil
    o.data.ValueInteger = nil
    o.data.ValueRange = nil
    o.data.ValueRatio = nil
    o.data.ValueSampledData = nil
    o.data.ValueTime = nil
    o.data.ValueDateTime = nil
    o.data.ValuePeriod = nil
}
```

---

## Part 4: Builder Pattern Implementation

### 4.1 Base Builder

```go
// pkg/r4/builders/base/element_builder.go

package base

import "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"

type ElementBuilder[T any] struct {
    data T
}

func (b *ElementBuilder[T]) SetID(id string) *ElementBuilder[T] {
    // Set ID on data
    return b
}

func (b *ElementBuilder[T]) AddExtension(ext types.Extension) *ElementBuilder[T] {
    // Add extension to data
    return b
}
```

### 4.2 Resource Builder

```go
// pkg/r4/builders/resources/patient_builder.go

package resources

import (
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types/valuesets"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/models/resources"
)

type PatientBuilder struct {
    data types.Patient
}

func NewPatientBuilder() *PatientBuilder {
    return &PatientBuilder{
        data: types.Patient{
            ResourceType: "Patient",
        },
    }
}

// Fluent setters - all return *PatientBuilder for chaining

func (b *PatientBuilder) SetID(id string) *PatientBuilder {
    b.data.ID = &id
    return b
}

func (b *PatientBuilder) SetMeta(meta types.Meta) *PatientBuilder {
    b.data.Meta = &meta
    return b
}

func (b *PatientBuilder) SetActive(active bool) *PatientBuilder {
    b.data.Active = &active
    return b
}

func (b *PatientBuilder) SetGender(gender valuesets.AdministrativeGender) *PatientBuilder {
    b.data.Gender = &gender
    return b
}

func (b *PatientBuilder) SetBirthDate(date string) *PatientBuilder {
    b.data.BirthDate = &date
    return b
}

// Array helpers
func (b *PatientBuilder) AddIdentifier(identifier types.Identifier) *PatientBuilder {
    b.data.Identifier = append(b.data.Identifier, identifier)
    return b
}

func (b *PatientBuilder) AddName(name types.HumanName) *PatientBuilder {
    b.data.Name = append(b.data.Name, name)
    return b
}

func (b *PatientBuilder) AddTelecom(telecom types.ContactPoint) *PatientBuilder {
    b.data.Telecom = append(b.data.Telecom, telecom)
    return b
}

func (b *PatientBuilder) AddAddress(address types.Address) *PatientBuilder {
    b.data.Address = append(b.data.Address, address)
    return b
}

func (b *PatientBuilder) AddContact(contact types.PatientContact) *PatientBuilder {
    b.data.Contact = append(b.data.Contact, contact)
    return b
}

// Extension support
func (b *PatientBuilder) AddExtension(ext types.Extension) *PatientBuilder {
    b.data.Extension = append(b.data.Extension, ext)
    return b
}

func (b *PatientBuilder) AddModifierExtension(ext types.Extension) *PatientBuilder {
    b.data.ModifierExtension = append(b.data.ModifierExtension, ext)
    return b
}

// Choice type: deceased[x]
func (b *PatientBuilder) SetDeceasedBoolean(deceased bool) *PatientBuilder {
    b.data.DeceasedBoolean = &deceased
    b.data.DeceasedDateTime = nil
    return b
}

func (b *PatientBuilder) SetDeceasedDateTime(dateTime string) *PatientBuilder {
    b.data.DeceasedDateTime = &dateTime
    b.data.DeceasedBoolean = nil
    return b
}

// Choice type: multipleBirth[x]
func (b *PatientBuilder) SetMultipleBirthBoolean(value bool) *PatientBuilder {
    b.data.MultipleBirthBoolean = &value
    b.data.MultipleBirthInteger = nil
    return b
}

func (b *PatientBuilder) SetMultipleBirthInteger(value int) *PatientBuilder {
    b.data.MultipleBirthInteger = &value
    b.data.MultipleBirthBoolean = nil
    return b
}

// Build returns the constructed model
func (b *PatientBuilder) Build() *resources.Patient {
    return resources.NewPatient(&b.data)
}

// BuildRaw returns the raw data struct
func (b *PatientBuilder) BuildRaw() types.Patient {
    return b.data
}
```

### 4.3 Datatype Builder Example

```go
// pkg/r4/builders/datatypes/codeable_concept_builder.go

package datatypes

import "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"

type CodeableConceptBuilder struct {
    data types.CodeableConcept
}

func NewCodeableConceptBuilder() *CodeableConceptBuilder {
    return &CodeableConceptBuilder{}
}

func (b *CodeableConceptBuilder) AddCoding(coding types.Coding) *CodeableConceptBuilder {
    b.data.Coding = append(b.data.Coding, coding)
    return b
}

func (b *CodeableConceptBuilder) SetText(text string) *CodeableConceptBuilder {
    b.data.Text = &text
    return b
}

func (b *CodeableConceptBuilder) Build() types.CodeableConcept {
    return b.data
}

// CodingBuilder for nested building
type CodingBuilder struct {
    data types.Coding
}

func NewCodingBuilder() *CodingBuilder {
    return &CodingBuilder{}
}

func (b *CodingBuilder) SetSystem(system string) *CodingBuilder {
    b.data.System = &system
    return b
}

func (b *CodingBuilder) SetCode(code string) *CodingBuilder {
    b.data.Code = &code
    return b
}

func (b *CodingBuilder) SetDisplay(display string) *CodingBuilder {
    b.data.Display = &display
    return b
}

func (b *CodingBuilder) SetVersion(version string) *CodingBuilder {
    b.data.Version = &version
    return b
}

func (b *CodingBuilder) SetUserSelected(selected bool) *CodingBuilder {
    b.data.UserSelected = &selected
    return b
}

func (b *CodingBuilder) Build() types.Coding {
    return b.data
}
```

### 4.4 Usage Example

```go
// Example usage of builders

package main

import (
    "fmt"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/builders/resources"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/builders/datatypes"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types/valuesets"
)

func main() {
    // Build a Patient using fluent API
    patient := resources.NewPatientBuilder().
        SetID("patient-123").
        SetActive(true).
        AddIdentifier(types.Identifier{
            System: strPtr("http://example.com/mrn"),
            Value:  strPtr("12345"),
        }).
        AddName(types.HumanName{
            Use:    strPtr("official"),
            Family: strPtr("García"),
            Given:  []string{"María", "José"},
        }).
        SetGender(valuesets.AdministrativeGenderFemale).
        SetBirthDate("1985-03-15").
        AddAddress(types.Address{
            Use:     strPtr("home"),
            City:    strPtr("Santiago"),
            Country: strPtr("Chile"),
        }).
        Build()

    json, _ := patient.ToJSON()
    fmt.Println(string(json))

    // Build an Observation with nested builders
    observation := resources.NewObservationBuilder().
        SetID("obs-456").
        SetStatus(valuesets.ObservationStatusFinal).
        SetCode(datatypes.NewCodeableConceptBuilder().
            AddCoding(datatypes.NewCodingBuilder().
                SetSystem("http://loinc.org").
                SetCode("8310-5").
                SetDisplay("Body temperature").
                Build()).
            Build()).
        SetSubject(types.Reference{
            Reference: strPtr("Patient/patient-123"),
        }).
        SetValueQuantity(&types.Quantity{
            Value:  floatPtr(37.5),
            Unit:   strPtr("°C"),
            System: strPtr("http://unitsofmeasure.org"),
            Code:   strPtr("Cel"),
        }).
        Build()

    json, _ = observation.ToJSON()
    fmt.Println(string(json))
}

func strPtr(s string) *string { return &s }
func floatPtr(f float64) *float64 { return &f }
```

---

## Part 4B: Functional Options Pattern (Alternative API)

In addition to the fluent builder pattern, implement a **Functional Options Pattern** for simpler, more idiomatic Go construction of FHIR resources. This pattern is preferred by many Go developers for its simplicity and composability.

### 4B.1 Functional Options Design

```go
// pkg/r4/resources/patient_options.go

package resources

import (
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types/valuesets"
)

// PatientOption is a functional option for configuring a Patient
type PatientOption func(*Patient)

// NewPatient creates a new Patient with the given options
func NewPatient(opts ...PatientOption) *Patient {
    p := &Patient{
        data: types.Patient{
            ResourceType: "Patient",
        },
    }
    for _, opt := range opts {
        opt(p)
    }
    return p
}

// WithID sets the patient's ID
func WithID(id string) PatientOption {
    return func(p *Patient) {
        p.data.ID = &id
    }
}

// WithMeta sets the patient's metadata
func WithMeta(meta types.Meta) PatientOption {
    return func(p *Patient) {
        p.data.Meta = &meta
    }
}

// WithActive sets whether the patient record is active
func WithActive(active bool) PatientOption {
    return func(p *Patient) {
        p.data.Active = &active
    }
}

// WithName adds a name to the patient
// First argument is family name, rest are given names
func WithName(family string, given ...string) PatientOption {
    return func(p *Patient) {
        name := types.HumanName{
            Family: &family,
            Given:  given,
        }
        p.data.Name = append(p.data.Name, name)
    }
}

// WithOfficialName adds an official name to the patient
func WithOfficialName(family string, given ...string) PatientOption {
    return func(p *Patient) {
        use := "official"
        name := types.HumanName{
            Use:    &use,
            Family: &family,
            Given:  given,
        }
        p.data.Name = append(p.data.Name, name)
    }
}

// WithIdentifier adds an identifier to the patient
func WithIdentifier(system, value string) PatientOption {
    return func(p *Patient) {
        p.data.Identifier = append(p.data.Identifier, types.Identifier{
            System: &system,
            Value:  &value,
        })
    }
}

// WithMRN adds a Medical Record Number identifier
func WithMRN(system, value string) PatientOption {
    return func(p *Patient) {
        use := "official"
        typeCode := types.CodeableConcept{
            Coding: []types.Coding{{
                System:  strPtr("http://terminology.hl7.org/CodeSystem/v2-0203"),
                Code:    strPtr("MR"),
                Display: strPtr("Medical Record Number"),
            }},
        }
        p.data.Identifier = append(p.data.Identifier, types.Identifier{
            Use:    &use,
            Type:   &typeCode,
            System: &system,
            Value:  &value,
        })
    }
}

// WithBirthDate sets the patient's birth date (YYYY-MM-DD format)
func WithBirthDate(date string) PatientOption {
    return func(p *Patient) {
        p.data.BirthDate = &date
    }
}

// WithGender sets the patient's gender
func WithGender(gender valuesets.AdministrativeGender) PatientOption {
    return func(p *Patient) {
        p.data.Gender = &gender
    }
}

// WithPhone adds a phone number to the patient
func WithPhone(number string) PatientOption {
    return func(p *Patient) {
        system := "phone"
        p.data.Telecom = append(p.data.Telecom, types.ContactPoint{
            System: &system,
            Value:  &number,
        })
    }
}

// WithMobilePhone adds a mobile phone number
func WithMobilePhone(number string) PatientOption {
    return func(p *Patient) {
        system := "phone"
        use := "mobile"
        p.data.Telecom = append(p.data.Telecom, types.ContactPoint{
            System: &system,
            Use:    &use,
            Value:  &number,
        })
    }
}

// WithEmail adds an email address to the patient
func WithEmail(email string) PatientOption {
    return func(p *Patient) {
        system := "email"
        p.data.Telecom = append(p.data.Telecom, types.ContactPoint{
            System: &system,
            Value:  &email,
        })
    }
}

// WithAddress adds an address to the patient
func WithAddress(line []string, city, state, postalCode, country string) PatientOption {
    return func(p *Patient) {
        p.data.Address = append(p.data.Address, types.Address{
            Line:       line,
            City:       &city,
            State:      &state,
            PostalCode: &postalCode,
            Country:    &country,
        })
    }
}

// WithHomeAddress adds a home address
func WithHomeAddress(line []string, city, state, postalCode, country string) PatientOption {
    return func(p *Patient) {
        use := "home"
        p.data.Address = append(p.data.Address, types.Address{
            Use:        &use,
            Line:       line,
            City:       &city,
            State:      &state,
            PostalCode: &postalCode,
            Country:    &country,
        })
    }
}

// WithDeceased marks the patient as deceased with a date
func WithDeceased(dateTime string) PatientOption {
    return func(p *Patient) {
        p.data.DeceasedDateTime = &dateTime
        p.data.DeceasedBoolean = nil
    }
}

// WithDeceasedBoolean marks the patient as deceased (boolean)
func WithDeceasedBoolean(deceased bool) PatientOption {
    return func(p *Patient) {
        p.data.DeceasedBoolean = &deceased
        p.data.DeceasedDateTime = nil
    }
}

// WithMaritalStatus sets the patient's marital status
func WithMaritalStatus(code, display string) PatientOption {
    return func(p *Patient) {
        p.data.MaritalStatus = &types.CodeableConcept{
            Coding: []types.Coding{{
                System:  strPtr("http://terminology.hl7.org/CodeSystem/v3-MaritalStatus"),
                Code:    &code,
                Display: &display,
            }},
        }
    }
}

// WithExtension adds an extension to the patient
func WithExtension(url string, value interface{}) PatientOption {
    return func(p *Patient) {
        ext := types.Extension{URL: url}
        // Set appropriate value based on type
        switch v := value.(type) {
        case string:
            ext.ValueString = &v
        case bool:
            ext.ValueBoolean = &v
        case int:
            ext.ValueInteger = &v
        case *types.Address:
            ext.ValueAddress = v
        case *types.CodeableConcept:
            ext.ValueCodeableConcept = v
        // Add more types as needed
        }
        p.data.Extension = append(p.data.Extension, ext)
    }
}

// WithGeneralPractitioner adds a general practitioner reference
func WithGeneralPractitioner(reference string) PatientOption {
    return func(p *Patient) {
        p.data.GeneralPractitioner = append(p.data.GeneralPractitioner, types.Reference{
            Reference: &reference,
        })
    }
}

// WithManagingOrganization sets the managing organization
func WithManagingOrganization(reference, display string) PatientOption {
    return func(p *Patient) {
        p.data.ManagingOrganization = &types.Reference{
            Reference: &reference,
            Display:   &display,
        }
    }
}

// Helper function
func strPtr(s string) *string { return &s }
```

### 4B.2 Observation Functional Options

```go
// pkg/r4/resources/observation_options.go

package resources

import (
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types/valuesets"
)

type ObservationOption func(*Observation)

// NewObservation creates a new Observation with the given options
func NewObservation(status valuesets.ObservationStatus, code types.CodeableConcept, opts ...ObservationOption) *Observation {
    o := &Observation{
        data: types.Observation{
            ResourceType: "Observation",
            Status:       &status,
            Code:         code,
        },
    }
    for _, opt := range opts {
        opt(o)
    }
    return o
}

// WithObsID sets the observation's ID
func WithObsID(id string) ObservationOption {
    return func(o *Observation) {
        o.data.ID = &id
    }
}

// WithSubject sets the subject reference
func WithSubject(reference string) ObservationOption {
    return func(o *Observation) {
        o.data.Subject = &types.Reference{
            Reference: &reference,
        }
    }
}

// WithSubjectDisplay sets the subject with display name
func WithSubjectDisplay(reference, display string) ObservationOption {
    return func(o *Observation) {
        o.data.Subject = &types.Reference{
            Reference: &reference,
            Display:   &display,
        }
    }
}

// WithEncounter sets the encounter reference
func WithEncounter(reference string) ObservationOption {
    return func(o *Observation) {
        o.data.Encounter = &types.Reference{
            Reference: &reference,
        }
    }
}

// WithEffectiveDateTime sets the effective date/time
func WithEffectiveDateTime(dateTime string) ObservationOption {
    return func(o *Observation) {
        o.data.EffectiveDateTime = &dateTime
        o.data.EffectivePeriod = nil
        o.data.EffectiveInstant = nil
    }
}

// WithValueQuantity sets a quantity value
func WithValueQuantity(value float64, unit, system, code string) ObservationOption {
    return func(o *Observation) {
        o.clearValue()
        o.data.ValueQuantity = &types.Quantity{
            Value:  &value,
            Unit:   &unit,
            System: &system,
            Code:   &code,
        }
    }
}

// WithValueString sets a string value
func WithValueString(value string) ObservationOption {
    return func(o *Observation) {
        o.clearValue()
        o.data.ValueString = &value
    }
}

// WithValueCodeableConcept sets a codeable concept value
func WithValueCodeableConcept(system, code, display string) ObservationOption {
    return func(o *Observation) {
        o.clearValue()
        o.data.ValueCodeableConcept = &types.CodeableConcept{
            Coding: []types.Coding{{
                System:  &system,
                Code:    &code,
                Display: &display,
            }},
        }
    }
}

// WithValueBoolean sets a boolean value
func WithValueBoolean(value bool) ObservationOption {
    return func(o *Observation) {
        o.clearValue()
        o.data.ValueBoolean = &value
    }
}

// WithCategory adds a category
func WithCategory(system, code, display string) ObservationOption {
    return func(o *Observation) {
        o.data.Category = append(o.data.Category, types.CodeableConcept{
            Coding: []types.Coding{{
                System:  &system,
                Code:    &code,
                Display: &display,
            }},
        })
    }
}

// WithVitalSignsCategory adds the vital-signs category
func WithVitalSignsCategory() ObservationOption {
    return WithCategory(
        "http://terminology.hl7.org/CodeSystem/observation-category",
        "vital-signs",
        "Vital Signs",
    )
}

// WithLaboratoryCategory adds the laboratory category
func WithLaboratoryCategory() ObservationOption {
    return WithCategory(
        "http://terminology.hl7.org/CodeSystem/observation-category",
        "laboratory",
        "Laboratory",
    )
}

// WithComponent adds a component to the observation
func WithComponent(code types.CodeableConcept, valueQuantity *types.Quantity) ObservationOption {
    return func(o *Observation) {
        o.data.Component = append(o.data.Component, types.ObservationComponent{
            Code:          code,
            ValueQuantity: valueQuantity,
        })
    }
}

// WithInterpretation adds an interpretation
func WithInterpretation(system, code, display string) ObservationOption {
    return func(o *Observation) {
        o.data.Interpretation = append(o.data.Interpretation, types.CodeableConcept{
            Coding: []types.Coding{{
                System:  &system,
                Code:    &code,
                Display: &display,
            }},
        })
    }
}

// WithNote adds a note/comment
func WithNote(text string) ObservationOption {
    return func(o *Observation) {
        o.data.Note = append(o.data.Note, types.Annotation{
            Text: text,
        })
    }
}

// WithPerformer adds a performer reference
func WithPerformer(reference string) ObservationOption {
    return func(o *Observation) {
        o.data.Performer = append(o.data.Performer, types.Reference{
            Reference: &reference,
        })
    }
}

// WithReferenceRange adds a reference range
func WithReferenceRange(low, high float64, unit string) ObservationOption {
    return func(o *Observation) {
        o.data.ReferenceRange = append(o.data.ReferenceRange, types.ObservationReferenceRange{
            Low: &types.Quantity{
                Value: &low,
                Unit:  &unit,
            },
            High: &types.Quantity{
                Value: &high,
                Unit:  &unit,
            },
        })
    }
}

// WithDataAbsentReason sets the data absent reason
func WithDataAbsentReason(code, display string) ObservationOption {
    return func(o *Observation) {
        o.clearValue()
        o.data.DataAbsentReason = &types.CodeableConcept{
            Coding: []types.Coding{{
                System:  strPtr("http://terminology.hl7.org/CodeSystem/data-absent-reason"),
                Code:    &code,
                Display: &display,
            }},
        }
    }
}
```

### 4B.3 Common LOINC Code Helpers

```go
// pkg/r4/resources/loinc_codes.go

package resources

import "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"

// Common LOINC codes as pre-built CodeableConcepts
var (
    // Vital Signs
    LOINCBodyWeight = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("29463-7"),
            Display: strPtr("Body weight"),
        }},
    }

    LOINCBodyHeight = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("8302-2"),
            Display: strPtr("Body height"),
        }},
    }

    LOINCBodyTemperature = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("8310-5"),
            Display: strPtr("Body temperature"),
        }},
    }

    LOINCBloodPressurePanel = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("85354-9"),
            Display: strPtr("Blood pressure panel"),
        }},
    }

    LOINCSystolicBP = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("8480-6"),
            Display: strPtr("Systolic blood pressure"),
        }},
    }

    LOINCDiastolicBP = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("8462-4"),
            Display: strPtr("Diastolic blood pressure"),
        }},
    }

    LOINCHeartRate = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("8867-4"),
            Display: strPtr("Heart rate"),
        }},
    }

    LOINCRespiratoryRate = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("9279-1"),
            Display: strPtr("Respiratory rate"),
        }},
    }

    LOINCOxygenSaturation = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("2708-6"),
            Display: strPtr("Oxygen saturation"),
        }},
    }

    LOINCBMI = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("39156-5"),
            Display: strPtr("Body mass index"),
        }},
    }

    // Laboratory
    LOINCGlucose = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("2339-0"),
            Display: strPtr("Glucose [Mass/volume] in Blood"),
        }},
    }

    LOINCHemoglobin = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("718-7"),
            Display: strPtr("Hemoglobin [Mass/volume] in Blood"),
        }},
    }

    LOINCHemoglobinA1c = types.CodeableConcept{
        Coding: []types.Coding{{
            System:  strPtr("http://loinc.org"),
            Code:    strPtr("4548-4"),
            Display: strPtr("Hemoglobin A1c/Hemoglobin.total in Blood"),
        }},
    }
)

// UCUM unit helpers
func Quantity(value float64, unit, code string) *types.Quantity {
    return &types.Quantity{
        Value:  &value,
        Unit:   &unit,
        System: strPtr("http://unitsofmeasure.org"),
        Code:   &code,
    }
}

func QuantityKg(value float64) *types.Quantity {
    return Quantity(value, "kg", "kg")
}

func QuantityCm(value float64) *types.Quantity {
    return Quantity(value, "cm", "cm")
}

func QuantityCelsius(value float64) *types.Quantity {
    return Quantity(value, "°C", "Cel")
}

func QuantityMmHg(value float64) *types.Quantity {
    return Quantity(value, "mmHg", "mm[Hg]")
}

func QuantityBPM(value float64) *types.Quantity {
    return Quantity(value, "beats/min", "/min")
}

func QuantityPercent(value float64) *types.Quantity {
    return Quantity(value, "%", "%")
}

func QuantityMgDL(value float64) *types.Quantity {
    return Quantity(value, "mg/dL", "mg/dL")
}

func QuantityGDL(value float64) *types.Quantity {
    return Quantity(value, "g/dL", "g/dL")
}
```

### 4B.4 Usage Examples

```go
// examples/functional_options/main.go

package main

import (
    "encoding/json"
    "fmt"

    "github.com/yourorg/fhir-toolkit-go/pkg/r4/resources"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types/valuesets"
)

func main() {
    // =============================================
    // Example 1: Create a Patient with functional options
    // =============================================

    patient := resources.NewPatient(
        resources.WithID("patient-001"),
        resources.WithActive(true),
        resources.WithMRN("http://hospital.example.org/mrn", "MRN-12345"),
        resources.WithOfficialName("García", "María", "Isabel"),
        resources.WithName("Mari"),  // Nickname
        resources.WithGender(valuesets.AdministrativeGenderFemale),
        resources.WithBirthDate("1985-03-15"),
        resources.WithMobilePhone("+56 9 1234 5678"),
        resources.WithEmail("maria.garcia@email.com"),
        resources.WithHomeAddress(
            []string{"Av. Libertador 1234", "Depto 567"},
            "Santiago",
            "Región Metropolitana",
            "8320000",
            "Chile",
        ),
        resources.WithMaritalStatus("M", "Married"),
        resources.WithManagingOrganization("Organization/hospital-1", "Hospital Central"),
    )

    patientJSON, _ := json.MarshalIndent(patient, "", "  ")
    fmt.Println("=== Patient ===")
    fmt.Println(string(patientJSON))

    // =============================================
    // Example 2: Create vital signs observations
    // =============================================

    // Body Temperature
    temperature := resources.NewObservation(
        valuesets.ObservationStatusFinal,
        resources.LOINCBodyTemperature,
        resources.WithObsID("temp-001"),
        resources.WithSubjectDisplay("Patient/patient-001", "María García"),
        resources.WithEffectiveDateTime("2024-01-15T10:30:00-03:00"),
        resources.WithVitalSignsCategory(),
        resources.WithValueQuantity(37.5, "°C", "http://unitsofmeasure.org", "Cel"),
    )

    tempJSON, _ := json.MarshalIndent(temperature, "", "  ")
    fmt.Println("\n=== Body Temperature ===")
    fmt.Println(string(tempJSON))

    // Blood Pressure (with components)
    bloodPressure := resources.NewObservation(
        valuesets.ObservationStatusFinal,
        resources.LOINCBloodPressurePanel,
        resources.WithObsID("bp-001"),
        resources.WithSubject("Patient/patient-001"),
        resources.WithEffectiveDateTime("2024-01-15T10:30:00-03:00"),
        resources.WithVitalSignsCategory(),
        resources.WithComponent(resources.LOINCSystolicBP, resources.QuantityMmHg(120)),
        resources.WithComponent(resources.LOINCDiastolicBP, resources.QuantityMmHg(80)),
        resources.WithPerformer("Practitioner/nurse-001"),
    )

    bpJSON, _ := json.MarshalIndent(bloodPressure, "", "  ")
    fmt.Println("\n=== Blood Pressure ===")
    fmt.Println(string(bpJSON))

    // Body Weight with reference range
    weight := resources.NewObservation(
        valuesets.ObservationStatusFinal,
        resources.LOINCBodyWeight,
        resources.WithObsID("weight-001"),
        resources.WithSubject("Patient/patient-001"),
        resources.WithEffectiveDateTime("2024-01-15T10:30:00-03:00"),
        resources.WithVitalSignsCategory(),
        resources.WithValueQuantity(70.5, "kg", "http://unitsofmeasure.org", "kg"),
        resources.WithReferenceRange(50, 90, "kg"),
        resources.WithNote("Measured with clothes on"),
    )

    weightJSON, _ := json.MarshalIndent(weight, "", "  ")
    fmt.Println("\n=== Body Weight ===")
    fmt.Println(string(weightJSON))

    // Lab Result - Glucose
    glucose := resources.NewObservation(
        valuesets.ObservationStatusFinal,
        resources.LOINCGlucose,
        resources.WithObsID("glucose-001"),
        resources.WithSubject("Patient/patient-001"),
        resources.WithEffectiveDateTime("2024-01-15T08:00:00-03:00"),
        resources.WithLaboratoryCategory(),
        resources.WithValueQuantity(95, "mg/dL", "http://unitsofmeasure.org", "mg/dL"),
        resources.WithReferenceRange(70, 100, "mg/dL"),
        resources.WithInterpretation(
            "http://terminology.hl7.org/CodeSystem/v3-ObservationInterpretation",
            "N",
            "Normal",
        ),
    )

    glucoseJSON, _ := json.MarshalIndent(glucose, "", "  ")
    fmt.Println("\n=== Glucose Lab Result ===")
    fmt.Println(string(glucoseJSON))

    // =============================================
    // Example 3: Observation with data absent reason
    // =============================================

    unknownObs := resources.NewObservation(
        valuesets.ObservationStatusFinal,
        resources.LOINCBodyWeight,
        resources.WithObsID("weight-unknown"),
        resources.WithSubject("Patient/patient-001"),
        resources.WithEffectiveDateTime("2024-01-15T10:30:00-03:00"),
        resources.WithVitalSignsCategory(),
        resources.WithDataAbsentReason("not-performed", "Not Performed"),
        resources.WithNote("Patient refused measurement"),
    )

    unknownJSON, _ := json.MarshalIndent(unknownObs, "", "  ")
    fmt.Println("\n=== Observation with Data Absent ===")
    fmt.Println(string(unknownJSON))

    // =============================================
    // Example 4: Combining functional options with builder
    // =============================================

    // You can also use functional options to create a base and then modify with builder
    basePatient := resources.NewPatient(
        resources.WithID("patient-template"),
        resources.WithActive(true),
    )

    // Then use builder for modifications
    modifiedPatient := basePatient.ToBuilder().
        AddIdentifier(types.Identifier{
            System: strPtr("http://custom-system"),
            Value:  strPtr("custom-value"),
        }).
        Build()

    modifiedJSON, _ := json.MarshalIndent(modifiedPatient, "", "  ")
    fmt.Println("\n=== Modified Patient ===")
    fmt.Println(string(modifiedJSON))
}
```

### 4B.5 Generic Functional Options for All Resources

```go
// pkg/r4/resources/common_options.go

package resources

import "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"

// Generic option type for any resource
type ResourceOption[T any] func(*T)

// Common option creators that work with any resource type

// WithResourceID creates an ID option for any resource
func WithResourceID[T any](id string, setter func(*T, string)) ResourceOption[T] {
    return func(r *T) {
        setter(r, id)
    }
}

// Compose combines multiple options into one
func Compose[T any](opts ...ResourceOption[T]) ResourceOption[T] {
    return func(r *T) {
        for _, opt := range opts {
            opt(r)
        }
    }
}

// When applies an option conditionally
func When[T any](condition bool, opt ResourceOption[T]) ResourceOption[T] {
    return func(r *T) {
        if condition {
            opt(r)
        }
    }
}

// Map applies a function to each element and creates options
func Map[T any, E any](elements []E, optFn func(E) ResourceOption[T]) ResourceOption[T] {
    return func(r *T) {
        for _, elem := range elements {
            optFn(elem)(r)
        }
    }
}

// Example: Conditional options
func ExampleConditionalOptions() {
    isVIP := true
    hasMobile := "+56 9 1234 5678"

    patient := NewPatient(
        WithID("patient-001"),
        WithActive(true),
        When(isVIP, WithExtension("http://example.org/vip", true)),
        When(hasMobile != "", WithMobilePhone(hasMobile)),
    )

    _ = patient
}

// Example: Mapping over collections
func ExampleMappingOptions() {
    phones := []string{"+56 9 1111 1111", "+56 9 2222 2222"}

    patient := NewPatient(
        WithID("patient-001"),
        Map(phones, func(phone string) PatientOption {
            return WithPhone(phone)
        }),
    )

    _ = patient
}
```

### 4B.6 Comparison: Builders vs Functional Options

| Aspect | Builder Pattern | Functional Options |
|--------|----------------|-------------------|
| **Syntax** | `NewBuilder().SetX().SetY().Build()` | `New(WithX(), WithY())` |
| **Chaining** | Method chaining | Function composition |
| **Intermediate state** | Mutable builder object | None (applied at construction) |
| **IDE support** | Good autocomplete | Good autocomplete |
| **Extensibility** | Need to modify builder | Just add new option functions |
| **Testing** | Test builder methods | Test option functions |
| **Go idiom** | Common in Go | More idiomatic for construction |
| **Use case** | Complex construction with many steps | Simple to moderate construction |
| **Best for** | When you need intermediate modifications | When you want immutable construction |

### 4B.7 Recommendation

Provide **both patterns** in the toolkit:

1. **Functional Options** (`NewPatient(opts...)`) - Preferred for simple, common cases
2. **Builders** (`NewPatientBuilder().Build()`) - For complex construction with conditional logic

Allow conversion between them:

```go
// From functional options to builder
patient := NewPatient(WithID("123"), WithActive(true))
builder := patient.ToBuilder()
builder.AddName(types.HumanName{...})
finalPatient := builder.Build()

// From builder to applying options
builder := NewPatientBuilder()
builder.Apply(WithID("123"), WithActive(true)) // Apply functional options to builder
```

---

## Part 5: FHIRPath Implementation

### 5.1 FHIRPath Overview

FHIRPath is an expression language for navigating and extracting data from FHIR resources. It's essential for:
- Validating FHIR constraints (invariants)
- Querying resource data
- Search parameter definitions
- Subscription criteria

### 5.2 FHIRPath Grammar (Subset)

```
expression      = term (operator term)*
term            = invocation | literal | '(' expression ')'
invocation      = (identifier | function) ('.' invocation)*
function        = identifier '(' (expression (',' expression)*)? ')'
identifier      = [A-Za-z_][A-Za-z0-9_]*
literal         = STRING | NUMBER | DATE | DATETIME | TIME | BOOLEAN
operator        = '=' | '!=' | '<' | '<=' | '>' | '>='
                | 'and' | 'or' | 'xor' | 'implies'
                | '+' | '-' | '*' | '/' | 'div' | 'mod'
                | '&' | '|' | 'in' | 'contains'
```

### 5.3 FHIRPath Lexer

```go
// pkg/fhirpath/parser/lexer.go

package parser

type TokenType int

const (
    TokenEOF TokenType = iota
    TokenIdentifier
    TokenString
    TokenNumber
    TokenBoolean
    TokenDateTime
    TokenDate
    TokenTime

    // Operators
    TokenPlus
    TokenMinus
    TokenStar
    TokenSlash
    TokenEquals
    TokenNotEquals
    TokenLessThan
    TokenLessOrEqual
    TokenGreaterThan
    TokenGreaterOrEqual
    TokenAnd
    TokenOr
    TokenXor
    TokenImplies
    TokenIn
    TokenContains
    TokenAs
    TokenIs
    TokenDiv
    TokenMod
    TokenUnion  // |
    TokenConcat // &

    // Delimiters
    TokenDot
    TokenComma
    TokenLParen
    TokenRParen
    TokenLBracket
    TokenRBracket
)

type Token struct {
    Type    TokenType
    Value   string
    Line    int
    Column  int
}

type Lexer struct {
    input   string
    pos     int
    line    int
    column  int
}

func NewLexer(input string) *Lexer {
    return &Lexer{
        input:  input,
        pos:    0,
        line:   1,
        column: 1,
    }
}

func (l *Lexer) NextToken() Token {
    l.skipWhitespace()

    if l.pos >= len(l.input) {
        return Token{Type: TokenEOF}
    }

    ch := l.input[l.pos]

    switch {
    case ch == '.':
        return l.singleChar(TokenDot)
    case ch == ',':
        return l.singleChar(TokenComma)
    case ch == '(':
        return l.singleChar(TokenLParen)
    case ch == ')':
        return l.singleChar(TokenRParen)
    case ch == '[':
        return l.singleChar(TokenLBracket)
    case ch == ']':
        return l.singleChar(TokenRBracket)
    case ch == '+':
        return l.singleChar(TokenPlus)
    case ch == '-':
        return l.singleChar(TokenMinus)
    case ch == '*':
        return l.singleChar(TokenStar)
    case ch == '/':
        return l.singleChar(TokenSlash)
    case ch == '|':
        return l.singleChar(TokenUnion)
    case ch == '&':
        return l.singleChar(TokenConcat)
    case ch == '=':
        return l.singleChar(TokenEquals)
    case ch == '!' && l.peek() == '=':
        return l.doubleChar(TokenNotEquals)
    case ch == '<':
        if l.peek() == '=' {
            return l.doubleChar(TokenLessOrEqual)
        }
        return l.singleChar(TokenLessThan)
    case ch == '>':
        if l.peek() == '=' {
            return l.doubleChar(TokenGreaterOrEqual)
        }
        return l.singleChar(TokenGreaterThan)
    case ch == '\'':
        return l.readString()
    case ch == '@':
        return l.readDateTime()
    case isDigit(ch):
        return l.readNumber()
    case isLetter(ch):
        return l.readIdentifier()
    }

    panic(fmt.Sprintf("unexpected character: %c at line %d, column %d",
        ch, l.line, l.column))
}

func (l *Lexer) readIdentifier() Token {
    start := l.pos
    for l.pos < len(l.input) && isAlphanumeric(l.input[l.pos]) {
        l.pos++
        l.column++
    }

    value := l.input[start:l.pos]

    // Check for keywords
    switch value {
    case "and":
        return Token{Type: TokenAnd, Value: value}
    case "or":
        return Token{Type: TokenOr, Value: value}
    case "xor":
        return Token{Type: TokenXor, Value: value}
    case "implies":
        return Token{Type: TokenImplies, Value: value}
    case "in":
        return Token{Type: TokenIn, Value: value}
    case "contains":
        return Token{Type: TokenContains, Value: value}
    case "as":
        return Token{Type: TokenAs, Value: value}
    case "is":
        return Token{Type: TokenIs, Value: value}
    case "div":
        return Token{Type: TokenDiv, Value: value}
    case "mod":
        return Token{Type: TokenMod, Value: value}
    case "true", "false":
        return Token{Type: TokenBoolean, Value: value}
    }

    return Token{Type: TokenIdentifier, Value: value}
}

// Additional helper methods...
```

### 5.4 FHIRPath AST

```go
// pkg/fhirpath/parser/ast.go

package parser

type Node interface {
    node()
}

// Expression nodes
type Expression interface {
    Node
    expr()
}

type BinaryExpr struct {
    Left     Expression
    Operator string
    Right    Expression
}

type UnaryExpr struct {
    Operator string
    Operand  Expression
}

type InvocationExpr struct {
    Target     Expression // nil for root
    Invocation Invocation
}

type Invocation interface {
    Node
    invocation()
}

type MemberInvocation struct {
    Identifier string
}

type FunctionInvocation struct {
    Name      string
    Arguments []Expression
}

type IndexerInvocation struct {
    Index Expression
}

// Literal nodes
type LiteralExpr struct {
    Value interface{}
    Type  string // "string", "number", "boolean", "datetime", etc.
}

type ParenExpr struct {
    Inner Expression
}

// Implement interfaces
func (BinaryExpr) node()          {}
func (BinaryExpr) expr()          {}
func (UnaryExpr) node()           {}
func (UnaryExpr) expr()           {}
func (InvocationExpr) node()      {}
func (InvocationExpr) expr()      {}
func (LiteralExpr) node()         {}
func (LiteralExpr) expr()         {}
func (ParenExpr) node()           {}
func (ParenExpr) expr()           {}
func (MemberInvocation) node()    {}
func (MemberInvocation) invocation() {}
func (FunctionInvocation) node()  {}
func (FunctionInvocation) invocation() {}
func (IndexerInvocation) node()   {}
func (IndexerInvocation) invocation() {}
```

### 5.5 FHIRPath Parser

```go
// pkg/fhirpath/parser/parser.go

package parser

type Parser struct {
    lexer   *Lexer
    current Token
    peek    Token
}

func NewParser(input string) *Parser {
    p := &Parser{
        lexer: NewLexer(input),
    }
    // Prime the parser
    p.advance()
    p.advance()
    return p
}

func (p *Parser) Parse() (Expression, error) {
    return p.parseExpression(0)
}

func (p *Parser) parseExpression(precedence int) (Expression, error) {
    left, err := p.parseUnary()
    if err != nil {
        return nil, err
    }

    for p.current.Type != TokenEOF {
        op := p.current
        opPrecedence := p.getPrecedence(op.Type)

        if opPrecedence < precedence {
            break
        }

        p.advance()

        right, err := p.parseExpression(opPrecedence + 1)
        if err != nil {
            return nil, err
        }

        left = &BinaryExpr{
            Left:     left,
            Operator: op.Value,
            Right:    right,
        }
    }

    return left, nil
}

func (p *Parser) parseUnary() (Expression, error) {
    if p.current.Type == TokenMinus || p.current.Type == TokenPlus {
        op := p.current
        p.advance()
        operand, err := p.parseUnary()
        if err != nil {
            return nil, err
        }
        return &UnaryExpr{
            Operator: op.Value,
            Operand:  operand,
        }, nil
    }
    return p.parseInvocation()
}

func (p *Parser) parseInvocation() (Expression, error) {
    term, err := p.parseTerm()
    if err != nil {
        return nil, err
    }

    for p.current.Type == TokenDot || p.current.Type == TokenLBracket {
        if p.current.Type == TokenDot {
            p.advance()

            inv, err := p.parseInvocationTail()
            if err != nil {
                return nil, err
            }

            term = &InvocationExpr{
                Target:     term,
                Invocation: inv,
            }
        } else {
            // Indexer [expr]
            p.advance() // consume [
            index, err := p.parseExpression(0)
            if err != nil {
                return nil, err
            }
            if err := p.expect(TokenRBracket); err != nil {
                return nil, err
            }

            term = &InvocationExpr{
                Target:     term,
                Invocation: &IndexerInvocation{Index: index},
            }
        }
    }

    return term, nil
}

func (p *Parser) parseTerm() (Expression, error) {
    switch p.current.Type {
    case TokenIdentifier:
        return p.parseIdentifierOrFunction()
    case TokenString:
        return p.parseLiteral("string")
    case TokenNumber:
        return p.parseLiteral("number")
    case TokenBoolean:
        return p.parseLiteral("boolean")
    case TokenDateTime:
        return p.parseLiteral("datetime")
    case TokenLParen:
        return p.parseParenExpr()
    default:
        return nil, fmt.Errorf("unexpected token: %v", p.current)
    }
}

// Additional parsing methods...
```

### 5.6 FHIRPath Evaluator

```go
// pkg/fhirpath/evaluator/evaluator.go

package evaluator

import (
    "github.com/yourorg/fhir-toolkit-go/pkg/fhirpath/parser"
)

type Context struct {
    Resource    interface{}          // The root resource
    RootContext interface{}          // $this context
    Environment map[string]interface{} // %variables
    FHIRVersion string               // "R4", "R4B", "R5"
}

type Evaluator struct {
    functions map[string]FHIRPathFunction
}

type FHIRPathFunction func(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error)

func NewEvaluator() *Evaluator {
    e := &Evaluator{
        functions: make(map[string]FHIRPathFunction),
    }
    e.registerBuiltinFunctions()
    return e
}

func (e *Evaluator) Evaluate(expr parser.Expression, ctx *Context) ([]interface{}, error) {
    return e.eval(expr, ctx, []interface{}{ctx.Resource})
}

func (e *Evaluator) eval(expr parser.Expression, ctx *Context, input []interface{}) ([]interface{}, error) {
    switch node := expr.(type) {
    case *parser.LiteralExpr:
        return []interface{}{node.Value}, nil

    case *parser.InvocationExpr:
        return e.evalInvocation(node, ctx, input)

    case *parser.BinaryExpr:
        return e.evalBinary(node, ctx, input)

    case *parser.UnaryExpr:
        return e.evalUnary(node, ctx, input)

    case *parser.ParenExpr:
        return e.eval(node.Inner, ctx, input)

    default:
        return nil, fmt.Errorf("unknown expression type: %T", expr)
    }
}

func (e *Evaluator) evalInvocation(node *parser.InvocationExpr, ctx *Context, input []interface{}) ([]interface{}, error) {
    // Evaluate target first
    var target []interface{}
    if node.Target != nil {
        var err error
        target, err = e.eval(node.Target, ctx, input)
        if err != nil {
            return nil, err
        }
    } else {
        target = input
    }

    switch inv := node.Invocation.(type) {
    case *parser.MemberInvocation:
        return e.evalMember(inv.Identifier, target)

    case *parser.FunctionInvocation:
        return e.evalFunction(inv, ctx, target)

    case *parser.IndexerInvocation:
        index, err := e.eval(inv.Index, ctx, input)
        if err != nil {
            return nil, err
        }
        return e.evalIndexer(target, index)
    }

    return nil, fmt.Errorf("unknown invocation type")
}

func (e *Evaluator) evalMember(name string, input []interface{}) ([]interface{}, error) {
    var result []interface{}

    for _, item := range input {
        // Use reflection to access fields
        val := reflect.ValueOf(item)
        if val.Kind() == reflect.Ptr {
            val = val.Elem()
        }

        if val.Kind() == reflect.Map {
            // Handle map[string]interface{} (JSON)
            if mapVal, ok := item.(map[string]interface{}); ok {
                if v, exists := mapVal[name]; exists && v != nil {
                    result = append(result, v)
                }
            }
        } else if val.Kind() == reflect.Struct {
            // Handle struct fields
            field := val.FieldByName(capitalize(name))
            if field.IsValid() && !field.IsNil() {
                result = append(result, field.Interface())
            }
        }
    }

    // Flatten arrays
    return flatten(result), nil
}

func (e *Evaluator) evalBinary(node *parser.BinaryExpr, ctx *Context, input []interface{}) ([]interface{}, error) {
    left, err := e.eval(node.Left, ctx, input)
    if err != nil {
        return nil, err
    }

    right, err := e.eval(node.Right, ctx, input)
    if err != nil {
        return nil, err
    }

    switch node.Operator {
    case "=":
        return e.equals(left, right)
    case "!=":
        result, _ := e.equals(left, right)
        return e.not(result)
    case "and":
        return e.and(left, right)
    case "or":
        return e.or(left, right)
    case "implies":
        return e.implies(left, right)
    case "+":
        return e.add(left, right)
    case "-":
        return e.subtract(left, right)
    case "*":
        return e.multiply(left, right)
    case "/":
        return e.divide(left, right)
    case "|":
        return e.union(left, right)
    case "in":
        return e.in(left, right)
    case "contains":
        return e.contains(left, right)
    // ... more operators
    }

    return nil, fmt.Errorf("unknown operator: %s", node.Operator)
}

func (e *Evaluator) registerBuiltinFunctions() {
    e.functions["exists"] = funcExists
    e.functions["empty"] = funcEmpty
    e.functions["count"] = funcCount
    e.functions["first"] = funcFirst
    e.functions["last"] = funcLast
    e.functions["tail"] = funcTail
    e.functions["take"] = funcTake
    e.functions["skip"] = funcSkip
    e.functions["where"] = funcWhere
    e.functions["select"] = funcSelect
    e.functions["all"] = funcAll
    e.functions["any"] = funcAny  // alias for exists with predicate
    e.functions["not"] = funcNot
    e.functions["iif"] = funcIif
    e.functions["trace"] = funcTrace

    // Type functions
    e.functions["ofType"] = funcOfType
    e.functions["as"] = funcAs
    e.functions["is"] = funcIs

    // String functions
    e.functions["startsWith"] = funcStartsWith
    e.functions["endsWith"] = funcEndsWith
    e.functions["contains"] = funcContainsString
    e.functions["matches"] = funcMatches
    e.functions["indexOf"] = funcIndexOf
    e.functions["substring"] = funcSubstring
    e.functions["replace"] = funcReplace
    e.functions["length"] = funcLength
    e.functions["upper"] = funcUpper
    e.functions["lower"] = funcLower
    e.functions["toChars"] = funcToChars

    // Math functions
    e.functions["abs"] = funcAbs
    e.functions["ceiling"] = funcCeiling
    e.functions["floor"] = funcFloor
    e.functions["round"] = funcRound
    e.functions["sqrt"] = funcSqrt
    e.functions["ln"] = funcLn
    e.functions["log"] = funcLog
    e.functions["exp"] = funcExp
    e.functions["power"] = funcPower
    e.functions["truncate"] = funcTruncate

    // Date/Time functions
    e.functions["now"] = funcNow
    e.functions["today"] = funcToday
    e.functions["timeOfDay"] = funcTimeOfDay

    // Aggregate functions
    e.functions["aggregate"] = funcAggregate
    e.functions["sum"] = funcSum
    e.functions["min"] = funcMin
    e.functions["max"] = funcMax
    e.functions["avg"] = funcAvg

    // FHIR-specific functions
    e.functions["resolve"] = funcResolve
    e.functions["extension"] = funcExtension
    e.functions["hasValue"] = funcHasValue
    e.functions["getValue"] = funcGetValue
    e.functions["memberOf"] = funcMemberOf
}
```

### 5.7 FHIRPath Built-in Functions

```go
// pkg/fhirpath/evaluator/functions.go

package evaluator

// exists() - Returns true if the collection has any elements
func funcExists(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error) {
    if len(args) == 0 {
        return []interface{}{len(input) > 0}, nil
    }

    // exists(criteria) - evaluates criteria for each element
    predicate := args[0].(parser.Expression)
    for _, item := range input {
        result, err := ctx.Evaluator.eval(predicate, ctx, []interface{}{item})
        if err != nil {
            return nil, err
        }
        if toBool(result) {
            return []interface{}{true}, nil
        }
    }
    return []interface{}{false}, nil
}

// empty() - Returns true if the collection is empty
func funcEmpty(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error) {
    return []interface{}{len(input) == 0}, nil
}

// count() - Returns the number of elements in the collection
func funcCount(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error) {
    return []interface{}{len(input)}, nil
}

// first() - Returns the first element
func funcFirst(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error) {
    if len(input) == 0 {
        return []interface{}{}, nil
    }
    return []interface{}{input[0]}, nil
}

// where(criteria) - Filters the collection
func funcWhere(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error) {
    if len(args) == 0 {
        return nil, fmt.Errorf("where() requires a criteria argument")
    }

    predicate := args[0].(parser.Expression)
    var result []interface{}

    for _, item := range input {
        evalResult, err := ctx.Evaluator.eval(predicate, ctx, []interface{}{item})
        if err != nil {
            return nil, err
        }
        if toBool(evalResult) {
            result = append(result, item)
        }
    }

    return result, nil
}

// select(projection) - Projects each element
func funcSelect(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error) {
    if len(args) == 0 {
        return nil, fmt.Errorf("select() requires a projection argument")
    }

    projection := args[0].(parser.Expression)
    var result []interface{}

    for _, item := range input {
        projected, err := ctx.Evaluator.eval(projection, ctx, []interface{}{item})
        if err != nil {
            return nil, err
        }
        result = append(result, projected...)
    }

    return result, nil
}

// all(criteria) - Returns true if all elements match
func funcAll(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error) {
    if len(args) == 0 {
        return nil, fmt.Errorf("all() requires a criteria argument")
    }

    if len(input) == 0 {
        return []interface{}{true}, nil // Empty collection returns true
    }

    predicate := args[0].(parser.Expression)

    for _, item := range input {
        result, err := ctx.Evaluator.eval(predicate, ctx, []interface{}{item})
        if err != nil {
            return nil, err
        }
        if !toBool(result) {
            return []interface{}{false}, nil
        }
    }

    return []interface{}{true}, nil
}

// ofType(type) - Filters by FHIR type
func funcOfType(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error) {
    if len(args) == 0 {
        return nil, fmt.Errorf("ofType() requires a type argument")
    }

    typeName := args[0].(string)
    var result []interface{}

    for _, item := range input {
        if m, ok := item.(map[string]interface{}); ok {
            if rt, exists := m["resourceType"]; exists && rt == typeName {
                result = append(result, item)
            }
        }
    }

    return result, nil
}

// extension(url) - Get extensions by URL
func funcExtension(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error) {
    if len(args) == 0 {
        return nil, fmt.Errorf("extension() requires a URL argument")
    }

    url := toString(args[0])
    var result []interface{}

    for _, item := range input {
        if m, ok := item.(map[string]interface{}); ok {
            if exts, exists := m["extension"]; exists {
                for _, ext := range exts.([]interface{}) {
                    if extMap, ok := ext.(map[string]interface{}); ok {
                        if extMap["url"] == url {
                            result = append(result, ext)
                        }
                    }
                }
            }
        }
    }

    return result, nil
}

// iif(criterion, true-result, otherwise-result) - Conditional
func funcIif(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error) {
    if len(args) < 2 {
        return nil, fmt.Errorf("iif() requires at least 2 arguments")
    }

    criterion := args[0].(parser.Expression)
    trueResult := args[1].(parser.Expression)

    condResult, err := ctx.Evaluator.eval(criterion, ctx, input)
    if err != nil {
        return nil, err
    }

    if toBool(condResult) {
        return ctx.Evaluator.eval(trueResult, ctx, input)
    }

    if len(args) >= 3 {
        falseResult := args[2].(parser.Expression)
        return ctx.Evaluator.eval(falseResult, ctx, input)
    }

    return []interface{}{}, nil
}

// String functions
func funcStartsWith(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error) {
    if len(input) == 0 || len(args) == 0 {
        return []interface{}{}, nil
    }

    str := toString(input[0])
    prefix := toString(args[0])

    return []interface{}{strings.HasPrefix(str, prefix)}, nil
}

func funcMatches(ctx *Context, input []interface{}, args []interface{}) ([]interface{}, error) {
    if len(input) == 0 || len(args) == 0 {
        return []interface{}{}, nil
    }

    str := toString(input[0])
    pattern := toString(args[0])

    matched, err := regexp.MatchString(pattern, str)
    if err != nil {
        return nil, err
    }

    return []interface{}{matched}, nil
}

// Additional function implementations...
```

### 5.8 FHIRPath Compiler with Caching

```go
// pkg/fhirpath/compiler/compiler.go

package compiler

import (
    "sync"

    "github.com/yourorg/fhir-toolkit-go/pkg/fhirpath/parser"
    "github.com/yourorg/fhir-toolkit-go/pkg/fhirpath/evaluator"
)

type CompiledExpression struct {
    ast       parser.Expression
    evaluator *evaluator.Evaluator
}

type Compiler struct {
    cache    *LRUCache
    mu       sync.RWMutex
    maxCache int
}

func NewCompiler(maxCache int) *Compiler {
    if maxCache <= 0 {
        maxCache = 500 // Default cache size
    }
    return &Compiler{
        cache:    NewLRUCache(maxCache),
        maxCache: maxCache,
    }
}

func (c *Compiler) Compile(expression string) (*CompiledExpression, error) {
    c.mu.RLock()
    if cached, ok := c.cache.Get(expression); ok {
        c.mu.RUnlock()
        return cached.(*CompiledExpression), nil
    }
    c.mu.RUnlock()

    // Parse the expression
    p := parser.NewParser(expression)
    ast, err := p.Parse()
    if err != nil {
        return nil, fmt.Errorf("parse error: %w", err)
    }

    compiled := &CompiledExpression{
        ast:       ast,
        evaluator: evaluator.NewEvaluator(),
    }

    c.mu.Lock()
    c.cache.Put(expression, compiled)
    c.mu.Unlock()

    return compiled, nil
}

func (ce *CompiledExpression) Evaluate(ctx *evaluator.Context) ([]interface{}, error) {
    return ce.evaluator.Evaluate(ce.ast, ctx)
}
```

### 5.9 FHIRPath Public API

```go
// pkg/fhirpath/fhirpath.go

package fhirpath

import (
    "github.com/yourorg/fhir-toolkit-go/pkg/fhirpath/compiler"
    "github.com/yourorg/fhir-toolkit-go/pkg/fhirpath/evaluator"
)

var defaultCompiler = compiler.NewCompiler(500)

// Evaluate evaluates a FHIRPath expression against a resource
func Evaluate(expression string, resource interface{}) ([]interface{}, error) {
    compiled, err := defaultCompiler.Compile(expression)
    if err != nil {
        return nil, err
    }

    ctx := &evaluator.Context{
        Resource:    resource,
        RootContext: resource,
        Environment: make(map[string]interface{}),
    }

    return compiled.Evaluate(ctx)
}

// EvaluateWithContext evaluates with a custom context
func EvaluateWithContext(expression string, ctx *evaluator.Context) ([]interface{}, error) {
    compiled, err := defaultCompiler.Compile(expression)
    if err != nil {
        return nil, err
    }
    return compiled.Evaluate(ctx)
}

// EvaluateToBoolean evaluates and returns a boolean result
func EvaluateToBoolean(expression string, resource interface{}) (bool, error) {
    result, err := Evaluate(expression, resource)
    if err != nil {
        return false, err
    }
    return toBool(result), nil
}

// EvaluateToString evaluates and returns a string result
func EvaluateToString(expression string, resource interface{}) (string, error) {
    result, err := Evaluate(expression, resource)
    if err != nil {
        return "", err
    }
    if len(result) == 0 {
        return "", nil
    }
    return fmt.Sprintf("%v", result[0]), nil
}

// Compile pre-compiles an expression for repeated use
func Compile(expression string) (*compiler.CompiledExpression, error) {
    return defaultCompiler.Compile(expression)
}

// NewContext creates a new evaluation context
func NewContext(resource interface{}) *evaluator.Context {
    return &evaluator.Context{
        Resource:    resource,
        RootContext: resource,
        Environment: make(map[string]interface{}),
    }
}
```

---

## Part 6: Validation System (YAFV)

### 6.1 Main Validator

```go
// pkg/validator/validator.go

package validator

import (
    "context"
    "fmt"

    "github.com/yourorg/fhir-toolkit-go/pkg/fhirpath"
    "github.com/yourorg/fhir-toolkit-go/pkg/validator/validators"
)

type FHIRVersion string

const (
    FHIRVersionR4  FHIRVersion = "R4"
    FHIRVersionR4B FHIRVersion = "R4B"
    FHIRVersionR5  FHIRVersion = "R5"
)

type ValidatorOptions struct {
    FHIRVersion          FHIRVersion
    ValidateConstraints  bool   // FHIRPath constraints
    ValidateTerminology  bool   // ValueSet bindings
    ErrorOnWarning       bool   // Treat warnings as errors
    TerminologyServer    string // External terminology server URL
    TerminologyCacheTTL  int    // Cache timeout in seconds
    TerminologyCacheSize int    // Max cached entries
}

func DefaultOptions(version FHIRVersion) *ValidatorOptions {
    return &ValidatorOptions{
        FHIRVersion:          version,
        ValidateConstraints:  true,
        ValidateTerminology:  false,
        ErrorOnWarning:       false,
        TerminologyCacheTTL:  3600,  // 1 hour
        TerminologyCacheSize: 1000,
    }
}

type FHIRValidator struct {
    options      *ValidatorOptions
    specRegistry *SpecRegistry
    validators   []validators.Validator
}

func NewValidator(options *ValidatorOptions) (*FHIRValidator, error) {
    if options == nil {
        options = DefaultOptions(FHIRVersionR4)
    }

    registry, err := NewSpecRegistry(options.FHIRVersion)
    if err != nil {
        return nil, fmt.Errorf("failed to initialize spec registry: %w", err)
    }

    v := &FHIRValidator{
        options:      options,
        specRegistry: registry,
    }

    // Initialize validators
    v.validators = []validators.Validator{
        validators.NewStructureValidator(registry),
        validators.NewPrimitiveValidator(),
        validators.NewReferenceValidator(),
        validators.NewExtensionValidator(registry),
    }

    if options.ValidateConstraints {
        v.validators = append(v.validators,
            validators.NewConstraintValidator(registry))
    }

    if options.ValidateTerminology {
        termValidator, err := validators.NewTerminologyValidator(
            options.TerminologyServer,
            options.TerminologyCacheSize,
            options.TerminologyCacheTTL,
        )
        if err != nil {
            return nil, err
        }
        v.validators = append(v.validators, termValidator)
    }

    return v, nil
}

func (v *FHIRValidator) Validate(ctx context.Context, resource interface{}) (*OperationOutcome, error) {
    outcome := NewOperationOutcome()

    // Get resource type
    resourceType, err := getResourceType(resource)
    if err != nil {
        outcome.AddIssue(IssueSeverityError, IssueCodeInvalid,
            "Unable to determine resource type", nil)
        return outcome, nil
    }

    // Load StructureDefinition
    structDef, err := v.specRegistry.GetStructureDefinition(resourceType)
    if err != nil {
        outcome.AddIssue(IssueSeverityError, IssueCodeNotSupported,
            fmt.Sprintf("Unknown resource type: %s", resourceType), nil)
        return outcome, nil
    }

    // Run all validators
    validationCtx := &validators.ValidationContext{
        Resource:        resource,
        ResourceType:    resourceType,
        StructureDef:    structDef,
        SpecRegistry:    v.specRegistry,
        FHIRVersion:     string(v.options.FHIRVersion),
    }

    for _, validator := range v.validators {
        issues, err := validator.Validate(ctx, validationCtx)
        if err != nil {
            return nil, err
        }
        outcome.Issues = append(outcome.Issues, issues...)
    }

    return outcome, nil
}

func (v *FHIRValidator) ValidateBundle(ctx context.Context, bundle interface{}) (*OperationOutcome, error) {
    // Special handling for Bundle validation
    bundleValidator := validators.NewBundleValidator(v)
    return bundleValidator.Validate(ctx, bundle)
}

// Helper to get resource type from interface{}
func getResourceType(resource interface{}) (string, error) {
    switch r := resource.(type) {
    case map[string]interface{}:
        if rt, ok := r["resourceType"].(string); ok {
            return rt, nil
        }
    default:
        // Use reflection for struct types
        val := reflect.ValueOf(resource)
        if val.Kind() == reflect.Ptr {
            val = val.Elem()
        }
        if val.Kind() == reflect.Struct {
            if field := val.FieldByName("ResourceType"); field.IsValid() {
                return field.String(), nil
            }
        }
    }
    return "", fmt.Errorf("cannot determine resource type")
}
```

### 6.2 OperationOutcome Result

```go
// pkg/validator/result.go

package validator

type IssueSeverity string

const (
    IssueSeverityFatal       IssueSeverity = "fatal"
    IssueSeverityError       IssueSeverity = "error"
    IssueSeverityWarning     IssueSeverity = "warning"
    IssueSeverityInformation IssueSeverity = "information"
)

type IssueCode string

const (
    IssueCodeInvalid       IssueCode = "invalid"
    IssueCodeStructure     IssueCode = "structure"
    IssueCodeRequired      IssueCode = "required"
    IssueCodeValue         IssueCode = "value"
    IssueCodeInvariant     IssueCode = "invariant"
    IssueCodeNotSupported  IssueCode = "not-supported"
    IssueCodeBusinessRule  IssueCode = "business-rule"
    IssueCodeCodeInvalid   IssueCode = "code-invalid"
    IssueCodeExtension     IssueCode = "extension"
)

type OperationOutcomeIssue struct {
    Severity    IssueSeverity          `json:"severity"`
    Code        IssueCode              `json:"code"`
    Details     *CodeableConcept       `json:"details,omitempty"`
    Diagnostics string                 `json:"diagnostics,omitempty"`
    Location    []string               `json:"location,omitempty"`
    Expression  []string               `json:"expression,omitempty"`
}

type OperationOutcome struct {
    ResourceType string                  `json:"resourceType"`
    Issues       []OperationOutcomeIssue `json:"issue"`
}

func NewOperationOutcome() *OperationOutcome {
    return &OperationOutcome{
        ResourceType: "OperationOutcome",
        Issues:       []OperationOutcomeIssue{},
    }
}

func (o *OperationOutcome) AddIssue(severity IssueSeverity, code IssueCode, message string, path []string) {
    o.Issues = append(o.Issues, OperationOutcomeIssue{
        Severity:    severity,
        Code:        code,
        Diagnostics: message,
        Location:    path,
        Expression:  path,
    })
}

func (o *OperationOutcome) HasErrors() bool {
    for _, issue := range o.Issues {
        if issue.Severity == IssueSeverityError || issue.Severity == IssueSeverityFatal {
            return true
        }
    }
    return false
}

func (o *OperationOutcome) HasWarnings() bool {
    for _, issue := range o.Issues {
        if issue.Severity == IssueSeverityWarning {
            return true
        }
    }
    return false
}

func (o *OperationOutcome) ErrorCount() int {
    count := 0
    for _, issue := range o.Issues {
        if issue.Severity == IssueSeverityError || issue.Severity == IssueSeverityFatal {
            count++
        }
    }
    return count
}

func (o *OperationOutcome) IsSuccess() bool {
    return !o.HasErrors()
}
```

### 6.3 Constraint Validator (FHIRPath)

```go
// pkg/validator/validators/constraint.go

package validators

import (
    "context"
    "fmt"

    "github.com/yourorg/fhir-toolkit-go/pkg/fhirpath"
)

type ConstraintValidator struct {
    specRegistry    *SpecRegistry
    expressionCache *fhirpath.Compiler
}

func NewConstraintValidator(registry *SpecRegistry) *ConstraintValidator {
    return &ConstraintValidator{
        specRegistry:    registry,
        expressionCache: fhirpath.NewCompiler(500),
    }
}

func (v *ConstraintValidator) Validate(ctx context.Context, vctx *ValidationContext) ([]OperationOutcomeIssue, error) {
    var issues []OperationOutcomeIssue

    // Get constraints from StructureDefinition
    constraints := v.extractConstraints(vctx.StructureDef)

    for _, constraint := range constraints {
        result, err := v.evaluateConstraint(constraint, vctx.Resource, vctx.ResourceType)
        if err != nil {
            issues = append(issues, OperationOutcomeIssue{
                Severity:    IssueSeverityWarning,
                Code:        IssueCodeInvariant,
                Diagnostics: fmt.Sprintf("Error evaluating constraint %s: %v", constraint.Key, err),
                Expression:  []string{constraint.Context},
            })
            continue
        }

        if !result {
            severity := IssueSeverityError
            if constraint.Severity == "warning" {
                severity = IssueSeverityWarning
            }

            issues = append(issues, OperationOutcomeIssue{
                Severity:    severity,
                Code:        IssueCodeInvariant,
                Diagnostics: fmt.Sprintf("Constraint %s failed: %s", constraint.Key, constraint.Human),
                Expression:  []string{constraint.Context, constraint.Expression},
            })
        }
    }

    return issues, nil
}

type Constraint struct {
    Key        string
    Severity   string
    Human      string
    Expression string
    Context    string
}

func (v *ConstraintValidator) extractConstraints(structDef *StructureDefinition) []Constraint {
    var constraints []Constraint

    for _, element := range structDef.Snapshot.Element {
        for _, constraint := range element.Constraint {
            constraints = append(constraints, Constraint{
                Key:        constraint.Key,
                Severity:   constraint.Severity,
                Human:      constraint.Human,
                Expression: constraint.Expression,
                Context:    element.Path,
            })
        }
    }

    return constraints
}

func (v *ConstraintValidator) evaluateConstraint(constraint Constraint, resource interface{}, resourceType string) (bool, error) {
    compiled, err := v.expressionCache.Compile(constraint.Expression)
    if err != nil {
        return false, err
    }

    ctx := fhirpath.NewContext(resource)
    ctx.Environment["resource"] = resource
    ctx.Environment["rootResource"] = resource

    result, err := compiled.Evaluate(ctx)
    if err != nil {
        return false, err
    }

    // FHIRPath constraint returns true if satisfied
    return toBool(result), nil
}
```

### 6.4 Primitive Type Validator

```go
// pkg/validator/validators/primitive.go

package validators

import (
    "context"
    "regexp"
)

type PrimitiveValidator struct {
    patterns map[string]*regexp.Regexp
}

func NewPrimitiveValidator() *PrimitiveValidator {
    v := &PrimitiveValidator{
        patterns: make(map[string]*regexp.Regexp),
    }
    v.initPatterns()
    return v
}

func (v *PrimitiveValidator) initPatterns() {
    // FHIR primitive type patterns
    v.patterns["id"] = regexp.MustCompile(`^[A-Za-z0-9\-\.]{1,64}$`)
    v.patterns["uri"] = regexp.MustCompile(`^\S*$`)
    v.patterns["url"] = regexp.MustCompile(`^\S*$`)
    v.patterns["canonical"] = regexp.MustCompile(`^\S*$`)
    v.patterns["code"] = regexp.MustCompile(`^[^\s]+(\s[^\s]+)*$`)
    v.patterns["oid"] = regexp.MustCompile(`^urn:oid:[0-2](\.(0|[1-9][0-9]*))+$`)
    v.patterns["uuid"] = regexp.MustCompile(`^urn:uuid:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
    v.patterns["date"] = regexp.MustCompile(`^([0-9]([0-9]([0-9][1-9]|[1-9]0)|[1-9]00)|[1-9]000)(-(0[1-9]|1[0-2])(-(0[1-9]|[1-2][0-9]|3[0-1]))?)?$`)
    v.patterns["dateTime"] = regexp.MustCompile(`^([0-9]([0-9]([0-9][1-9]|[1-9]0)|[1-9]00)|[1-9]000)(-(0[1-9]|1[0-2])(-(0[1-9]|[1-2][0-9]|3[0-1])(T([01][0-9]|2[0-3]):[0-5][0-9]:([0-5][0-9]|60)(\.[0-9]+)?(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00)))?)?)?$`)
    v.patterns["time"] = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]:([0-5][0-9]|60)(\.[0-9]+)?$`)
    v.patterns["instant"] = regexp.MustCompile(`^([0-9]([0-9]([0-9][1-9]|[1-9]0)|[1-9]00)|[1-9]000)-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1])T([01][0-9]|2[0-3]):[0-5][0-9]:([0-5][0-9]|60)(\.[0-9]+)?(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00))$`)
    v.patterns["base64Binary"] = regexp.MustCompile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`)
    v.patterns["positiveInt"] = regexp.MustCompile(`^[1-9][0-9]*$`)
    v.patterns["unsignedInt"] = regexp.MustCompile(`^[0-9]+$`)
    v.patterns["markdown"] = regexp.MustCompile(`^[\s\S]+$`)
    v.patterns["xhtml"] = regexp.MustCompile(`^<div[^>]*>[\s\S]*</div>$`)
}

func (v *PrimitiveValidator) Validate(ctx context.Context, vctx *ValidationContext) ([]OperationOutcomeIssue, error) {
    var issues []OperationOutcomeIssue

    // Walk through all elements and validate primitive types
    issues = append(issues, v.validateElement(vctx.Resource, "", vctx.StructureDef)...)

    return issues, nil
}

func (v *PrimitiveValidator) validateElement(element interface{}, path string, structDef *StructureDefinition) []OperationOutcomeIssue {
    var issues []OperationOutcomeIssue

    switch val := element.(type) {
    case map[string]interface{}:
        for key, value := range val {
            elementPath := path + "." + key
            if path == "" {
                elementPath = key
            }
            issues = append(issues, v.validateElement(value, elementPath, structDef)...)
        }

    case []interface{}:
        for i, item := range val {
            elementPath := fmt.Sprintf("%s[%d]", path, i)
            issues = append(issues, v.validateElement(item, elementPath, structDef)...)
        }

    case string:
        // Look up element type from StructureDefinition
        elementType := v.getElementType(path, structDef)
        if pattern, ok := v.patterns[elementType]; ok {
            if !pattern.MatchString(val) {
                issues = append(issues, OperationOutcomeIssue{
                    Severity:    IssueSeverityError,
                    Code:        IssueCodeValue,
                    Diagnostics: fmt.Sprintf("Invalid %s value: %s", elementType, val),
                    Location:    []string{path},
                    Expression:  []string{path},
                })
            }
        }
    }

    return issues
}

func (v *PrimitiveValidator) getElementType(path string, structDef *StructureDefinition) string {
    // Look up element definition from StructureDefinition
    // Implementation omitted for brevity
    return ""
}
```

---

## Part 7: Code Generation

### 7.1 Generator Main

```go
// internal/codegen/generator.go

package codegen

import (
    "encoding/json"
    "os"
    "path/filepath"
    "text/template"
)

type Generator struct {
    specsDir    string
    outputDir   string
    version     string
    templates   *template.Template
}

func NewGenerator(specsDir, outputDir, version string) (*Generator, error) {
    g := &Generator{
        specsDir:  specsDir,
        outputDir: outputDir,
        version:   version,
    }

    if err := g.loadTemplates(); err != nil {
        return nil, err
    }

    return g, nil
}

func (g *Generator) Generate() error {
    // Load all StructureDefinitions
    structDefs, err := g.loadStructureDefinitions()
    if err != nil {
        return err
    }

    // Generate types
    if err := g.generateTypes(structDefs); err != nil {
        return err
    }

    // Generate models
    if err := g.generateModels(structDefs); err != nil {
        return err
    }

    // Generate builders
    if err := g.generateBuilders(structDefs); err != nil {
        return err
    }

    // Generate ValueSets
    valueSets, err := g.loadValueSets()
    if err != nil {
        return err
    }

    if err := g.generateValueSets(valueSets); err != nil {
        return err
    }

    return nil
}

func (g *Generator) loadStructureDefinitions() ([]*StructureDefinition, error) {
    var structDefs []*StructureDefinition

    pattern := filepath.Join(g.specsDir, "StructureDefinition-*.json")
    files, err := filepath.Glob(pattern)
    if err != nil {
        return nil, err
    }

    for _, file := range files {
        data, err := os.ReadFile(file)
        if err != nil {
            return nil, err
        }

        var sd StructureDefinition
        if err := json.Unmarshal(data, &sd); err != nil {
            return nil, err
        }

        structDefs = append(structDefs, &sd)
    }

    return structDefs, nil
}

func (g *Generator) generateTypes(structDefs []*StructureDefinition) error {
    for _, sd := range structDefs {
        if sd.Kind == "resource" || sd.Kind == "complex-type" {
            if err := g.generateType(sd); err != nil {
                return err
            }
        }
    }
    return nil
}

func (g *Generator) generateType(sd *StructureDefinition) error {
    // Parse elements and generate Go struct
    typeData := g.parseStructureDefinition(sd)

    outputPath := g.getOutputPath("types", sd.Kind, sd.Name)

    file, err := os.Create(outputPath)
    if err != nil {
        return err
    }
    defer file.Close()

    return g.templates.ExecuteTemplate(file, "struct.tmpl", typeData)
}

// Similar methods for generateModels and generateBuilders...
```

### 7.2 Template Example

```go
// internal/codegen/templates/struct.tmpl

// Code generated by fhir-toolkit-go codegen. DO NOT EDIT.

package {{.Package}}

{{if .Imports}}
import (
{{range .Imports}}
    "{{.}}"
{{end}}
)
{{end}}

// {{.Name}} {{.Description}}
type {{.Name}} struct {
{{- if .BaseType}}
    {{.BaseType}}
{{- end}}
{{range .Properties}}
    {{.GoName}} {{.GoType}} `json:"{{.JsonName}},omitempty"`
{{- if .HasExtension}}
    {{.GoName}}Ext *Element `json:"_{{.JsonName}},omitempty"`
{{- end}}
{{end}}
}

{{if .IsResource}}
// ResourceType returns the FHIR resource type
func (r *{{.Name}}) GetResourceType() string {
    return "{{.Name}}"
}
{{end}}
```

---

## Part 8: Testing Strategy

### 8.1 Model Tests

```go
// pkg/r4/models/resources/patient_test.go

package resources_test

import (
    "encoding/json"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types/valuesets"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/models/resources"
)

func TestPatientConstruction(t *testing.T) {
    patient := resources.NewPatient(&types.Patient{
        ID:     strPtr("test-123"),
        Active: boolPtr(true),
        Gender: genderPtr(valuesets.AdministrativeGenderMale),
        Name: []types.HumanName{
            {Family: strPtr("Smith"), Given: []string{"John"}},
        },
    })

    assert.Equal(t, "test-123", *patient.ID())
    assert.Equal(t, true, *patient.Active())
    assert.Equal(t, valuesets.AdministrativeGenderMale, *patient.Gender())
    assert.Len(t, patient.Names(), 1)
    assert.Equal(t, "Smith", *patient.Names()[0].Family)
}

func TestPatientSerialization(t *testing.T) {
    patient := resources.NewPatient(&types.Patient{
        ID:        strPtr("test-123"),
        Active:    boolPtr(true),
        BirthDate: strPtr("1990-01-15"),
    })

    jsonData, err := patient.ToJSON()
    require.NoError(t, err)

    var parsed map[string]interface{}
    err = json.Unmarshal(jsonData, &parsed)
    require.NoError(t, err)

    assert.Equal(t, "Patient", parsed["resourceType"])
    assert.Equal(t, "test-123", parsed["id"])
    assert.Equal(t, true, parsed["active"])
    assert.Equal(t, "1990-01-15", parsed["birthDate"])
}

func TestPatientRoundTrip(t *testing.T) {
    original := resources.NewPatient(&types.Patient{
        ID:     strPtr("roundtrip-test"),
        Active: boolPtr(true),
        Gender: genderPtr(valuesets.AdministrativeGenderFemale),
    })

    // Serialize
    jsonData, err := original.ToJSON()
    require.NoError(t, err)

    // Deserialize
    var parsed types.Patient
    err = json.Unmarshal(jsonData, &parsed)
    require.NoError(t, err)

    restored := resources.NewPatient(&parsed)

    assert.Equal(t, *original.ID(), *restored.ID())
    assert.Equal(t, *original.Active(), *restored.Active())
    assert.Equal(t, *original.Gender(), *restored.Gender())
}

func TestPatientClone(t *testing.T) {
    original := resources.NewPatient(&types.Patient{
        ID:     strPtr("clone-test"),
        Active: boolPtr(true),
    })

    clone := original.Clone()

    // Modify clone
    clone.SetActive(false)

    // Original should be unchanged
    assert.Equal(t, true, *original.Active())
    assert.Equal(t, false, *clone.Active())
}

func TestPatientImmutableWith(t *testing.T) {
    original := resources.NewPatient(&types.Patient{
        ID:     strPtr("with-test"),
        Active: boolPtr(true),
    })

    modified := original.With(func(p *resources.Patient) {
        p.SetActive(false)
        p.SetGender(valuesets.AdministrativeGenderOther)
    })

    // Original unchanged
    assert.Equal(t, true, *original.Active())
    assert.Nil(t, original.Gender())

    // Modified has changes
    assert.Equal(t, false, *modified.Active())
    assert.Equal(t, valuesets.AdministrativeGenderOther, *modified.Gender())
}

// Helper functions
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool { return &b }
func genderPtr(g valuesets.AdministrativeGender) *valuesets.AdministrativeGender { return &g }
```

### 8.2 Builder Tests

```go
// pkg/r4/builders/resources/patient_builder_test.go

package resources_test

import (
    "testing"

    "github.com/stretchr/testify/assert"

    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types/valuesets"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/builders/resources"
)

func TestPatientBuilderFluent(t *testing.T) {
    patient := resources.NewPatientBuilder().
        SetID("builder-test").
        SetActive(true).
        SetGender(valuesets.AdministrativeGenderMale).
        SetBirthDate("1985-06-15").
        AddName(types.HumanName{
            Use:    strPtr("official"),
            Family: strPtr("Builder"),
            Given:  []string{"Test", "User"},
        }).
        AddIdentifier(types.Identifier{
            System: strPtr("http://example.com/mrn"),
            Value:  strPtr("MRN-12345"),
        }).
        Build()

    assert.Equal(t, "builder-test", *patient.ID())
    assert.Equal(t, true, *patient.Active())
    assert.Equal(t, valuesets.AdministrativeGenderMale, *patient.Gender())
    assert.Equal(t, "1985-06-15", *patient.BirthDate())
    assert.Len(t, patient.Names(), 1)
    assert.Len(t, patient.Identifiers(), 1)
}

func TestPatientBuilderChoiceTypes(t *testing.T) {
    // Test deceased[x] as boolean
    patient1 := resources.NewPatientBuilder().
        SetID("deceased-bool").
        SetDeceasedBoolean(true).
        Build()

    deceased, deceasedType := patient1.Deceased()
    assert.Equal(t, "Boolean", deceasedType)
    assert.Equal(t, true, deceased)

    // Test deceased[x] as dateTime
    patient2 := resources.NewPatientBuilder().
        SetID("deceased-datetime").
        SetDeceasedDateTime("2023-01-15").
        Build()

    deceased2, deceasedType2 := patient2.Deceased()
    assert.Equal(t, "DateTime", deceasedType2)
    assert.Equal(t, "2023-01-15", deceased2)
}
```

### 8.3 FHIRPath Tests

```go
// pkg/fhirpath/fhirpath_test.go

package fhirpath_test

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/yourorg/fhir-toolkit-go/pkg/fhirpath"
)

func TestFHIRPathBasicNavigation(t *testing.T) {
    patient := map[string]interface{}{
        "resourceType": "Patient",
        "id":           "test-123",
        "active":       true,
        "name": []interface{}{
            map[string]interface{}{
                "family": "Smith",
                "given":  []interface{}{"John", "William"},
            },
        },
    }

    tests := []struct {
        expr     string
        expected interface{}
    }{
        {"id", "test-123"},
        {"active", true},
        {"name.family", "Smith"},
        {"name.given.first()", "John"},
        {"name.given.count()", 2},
        {"active.not()", false},
    }

    for _, tt := range tests {
        t.Run(tt.expr, func(t *testing.T) {
            result, err := fhirpath.Evaluate(tt.expr, patient)
            require.NoError(t, err)
            require.Len(t, result, 1)
            assert.Equal(t, tt.expected, result[0])
        })
    }
}

func TestFHIRPathExists(t *testing.T) {
    patient := map[string]interface{}{
        "resourceType": "Patient",
        "id":           "test",
        "name": []interface{}{
            map[string]interface{}{"family": "Smith"},
        },
    }

    tests := []struct {
        expr     string
        expected bool
    }{
        {"id.exists()", true},
        {"name.exists()", true},
        {"birthDate.exists()", false},
        {"name.family.exists()", true},
        {"name.given.exists()", false},
    }

    for _, tt := range tests {
        t.Run(tt.expr, func(t *testing.T) {
            result, err := fhirpath.EvaluateToBoolean(tt.expr, patient)
            require.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestFHIRPathWhere(t *testing.T) {
    bundle := map[string]interface{}{
        "resourceType": "Bundle",
        "entry": []interface{}{
            map[string]interface{}{
                "resource": map[string]interface{}{
                    "resourceType": "Patient",
                    "id":           "p1",
                    "active":       true,
                },
            },
            map[string]interface{}{
                "resource": map[string]interface{}{
                    "resourceType": "Patient",
                    "id":           "p2",
                    "active":       false,
                },
            },
            map[string]interface{}{
                "resource": map[string]interface{}{
                    "resourceType": "Observation",
                    "id":           "o1",
                },
            },
        },
    }

    // Find active patients
    result, err := fhirpath.Evaluate(
        "entry.resource.where(resourceType = 'Patient' and active = true).id",
        bundle,
    )
    require.NoError(t, err)
    assert.Equal(t, []interface{}{"p1"}, result)
}

func TestFHIRPathConstraints(t *testing.T) {
    // Test common FHIR constraints

    // obs-6: dataAbsentReason SHALL only be present if value is not present
    observation := map[string]interface{}{
        "resourceType":      "Observation",
        "status":            "final",
        "code":              map[string]interface{}{"text": "Test"},
        "valueQuantity":     map[string]interface{}{"value": 100},
        "dataAbsentReason":  map[string]interface{}{"text": "Unknown"}, // Invalid!
    }

    expr := "dataAbsentReason.empty() or value.empty()"
    result, err := fhirpath.EvaluateToBoolean(expr, observation)
    require.NoError(t, err)
    assert.False(t, result) // Constraint should fail

    // Valid observation
    validObs := map[string]interface{}{
        "resourceType": "Observation",
        "status":       "final",
        "code":         map[string]interface{}{"text": "Test"},
        "valueQuantity": map[string]interface{}{"value": 100},
    }

    result, err = fhirpath.EvaluateToBoolean(expr, validObs)
    require.NoError(t, err)
    assert.True(t, result) // Constraint should pass
}

func TestFHIRPathFunctions(t *testing.T) {
    data := map[string]interface{}{
        "values": []interface{}{1, 2, 3, 4, 5},
        "text":   "Hello World",
    }

    tests := []struct {
        expr     string
        expected interface{}
    }{
        // Collection functions
        {"values.count()", 5},
        {"values.first()", 1},
        {"values.last()", 5},
        {"values.tail().first()", 2},
        {"values.take(3).count()", 3},
        {"values.skip(2).first()", 3},

        // String functions
        {"text.startsWith('Hello')", true},
        {"text.endsWith('World')", true},
        {"text.contains('lo Wo')", true},
        {"text.length()", 11},
        {"text.upper()", "HELLO WORLD"},
        {"text.lower()", "hello world"},

        // Math functions
        {"values.sum()", 15},
        {"values.min()", 1},
        {"values.max()", 5},
    }

    for _, tt := range tests {
        t.Run(tt.expr, func(t *testing.T) {
            result, err := fhirpath.Evaluate(tt.expr, data)
            require.NoError(t, err)
            require.Len(t, result, 1)
            assert.Equal(t, tt.expected, result[0])
        })
    }
}
```

### 8.4 Validation Tests

```go
// pkg/validator/validator_test.go

package validator_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "github.com/yourorg/fhir-toolkit-go/pkg/validator"
)

func TestValidatorBasic(t *testing.T) {
    v, err := validator.NewValidator(validator.DefaultOptions(validator.FHIRVersionR4))
    require.NoError(t, err)

    patient := map[string]interface{}{
        "resourceType": "Patient",
        "id":           "test-123",
        "active":       true,
    }

    outcome, err := v.Validate(context.Background(), patient)
    require.NoError(t, err)
    assert.True(t, outcome.IsSuccess())
}

func TestValidatorMissingRequired(t *testing.T) {
    v, err := validator.NewValidator(validator.DefaultOptions(validator.FHIRVersionR4))
    require.NoError(t, err)

    // Observation without required 'status' and 'code'
    observation := map[string]interface{}{
        "resourceType": "Observation",
        "id":           "obs-123",
    }

    outcome, err := v.Validate(context.Background(), observation)
    require.NoError(t, err)
    assert.True(t, outcome.HasErrors())

    // Should have errors for missing status and code
    assert.GreaterOrEqual(t, outcome.ErrorCount(), 2)
}

func TestValidatorInvalidPrimitive(t *testing.T) {
    v, err := validator.NewValidator(validator.DefaultOptions(validator.FHIRVersionR4))
    require.NoError(t, err)

    patient := map[string]interface{}{
        "resourceType": "Patient",
        "id":           "invalid id with spaces!", // Invalid id format
        "birthDate":    "not-a-date",              // Invalid date format
    }

    outcome, err := v.Validate(context.Background(), patient)
    require.NoError(t, err)
    assert.True(t, outcome.HasErrors())
}

func TestValidatorConstraints(t *testing.T) {
    opts := validator.DefaultOptions(validator.FHIRVersionR4)
    opts.ValidateConstraints = true

    v, err := validator.NewValidator(opts)
    require.NoError(t, err)

    // Invalid: both value and dataAbsentReason present
    observation := map[string]interface{}{
        "resourceType": "Observation",
        "status":       "final",
        "code":         map[string]interface{}{"text": "Test"},
        "valueQuantity": map[string]interface{}{
            "value": 100,
            "unit":  "mg",
        },
        "dataAbsentReason": map[string]interface{}{
            "text": "Unknown",
        },
    }

    outcome, err := v.Validate(context.Background(), observation)
    require.NoError(t, err)
    assert.True(t, outcome.HasErrors())

    // Check for constraint violation
    hasConstraintError := false
    for _, issue := range outcome.Issues {
        if issue.Code == validator.IssueCodeInvariant {
            hasConstraintError = true
            break
        }
    }
    assert.True(t, hasConstraintError)
}
```

---

## Part 9: CLI Tool

```go
// cmd/fhir-cli/main.go

package main

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "github.com/yourorg/fhir-toolkit-go/pkg/validator"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "fhir-cli",
        Short: "FHIR Toolkit CLI",
    }

    validateCmd := &cobra.Command{
        Use:   "validate [file]",
        Short: "Validate a FHIR resource",
        Args:  cobra.ExactArgs(1),
        RunE:  runValidate,
    }

    validateCmd.Flags().StringP("version", "v", "R4", "FHIR version (R4, R4B, R5)")
    validateCmd.Flags().Bool("constraints", true, "Validate FHIRPath constraints")
    validateCmd.Flags().Bool("terminology", false, "Validate terminology bindings")

    rootCmd.AddCommand(validateCmd)

    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func runValidate(cmd *cobra.Command, args []string) error {
    filePath := args[0]

    version, _ := cmd.Flags().GetString("version")
    constraints, _ := cmd.Flags().GetBool("constraints")
    terminology, _ := cmd.Flags().GetBool("terminology")

    // Read file
    data, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("failed to read file: %w", err)
    }

    var resource interface{}
    if err := json.Unmarshal(data, &resource); err != nil {
        return fmt.Errorf("failed to parse JSON: %w", err)
    }

    // Create validator
    opts := &validator.ValidatorOptions{
        FHIRVersion:         validator.FHIRVersion(version),
        ValidateConstraints: constraints,
        ValidateTerminology: terminology,
    }

    v, err := validator.NewValidator(opts)
    if err != nil {
        return fmt.Errorf("failed to create validator: %w", err)
    }

    // Validate
    outcome, err := v.Validate(context.Background(), resource)
    if err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    // Output results
    outputJSON, _ := json.MarshalIndent(outcome, "", "  ")
    fmt.Println(string(outputJSON))

    if outcome.HasErrors() {
        os.Exit(1)
    }

    return nil
}
```

---

## Part 10: Implementation Priorities

### Phase 1: Core Foundation
1. Set up Go module structure
2. Implement base types (Element, Resource, DomainResource)
3. Implement primitive type handling with extensions
4. Create code generation pipeline from StructureDefinitions
5. Generate R4 types for all resources and datatypes

### Phase 2: Models & Builders
1. Implement model base classes with serialization
2. Generate all R4 models with proper JSON handling
3. Implement builder pattern with fluent API
4. Generate all R4 builders
5. Add clone, with, and immutable update methods

### Phase 3: FHIRPath Engine
1. Implement lexer and parser
2. Build AST representation
3. Implement expression evaluator
4. Add all standard FHIRPath functions
5. Implement expression caching with LRU

### Phase 4: Validation System
1. Implement SpecRegistry for loading definitions
2. Build structural validator
3. Implement primitive type validation
4. Integrate FHIRPath constraint validation
5. Add extension and reference validation

### Phase 5: Multi-Version & Extras
1. Generate R4B and R5 packages
2. Implement terminology validation
3. Add Bundle validation
4. Build CLI tool
5. Implement Implementation Guide loading

---

## Key Considerations for Go Implementation

### 1. Performance Optimizations
- Use sync.Pool for frequently allocated objects
- Implement efficient JSON marshaling with custom encoders
- Use concurrent validation for large resources
- Cache compiled FHIRPath expressions

### 2. Memory Management
- Use pointer types judiciously to reduce memory
- Implement lazy loading for large spec files
- Use streaming JSON parsing for large files

### 3. Concurrency
- Make validators goroutine-safe
- Use context.Context for cancellation
- Implement concurrent Bundle validation

### 4. Go Idioms
- Use interfaces for extensibility
- Follow Go naming conventions
- Use errors instead of exceptions
- Implement proper error wrapping

### 5. Testing
- Use table-driven tests
- Implement benchmarks for performance-critical code
- Use testify for assertions
- Aim for >80% code coverage

---

---

## Part 11: Advanced JSON Serialization

### 11.1 FHIR Property Ordering

FHIR requires properties to be serialized in a specific order. Go's `encoding/json` doesn't guarantee order, so implement custom marshaling:

```go
// pkg/common/json/ordered_marshal.go

package json

import (
    "bytes"
    "encoding/json"
    "reflect"
    "sort"
)

// PropertyOrder defines the FHIR-compliant property order for each resource type
var PropertyOrder = map[string][]string{
    "Patient": {
        "resourceType", "id", "meta", "implicitRules", "_implicitRules",
        "language", "_language", "text", "contained", "extension",
        "modifierExtension", "identifier", "active", "_active", "name",
        "telecom", "gender", "_gender", "birthDate", "_birthDate",
        "deceased", "address", "maritalStatus", "multipleBirth",
        "photo", "contact", "communication", "generalPractitioner",
        "managingOrganization", "link",
    },
    "Observation": {
        "resourceType", "id", "meta", "implicitRules", "_implicitRules",
        "language", "_language", "text", "contained", "extension",
        "modifierExtension", "identifier", "basedOn", "partOf", "status",
        "_status", "category", "code", "subject", "focus", "encounter",
        "effective", "issued", "_issued", "performer", "value",
        "dataAbsentReason", "interpretation", "note", "bodySite", "method",
        "specimen", "device", "referenceRange", "hasMember", "derivedFrom",
        "component",
    },
    // ... more resource types
}

// OrderedMarshal serializes a FHIR resource with properties in FHIR-compliant order
func OrderedMarshal(resourceType string, data map[string]interface{}) ([]byte, error) {
    order, exists := PropertyOrder[resourceType]
    if !exists {
        // Fall back to alphabetical if no order defined
        return json.Marshal(data)
    }

    var buf bytes.Buffer
    buf.WriteByte('{')

    first := true
    written := make(map[string]bool)

    // Write properties in defined order
    for _, key := range order {
        if val, ok := data[key]; ok && !isEmptyValue(val) {
            if !first {
                buf.WriteByte(',')
            }
            first = false
            written[key] = true

            keyJSON, _ := json.Marshal(key)
            valJSON, _ := json.Marshal(val)
            buf.Write(keyJSON)
            buf.WriteByte(':')
            buf.Write(valJSON)
        }
    }

    // Write any remaining properties (extensions, custom fields)
    var remaining []string
    for key := range data {
        if !written[key] {
            remaining = append(remaining, key)
        }
    }
    sort.Strings(remaining)

    for _, key := range remaining {
        val := data[key]
        if !isEmptyValue(val) {
            if !first {
                buf.WriteByte(',')
            }
            first = false

            keyJSON, _ := json.Marshal(key)
            valJSON, _ := json.Marshal(val)
            buf.Write(keyJSON)
            buf.WriteByte(':')
            buf.Write(valJSON)
        }
    }

    buf.WriteByte('}')
    return buf.Bytes(), nil
}

func isEmptyValue(v interface{}) bool {
    if v == nil {
        return true
    }

    val := reflect.ValueOf(v)
    switch val.Kind() {
    case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
        return val.Len() == 0
    case reflect.Ptr, reflect.Interface:
        return val.IsNil()
    }
    return false
}
```

### 11.2 Primitive Extension Serialization

FHIR primitives can have extensions via underscore-prefixed properties:

```go
// pkg/common/json/primitive.go

package json

import (
    "encoding/json"
)

// PrimitiveWithExtension represents a FHIR primitive that may have extensions
type PrimitiveWithExtension[T any] struct {
    Value     *T
    ID        *string
    Extension []Extension
}

// MarshalFHIR returns both the value and extension as separate JSON properties
func (p PrimitiveWithExtension[T]) MarshalFHIR(propertyName string) (map[string]json.RawMessage, error) {
    result := make(map[string]json.RawMessage)

    // Marshal the value
    if p.Value != nil {
        valueJSON, err := json.Marshal(p.Value)
        if err != nil {
            return nil, err
        }
        result[propertyName] = valueJSON
    }

    // Marshal the extension element
    if p.ID != nil || len(p.Extension) > 0 {
        extElement := struct {
            ID        *string     `json:"id,omitempty"`
            Extension []Extension `json:"extension,omitempty"`
        }{
            ID:        p.ID,
            Extension: p.Extension,
        }
        extJSON, err := json.Marshal(extElement)
        if err != nil {
            return nil, err
        }
        result["_"+propertyName] = extJSON
    }

    return result, nil
}

// Example usage in a Patient model:
type PatientData struct {
    BirthDate    PrimitiveWithExtension[string]
    Active       PrimitiveWithExtension[bool]
    // ... other fields
}

func (p *PatientData) MarshalJSON() ([]byte, error) {
    result := make(map[string]interface{})
    result["resourceType"] = "Patient"

    // Handle birthDate with potential extension
    if p.BirthDate.Value != nil {
        result["birthDate"] = *p.BirthDate.Value
    }
    if p.BirthDate.ID != nil || len(p.BirthDate.Extension) > 0 {
        result["_birthDate"] = map[string]interface{}{
            "id":        p.BirthDate.ID,
            "extension": p.BirthDate.Extension,
        }
    }

    return OrderedMarshal("Patient", result)
}
```

### 11.3 Choice Type JSON Handling

```go
// pkg/common/json/choice.go

package json

import (
    "encoding/json"
    "fmt"
    "reflect"
    "strings"
)

// ChoiceType represents FHIR's polymorphic value[x] pattern
type ChoiceType struct {
    TypeName string
    Value    interface{}
}

// MarshalChoice serializes a choice type with the correct property name
func (c ChoiceType) MarshalChoice(baseName string) (string, json.RawMessage, error) {
    if c.Value == nil {
        return "", nil, nil
    }

    propertyName := baseName + c.TypeName
    valueJSON, err := json.Marshal(c.Value)
    if err != nil {
        return "", nil, err
    }

    return propertyName, valueJSON, nil
}

// UnmarshalChoice deserializes a choice type from JSON
func UnmarshalChoice(data map[string]json.RawMessage, baseName string, allowedTypes map[string]reflect.Type) (*ChoiceType, error) {
    for typeName, targetType := range allowedTypes {
        propertyName := baseName + typeName
        if rawValue, exists := data[propertyName]; exists {
            value := reflect.New(targetType).Interface()
            if err := json.Unmarshal(rawValue, value); err != nil {
                return nil, err
            }
            return &ChoiceType{
                TypeName: typeName,
                Value:    reflect.ValueOf(value).Elem().Interface(),
            }, nil
        }
    }
    return nil, nil // No choice value present
}

// Observation.value[x] allowed types
var ObservationValueTypes = map[string]reflect.Type{
    "Quantity":         reflect.TypeOf(Quantity{}),
    "CodeableConcept":  reflect.TypeOf(CodeableConcept{}),
    "String":           reflect.TypeOf(""),
    "Boolean":          reflect.TypeOf(true),
    "Integer":          reflect.TypeOf(0),
    "Range":            reflect.TypeOf(Range{}),
    "Ratio":            reflect.TypeOf(Ratio{}),
    "SampledData":      reflect.TypeOf(SampledData{}),
    "Time":             reflect.TypeOf(""),
    "DateTime":         reflect.TypeOf(""),
    "Period":           reflect.TypeOf(Period{}),
}
```

---

## Part 12: Concurrency Patterns

### 12.1 Thread-Safe Caches

```go
// pkg/common/cache/lru.go

package cache

import (
    "container/list"
    "sync"
)

// LRUCache is a thread-safe LRU cache
type LRUCache[K comparable, V any] struct {
    capacity int
    cache    map[K]*list.Element
    list     *list.List
    mu       sync.RWMutex
}

type entry[K comparable, V any] struct {
    key   K
    value V
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
    return &LRUCache[K, V]{
        capacity: capacity,
        cache:    make(map[K]*list.Element),
        list:     list.New(),
    }
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
    c.mu.RLock()
    elem, exists := c.cache[key]
    c.mu.RUnlock()

    if !exists {
        var zero V
        return zero, false
    }

    c.mu.Lock()
    c.list.MoveToFront(elem)
    c.mu.Unlock()

    return elem.Value.(*entry[K, V]).value, true
}

func (c *LRUCache[K, V]) Put(key K, value V) {
    c.mu.Lock()
    defer c.mu.Unlock()

    if elem, exists := c.cache[key]; exists {
        c.list.MoveToFront(elem)
        elem.Value.(*entry[K, V]).value = value
        return
    }

    if c.list.Len() >= c.capacity {
        // Remove oldest
        oldest := c.list.Back()
        if oldest != nil {
            c.list.Remove(oldest)
            delete(c.cache, oldest.Value.(*entry[K, V]).key)
        }
    }

    elem := c.list.PushFront(&entry[K, V]{key: key, value: value})
    c.cache[key] = elem
}

func (c *LRUCache[K, V]) Len() int {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.list.Len()
}

func (c *LRUCache[K, V]) Clear() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.cache = make(map[K]*list.Element)
    c.list.Init()
}
```

### 12.2 Concurrent Bundle Validation

```go
// pkg/validator/validators/bundle.go

package validators

import (
    "context"
    "sync"
)

type BundleValidator struct {
    validator      *FHIRValidator
    maxConcurrency int
}

func NewBundleValidator(v *FHIRValidator, maxConcurrency int) *BundleValidator {
    if maxConcurrency <= 0 {
        maxConcurrency = 10 // Default
    }
    return &BundleValidator{
        validator:      v,
        maxConcurrency: maxConcurrency,
    }
}

type entryResult struct {
    index  int
    issues []OperationOutcomeIssue
    err    error
}

func (v *BundleValidator) Validate(ctx context.Context, bundle map[string]interface{}) (*OperationOutcome, error) {
    outcome := NewOperationOutcome()

    // Validate bundle-level properties
    if err := v.validateBundleType(bundle, outcome); err != nil {
        return nil, err
    }

    // Get entries
    entries, ok := bundle["entry"].([]interface{})
    if !ok || len(entries) == 0 {
        return outcome, nil
    }

    // Concurrent validation with semaphore
    sem := make(chan struct{}, v.maxConcurrency)
    results := make(chan entryResult, len(entries))
    var wg sync.WaitGroup

    for i, entry := range entries {
        wg.Add(1)
        go func(idx int, e interface{}) {
            defer wg.Done()

            // Acquire semaphore
            select {
            case sem <- struct{}{}:
                defer func() { <-sem }()
            case <-ctx.Done():
                results <- entryResult{index: idx, err: ctx.Err()}
                return
            }

            issues, err := v.validateEntry(ctx, e, idx)
            results <- entryResult{index: idx, issues: issues, err: err}
        }(i, entry)
    }

    // Close results channel when all goroutines complete
    go func() {
        wg.Wait()
        close(results)
    }()

    // Collect results
    for result := range results {
        if result.err != nil {
            if result.err == context.Canceled || result.err == context.DeadlineExceeded {
                outcome.AddIssue(IssueSeverityWarning, IssueCodeTimeout,
                    "Validation cancelled", []string{fmt.Sprintf("Bundle.entry[%d]", result.index)})
            } else {
                return nil, result.err
            }
        }
        outcome.Issues = append(outcome.Issues, result.issues...)
    }

    // Sort issues by entry index for deterministic output
    sort.Slice(outcome.Issues, func(i, j int) bool {
        return outcome.Issues[i].Location[0] < outcome.Issues[j].Location[0]
    })

    return outcome, nil
}

func (v *BundleValidator) validateEntry(ctx context.Context, entry interface{}, index int) ([]OperationOutcomeIssue, error) {
    var issues []OperationOutcomeIssue
    entryPath := fmt.Sprintf("Bundle.entry[%d]", index)

    entryMap, ok := entry.(map[string]interface{})
    if !ok {
        issues = append(issues, OperationOutcomeIssue{
            Severity:    IssueSeverityError,
            Code:        IssueCodeStructure,
            Diagnostics: "Entry is not a valid object",
            Location:    []string{entryPath},
        })
        return issues, nil
    }

    // Validate resource within entry
    if resource, ok := entryMap["resource"]; ok {
        resourceOutcome, err := v.validator.Validate(ctx, resource)
        if err != nil {
            return nil, err
        }

        // Prefix issue locations with entry path
        for _, issue := range resourceOutcome.Issues {
            issue.Location = prefixPaths(issue.Location, entryPath+".resource")
            issue.Expression = prefixPaths(issue.Expression, entryPath+".resource")
            issues = append(issues, issue)
        }
    }

    return issues, nil
}

func prefixPaths(paths []string, prefix string) []string {
    result := make([]string, len(paths))
    for i, path := range paths {
        result[i] = prefix + "." + path
    }
    return result
}
```

### 12.3 Parallel Spec Loading

```go
// pkg/validator/spec_registry.go

package validator

import (
    "context"
    "embed"
    "encoding/json"
    "fmt"
    "sync"
)

//go:embed specs/r4/*.json
var r4Specs embed.FS

//go:embed specs/r4b/*.json
var r4bSpecs embed.FS

//go:embed specs/r5/*.json
var r5Specs embed.FS

type SpecRegistry struct {
    version       FHIRVersion
    structureDefs *LRUCache[string, *StructureDefinition]
    valueSets     *LRUCache[string, *ValueSet]
    codeSystems   *LRUCache[string, *CodeSystem]
    specs         embed.FS
    mu            sync.RWMutex
    loading       map[string]chan struct{}
}

func NewSpecRegistry(version FHIRVersion) (*SpecRegistry, error) {
    var specs embed.FS
    switch version {
    case FHIRVersionR4:
        specs = r4Specs
    case FHIRVersionR4B:
        specs = r4bSpecs
    case FHIRVersionR5:
        specs = r5Specs
    default:
        return nil, fmt.Errorf("unsupported FHIR version: %s", version)
    }

    return &SpecRegistry{
        version:       version,
        structureDefs: NewLRUCache[string, *StructureDefinition](200),
        valueSets:     NewLRUCache[string, *ValueSet](500),
        codeSystems:   NewLRUCache[string, *CodeSystem](100),
        specs:         specs,
        loading:       make(map[string]chan struct{}),
    }, nil
}

func (r *SpecRegistry) GetStructureDefinition(resourceType string) (*StructureDefinition, error) {
    // Check cache first
    if sd, ok := r.structureDefs.Get(resourceType); ok {
        return sd, nil
    }

    // Acquire loading lock to prevent duplicate loads
    r.mu.Lock()
    if ch, loading := r.loading[resourceType]; loading {
        r.mu.Unlock()
        <-ch // Wait for other goroutine to finish loading
        return r.structureDefs.Get(resourceType)
    }

    ch := make(chan struct{})
    r.loading[resourceType] = ch
    r.mu.Unlock()

    defer func() {
        r.mu.Lock()
        delete(r.loading, resourceType)
        close(ch)
        r.mu.Unlock()
    }()

    // Load from embedded files
    filename := fmt.Sprintf("specs/%s/StructureDefinition-%s.json",
        strings.ToLower(string(r.version)), resourceType)

    data, err := r.specs.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("StructureDefinition not found: %s", resourceType)
    }

    var sd StructureDefinition
    if err := json.Unmarshal(data, &sd); err != nil {
        return nil, fmt.Errorf("failed to parse StructureDefinition: %w", err)
    }

    r.structureDefs.Put(resourceType, &sd)
    return &sd, nil
}

// PreloadCommonSpecs loads frequently used specifications in parallel
func (r *SpecRegistry) PreloadCommonSpecs(ctx context.Context) error {
    commonResources := []string{
        "Patient", "Observation", "Encounter", "Condition",
        "Procedure", "MedicationRequest", "DiagnosticReport",
        "Immunization", "AllergyIntolerance", "CarePlan",
    }

    var wg sync.WaitGroup
    errCh := make(chan error, len(commonResources))

    for _, resource := range commonResources {
        wg.Add(1)
        go func(r string) {
            defer wg.Done()
            select {
            case <-ctx.Done():
                errCh <- ctx.Err()
                return
            default:
            }
            if _, err := r.GetStructureDefinition(r); err != nil {
                errCh <- err
            }
        }(resource)
    }

    wg.Wait()
    close(errCh)

    // Return first error if any
    for err := range errCh {
        if err != nil {
            return err
        }
    }

    return nil
}
```

---

## Part 13: Implementation Guide Support

### 13.1 IG Loader

```go
// pkg/validator/ig_loader.go

package validator

import (
    "archive/tar"
    "compress/gzip"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "strings"
)

type ImplementationGuide struct {
    ID              string
    Version         string
    FHIRVersion     string
    StructureDefs   map[string]*StructureDefinition
    ValueSets       map[string]*ValueSet
    CodeSystems     map[string]*CodeSystem
    SearchParams    map[string]*SearchParameter
}

type IGLoader struct {
    cacheDir string
    client   *http.Client
}

func NewIGLoader(cacheDir string) *IGLoader {
    if cacheDir == "" {
        cacheDir = filepath.Join(os.TempDir(), "fhir-igs")
    }
    os.MkdirAll(cacheDir, 0755)

    return &IGLoader{
        cacheDir: cacheDir,
        client:   &http.Client{Timeout: 60 * time.Second},
    }
}

// LoadFromPackage loads an IG from a FHIR package (.tgz)
func (l *IGLoader) LoadFromPackage(path string) (*ImplementationGuide, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("failed to open package: %w", err)
    }
    defer file.Close()

    return l.loadFromTarGz(file)
}

// LoadFromRegistry loads an IG from the FHIR package registry
func (l *IGLoader) LoadFromRegistry(packageName, version string) (*ImplementationGuide, error) {
    // Check cache first
    cachePath := filepath.Join(l.cacheDir, fmt.Sprintf("%s-%s.tgz", packageName, version))
    if _, err := os.Stat(cachePath); err == nil {
        return l.LoadFromPackage(cachePath)
    }

    // Download from registry
    url := fmt.Sprintf("https://packages.fhir.org/%s/%s", packageName, version)
    resp, err := l.client.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to download package: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("package not found: %s@%s", packageName, version)
    }

    // Save to cache
    cacheFile, err := os.Create(cachePath)
    if err != nil {
        return nil, err
    }
    defer cacheFile.Close()

    if _, err := io.Copy(cacheFile, resp.Body); err != nil {
        return nil, err
    }

    return l.LoadFromPackage(cachePath)
}

// LoadFromDirectory loads an IG from a directory of JSON files
func (l *IGLoader) LoadFromDirectory(dir string) (*ImplementationGuide, error) {
    ig := &ImplementationGuide{
        StructureDefs: make(map[string]*StructureDefinition),
        ValueSets:     make(map[string]*ValueSet),
        CodeSystems:   make(map[string]*CodeSystem),
        SearchParams:  make(map[string]*SearchParameter),
    }

    err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil || info.IsDir() || !strings.HasSuffix(path, ".json") {
            return err
        }

        data, err := os.ReadFile(path)
        if err != nil {
            return err
        }

        return l.processResource(ig, data)
    })

    if err != nil {
        return nil, err
    }

    return ig, nil
}

func (l *IGLoader) loadFromTarGz(r io.Reader) (*ImplementationGuide, error) {
    ig := &ImplementationGuide{
        StructureDefs: make(map[string]*StructureDefinition),
        ValueSets:     make(map[string]*ValueSet),
        CodeSystems:   make(map[string]*CodeSystem),
        SearchParams:  make(map[string]*SearchParameter),
    }

    gzr, err := gzip.NewReader(r)
    if err != nil {
        return nil, err
    }
    defer gzr.Close()

    tr := tar.NewReader(gzr)

    for {
        header, err := tr.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, err
        }

        // Only process JSON files in package/ directory
        if !strings.HasPrefix(header.Name, "package/") ||
           !strings.HasSuffix(header.Name, ".json") ||
           header.Typeflag != tar.TypeReg {
            continue
        }

        data, err := io.ReadAll(tr)
        if err != nil {
            return nil, err
        }

        if err := l.processResource(ig, data); err != nil {
            // Log warning but continue
            continue
        }
    }

    return ig, nil
}

func (l *IGLoader) processResource(ig *ImplementationGuide, data []byte) error {
    // Peek at resourceType
    var peek struct {
        ResourceType string `json:"resourceType"`
        URL          string `json:"url"`
        ID           string `json:"id"`
    }
    if err := json.Unmarshal(data, &peek); err != nil {
        return err
    }

    switch peek.ResourceType {
    case "StructureDefinition":
        var sd StructureDefinition
        if err := json.Unmarshal(data, &sd); err != nil {
            return err
        }
        ig.StructureDefs[peek.URL] = &sd

    case "ValueSet":
        var vs ValueSet
        if err := json.Unmarshal(data, &vs); err != nil {
            return err
        }
        ig.ValueSets[peek.URL] = &vs

    case "CodeSystem":
        var cs CodeSystem
        if err := json.Unmarshal(data, &cs); err != nil {
            return err
        }
        ig.CodeSystems[peek.URL] = &cs

    case "ImplementationGuide":
        var igMeta struct {
            ID          string `json:"id"`
            Version     string `json:"version"`
            FHIRVersion []string `json:"fhirVersion"`
        }
        if err := json.Unmarshal(data, &igMeta); err != nil {
            return err
        }
        ig.ID = igMeta.ID
        ig.Version = igMeta.Version
        if len(igMeta.FHIRVersion) > 0 {
            ig.FHIRVersion = igMeta.FHIRVersion[0]
        }
    }

    return nil
}
```

### 13.2 Registry Integration with IGs

```go
// pkg/validator/spec_registry_ig.go

package validator

// LoadImplementationGuide adds an IG's definitions to the registry
func (r *SpecRegistry) LoadImplementationGuide(ig *ImplementationGuide) error {
    for url, sd := range ig.StructureDefs {
        r.structureDefs.Put(url, sd)
        // Also index by resource type for profiles
        if sd.Type != "" {
            r.structureDefs.Put(sd.Type+":"+url, sd)
        }
    }

    for url, vs := range ig.ValueSets {
        r.valueSets.Put(url, vs)
    }

    for url, cs := range ig.CodeSystems {
        r.codeSystems.Put(url, cs)
    }

    return nil
}

// GetProfile retrieves a profile for a resource type
func (r *SpecRegistry) GetProfile(resourceType, profileURL string) (*StructureDefinition, error) {
    key := resourceType + ":" + profileURL
    if sd, ok := r.structureDefs.Get(key); ok {
        return sd, nil
    }

    if sd, ok := r.structureDefs.Get(profileURL); ok {
        return sd, nil
    }

    return nil, fmt.Errorf("profile not found: %s", profileURL)
}
```

---

## Part 14: Terminology Validation

### 14.1 Terminology Client

```go
// pkg/validator/terminology/client.go

package terminology

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type TerminologyClient struct {
    baseURL    string
    client     *http.Client
    cache      *LRUCache[string, *ValidationResult]
    cacheTTL   time.Duration
}

type ValidationResult struct {
    Valid   bool
    Display string
    Message string
    CachedAt time.Time
}

func NewTerminologyClient(baseURL string, cacheSize int, cacheTTL time.Duration) *TerminologyClient {
    return &TerminologyClient{
        baseURL:  baseURL,
        client:   &http.Client{Timeout: 30 * time.Second},
        cache:    NewLRUCache[string, *ValidationResult](cacheSize),
        cacheTTL: cacheTTL,
    }
}

func (c *TerminologyClient) ValidateCode(ctx context.Context, system, code, display string) (*ValidationResult, error) {
    cacheKey := fmt.Sprintf("%s|%s", system, code)

    // Check cache
    if cached, ok := c.cache.Get(cacheKey); ok {
        if time.Since(cached.CachedAt) < c.cacheTTL {
            return cached, nil
        }
    }

    // Build Parameters resource
    params := map[string]interface{}{
        "resourceType": "Parameters",
        "parameter": []map[string]interface{}{
            {"name": "system", "valueUri": system},
            {"name": "code", "valueCode": code},
        },
    }

    if display != "" {
        params["parameter"] = append(params["parameter"].([]map[string]interface{}),
            map[string]interface{}{"name": "display", "valueString": display})
    }

    body, _ := json.Marshal(params)

    req, err := http.NewRequestWithContext(ctx, "POST",
        c.baseURL+"/CodeSystem/$validate-code", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/fhir+json")
    req.Header.Set("Accept", "application/fhir+json")

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Parameter []struct {
            Name         string `json:"name"`
            ValueBoolean bool   `json:"valueBoolean"`
            ValueString  string `json:"valueString"`
        } `json:"parameter"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    validationResult := &ValidationResult{CachedAt: time.Now()}
    for _, param := range result.Parameter {
        switch param.Name {
        case "result":
            validationResult.Valid = param.ValueBoolean
        case "display":
            validationResult.Display = param.ValueString
        case "message":
            validationResult.Message = param.ValueString
        }
    }

    c.cache.Put(cacheKey, validationResult)
    return validationResult, nil
}

func (c *TerminologyClient) ValidateCoding(ctx context.Context, valueSetURL string, coding map[string]interface{}) (*ValidationResult, error) {
    system, _ := coding["system"].(string)
    code, _ := coding["code"].(string)
    display, _ := coding["display"].(string)

    cacheKey := fmt.Sprintf("%s|%s|%s", valueSetURL, system, code)

    // Check cache
    if cached, ok := c.cache.Get(cacheKey); ok {
        if time.Since(cached.CachedAt) < c.cacheTTL {
            return cached, nil
        }
    }

    params := map[string]interface{}{
        "resourceType": "Parameters",
        "parameter": []map[string]interface{}{
            {"name": "url", "valueUri": valueSetURL},
            {"name": "coding", "valueCoding": coding},
        },
    }

    body, _ := json.Marshal(params)

    req, err := http.NewRequestWithContext(ctx, "POST",
        c.baseURL+"/ValueSet/$validate-code", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    req.Header.Set("Content-Type", "application/fhir+json")
    req.Header.Set("Accept", "application/fhir+json")

    resp, err := c.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Parameter []struct {
            Name         string `json:"name"`
            ValueBoolean bool   `json:"valueBoolean"`
            ValueString  string `json:"valueString"`
        } `json:"parameter"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    validationResult := &ValidationResult{CachedAt: time.Now()}
    for _, param := range result.Parameter {
        switch param.Name {
        case "result":
            validationResult.Valid = param.ValueBoolean
        case "display":
            validationResult.Display = param.ValueString
        case "message":
            validationResult.Message = param.ValueString
        }
    }

    c.cache.Put(cacheKey, validationResult)
    return validationResult, nil
}
```

### 14.2 Terminology Validator

```go
// pkg/validator/validators/terminology.go

package validators

import (
    "context"
    "fmt"

    "github.com/yourorg/fhir-toolkit-go/pkg/validator/terminology"
)

type TerminologyValidator struct {
    client      *terminology.TerminologyClient
    specRegistry *SpecRegistry
}

func NewTerminologyValidator(serverURL string, cacheSize int, cacheTTL int) (*TerminologyValidator, error) {
    if serverURL == "" {
        serverURL = "https://tx.fhir.org/r4" // Default
    }

    return &TerminologyValidator{
        client: terminology.NewTerminologyClient(serverURL, cacheSize, time.Duration(cacheTTL)*time.Second),
    }, nil
}

func (v *TerminologyValidator) Validate(ctx context.Context, vctx *ValidationContext) ([]OperationOutcomeIssue, error) {
    var issues []OperationOutcomeIssue

    // Walk through resource looking for coded elements
    bindings := v.extractBindings(vctx.StructureDef)

    for _, binding := range bindings {
        bindingIssues, err := v.validateBinding(ctx, vctx.Resource, binding)
        if err != nil {
            return nil, err
        }
        issues = append(issues, bindingIssues...)
    }

    return issues, nil
}

type ElementBinding struct {
    Path          string
    ValueSetURL   string
    Strength      string // required, extensible, preferred, example
}

func (v *TerminologyValidator) extractBindings(sd *StructureDefinition) []ElementBinding {
    var bindings []ElementBinding

    for _, element := range sd.Snapshot.Element {
        if element.Binding != nil && element.Binding.ValueSet != "" {
            bindings = append(bindings, ElementBinding{
                Path:        element.Path,
                ValueSetURL: element.Binding.ValueSet,
                Strength:    element.Binding.Strength,
            })
        }
    }

    return bindings
}

func (v *TerminologyValidator) validateBinding(ctx context.Context, resource interface{}, binding ElementBinding) ([]OperationOutcomeIssue, error) {
    var issues []OperationOutcomeIssue

    // Get values at the binding path
    values := getValuesAtPath(resource, binding.Path)

    for i, value := range values {
        coding, ok := value.(map[string]interface{})
        if !ok {
            continue
        }

        result, err := v.client.ValidateCoding(ctx, binding.ValueSetURL, coding)
        if err != nil {
            // Network error - add warning
            issues = append(issues, OperationOutcomeIssue{
                Severity:    IssueSeverityWarning,
                Code:        IssueCodeBusinessRule,
                Diagnostics: fmt.Sprintf("Unable to validate code: %v", err),
                Location:    []string{fmt.Sprintf("%s[%d]", binding.Path, i)},
            })
            continue
        }

        if !result.Valid {
            severity := getSeverityForBindingStrength(binding.Strength)
            issues = append(issues, OperationOutcomeIssue{
                Severity:    severity,
                Code:        IssueCodeCodeInvalid,
                Diagnostics: fmt.Sprintf("Code '%s' from system '%s' is not in valueset %s: %s",
                    coding["code"], coding["system"], binding.ValueSetURL, result.Message),
                Location:    []string{fmt.Sprintf("%s[%d]", binding.Path, i)},
            })
        }
    }

    return issues, nil
}

func getSeverityForBindingStrength(strength string) IssueSeverity {
    switch strength {
    case "required":
        return IssueSeverityError
    case "extensible":
        return IssueSeverityWarning
    case "preferred", "example":
        return IssueSeverityInformation
    default:
        return IssueSeverityWarning
    }
}
```

---

## Part 15: Complete Example Application

```go
// examples/complete/main.go

package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"

    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/types/valuesets"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/builders/resources"
    "github.com/yourorg/fhir-toolkit-go/pkg/r4/builders/datatypes"
    "github.com/yourorg/fhir-toolkit-go/pkg/validator"
    "github.com/yourorg/fhir-toolkit-go/pkg/fhirpath"
)

func main() {
    // ============================================
    // Example 1: Building a Patient with Fluent API
    // ============================================

    patient := resources.NewPatientBuilder().
        SetID("example-patient-001").
        SetActive(true).
        AddIdentifier(types.Identifier{
            Use:    strPtr("official"),
            System: strPtr("http://hospital.example.org/mrn"),
            Value:  strPtr("MRN-12345678"),
        }).
        AddIdentifier(types.Identifier{
            Use:    strPtr("secondary"),
            System: strPtr("http://national-id.example.org"),
            Value:  strPtr("123-45-6789"),
        }).
        AddName(types.HumanName{
            Use:    strPtr("official"),
            Family: strPtr("García"),
            Given:  []string{"María", "Isabel"},
            Prefix: []string{"Dra."},
        }).
        AddName(types.HumanName{
            Use:    strPtr("nickname"),
            Given:  []string{"Mari"},
        }).
        SetGender(valuesets.AdministrativeGenderFemale).
        SetBirthDate("1985-03-15").
        AddTelecom(types.ContactPoint{
            System: strPtr("phone"),
            Value:  strPtr("+56 9 1234 5678"),
            Use:    strPtr("mobile"),
        }).
        AddTelecom(types.ContactPoint{
            System: strPtr("email"),
            Value:  strPtr("maria.garcia@email.com"),
            Use:    strPtr("home"),
        }).
        AddAddress(types.Address{
            Use:        strPtr("home"),
            Type:       strPtr("physical"),
            Line:       []string{"Av. Libertador Bernardo O'Higgins 1234", "Depto 567"},
            City:       strPtr("Santiago"),
            State:      strPtr("Región Metropolitana"),
            PostalCode: strPtr("8320000"),
            Country:    strPtr("Chile"),
        }).
        AddContact(types.PatientContact{
            Relationship: []types.CodeableConcept{{
                Coding: []types.Coding{{
                    System:  strPtr("http://terminology.hl7.org/CodeSystem/v2-0131"),
                    Code:    strPtr("C"),
                    Display: strPtr("Emergency Contact"),
                }},
            }},
            Name: &types.HumanName{
                Family: strPtr("García"),
                Given:  []string{"Pedro"},
            },
            Telecom: []types.ContactPoint{{
                System: strPtr("phone"),
                Value:  strPtr("+56 9 8765 4321"),
            }},
        }).
        AddExtension(types.Extension{
            URL:         "http://hl7.org/fhir/StructureDefinition/patient-birthPlace",
            ValueAddress: &types.Address{
                City:    strPtr("Valparaíso"),
                Country: strPtr("Chile"),
            },
        }).
        Build()

    // Serialize to JSON
    patientJSON, err := patient.ToJSON()
    if err != nil {
        log.Fatalf("Failed to serialize patient: %v", err)
    }

    fmt.Println("=== Patient Resource ===")
    printPrettyJSON(patientJSON)

    // ============================================
    // Example 2: Building an Observation
    // ============================================

    observation := resources.NewObservationBuilder().
        SetID("blood-pressure-001").
        SetStatus(valuesets.ObservationStatusFinal).
        AddCategory(datatypes.NewCodeableConceptBuilder().
            AddCoding(datatypes.NewCodingBuilder().
                SetSystem("http://terminology.hl7.org/CodeSystem/observation-category").
                SetCode("vital-signs").
                SetDisplay("Vital Signs").
                Build()).
            Build()).
        SetCode(datatypes.NewCodeableConceptBuilder().
            AddCoding(datatypes.NewCodingBuilder().
                SetSystem("http://loinc.org").
                SetCode("85354-9").
                SetDisplay("Blood pressure panel").
                Build()).
            SetText("Blood Pressure").
            Build()).
        SetSubject(types.Reference{
            Reference: strPtr("Patient/example-patient-001"),
            Display:   strPtr("María García"),
        }).
        SetEffectiveDateTime("2024-01-15T10:30:00-03:00").
        AddComponent(types.ObservationComponent{
            Code: types.CodeableConcept{
                Coding: []types.Coding{{
                    System:  strPtr("http://loinc.org"),
                    Code:    strPtr("8480-6"),
                    Display: strPtr("Systolic blood pressure"),
                }},
            },
            ValueQuantity: &types.Quantity{
                Value:  floatPtr(120),
                Unit:   strPtr("mmHg"),
                System: strPtr("http://unitsofmeasure.org"),
                Code:   strPtr("mm[Hg]"),
            },
        }).
        AddComponent(types.ObservationComponent{
            Code: types.CodeableConcept{
                Coding: []types.Coding{{
                    System:  strPtr("http://loinc.org"),
                    Code:    strPtr("8462-4"),
                    Display: strPtr("Diastolic blood pressure"),
                }},
            },
            ValueQuantity: &types.Quantity{
                Value:  floatPtr(80),
                Unit:   strPtr("mmHg"),
                System: strPtr("http://unitsofmeasure.org"),
                Code:   strPtr("mm[Hg]"),
            },
        }).
        Build()

    obsJSON, _ := observation.ToJSON()
    fmt.Println("\n=== Observation Resource ===")
    printPrettyJSON(obsJSON)

    // ============================================
    // Example 3: FHIRPath Queries
    // ============================================

    fmt.Println("\n=== FHIRPath Queries ===")

    var patientMap map[string]interface{}
    json.Unmarshal(patientJSON, &patientMap)

    // Query 1: Get patient's official name
    result, _ := fhirpath.Evaluate("name.where(use = 'official').given.first()", patientMap)
    fmt.Printf("Official first name: %v\n", result)

    // Query 2: Get mobile phone
    result, _ = fhirpath.Evaluate("telecom.where(system = 'phone' and use = 'mobile').value", patientMap)
    fmt.Printf("Mobile phone: %v\n", result)

    // Query 3: Check if patient is active
    active, _ := fhirpath.EvaluateToBoolean("active = true", patientMap)
    fmt.Printf("Is active: %v\n", active)

    // Query 4: Count identifiers
    result, _ = fhirpath.Evaluate("identifier.count()", patientMap)
    fmt.Printf("Number of identifiers: %v\n", result)

    // Query 5: Get all coding systems used
    var obsMap map[string]interface{}
    json.Unmarshal(obsJSON, &obsMap)
    result, _ = fhirpath.Evaluate("component.code.coding.system.distinct()", obsMap)
    fmt.Printf("Coding systems in observation: %v\n", result)

    // ============================================
    // Example 4: Validation
    // ============================================

    fmt.Println("\n=== Validation ===")

    opts := validator.DefaultOptions(validator.FHIRVersionR4)
    opts.ValidateConstraints = true

    v, err := validator.NewValidator(opts)
    if err != nil {
        log.Fatalf("Failed to create validator: %v", err)
    }

    // Validate valid patient
    outcome, err := v.Validate(context.Background(), patientMap)
    if err != nil {
        log.Fatalf("Validation failed: %v", err)
    }

    if outcome.IsSuccess() {
        fmt.Println("✓ Patient validation passed")
    } else {
        fmt.Printf("✗ Patient validation failed with %d errors\n", outcome.ErrorCount())
        for _, issue := range outcome.Issues {
            fmt.Printf("  - [%s] %s at %v\n", issue.Severity, issue.Diagnostics, issue.Location)
        }
    }

    // Validate observation
    outcome, _ = v.Validate(context.Background(), obsMap)
    if outcome.IsSuccess() {
        fmt.Println("✓ Observation validation passed")
    }

    // Create an invalid resource to test validation
    invalidPatient := map[string]interface{}{
        "resourceType": "Patient",
        "id":           "invalid id with spaces!", // Invalid!
        "birthDate":    "not-a-date",              // Invalid!
    }

    outcome, _ = v.Validate(context.Background(), invalidPatient)
    fmt.Printf("\n✗ Invalid patient has %d errors:\n", outcome.ErrorCount())
    for _, issue := range outcome.Issues {
        fmt.Printf("  - [%s] %s\n", issue.Severity, issue.Diagnostics)
    }

    // ============================================
    // Example 5: Immutable Updates
    // ============================================

    fmt.Println("\n=== Immutable Updates ===")

    // Original patient
    fmt.Printf("Original patient active: %v\n", *patient.Active())

    // Create modified copy without changing original
    modifiedPatient := patient.With(func(p *resources.Patient) {
        p.SetActive(false)
        p.SetGender(valuesets.AdministrativeGenderUnknown)
    })

    fmt.Printf("Original patient active (unchanged): %v\n", *patient.Active())
    fmt.Printf("Modified patient active: %v\n", *modifiedPatient.Active())

    // ============================================
    // Example 6: Clone
    // ============================================

    fmt.Println("\n=== Deep Clone ===")

    clone := patient.Clone()
    clone.SetID("cloned-patient-002")

    fmt.Printf("Original ID: %s\n", *patient.ID())
    fmt.Printf("Clone ID: %s\n", *clone.ID())
}

// Helper functions
func strPtr(s string) *string       { return &s }
func floatPtr(f float64) *float64   { return &f }

func printPrettyJSON(data []byte) {
    var pretty bytes.Buffer
    json.Indent(&pretty, data, "", "  ")
    fmt.Println(pretty.String())
}
```

---

## Part 16: Makefile for Build Automation

```makefile
# Makefile

.PHONY: all build test generate lint clean install

# Variables
GO := go
GOFLAGS := -v
BINARY := fhir-cli
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

# Directories
PKG_DIR := ./pkg
CMD_DIR := ./cmd
CODEGEN_DIR := ./internal/codegen
SPECS_DIR := ./specs

all: generate build test

# Build
build:
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o bin/$(BINARY) $(CMD_DIR)/fhir-cli

build-all:
	GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o bin/$(BINARY)-linux-amd64 $(CMD_DIR)/fhir-cli
	GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o bin/$(BINARY)-darwin-amd64 $(CMD_DIR)/fhir-cli
	GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o bin/$(BINARY)-darwin-arm64 $(CMD_DIR)/fhir-cli
	GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS) -o bin/$(BINARY)-windows-amd64.exe $(CMD_DIR)/fhir-cli

# Test
test:
	$(GO) test $(GOFLAGS) ./...

test-coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

test-race:
	$(GO) test -race ./...

# Benchmarks
bench:
	$(GO) test -bench=. -benchmem ./...

bench-fhirpath:
	$(GO) test -bench=. -benchmem ./pkg/fhirpath/...

# Code generation
generate: generate-r4 generate-r4b generate-r5

generate-r4:
	$(GO) run $(CMD_DIR)/codegen -version=R4 -specs=$(SPECS_DIR)/r4 -output=$(PKG_DIR)/r4

generate-r4b:
	$(GO) run $(CMD_DIR)/codegen -version=R4B -specs=$(SPECS_DIR)/r4b -output=$(PKG_DIR)/r4b

generate-r5:
	$(GO) run $(CMD_DIR)/codegen -version=R5 -specs=$(SPECS_DIR)/r5 -output=$(PKG_DIR)/r5

# Download FHIR specs
download-specs:
	mkdir -p $(SPECS_DIR)/r4 $(SPECS_DIR)/r4b $(SPECS_DIR)/r5
	curl -L https://www.hl7.org/fhir/R4/definitions.json.zip -o /tmp/r4-defs.zip
	unzip -o /tmp/r4-defs.zip -d $(SPECS_DIR)/r4
	curl -L https://www.hl7.org/fhir/R4B/definitions.json.zip -o /tmp/r4b-defs.zip
	unzip -o /tmp/r4b-defs.zip -d $(SPECS_DIR)/r4b
	curl -L https://www.hl7.org/fhir/R5/definitions.json.zip -o /tmp/r5-defs.zip
	unzip -o /tmp/r5-defs.zip -d $(SPECS_DIR)/r5

# Lint
lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

# Format
fmt:
	$(GO) fmt ./...
	goimports -w .

# Clean
clean:
	rm -rf bin/
	rm -rf coverage.out coverage.html
	$(GO) clean -cache

# Install dependencies
install:
	$(GO) mod download
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install golang.org/x/tools/cmd/goimports@latest

# Documentation
docs:
	godoc -http=:6060

# Release
release: test lint build-all
	@echo "Release $(VERSION) built successfully"
```

---

## References

- FHIR Specification: https://hl7.org/fhir/
- FHIRPath Specification: https://hl7.org/fhirpath/
- Go JSON handling: https://pkg.go.dev/encoding/json
- Go templates: https://pkg.go.dev/text/template
- Go embed: https://pkg.go.dev/embed
- Go generics: https://go.dev/doc/tutorial/generics
- Effective Go: https://go.dev/doc/effective_go

---

## Appendix A: FHIR Primitive Type Regex Patterns

```go
// Complete regex patterns for all FHIR primitive types

var FHIRPrimitivePatterns = map[string]string{
    "id":           `^[A-Za-z0-9\-\.]{1,64}$`,
    "string":       `^[\s\S]+$`,
    "uri":          `^\S*$`,
    "url":          `^\S*$`,
    "canonical":    `^\S*$`,
    "oid":          `^urn:oid:[0-2](\.(0|[1-9][0-9]*))+$`,
    "uuid":         `^urn:uuid:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
    "code":         `^[^\s]+(\s[^\s]+)*$`,
    "markdown":     `^[\s\S]+$`,
    "base64Binary": `^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`,
    "integer":      `^[0]|[-+]?[1-9][0-9]*$`,
    "integer64":    `^[0]|[-+]?[1-9][0-9]*$`,
    "unsignedInt":  `^[0-9]+$`,
    "positiveInt":  `^[1-9][0-9]*$`,
    "decimal":      `^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?$`,
    "boolean":      `^true|false$`,
    "date":         `^([0-9]([0-9]([0-9][1-9]|[1-9]0)|[1-9]00)|[1-9]000)(-(0[1-9]|1[0-2])(-(0[1-9]|[1-2][0-9]|3[0-1]))?)?$`,
    "dateTime":     `^([0-9]([0-9]([0-9][1-9]|[1-9]0)|[1-9]00)|[1-9]000)(-(0[1-9]|1[0-2])(-(0[1-9]|[1-2][0-9]|3[0-1])(T([01][0-9]|2[0-3]):[0-5][0-9]:([0-5][0-9]|60)(\.[0-9]+)?(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00)))?)?)?$`,
    "time":         `^([01][0-9]|2[0-3]):[0-5][0-9]:([0-5][0-9]|60)(\.[0-9]+)?$`,
    "instant":      `^([0-9]([0-9]([0-9][1-9]|[1-9]0)|[1-9]00)|[1-9]000)-(0[1-9]|1[0-2])-(0[1-9]|[1-2][0-9]|3[0-1])T([01][0-9]|2[0-3]):[0-5][0-9]:([0-5][0-9]|60)(\.[0-9]+)?(Z|(\+|-)((0[0-9]|1[0-3]):[0-5][0-9]|14:00))$`,
    "xhtml":        `^<div[^>]*>[\s\S]*</div>$`,
}
```

---

## Appendix B: FHIRPath Function Reference

| Function | Description | Example |
|----------|-------------|---------|
| `empty()` | True if collection is empty | `name.empty()` |
| `exists()` | True if collection has elements | `identifier.exists()` |
| `exists(criteria)` | True if any element matches | `name.exists(use = 'official')` |
| `all(criteria)` | True if all elements match | `telecom.all(system = 'phone')` |
| `where(criteria)` | Filter collection | `name.where(use = 'official')` |
| `select(projection)` | Project elements | `name.select(given)` |
| `first()` | First element | `name.first()` |
| `last()` | Last element | `name.last()` |
| `tail()` | All but first | `name.tail()` |
| `take(n)` | First n elements | `name.take(2)` |
| `skip(n)` | Skip first n | `name.skip(1)` |
| `count()` | Number of elements | `identifier.count()` |
| `distinct()` | Remove duplicates | `telecom.system.distinct()` |
| `isDistinct()` | True if all unique | `identifier.value.isDistinct()` |
| `not()` | Boolean negation | `active.not()` |
| `iif(c, t, f)` | Conditional | `iif(active, 'Yes', 'No')` |
| `startsWith(s)` | String starts with | `name.family.startsWith('Gar')` |
| `endsWith(s)` | String ends with | `name.family.endsWith('ez')` |
| `contains(s)` | String contains | `name.family.contains('arc')` |
| `matches(regex)` | Regex match | `id.matches('^[A-Z].*')` |
| `length()` | String length | `name.family.length()` |
| `upper()` | Uppercase | `name.family.upper()` |
| `lower()` | Lowercase | `name.family.lower()` |
| `ofType(type)` | Filter by FHIR type | `entry.resource.ofType(Patient)` |
| `as(type)` | Cast to type | `value.as(Quantity)` |
| `is(type)` | Type check | `value.is(Quantity)` |
| `extension(url)` | Get extension by URL | `extension('http://...')` |
| `resolve()` | Resolve reference | `subject.resolve()` |
| `sum()` | Sum of numbers | `component.valueQuantity.value.sum()` |
| `min()` | Minimum value | `observation.valueQuantity.value.min()` |
| `max()` | Maximum value | `observation.valueQuantity.value.max()` |
| `avg()` | Average value | `observation.valueQuantity.value.avg()` |
