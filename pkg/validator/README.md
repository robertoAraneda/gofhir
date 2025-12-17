# FHIR Validator

Comprehensive FHIR resource validation against StructureDefinitions.

## Overview

This package validates FHIR resources against their StructureDefinitions, including structure validation, cardinality checks, type validation, FHIRPath constraint evaluation, terminology binding verification, and reference resolution.

## Installation

```go
import "github.com/robertoaraneda/gofhir/pkg/validator"
```

## Quick Start

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

    // Validate a resource
    patient := []byte(`{
        "resourceType": "Patient",
        "id": "123",
        "name": [{"family": "Doe", "given": ["John"]}]
    }`)

    result, err := v.Validate(context.Background(), patient)
    if err != nil {
        panic(err)
    }

    if result.Valid {
        fmt.Println("Resource is valid!")
    } else {
        for _, issue := range result.Issues {
            fmt.Printf("[%s] %s: %s\n",
                issue.Severity, issue.Code, issue.Diagnostics)
        }
    }
}
```

## Validator Options

```go
type ValidatorOptions struct {
    // ValidateConstraints enables FHIRPath constraint validation
    ValidateConstraints bool

    // ValidateTerminology enables terminology binding validation
    ValidateTerminology bool

    // TerminologyService specifies which embedded terminology to use
    TerminologyService TerminologyServiceType

    // ValidateReferences enables reference validation
    ValidateReferences bool

    // ValidateExtensions enables extension validation
    ValidateExtensions bool

    // StrictMode treats warnings as errors
    StrictMode bool

    // MaxErrors stops validation after N errors (0 = unlimited)
    MaxErrors int

    // Profile is an optional profile URL to validate against
    Profile string
}
```

### Default Options

```go
opts := validator.DefaultValidatorOptions()
// Returns:
// ValidateConstraints: true
// ValidateTerminology: false
// ValidateReferences:  false
// ValidateExtensions:  true
// StrictMode:          false
// MaxErrors:           0
```

### Full Validation Example

```go
opts := validator.ValidatorOptions{
    ValidateConstraints: true,
    ValidateTerminology: true,
    TerminologyService:  validator.TerminologyEmbeddedR4,
    ValidateReferences:  true,
    ValidateExtensions:  true,
    StrictMode:          false,
    MaxErrors:           100,
}

v := validator.NewValidator(registry, opts).
    WithReferenceResolver(myResolver)
```

## Validation Types

### 1. Structural Validation

Validates element presence and cardinality:

```go
// Missing required element
// Issue: [error] required: Patient.identifier is required (min=1)

// Too many elements
// Issue: [error] structure: Patient.active exceeds max cardinality (max=1, found=2)
```

### 2. Type Validation

Validates element types match StructureDefinition:

```go
// Wrong type
// Issue: [error] value: Patient.birthDate expected date, got "not-a-date"
```

### 3. Primitive Type Validation

Validates primitive type formats:

| Type | Pattern | Example |
|------|---------|---------|
| `date` | `YYYY[-MM[-DD]]` | `2024-01-15` |
| `dateTime` | ISO 8601 with optional time | `2024-01-15T10:30:00Z` |
| `instant` | Full datetime with timezone | `2024-01-15T10:30:00.000Z` |
| `time` | `HH:MM:SS[.sss]` | `14:30:00.000` |

```go
// Invalid date format
// Issue: [error] value: Patient.birthDate is not a valid date: "01-15-2024"
```

### 4. FHIRPath Constraint Validation

Evaluates FHIRPath invariants from StructureDefinitions:

```go
// Example: Patient constraint pat-1
// "SHALL at least contain a contact's details or a reference to an organization"
// Expression: name.exists() or telecom.exists()

// Issue: [error] invariant: pat-1: Patient must have name or telecom
```

Common constraints validated:

| Resource | Key | Description |
|----------|-----|-------------|
| Patient | pat-1 | Must have name or telecom |
| Bundle | bdl-1 | Total only when searchset/history |
| Bundle | bdl-3 | Entry.request only for batch/transaction |
| Bundle | bdl-4 | Entry.response only for batch-response |
| Observation | obs-6 | dataAbsentReason only when no value |
| Observation | obs-7 | If Observation.code is same as component, no value |

### 5. Terminology Binding Validation

Validates codes against ValueSets based on binding strength:

| Strength | Behavior |
|----------|----------|
| `required` | Code MUST be from ValueSet |
| `extensible` | SHOULD use ValueSet, can extend |
| `preferred` | Recommended to use ValueSet |
| `example` | Informational only |

```go
opts := validator.ValidatorOptions{
    ValidateTerminology: true,
    TerminologyService:  validator.TerminologyEmbeddedR4,
}

