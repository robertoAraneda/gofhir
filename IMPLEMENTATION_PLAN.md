# GoFHIR - Plan de Implementacion por Sprints

## Resumen Ejecutivo

Este documento presenta el plan de implementacion para **GoFHIR**, un toolkit FHIR de grado produccion en Go que incluye:

- Tipos fuertemente tipados para todos los recursos FHIR (R4, R4B, R5)
- Patron Builder fluido para construccion de recursos
- Motor FHIRPath completo para evaluacion de expresiones
- Sistema de validacion robusto (YAFV)
- Pipeline de generacion de codigo desde StructureDefinitions

### Principio de Diseno Clave

> **TODO desde CodeGen**: Todos los recursos, datatypes, y backbone elements se generan automaticamente desde los StructureDefinitions oficiales de FHIR. **No hay herencia ni embedding de structs** - cada tipo es una estructura plana y auto-contenida para garantizar serializacion JSON correcta.
>
> **Soporte Completo de Extensiones**: Todos los primitivos soportan extensiones via el patron `_field` de FHIR. Esto es fundamental para uso como base de servidor FHIR.

Solo se crean manualmente:

- Helpers y utilidades (`pkg/common`)
- Lo que NO existe en StructureDefinitions (interfaces, registries)
- Motor FHIRPath
- Sistema de validacion

### Arquitectura de Packages

```
github.com/robertoaraneda/gofhir/
├── pkg/
│   ├── fhir/                    # Package principal - API publica
│   │   ├── r4/                  # FHIR R4 (4.0.1) - 100% generado
│   │   │   ├── datatypes.go    # Todos los datatypes + backbones
│   │   │   ├── resources.go    # Todos los resources
│   │   │   ├── builders.go     # Todos los builders
│   │   │   ├── registry.go     # Factory + UnmarshalResource
│   │   │   ├── codes.go        # ValueSets principales (required)
│   │   │   └── interfaces.go   # Resource, Element (manual)
│   │   ├── r4b/                 # FHIR R4B (4.3.0) - 100% generado
│   │   └── r5/                  # FHIR R5 (5.0.0) - 100% generado
│   ├── fhirpath/                # Motor FHIRPath (manual)
│   ├── validator/               # Sistema de validacion (manual)
│   └── common/                  # Utilidades compartidas (manual)
├── internal/
│   └── codegen/                 # Generacion de codigo (manual)
└── cmd/
    └── gofhir/                  # CLI tool (manual)
```

**Nota sobre estructura simplificada**: En lugar de múltiples subdirectorios (`datatypes/`, `resources/`, etc.), usamos archivos grandes dentro de cada versión. Esto es idiomático en Go y simplifica imports:

```go
// Import simple
import "github.com/robertoaraneda/gofhir/pkg/fhir/r4"

// Uso directo
patient := &r4.Patient{}
coding := &r4.Coding{}
```

---

## Sprint 0: Fundacion del Proyecto (1 semana)

### Objetivos
- Establecer estructura base del proyecto
- Configurar tooling y CI/CD
- Descargar especificaciones FHIR

### Tareas

#### 0.1 Inicializacion del Proyecto
- [ ] Crear `go.mod` con `github.com/robertoaraneda/gofhir`
- [ ] Configurar estructura de directorios segun arquitectura
- [ ] Crear `Makefile` con targets: `build`, `test`, `generate`, `lint`
- [ ] Configurar `.golangci.yml` para linting

#### 0.2 Descarga de Especificaciones FHIR
- [ ] Crear script para descargar specs de hl7.org/fhir
- [ ] Descargar R4 (4.0.1): StructureDefinitions, ValueSets, CodeSystems
- [ ] Descargar R4B (4.3.0)
- [ ] Descargar R5 (5.0.0)
- [ ] Almacenar en `specs/r4/`, `specs/r4b/`, `specs/r5/`

#### 0.3 Configuracion CI/CD
- [ ] GitHub Actions workflow para tests
- [ ] GitHub Actions workflow para linting
- [ ] Configurar codecov para cobertura

### Entregables
- Repositorio inicializado con estructura base
- Specs FHIR descargadas
- CI/CD funcionando

---

## Patron Critico: Primitive Elements con Extensiones

Antes de comenzar Sprint 1, es fundamental entender el patron de extensiones en primitivos FHIR. Este patron es **obligatorio** para un servidor FHIR completo.

### El Problema

En FHIR, cualquier primitivo puede tener extensiones. El JSON usa el prefijo `_` para el elemento que contiene las extensiones:

```json
{
  "resourceType": "Patient",
  "birthDate": "1990-01-01",
  "_birthDate": {
    "id": "bd-1",
    "extension": [
      {
        "url": "http://example.org/birth-time",
        "valueTime": "14:30:00"
      }
    ]
  }
}
```

### Solucion: Campos Paralelos

Para cada campo primitivo, generamos un campo `_Field` correspondiente:

```go
// pkg/fhir/r4/datatypes.go

// Element es la base para extensiones en primitivos
// NO es herencia - es un tipo separado usado en campos _field
type Element struct {
    ID        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
}

// Extension representa una extension FHIR
type Extension struct {
    ID        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"` // nested extensions
    URL       string      `json:"url"`
    // Choice type para value[x] - solo uno debe estar presente
    ValueString          *string          `json:"valueString,omitempty"`
    ValueInteger         *int             `json:"valueInteger,omitempty"`
    ValueBoolean         *bool            `json:"valueBoolean,omitempty"`
    ValueDecimal         *float64         `json:"valueDecimal,omitempty"`
    ValueCode            *string          `json:"valueCode,omitempty"`
    ValueUri             *string          `json:"valueUri,omitempty"`
    ValueDateTime        *string          `json:"valueDateTime,omitempty"`
    ValueDate            *string          `json:"valueDate,omitempty"`
    ValueTime            *string          `json:"valueTime,omitempty"`
    ValueInstant         *string          `json:"valueInstant,omitempty"`
    ValueCoding          *Coding          `json:"valueCoding,omitempty"`
    ValueCodeableConcept *CodeableConcept `json:"valueCodeableConcept,omitempty"`
    ValueQuantity        *Quantity        `json:"valueQuantity,omitempty"`
    ValueReference       *Reference       `json:"valueReference,omitempty"`
    ValuePeriod          *Period          `json:"valuePeriod,omitempty"`
    ValueIdentifier      *Identifier      `json:"valueIdentifier,omitempty"`
    ValueHumanName       *HumanName       `json:"valueHumanName,omitempty"`
    ValueAddress         *Address         `json:"valueAddress,omitempty"`
    ValueContactPoint    *ContactPoint    `json:"valueContactPoint,omitempty"`
    ValueAttachment      *Attachment      `json:"valueAttachment,omitempty"`
    // ... todos los tipos de value[x]
}
```

### Ejemplo de Resource Generado

```go
// pkg/fhir/r4/resources.go (GENERADO)

type Patient struct {
    ResourceType string `json:"resourceType"`

    // Primitivo + su elemento de extension
    ID    *string  `json:"id,omitempty"`
    IDExt *Element `json:"_id,omitempty"`

    // Primitivo + su elemento de extension
    Active    *bool    `json:"active,omitempty"`
    ActiveExt *Element `json:"_active,omitempty"`

    // Primitivo + su elemento de extension
    BirthDate    *string  `json:"birthDate,omitempty"`
    BirthDateExt *Element `json:"_birthDate,omitempty"`

    // Array de primitivos: extension es array paralelo
    Given    []string   `json:"given,omitempty"`
    GivenExt []*Element `json:"_given,omitempty"` // mismo indice, nil si no hay ext

    // Tipos complejos NO necesitan _field (ya tienen Extension internamente)
    Name []HumanName `json:"name,omitempty"`

    // Meta, contained, etc
    Meta      *Meta             `json:"meta,omitempty"`
    Contained []json.RawMessage `json:"contained,omitempty"` // lazy deserialize
    // ...
}
```

### Reglas de Generacion

1. **Primitivos simples**: Generar campo + campo `_Ext`
2. **Arrays de primitivos**: Generar array + array paralelo de `*Element`
3. **Tipos complejos**: NO generar `_field` (ya tienen Extension internamente)
4. **Contained**: Usar `json.RawMessage` para deserializacion lazy

### JSON Custom Marshaling (Opcional pero Recomendado)

Para arrays de primitivos, necesitamos manejar el caso donde `_given[1]` es null:

```json
{
  "given": ["Juan", "Carlos", "Maria"],
  "_given": [null, {"extension": [...]}, null]
}
```

El codegen puede generar custom marshaler si es necesario, o manejarlo en la validacion.

---

## Sprint 1: Helpers y Generador de Codigo Base (2 semanas)

### Objetivos
- Crear utilidades manuales en `pkg/common`
- Implementar el parser de StructureDefinitions
- Crear la base del generador de codigo
- **Implementar soporte para patron _field de extensiones**

### Tareas

#### 1.1 Package `pkg/common` (Manual)
```go
// pkg/common/errors.go
package common

