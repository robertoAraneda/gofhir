// Package validator provides FHIR resource validation based on StructureDefinitions.
package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Bundle resource type constant.
const ResourceTypeBundle = "Bundle"

// Bundle type constants as defined in FHIR specification.
const (
	BundleTypeDocument            = "document"
	BundleTypeMessage             = "message"
	BundleTypeTransaction         = "transaction"
	BundleTypeTransactionResponse = "transaction-response"
	BundleTypeBatch               = "batch"
	BundleTypeBatchResponse       = "batch-response"
	BundleTypeHistory             = "history"
	BundleTypeSearchset           = "searchset"
	BundleTypeCollection          = "collection"
)

// validBundleTypes is a set of valid Bundle.type values for O(1) lookup.
var validBundleTypes = map[string]bool{
	BundleTypeDocument:            true,
	BundleTypeMessage:             true,
	BundleTypeTransaction:         true,
	BundleTypeTransactionResponse: true,
	BundleTypeBatch:               true,
	BundleTypeBatchResponse:       true,
	BundleTypeHistory:             true,
	BundleTypeSearchset:           true,
	BundleTypeCollection:          true,
}

// bundleTypesRequiringRequest are types where entry.request is mandatory.
var bundleTypesRequiringRequest = map[string]bool{
	BundleTypeTransaction: true,
	BundleTypeBatch:       true,
	BundleTypeHistory:     true,
}

// bundleTypesRequiringResponse are types where entry.response is mandatory.
var bundleTypesRequiringResponse = map[string]bool{
	BundleTypeTransactionResponse: true,
	BundleTypeBatchResponse:       true,
	BundleTypeHistory:             true,
}

// bundleTypesAllowingTotal are types where Bundle.total is allowed.
var bundleTypesAllowingTotal = map[string]bool{
	BundleTypeSearchset: true,
	BundleTypeHistory:   true,
}

// bundleTypesAllowingSearch are types where entry.search is allowed.
var bundleTypesAllowingSearch = map[string]bool{
	BundleTypeSearchset: true,
}

// validateBundle performs Bundle-specific validation after standard validation.
// This method is called automatically by Validate() when resourceType is "Bundle".
func (v *Validator) validateBundle(ctx context.Context, vctx *validationContext, result *ValidationResult) {
	bundle := vctx.parsed

	// Get Bundle.type (required field - already validated by structure validation)
	bundleType, _ := bundle["type"].(string)
	if bundleType == "" {
		// Type is required but missing - structure validation already reported this
		return
	}

	// Validate Bundle.type is a valid value
	if !validBundleTypes[bundleType] {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeCodeInvalid,
			Diagnostics: fmt.Sprintf("Invalid Bundle.type: '%s'", bundleType),
			Expression:  []string{"Bundle.type"},
		})
		return
	}

	// Apply Bundle-specific constraints (bdl-*)
	v.validateBundleConstraints(ctx, bundle, bundleType, result)

	// Validate each entry and its resource recursively
	v.validateBundleEntries(ctx, vctx, bundle, bundleType, result)
}

// validateBundleConstraints validates Bundle-level constraints (bdl-1, bdl-2, bdl-9, bdl-10).
func (v *Validator) validateBundleConstraints(_ context.Context, bundle map[string]interface{}, bundleType string, result *ValidationResult) {
	// bdl-1: total only when a search or history
	if _, hasTotal := bundle["total"]; hasTotal {
		if !bundleTypesAllowingTotal[bundleType] {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeInvariant,
				Diagnostics: fmt.Sprintf("Constraint bdl-1 violated: Bundle.total is only allowed for searchset or history bundles, not '%s'", bundleType),
				Expression:  []string{"Bundle.total"},
			})
		}
	}

	// bdl-9: A document must have an identifier with a system and a value
	if bundleType == BundleTypeDocument {
		v.validateDocumentIdentifier(bundle, result)
	}

	// bdl-10: A document must have a date
	if bundleType == BundleTypeDocument {
		if _, hasTimestamp := bundle["timestamp"]; !hasTimestamp {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeInvariant,
				Diagnostics: "Constraint bdl-10 violated: A document Bundle must have a timestamp",
				Expression:  []string{"Bundle.timestamp"},
			})
		}
	}
}

