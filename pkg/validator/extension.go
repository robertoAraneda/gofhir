// Package validator provides FHIR resource validation based on StructureDefinitions.
package validator

import (
	"context"
	"fmt"
	"strings"
)

// hl7ExtensionPrefix is the prefix for HL7-defined extensions.
const hl7ExtensionPrefix = "http://hl7.org/fhir/StructureDefinition/"

// ParsedExtension contains the parsed components of a FHIR extension.
type ParsedExtension struct {
	// URL is the extension URL
	URL string
	// Value is the extension value (one of value[x])
	Value interface{}
	// NestedExtensions are nested extensions (for complex extensions)
	NestedExtensions []ParsedExtension
	// Valid indicates if the extension format is valid
	Valid bool
	// IsComplex indicates if this is a complex extension (has nested extensions)
	IsComplex bool
}

// validateExtensions validates all Extension elements in the resource.
func (v *Validator) validateExtensions(ctx context.Context, vctx *validationContext, result *ValidationResult) {
	// Recursively find and validate all extensions
	v.validateExtensionsInNode(ctx, vctx, vctx.parsed, vctx.resourceType, result)
}

// validateExtensionsInNode recursively validates extensions in a node.
func (v *Validator) validateExtensionsInNode(ctx context.Context, vctx *validationContext, node interface{}, path string, result *ValidationResult) {
	if v.options.MaxErrors > 0 && result.ErrorCount() >= v.options.MaxErrors {
		return
	}

	switch val := node.(type) {
	case map[string]interface{}:
		// Check for "extension" field
		if extensions, ok := val["extension"].([]interface{}); ok {
			v.validateExtensionArray(ctx, vctx, extensions, path+".extension", result)
		}

		// Check for "modifierExtension" field
		if modExtensions, ok := val["modifierExtension"].([]interface{}); ok {
			v.validateExtensionArray(ctx, vctx, modExtensions, path+".modifierExtension", result)
		}

		// Recursively check children (skip extension fields themselves)
		for key, child := range val {
			if key == "extension" || key == "modifierExtension" {
				continue
			}
			childPath := path + "." + key
			v.validateExtensionsInNode(ctx, vctx, child, childPath, result)
		}

	case []interface{}:
		for i, item := range val {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			v.validateExtensionsInNode(ctx, vctx, item, itemPath, result)
		}
	}
}

// validateExtensionArray validates an array of extensions.
func (v *Validator) validateExtensionArray(ctx context.Context, vctx *validationContext, extensions []interface{}, path string, result *ValidationResult) {
	for i, ext := range extensions {
		extPath := fmt.Sprintf("%s[%d]", path, i)
		if extMap, ok := ext.(map[string]interface{}); ok {
			v.validateSingleExtension(ctx, vctx, extMap, extPath, result)
		} else {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeStructure,
				Diagnostics: "Extension must be an object",
				Expression:  []string{extPath},
			})
		}
	}
}

// validateSingleExtension validates a single extension object.
func (v *Validator) validateSingleExtension(ctx context.Context, vctx *validationContext, ext map[string]interface{}, path string, result *ValidationResult) {
	// 1. Validate URL is present and valid format
	url, hasURL := ext["url"].(string)
	if !hasURL || url == "" {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeRequired,
			Diagnostics: "Extension must have a 'url' field",
			Expression:  []string{path},
		})
		return
	}

	// 2. Validate URL format
	if !isValidExtensionURL(url) {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeValue,
			Diagnostics: fmt.Sprintf("Invalid extension URL format: '%s'", url),
			Expression:  []string{path + ".url"},
		})
	}

	// 3. Check for value[x] or nested extensions (mutually exclusive)
	hasValue := hasExtensionValue(ext)
	hasNestedExt := hasNestedExtensions(ext)

	if hasValue && hasNestedExt {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeStructure,
			Diagnostics: "Extension cannot have both a value and nested extensions",
			Expression:  []string{path},
		})
	}

	if !hasValue && !hasNestedExt {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeRequired,
			Diagnostics: "Extension must have either a value[x] or nested extensions",
			Expression:  []string{path},
		})
	}

	// 4. Validate nested extensions recursively
	if nestedExts, ok := ext["extension"].([]interface{}); ok {
		for i, nested := range nestedExts {
			nestedPath := fmt.Sprintf("%s.extension[%d]", path, i)
			if nestedMap, ok := nested.(map[string]interface{}); ok {
				v.validateSingleExtension(ctx, vctx, nestedMap, nestedPath, result)
			}
		}
	}

	// 5. Validate against StructureDefinition if available
	v.validateExtensionAgainstDefinition(ctx, vctx, ext, url, path, result)
}

