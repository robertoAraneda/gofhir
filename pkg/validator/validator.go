// Package validator provides FHIR resource validation based on StructureDefinitions.
package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

// Package-level constants to avoid allocations in hot paths

// choiceSuffixes contains all possible type suffixes for choice types (value[x]).
// Defined once at package level to avoid repeated allocations.
var choiceSuffixes = []string{
	"Boolean", "Integer", "String", "Date", "DateTime", "Time",
	"Decimal", "Uri", "Url", "Canonical", "Code", "Oid", "Id",
	"Markdown", "Base64Binary", "Instant", "PositiveInt", "UnsignedInt",
	"CodeableConcept", "Coding", "Quantity", "Range", "Period",
	"Ratio", "SampledData", "Attachment", "Reference", "Identifier",
	"HumanName", "Address", "ContactPoint", "Timing", "Signature",
	"Annotation", "Money", "Age", "Distance", "Duration", "Count",
}

// complexTypes is a lookup map for FHIR complex types.
// Defined at package level to avoid allocation on each isComplexType call.
var complexTypes = map[string]bool{
	// Data Types
	"Address":             true,
	"Age":                 true,
	"Annotation":          true,
	"Attachment":          true,
	"CodeableConcept":     true,
	"CodeableReference":   true,
	"Coding":              true,
	"ContactDetail":       true,
	"ContactPoint":        true,
	"Contributor":         true,
	"Count":               true,
	"DataRequirement":     true,
	"Distance":            true,
	"Dosage":              true,
	"Duration":            true,
	"Element":             true,
	"ElementDefinition":   true,
	"Expression":          true,
	"Extension":           true,
	"HumanName":           true,
	"Identifier":          true,
	"Meta":                true,
	"Money":               true,
	"Narrative":           true,
	"ParameterDefinition": true,
	"Period":              true,
	"Population":          true,
	"ProdCharacteristic":  true,
	"ProductShelfLife":    true,
	"Quantity":            true,
	"Range":               true,
	"Ratio":               true,
	"RatioRange":          true,
	"Reference":           true,
	"RelatedArtifact":     true,
	"SampledData":         true,
	"Signature":           true,
	"Timing":              true,
	"TriggerDefinition":   true,
	"UsageContext":        true,
	// Backbone elements are also complex
	"BackboneElement": true,
}

// Validator validates FHIR resources against StructureDefinitions.
type Validator struct {
	// Registry provides StructureDefinitions
	registry StructureDefinitionProvider
	// Options configures validation behavior
	options ValidatorOptions
	// TermService validates terminology bindings
	termService TerminologyService
	// RefResolver resolves references
	refResolver ReferenceResolver
	// exprCache caches compiled FHIRPath expressions
	exprCache *expressionCache
}

// expressionCache is a simple thread-safe cache for compiled FHIRPath expressions.
type expressionCache struct {
	mu    sync.RWMutex
	cache map[string]*fhirpath.Expression
	limit int
}

// newExpressionCache creates a new expression cache with the given size limit.
func newExpressionCache(limit int) *expressionCache {
	return &expressionCache{
		cache: make(map[string]*fhirpath.Expression),
		limit: limit,
	}
}

// get retrieves a compiled expression from the cache.
func (c *expressionCache) get(expr string) (*fhirpath.Expression, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	compiled, ok := c.cache[expr]
	return compiled, ok
}

// set stores a compiled expression in the cache.
func (c *expressionCache) set(expr string, compiled *fhirpath.Expression) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Simple eviction: clear cache if it exceeds limit
	if len(c.cache) >= c.limit {
		c.cache = make(map[string]*fhirpath.Expression)
	}
	c.cache[expr] = compiled
}

// validationContext holds parsed data to avoid re-parsing JSON multiple times.
type validationContext struct {
	raw          []byte
	parsed       map[string]interface{}
	resourceType string
	sd           *StructureDef
	index        elementIndex
}

// ValidatorOptions configures validation behavior.
//
//nolint:revive // Keeping ValidatorOptions name for API compatibility
type ValidatorOptions struct {
	// ValidateConstraints enables FHIRPath constraint validation
	ValidateConstraints bool
	// ValidateTerminology enables terminology binding validation
	ValidateTerminology bool
	// ValidateReferences enables reference validation
	ValidateReferences bool
	// StrictMode treats warnings as errors
	StrictMode bool
	// MaxErrors stops validation after this many errors (0 = unlimited)
	MaxErrors int
	// Profile is an optional profile URL to validate against
	Profile string
}