// validateDocumentIdentifier validates bdl-9: document identifier requirements.
func (v *Validator) validateDocumentIdentifier(bundle map[string]interface{}, result *ValidationResult) {
	identifier, hasIdentifier := bundle["identifier"]
	if !hasIdentifier {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: "Constraint bdl-9 violated: A document Bundle must have an identifier",
			Expression:  []string{"Bundle.identifier"},
		})
		return
	}

	identifierMap, ok := identifier.(map[string]interface{})
	if !ok {
		return
	}

	system, hasSystem := identifierMap["system"].(string)
	value, hasValue := identifierMap["value"].(string)

	if !hasSystem || system == "" {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: "Constraint bdl-9 violated: A document Bundle identifier must have a system",
			Expression:  []string{"Bundle.identifier.system"},
		})
	}

	if !hasValue || value == "" {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: "Constraint bdl-9 violated: A document Bundle identifier must have a value",
			Expression:  []string{"Bundle.identifier.value"},
		})
	}
}

// validateBundleEntries validates all Bundle.entry elements.
func (v *Validator) validateBundleEntries(ctx context.Context, vctx *validationContext, bundle map[string]interface{}, bundleType string, result *ValidationResult) {
	entries, ok := bundle["entry"].([]interface{})
	if !ok || len(entries) == 0 {
		return
	}

	// Track fullURLs for uniqueness validation (bdl-7)
	fullURLSet := make(map[string]bool)

	for i, entry := range entries {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}

		entryPath := fmt.Sprintf("Bundle.entry[%d]", i)
		v.validateBundleEntry(ctx, vctx, entryMap, entryPath, bundleType, i, fullURLSet, result)
	}

	// bdl-11: Document bundle first entry must be Composition
	if bundleType == BundleTypeDocument && len(entries) > 0 {
		v.validateDocumentFirstEntry(entries[0], result)
	}

	// bdl-12: Message bundle first entry must be MessageHeader
	if bundleType == BundleTypeMessage && len(entries) > 0 {
		v.validateMessageFirstEntry(entries[0], result)
	}
}

// validateBundleEntry validates a single Bundle.entry element.
func (v *Validator) validateBundleEntry(
	ctx context.Context,
	vctx *validationContext,
	entry map[string]interface{},
	entryPath string,
	bundleType string,
	_ int, // entryIndex - reserved for future use
	fullURLSet map[string]bool,
	result *ValidationResult,
) {
	resource, hasResource := entry["resource"].(map[string]interface{})
	request, hasRequest := entry["request"].(map[string]interface{})
	response, hasResponse := entry["response"].(map[string]interface{})
	search, hasSearch := entry["search"].(map[string]interface{})
	fullURL, hasFullURL := entry["fullUrl"].(string)

	// bdl-5: must be a resource unless there's a request or response
	if !hasResource && !hasRequest && !hasResponse {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: "Constraint bdl-5 violated: Bundle entry must have a resource, request, or response",
			Expression:  []string{entryPath},
		})
	}

	// bdl-7: fullUrl must be unique (except for history bundles)
	if hasFullURL && bundleType != BundleTypeHistory {
		v.validateFullURLUniqueness(entry, entryPath, fullURL, fullURLSet, result)
	}

	// bdl-8: fullUrl cannot be a version specific reference
	if hasFullURL && strings.Contains(fullURL, "/_history/") {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: "Constraint bdl-8 violated: fullUrl cannot be a version specific reference (contains /_history/)",
			Expression:  []string{entryPath + ".fullUrl"},
		})
	}

	// bdl-2: entry.search only when a search
	if hasSearch && !bundleTypesAllowingSearch[bundleType] {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: fmt.Sprintf("Constraint bdl-2 violated: entry.search is only allowed in searchset bundles, not '%s'", bundleType),
			Expression:  []string{entryPath + ".search"},
		})
	}

	// bdl-3: entry.request mandatory for batch/transaction/history, otherwise prohibited
	v.validateEntryRequest(entry, entryPath, bundleType, hasRequest, request, result)

	// bdl-4: entry.response mandatory for batch-response/transaction-response/history, otherwise prohibited
	v.validateEntryResponse(entry, entryPath, bundleType, hasResponse, response, result)

	// Validate search element if present
	if hasSearch {
		v.validateEntrySearch(search, entryPath, result)
	}

	// Recursively validate entry.resource if present and option enabled
	if hasResource {
		v.validateEntryResource(ctx, vctx, resource, entryPath, result)
	}
}