type FHIRError struct {
    Code    string
    Message string
    Path    string
}

func (e *FHIRError) Error() string

// pkg/common/ptr.go - Helpers para punteros
func String(s string) *string { return &s }
func Bool(b bool) *bool { return &b }
func Int(i int) *int { return &i }
func Int64(i int64) *int64 { return &i }
func Float64(f float64) *float64 { return &f }

// Helpers inversos
func StringVal(s *string) string
func BoolVal(b *bool) bool
func IntVal(i *int) int

// pkg/common/clone.go - Clone generico via JSON
func Clone[T any](v *T) *T {
    if v == nil {
        return nil
    }
    data, _ := json.Marshal(v)
    var clone T
    _ = json.Unmarshal(data, &clone)
    return &clone
}

// pkg/common/errors.go - Error wrapping con path
type PathError struct {
    Path string
    Err  error
}

func (e *PathError) Error() string {
    return fmt.Sprintf("at %s: %v", e.Path, e.Err)
}

func (e *PathError) Unwrap() error {
    return e.Err
}

func WrapPath(path string, err error) error {
    if err == nil {
        return nil
    }
    return &PathError{Path: path, Err: err}
}
```

- [ ] Implementar `errors.go` - tipos de error FHIR con wrapping
- [ ] Implementar `ptr.go` - helpers para punteros
- [ ] Implementar `json.go` - utilidades JSON (ordenamiento, omitempty)
- [ ] Implementar `clone.go` - funcion generica Clone[T]

#### 1.2 Parser de StructureDefinitions
```go
// internal/codegen/parser/structuredefinition.go
package parser

// Estructuras para parsear los JSON de FHIR specs
type StructureDefinition struct {
    ResourceType   string       `json:"resourceType"`
    ID             string       `json:"id"`
    URL            string       `json:"url"`
    Name           string       `json:"name"`
    Kind           string       `json:"kind"` // primitive-type, complex-type, resource
    Abstract       bool         `json:"abstract"`
    Type           string       `json:"type"`
    BaseDefinition string       `json:"baseDefinition"`
    Snapshot       *Snapshot    `json:"snapshot"`
    Differential   *Differential `json:"differential"`
}

type Snapshot struct {
    Element []ElementDefinition `json:"element"`
}

type ElementDefinition struct {
    ID           string        `json:"id"`
    Path         string        `json:"path"`
    Short        string        `json:"short"`
    Definition   string        `json:"definition"`
    Min          int           `json:"min"`
    Max          string        `json:"max"`
    Type         []TypeRef     `json:"type"`
    Binding      *Binding      `json:"binding"`
    Constraint   []Constraint  `json:"constraint"`
    IsSummary    bool          `json:"isSummary"`
    IsModifier   bool          `json:"isModifier"`
}

type TypeRef struct {
    Code        string   `json:"code"`
    TargetProfile []string `json:"targetProfile"`
}

type Binding struct {
    Strength   string `json:"strength"`
    ValueSet   string `json:"valueSet"`
}

type Constraint struct {
    Key        string `json:"key"`
    Severity   string `json:"severity"`
    Human      string `json:"human"`
    Expression string `json:"expression"`
}

func ParseStructureDefinition(data []byte) (*StructureDefinition, error)
func LoadAllStructureDefinitions(specsDir string) ([]*StructureDefinition, error)
```

- [ ] Implementar structs para StructureDefinition completa
- [ ] Implementar parser JSON
- [ ] Implementar carga de todos los specs de un directorio
- [ ] Implementar filtrado por Kind (primitive, complex-type, resource)

#### 1.3 Analizador de Elementos
```go
// internal/codegen/analyzer/analyzer.go
package analyzer

type AnalyzedType struct {
    Name           string
    Kind           string // primitive, datatype, resource, backbone
    Properties     []AnalyzedProperty
    Description    string
    IsAbstract     bool
    Constraints    []Constraint
}

type AnalyzedProperty struct {
    Name         string   // Nombre Go (PascalCase)
    JsonName     string   // Nombre JSON (camelCase)
    GoType       string   // Tipo Go completo
    IsPointer    bool     // *Type vs Type
    IsArray      bool     // []Type
    IsChoice     bool     // value[x]
    ChoiceTypes  []string // Para choice types
    Description  string
    Min          int
    Max          string
    Binding      *Binding
}

type Analyzer struct {
    definitions map[string]*StructureDefinition
}

func NewAnalyzer(definitions []*StructureDefinition) *Analyzer
func (a *Analyzer) Analyze(sd *StructureDefinition) (*AnalyzedType, error)
func (a *Analyzer) ResolveGoType(fhirType string) string
func (a *Analyzer) IsChoiceType(element *ElementDefinition) bool
func (a *Analyzer) GetChoiceTypes(element *ElementDefinition) []string
```

- [ ] Implementar conversion FHIR type -> Go type
- [ ] Implementar deteccion de choice types (value[x])
- [ ] Implementar expansion de choice types a propiedades individuales
- [ ] Implementar calculo de cardinalidad (pointer vs array)
- [ ] Implementar generacion de nombres Go validos

#### 1.4 Mapa de Tipos FHIR -> Go
```go
// internal/codegen/analyzer/typemap.go
package analyzer

// Mapeo de tipos primitivos FHIR a Go
var PrimitiveTypeMap = map[string]string{
    "boolean":      "bool",
    "integer":      "int",
    "integer64":    "int64",
    "string":       "string",
    "decimal":      "float64",
    "uri":          "string",
    "url":          "string",
    "canonical":    "string",
    "base64Binary": "string",
    "instant":      "string",
    "date":         "string",
    "dateTime":     "string",
    "time":         "string",
    "code":         "string",
    "oid":          "string",
    "id":           "string",
    "markdown":     "string",
    "unsignedInt":  "uint32",
    "positiveInt":  "uint32",
    "uuid":         "string",
    "xhtml":        "string",
}

// Tipos que requieren puntero cuando min=0
func RequiresPointer(goType string, isArray bool, min int) bool
```

- [ ] Definir mapeo completo de primitivos
- [ ] Implementar logica de punteros
- [ ] Implementar logica para arrays

### Tests Sprint 1
- [ ] Tests de parsing de StructureDefinitions
- [ ] Tests de analisis de elementos
- [ ] Tests de conversion de tipos
- [ ] Tests de deteccion de choice types

### Entregables
- Package `pkg/common` con helpers
- Parser de StructureDefinitions funcional
- Analizador de tipos funcional
- Base para generacion de codigo

---

## Sprint 2: Templates y Generacion Completa R4 (2 semanas)

### Objetivos
- Crear templates de generacion Go
- Implementar generador completo
- Generar TODOS los tipos R4 desde StructureDefinitions

### Tareas

#### 2.1 Templates de Generacion

```go
// internal/codegen/templates/datatype.go.tmpl

// Code generated by gofhir codegen. DO NOT EDIT.
// Source: {{.SourceURL}}

package datatypes

{{if .Imports}}
import (
{{range .Imports}}    "{{.}}"
{{end}})
{{end}}

// {{.Name}} - {{.Description}}
// FHIR Path: {{.FHIRPath}}
type {{.Name}} struct {
{{range .Properties}}
    // {{.Description}}
    {{.GoName}} {{.GoType}} `json:"{{.JsonName}},omitempty"`
{{end}}
}
```

```go
// internal/codegen/templates/resource.go.tmpl

// Code generated by gofhir codegen. DO NOT EDIT.
// Source: {{.SourceURL}}

package resources

import (
    "github.com/robertoaraneda/gofhir/pkg/fhir/r4/datatypes"
{{range .ExtraImports}}    "{{.}}"
{{end}})

// {{.Name}} - {{.Description}}
type {{.Name}} struct {
    // ResourceType is always "{{.Name}}"
    ResourceType string `json:"resourceType"`
{{range .Properties}}
    // {{.Description}}
    {{.GoName}} {{.GoType}} `json:"{{.JsonName}},omitempty"`
{{end}}
}

// GetResourceType returns the FHIR resource type
func (r *{{.Name}}) GetResourceType() string {
    return "{{.Name}}"
}
```

- [ ] Template para datatypes (sin herencia, struct plano)
- [ ] Template para resources (con ResourceType field)
- [ ] Template para backbone elements (inline o separado)
- [ ] Template para ValueSets como constantes tipadas

#### 2.2 Generador Principal

```go
// internal/codegen/generator/generator.go
package generator

