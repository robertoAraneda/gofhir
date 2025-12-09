// Package validator provides FHIR resource validation based on StructureDefinitions.
// It supports validation against base FHIR specs and custom Implementation Guides.
package validator

import (
	"context"
)

// ReferenceResolver allows resolving external references.
// Useful for tests and FHIR server implementations.
type ReferenceResolver interface {
	// Resolve resolves a reference string to a resource.
	// Returns nil, nil if the reference cannot be resolved (not an error).
	Resolve(ctx context.Context, reference string) (interface{}, error)
}

// TerminologyService allows validating codes against ValueSets and CodeSystems.
// Implementations: LocalTerminologyService, RemoteTerminologyService (tx.fhir.org)
type TerminologyService interface {
	// ValidateCode checks if a code is valid in the given ValueSet.
	ValidateCode(ctx context.Context, system, code, valueSetURL string) (bool, error)

	// ExpandValueSet returns all codes in the ValueSet.
	ExpandValueSet(ctx context.Context, valueSetURL string) ([]CodeInfo, error)

	// LookupCode returns information about a specific code.
	LookupCode(ctx context.Context, system, code string) (*CodeInfo, error)
}

// CodeInfo contains information about a terminology code.
type CodeInfo struct {
	System  string `json:"system"`
	Code    string `json:"code"`
	Display string `json:"display,omitempty"`
	Active  bool   `json:"active"`
}

// StructureDefinitionProvider allows loading StructureDefinitions from different sources.
// Uses internal ElementDef model to support all FHIR versions (R4, R4B, R5).
type StructureDefinitionProvider interface {
	// Get returns a StructureDefinition by URL.
	Get(ctx context.Context, url string) (*StructureDef, error)

	// GetByType returns the base StructureDefinition for a resource type.
	GetByType(ctx context.Context, resourceType string) (*StructureDef, error)

	// List returns all available StructureDefinition URLs.
	List(ctx context.Context) ([]string, error)
}

// NoopReferenceResolver does not resolve any references (for local validation).
type NoopReferenceResolver struct{}

// Resolve always returns nil, nil.
func (n *NoopReferenceResolver) Resolve(ctx context.Context, ref string) (interface{}, error) {
	return nil, nil
}

// NoopTerminologyService does not validate terminology (skips validation).
type NoopTerminologyService struct{}

// ValidateCode always returns true (skip validation).
func (n *NoopTerminologyService) ValidateCode(ctx context.Context, system, code, valueSetURL string) (bool, error) {
	return true, nil
}

// ExpandValueSet returns empty (no expansion available).
func (n *NoopTerminologyService) ExpandValueSet(ctx context.Context, valueSetURL string) ([]CodeInfo, error) {
	return nil, nil
}

// LookupCode returns nil (no lookup available).
func (n *NoopTerminologyService) LookupCode(ctx context.Context, system, code string) (*CodeInfo, error) {
	return nil, nil
}
