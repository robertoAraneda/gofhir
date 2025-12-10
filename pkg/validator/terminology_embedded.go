// Package validator provides FHIR resource validation.
package validator

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// embeddedValueSetRegistry holds all registered embedded ValueSets by FHIR version.
var (
	embeddedValueSetRegistry = make(map[string]map[string]map[string]bool)
	embeddedRegistryMu       sync.RWMutex
)

// registerEmbeddedValueSets registers ValueSets for a FHIR version.
// Called by init() functions in generated terminology_embedded_*.go files.
func registerEmbeddedValueSets(fhirVersion string, valueSets map[string]map[string]bool) {
	embeddedRegistryMu.Lock()
	defer embeddedRegistryMu.Unlock()
	embeddedValueSetRegistry[fhirVersion] = valueSets
}

// EmbeddedTerminologyService provides terminology validation using embedded ValueSets.
// This is more efficient than LocalTerminologyService as it doesn't require file I/O.
type EmbeddedTerminologyService struct {
	fhirVersion string
	valueSets   map[string]map[string]bool
}

// NewEmbeddedTerminologyService creates a new embedded terminology service for the specified FHIR version.
// Supported versions: "4.0.1" (R4), "4.3.0" (R4B), "5.0.0" (R5)
func NewEmbeddedTerminologyService(fhirVersion string) (*EmbeddedTerminologyService, error) {
	embeddedRegistryMu.RLock()
	defer embeddedRegistryMu.RUnlock()

	valueSets, ok := embeddedValueSetRegistry[fhirVersion]
	if !ok {
		available := make([]string, 0, len(embeddedValueSetRegistry))
		for v := range embeddedValueSetRegistry {
			available = append(available, v)
		}
		return nil, fmt.Errorf("no embedded ValueSets for FHIR version %s (available: %v)", fhirVersion, available)
	}

	return &EmbeddedTerminologyService{
		fhirVersion: fhirVersion,
		valueSets:   valueSets,
	}, nil
}

// NewEmbeddedTerminologyServiceR4 creates an embedded terminology service for FHIR R4 (4.0.1).
func NewEmbeddedTerminologyServiceR4() *EmbeddedTerminologyService {
	svc, err := NewEmbeddedTerminologyService("4.0.1")
	if err != nil {
		// This should never happen if the R4 file is generated
		panic(fmt.Sprintf("failed to create R4 embedded terminology service: %v", err))
	}
	return svc
}

// NewEmbeddedTerminologyServiceR4B creates an embedded terminology service for FHIR R4B (4.3.0).
func NewEmbeddedTerminologyServiceR4B() *EmbeddedTerminologyService {
	svc, err := NewEmbeddedTerminologyService("4.3.0")
	if err != nil {
		panic(fmt.Sprintf("failed to create R4B embedded terminology service: %v", err))
	}
	return svc
}

// NewEmbeddedTerminologyServiceR5 creates an embedded terminology service for FHIR R5 (5.0.0).
func NewEmbeddedTerminologyServiceR5() *EmbeddedTerminologyService {
	svc, err := NewEmbeddedTerminologyService("5.0.0")
	if err != nil {
		panic(fmt.Sprintf("failed to create R5 embedded terminology service: %v", err))
	}
	return svc
}

// ValidateCode checks if a code is valid in the given ValueSet.
func (s *EmbeddedTerminologyService) ValidateCode(_ context.Context, _, code, valueSetURL string) (bool, error) {
	vsURL := normalizeEmbeddedURL(valueSetURL)

	codes, ok := s.valueSets[vsURL]
	if !ok {
		return false, fmt.Errorf("ValueSet not found: %s", valueSetURL)
	}

	return codes[code], nil
}

// ExpandValueSet returns all codes in the ValueSet.
func (s *EmbeddedTerminologyService) ExpandValueSet(_ context.Context, valueSetURL string) ([]CodeInfo, error) {
	vsURL := normalizeEmbeddedURL(valueSetURL)

	codes, ok := s.valueSets[vsURL]
	if !ok {
		return nil, fmt.Errorf("ValueSet not found: %s", valueSetURL)
	}

	result := make([]CodeInfo, 0, len(codes))
	for code := range codes {
		result = append(result, CodeInfo{Code: code, Active: true})
	}
	return result, nil
}

// LookupCode returns information about a specific code.
// Note: Embedded service only stores codes, not full CodeInfo with display/system.
func (s *EmbeddedTerminologyService) LookupCode(_ context.Context, _, _ string) (*CodeInfo, error) {
	// Embedded service doesn't track CodeSystems, only ValueSets
	return nil, nil
}

// HasValueSet returns true if the ValueSet is available.
func (s *EmbeddedTerminologyService) HasValueSet(url string) bool {
	_, ok := s.valueSets[normalizeEmbeddedURL(url)]
	return ok
}

// FHIRVersion returns the FHIR version this service is configured for.
func (s *EmbeddedTerminologyService) FHIRVersion() string {
	return s.fhirVersion
}

// Stats returns statistics about embedded terminology.
func (s *EmbeddedTerminologyService) Stats() (valueSets, totalCodes int) {
	valueSets = len(s.valueSets)
	for _, codes := range s.valueSets {
		totalCodes += len(codes)
	}
	return
}

// normalizeEmbeddedURL removes version suffix from ValueSet URL.
func normalizeEmbeddedURL(url string) string {
	if idx := strings.Index(url, "|"); idx != -1 {
		return url[:idx]
	}
	return url
}

// AvailableEmbeddedVersions returns a list of FHIR versions with embedded ValueSets.
func AvailableEmbeddedVersions() []string {
	embeddedRegistryMu.RLock()
	defer embeddedRegistryMu.RUnlock()

	versions := make([]string, 0, len(embeddedValueSetRegistry))
	for v := range embeddedValueSetRegistry {
		versions = append(versions, v)
	}
	return versions
}