type Generator struct {
    analyzer    *analyzer.Analyzer
    templates   *template.Template
    specsDir    string
    outputDir   string
    version     string
}

type GeneratorConfig struct {
    SpecsDir    string   // ./specs/r4
    OutputDir   string   // ./pkg/fhir/r4
    Version     string   // "R4"
    Include     []string // Filtro opcional
    Exclude     []string // Exclusiones
}

func NewGenerator(config *GeneratorConfig) (*Generator, error)
func (g *Generator) Generate() error
func (g *Generator) GenerateDatatypes() error
func (g *Generator) GenerateResources() error
func (g *Generator) GenerateBackbones() error
func (g *Generator) GenerateValueSets() error
```

- [ ] Implementar carga de templates
- [ ] Implementar orquestacion de generacion
- [ ] Implementar resolucion de dependencias (datatypes antes que resources)
- [ ] Implementar formateo automatico con gofmt/goimports

#### 2.3 Manejo de Casos Especiales

```go
// internal/codegen/generator/special_cases.go

// Backbone elements: elementos anidados que se generan como tipos separados
// Ejemplo: Patient.contact -> PatientContact
func (g *Generator) handleBackboneElements(sd *StructureDefinition) []AnalyzedType

// Choice types: value[x] se expande a multiples campos
// Ejemplo: value[x] -> ValueString, ValueQuantity, ValueCodeableConcept...
func (g *Generator) expandChoiceType(element *ElementDefinition) []AnalyzedProperty

// Contenido recursivo: tipos que se referencian a si mismos
// Ejemplo: QuestionnaireItem contiene []QuestionnaireItem
func (g *Generator) handleRecursiveTypes(sd *StructureDefinition) error

// Referencias: Reference con targetProfile
// Se genera como Reference simple (no generics en Go)
func (g *Generator) handleReferences(element *ElementDefinition) AnalyzedProperty
```

- [ ] Implementar extraccion de backbone elements
- [ ] Implementar expansion de choice types
- [ ] Implementar manejo de tipos recursivos
- [ ] Implementar manejo de contained resources

#### 2.4 Generacion de ValueSets

```go
// internal/codegen/generator/valueset_generator.go

// Genera constantes tipadas para ValueSets required/extensible
// Ejemplo:
// type AdministrativeGender string
// const (
//     AdministrativeGenderMale    AdministrativeGender = "male"
//     AdministrativeGenderFemale  AdministrativeGender = "female"
//     AdministrativeGenderOther   AdministrativeGender = "other"
//     AdministrativeGenderUnknown AdministrativeGender = "unknown"
// )

type ValueSetGenerator struct {
    templates *template.Template
    outputDir string
}

func (g *ValueSetGenerator) Generate(vs *ValueSet) error
func (g *ValueSetGenerator) GenerateCodeSystem(cs *CodeSystem) error
```

- [ ] Parsear ValueSets y CodeSystems
- [ ] Generar tipos string tipados
- [ ] Generar constantes para cada codigo
- [ ] Generar metodo IsValid() opcional

#### 2.5 Ejecutar Generacion R4 Completa

- [ ] Generar todos los datatypes (~50 tipos)
- [ ] Generar todos los resources (~150 tipos)
- [ ] Generar todos los backbone elements (~300 tipos)
- [ ] Generar ValueSets principales (~100 tipos)
- [ ] Verificar que compila sin errores
- [ ] Verificar imports correctos

### Tests Sprint 2

- [ ] Tests de templates individuales
- [ ] Tests de generacion de un datatype completo (Coding, CodeableConcept)
- [ ] Tests de generacion de un resource completo (Patient, Observation)
- [ ] Tests de choice types (Observation.value[x])
- [ ] Test de compilacion de todo el codigo generado
- [ ] Tests de serializacion JSON round-trip

### Entregables

- Templates de generacion completos
- Generador funcional
- `pkg/fhir/r4/datatypes/` - 100% generado
- `pkg/fhir/r4/resources/` - 100% generado
- `pkg/fhir/r4/backbones/` - 100% generado
- `pkg/fhir/r4/valuesets/` - principales generados
- `make generate` funcionando

---

## Sprint 3: Metodos Generados y Registry (2 semanas)

### Objetivos
- Extender templates para generar metodos utiles
- Generar helpers para choice types
- Crear registry de recursos e interfaces

### Tareas

#### 3.1 Interfaces Manuales (pkg/fhir/r4/interfaces.go)

```go
// pkg/fhir/r4/interfaces.go - MANUAL, no generado
package r4

// Resource es la interface minima para polimorfismo de recursos
// Solo incluye GetResourceType() porque es el unico metodo necesario
// para identificar el tipo en runtime. Acceso a campos se hace directo.
type Resource interface {
    GetResourceType() string
}

// NO definimos interface Cloneable - usamos funcion generica common.Clone[T]
// Esto evita generar ~500 metodos Clone() en cada struct
```

**Nota sobre Go idiomatico**: No generamos GetID/SetID/ToJSON/Clone porque:

- `patient.ID` es mas idiomatico que `patient.GetID()`
- `json.Marshal(patient)` es estandar Go
- `common.Clone(patient)` es mas eficiente que generar Clone() en cada tipo
- Getters/setters sin logica son anti-pattern en Go

Tareas:

- [ ] Definir interface Resource (solo GetResourceType)
- [ ] Verificar que common.Clone[T] funciona con todos los tipos

#### 3.2 Extender Templates con Metodos (Solo los Utiles)

```go
// internal/codegen/templates/resource_methods.go.tmpl

// Metodos generados automaticamente para {{.Name}}
// Solo generamos GetResourceType - Clone se maneja con common.Clone[T]

// GetResourceType - necesario para polimorfismo via interface Resource
func (r *{{.Name}}) GetResourceType() string {
    return "{{.Name}}"
}

// NO generamos Clone() - usar common.Clone(resource) en su lugar
// Ejemplo de uso:
//   patient2 := common.Clone(patient)
//   observation2 := common.Clone(observation)

// NO generamos (anti-pattern en Go):
// - GetID/SetID: usar r.ID directamente
// - ToJSON: usar json.Marshal(r)
// - GetMeta/SetMeta: usar r.Meta directamente
// - Clone(): usar common.Clone(r)
```

Tareas:

- [ ] Agregar GetResourceType() a template de resources
- [ ] Regenerar todos los recursos con GetResourceType()

#### 3.3 Helpers de Choice Types (Generados)

```go
// internal/codegen/templates/choice_helpers.go.tmpl

{{range .ChoiceTypes}}
// Get{{.BaseName}} retorna el valor y tipo del choice type {{.BaseName}}[x]
func (r *{{$.Name}}) Get{{.BaseName}}() (interface{}, string) {
    {{range .Options}}
    if r.{{.FieldName}} != nil {
        return r.{{.FieldName}}, "{{.TypeName}}"
    }
    {{end}}
    return nil, ""
}

// Has{{.BaseName}} indica si alguno de los campos {{.BaseName}}[x] tiene valor
func (r *{{$.Name}}) Has{{.BaseName}}() bool {
    _, typeName := r.Get{{.BaseName}}()
    return typeName != ""
}

// Clear{{.BaseName}} limpia todos los campos {{.BaseName}}[x]
func (r *{{$.Name}}) clear{{.BaseName}}() {
    {{range .Options}}
    r.{{.FieldName}} = nil
    {{end}}
}

{{range .Options}}
// Set{{.FieldName}} establece el valor como {{.TypeName}}
func (r *{{$.Name}}) Set{{.FieldName}}(v {{.GoType}}) {
    r.clear{{$.BaseName}}()
    r.{{.FieldName}} = {{if .IsPointer}}&{{end}}v
}
{{end}}
{{end}}
```

- [ ] Detectar choice types en analyzer
- [ ] Generar GetValue/GetEffective/GetDeceased etc
- [ ] Generar HasValue/HasEffective etc
- [ ] Generar SetValueQuantity/SetValueString etc
- [ ] Generar clearValue helper privado

#### 3.4 Registry de Recursos (Generado)

```go
// pkg/fhir/r4/registry.go - GENERADO

package r4

import (
    "encoding/json"
    "fmt"
    "github.com/robertoaraneda/gofhir/pkg/fhir/r4/resources"
)

// resourceFactories mapea resourceType a funcion factory
var resourceFactories = map[string]func() Resource{
{{range .Resources}}
    "{{.Name}}": func() Resource { return &resources.{{.Name}}{ResourceType: "{{.Name}}"} },
{{end}}
}

