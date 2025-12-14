// Package fhir provides version-agnostic interfaces for FHIR resources.
// This allows the FHIR server to support multiple FHIR versions (R4, R4B, R5, R6)
// through a common abstraction layer.
package fhir

import "fmt"

// Version represents a FHIR specification version.
type Version string

const (
	R4  Version = "R4"
	R4B Version = "R4B"
	R5  Version = "R5"
	R6  Version = "R6"
)

// Resource is the version-agnostic interface for all FHIR resources.
// Per FHIR spec, Resource contains: id, meta, implicitRules, language.
type Resource interface {
	GetResourceType() string
	GetID() *string
	SetID(string)
	GetMeta() Meta
	SetMeta(Meta)
}

// Meta is the version-agnostic interface for FHIR Meta element.
type Meta interface {
	GetVersionID() *string
	SetVersionID(string)
	GetLastUpdated() *string
	SetLastUpdated(string)
}

// SearchEntryMode represents the mode of a search entry (match, include, outcome).
type SearchEntryMode string

const (
	SearchEntryModeMatch   SearchEntryMode = "match"
	SearchEntryModeInclude SearchEntryMode = "include"
	SearchEntryModeOutcome SearchEntryMode = "outcome"
)

// SearchBundleConfig contains configuration for building a search result Bundle.
type SearchBundleConfig struct {
	BaseURL           string       // Base URL for fullUrl generation
	Total             int          // Total number of matching resources
	IncludeTotal      bool         // Whether to include total in the Bundle (default true)
	Resources         [][]byte     // Raw JSON of resources (already with correct field order)
	IncludedResources [][]byte     // Resources included via _include/_revinclude
	Links             []LinkConfig // Bundle links (self, next, previous, etc.)
}

// HistoryBundleConfig contains configuration for building a history Bundle.
type HistoryBundleConfig struct {
	BaseURL  string          // Base URL for fullUrl generation
	Total    int             // Total number of versions
	Versions []VersionConfig // Version entries
	Links    []LinkConfig    // Bundle links (self, next, previous, etc.)
}

// VersionConfig represents a resource version for history bundles.
type VersionConfig struct {
	Resource    []byte // Raw JSON of the resource version
	Method      string // HTTP method (POST, PUT, DELETE)
	LastUpdated string // When this version was created
}

// LinkConfig represents a Bundle link.
type LinkConfig struct {
	Relation string // self, next, previous, first, last
	URL      string // The URL for this link
}

// OutcomeConfig contains configuration for building an OperationOutcome.
type OutcomeConfig struct {
	Severity    string // fatal, error, warning, information
	Code        string // Issue type code
	Diagnostics string // Additional diagnostic information
}

// ResourceFactory creates and deserializes FHIR resources for a specific version.
type ResourceFactory interface {
	// Version returns the FHIR version this factory handles.
	Version() Version

	// NewResource creates an empty instance of the specified resource type.
	NewResource(resourceType string) (Resource, error)

	// UnmarshalResource deserializes JSON to the correct resource type.
	UnmarshalResource(data []byte) (Resource, error)

	// GetResourceType extracts the resourceType from JSON without fully deserializing.
	GetResourceType(data []byte) (string, error)

	// IsKnownResourceType returns true if the resource type is known in this version.
	IsKnownResourceType(resourceType string) bool

	// NewMeta creates a new Meta instance for this version.
	NewMeta() Meta

	// BuildSearchBundle builds a searchset Bundle with correct FHIR field ordering.
	// Resources are passed through as-is (must already have correct field order).
	BuildSearchBundle(cfg SearchBundleConfig) ([]byte, error)

	// BuildHistoryBundle builds a history Bundle with correct FHIR field ordering.
	BuildHistoryBundle(cfg HistoryBundleConfig) ([]byte, error)

	// BuildOperationOutcome builds an OperationOutcome with correct FHIR field ordering.
	BuildOperationOutcome(cfg OutcomeConfig) ([]byte, error)
}

// registry holds registered factories by version.
var registry = make(map[Version]ResourceFactory)

// RegisterFactory registers a ResourceFactory for a specific FHIR version.
// This should be called by each version package's init() function.
func RegisterFactory(factory ResourceFactory) {
	registry[factory.Version()] = factory
}

// GetFactory returns the ResourceFactory for the specified FHIR version.
// Returns an error if the version is not registered.
func GetFactory(version Version) (ResourceFactory, error) {
	factory, ok := registry[version]
	if !ok {
		return nil, fmt.Errorf("FHIR version %s is not registered", version)
	}
	return factory, nil
}

// SupportedVersions returns a list of all registered FHIR versions.
func SupportedVersions() []Version {
	versions := make([]Version, 0, len(registry))
	for v := range registry {
		versions = append(versions, v)
	}
	return versions
}

// IsVersionSupported returns true if the specified version has a registered factory.
func IsVersionSupported(version Version) bool {
	_, ok := registry[version]
	return ok
}
