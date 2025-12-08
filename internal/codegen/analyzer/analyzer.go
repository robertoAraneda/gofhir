// Package analyzer analyzes FHIR StructureDefinitions and determines Go types.
package analyzer

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/robertoaraneda/gofhir/internal/codegen/parser"
)

// Analyzer processes StructureDefinitions and produces analyzed types for code generation.
type Analyzer struct {
	definitions  map[string]*parser.StructureDefinition
	valueSets    *parser.ValueSetRegistry
	UsedBindings map[string]bool // Track which bindings are used (exported for generator)
}

// NewAnalyzer creates a new Analyzer with the given StructureDefinitions and ValueSets.
func NewAnalyzer(definitions []*parser.StructureDefinition, valueSets *parser.ValueSetRegistry) *Analyzer {
	defMap := make(map[string]*parser.StructureDefinition)
	for _, sd := range definitions {
		defMap[sd.URL] = sd
		defMap[sd.Name] = sd
		defMap[sd.Type] = sd
	}
	return &Analyzer{
		definitions:  defMap,
		valueSets:    valueSets,
		UsedBindings: make(map[string]bool),
	}
}

// AnalyzedType represents a fully analyzed type ready for code generation.
type AnalyzedType struct {
	Name        string             // Go type name (PascalCase)
	FHIRName    string             // Original FHIR name
	Kind        string             // primitive, datatype, resource, backbone
	Description string             // Documentation
	URL         string             // Canonical URL
	IsAbstract  bool               // Whether this is an abstract type
	Properties  []AnalyzedProperty // Fields of this type
	Constraints []AnalyzedConstraint
}

// AnalyzedProperty represents a single property of a type.
type AnalyzedProperty struct {
	Name         string   // Go field name (PascalCase)
	JSONName     string   // JSON field name (camelCase)
	GoType       string   // Complete Go type (e.g., "*string", "[]Coding")
	Description  string   // Documentation
	IsPointer    bool     // Whether this field is a pointer
	IsArray      bool     // Whether this field is an array
	IsRequired   bool     // Whether min >= 1
	IsPrimitive  bool     // Whether the base type is a primitive
	IsChoice     bool     // Whether this is a choice type field
	ChoiceTypes  []string // For choice types, the list of allowed types
	FHIRType     string   // Original FHIR type code
	Binding      *AnalyzedBinding
	HasExtension bool // Whether this primitive needs a _field for extensions
}

// AnalyzedBinding represents a value set binding.
type AnalyzedBinding struct {
	Strength string // required, extensible, preferred, example
	ValueSet string // ValueSet URL
}

// AnalyzedConstraint represents a FHIRPath constraint.
type AnalyzedConstraint struct {
	Key        string
	Severity   string
	Human      string
	Expression string
}

// Analyze processes a StructureDefinition and returns an AnalyzedType.
func (a *Analyzer) Analyze(sd *parser.StructureDefinition) (*AnalyzedType, error) {
	if sd == nil {
		return nil, fmt.Errorf("StructureDefinition is nil")
	}

	kind := a.determineKind(sd)

	analyzed := &AnalyzedType{
		Name:        sd.Name,
		FHIRName:    sd.Name,
		Kind:        kind,
		Description: sd.Title,
		URL:         sd.URL,
		IsAbstract:  sd.Abstract,
	}

	elements := sd.GetElements()
	if len(elements) == 0 {
		return analyzed, nil
	}

	// Skip the root element (first element is always the type itself)
	for i := 1; i < len(elements); i++ {
		elem := elements[i]

		// Skip slices for now
		if elem.SliceName != "" {
			continue
		}

		// Skip nested elements (backbone children) - they'll be handled separately
		if a.isNestedElement(elem.Path, sd.Type) {
			continue
		}

		props, err := a.analyzeElement(&elem, sd.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze element %s: %w", elem.Path, err)
		}
		analyzed.Properties = append(analyzed.Properties, props...)
	}

	// Extract constraints from the root element
	if len(elements) > 0 {
		for _, c := range elements[0].Constraint {
			analyzed.Constraints = append(analyzed.Constraints, AnalyzedConstraint{
				Key:        c.Key,
				Severity:   c.Severity,
				Human:      c.Human,
				Expression: c.Expression,
			})
		}
	}

	return analyzed, nil
}