// validateFullURLUniqueness validates bdl-7: fullUrl uniqueness.
func (v *Validator) validateFullURLUniqueness(entry map[string]interface{}, entryPath, fullURL string, fullURLSet map[string]bool, result *ValidationResult) {
	// For uniqueness check, combine fullUrl with versionId if present
	uniqueKey := fullURL

	if resource, ok := entry["resource"].(map[string]interface{}); ok {
		if meta, ok := resource["meta"].(map[string]interface{}); ok {
			if versionID, ok := meta["versionId"].(string); ok && versionID != "" {
				uniqueKey = fullURL + "&" + versionID
			}
		}
	}

	if fullURLSet[uniqueKey] {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: fmt.Sprintf("Constraint bdl-7 violated: duplicate fullUrl '%s' in bundle", fullURL),
			Expression:  []string{entryPath + ".fullUrl"},
		})
	}
	fullURLSet[uniqueKey] = true
}

// validateEntryRequest validates bdl-3: request presence rules.
func (v *Validator) validateEntryRequest(_ map[string]interface{}, entryPath, bundleType string, hasRequest bool, request map[string]interface{}, result *ValidationResult) {
	requiresRequest := bundleTypesRequiringRequest[bundleType]

	if requiresRequest && !hasRequest {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: fmt.Sprintf("Constraint bdl-3 violated: entry.request is required for '%s' bundles", bundleType),
			Expression:  []string{entryPath + ".request"},
		})
	} else if !requiresRequest && hasRequest && bundleType != BundleTypeHistory {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: fmt.Sprintf("Constraint bdl-3 violated: entry.request is not allowed for '%s' bundles", bundleType),
			Expression:  []string{entryPath + ".request"},
		})
	}

	// Validate request content if present
	if hasRequest && request != nil {
		v.validateRequestContent(request, entryPath, result)
	}
}

// validateEntryResponse validates bdl-4: response presence rules.
func (v *Validator) validateEntryResponse(_ map[string]interface{}, entryPath, bundleType string, hasResponse bool, response map[string]interface{}, result *ValidationResult) {
	requiresResponse := bundleTypesRequiringResponse[bundleType]

	if requiresResponse && !hasResponse {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: fmt.Sprintf("Constraint bdl-4 violated: entry.response is required for '%s' bundles", bundleType),
			Expression:  []string{entryPath + ".response"},
		})
	} else if !requiresResponse && hasResponse && bundleType != BundleTypeHistory {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: fmt.Sprintf("Constraint bdl-4 violated: entry.response is not allowed for '%s' bundles", bundleType),
			Expression:  []string{entryPath + ".response"},
		})
	}

	// Validate response content if present
	if hasResponse && response != nil {
		v.validateResponseContent(response, entryPath, result)
	}
}

// validateRequestContent validates entry.request required fields.
func (v *Validator) validateRequestContent(request map[string]interface{}, entryPath string, result *ValidationResult) {
	method, hasMethod := request["method"].(string)
	requestURL, hasURL := request["url"].(string)

	if !hasMethod || method == "" {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeRequired,
			Diagnostics: "Bundle.entry.request.method is required",
			Expression:  []string{entryPath + ".request.method"},
		})
	} else {
		// Validate method is a valid HTTP verb
		validMethods := map[string]bool{
			"GET": true, "HEAD": true, "POST": true,
			"PUT": true, "DELETE": true, "PATCH": true,
		}
		if !validMethods[method] {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeCodeInvalid,
				Diagnostics: fmt.Sprintf("Invalid request method: '%s'", method),
				Expression:  []string{entryPath + ".request.method"},
			})
		}
	}

	if !hasURL || requestURL == "" {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeRequired,
			Diagnostics: "Bundle.entry.request.url is required",
			Expression:  []string{entryPath + ".request.url"},
		})
	}
}