// DefaultValidatorOptions returns sensible default options.
func DefaultValidatorOptions() ValidatorOptions {
	return ValidatorOptions{
		ValidateConstraints: true,
		ValidateTerminology: false, // Requires terminology service
		ValidateReferences:  false, // Requires reference resolver
		StrictMode:          false,
		MaxErrors:           0,
	}
}

// NewValidator creates a new Validator with the given registry and options.
func NewValidator(registry StructureDefinitionProvider, opts ValidatorOptions) *Validator {
	return &Validator{
		registry:    registry,
		options:     opts,
		termService: &NoopTerminologyService{},
		refResolver: &NoopReferenceResolver{},
		exprCache:   newExpressionCache(1000), // Cache up to 1000 expressions
	}
}

// WithTerminologyService sets the terminology service.
func (v *Validator) WithTerminologyService(ts TerminologyService) *Validator {
	v.termService = ts
	return v
}

// WithReferenceResolver sets the reference resolver.
func (v *Validator) WithReferenceResolver(rr ReferenceResolver) *Validator {
	v.refResolver = rr
	return v
}

// Validate validates a FHIR resource (as JSON) against its StructureDefinition.
func (v *Validator) Validate(ctx context.Context, resource []byte) (*ValidationResult, error) {
	result := NewValidationResult()

	// Parse the resource once - reuse throughout validation
	var parsed map[string]any
	if err := json.Unmarshal(resource, &parsed); err != nil {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityFatal,
			Code:        IssueCodeStructure,
			Diagnostics: fmt.Sprintf("Invalid JSON: %v", err),
		})
		return result, nil
	}

	resourceType, ok := parsed["resourceType"].(string)
	if !ok || resourceType == "" {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityFatal,
			Code:        IssueCodeRequired,
			Diagnostics: "Resource must have a resourceType",
			Expression:  []string{"resourceType"},
		})
		return result, nil
	}

	// Get the StructureDefinition
	var sd *StructureDef
	var err error

	if v.options.Profile != "" {
		// Validate against specific profile
		sd, err = v.registry.Get(ctx, v.options.Profile)
		if err != nil {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityFatal,
				Code:        IssueCodeNotFound,
				Diagnostics: fmt.Sprintf("Profile not found: %s", v.options.Profile),
			})
			return result, nil
		}
	} else {
		// Validate against base resource type
		sd, err = v.registry.GetByType(ctx, resourceType)
		if err != nil {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityFatal,
				Code:        IssueCodeNotFound,
				Diagnostics: fmt.Sprintf("Unknown resource type: %s", resourceType),
			})
			return result, nil
		}
	}

	// Build element index for faster lookup
	elemIndex := v.buildElementIndex(sd)

	// Create validation context to pass parsed data (avoids re-parsing)
	vctx := &validationContext{
		raw:          resource,
		parsed:       parsed,
		resourceType: resourceType,
		sd:           sd,
		index:        elemIndex,
	}

	// Validate structure (cardinality, required fields, unknown elements)
	v.validateStructure(ctx, vctx, result)

	// Check max errors
	if v.options.MaxErrors > 0 && result.ErrorCount() >= v.options.MaxErrors {
		return result, nil
	}

	// Validate primitive types
	v.validatePrimitives(ctx, vctx, result)

	// Validate constraints (FHIRPath)
	if v.options.ValidateConstraints {
		v.validateConstraints(ctx, vctx, result)
	}

	// Validate terminology bindings
	if v.options.ValidateTerminology {
		v.validateTerminology(ctx, vctx, result)
	}

	// Validate references
	if v.options.ValidateReferences {
		v.validateReferences(ctx, vctx, result)
	}

	return result, nil
}