// NewResource crea una instancia vacia del recurso especificado
func NewResource(resourceType string) (Resource, error) {
    factory, ok := resourceFactories[resourceType]
    if !ok {
        return nil, fmt.Errorf("unknown resource type: %s", resourceType)
    }
    return factory(), nil
}

// UnmarshalResource deserializa JSON a el tipo correcto de recurso
func UnmarshalResource(data []byte) (Resource, error) {
    // Primero extraer resourceType
    var peek struct {
        ResourceType string `json:"resourceType"`
    }
    if err := json.Unmarshal(data, &peek); err != nil {
        return nil, err
    }

    resource, err := NewResource(peek.ResourceType)
    if err != nil {
        return nil, err
    }

    if err := json.Unmarshal(data, resource); err != nil {
        return nil, err
    }

    return resource, nil
}

// GetResourceType extrae el resourceType de JSON sin deserializar completo
func GetResourceType(data []byte) (string, error) {
    var peek struct {
        ResourceType string `json:"resourceType"`
    }
    if err := json.Unmarshal(data, &peek); err != nil {
        return "", err
    }
    return peek.ResourceType, nil
}
```

- [ ] Crear template para registry
- [ ] Generar map de factories
- [ ] Implementar NewResource
- [ ] Implementar UnmarshalResource dinamico
- [ ] Implementar GetResourceType helper

#### 3.5 Regenerar Todo con Metodos

- [ ] Actualizar generador para incluir metodos
- [ ] Regenerar todos los datatypes
- [ ] Regenerar todos los resources
- [ ] Regenerar registry
- [ ] Verificar compilacion

### Tests Sprint 3

- [ ] Tests de interface Resource (GetResourceType)
- [ ] Tests de common.Clone[T] con resources
- [ ] Tests de common.Clone[T] con datatypes
- [ ] Tests de choice type helpers (GetValue, HasValue, SetValue*)
- [ ] Tests de registry (NewResource, UnmarshalResource)
- [ ] Tests de round-trip JSON con json.Marshal/Unmarshal
- [ ] Tests de Extension en primitivos (_field)

### Entregables

- Interface Resource minima (solo GetResourceType)
- Funcion generica common.Clone[T] (NO metodos en cada struct)
- Helpers de choice types generados
- Registry de recursos generado
- Todos los tipos regenerados con soporte _field

---

## Sprint 4: Builders Generados (2 semanas)

### Objetivos
- Generar Builders automaticamente para todos los tipos
- Crear API fluida desde templates
- Agregar helpers clinicos manuales

### Tareas

#### 4.1 Template de Builders para Resources

```go
// internal/codegen/templates/resource_builder.go.tmpl

// Code generated by gofhir codegen. DO NOT EDIT.

package builders

import (
    "github.com/robertoaraneda/gofhir/pkg/fhir/r4/resources"
    "github.com/robertoaraneda/gofhir/pkg/fhir/r4/datatypes"
)

// {{.Name}}Builder construye recursos {{.Name}} de forma fluida
type {{.Name}}Builder struct {
    resource *resources.{{.Name}}
}

// New{{.Name}}Builder crea un nuevo builder para {{.Name}}
func New{{.Name}}Builder() *{{.Name}}Builder {
    return &{{.Name}}Builder{
        resource: &resources.{{.Name}}{
            ResourceType: "{{.Name}}",
        },
    }
}

// Build retorna el recurso construido (copia)
func (b *{{.Name}}Builder) Build() *resources.{{.Name}} {
    return b.resource.Clone()
}

// BuildDirect retorna el recurso sin clonar (mas eficiente, pero mutable)
func (b *{{.Name}}Builder) BuildDirect() *resources.{{.Name}} {
    return b.resource
}

{{range .Properties}}
{{if .IsArray}}
// Add{{.GoName}} agrega un elemento a {{.JsonName}}
func (b *{{$.Name}}Builder) Add{{.GoName}}(value {{.ElementType}}) *{{$.Name}}Builder {
    b.resource.{{.GoName}} = append(b.resource.{{.GoName}}, value)
    return b
}

// Set{{.GoName}} reemplaza todo el array {{.JsonName}}
func (b *{{$.Name}}Builder) Set{{.GoName}}(values []{{.ElementType}}) *{{$.Name}}Builder {
    b.resource.{{.GoName}} = values
    return b
}
{{else if .IsPointer}}
// Set{{.GoName}} establece {{.JsonName}}
func (b *{{$.Name}}Builder) Set{{.GoName}}(value {{.BaseType}}) *{{$.Name}}Builder {
    b.resource.{{.GoName}} = &value
    return b
}

// Set{{.GoName}}Ptr establece {{.JsonName}} desde puntero
func (b *{{$.Name}}Builder) Set{{.GoName}}Ptr(value *{{.BaseType}}) *{{$.Name}}Builder {
    b.resource.{{.GoName}} = value
    return b
}
{{else}}
// Set{{.GoName}} establece {{.JsonName}}
func (b *{{$.Name}}Builder) Set{{.GoName}}(value {{.GoType}}) *{{$.Name}}Builder {
    b.resource.{{.GoName}} = value
    return b
}
{{end}}
{{end}}

{{range .ChoiceTypes}}
{{range .Options}}
// Set{{.FieldName}} establece {{$.BaseName}}[x] como {{.TypeName}}
func (b *{{$.ResourceName}}Builder) Set{{.FieldName}}(value {{.GoType}}) *{{$.ResourceName}}Builder {
    b.resource.clear{{$.BaseName}}()
    b.resource.{{.FieldName}} = {{if .NeedsPointer}}&{{end}}value
    return b
}
{{end}}
{{end}}
```

- [ ] Crear template para resource builders
- [ ] Generar Set* para campos singulares con puntero
- [ ] Generar Add* para campos array
- [ ] Generar Set* para choice types (SetValueQuantity, SetValueString, etc)

#### 4.2 Template de Builders para Datatypes

```go
// internal/codegen/templates/datatype_builder.go.tmpl

// {{.Name}}Builder construye {{.Name}} de forma fluida
type {{.Name}}Builder struct {
    data *datatypes.{{.Name}}
}

func New{{.Name}}Builder() *{{.Name}}Builder {
    return &{{.Name}}Builder{
        data: &datatypes.{{.Name}}{},
    }
}

func (b *{{.Name}}Builder) Build() datatypes.{{.Name}} {
    return *b.data
}

func (b *{{.Name}}Builder) BuildPtr() *datatypes.{{.Name}} {
    return b.data
}

{{range .Properties}}
// ... similar a resources
{{end}}
```

- [ ] Crear template para datatype builders
- [ ] Generar builders para todos los datatypes complejos

#### 4.3 Generar Todos los Builders

- [ ] Generar builders para ~150 resources
- [ ] Generar builders para ~50 datatypes
- [ ] Generar builders para backbone elements importantes
- [ ] Verificar compilacion
- [ ] Verificar que Build() retorna tipos correctos

#### 4.4 Helpers Clinicos Manuales (pkg/fhir/r4/helpers/)

```go
// pkg/fhir/r4/helpers/loinc.go - MANUAL
package helpers

import (
    "github.com/robertoaraneda/gofhir/pkg/fhir/r4/datatypes"
    "github.com/robertoaraneda/gofhir/pkg/common"
)

// Codigos LOINC comunes para signos vitales
var (
    LOINCBodyWeight = datatypes.CodeableConcept{
        Coding: []datatypes.Coding{{
            System:  common.String("http://loinc.org"),
            Code:    common.String("29463-7"),
            Display: common.String("Body weight"),
        }},
    }
    LOINCBodyHeight = datatypes.CodeableConcept{...}
    LOINCBodyTemperature = datatypes.CodeableConcept{...}
    LOINCBloodPressurePanel = datatypes.CodeableConcept{...}
    LOINCSystolicBP = datatypes.CodeableConcept{...}
    LOINCDiastolicBP = datatypes.CodeableConcept{...}
    LOINCHeartRate = datatypes.CodeableConcept{...}
    LOINCRespiratoryRate = datatypes.CodeableConcept{...}
    LOINCOxygenSaturation = datatypes.CodeableConcept{...}
    LOINCBMI = datatypes.CodeableConcept{...}
    LOINCGlucose = datatypes.CodeableConcept{...}
    LOINCHemoglobinA1c = datatypes.CodeableConcept{...}
)
```

```go
// pkg/fhir/r4/helpers/quantity.go - MANUAL
package helpers