// validateResponseContent validates entry.response required fields.
func (v *Validator) validateResponseContent(response map[string]interface{}, entryPath string, result *ValidationResult) {
	status, hasStatus := response["status"].(string)

	if !hasStatus || status == "" {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeRequired,
			Diagnostics: "Bundle.entry.response.status is required",
			Expression:  []string{entryPath + ".response.status"},
		})
	}
}

// validateEntrySearch validates entry.search content.
func (v *Validator) validateEntrySearch(search map[string]interface{}, entryPath string, result *ValidationResult) {
	if mode, hasMode := search["mode"].(string); hasMode {
		validModes := map[string]bool{
			"match": true, "include": true, "outcome": true,
		}
		if !validModes[mode] {
			result.AddIssue(ValidationIssue{
				Severity:    SeverityError,
				Code:        IssueCodeCodeInvalid,
				Diagnostics: fmt.Sprintf("Invalid search mode: '%s'", mode),
				Expression:  []string{entryPath + ".search.mode"},
			})
		}
	}

	if score, hasScore := search["score"]; hasScore {
		if scoreFloat, ok := score.(float64); ok {
			if scoreFloat < 0 || scoreFloat > 1 {
				result.AddIssue(ValidationIssue{
					Severity:    SeverityError,
					Code:        IssueCodeValue,
					Diagnostics: "search.score must be between 0 and 1",
					Expression:  []string{entryPath + ".search.score"},
				})
			}
		}
	}
}

// validateEntryResource recursively validates the resource within an entry.
func (v *Validator) validateEntryResource(ctx context.Context, vctx *validationContext, resource map[string]interface{}, entryPath string, result *ValidationResult) {
	resourceType, ok := resource["resourceType"].(string)
	if !ok || resourceType == "" {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeRequired,
			Diagnostics: "Bundle entry resource must have a resourceType",
			Expression:  []string{entryPath + ".resource.resourceType"},
		})
		return
	}

	// Get StructureDefinition for the resource type
	sd, err := v.registry.GetByType(ctx, resourceType)
	if err != nil {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeNotFound,
			Diagnostics: fmt.Sprintf("Unknown resource type in entry: %s", resourceType),
			Expression:  []string{entryPath + ".resource"},
		})
		return
	}

	// Create a new validation context for the nested resource
	nestedIndex := v.buildElementIndex(sd)
	nestedVctx := &validationContext{
		raw:          vctx.raw, // Keep original raw for reference resolution
		parsed:       resource,
		resourceType: resourceType,
		sd:           sd,
		index:        nestedIndex,
	}

	// Track present elements for structure validation
	presentElements := make(map[string]bool)

	// Validate structure recursively
	v.validateNode(ctx, resource, sd, nestedIndex, resourceType, "", presentElements, result)

	// Validate primitives
	v.validatePrimitiveNode(ctx, resource, nestedIndex, resourceType, result)

	// Validate ele-1
	v.checkEle1Recursive(resource, entryPath+".resource", result)

	// Validate constraints if enabled
	if v.options.ValidateConstraints {
		v.validateNestedConstraints(ctx, nestedVctx, entryPath, result)
	}

	// Validate terminology if enabled
	if v.options.ValidateTerminology {
		v.validateTerminology(ctx, nestedVctx, result)
	}

	// Validate extensions if enabled
	if v.options.ValidateExtensions {
		v.validateExtensions(ctx, nestedVctx, result)
	}

	// Recursively validate nested Bundles
	if resourceType == ResourceTypeBundle {
		v.validateBundle(ctx, nestedVctx, result)
	}
}