// ValidateResource validates a parsed resource map.
func (v *Validator) ValidateResource(ctx context.Context, resource map[string]interface{}) (*ValidationResult, error) {
	data, err := json.Marshal(resource)
	if err != nil {
		result := NewValidationResult()
		result.AddIssue(ValidationIssue{
			Severity:    SeverityFatal,
			Code:        IssueCodeProcessing,
			Diagnostics: fmt.Sprintf("Failed to serialize resource: %v", err),
		})
		return result, nil
	}
	return v.Validate(ctx, data)
}

// elementIndex maps element path to ElementDef for quick lookup.
type elementIndex map[string]*ElementDef

// buildElementIndex creates an index of elements by path.
func (v *Validator) buildElementIndex(sd *StructureDef) elementIndex {
	index := make(elementIndex)
	for i := range sd.Snapshot {
		elem := &sd.Snapshot[i]
		index[elem.Path] = elem
	}
	return index
}

// validateStructure validates cardinality and required fields.
func (v *Validator) validateStructure(ctx context.Context, vctx *validationContext, result *ValidationResult) {
	// Track which required elements are present
	presentElements := make(map[string]bool)

	// Recursively validate the resource structure
	v.validateNode(ctx, vctx.parsed, vctx.sd, vctx.index, vctx.resourceType, "", presentElements, result)

	// Check for missing required elements
	for _, elem := range vctx.sd.Snapshot {
		if elem.Min > 0 {
			// Element is required
			if !presentElements[elem.Path] && !isChildPath(elem.Path, vctx.resourceType) {
				// Only report if it's a direct child of what we're validating
				parentPath := getParentPath(elem.Path)
				if parentPath == vctx.resourceType || presentElements[parentPath] {
					// Check if this is a choice element that might be satisfied by another choice
					if !v.isChoiceElementSatisfied(elem.Path, presentElements) {
						result.AddIssue(ValidationIssue{
							Severity:    SeverityError,
							Code:        IssueCodeRequired,
							Diagnostics: fmt.Sprintf("Missing required element: %s (min=%d)", elem.Path, elem.Min),
							Expression:  []string{elem.Path},
						})
					}
				}
			}
		}
	}
}

// validateNode recursively validates a node in the resource.
//
//nolint:unparam // ctx passed to recursive calls for future cancellation support
func (v *Validator) validateNode(ctx context.Context, node interface{}, sd *StructureDef, index elementIndex, basePath, currentPath string, presentElements map[string]bool, result *ValidationResult) {
	if v.options.MaxErrors > 0 && result.ErrorCount() >= v.options.MaxErrors {
		return
	}

	val, ok := node.(map[string]interface{})
	if !ok {
		return
	}

	for key, child := range val {
		// Skip internal fields
		if key == "resourceType" && currentPath == "" {
			continue
		}
		if strings.HasPrefix(key, "_") {
			// Extension element - validate separately
			continue
		}

		var childPath string
		if currentPath != "" {
			childPath = currentPath + "." + key
		} else {
			childPath = basePath + "." + key
		}

		// Mark element as present
		presentElements[childPath] = true

		// Look up element definition
		elemDef := v.findElementDef(index, childPath, basePath)

		if elemDef == nil {
			// Unknown element
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeStructure,
				Diagnostics: fmt.Sprintf("Unknown element: %s", childPath),
				Expression:  []string{childPath},
			})
			continue
		}

		// Validate cardinality
		v.validateCardinality(child, elemDef, childPath, result)

		// Recursively validate children
		if arr, ok := child.([]interface{}); ok {
			for i, item := range arr {
				itemPath := fmt.Sprintf("%s[%d]", childPath, i)
				v.validateNode(ctx, item, sd, index, basePath, childPath, presentElements, result)
				_ = itemPath // Used for error reporting in more detailed validation
			}
		} else {
			v.validateNode(ctx, child, sd, index, basePath, childPath, presentElements, result)
		}
	}
}