// Invalid code for required binding
// Issue: [error] code-invalid: Patient.gender code "X" not in required ValueSet
```

### 6. Reference Validation

Validates FHIR references can be resolved:

```go
type ReferenceResolver interface {
    Resolve(ctx context.Context, reference string) (interface{}, error)
}

// Implement for your FHIR server
type MyResolver struct {
    store FHIRStore
}

func (r *MyResolver) Resolve(ctx context.Context, ref string) (interface{}, error) {
    // Parse reference: "Patient/123" → type="Patient", id="123"
    return r.store.Read(ctx, resourceType, id)
}

v := validator.NewValidator(registry, opts).
    WithReferenceResolver(&MyResolver{store})

// Unresolved reference
// Issue: [warning] not-found: Reference Patient/unknown could not be resolved
```

Reference formats supported:
- Relative: `Patient/123`
- Absolute: `http://example.com/fhir/Patient/123`
- Contained: `#patient-1`
- URN: `urn:uuid:550e8400-e29b-41d4-a716-446655440000`

### 7. Extension Validation

Validates extensions against their StructureDefinitions:

```go
// Unknown extension
// Issue: [warning] extension: Unknown extension URL: http://custom.org/unknown

// Invalid extension value type
// Issue: [error] extension: Extension value type mismatch
```

### 8. Bundle Validation

Validates Bundle-specific constraints:

```go
// Bundle type constraints
// bdl-1: total only for searchset/history
// bdl-3: entry.request only for batch/transaction
// bdl-4: entry.response only for batch-response/transaction-response

// Issue: [error] invariant: bdl-3: entry.request SHALL only be present for batch/transaction
```

## Validation Result

```go
type ValidationResult struct {
    Valid  bool              // true if no errors (warnings allowed)
    Issues []ValidationIssue // All validation issues
}

type ValidationIssue struct {
    Severity    string   // fatal | error | warning | information
    Code        string   // Issue type code
    Diagnostics string   // Human-readable message
    Location    []string // JSON path
    Expression  []string // FHIRPath expression
}
```

### Result Methods

```go
result, _ := v.Validate(ctx, resource)

// Check validity
if result.Valid {
    fmt.Println("No errors")
}

// Check for errors
if result.HasErrors() {
    fmt.Printf("Found %d errors\n", result.ErrorCount())
}

// Check for warnings
if result.HasWarnings() {
    fmt.Printf("Found %d warnings\n", result.WarningCount())
}

// Iterate issues
for _, issue := range result.Issues {
    fmt.Printf("[%s] %s at %s: %s\n",
        issue.Severity,
        issue.Code,
        strings.Join(issue.Expression, ", "),
        issue.Diagnostics)
}
```

### Issue Severity

| Severity | Description |
|----------|-------------|
| `fatal` | Cannot continue processing |
| `error` | Invalid content, must be fixed |
| `warning` | Potential issue, should be reviewed |
| `information` | Informational message |

### Issue Codes

| Code | Description |
|------|-------------|
| `structure` | Structural/schema issue |
| `required` | Required element missing |
| `value` | Invalid value |
| `invariant` | Constraint violation |
| `code-invalid` | Invalid terminology code |
| `not-found` | Reference not found |
| `extension` | Extension issue |
| `processing` | Processing error |

## Terminology Services

### Embedded Terminology

Pre-packaged ValueSets and CodeSystems for offline validation:

```go
// R4 (FHIR 4.0.1)
opts.TerminologyService = validator.TerminologyEmbeddedR4

// R4B (FHIR 4.3.0)
opts.TerminologyService = validator.TerminologyEmbeddedR4B

// R5 (FHIR 5.0.0)
opts.TerminologyService = validator.TerminologyEmbeddedR5
```

### Custom Terminology Service

