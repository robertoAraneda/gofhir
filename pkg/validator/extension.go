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
		// Even without the extension definition, validate the value's primitive/complex type
		v.validateExtensionValueBasicType(ctx, ext, path, result)
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

// validateExtensionValueBasicType validates extension values without a StructureDefinition.
// This performs basic type validation for primitive and complex types.
func (v *Validator) validateExtensionValueBasicType(ctx context.Context, ext map[string]interface{}, path string, result *ValidationResult) {
	// Get the actual value type from the extension
	actualValueType := getExtensionValueType(ext)
	if actualValueType == "" {
		return // No value present (nested extensions instead)
	}

	// Validate the internal structure of the value against its type definition
	valueKey := "value" + actualValueType
	if value, ok := ext[valueKey]; ok {
		v.validateExtensionValueContent(ctx, value, actualValueType, path+"."+valueKey, result)
	}
}

// validateExtensionValueType validates the extension value against allowed types.
func (v *Validator) validateExtensionValueType(ctx context.Context, ext map[string]interface{}, sd *StructureDef, path string, result *ValidationResult) {
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
			return // Don't validate content if type is wrong
		}
	}

	// Validate the internal structure of the value against its type definition
	valueKey := "value" + actualValueType
	if value, ok := ext[valueKey]; ok {
		v.validateExtensionValueContent(ctx, value, actualValueType, path+"."+valueKey, result)
	}
}

// validateExtensionValueContent validates the internal structure of an extension value.
func (v *Validator) validateExtensionValueContent(ctx context.Context, value interface{}, typeName, path string, result *ValidationResult) {
	// Get the StructureDefinition for the type
	typeURL := "http://hl7.org/fhir/StructureDefinition/" + typeName
	typeDef, err := v.registry.Get(ctx, typeURL)
	if err != nil || typeDef == nil {
		// Type definition not found - can't validate deeply
		// This is expected for primitive types like "String", "Boolean", etc.
		// For primitives, just validate the value type
		v.validatePrimitiveExtensionValue(value, typeName, path, result)
		return
	}

	// For complex types, validate the structure
	valueMap, ok := value.(map[string]interface{})
	if !ok {
		// Complex type must be an object
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeStructure,
			Diagnostics: fmt.Sprintf("Expected object for type '%s', got %T", typeName, value),
			Expression:  []string{path},
		})
		return
	}

	// Build element index for the type
	index := make(map[string]*ElementDef)
	for i := range typeDef.Snapshot {
		elem := &typeDef.Snapshot[i]
		index[elem.Path] = elem
	}

	// Validate each field in the value
	v.validateExtensionFields(ctx, valueMap, typeName, path, index, result)

	// Check required fields
	v.validateExtensionRequiredFields(typeDef, valueMap, typeName, path, result)
}

// validateExtensionFields validates all fields in an extension value.
func (v *Validator) validateExtensionFields(ctx context.Context, valueMap map[string]interface{}, typeName, path string, index map[string]*ElementDef, result *ValidationResult) {
	for fieldName, fieldValue := range valueMap {
		// Skip extension and id fields (they're always allowed)
		if fieldName == "extension" || fieldName == "id" || fieldName == "_"+fieldName {
			continue
		}

		fieldPath := typeName + "." + fieldName
		elemDef := v.findElementDefForType(index, fieldPath)

		if elemDef == nil {
			// Unknown field
			if v.options.StrictMode {
				result.AddIssue(ValidationIssue{
					Severity:    SeverityError,
					Code:        IssueCodeStructure,
					Diagnostics: fmt.Sprintf("Unknown element '%s' in type '%s'", fieldName, typeName),
					Expression:  []string{path + "." + fieldName},
				})
			}
			continue
		}

		// Validate the field type
		v.validateExtensionFieldType(ctx, fieldValue, elemDef, path+"."+fieldName, result)
	}
}

// validateExtensionRequiredFields checks that all required fields are present.
func (v *Validator) validateExtensionRequiredFields(typeDef *StructureDef, valueMap map[string]interface{}, typeName, path string, result *ValidationResult) {
	for i := range typeDef.Snapshot {
		elem := &typeDef.Snapshot[i]
		if elem.Min > 0 && elem.Path != typeName {
			// This is a required field
			fieldName := strings.TrimPrefix(elem.Path, typeName+".")
			if !strings.Contains(fieldName, ".") { // Only check direct children
				if _, ok := valueMap[fieldName]; !ok {
					result.AddIssue(ValidationIssue{
						Severity:    SeverityError,
						Code:        IssueCodeRequired,
						Diagnostics: fmt.Sprintf("Missing required element '%s' in type '%s'", fieldName, typeName),
						Expression:  []string{path},
					})
				}
			}
		}
	}
}