// findElementDef finds the ElementDef for a path, handling choice types and complex types.
func (v *Validator) findElementDef(index elementIndex, path, _ string) *ElementDef {
	// Direct match
	if elem, ok := index[path]; ok {
		return elem
	}

	parts := strings.Split(path, ".")

	// Try choice type (e.g., "Patient.deceased" for "Patient.deceasedBoolean")
	// Uses package-level choiceSuffixes to avoid allocation
	if len(parts) >= 2 {
		lastPart := parts[len(parts)-1]
		for _, suffix := range choiceSuffixes {
			if strings.HasSuffix(lastPart, suffix) {
				// Try to find the [x] version
				baseName := strings.TrimSuffix(lastPart, suffix)
				choicePath := strings.Join(parts[:len(parts)-1], ".") + "." + baseName + "[x]"
				if elem, ok := index[choicePath]; ok {
					return elem
				}
			}
		}
	}

	// For nested elements of complex types (e.g., Patient.name.family or Observation.code.coding.system),
	// check if any ancestor is a complex type.
	if len(parts) >= 3 {
		// Walk backwards through the path to find a complex type ancestor
		for i := len(parts) - 1; i >= 2; i-- {
			ancestorPath := strings.Join(parts[:i], ".")

			// First check direct index
			if ancestorElem, ok := index[ancestorPath]; ok {
				if len(ancestorElem.Types) > 0 {
					typeCode := ancestorElem.Types[0].Code
					if isComplexType(typeCode) {
						// This is a child of a complex type - return synthetic ElementDef
						return &ElementDef{
							Path: path,
							Min:  0,
							Max:  "*",
						}
					}
				}
			}

			// If not found directly, try to find via choice type resolution
			// E.g., Observation.valueQuantity.value -> ancestorPath = Observation.valueQuantity
			ancestorParts := strings.Split(ancestorPath, ".")
			if len(ancestorParts) >= 2 {
				ancestorLastPart := ancestorParts[len(ancestorParts)-1]
				// Uses package-level choiceSuffixes to avoid allocation
				for _, suffix := range choiceSuffixes {
					if strings.HasSuffix(ancestorLastPart, suffix) {
						baseName := strings.TrimSuffix(ancestorLastPart, suffix)
						choicePath := strings.Join(ancestorParts[:len(ancestorParts)-1], ".") + "." + baseName + "[x]"
						if choiceElem, ok := index[choicePath]; ok {
							// Found the choice type element - check if the suffix type is complex
							if isComplexType(suffix) {
								return &ElementDef{
									Path: path,
									Min:  0,
									Max:  "*",
								}
							}
							// Also check if any of the choice types is complex
							for _, t := range choiceElem.Types {
								if t.Code == suffix && isComplexType(t.Code) {
									return &ElementDef{
										Path: path,
										Min:  0,
										Max:  "*",
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// isComplexType returns true if the type code is a complex FHIR type.
// Uses package-level complexTypes map to avoid allocation on each call.
func isComplexType(typeCode string) bool {
	return complexTypes[typeCode]
}

// validateCardinality checks if the value satisfies min/max cardinality.
func (v *Validator) validateCardinality(value interface{}, elem *ElementDef, path string, result *ValidationResult) {
	var count int

	switch val := value.(type) {
	case []interface{}:
		count = len(val)
	case nil:
		count = 0
	default:
		count = 1
	}

	// Check min
	if count < elem.Min {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeRequired,
			Diagnostics: fmt.Sprintf("Element '%s' has %d items but minimum is %d", path, count, elem.Min),
			Expression:  []string{path},
		})
	}

	// Check max
	if elem.Max != "*" && elem.Max != "" {
		var maxVal int
		if _, err := fmt.Sscanf(elem.Max, "%d", &maxVal); err == nil && maxVal > 0 && count > maxVal {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeStructure,
				Diagnostics: fmt.Sprintf("Element '%s' has %d items but maximum is %d", path, count, maxVal),
				Expression:  []string{path},
			})
		}
	}
}

// validatePrimitives validates primitive type values.
func (v *Validator) validatePrimitives(_ context.Context, vctx *validationContext, result *ValidationResult) {
	v.validatePrimitiveNode(vctx.parsed, vctx.index, vctx.resourceType, result)
}

// validatePrimitiveNode recursively validates primitive values.
func (v *Validator) validatePrimitiveNode(node interface{}, index elementIndex, path string, result *ValidationResult) {
	switch val := node.(type) {
	case map[string]interface{}:
		for key, child := range val {
			if key == "resourceType" || strings.HasPrefix(key, "_") {
				continue
			}
			childPath := path + "." + key
			v.validatePrimitiveNode(child, index, childPath, result)
		}
	case []interface{}:
		for _, item := range val {
			v.validatePrimitiveNode(item, index, path, result)
		}
	default:
		// Validate primitive value against type
		elemDef := v.findElementDef(index, path, strings.Split(path, ".")[0])
		if elemDef != nil && len(elemDef.Types) > 0 {
			v.validatePrimitiveValue(val, elemDef.Types[0].Code, path, result)
		}
	}
}

// validatePrimitiveValue validates a primitive value against its type.
func (v *Validator) validatePrimitiveValue(value interface{}, typeCode, path string, result *ValidationResult) {
	// Type validation based on FHIR primitive types
	switch typeCode {
	case "boolean":
		if _, ok := value.(bool); !ok {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Element '%s' must be a boolean", path),
				Expression:  []string{path},
			})
		}
	case "integer", "positiveInt", "unsignedInt":
		switch v := value.(type) {
		case float64:
			if v != float64(int(v)) {
				result.AddIssue(ValidationIssue{
					Severity:    SeverityError,
					Code:        IssueCodeValue,
					Diagnostics: fmt.Sprintf("Element '%s' must be an integer", path),
					Expression:  []string{path},
				})
			}
			if typeCode == "positiveInt" && v <= 0 {
				result.AddIssue(ValidationIssue{
					Severity:    SeverityError,
					Code:        IssueCodeValue,
					Diagnostics: fmt.Sprintf("Element '%s' must be a positive integer", path),
					Expression:  []string{path},
				})
			}
			if typeCode == "unsignedInt" && v < 0 {
				result.AddIssue(ValidationIssue{
					Severity:    SeverityError,
					Code:        IssueCodeValue,
					Diagnostics: fmt.Sprintf("Element '%s' must be a non-negative integer", path),
					Expression:  []string{path},
				})
			}
		default:
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Element '%s' must be an integer", path),
				Expression:  []string{path},
			})
		}
	case "decimal":
		if _, ok := value.(float64); !ok {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Element '%s' must be a decimal number", path),
				Expression:  []string{path},
			})
		}
	case "string", "code", "id", "markdown", "uri", "url", "canonical", "oid", "uuid":
		if _, ok := value.(string); !ok {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Element '%s' must be a string", path),
				Expression:  []string{path},
			})
		}
		// TODO: Add regex validation for specific string types
	}
}