// validateNestedConstraints validates FHIRPath constraints for nested resources.
func (v *Validator) validateNestedConstraints(_ context.Context, vctx *validationContext, basePath string, result *ValidationResult) {
	for _, elem := range vctx.sd.Snapshot {
		for _, constraint := range elem.Constraints {
			if constraint.Expression == "" {
				continue
			}

			if constraint.Source != "" && constraint.Source != vctx.sd.URL {
				continue
			}

			if elem.Path != vctx.resourceType && !elementExistsInResource(vctx.parsed, elem.Path, vctx.resourceType) {
				continue
			}

			// For nested resources, we need to marshal back to JSON for FHIRPath evaluation
			// This is a performance tradeoff for correctness
			valid, err := v.evaluateConstraintOnParsed(vctx.parsed, elem.Path, vctx.resourceType, constraint)
			if err != nil {
				result.AddIssue(ValidationIssue{
					Severity:    SeverityWarning,
					Code:        IssueCodeProcessing,
					Diagnostics: fmt.Sprintf("Failed to evaluate constraint %s on %s: %v", constraint.Key, basePath+"."+elem.Path, err),
					Expression:  []string{basePath + "." + elem.Path},
				})
				continue
			}

			if !valid {
				severity := SeverityError
				if constraint.Severity == "warning" {
					severity = SeverityWarning
				}

				result.AddIssue(ValidationIssue{
					Severity:    severity,
					Code:        IssueCodeInvariant,
					Diagnostics: fmt.Sprintf("Constraint %s violated: %s", constraint.Key, constraint.Human),
					Expression:  []string{basePath + "." + elem.Path},
				})
			}
		}
	}
}

// evaluateConstraintOnParsed evaluates a FHIRPath constraint on a parsed resource map.
func (v *Validator) evaluateConstraintOnParsed(resource map[string]interface{}, elementPath, resourceType string, constraint ElementConstraint) (bool, error) {
	// Marshal back to JSON for FHIRPath evaluation
	// This is necessary because our FHIRPath engine works with JSON bytes
	jsonBytes, err := json.Marshal(resource)
	if err != nil {
		return false, fmt.Errorf("failed to marshal resource: %w", err)
	}

	return v.evaluateConstraint(jsonBytes, elementPath, resourceType, constraint)
}

// validateDocumentFirstEntry validates bdl-11: first entry must be Composition.
func (v *Validator) validateDocumentFirstEntry(firstEntry interface{}, result *ValidationResult) {
	entry, ok := firstEntry.(map[string]interface{})
	if !ok {
		return
	}

	resource, ok := entry["resource"].(map[string]interface{})
	if !ok {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: "Constraint bdl-11 violated: document Bundle first entry must have a resource",
			Expression:  []string{"Bundle.entry[0].resource"},
		})
		return
	}

	resourceType, _ := resource["resourceType"].(string)
	if resourceType != "Composition" {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: fmt.Sprintf("Constraint bdl-11 violated: document Bundle first entry must be a Composition, got '%s'", resourceType),
			Expression:  []string{"Bundle.entry[0].resource"},
		})
	}
}

// validateMessageFirstEntry validates bdl-12: first entry must be MessageHeader.
func (v *Validator) validateMessageFirstEntry(firstEntry interface{}, result *ValidationResult) {
	entry, ok := firstEntry.(map[string]interface{})
	if !ok {
		return
	}

	resource, ok := entry["resource"].(map[string]interface{})
	if !ok {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: "Constraint bdl-12 violated: message Bundle first entry must have a resource",
			Expression:  []string{"Bundle.entry[0].resource"},
		})
		return
	}

	resourceType, _ := resource["resourceType"].(string)
	if resourceType != "MessageHeader" {
		result.AddIssue(ValidationIssue{
			Severity:    SeverityError,
			Code:        IssueCodeInvariant,
			Diagnostics: fmt.Sprintf("Constraint bdl-12 violated: message Bundle first entry must be a MessageHeader, got '%s'", resourceType),
			Expression:  []string{"Bundle.entry[0].resource"},
		})
	}
}
