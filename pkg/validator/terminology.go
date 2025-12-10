// Package validator provides FHIR resource validation based on StructureDefinitions.
package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// LocalTerminologyService provides terminology validation using locally loaded ValueSets
// and CodeSystems from FHIR specification bundles.
//
// This implementation:
// - Loads ValueSets and CodeSystems from specs/{version}/valuesets.json
// - Supports required, extensible, preferred, and example bindings
// - Resolves ValueSets that reference CodeSystems (most common pattern)
// - Handles versioned ValueSet URLs (e.g., http://hl7.org/fhir/ValueSet/address-use|4.0.1)
//
// Example usage:
//
//	termService := NewLocalTerminologyService()
//	err := termService.LoadFromFile("specs/r4/valuesets.json")
//	validator := NewValidator(registry, opts).WithTerminologyService(termService)
type LocalTerminologyService struct {
	mu sync.RWMutex

	// codeSystems maps CodeSystem URL to its codes (system URL -> code -> CodeInfo)
	codeSystems map[string]map[string]*CodeInfo

	// valueSets maps ValueSet URL to its expanded codes (valueSet URL -> []CodeInfo)
	valueSets map[string][]*CodeInfo

	// valueSetIndex maps ValueSet URL to the systems it includes
	// Used for ValidateCode when only system+code provided without valueSet
	valueSetSystems map[string][]string
}

// NewLocalTerminologyService creates a new local terminology service.
func NewLocalTerminologyService() *LocalTerminologyService {
	return &LocalTerminologyService{
		codeSystems:     make(map[string]map[string]*CodeInfo),
		valueSets:       make(map[string][]*CodeInfo),
		valueSetSystems: make(map[string][]string),
	}
}

// LoadFromFile loads ValueSets and CodeSystems from a FHIR bundle file.
func (s *LocalTerminologyService) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return s.LoadFromBundle(data)
}

// LoadFromDirectory loads all valuesets.json files from a specs directory.
// Expects structure: specsDir/{version}/valuesets.json
func (s *LocalTerminologyService) LoadFromDirectory(specsDir string) error {
	versions := []string{"r4", "r4b", "r5"}
	for _, version := range versions {
		path := filepath.Join(specsDir, version, "valuesets.json")
		if _, err := os.Stat(path); err == nil {
			if err := s.LoadFromFile(path); err != nil {
				return fmt.Errorf("failed to load %s valuesets: %w", version, err)
			}
		}
	}
	return nil
}

