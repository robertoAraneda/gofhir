// Package validator provides FHIR resource validation based on StructureDefinitions.
package validator

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FHIRVersion represents a FHIR specification version.
type FHIRVersion string

const (
	FHIRVersionR4  FHIRVersion = "R4"
	FHIRVersionR4B FHIRVersion = "R4B"
	FHIRVersionR5  FHIRVersion = "R5"

	// resourceTypeStructureDefinition is the FHIR resource type for StructureDefinition.
	resourceTypeStructureDefinition = "StructureDefinition"
)

// Registry is a StructureDefinitionProvider that loads definitions from
// embedded specs or external files. Thread-safe for concurrent access.
type Registry struct {
	mu sync.RWMutex
	// byURL maps canonical URL to StructureDef
	byURL map[string]*StructureDef
	// byType maps resource type to base StructureDef
	byType map[string]*StructureDef
	// version is the FHIR version for this registry
	version FHIRVersion
}

// NewRegistry creates a new empty registry.
func NewRegistry(version FHIRVersion) *Registry {
	return &Registry{
		byURL:   make(map[string]*StructureDef),
		byType:  make(map[string]*StructureDef),
		version: version,
	}
}

// Get returns a StructureDefinition by canonical URL.
func (r *Registry) Get(ctx context.Context, url string) (*StructureDef, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if sd, ok := r.byURL[url]; ok {
		return sd, nil
	}
	return nil, fmt.Errorf("StructureDefinition not found: %s", url)
}

// GetByType returns the base StructureDefinition for a resource type.
func (r *Registry) GetByType(ctx context.Context, resourceType string) (*StructureDef, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if sd, ok := r.byType[resourceType]; ok {
		return sd, nil
	}
	return nil, fmt.Errorf("StructureDefinition not found for type: %s", resourceType)
}

// List returns all available StructureDefinition URLs.
func (r *Registry) List(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	urls := make([]string, 0, len(r.byURL))
	for url := range r.byURL {
		urls = append(urls, url)
	}
	return urls, nil
}

// Register adds a StructureDefinition to the registry.
func (r *Registry) Register(sd *StructureDef) error {
	if sd == nil {
		return fmt.Errorf("cannot register nil StructureDefinition")
	}
	if sd.URL == "" {
		return fmt.Errorf("StructureDefinition must have a URL")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.byURL[sd.URL] = sd

	// Also index by type for base definitions (non-profiles)
	if sd.Type != "" && sd.Kind == "resource" && !strings.Contains(sd.URL, "/profile/") {
		// Only register as base type if not already registered or if this is the canonical URL
		if existing, ok := r.byType[sd.Type]; !ok || isCanonicalURL(sd.URL, sd.Type) {
			if existing == nil || isCanonicalURL(sd.URL, sd.Type) {
				r.byType[sd.Type] = sd
			}
		}
	}

	return nil
}

// isCanonicalURL checks if URL is the canonical HL7 FHIR URL for a type
func isCanonicalURL(url, resourceType string) bool {
	canonical := "http://hl7.org/fhir/StructureDefinition/" + resourceType
	return url == canonical
}

// Size returns the number of registered StructureDefinitions.
func (r *Registry) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byURL)
}