```go
type TerminologyService interface {
    ValidateCode(ctx context.Context, system, code, valueSetURL string) (bool, error)
    ExpandValueSet(ctx context.Context, valueSetURL string) ([]CodeInfo, error)
    LookupCode(ctx context.Context, system, code string) (*CodeInfo, error)
}

// Implement for tx.fhir.org or your terminology server
type RemoteTerminologyService struct {
    baseURL string
}

func (s *RemoteTerminologyService) ValidateCode(ctx context.Context, system, code, vs string) (bool, error) {
    // Call $validate-code operation
}

v := validator.NewValidator(registry, opts).
    WithTerminologyService(&RemoteTerminologyService{baseURL: "https://tx.fhir.org/r4"})
```

## StructureDefinition Providers

### Embedded Provider

Uses bundled StructureDefinitions:

```go
registry := validator.NewEmbeddedRegistry("R4")  // or "R4B", "R5"
```

### Custom Provider

```go
type StructureDefinitionProvider interface {
    Get(ctx context.Context, url string) (*StructureDef, error)
    GetByType(ctx context.Context, resourceType string) (*StructureDef, error)
    List(ctx context.Context) ([]string, error)
}

// Implement for your StructureDefinition source
type FileSystemProvider struct {
    path string
}

func (p *FileSystemProvider) GetByType(ctx context.Context, rt string) (*StructureDef, error) {
    // Load from filesystem
}
```

## Performance

### Expression Caching

FHIRPath expressions are cached for performance:

```go
// Default cache: 1000 expressions
// Automatic eviction when full
```

### Validation Context

Resources are parsed once and reused throughout validation to avoid repeated JSON parsing.

### Parallel Validation

The validator is thread-safe and can validate multiple resources concurrently:

```go
var wg sync.WaitGroup
for _, resource := range resources {
    wg.Add(1)
    go func(r []byte) {
        defer wg.Done()
        result, _ := v.Validate(ctx, r)
        // process result
    }(resource)
}
wg.Wait()
```

## Examples

### Validating a Patient

```go
patient := []byte(`{
    "resourceType": "Patient",
    "id": "example",
    "identifier": [{
        "system": "http://hospital.org/mrn",
        "value": "12345"
    }],
    "active": true,
    "name": [{
        "use": "official",
        "family": "Doe",
        "given": ["John"]
    }],
    "gender": "male",
    "birthDate": "1990-05-15"
}`)

result, err := v.Validate(ctx, patient)
```

### Validating a Bundle

```go
bundle := []byte(`{
    "resourceType": "Bundle",
    "type": "searchset",
    "total": 1,
    "entry": [{
        "fullUrl": "http://example.com/Patient/123",
        "resource": {
            "resourceType": "Patient",
            "id": "123"
        }
    }]
}`)

result, err := v.Validate(ctx, bundle)
```

### Validating Against a Profile

```go
opts := validator.ValidatorOptions{
    Profile: "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient",
}
v := validator.NewValidator(registry, opts)
result, err := v.Validate(ctx, patient)
```

## Data Models

### StructureDef

```go
type StructureDef struct {
    URL            string       // Canonical URL
    Name           string       // Computer-friendly name
    Type           string       // Resource type (e.g., "Patient")
    Kind           string       // primitive-type | complex-type | resource | logical
    Abstract       bool         // Is abstract type
    BaseDefinition string       // Parent StructureDefinition URL
    FHIRVersion    string       // FHIR version
    Snapshot       []ElementDef // Full element definitions
    Differential   []ElementDef // Changed elements (profiles)
}
```

### ElementDef

```go
type ElementDef struct {
    ID          string              // Unique ID
    Path        string              // Element path
    SliceName   string              // For sliced elements
    Min         int                 // Min cardinality
    Max         string              // Max cardinality
    Types       []TypeRef           // Allowed types
    Fixed       interface{}         // Fixed value
    Pattern     interface{}         // Pattern constraint
    Binding     *ElementBinding     // Terminology binding
    Constraints []ElementConstraint // FHIRPath invariants
    MustSupport bool
    IsModifier  bool
    IsSummary   bool
}
```

### ElementConstraint

```go
type ElementConstraint struct {
    Key        string // Constraint ID (e.g., "pat-1")
    Severity   string // error | warning
    Human      string // Description
    Expression string // FHIRPath expression
    Source     string // Source URL
}
```

## Supported FHIR Versions

| Version | StructureDefinitions | Terminology |
|---------|---------------------|-------------|
| R4 (4.0.1) | Full | Full |
| R4B (4.3.0) | Full | Full |
| R5 (5.0.0) | Full | Full |

## License

See repository root for license information.