// validatePrimitiveExtensionValue validates primitive type values in extensions.
func (v *Validator) validatePrimitiveExtensionValue(value interface{}, typeName, path string, result *ValidationResult) {
	switch strings.ToLower(typeName) {
	case "string", "code", "id", "markdown", "uri", "url", "canonical", "oid", "uuid":
		if _, ok := value.(string); !ok {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Expected string for type '%s', got %T", typeName, value),
				Expression:  []string{path},
			})
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Expected boolean for type '%s', got %T", typeName, value),
				Expression:  []string{path},
			})
		}
	case "integer", "positiveint", "unsignedint":
		switch v := value.(type) {
		case float64:
			if v != float64(int(v)) {
				result.AddIssue(ValidationIssue{
					Severity:    SeverityError,
					Code:        IssueCodeValue,
					Diagnostics: fmt.Sprintf("Expected integer for type '%s', got decimal", typeName),
					Expression:  []string{path},
				})
			}
		default:
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Expected integer for type '%s', got %T", typeName, value),
				Expression:  []string{path},
			})
		}
	case "decimal":
		if _, ok := value.(float64); !ok {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Expected number for type '%s', got %T", typeName, value),
				Expression:  []string{path},
			})
		}
	case "date", "datetime", "time", "instant":
		if _, ok := value.(string); !ok {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Expected string for type '%s', got %T", typeName, value),
				Expression:  []string{path},
			})
		}
	case "base64binary":
		if _, ok := value.(string); !ok {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Expected string for type '%s', got %T", typeName, value),
				Expression:  []string{path},
			})
		}
	}
	// For other types (complex types), validation is handled by validateExtensionValueContent
}

// findElementDefForType finds an element definition within a type's snapshot.
func (v *Validator) findElementDefForType(index map[string]*ElementDef, path string) *ElementDef {
	// Direct match
	if elem, ok := index[path]; ok {
		return elem
	}

	// Try choice type
	parts := strings.Split(path, ".")
	if len(parts) >= 2 {
		lastPart := parts[len(parts)-1]
		for _, suffix := range []string{"String", "Boolean", "Integer", "Decimal", "DateTime", "Date", "Time",
			"Code", "Uri", "Url", "Canonical", "Reference", "CodeableConcept", "Coding", "Quantity",
			"Period", "Range", "Ratio", "Identifier", "HumanName", "Address", "ContactPoint",
			"Attachment", "Annotation", "Signature", "Money", "Age", "Duration", "Count", "Distance"} {
			if strings.HasSuffix(lastPart, suffix) {
				baseName := strings.TrimSuffix(lastPart, suffix)
				choicePath := strings.Join(parts[:len(parts)-1], ".") + "." + baseName + "[x]"
				if elem, ok := index[choicePath]; ok {
					return elem
				}
			}
		}
	}

	return nil
}

// validateExtensionFieldType validates the type of a field within an extension value.
func (v *Validator) validateExtensionFieldType(ctx context.Context, value interface{}, elemDef *ElementDef, path string, result *ValidationResult) {
	if len(elemDef.Types) == 0 {
		return
	}

	expectedType := elemDef.Types[0].Code

	// Check if value matches expected type
	switch expectedType {
	case "string", "code", "id", "markdown", "uri", "url", "canonical", "oid", "uuid":
		if _, ok := value.(string); !ok {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Expected string for '%s', got %T", path, value),
				Expression:  []string{path},
			})
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Expected boolean for '%s', got %T", path, value),
				Expression:  []string{path},
			})
		}
	case "integer", "positiveInt", "unsignedInt":
		if num, ok := value.(float64); ok {
			if num != float64(int(num)) {
				result.AddIssue(ValidationIssue{
					Severity:    SeverityError,
					Code:        IssueCodeValue,
					Diagnostics: fmt.Sprintf("Expected integer for '%s', got decimal", path),
					Expression:  []string{path},
				})
			}
		} else {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Expected integer for '%s', got %T", path, value),
				Expression:  []string{path},
			})
		}
	case "decimal":
		if _, ok := value.(float64); !ok {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeValue,
				Diagnostics: fmt.Sprintf("Expected number for '%s', got %T", path, value),
				Expression:  []string{path},
			})
		}
	default:
		// For complex types, recursively validate
		switch typedValue := value.(type) {
		case map[string]interface{}:
			v.validateExtensionValueContent(ctx, typedValue, expectedType, path, result)
		case []interface{}:
			// Arrays are handled at a higher level
		default:
			if expectedType != "" && !isPrimitiveType(expectedType) {
				result.AddIssue(ValidationIssue{
					Severity:    SeverityError,
					Code:        IssueCodeStructure,
					Diagnostics: fmt.Sprintf("Expected object for '%s' of type '%s', got %T", path, expectedType, value),
					Expression:  []string{path},
				})
			}
		}
	}
}

// isPrimitiveType returns true if the type is a FHIR primitive type.
func isPrimitiveType(typeName string) bool {
	primitives := map[string]bool{
		"boolean": true, "integer": true, "string": true, "decimal": true,
		"uri": true, "url": true, "canonical": true, "base64Binary": true,
		"instant": true, "date": true, "dateTime": true, "time": true,
		"code": true, "oid": true, "id": true, "markdown": true,
		"unsignedInt": true, "positiveInt": true, "uuid": true,
	}
	return primitives[typeName]
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