// LoadFromBundle loads StructureDefinitions from a FHIR Bundle JSON.
// This is the format used in profiles-resources.json, profiles-types.json, etc.
func (r *Registry) LoadFromBundle(data []byte) (int, error) {
	var bundle struct {
		Entry []struct {
			Resource json.RawMessage `json:"resource"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(data, &bundle); err != nil {
		return 0, fmt.Errorf("failed to parse bundle: %w", err)
	}

	count := 0
	for _, entry := range bundle.Entry {
		// Check if this is a StructureDefinition
		var resourceType struct {
			ResourceType string `json:"resourceType"`
		}
		if err := json.Unmarshal(entry.Resource, &resourceType); err != nil {
			continue
		}
		if resourceType.ResourceType != resourceTypeStructureDefinition {
			continue
		}

		sd, err := ParseStructureDefinition(entry.Resource)
		if err != nil {
			continue // Skip invalid entries
		}

		if err := r.Register(sd); err != nil {
			continue
		}
		count++
	}

	return count, nil
}

// LoadFromFile loads StructureDefinitions from a JSON file.
// Supports both single StructureDefinition and Bundle formats.
func (r *Registry) LoadFromFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	return r.LoadFromJSON(data)
}

// LoadFromJSON loads StructureDefinitions from JSON data.
// Auto-detects Bundle vs single StructureDefinition format.
func (r *Registry) LoadFromJSON(data []byte) (int, error) {
	// Try to detect format
	var probe struct {
		ResourceType string `json:"resourceType"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return 0, fmt.Errorf("invalid JSON: %w", err)
	}

	switch probe.ResourceType {
	case "Bundle":
		return r.LoadFromBundle(data)
	case resourceTypeStructureDefinition:
		sd, err := ParseStructureDefinition(data)
		if err != nil {
			return 0, err
		}
		if err := r.Register(sd); err != nil {
			return 0, err
		}
		return 1, nil
	default:
		return 0, fmt.Errorf("unsupported resourceType: %s", probe.ResourceType)
	}
}

// LoadFromDirectory loads all JSON files from a directory.
func (r *Registry) LoadFromDirectory(dirPath string) (int, error) {
	total := 0
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		count, err := r.LoadFromFile(path)
		if err != nil {
			// Log but continue
			return nil
		}
		total += count
		return nil
	})

	return total, err
}

// LoadFromFS loads StructureDefinitions from an embedded filesystem.
func (r *Registry) LoadFromFS(fsys embed.FS, root string) (int, error) {
	total := 0
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := fsys.ReadFile(path)
		if err != nil {
			return nil // Skip files we can't read
		}

		count, err := r.LoadFromJSON(data)
		if err != nil {
			return nil // Skip invalid files
		}
		total += count
		return nil
	})

	return total, err
}

// ParseStructureDefinition parses a single StructureDefinition from JSON.
// Works with any FHIR version (R4, R4B, R5) by extracting common fields.
func ParseStructureDefinition(data []byte) (*StructureDef, error) {
	// Use a generic map to handle version differences
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse StructureDefinition: %w", err)
	}

	// Verify it's a StructureDefinition
	if rt, _ := raw["resourceType"].(string); rt != resourceTypeStructureDefinition {
		return nil, fmt.Errorf("not a StructureDefinition: %s", rt)
	}

	sd := &StructureDef{}

	// Extract basic fields
	sd.URL, _ = raw["url"].(string)
	sd.Name, _ = raw["name"].(string)
	sd.Type, _ = raw["type"].(string)
	sd.Kind, _ = raw["kind"].(string)
	sd.Abstract, _ = raw["abstract"].(bool)
	sd.BaseDefinition, _ = raw["baseDefinition"].(string)
	sd.FHIRVersion, _ = raw["fhirVersion"].(string)

	// Parse snapshot elements
	if snapshot, ok := raw["snapshot"].(map[string]interface{}); ok {
		if elements, ok := snapshot["element"].([]interface{}); ok {
			sd.Snapshot = parseElements(elements)
		}
	}

	// Parse differential elements
	if differential, ok := raw["differential"].(map[string]interface{}); ok {
		if elements, ok := differential["element"].([]interface{}); ok {
			sd.Differential = parseElements(elements)
		}
	}

	return sd, nil
}

