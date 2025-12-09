// Package validator provides FHIR resource validation based on StructureDefinitions.
package validator

// StructureDef is a version-agnostic internal model for StructureDefinition.
// It extracts only the fields needed for validation, working across R4, R4B, and R5.
type StructureDef struct {
	// URL is the canonical identifier for this StructureDefinition
	URL string `json:"url"`
	// Name is the computer-friendly name
	Name string `json:"name"`
	// Type is the type defined or constrained (e.g., "Patient", "Observation")
	Type string `json:"type"`
	// Kind is the structure kind: primitive-type, complex-type, resource, logical
	Kind string `json:"kind"`
	// Abstract indicates if this is an abstract type
	Abstract bool `json:"abstract"`
	// BaseDefinition is the URL of the parent StructureDefinition
	BaseDefinition string `json:"baseDefinition,omitempty"`
	// FHIRVersion is the FHIR version this definition targets
	FHIRVersion string `json:"fhirVersion,omitempty"`
	// Snapshot contains the full element definitions
	Snapshot []ElementDef `json:"snapshot,omitempty"`
	// Differential contains only the changed elements (for profiles)
	Differential []ElementDef `json:"differential,omitempty"`
}

// ElementDef is a version-agnostic internal model for ElementDefinition.
// Contains all fields needed for validation across FHIR versions.
type ElementDef struct {
	// ID is the unique identifier within the StructureDefinition
	ID string `json:"id,omitempty"`
	// Path is the element path (e.g., "Patient.name", "Patient.name.given")
	Path string `json:"path"`
	// SliceName for sliced elements
	SliceName string `json:"sliceName,omitempty"`
	// Min cardinality (0 or 1 typically)
	Min int `json:"min"`
	// Max cardinality ("*" = unbounded, "0" = prohibited, "1" = single)
	Max string `json:"max"`
	// Types allowed for this element
	Types []TypeRef `json:"type,omitempty"`
	// Short description
	Short string `json:"short,omitempty"`
	// Definition (full description)
	Definition string `json:"definition,omitempty"`
	// Fixed value (if element must have exact value)
	Fixed interface{} `json:"fixed,omitempty"`
	// Pattern value (if element must match pattern)
	Pattern interface{} `json:"pattern,omitempty"`
	// Binding to a ValueSet
	Binding *ElementBinding `json:"binding,omitempty"`
	// Constraints (FHIRPath invariants)
	Constraints []ElementConstraint `json:"constraint,omitempty"`
	// MustSupport indicates if the element is required for conformance
	MustSupport bool `json:"mustSupport,omitempty"`
	// IsModifier indicates if the element can modify other elements' meaning
	IsModifier bool `json:"isModifier,omitempty"`
	// IsSummary indicates if the element is part of the summary view
	IsSummary bool `json:"isSummary,omitempty"`
}

// TypeRef represents a type reference for an element.
type TypeRef struct {
	// Code is the type code (e.g., "string", "Reference", "CodeableConcept")
	Code string `json:"code"`
	// TargetProfile for Reference types - what resources can be referenced
	TargetProfile []string `json:"targetProfile,omitempty"`
	// Profile for complex types - what profiles must be followed
	Profile []string `json:"profile,omitempty"`
}

// ElementBinding represents a terminology binding for an element.
type ElementBinding struct {
	// Strength: required | extensible | preferred | example
	Strength string `json:"strength"`
	// ValueSet URL
	ValueSet string `json:"valueSet,omitempty"`
	// Description of the binding
	Description string `json:"description,omitempty"`
}

// ElementConstraint represents a FHIRPath constraint on an element.
type ElementConstraint struct {
	// Key is the unique constraint identifier (e.g., "ele-1", "pat-1")
	Key string `json:"key"`
	// Severity: error | warning
	Severity string `json:"severity"`
	// Human readable description
	Human string `json:"human,omitempty"`
	// FHIRPath expression to evaluate
	Expression string `json:"expression,omitempty"`
	// XPath expression (legacy, optional)
	XPath string `json:"xpath,omitempty"`
	// Source URL of the constraint definition
	Source string `json:"source,omitempty"`
}

// ValidationIssue represents a single validation issue found during validation.
// This is version-agnostic and maps to OperationOutcome.issue in any FHIR version.
type ValidationIssue struct {
	// Severity: fatal | error | warning | information
	Severity string `json:"severity"`
	// Code: structure | required | value | invariant | processing | etc.
	Code string `json:"code"`
	// Diagnostics message (human readable)
	Diagnostics string `json:"diagnostics,omitempty"`
	// Location in the resource (FHIRPath expression)
	Location []string `json:"location,omitempty"`
	// Expression (FHIRPath) that identifies the element
	Expression []string `json:"expression,omitempty"`
}

// ValidationResult contains the result of validating a resource.
type ValidationResult struct {
	// Valid is true if no errors were found (warnings are allowed)
	Valid bool `json:"valid"`
	// Issues contains all validation issues found
	Issues []ValidationIssue `json:"issues,omitempty"`
}

// Severity constants for ValidationIssue
const (
	SeverityFatal       = "fatal"
	SeverityError       = "error"
	SeverityWarning     = "warning"
	SeverityInformation = "information"
)

// Issue code constants (subset of OperationOutcome issue types)
const (
	IssueCodeStructure   = "structure"    // Structural issue
	IssueCodeRequired    = "required"     // Required element missing
	IssueCodeValue       = "value"        // Invalid value
	IssueCodeInvariant   = "invariant"    // Invariant/constraint violation
	IssueCodeProcessing  = "processing"   // Processing error
	IssueCodeInvalid     = "invalid"      // Invalid content
	IssueCodeNotFound    = "not-found"    // Reference not found
	IssueCodeCodeInvalid = "code-invalid" // Invalid code
	IssueCodeExtension   = "extension"    // Extension error
)

// HasErrors returns true if there are any fatal or error severity issues.
func (r *ValidationResult) HasErrors() bool {
	for _, issue := range r.Issues {
		if issue.Severity == SeverityFatal || issue.Severity == SeverityError {
			return true
		}
	}
	return false
}

// HasWarnings returns true if there are any warning severity issues.
func (r *ValidationResult) HasWarnings() bool {
	for _, issue := range r.Issues {
		if issue.Severity == SeverityWarning {
			return true
		}
	}
	return false
}

// ErrorCount returns the number of fatal and error issues.
func (r *ValidationResult) ErrorCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == SeverityFatal || issue.Severity == SeverityError {
			count++
		}
	}
	return count
}

// WarningCount returns the number of warning issues.
func (r *ValidationResult) WarningCount() int {
	count := 0
	for _, issue := range r.Issues {
		if issue.Severity == SeverityWarning {
			count++
		}
	}
	return count
}

// AddIssue adds a validation issue to the result.
func (r *ValidationResult) AddIssue(issue ValidationIssue) {
	r.Issues = append(r.Issues, issue)
	if issue.Severity == SeverityFatal || issue.Severity == SeverityError {
		r.Valid = false
	}
}

// NewValidationResult creates a new validation result (initially valid).
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:  true,
		Issues: []ValidationIssue{},
	}
}

// Merge combines another validation result into this one.
func (r *ValidationResult) Merge(other *ValidationResult) {
	if other == nil {
		return
	}
	for _, issue := range other.Issues {
		r.AddIssue(issue)
	}
}