// validateConstraints validates FHIRPath constraints defined in the StructureDefinition.
// Uses validationContext to avoid re-parsing JSON.
func (v *Validator) validateConstraints(_ context.Context, vctx *validationContext, result *ValidationResult) {
	// Collect all constraints from snapshot elements
	for _, elem := range vctx.sd.Snapshot {
		for _, constraint := range elem.Constraints {
			// Skip constraints without expressions
			if constraint.Expression == "" {
				continue
			}

			// Skip constraints from external sources (they're validated by the source profile)
			// Only validate constraints defined in this StructureDefinition
			if constraint.Source != "" && constraint.Source != vctx.sd.URL {
				continue
			}

			// Only validate constraints for elements that exist in the resource
			// Root level constraints (e.g., Patient) always apply
			if elem.Path != vctx.resourceType && !elementExistsInResource(vctx.parsed, elem.Path, vctx.resourceType) {
				continue
			}

			// Evaluate the FHIRPath expression
			valid, err := v.evaluateConstraint(vctx.raw, elem.Path, vctx.resourceType, constraint)
			if err != nil {
				// If expression fails to evaluate, report as warning
				result.AddIssue(ValidationIssue{
					Severity:    SeverityWarning,
					Code:        IssueCodeProcessing,
					Diagnostics: fmt.Sprintf("Failed to evaluate constraint %s on %s: %v", constraint.Key, elem.Path, err),
					Expression:  []string{elem.Path},
				})
				continue
			}

			if !valid {
				// Constraint violated
				severity := SeverityError
				if constraint.Severity == "warning" {
					severity = SeverityWarning
				}

				result.AddIssue(ValidationIssue{
					Severity:    severity,
					Code:        IssueCodeInvariant,
					Diagnostics: fmt.Sprintf("Constraint %s violated: %s", constraint.Key, constraint.Human),
					Expression:  []string{elem.Path},
				})
			}
		}
	}
}