// determineKind determines the kind of type (primitive, datatype, resource, backbone).
func (a *Analyzer) determineKind(sd *parser.StructureDefinition) string {
	switch sd.Kind {
	case parser.KindPrimitiveType:
		return "primitive"
	case parser.KindResource:
		return "resource"
	case parser.KindComplexType:
		// Check if it's a backbone element
		if strings.Contains(sd.BaseDefinition, "BackboneElement") {
			return "backbone"
		}
		return "datatype"
	default:
		return "datatype"
	}
}

// isNestedElement checks if an element path indicates a nested (backbone) element.
func (a *Analyzer) isNestedElement(path, rootType string) bool {
	// Remove the root type prefix
	suffix := strings.TrimPrefix(path, rootType+".")
	if suffix == path {
		return false
	}
	// If there's still a dot, it's nested
	return strings.Contains(suffix, ".")
}

// analyzeElement analyzes a single element and returns properties.
// May return multiple properties for choice types.
func (a *Analyzer) analyzeElement(elem *parser.ElementDefinition, rootType string) ([]AnalyzedProperty, error) {
	// Get the field name from the path
	fieldName := a.extractFieldName(elem.Path, rootType)
	if fieldName == "" {
		return nil, nil
	}

	// Handle choice types (value[x], effective[x], etc.)
	if elem.IsChoiceType() {
		return a.analyzeChoiceType(elem, fieldName)
	}

	// Handle content references
	if elem.ContentReference != "" {
		return a.analyzeContentReference(elem, fieldName)
	}

	// Regular element
	if len(elem.Type) == 0 {
		// Backbone element - will be handled separately
		return nil, nil
	}

	prop := a.createProperty(elem, fieldName, elem.Type[0])
	return []AnalyzedProperty{prop}, nil
}

// analyzeChoiceType handles choice type elements like value[x].
func (a *Analyzer) analyzeChoiceType(elem *parser.ElementDefinition, baseName string) ([]AnalyzedProperty, error) {
	props := make([]AnalyzedProperty, 0, len(elem.Type)*2) // *2 for extension fields
	choiceTypes := make([]string, 0, len(elem.Type))

	for _, typeRef := range elem.Type {
		choiceTypes = append(choiceTypes, typeRef.Code)
	}

	// Generate a property for each possible type
	for _, typeRef := range elem.Type {
		typeName := typeRef.Code
		// Field name: PascalCase(baseName) + PascalCase(typeName)
		// e.g., "deceased" + "Boolean" = "DeceasedBoolean"
		fieldName := toPascalCase(baseName) + toPascalCase(typeName)

		prop := AnalyzedProperty{
			Name:         fieldName,
			JSONName:     toLowerFirst(baseName) + toPascalCase(typeName),
			GoType:       a.resolveGoType(typeName, true, false), // Choice types are always pointers
			Description:  elem.Short,
			IsPointer:    true, // Choice types are always optional
			IsArray:      false,
			IsRequired:   false,
			IsPrimitive:  IsPrimitiveType(typeName),
			IsChoice:     true,
			ChoiceTypes:  choiceTypes,
			FHIRType:     typeName,
			HasExtension: IsPrimitiveType(typeName),
		}

		if elem.Binding != nil {
			prop.Binding = &AnalyzedBinding{
				Strength: elem.Binding.Strength,
				ValueSet: elem.Binding.ValueSet,
			}
		}

		props = append(props, prop)

		// Add extension field for primitives
		if prop.HasExtension {
			extProp := AnalyzedProperty{
				Name:        fieldName + "Ext",
				JSONName:    "_" + toLowerFirst(baseName) + toPascalCase(typeName),
				GoType:      "*Element",
				Description: fmt.Sprintf("Extension for %s", fieldName),
				IsPointer:   true,
				IsArray:     false,
				IsPrimitive: false,
				FHIRType:    "Element",
			}
			props = append(props, extProp)
		}
	}

	return props, nil
}

// analyzeContentReference handles content references.
func (a *Analyzer) analyzeContentReference(elem *parser.ElementDefinition, fieldName string) ([]AnalyzedProperty, error) {
	// Content references point to another element's definition
	// For now, treat as a generic type that will be resolved later
	prop := AnalyzedProperty{
		Name:        toGoFieldName(fieldName),
		JSONName:    toLowerFirst(fieldName),
		GoType:      "*interface{}", // Will be resolved during generation
		Description: elem.Short,
		IsPointer:   true,
		IsArray:     elem.IsArray(),
		FHIRType:    "ContentReference",
	}
	return []AnalyzedProperty{prop}, nil
}

