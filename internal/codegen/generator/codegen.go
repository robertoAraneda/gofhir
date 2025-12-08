// Package generator implements FHIR to Go code generation.
package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/robertoaraneda/gofhir/internal/codegen/analyzer"
	"github.com/robertoaraneda/gofhir/internal/codegen/parser"
)

// Config holds code generation configuration.
type Config struct {
	// SpecsDir is the directory containing FHIR specifications
	SpecsDir string
	// OutputDir is the directory to write generated code
	OutputDir string
	// PackageName is the Go package name for generated code
	PackageName string
	// Version is the FHIR version (r4, r4b, r5)
	Version string
}

// CodeGen generates Go code from FHIR specifications.
type CodeGen struct {
	config       Config
	analyzer     *analyzer.Analyzer
	types        []*analyzer.AnalyzedType
	valueSets    *parser.ValueSetRegistry
	usedBindings map[string]bool // Track which bindings are actually used
}

// New creates a new CodeGen instance.
func New(config Config) *CodeGen {
	return &CodeGen{
		config:       config,
		types:        make([]*analyzer.AnalyzedType, 0),
		valueSets:    parser.NewValueSetRegistry(),
		usedBindings: make(map[string]bool),
	}
}

// LoadTypes loads and analyzes all StructureDefinitions from the specs directory.
func (c *CodeGen) LoadTypes() error {
	specsDir := filepath.Join(c.config.SpecsDir, c.config.Version)

	// Load ValueSets first (needed for binding resolution)
	valueSetsFile := filepath.Join(specsDir, "valuesets.json")
	if data, err := os.ReadFile(valueSetsFile); err == nil {
		if err := c.valueSets.LoadFromBundle(data); err != nil {
			// Non-fatal: continue without value sets
			fmt.Printf("Warning: failed to load value sets: %v\n", err)
		}
	}

	// Collect all StructureDefinitions from both bundles
	var allSDs []*parser.StructureDefinition

	// Load datatypes from profiles-types.json
	typesFile := filepath.Join(specsDir, "profiles-types.json")
	typeSDs, err := c.loadStructureDefinitions(typesFile)
	if err != nil {
		return fmt.Errorf("failed to load types: %w", err)
	}
	allSDs = append(allSDs, typeSDs...)

	// Load resources from profiles-resources.json
	resourcesFile := filepath.Join(specsDir, "profiles-resources.json")
	resourceSDs, err := c.loadStructureDefinitions(resourcesFile)
	if err != nil {
		return fmt.Errorf("failed to load resources: %w", err)
	}
	allSDs = append(allSDs, resourceSDs...)

	// Create ONE analyzer with ALL definitions and value sets
	c.analyzer = analyzer.NewAnalyzer(allSDs, c.valueSets)

	// Analyze each StructureDefinition
	for _, sd := range allSDs {
		analyzed, err := c.analyzer.Analyze(sd)
		if err != nil {
			// Skip types that fail analysis (e.g., incomplete definitions)
			continue
		}
		c.types = append(c.types, analyzed)
	}

	return nil
}

// loadStructureDefinitions loads and filters StructureDefinitions from a Bundle file.
func (c *CodeGen) loadStructureDefinitions(path string) ([]*parser.StructureDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	bundle, err := parser.ParseBundle(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bundle: %w", err)
	}

	sds, err := parser.ExtractStructureDefinitions(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to extract definitions: %w", err)
	}

	// Filter StructureDefinitions:
	// - Keep abstract base types: Element, BackboneElement
	// - Skip primitive types (they map to Go builtins)
	// - Skip other abstract types
	var filtered []*parser.StructureDefinition
	for _, sd := range sds {
		// Keep base types even if abstract
		if sd.Name == "Element" || sd.Name == "BackboneElement" {
			filtered = append(filtered, sd)
			continue
		}

		// Skip primitive types - they map to Go builtins
		if sd.Kind == parser.KindPrimitiveType {
			continue
		}

		// Filter out other abstract types
		if !sd.Abstract {
			filtered = append(filtered, sd)
		}
	}

	return filtered, nil
}