// Helpers para crear Quantities con UCUM
func QuantityKg(value float64) *datatypes.Quantity {
    return &datatypes.Quantity{
        Value:  &value,
        Unit:   common.String("kg"),
        System: common.String("http://unitsofmeasure.org"),
        Code:   common.String("kg"),
    }
}

func QuantityCm(value float64) *datatypes.Quantity
func QuantityMmHg(value float64) *datatypes.Quantity
func QuantityCelsius(value float64) *datatypes.Quantity
func QuantityPercent(value float64) *datatypes.Quantity
func QuantityBPM(value float64) *datatypes.Quantity
func QuantityMgDL(value float64) *datatypes.Quantity
```

```go
// pkg/fhir/r4/helpers/categories.go - MANUAL
package helpers

// Categorias comunes de Observation
var (
    CategoryVitalSigns = datatypes.CodeableConcept{...}
    CategoryLaboratory = datatypes.CodeableConcept{...}
    CategorySocialHistory = datatypes.CodeableConcept{...}
    CategoryImaging = datatypes.CodeableConcept{...}
)
```

- [ ] Crear helpers LOINC para signos vitales (~20 codigos)
- [ ] Crear helpers UCUM para unidades (~15 funciones)
- [ ] Crear helpers para categorias de Observation
- [ ] Crear helpers para identifier types comunes

### Tests Sprint 4

- [ ] Tests de builders de resources (Patient, Observation, Bundle)
- [ ] Tests de builders de datatypes (Coding, CodeableConcept, Quantity)
- [ ] Tests de choice types en builders
- [ ] Tests de helpers clinicos
- [ ] Tests de integracion (builder + ToJSON + UnmarshalResource)

### Entregables

- Template de builders para resources
- Template de builders para datatypes
- `pkg/fhir/r4/builders/` - 100% generado
- `pkg/fhir/r4/helpers/` - manual con helpers clinicos
- Ejemplo de uso documentado

---

## Sprint 5: Motor FHIRPath (3 semanas)

### Objetivos
- Implementar lexer y parser FHIRPath
- Implementar evaluador de expresiones
- Implementar todas las funciones built-in

### Tareas

#### 5.1 Lexer FHIRPath
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
    TokenDot
    TokenComma
    TokenLParen
    TokenRParen
    TokenLBracket
    TokenRBracket
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
    TokenUnion
    // ...
)

type Token struct {
    Type   TokenType
    Value  string
    Line   int
    Column int
}

type Lexer struct {
    input  string
    pos    int
    line   int
    column int
}

func NewLexer(input string) *Lexer
func (l *Lexer) NextToken() Token
```

- [ ] Implementar tokenizacion basica
- [ ] Implementar strings con escape
- [ ] Implementar numeros (int y decimal)
- [ ] Implementar literales datetime (@)
- [ ] Implementar keywords (and, or, implies, etc)
- [ ] Implementar operadores

#### 5.2 AST FHIRPath
```go
// pkg/fhirpath/parser/ast.go
package parser

type Node interface {
    node()
}

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
    Target     Expression
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

type LiteralExpr struct {
    Value interface{}
    Type  string
}

type ParenExpr struct {
    Inner Expression
}

type ThisExpr struct{}
type IndexExpr struct{}
type TotalExpr struct{}
```

- [ ] Definir todos los nodos AST
- [ ] Implementar visitor pattern
- [ ] Implementar String() para debugging

#### 5.3 Parser FHIRPath
```go
// pkg/fhirpath/parser/parser.go
package parser

type Parser struct {
    lexer   *Lexer
    current Token
    peek    Token
    errors  []string
}

func NewParser(input string) *Parser
func (p *Parser) Parse() (Expression, error)
func (p *Parser) parseExpression(precedence int) (Expression, error)
func (p *Parser) parseUnary() (Expression, error)
func (p *Parser) parseInvocation() (Expression, error)
func (p *Parser) parseTerm() (Expression, error)
func (p *Parser) parseFunction() (*FunctionInvocation, error)
```

- [ ] Implementar parser recursivo descendente
- [ ] Implementar precedencia de operadores
- [ ] Implementar parsing de funciones
- [ ] Implementar parsing de indexers
- [ ] Implementar manejo de errores con posicion

#### 5.4 Evaluador FHIRPath
```go
// pkg/fhirpath/evaluator/evaluator.go
package evaluator

type Context struct {
    Resource    interface{}
    RootContext interface{}
    This        interface{}
    Index       *int
    Total       *int
    Environment map[string]interface{}
    FHIRVersion string
}

type Evaluator struct {
    functions map[string]FHIRPathFunction
}

type FHIRPathFunction func(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)

func NewEvaluator() *Evaluator
func (e *Evaluator) Evaluate(expr parser.Expression, ctx *Context) ([]interface{}, error)
func (e *Evaluator) evalBinary(node *parser.BinaryExpr, ctx *Context, input []interface{}) ([]interface{}, error)
func (e *Evaluator) evalMember(name string, input []interface{}) ([]interface{}, error)
func (e *Evaluator) evalFunction(fn *parser.FunctionInvocation, ctx *Context, input []interface{}) ([]interface{}, error)
```

- [ ] Implementar evaluador base
- [ ] Implementar navegacion de propiedades via reflection
- [ ] Implementar navegacion en map[string]interface{}
- [ ] Implementar operadores aritmeticos
- [ ] Implementar operadores de comparacion
- [ ] Implementar operadores logicos
- [ ] Implementar operador union (|)
- [ ] Implementar operador in/contains

#### 5.5 Funciones Built-in
```go
// pkg/fhirpath/evaluator/functions.go

// Existence
func funcExists(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcEmpty(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcNot(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)

// Filtering/Projection
func funcWhere(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcSelect(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcAll(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcAny(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcRepeat(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)

// Subsetting
func funcFirst(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcLast(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcTail(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcTake(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcSkip(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcSingle(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)

// Aggregates
func funcCount(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcSum(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcMin(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcMax(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcAvg(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)

// String functions
func funcStartsWith(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcEndsWith(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcContains(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcMatches(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcReplace(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcSubstring(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcLength(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcUpper(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcLower(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcToChars(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcIndexOf(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)

// Type functions
func funcOfType(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcAs(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcIs(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)

// Math functions
func funcAbs(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcCeiling(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcFloor(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcRound(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcSqrt(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcLn(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcLog(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcPower(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcTruncate(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcExp(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)

// Date/Time functions
func funcNow(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcToday(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcTimeOfDay(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)

// Utility
func funcIif(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcTrace(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)

// FHIR-specific
func funcExtension(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcHasValue(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcGetValue(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcResolve(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
func funcMemberOf(ctx *Context, input []interface{}, args []parser.Expression) ([]interface{}, error)
```

- [ ] Implementar funciones de existencia (8 funciones)
- [ ] Implementar funciones de filtrado (5 funciones)
- [ ] Implementar funciones de subsetting (6 funciones)
- [ ] Implementar funciones de agregacion (5 funciones)
- [ ] Implementar funciones de string (12 funciones)
- [ ] Implementar funciones de tipo (3 funciones)
- [ ] Implementar funciones matematicas (10 funciones)
- [ ] Implementar funciones de fecha (3 funciones)
- [ ] Implementar funciones utilitarias (2 funciones)
- [ ] Implementar funciones FHIR-especificas (5 funciones)

#### 5.6 Cache de Expresiones Compiladas
```go
// pkg/fhirpath/compiler/compiler.go
package compiler

type CompiledExpression struct {
    ast       parser.Expression
    source    string
}

type Compiler struct {
    cache    *lru.Cache
    mu       sync.RWMutex
}

func NewCompiler(cacheSize int) *Compiler
func (c *Compiler) Compile(expression string) (*CompiledExpression, error)
```

- [ ] Implementar LRU cache
- [ ] Implementar compilacion con cache
- [ ] Implementar estadisticas de cache

#### 5.7 API Publica FHIRPath
```go
// pkg/fhirpath/fhirpath.go
package fhirpath

var defaultCompiler = compiler.NewCompiler(500)

func Evaluate(expression string, resource interface{}) ([]interface{}, error)
func EvaluateWithContext(expression string, ctx *evaluator.Context) ([]interface{}, error)
func EvaluateToBoolean(expression string, resource interface{}) (bool, error)
func EvaluateToString(expression string, resource interface{}) (string, error)
func Compile(expression string) (*compiler.CompiledExpression, error)
func NewContext(resource interface{}) *evaluator.Context
```

- [ ] Implementar API de alto nivel
- [ ] Implementar helpers de conversion
- [ ] Documentar API publica