// parseElements converts raw JSON elements to ElementDef slice.
func parseElements(elements []interface{}) []ElementDef {
	result := make([]ElementDef, 0, len(elements))

	for _, elem := range elements {
		elemMap, ok := elem.(map[string]interface{})
		if !ok {
			continue
		}

		ed := ElementDef{}
		ed.ID, _ = elemMap["id"].(string)
		ed.Path, _ = elemMap["path"].(string)
		ed.SliceName, _ = elemMap["sliceName"].(string)

		if minVal, ok := elemMap["min"].(float64); ok {
			ed.Min = int(minVal)
		}
		ed.Max, _ = elemMap["max"].(string)

		ed.Short, _ = elemMap["short"].(string)
		ed.Definition, _ = elemMap["definition"].(string)
		ed.MustSupport, _ = elemMap["mustSupport"].(bool)
		ed.IsModifier, _ = elemMap["isModifier"].(bool)
		ed.IsSummary, _ = elemMap["isSummary"].(bool)

		// Parse types
		if types, ok := elemMap["type"].([]interface{}); ok {
			ed.Types = parseTypes(types)
		}

		// Parse binding
		if binding, ok := elemMap["binding"].(map[string]interface{}); ok {
			ed.Binding = parseBinding(binding)
		}

		// Parse constraints
		if constraints, ok := elemMap["constraint"].([]interface{}); ok {
			ed.Constraints = parseConstraints(constraints)
		}

		// Handle fixed[x] and pattern[x] values
		for key, val := range elemMap {
			if strings.HasPrefix(key, "fixed") {
				ed.Fixed = val
			}
			if strings.HasPrefix(key, "pattern") {
				ed.Pattern = val
			}
		}

		result = append(result, ed)
	}

	return result
}

// parseTypes converts raw type references to TypeRef slice.
func parseTypes(types []interface{}) []TypeRef {
	result := make([]TypeRef, 0, len(types))

	for _, t := range types {
		typeMap, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		tr := TypeRef{}
		tr.Code, _ = typeMap["code"].(string)

		// Parse targetProfile (for Reference types)
		if targets, ok := typeMap["targetProfile"].([]interface{}); ok {
			for _, target := range targets {
				if s, ok := target.(string); ok {
					tr.TargetProfile = append(tr.TargetProfile, s)
				}
			}
		}

		// Parse profile
		if profiles, ok := typeMap["profile"].([]interface{}); ok {
			for _, profile := range profiles {
				if s, ok := profile.(string); ok {
					tr.Profile = append(tr.Profile, s)
				}
			}
		}

		result = append(result, tr)
	}

	return result
}

// parseBinding converts raw binding to ElementBinding.
func parseBinding(binding map[string]interface{}) *ElementBinding {
	eb := &ElementBinding{}
	eb.Strength, _ = binding["strength"].(string)
	eb.ValueSet, _ = binding["valueSet"].(string)
	eb.Description, _ = binding["description"].(string)
	return eb
}

// LoadR4Specs loads all standard R4 StructureDefinitions from a specs directory.
// This includes profiles-resources.json, profiles-types.json, and extension-definitions.json.
func (r *Registry) LoadR4Specs(specsDir string) (int, error) {
	total := 0

	// Load resource definitions
	resourcesPath := filepath.Join(specsDir, "profiles-resources.json")
	if count, err := r.LoadFromFile(resourcesPath); err == nil {
		total += count
	}

	// Load type definitions
	typesPath := filepath.Join(specsDir, "profiles-types.json")
	if count, err := r.LoadFromFile(typesPath); err == nil {
		total += count
	}

	// Load extension definitions
	extensionsPath := filepath.Join(specsDir, "extension-definitions.json")
	if count, err := r.LoadFromFile(extensionsPath); err == nil {
		total += count
	}

	return total, nil
}

// parseConstraints converts raw constraints to ElementConstraint slice.
func parseConstraints(constraints []interface{}) []ElementConstraint {
	result := make([]ElementConstraint, 0, len(constraints))

	for _, c := range constraints {
		cMap, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		ec := ElementConstraint{}
		ec.Key, _ = cMap["key"].(string)
		ec.Severity, _ = cMap["severity"].(string)
		ec.Human, _ = cMap["human"].(string)
		ec.Expression, _ = cMap["expression"].(string)
		ec.XPath, _ = cMap["xpath"].(string)
		ec.Source, _ = cMap["source"].(string)

		result = append(result, ec)
	}

	return result
}