// Generate writes all generated code to the output directory.
func (c *CodeGen) Generate() error {
	if err := os.MkdirAll(c.config.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate codesystems.go first (types used by datatypes and resources)
	if err := c.generateCodeSystems(); err != nil {
		return fmt.Errorf("failed to generate codesystems: %w", err)
	}

	// Generate datatypes.go
	if err := c.generateDatatypes(); err != nil {
		return fmt.Errorf("failed to generate datatypes: %w", err)
	}

	// Generate resources.go
	if err := c.generateResources(); err != nil {
		return fmt.Errorf("failed to generate resources: %w", err)
	}

	// Generate interfaces.go (manual - Resource interface)
	if err := c.generateInterfaces(); err != nil {
		return fmt.Errorf("failed to generate interfaces: %w", err)
	}

	return nil
}

// generateDatatypes generates datatypes.go
func (c *CodeGen) generateDatatypes() error {
	var datatypes []*analyzer.AnalyzedType
	for _, t := range c.types {
		if t.Kind == "datatype" || t.Kind == "primitive" || t.Kind == "backbone" {
			datatypes = append(datatypes, t)
		}
	}

	// Sort alphabetically
	sort.Slice(datatypes, func(i, j int) bool {
		return datatypes[i].Name < datatypes[j].Name
	})

	path := filepath.Join(c.config.OutputDir, "datatypes.go")
	return c.writeTypesFile(path, datatypes, "datatypes")
}

// generateResources generates resources.go
func (c *CodeGen) generateResources() error {
	var resources []*analyzer.AnalyzedType
	for _, t := range c.types {
		if t.Kind == "resource" {
			resources = append(resources, t)
		}
	}

	// Sort alphabetically
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})

	path := filepath.Join(c.config.OutputDir, "resources.go")
	return c.writeTypesFile(path, resources, "resources")
}

// generateCodeSystems generates codesystems.go with enum types for required bindings.
func (c *CodeGen) generateCodeSystems() error {
	if c.analyzer == nil || len(c.analyzer.UsedBindings) == 0 {
		return nil // No code systems to generate
	}

	var buf bytes.Buffer

	// Write header
	fmt.Fprintf(&buf, "// Code generated by gofhir. DO NOT EDIT.\n")
	fmt.Fprintf(&buf, "// Source: FHIR ValueSets (codesystems)\n")
	fmt.Fprintf(&buf, "// Package: %s\n\n", c.config.PackageName)
	fmt.Fprintf(&buf, "package %s\n\n", c.config.PackageName)

	// Collect and sort used value sets
	valueSetURLs := make([]string, 0, len(c.analyzer.UsedBindings))
	for url := range c.analyzer.UsedBindings {
		valueSetURLs = append(valueSetURLs, url)
	}
	sort.Strings(valueSetURLs)

	// Track generated type names to avoid duplicates
	generatedTypes := make(map[string]bool)

	// Generate types for each used value set
	for _, url := range valueSetURLs {
		vs := c.valueSets.Get(url)
		if vs == nil {
			continue
		}

		// Skip if we already generated a type with this name
		typeName := sanitizeTypeName(vs.Name)
		if generatedTypes[typeName] {
			continue
		}
		generatedTypes[typeName] = true

		c.writeCodeSystemType(&buf, vs)
	}

	// Format code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		path := filepath.Join(c.config.OutputDir, "codesystems.go.unformatted")
		if writeErr := os.WriteFile(path, buf.Bytes(), 0o600); writeErr != nil {
			return fmt.Errorf("format error: %w (also failed to write debug file: %v)", err, writeErr)
		}
		return fmt.Errorf("failed to format codesystems (saved to %s): %w", path, err)
	}

	path := filepath.Join(c.config.OutputDir, "codesystems.go")
	return os.WriteFile(path, formatted, 0o600)
}

// writeCodeSystemType writes a single code system type with constants.
func (c *CodeGen) writeCodeSystemType(w io.Writer, vs *parser.ParsedValueSet) {
	// Sanitize type name to be a valid Go identifier
	typeName := sanitizeTypeName(vs.Name)

	// Write type definition
	if vs.Title != "" {
		fmt.Fprintf(w, "// %s represents %s.\n", typeName, vs.Title)
	} else {
		fmt.Fprintf(w, "// %s represents allowed values for the %s code system.\n", typeName, vs.Name)
	}
	fmt.Fprintf(w, "type %s string\n\n", typeName)

	// Write constants
	fmt.Fprintf(w, "// %s values.\n", typeName)
	fmt.Fprintf(w, "const (\n")
	for _, code := range vs.Codes {
		constName := typeName + toPascalCaseCode(code.Code)
		if code.Display != "" {
			fmt.Fprintf(w, "\t// %s - %s\n", constName, code.Display)
		}
		fmt.Fprintf(w, "\t%s %s = %q\n", constName, typeName, code.Code)
	}
	fmt.Fprintf(w, ")\n\n")
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
		r, _ := utf8.DecodeRuneInString(strings.ToUpper(string(runes[0])))
		runes[0] = r
		name = string(runes)
	}

	return name
}