### Tests Sprint 5
- [ ] Tests exhaustivos del lexer
- [ ] Tests del parser con expresiones complejas
- [ ] Tests de cada funcion built-in
- [ ] Tests con constraints FHIR reales
- [ ] Benchmarks de evaluacion
- [ ] Tests con suite oficial FHIRPath (si disponible)

### Entregables
- Package `pkg/fhirpath` completo
- 60+ funciones built-in
- Cache LRU de expresiones
- 95%+ cobertura de tests
- Documentacion completa

---

## Sprint 6: Sistema de Validacion (3 semanas)

### Objetivos
- Implementar validador estructural
- Integrar validacion FHIRPath
- Implementar validacion de primitivos
- **Definir interfaces para testing/mocking**

### Tareas

#### 6.0 Interfaces para Testing y Extensibilidad

```go
// pkg/validator/interfaces.go - MANUAL
package validator

import "context"

// ReferenceResolver permite mockear resolucion de referencias externas
// Util para tests y para implementaciones de servidor FHIR
type ReferenceResolver interface {
    Resolve(ctx context.Context, reference string) (interface{}, error)
}

// TerminologyService permite mockear validacion de terminologia
// Implementaciones: LocalTerminology, RemoteTerminology (tx.fhir.org)
type TerminologyService interface {
    ValidateCode(ctx context.Context, system, code string, valueSetURL string) (bool, error)
    ExpandValueSet(ctx context.Context, valueSetURL string) ([]string, error)
    LookupCode(ctx context.Context, system, code string) (*CodeInfo, error)
}

// CodeInfo contiene informacion de un codigo
type CodeInfo struct {
    System  string
    Code    string
    Display string
    Active  bool
}

// StructureDefinitionProvider permite cargar SDs desde diferentes fuentes
type StructureDefinitionProvider interface {
    Get(ctx context.Context, url string) (*StructureDefinition, error)
    List(ctx context.Context) ([]string, error)
}

// NoopReferenceResolver no resuelve nada (para validacion local)
type NoopReferenceResolver struct{}

func (n *NoopReferenceResolver) Resolve(ctx context.Context, ref string) (interface{}, error) {
    return nil, nil // No error, simplemente no resuelve
}

// LocalTerminologyService valida contra ValueSets embebidos
type LocalTerminologyService struct {
    valueSets map[string]*ValueSet
}

func NewLocalTerminologyService() *LocalTerminologyService
func (l *LocalTerminologyService) LoadValueSet(vs *ValueSet) error
```

- [ ] Definir interface ReferenceResolver
- [ ] Definir interface TerminologyService
- [ ] Definir interface StructureDefinitionProvider
- [ ] Implementar NoopReferenceResolver
- [ ] Implementar LocalTerminologyService

#### 6.1 Registry de Especificaciones
```go
// pkg/validator/registry.go
package validator

type SpecRegistry struct {
    structureDefinitions map[string]*StructureDefinition
    valueSets           map[string]*ValueSet
    codeSystems         map[string]*CodeSystem
    version             string
    mu                  sync.RWMutex
}

func NewSpecRegistry(version string) (*SpecRegistry, error)
func (r *SpecRegistry) GetStructureDefinition(name string) (*StructureDefinition, error)
func (r *SpecRegistry) GetValueSet(url string) (*ValueSet, error)
func (r *SpecRegistry) GetCodeSystem(url string) (*CodeSystem, error)
func (r *SpecRegistry) LoadImplementationGuide(path string) error
```

- [ ] Implementar carga lazy de specs
- [ ] Implementar cache de specs cargadas
- [ ] Implementar resolucion de URLs canonicas
- [ ] Implementar carga de IGs personalizados

#### 6.2 Validador Principal
```go
// pkg/validator/validator.go
package validator

type ValidatorOptions struct {
    FHIRVersion          string
    ValidateConstraints  bool
    ValidateTerminology  bool
    ValidateReferences   bool
    ErrorOnWarning       bool
    TerminologyServer    string
    MaxErrors            int
}

type FHIRValidator struct {
    options      *ValidatorOptions
    specRegistry *SpecRegistry
    validators   []SubValidator
    fhirpath     *fhirpath.Compiler
}

type SubValidator interface {
    Validate(ctx context.Context, vctx *ValidationContext) ([]Issue, error)
}

type ValidationContext struct {
    Resource        interface{}
    ResourceType    string
    StructureDef    *StructureDefinition
    Path            string
    SpecRegistry    *SpecRegistry
    FHIRVersion     string
}

func NewValidator(options *ValidatorOptions) (*FHIRValidator, error)
func (v *FHIRValidator) Validate(ctx context.Context, resource interface{}) (*OperationOutcome, error)
func (v *FHIRValidator) ValidateBundle(ctx context.Context, bundle interface{}) (*OperationOutcome, error)
```

- [ ] Implementar orquestador de validacion
- [ ] Implementar contexto de validacion
- [ ] Implementar agregacion de issues

#### 6.3 OperationOutcome
```go
// pkg/validator/outcome.go
package validator

type Severity string
const (
    SeverityFatal       Severity = "fatal"
    SeverityError       Severity = "error"
    SeverityWarning     Severity = "warning"
    SeverityInformation Severity = "information"
)

type IssueCode string
const (
    CodeInvalid      IssueCode = "invalid"
    CodeStructure    IssueCode = "structure"
    CodeRequired     IssueCode = "required"
    CodeValue        IssueCode = "value"
    CodeInvariant    IssueCode = "invariant"
    CodeCodeInvalid  IssueCode = "code-invalid"
    // ...
)

type Issue struct {
    Severity    Severity
    Code        IssueCode
    Diagnostics string
    Location    []string
    Expression  []string
}

type OperationOutcome struct {
    ResourceType string  `json:"resourceType"`
    Issues       []Issue `json:"issue"`
}

func NewOperationOutcome() *OperationOutcome
func (o *OperationOutcome) AddIssue(severity Severity, code IssueCode, msg string, path []string)
func (o *OperationOutcome) HasErrors() bool
func (o *OperationOutcome) IsSuccess() bool
func (o *OperationOutcome) ErrorCount() int
```

- [ ] Implementar struct OperationOutcome
- [ ] Implementar todos los severity levels
- [ ] Implementar todos los issue codes FHIR
- [ ] Implementar metodos de conveniencia

#### 6.4 Validador Estructural
```go
// pkg/validator/validators/structure.go
package validators

type StructureValidator struct {
    registry *SpecRegistry
}

func NewStructureValidator(registry *SpecRegistry) *StructureValidator

func (v *StructureValidator) Validate(ctx context.Context, vctx *ValidationContext) ([]Issue, error) {
    // Validar:
    // - Campos requeridos presentes (min >= 1)
    // - Cardinalidad maxima respetada (max)
    // - Solo campos definidos presentes
    // - Tipos correctos
}
```

- [ ] Implementar validacion de campos requeridos
- [ ] Implementar validacion de cardinalidad
- [ ] Implementar validacion de campos desconocidos
- [ ] Implementar validacion de tipos

#### 6.5 Validador de Primitivos
```go
// pkg/validator/validators/primitive.go
package validators

type PrimitiveValidator struct {
    patterns map[string]*regexp.Regexp
}

func NewPrimitiveValidator() *PrimitiveValidator

var primitivePatterns = map[string]string{
    "id":           `^[A-Za-z0-9\-\.]{1,64}$`,
    "uri":          `^\S*$`,
    "url":          `^\S*$`,
    "canonical":    `^\S*$`,
    "code":         `^[^\s]+(\s[^\s]+)*$`,
    "oid":          `^urn:oid:[0-2](\.(0|[1-9][0-9]*))+$`,
    "uuid":         `^urn:uuid:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
    "date":         `^([0-9]{4})(-[0-9]{2}(-[0-9]{2})?)?$`,
    "dateTime":     `...regex complejo...`,
    "time":         `^([01][0-9]|2[0-3]):[0-5][0-9]:([0-5][0-9]|60)(\.[0-9]+)?$`,
    "instant":      `...regex complejo...`,
    "base64Binary": `^[A-Za-z0-9+/]*={0,2}$`,
    "positiveInt":  `^[1-9][0-9]*$`,
    "unsignedInt":  `^[0-9]+$`,
}

func (v *PrimitiveValidator) Validate(ctx context.Context, vctx *ValidationContext) ([]Issue, error)
```

- [ ] Implementar patrones regex para todos los primitivos
- [ ] Implementar validacion recursiva
- [ ] Implementar mensajes de error claros

#### 6.6 Validador de Constraints (FHIRPath)
```go
// pkg/validator/validators/constraint.go
package validators