// createProperty creates an AnalyzedProperty from an element and type reference.
func (a *Analyzer) createProperty(elem *parser.ElementDefinition, fieldName string, typeRef parser.TypeRef) AnalyzedProperty {
	typeName := typeRef.Code
	isArray := elem.IsArray()
	isPrimitive := IsPrimitiveType(typeName)

	// Determine if pointer is needed
	// - Arrays don't need pointer (nil slice is fine)
	// - Required primitives could be non-pointer, but we use pointer for JSON omitempty
	// - Complex types are always pointers when optional
	isPointer := !isArray && (elem.Min == 0 || isPrimitive)

	// Check for required binding with code type - use custom type
	goType := a.resolveGoTypeWithBinding(typeName, isPointer, isArray, elem.Binding)

	prop := AnalyzedProperty{
		Name:         toGoFieldName(fieldName),
		JSONName:     toLowerFirst(fieldName),
		GoType:       goType,
		Description:  elem.Short,
		IsPointer:    isPointer,
		IsArray:      isArray,
		IsRequired:   elem.IsRequired(),
		IsPrimitive:  isPrimitive,
		IsChoice:     false,
		FHIRType:     typeName,
		HasExtension: isPrimitive,
	}

	if elem.Binding != nil {
		prop.Binding = &AnalyzedBinding{
			Strength: elem.Binding.Strength,
			ValueSet: elem.Binding.ValueSet,
		}
	}

	return prop
}

// resolveGoTypeWithBinding resolves Go type, using custom types for required bindings.
func (a *Analyzer) resolveGoTypeWithBinding(fhirType string, isPointer, isArray bool, binding *parser.Binding) string {
	// Only apply custom types for code fields with required binding
	if fhirType == "code" && binding != nil && binding.Strength == "required" {
		if vs := a.getValueSetForBinding(binding.ValueSet); vs != nil {
			// Track that this binding is used
			a.UsedBindings[binding.ValueSet] = true

			// Sanitize the type name to match what generator produces
			customType := sanitizeTypeName(vs.Name)
			if isArray {
				return "[]" + customType
			}
			if isPointer {
				return "*" + customType
			}
			return customType
		}
	}

	return a.resolveGoType(fhirType, isPointer, isArray)
}

// sanitizeTypeName converts a ValueSet name to a valid Go type name.
func sanitizeTypeName(name string) string {
	// Remove/replace invalid characters
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, ".", "")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	name = strings.ReplaceAll(name, "/", "")

	// Ensure first character is uppercase
	if name != "" {
		runes := []rune(name)
		runes[0] = unicode.ToUpper(runes[0])
		name = string(runes)
	}

	return name
}

// getValueSetForBinding retrieves and validates a ValueSet for use as a Go type.
func (a *Analyzer) getValueSetForBinding(url string) *parser.ParsedValueSet {
	if a.valueSets == nil {
		return nil
	}

	vs := a.valueSets.Get(url)
	if vs == nil || len(vs.Codes) == 0 {
		return nil
	}

	// Skip very large value sets (like all-types, mimetypes)
	if len(vs.Codes) > 100 {
		return nil
	}

	return vs
}

// resolveGoType converts a FHIR type to a Go type string.
func (a *Analyzer) resolveGoType(fhirType string, isPointer, isArray bool) string {
	goType := FHIRToGoType(fhirType)

	if isArray {
		return "[]" + goType
	}
	if isPointer {
		return "*" + goType
	}
	return goType
}

// extractFieldName extracts the field name from an element path.
func (a *Analyzer) extractFieldName(path, rootType string) string {
	suffix := strings.TrimPrefix(path, rootType+".")
	if suffix == path || suffix == "" {
		return ""
	}
	// Remove [x] suffix
	suffix = strings.TrimSuffix(suffix, "[x]")
	return suffix
}

// toGoFieldName converts a FHIR field name to a Go field name.
func toGoFieldName(name string) string {
	// Handle special cases
	switch name {
	case "class":
		return "Class"
	case "import":
		return "Import"
	case "type":
		return "Type"
	case "package":
		return "Package"
	case "interface":
		return "Interface"
	}

	// Convert to PascalCase
	return toPascalCase(name)
}

// toPascalCase converts a string to PascalCase.
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// toLowerFirst returns the string with the first character lowercased.
func toLowerFirst(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
