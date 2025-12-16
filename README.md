# GoFHIR

[![CI](https://github.com/robertoaraneda/gofhir/actions/workflows/ci.yml/badge.svg)](https://github.com/robertoaraneda/gofhir/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/robertoaraneda/gofhir.svg)](https://pkg.go.dev/github.com/robertoaraneda/gofhir)
[![codecov](https://codecov.io/gh/robertoaraneda/gofhir/branch/main/graph/badge.svg)](https://codecov.io/gh/robertoaraneda/gofhir)

Production-grade FHIR toolkit for Go.

## Features

- **Strongly-typed resources**: All FHIR R4, R4B, and R5 resources as Go structs
- **Multi-version abstraction**: Common interfaces for version-agnostic code
- **Fluent builders**: Construct resources with a fluent API
- **FHIRPath engine**: Evaluate FHIRPath expressions
- **Validation**: Validate resources against StructureDefinitions
- **Full extension support**: Primitive elements support extensions via `_field` pattern
- **JSON field ordering**: Typed structs guarantee FHIR-compliant field order
- **FHIR Server**: Production-grade server with CRUD, search, and validation

## Installation

```bash
go get github.com/robertoaraneda/gofhir
```

## Quick Start

```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/robertoaraneda/gofhir/pkg/common"
    "github.com/robertoaraneda/gofhir/pkg/fhir/r4"
)

func main() {
    // Create a patient using the builder
    patient := r4.NewPatientBuilder().
        SetID("example-001").
        SetActive(true).
        AddName(r4.HumanName{
            Family: common.String("Garcia"),
            Given:  []string{"Maria"},
        }).
        SetBirthDate("1985-03-15").
        Build()

    // Serialize to JSON
    data, _ := json.MarshalIndent(patient, "", "  ")
    fmt.Println(string(data))
}
```

## Packages

| Package | Description |
|---------|-------------|
| `pkg/fhir` | Multi-version abstraction (ResourceFactory, Resource, Meta interfaces) |
| `pkg/fhir/r4` | FHIR R4 (4.0.1) types, builders, and version adapter |
| `pkg/fhir/r4b` | FHIR R4B (4.3.0) types, builders, and version adapter |
| `pkg/fhir/r5` | FHIR R5 (5.0.0) types, builders, and version adapter |
| `pkg/fhirpath` | FHIRPath expression evaluator |
| `pkg/validator` | Resource validation |
| `pkg/common` | Shared utilities |
| `fhir-server/` | Production-grade FHIR server (see [fhir-server/README.md](fhir-server/README.md)) |

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

# Download FHIR specifications
make download-specs

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
│   │   ├── r4/          # R4 types + adapter
│   │   ├── r4b/         # R4B types + adapter
│   │   └── r5/          # R5 types + adapter
│   ├── fhirpath/        # FHIRPath engine
│   ├── validator/       # Validation
│   └── common/          # Utilities
├── internal/
│   └── codegen/         # Code generation
├── fhir-server/         # Production FHIR server
│   ├── cmd/server/      # Server entry point
│   ├── internal/        # Server implementation
│   └── docs/            # Architecture documentation
├── specs/               # FHIR specifications
└── scripts/             # Build scripts
```

## License

MIT License - see [LICENSE](LICENSE) for details.