type ConstraintValidator struct {
    registry *SpecRegistry
    compiler *fhirpath.Compiler
}

type Constraint struct {
    Key        string
    Severity   string
    Human      string
    Expression string
    Context    string
}

func NewConstraintValidator(registry *SpecRegistry) *ConstraintValidator

func (v *ConstraintValidator) Validate(ctx context.Context, vctx *ValidationContext) ([]Issue, error) {
    constraints := v.extractConstraints(vctx.StructureDef)

    for _, c := range constraints {
        result, err := v.evaluate(c.Expression, vctx.Resource)
        if err != nil {
            // Warning: error evaluando constraint
        }
        if !result {
            // Error o warning segun c.Severity
        }
    }
}
```

- [ ] Implementar extraccion de constraints de StructureDefinition
- [ ] Implementar evaluacion de constraints
- [ ] Implementar manejo de errores de evaluacion
- [ ] Implementar contexto correcto ($this, etc)

#### 6.7 Validador de Referencias
```go
// pkg/validator/validators/reference.go
package validators

type ReferenceValidator struct {
    registry *SpecRegistry
}

func NewReferenceValidator(registry *SpecRegistry) *ReferenceValidator

func (v *ReferenceValidator) Validate(ctx context.Context, vctx *ValidationContext) ([]Issue, error) {
    // Validar:
    // - Formato de referencia correcto
    // - Tipo de recurso referenciado permitido
    // - Referencias internas resolvibles (contained, Bundle)
}
```

- [ ] Implementar validacion de formato de referencia
- [ ] Implementar validacion de tipos permitidos
- [ ] Implementar validacion de referencias contenidas

#### 6.8 Validador de Extensiones
```go
// pkg/validator/validators/extension.go
package validators

type ExtensionValidator struct {
    registry *SpecRegistry
}

func NewExtensionValidator(registry *SpecRegistry) *ExtensionValidator