// elementExistsInResource checks if an element path exists in the resource.
func elementExistsInResource(resource map[string]interface{}, elementPath, resourceType string) bool {
	// Remove resource type prefix
	path := strings.TrimPrefix(elementPath, resourceType+".")
	if path == elementPath {
		// Path doesn't start with resource type
		return false
	}

	parts := strings.Split(path, ".")
	current := interface{}(resource)

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var found bool
			current, found = v[part]
			if !found {
				// Try choice type variants
				for key := range v {
					if strings.HasPrefix(key, part) {
						current = v[key]
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		case []interface{}:
			// For arrays, check if any element has the path
			if len(v) == 0 {
				return false
			}
			// Check first element
			if m, ok := v[0].(map[string]interface{}); ok {
				if val, found := m[part]; found {
					current = val
				} else {
					return false
				}
			} else {
				return false
			}
		default:
			return false
		}
	}

	return true
}

// evaluateConstraint evaluates a single FHIRPath constraint.
// For element-level constraints, wraps the expression to evaluate in the context of that element.
// Uses expression cache to avoid recompiling the same expressions.
func (v *Validator) evaluateConstraint(resource []byte, elementPath, resourceType string, constraint ElementConstraint) (bool, error) {
	// Build the full FHIRPath expression
	// For root-level constraints (e.g., Patient), use the expression directly
	// For element-level constraints (e.g., Patient.contact), wrap with .all()
	fullExpr := constraint.Expression
	if elementPath != resourceType {
		// Element-level constraint - need to evaluate in context of the element
		// Convert "Patient.contact" -> "contact" relative path
		relativePath := strings.TrimPrefix(elementPath, resourceType+".")
		// Wrap: contact.all(name.exists() or telecom.exists() ...)
		fullExpr = fmt.Sprintf("%s.all(%s)", relativePath, constraint.Expression)
	}

	// Try to get compiled expression from cache
	var expr *fhirpath.Expression
	var err error

	if cached, ok := v.exprCache.get(fullExpr); ok {
		expr = cached
	} else {
		// Compile the FHIRPath expression
		expr, err = fhirpath.Compile(fullExpr)
		if err != nil {
			return false, fmt.Errorf("compile error: %w", err)
		}
		// Store in cache for future use
		v.exprCache.set(fullExpr, expr)
	}

	// Evaluate the expression
	result, err := expr.Evaluate(resource)
	if err != nil {
		return false, fmt.Errorf("evaluation error: %w", err)
	}

	// Check the result
	return isTruthy(result), nil
}

// isTruthy determines if a FHIRPath result is truthy for constraint evaluation.
// Per FHIRPath spec: empty = false, single boolean = its value, otherwise = true
func isTruthy(result types.Collection) bool {
	if result.Empty() {
		return false
	}

	// If single boolean, return its value
	if len(result) == 1 {
		if b, ok := result[0].(types.Boolean); ok {
			return b.Bool()
		}
	}

	// Non-empty collection is truthy
	return true
}

// validateTerminology validates terminology bindings.
func (v *Validator) validateTerminology(ctx context.Context, vctx *validationContext, result *ValidationResult) {
	// TODO: Implement terminology validation using termService
	// For each element with a binding, validate the code against the ValueSet
}

// validateReferences validates resource references.
func (v *Validator) validateReferences(ctx context.Context, vctx *validationContext, result *ValidationResult) {
	// TODO: Implement reference validation using refResolver
	// For each Reference element, validate that the reference can be resolved
}

// Helper functions

func isChildPath(path, parent string) bool {
	return strings.HasPrefix(path, parent+".")
}

func getParentPath(path string) string {
	lastDot := strings.LastIndex(path, ".")
	if lastDot == -1 {
		return ""
	}
	return path[:lastDot]
}

func (v *Validator) isChoiceElementSatisfied(path string, present map[string]bool) bool {
	// Check if this is a [x] path and if any variant is present
	if !strings.HasSuffix(path, "[x]") {
		return false
	}

	basePath := strings.TrimSuffix(path, "[x]")
	for presentPath := range present {
		if strings.HasPrefix(presentPath, basePath) && presentPath != path {
			return true
		}
	}
	return false
}