// toPascalCaseCode converts a code value to PascalCase for use as a constant name.
func toPascalCaseCode(code string) string {
	// Handle special symbol codes first
	symbolMap := map[string]string{
		"<":  "LessThan",
		"<=": "LessOrEqual",
		">":  "GreaterThan",
		">=": "GreaterOrEqual",
		"=":  "Equal",
		"!=": "NotEqual",
		"+":  "Plus",
		"-":  "Minus",
		"*":  "Asterisk",
		"/":  "Slash",
		"#":  "Hash",
		"&":  "Ampersand",
		"|":  "Pipe",
	}

	if replacement, ok := symbolMap[code]; ok {
		return replacement
	}

	// Handle common patterns
	code = strings.ReplaceAll(code, "-", " ")
	code = strings.ReplaceAll(code, "_", " ")
	code = strings.ReplaceAll(code, ".", " ")
	code = strings.ReplaceAll(code, "/", " ")
	code = strings.ReplaceAll(code, "(", " ")
	code = strings.ReplaceAll(code, ")", " ")

	words := strings.Fields(code)
	for i, word := range words {
		if word != "" {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}

// generateInterfaces generates interfaces.go
func (c *CodeGen) generateInterfaces() error {
	path := filepath.Join(c.config.OutputDir, "interfaces.go")

	content := fmt.Sprintf(`// Code generated by gofhir. DO NOT EDIT.
// Package %s contains FHIR %s types.

package %s

// Resource is the interface implemented by all FHIR resources.
type Resource interface {
	GetResourceType() string
}
`, c.config.PackageName, strings.ToUpper(c.config.Version), c.config.PackageName)

	return os.WriteFile(path, []byte(content), 0o600)
}

// writeTypesFile writes a Go file with the given types.
func (c *CodeGen) writeTypesFile(path string, types []*analyzer.AnalyzedType, fileType string) error {
	var buf bytes.Buffer

	// Write header
	fmt.Fprintf(&buf, "// Code generated by gofhir. DO NOT EDIT.\n")
	fmt.Fprintf(&buf, "// Source: FHIR StructureDefinitions (%s)\n", fileType)
	fmt.Fprintf(&buf, "// Package: %s\n\n", c.config.PackageName)
	fmt.Fprintf(&buf, "package %s\n\n", c.config.PackageName)

	// Write each type
	for _, t := range types {
		c.writeType(&buf, t)
	}

	// Format code
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Write unformatted for debugging
		if writeErr := os.WriteFile(path+".unformatted", buf.Bytes(), 0o600); writeErr != nil {
			return fmt.Errorf("format error: %w (also failed to write debug file: %v)", err, writeErr)
		}
		return fmt.Errorf("failed to format code (saved to %s.unformatted): %w", path, err)
	}

	return os.WriteFile(path, formatted, 0o600)
}

// writeType writes a single struct type.
func (c *CodeGen) writeType(w io.Writer, t *analyzer.AnalyzedType) {
	// Write doc comment
	if t.Description != "" {
		fmt.Fprintf(w, "// %s %s\n", t.Name, t.Description)
	} else {
		fmt.Fprintf(w, "// %s represents FHIR %s.\n", t.Name, t.FHIRName)
	}

	fmt.Fprintf(w, "type %s struct {\n", t.Name)

	for _, prop := range t.Properties {
		// Write field comment
		if prop.Description != "" {
			fmt.Fprintf(w, "\t// %s\n", prop.Description)
		}

		// Write field
		jsonTag := c.jsonTag(prop)
		fmt.Fprintf(w, "\t%s %s %s\n", prop.Name, prop.GoType, jsonTag)

		// Add extension field for primitives (except for choice types which already have them)
		if prop.HasExtension && !prop.IsChoice {
			extName := prop.Name + "Ext"
			extJSONName := "_" + prop.JSONName
			if prop.IsArray {
				fmt.Fprintf(w, "\t// Extension for %s\n", prop.Name)
				fmt.Fprintf(w, "\t%s []Element `json:%q`\n", extName, extJSONName+",omitempty")
			} else {
				fmt.Fprintf(w, "\t// Extension for %s\n", prop.Name)
				fmt.Fprintf(w, "\t%s *Element `json:%q`\n", extName, extJSONName+",omitempty")
			}
		}
	}

	fmt.Fprintf(w, "}\n\n")

	// Add GetResourceType for resources
	if t.Kind == "resource" {
		fmt.Fprintf(w, "// GetResourceType returns the FHIR resource type.\n")
		fmt.Fprintf(w, "func (r *%s) GetResourceType() string {\n", t.Name)
		fmt.Fprintf(w, "\treturn %q\n", t.FHIRName)
		fmt.Fprintf(w, "}\n\n")
	}
}

// jsonTag generates the JSON struct tag for a property.
func (c *CodeGen) jsonTag(prop analyzer.AnalyzedProperty) string {
	if prop.IsArray || prop.IsPointer {
		return fmt.Sprintf("`json:%s`", fmt.Sprintf("%q", prop.JSONName+",omitempty"))
	}
	return fmt.Sprintf("`json:%s`", fmt.Sprintf("%q", prop.JSONName))
}