func (v *ExtensionValidator) Validate(ctx context.Context, vctx *ValidationContext) ([]Issue, error) {
    // Validar:
    // - URL de extension valida
    // - StructureDefinition de extension existe (si conocida)
    // - Tipo de valor correcto
    // - Contexto de uso correcto
}
```

- [ ] Implementar validacion de URLs de extension
- [ ] Implementar validacion contra StructureDefinition
- [ ] Implementar validacion de contexto

### Tests Sprint 6
- [ ] Tests de validacion estructural
- [ ] Tests de validacion de primitivos
- [ ] Tests de constraints FHIRPath
- [ ] Tests con recursos validos e invalidos
- [ ] Tests con recursos FHIR de ejemplo oficiales
- [ ] Benchmarks de validacion

### Entregables
- Package `pkg/validator` completo
- Validadores: estructura, primitivos, constraints, referencias, extensiones
- OperationOutcome completo
- 90%+ cobertura de tests

---

## Sprint 7: Multi-Version y CLI (2 semanas)

### Objetivos
- Generar packages R4B y R5
- Implementar CLI tool
- Agregar validacion de terminologia

### Tareas

#### 7.1 Generacion R4B
- [ ] Adaptar generador para diferencias R4B
- [ ] Generar `pkg/fhir/r4b/datatypes/`
- [ ] Generar `pkg/fhir/r4b/resources/`
- [ ] Generar `pkg/fhir/r4b/builders/`
- [ ] Generar `pkg/fhir/r4b/valuesets/`
- [ ] Actualizar registry para R4B

#### 7.2 Generacion R5
- [ ] Adaptar generador para diferencias R5
- [ ] Manejar nuevos recursos R5
- [ ] Manejar cambios de estructura R5
- [ ] Generar `pkg/fhir/r5/*`

#### 7.3 Validador de Terminologia
```go
// pkg/validator/validators/terminology.go
package validators

type TerminologyValidator struct {
    client      *TerminologyClient
    cache       *lru.Cache
    codeSystem  map[string]*CodeSystem // Embebidos
}

type TerminologyClient struct {
    serverURL string
    http      *http.Client
}

func NewTerminologyValidator(serverURL string, cacheSize int) (*TerminologyValidator, error)

func (v *TerminologyValidator) Validate(ctx context.Context, vctx *ValidationContext) ([]Issue, error) {
    // Validar bindings:
    // - required: codigo debe estar en ValueSet
    // - extensible: warning si no esta
    // - preferred: info si no esta
    // - example: sin validacion
}

func (v *TerminologyValidator) ValidateCode(ctx context.Context, system, code string, valueSetURL string) (bool, error)
```

- [ ] Implementar cliente de terminologia (tx.fhir.org)
- [ ] Implementar cache de validaciones
- [ ] Implementar ValueSets embebidos comunes
- [ ] Implementar validacion por binding strength

#### 7.4 CLI Tool
```go
// cmd/gofhir/main.go
package main

import "github.com/spf13/cobra"

func main() {
    rootCmd := &cobra.Command{
        Use:   "gofhir",
        Short: "GoFHIR - FHIR Toolkit for Go",
    }

    // Subcommands
    rootCmd.AddCommand(
        validateCmd(),
        fhirpathCmd(),
        generateCmd(),
        versionCmd(),
    )

    rootCmd.Execute()
}

// validate: Validar recursos FHIR
func validateCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "validate [file]",
        Short: "Validate a FHIR resource",
        RunE:  runValidate,
    }
    cmd.Flags().StringP("version", "v", "R4", "FHIR version")
    cmd.Flags().Bool("constraints", true, "Validate FHIRPath constraints")
    cmd.Flags().Bool("terminology", false, "Validate terminology bindings")
    cmd.Flags().StringP("output", "o", "text", "Output format (text, json)")
    return cmd
}

// fhirpath: Evaluar expresiones FHIRPath
func fhirpathCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "fhirpath [expression] [file]",
        Short: "Evaluate a FHIRPath expression",
        RunE:  runFHIRPath,
    }
    return cmd
}

// generate: Regenerar tipos desde specs
func generateCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "generate",
        Short: "Generate Go types from FHIR specs",
        RunE:  runGenerate,
    }
    cmd.Flags().String("specs", "./specs", "Path to FHIR specs")
    cmd.Flags().String("output", "./pkg/fhir", "Output directory")
    cmd.Flags().StringSlice("versions", []string{"r4"}, "FHIR versions to generate")
    return cmd
}
```

- [ ] Implementar comando `validate`
- [ ] Implementar comando `fhirpath`
- [ ] Implementar comando `generate`
- [ ] Implementar output formateado (text, json)
- [ ] Implementar colores en terminal
- [ ] Agregar ejemplos en help

### Tests Sprint 7
- [ ] Tests de generacion R4B
- [ ] Tests de generacion R5
- [ ] Tests de validacion de terminologia
- [ ] Tests E2E del CLI
- [ ] Tests de integracion multi-version

### Entregables
- Packages R4B y R5 generados
- Validador de terminologia
- CLI tool funcional
- Documentacion de CLI

---

## Sprint 8: Bundle, Polish y Documentacion (2 semanas)

### Objetivos
- Validacion especial de Bundle
- Pulir API y corregir edge cases
- Crear documentacion completa

### Tareas

#### 8.1 Validador de Bundle
```go
// pkg/validator/validators/bundle.go
package validators

type BundleValidator struct {
    validator *FHIRValidator
}

func NewBundleValidator(v *FHIRValidator) *BundleValidator

func (v *BundleValidator) Validate(ctx context.Context, bundle interface{}) (*OperationOutcome, error) {
    // Validaciones especiales:
    // - fullUrl unico
    // - Referencias entre entries resolvibles
    // - Validacion de cada entry.resource
    // - Reglas especificas por bundle.type (transaction, document, etc)
}
```

- [ ] Implementar validacion de fullUrl
- [ ] Implementar validacion de referencias internas
- [ ] Implementar reglas por tipo de Bundle
- [ ] Implementar validacion transaccional

#### 8.2 Slicing Support
```go
// pkg/validator/validators/slicing.go
package validators

type SlicingValidator struct {
    registry *SpecRegistry
}

func (v *SlicingValidator) Validate(ctx context.Context, vctx *ValidationContext) ([]Issue, error) {
    // Validar discriminadores de slicing
    // Validar cardinalidad por slice
}
```

- [ ] Implementar parsing de slicing rules
- [ ] Implementar validacion de discriminadores
- [ ] Implementar validacion de cardinalidad por slice

#### 8.3 API Polish
- [ ] Revisar y unificar nombres de funciones
- [ ] Agregar godoc a todas las funciones publicas
- [ ] Revisar manejo de nil/empty
- [ ] Agregar mas examples en godoc
- [ ] Crear package-level doc

#### 8.4 Documentacion
```
docs/
├── getting-started.md
├── api-reference.md
├── builders.md
├── fhirpath.md
├── validation.md
├── code-generation.md
└── examples/
    ├── basic-usage.md
    ├── creating-resources.md
    ├── validation.md
    └── fhirpath-queries.md
```

- [ ] Escribir Getting Started
- [ ] Documentar API de tipos
- [ ] Documentar Builders
- [ ] Documentar FHIRPath
- [ ] Documentar Validacion
- [ ] Crear ejemplos completos

#### 8.5 Examples
```go
// examples/basic/main.go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/robertoaraneda/gofhir/pkg/common"
    "github.com/robertoaraneda/gofhir/pkg/fhir/r4"
)

func main() {
    // Crear paciente usando builder (API fluida)
    patient := r4.NewPatientBuilder().
        SetID("example-001").
        SetActive(true).
        AddName(r4.HumanName{
            Family: common.String("Garcia"),
            Given:  []string{"Maria"},
        }).
        SetGender(r4.AdministrativeGenderFemale).
        SetBirthDate("1985-03-15").
        Build()

    // Serializar a JSON (stdlib, no metodo custom)
    data, _ := json.MarshalIndent(patient, "", "  ")
    fmt.Println(string(data))

    // Clonar paciente (funcion generica)
    patient2 := common.Clone(patient)
    patient2.ID = common.String("example-002")

    // Acceso directo a campos (no getters)
    fmt.Printf("Patient ID: %s\n", *patient.ID)
    fmt.Printf("Patient Active: %v\n", *patient.Active)
}
```

- [ ] Crear ejemplo basico
- [ ] Crear ejemplo de builders
- [ ] Crear ejemplo de validacion
- [ ] Crear ejemplo de FHIRPath
- [ ] Crear ejemplo de Bundle

### Tests Sprint 8
- [ ] Tests de Bundle validation
- [ ] Tests de slicing
- [ ] Tests de documentacion (ejemplos funcionan)
- [ ] Tests de performance final
- [ ] Revision de cobertura total

### Entregables
- Validacion de Bundle completa
- Soporte de slicing
- Documentacion completa
- Ejemplos funcionales
- README completo

---

## Metricas de Exito

### Cobertura de Codigo
- Minimo 80% cobertura total
- 95%+ en packages criticos (fhirpath, validator)

### Performance
- Parse FHIRPath < 1ms para expresiones comunes
- Evaluacion FHIRPath < 5ms para recursos tipicos
- Validacion completa < 50ms para recursos tipicos
- Serializacion JSON < 1ms para recursos tipicos

### Compatibilidad
- 100% de recursos R4 generados
- 100% de constraints R4 evaluables
- Compatible con Go 1.21+

### Calidad
- Zero panics en uso normal
- Errores descriptivos con paths
- API consistente e idiomatica

---

## Dependencias Externas

```go
// go.mod
module github.com/robertoaraneda/gofhir

go 1.22

require (
    // CLI - cobra es el estandar de facto para CLIs en Go
    github.com/spf13/cobra v1.8.0

    // Testing - testify para assertions mas legibles
    github.com/stretchr/testify v1.9.0

    // Cache LRU - para cache de expresiones FHIRPath compiladas
    github.com/hashicorp/golang-lru/v2 v2.0.7

    // HTTP Client mejorado - para TerminologyService remoto
    github.com/hashicorp/go-retryablehttp v0.7.5
)
```

### Justificacion de Dependencias

| Dependencia | Uso | Alternativa stdlib |
|-------------|-----|-------------------|
| `cobra` | CLI con subcomandos, flags, help automatico | `flag` - menos features |
| `testify` | Assertions, mocks, suites | `testing` - mas verbose |
| `golang-lru/v2` | Cache LRU thread-safe con generics | `sync.Map` - sin limite de tamaño |
| `go-retryablehttp` | Retry automatico, backoff exponencial | `net/http` - retry manual |

### Dependencias NO incluidas (decision consciente)

- **No reflection libraries**: Usamos codegen, no reflection en runtime
- **No ORM**: Esto es un toolkit FHIR, no maneja persistencia
- **No logging framework**: Dejamos que el usuario elija (slog, zap, zerolog)
- **No validation framework**: Implementamos nuestro propio validador FHIR

---

## Riesgos y Mitigaciones

| Riesgo | Impacto | Probabilidad | Mitigacion |
|--------|---------|--------------|------------|
| Complejidad FHIRPath mayor a esperada | Alto | Media | Implementar subset primero, agregar funciones incrementalmente |
| Diferencias entre versiones FHIR | Medio | Alta | Abstraer generador para manejar diferencias |
| Performance de validacion | Medio | Media | Profiling temprano, optimizacion de hot paths |
| Edge cases en JSON serialization | Bajo | Alta | Tests exhaustivos con fixtures oficiales |

---

## Timeline Resumen

| Sprint | Duracion | Entregable Principal |
|--------|----------|---------------------|
| Sprint 0 | 1 semana | Fundacion: estructura, CI/CD, specs FHIR |
| Sprint 1 | 2 semanas | Parser de StructureDefinitions + Analyzer |
| Sprint 2 | 2 semanas | Templates + Generacion completa R4 (100% codegen) |
| Sprint 3 | 2 semanas | Metodos generados + Registry + Interfaces |
| Sprint 4 | 2 semanas | Builders generados + Helpers clinicos |
| Sprint 5 | 3 semanas | Motor FHIRPath completo |
| Sprint 6 | 3 semanas | Sistema de validacion YAFV |
| Sprint 7 | 2 semanas | Multi-version (R4B, R5) + CLI |
| Sprint 8 | 2 semanas | Bundle validation, polish, docs |

**Total estimado: 19 semanas**

### Principio Clave del Timeline

```text
Sprint 0-4: TODO el codigo FHIR (datatypes, resources, backbones, builders)
            se genera automaticamente desde StructureDefinitions.
            NO hay herencia, NO hay embedding, structs planos.

Sprint 5-8: Funcionalidad adicional (FHIRPath, Validator, CLI)
            se implementa manualmente.
```

---

## Uso de Packages

### Package Standalone: Solo Tipos
```go
import (
    "github.com/robertoaraneda/gofhir/pkg/common"
    "github.com/robertoaraneda/gofhir/pkg/fhir/r4"
)

// Crear recurso directamente (struct literal)
patient := &r4.Patient{
    ResourceType: "Patient",
    ID:           common.String("123"),
    Active:       common.Bool(true),
    BirthDate:    common.String("1990-01-01"),
}

// Con extension en primitivo
patient.BirthDateExt = &r4.Element{
    Extension: []r4.Extension{{
        URL:       "http://example.org/birth-time",
        ValueTime: common.String("14:30:00"),
    }},
}
```

### Package Standalone: Solo FHIRPath
```go
import "github.com/robertoaraneda/gofhir/pkg/fhirpath"

result, _ := fhirpath.Evaluate("name.given.first()", patient)
names, _ := fhirpath.EvaluateToStrings("name.family", patient)
exists, _ := fhirpath.EvaluateToBoolean("active.exists()", patient)
```

### Package Standalone: Solo Validacion
```go
import "github.com/robertoaraneda/gofhir/pkg/validator"

v, _ := validator.NewValidator(&validator.Options{
    FHIRVersion:         "R4",
    ValidateConstraints:  true,
    ValidateTerminology:  false, // Sin servidor de terminologia
})
outcome, _ := v.Validate(ctx, patient)

if outcome.HasErrors() {
    for _, issue := range outcome.Issues {
        fmt.Printf("[%s] %s: %s\n", issue.Severity, issue.Location, issue.Diagnostics)
    }
}
```

### Full Toolkit con Builders

```go
import (
    "github.com/robertoaraneda/gofhir/pkg/common"
    "github.com/robertoaraneda/gofhir/pkg/fhir/r4"
    "github.com/robertoaraneda/gofhir/pkg/fhirpath"
    "github.com/robertoaraneda/gofhir/pkg/validator"
)

// Builder fluido
patient := r4.NewPatientBuilder().
    SetID("123").
    SetActive(true).
    SetBirthDate("1990-01-01").
    AddName(r4.HumanName{
        Family: common.String("Garcia"),
        Given:  []string{"Maria", "Elena"},
    }).
    Build()

// Clonar
patient2 := common.Clone(patient)

// FHIRPath
names, _ := fhirpath.EvaluateToStrings("name.given", patient)

// Validar
v, _ := validator.NewValidator(validator.DefaultOptions("R4"))
outcome, _ := v.Validate(ctx, patient)
```