// LoadFromBundle loads ValueSets and CodeSystems from a FHIR Bundle JSON.
func (s *LocalTerminologyService) LoadFromBundle(data []byte) error {
	var bundle struct {
		ResourceType string `json:"resourceType"`
		Entry        []struct {
			Resource json.RawMessage `json:"resource"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(data, &bundle); err != nil {
		return fmt.Errorf("failed to parse bundle: %w", err)
	}

	if bundle.ResourceType != "Bundle" {
		return fmt.Errorf("expected Bundle, got %s", bundle.ResourceType)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// First pass: load all CodeSystems
	for _, entry := range bundle.Entry {
		if entry.Resource == nil {
			continue
		}

		var base struct {
			ResourceType string `json:"resourceType"`
		}
		if err := json.Unmarshal(entry.Resource, &base); err != nil {
			continue
		}

		if base.ResourceType == "CodeSystem" {
			if err := s.loadCodeSystem(entry.Resource); err != nil {
				// Log but continue loading other resources
				continue
			}
		}
	}

	// Second pass: load ValueSets and resolve CodeSystem references
	for _, entry := range bundle.Entry {
		if entry.Resource == nil {
			continue
		}

		var base struct {
			ResourceType string `json:"resourceType"`
		}
		if err := json.Unmarshal(entry.Resource, &base); err != nil {
			continue
		}

		if base.ResourceType == "ValueSet" {
			if err := s.loadValueSet(entry.Resource); err != nil {
				// Log but continue loading other resources
				continue
			}
		}
	}

	return nil
}

// codeSystemResource represents a FHIR CodeSystem for parsing.
type codeSystemResource struct {
	ResourceType string              `json:"resourceType"`
	URL          string              `json:"url"`
	Name         string              `json:"name"`
	Status       string              `json:"status"`
	Content      string              `json:"content"`
	Concept      []codeSystemConcept `json:"concept,omitempty"`
}

type codeSystemConcept struct {
	Code       string              `json:"code"`
	Display    string              `json:"display,omitempty"`
	Definition string              `json:"definition,omitempty"`
	Concept    []codeSystemConcept `json:"concept,omitempty"` // Nested concepts
}

// loadCodeSystem parses and stores a CodeSystem.
func (s *LocalTerminologyService) loadCodeSystem(data []byte) error {
	var cs codeSystemResource
	if err := json.Unmarshal(data, &cs); err != nil {
		return err
	}

	if cs.URL == "" {
		return nil // Skip CodeSystems without URL
	}

	// Only load CodeSystems with actual content
	if cs.Content != "complete" && cs.Content != "fragment" {
		// "not-present" or "example" - codes are not in the resource
		return nil
	}

	codes := make(map[string]*CodeInfo)
	s.flattenConcepts(cs.URL, cs.Concept, codes)

	if len(codes) > 0 {
		s.codeSystems[cs.URL] = codes
	}

	return nil
}

// flattenConcepts recursively flattens nested concepts into a map.
func (s *LocalTerminologyService) flattenConcepts(system string, concepts []codeSystemConcept, codes map[string]*CodeInfo) {
	for _, c := range concepts {
		codes[c.Code] = &CodeInfo{
			System:  system,
			Code:    c.Code,
			Display: c.Display,
			Active:  true,
		}
		// Recursively add nested concepts
		if len(c.Concept) > 0 {
			s.flattenConcepts(system, c.Concept, codes)
		}
	}
}

// valueSetResource represents a FHIR ValueSet for parsing.
type valueSetResource struct {
	ResourceType string             `json:"resourceType"`
	URL          string             `json:"url"`
	Name         string             `json:"name"`
	Status       string             `json:"status"`
	Compose      *valueSetCompose   `json:"compose,omitempty"`
	Expansion    *valueSetExpansion `json:"expansion,omitempty"`
}

type valueSetCompose struct {
	Include []valueSetInclude `json:"include,omitempty"`
	Exclude []valueSetInclude `json:"exclude,omitempty"`
}

type valueSetInclude struct {
	System  string            `json:"system,omitempty"`
	Version string            `json:"version,omitempty"`
	Concept []valueSetConcept `json:"concept,omitempty"`
	Filter  []valueSetFilter  `json:"filter,omitempty"`
}

type valueSetConcept struct {
	Code    string `json:"code"`
	Display string `json:"display,omitempty"`
}

type valueSetFilter struct {
	Property string `json:"property"`
	Op       string `json:"op"`
	Value    string `json:"value"`
}

type valueSetExpansion struct {
	Contains []expansionContains `json:"contains,omitempty"`
}

type expansionContains struct {
	System  string `json:"system,omitempty"`
	Code    string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
}

// loadValueSet parses and stores a ValueSet with its expanded codes.
func (s *LocalTerminologyService) loadValueSet(data []byte) error {
	var vs valueSetResource
	if err := json.Unmarshal(data, &vs); err != nil {
		return err
	}

	if vs.URL == "" {
		return nil // Skip ValueSets without URL
	}

	var codes []*CodeInfo
	var systems []string

	// First, try to use pre-computed expansion (most efficient)
	if vs.Expansion != nil && len(vs.Expansion.Contains) > 0 {
		codes = s.expandFromExpansion(vs.Expansion)
	} else if vs.Compose != nil {
		// Otherwise, expand from compose
		codes, systems = s.expandFromCompose(vs.Compose)
	}

	if len(codes) > 0 {
		s.valueSets[vs.URL] = codes
		if len(systems) > 0 {
			s.valueSetSystems[vs.URL] = systems
		}
	}

	return nil
}

// expandFromExpansion extracts codes from a pre-computed ValueSet expansion.
func (s *LocalTerminologyService) expandFromExpansion(expansion *valueSetExpansion) []*CodeInfo {
	codes := make([]*CodeInfo, 0, len(expansion.Contains))
	for _, c := range expansion.Contains {
		codes = append(codes, &CodeInfo{
			System:  c.System,
			Code:    c.Code,
			Display: c.Display,
			Active:  true,
		})
	}
	return codes
}

// expandFromCompose expands codes from ValueSet.compose definition.
func (s *LocalTerminologyService) expandFromCompose(compose *valueSetCompose) (codes []*CodeInfo, systems []string) {
	systemSet := make(map[string]bool)

	for _, include := range compose.Include {
		if include.System == "" {
			continue
		}

		systemSet[include.System] = true
		codes = append(codes, s.expandInclude(include)...)
	}

	// Convert system set to slice
	systems = make([]string, 0, len(systemSet))
	for system := range systemSet {
		systems = append(systems, system)
	}

	return codes, systems
}

// expandInclude expands a single include clause from ValueSet.compose.
func (s *LocalTerminologyService) expandInclude(include valueSetInclude) []*CodeInfo {
	// If explicit concepts are listed, use them
	if len(include.Concept) > 0 {
		codes := make([]*CodeInfo, 0, len(include.Concept))
		for _, c := range include.Concept {
			codes = append(codes, &CodeInfo{
				System:  include.System,
				Code:    c.Code,
				Display: c.Display,
				Active:  true,
			})
		}
		return codes
	}

	// Try to get codes from CodeSystem
	csCodes, ok := s.codeSystems[include.System]
	if !ok {
		return nil
	}

	// If no filters, include all codes from CodeSystem
	if len(include.Filter) == 0 {
		codes := make([]*CodeInfo, 0, len(csCodes))
		for _, code := range csCodes {
			codes = append(codes, code)
		}
		return codes
	}

	// Apply filters
	return s.applyFilters(csCodes, include.Filter)
}

// applyFilters applies ValueSet filters to CodeSystem codes.
// This is a simplified implementation supporting common filters.
func (s *LocalTerminologyService) applyFilters(codes map[string]*CodeInfo, filters []valueSetFilter) []*CodeInfo {
	var result []*CodeInfo

	for _, code := range codes {
		include := true
		for _, filter := range filters {
			switch filter.Op {
			case "=":
				// Property equals value (for code property, match the code)
				if filter.Property == "code" && code.Code != filter.Value {
					include = false
				}
			case "in":
				// Code is in a comma-separated list
				if filter.Property == "code" {
					values := strings.Split(filter.Value, ",")
					found := false
					for _, v := range values {
						if strings.TrimSpace(v) == code.Code {
							found = true
							break
						}
					}
					if !found {
						include = false
					}
				}
				// "is-a", "descendent-of", "is-not-a" etc. require hierarchy info
				// which we don't track - include all codes for now
			}
		}
		if include {
			result = append(result, code)
		}
	}

	return result
}

// ValidateCode checks if a code is valid in the given ValueSet.
// Implements TerminologyService.ValidateCode.
func (s *LocalTerminologyService) ValidateCode(_ context.Context, system, code, valueSetURL string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Normalize ValueSet URL (remove version suffix)
	vsURL := normalizeValueSetURL(valueSetURL)

	// Look up ValueSet
	codes, ok := s.valueSets[vsURL]
	if !ok {
		// ValueSet not found - cannot validate
		return false, fmt.Errorf("ValueSet not found: %s", valueSetURL)
	}

	// Search for the code
	for _, c := range codes {
		// If system is provided, must match
		if system != "" && c.System != system {
			continue
		}
		if c.Code == code {
			return true, nil
		}
	}

	return false, nil
}

// ExpandValueSet returns all codes in the ValueSet.
// Implements TerminologyService.ExpandValueSet.
func (s *LocalTerminologyService) ExpandValueSet(_ context.Context, valueSetURL string) ([]CodeInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vsURL := normalizeValueSetURL(valueSetURL)

	codes, ok := s.valueSets[vsURL]
	if !ok {
		return nil, fmt.Errorf("ValueSet not found: %s", valueSetURL)
	}

	result := make([]CodeInfo, len(codes))
	for i, c := range codes {
		result[i] = *c
	}

	return result, nil
}

// LookupCode returns information about a specific code from a CodeSystem.
// Implements TerminologyService.LookupCode.
func (s *LocalTerminologyService) LookupCode(_ context.Context, system, code string) (*CodeInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	codes, ok := s.codeSystems[system]
	if !ok {
		return nil, fmt.Errorf("CodeSystem not found: %s", system)
	}

	codeInfo, ok := codes[code]
	if !ok {
		return nil, nil // Code not found in system
	}

	// Return a copy
	return &CodeInfo{
		System:  codeInfo.System,
		Code:    codeInfo.Code,
		Display: codeInfo.Display,
		Active:  codeInfo.Active,
	}, nil
}

// Stats returns statistics about loaded terminology resources.
func (s *LocalTerminologyService) Stats() (codeSystems, valueSets, totalCodes int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	codeSystems = len(s.codeSystems)
	valueSets = len(s.valueSets)

	for _, codes := range s.codeSystems {
		totalCodes += len(codes)
	}

	return
}

// HasValueSet returns true if the ValueSet is loaded.
func (s *LocalTerminologyService) HasValueSet(url string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.valueSets[normalizeValueSetURL(url)]
	return ok
}

// HasCodeSystem returns true if the CodeSystem is loaded.
func (s *LocalTerminologyService) HasCodeSystem(url string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.codeSystems[url]
	return ok
}

// normalizeValueSetURL removes version suffix from ValueSet URL.
// E.g., "http://hl7.org/fhir/ValueSet/address-use|4.0.1" -> "http://hl7.org/fhir/ValueSet/address-use"
func normalizeValueSetURL(url string) string {
	if idx := strings.Index(url, "|"); idx != -1 {
		return url[:idx]
	}
	return url
}