// validateExtensionAgainstDefinition validates an extension against its StructureDefinition.
func (v *Validator) validateExtensionAgainstDefinition(ctx context.Context, vctx *validationContext, ext map[string]interface{}, url, path string, result *ValidationResult) {
	// Try to get the extension's StructureDefinition from the registry
	sd, err := v.registry.Get(ctx, url)
	if err != nil || sd == nil {
		// Extension definition not found - this is a warning, not an error
		// Unknown extensions are allowed in FHIR
		if v.options.StrictMode {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityWarning,
				Code:        IssueCodeExtension,
				Diagnostics: fmt.Sprintf("Extension definition not found: '%s'", url),
				Expression:  []string{path},
			})
		}
		return
	}

	// Verify this is actually an Extension definition
	if sd.Type != "Extension" {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeExtension,
			Diagnostics: fmt.Sprintf("URL '%s' does not define an Extension (type: %s)", url, sd.Type),
			Expression:  []string{path},
		})
		return
	}

	// Validate context if defined in the StructureDefinition
	v.validateExtensionContext(ctx, vctx, sd, path, result)

	// Validate value type against allowed types in the definition
	v.validateExtensionValueType(ctx, ext, sd, path, result)
}

// validateExtensionContext validates that the extension is used in an allowed context.
func (v *Validator) validateExtensionContext(_ context.Context, _ *validationContext, sd *StructureDef, _ string, _ *ValidationResult) {
	// Context validation requires parsing the extension's context from StructureDefinition
	// For now, we extract context from the Extension.extension element definitions
	// The context is typically defined in the StructureDefinition.context field (R4+)

	// Find context restrictions in the snapshot
	for _, elem := range sd.Snapshot {
		if elem.Path == "Extension" && len(elem.Types) > 0 {
			// Check if there are context restrictions
			// This would be in the StructureDefinition.context array in the original JSON
			// For now, we skip detailed context validation as it requires additional parsing
			break
		}
	}
}

// validateExtensionValueType validates the extension value against allowed types.
func (v *Validator) validateExtensionValueType(_ context.Context, ext map[string]interface{}, sd *StructureDef, path string, result *ValidationResult) {
	// Find the Extension.value[x] element definition
	var valueElement *ElementDef
	for i := range sd.Snapshot {
		elem := &sd.Snapshot[i]
		if strings.HasPrefix(elem.Path, "Extension.value") {
			valueElement = elem
			break
		}
	}

	if valueElement == nil {
		// No value element defined - this is a complex extension
		return
	}

	// Get the actual value type from the extension
	actualValueType := getExtensionValueType(ext)
	if actualValueType == "" {
		return // No value present (nested extensions instead)
	}

	// Check if the value type is allowed
	if len(valueElement.Types) > 0 {
		allowed := false
		var allowedTypes []string
		for _, t := range valueElement.Types {
			allowedTypes = append(allowedTypes, t.Code)
			if strings.EqualFold(t.Code, actualValueType) {
				allowed = true
				break
			}
		}

		if !allowed {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Extension value type '%s' not allowed; expected one of: %s", actualValueType, strings.Join(allowedTypes, ", ")),
				Expression:  []string{path},
			})
		}
	}
}

// isValidExtensionURL checks if an extension URL has valid format.
// For top-level extensions, URL must be absolute (http/https/urn).
// For nested extensions within complex extensions, URL can be a simple name.
func isValidExtensionURL(url string) bool {
	// Must not be empty
	if url == "" {
		return false
	}

	// Check for absolute URLs (required for top-level extensions)
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return true
	}
	if strings.HasPrefix(url, "urn:") {
		return true
	}

	// For nested extensions, simple names are allowed (e.g., "latitude", "longitude")
	// These are relative to the parent extension's definition
	// A simple name is alphanumeric with optional hyphens/underscores
	if isSimpleExtensionName(url) {
		return true
	}

	return false
}

// isSimpleExtensionName checks if a string is a valid simple extension name.
// Used for nested extensions within complex extensions.
func isSimpleExtensionName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// hasExtensionValue checks if an extension has a value[x] field.
func hasExtensionValue(ext map[string]interface{}) bool {
	for key := range ext {
		if strings.HasPrefix(key, "value") && key != "value" {
			// value followed by a type name (valueString, valueCode, etc.)
			return true
		}
	}
	return false
}

// hasNestedExtensions checks if an extension has nested extensions.
func hasNestedExtensions(ext map[string]interface{}) bool {
	if nested, ok := ext["extension"].([]interface{}); ok {
		return len(nested) > 0
	}
	return false
}

// getExtensionValueType returns the type of value in an extension (e.g., "String" from "valueString").
func getExtensionValueType(ext map[string]interface{}) string {
	for key := range ext {
		if strings.HasPrefix(key, "value") && key != "value" {
			// Extract type from "valueXxx" -> "Xxx"
			return key[5:] // Remove "value" prefix
		}
	}
	return ""
}

// IsHL7Extension checks if the URL is an HL7-defined extension.
func IsHL7Extension(url string) bool {
	return strings.HasPrefix(url, hl7ExtensionPrefix)
}

// ExtractExtensionName extracts the extension name from a URL.
// e.g., "http://hl7.org/fhir/StructureDefinition/patient-birthPlace" -> "patient-birthPlace"
func ExtractExtensionName(url string) string {
	if strings.HasPrefix(url, hl7ExtensionPrefix) {
		return strings.TrimPrefix(url, hl7ExtensionPrefix)
	}

	// For other URLs, return the last path segment
	if idx := strings.LastIndex(url, "/"); idx != -1 {
		return url[idx+1:]
	}

	return url
}
