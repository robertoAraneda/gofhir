// Package validator provides FHIR resource validation based on StructureDefinitions.
package validator

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Reference format patterns according to FHIR specification.
// https://www.hl7.org/fhir/references.html
var (
	// relativeRefPattern matches: ResourceType/id (e.g., "Patient/123")
	relativeRefPattern = regexp.MustCompile(`^([A-Za-z]+)/([A-Za-z0-9\-.]+)$`)

	// absoluteRefPattern matches: http(s)://server/path/ResourceType/id
	absoluteRefPattern = regexp.MustCompile(`^https?://[^/]+/.*/([A-Za-z]+)/([A-Za-z0-9\-.]+)$`)

	// containedRefPattern matches: #id (reference to contained resource)
	containedRefPattern = regexp.MustCompile(`^#([A-Za-z0-9\-.]+)$`)

	// urnUUIDPattern matches: urn:uuid:xxxx (used in Bundles)
	urnUUIDPattern = regexp.MustCompile(`^urn:uuid:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	// urnOIDPattern matches: urn:oid:x.x.x.x
	urnOIDPattern = regexp.MustCompile(`^urn:oid:[012](\.\d+)+$`)
)

// ParsedReference contains the parsed components of a FHIR reference.
type ParsedReference struct {
	// Type is the reference type (relative, absolute, contained, urn-uuid, urn-oid, canonical)
	Type string
	// ResourceType is the referenced resource type (if extractable)
	ResourceType string
	// ID is the resource ID (if extractable)
	ID string
	// Raw is the original reference string
	Raw string
	// Valid indicates if the reference format is valid
	Valid bool
	// Version for canonical references
	Version string
}

// ReferenceType constants
const (
	RefTypeRelative  = "relative"
	RefTypeAbsolute  = "absolute"
	RefTypeContained = "contained"
	RefTypeUrnUUID   = "urn-uuid"
	RefTypeUrnOID    = "urn-oid"
	RefTypeCanonical = "canonical"
	RefTypeUnknown   = "unknown"
)

// ParseReference parses a FHIR reference string and extracts its components.
func ParseReference(ref string) *ParsedReference {
	if ref == "" {
		return &ParsedReference{Raw: ref, Valid: false, Type: RefTypeUnknown}
	}

	// Try contained reference first (#id)
	if matches := containedRefPattern.FindStringSubmatch(ref); matches != nil {
		return &ParsedReference{
			Type:  RefTypeContained,
			ID:    matches[1],
			Raw:   ref,
			Valid: true,
		}
	}

	// Try relative reference (ResourceType/id)
	if matches := relativeRefPattern.FindStringSubmatch(ref); matches != nil {
		return &ParsedReference{
			Type:         RefTypeRelative,
			ResourceType: matches[1],
			ID:           matches[2],
			Raw:          ref,
			Valid:        true,
		}
	}

	// Try URN:UUID
	if urnUUIDPattern.MatchString(ref) {
		return &ParsedReference{
			Type:  RefTypeUrnUUID,
			ID:    strings.TrimPrefix(ref, "urn:uuid:"),
			Raw:   ref,
			Valid: true,
		}
	}

	// Try URN:OID
	if urnOIDPattern.MatchString(ref) {
		return &ParsedReference{
			Type:  RefTypeUrnOID,
			ID:    strings.TrimPrefix(ref, "urn:oid:"),
			Raw:   ref,
			Valid: true,
		}
	}

	// Try absolute reference (http://server/path/ResourceType/id)
	// Must be checked AFTER URN patterns
	if matches := absoluteRefPattern.FindStringSubmatch(ref); matches != nil {
		return &ParsedReference{
			Type:         RefTypeAbsolute,
			ResourceType: matches[1],
			ID:           matches[2],
			Raw:          ref,
			Valid:        true,
		}
	}

	// Try canonical URL - HTTP/HTTPS URLs that don't match absolute pattern
	// (e.g., StructureDefinition URLs without ResourceType/id pattern)
	if strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://") {
		parsed := &ParsedReference{
			Type:  RefTypeCanonical,
			Raw:   ref,
			Valid: true,
		}
		// Check for version suffix
		if idx := strings.LastIndex(ref, "|"); idx != -1 {
			parsed.Version = ref[idx+1:]
		}
		return parsed
	}

	// Unknown format
	return &ParsedReference{Raw: ref, Valid: false, Type: RefTypeUnknown}
}

// validateReferences validates all Reference elements in the resource.
func (v *Validator) validateReferences(ctx context.Context, vctx *validationContext, result *ValidationResult) {
	// Extract contained resources for local reference validation
	containedIDs := v.extractContainedIDs(vctx.parsed)

	// Recursively find and validate all references
	v.validateReferencesInNode(ctx, vctx, vctx.parsed, vctx.resourceType, containedIDs, result)
}

// extractContainedIDs extracts IDs of all contained resources.
func (v *Validator) extractContainedIDs(resource map[string]interface{}) map[string]string {
	contained := make(map[string]string)

	if containedArr, ok := resource["contained"].([]interface{}); ok {
		for _, item := range containedArr {
			if res, ok := item.(map[string]interface{}); ok {
				if id, ok := res["id"].(string); ok {
					if rt, ok := res["resourceType"].(string); ok {
						contained[id] = rt
					}
				}
			}
		}
	}

	return contained
}

// validateReferencesInNode recursively validates references in a node.
func (v *Validator) validateReferencesInNode(ctx context.Context, vctx *validationContext, node interface{}, path string, containedIDs map[string]string, result *ValidationResult) {
	if v.options.MaxErrors > 0 && result.ErrorCount() >= v.options.MaxErrors {
		return
	}

	switch val := node.(type) {
	case map[string]interface{}:
		// Check if this is a Reference type (has "reference" field)
		if refStr, ok := val["reference"].(string); ok {
			v.validateSingleReference(ctx, vctx, refStr, path, containedIDs, result)
		}

		// Recursively check children
		for key, child := range val {
			if key == "contained" {
				// Skip contained - we already extracted IDs
				continue
			}
			childPath := path + "." + key
			v.validateReferencesInNode(ctx, vctx, child, childPath, containedIDs, result)
		}

	case []interface{}:
		for i, item := range val {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			v.validateReferencesInNode(ctx, vctx, item, itemPath, containedIDs, result)
		}
	}
}

// validateSingleReference validates a single reference string.
func (v *Validator) validateSingleReference(ctx context.Context, vctx *validationContext, refStr, path string, containedIDs map[string]string, result *ValidationResult) {
	parsed := ParseReference(refStr)

	// 1. Validate format
	if !parsed.Valid {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeValue,
			Diagnostics: fmt.Sprintf("Invalid reference format: '%s'", refStr),
			Expression:  []string{path + ".reference"},
		})
		return
	}

	// 2. Validate contained references
	if parsed.Type == RefTypeContained {
		if _, exists := containedIDs[parsed.ID]; !exists {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeNotFound,
				Diagnostics: fmt.Sprintf("Contained resource not found: '%s'", refStr),
				Expression:  []string{path + ".reference"},
			})
		}
		return
	}

	// 3. Validate target type against allowed types (if we have type info in the path)
	if parsed.ResourceType != "" {
		v.validateReferenceTargetType(vctx, parsed, path, result)
	}

	// 4. Optional: resolve reference if resolver is configured
	// This is skipped by default (NoopReferenceResolver)
	if _, isNoop := v.refResolver.(*NoopReferenceResolver); !isNoop {
		_, err := v.refResolver.Resolve(ctx, refStr)
		if err != nil {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityWarning,
				Code:        IssueCodeNotFound,
				Diagnostics: fmt.Sprintf("Could not resolve reference '%s': %v", refStr, err),
				Expression:  []string{path + ".reference"},
			})
		}
	}
}

// validateReferenceTargetType validates that the referenced resource type is allowed.
func (v *Validator) validateReferenceTargetType(vctx *validationContext, parsed *ParsedReference, path string, result *ValidationResult) {
	// Find the element definition for this reference
	elemPath := pathWithoutArrayIndices(path)
	elemDef := v.findElementDef(vctx.index, elemPath, vctx.resourceType)

	if elemDef == nil {
		return // Can't validate without element definition
	}

	// Check if any of the types is Reference with targetProfile
	for _, typeRef := range elemDef.Types {
		if typeRef.Code == "Reference" {
			// If no target profiles specified, any type is allowed
			if len(typeRef.TargetProfile) == 0 {
				return
			}

			// Check if the referenced type matches any allowed target
			for _, profile := range typeRef.TargetProfile {
				// Extract resource type from profile URL
				// e.g., "http://hl7.org/fhir/StructureDefinition/Patient" -> "Patient"
				allowedType := extractResourceTypeFromProfile(profile)
				if allowedType == parsed.ResourceType || allowedType == "Resource" {
					return // Match found
				}
			}

			// No match found - report error
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Reference to '%s' not allowed; expected one of: %s", parsed.ResourceType, formatAllowedTypes(typeRef.TargetProfile)),
				Expression:  []string{path + ".reference"},
			})
			return
		}
	}
}

// pathWithoutArrayIndices removes array indices from a path.
// e.g., "Patient.contact[0].reference" -> "Patient.contact.reference"
func pathWithoutArrayIndices(path string) string {
	// Simple regex to remove [n] patterns
	indexPattern := regexp.MustCompile(`\[\d+\]`)
	return indexPattern.ReplaceAllString(path, "")
}

// extractResourceTypeFromProfile extracts the resource type from a StructureDefinition URL.
func extractResourceTypeFromProfile(profile string) string {
	// Handle standard FHIR profiles
	if strings.Contains(profile, "/StructureDefinition/") {
		parts := strings.Split(profile, "/StructureDefinition/")
		if len(parts) == 2 {
			// Handle version suffix (|4.0.1)
			typePart := strings.Split(parts[1], "|")[0]
			return typePart
		}
	}

	// Handle simple resource type names
	if !strings.Contains(profile, "/") {
		return profile
	}

	// Last segment of URL
	parts := strings.Split(profile, "/")
	return parts[len(parts)-1]
}

// formatAllowedTypes formats the list of allowed target profiles for error messages.
func formatAllowedTypes(profiles []string) string {
	types := make([]string, 0, len(profiles))
	for _, p := range profiles {
		types = append(types, extractResourceTypeFromProfile(p))
	}
	return strings.Join(types, ", ")
}
