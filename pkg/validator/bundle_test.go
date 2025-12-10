package validator

import (
	"context"
	"strings"
	"testing"
)

// ============================================================================
// Bundle Type Validation Tests
// ============================================================================

func TestValidateBundleType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		bundleType  string
		expectError bool
	}{
		{"document", BundleTypeDocument, false},
		{"message", BundleTypeMessage, false},
		{"transaction", BundleTypeTransaction, false},
		{"transaction-response", BundleTypeTransactionResponse, false},
		{"batch", BundleTypeBatch, false},
		{"batch-response", BundleTypeBatchResponse, false},
		{"history", BundleTypeHistory, false},
		{"searchset", BundleTypeSearchset, false},
		{"collection", BundleTypeCollection, false},
		{"invalid-type", "invalid-type", true},
		{"empty-type", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-` + tt.name + `",
				"type": "` + tt.bundleType + `"
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasTypeError := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeCodeInvalid && strings.Contains(issue.Diagnostics, "Bundle.type") {
					hasTypeError = true
					break
				}
			}

			if tt.expectError && !hasTypeError && tt.bundleType != "" {
				t.Errorf("Expected type error for '%s', got none", tt.bundleType)
			}
			if !tt.expectError && hasTypeError {
				t.Errorf("Unexpected type error for '%s'", tt.bundleType)
			}
		})
	}
}

// ============================================================================
// bdl-1: total only when searchset or history
// ============================================================================

func TestValidateBdl1TotalConstraint(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		bundleType  string
		hasTotal    bool
		expectError bool
	}{
		// Total allowed in searchset and history
		{"searchset-with-total", BundleTypeSearchset, true, false},
		{"history-with-total", BundleTypeHistory, true, false},
		{"searchset-without-total", BundleTypeSearchset, false, false},
		{"history-without-total", BundleTypeHistory, false, false},
		// Total NOT allowed in other types
		{"collection-with-total", BundleTypeCollection, true, true},
		{"transaction-with-total", BundleTypeTransaction, true, true},
		{"batch-with-total", BundleTypeBatch, true, true},
		// Without total should be fine for any type
		{"collection-without-total", BundleTypeCollection, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalField := ""
			if tt.hasTotal {
				totalField = `"total": 10,`
			}

			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-bdl1",
				"type": "` + tt.bundleType + `",
				` + totalField + `
				"entry": []
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasBdl1Error := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "bdl-1") {
					hasBdl1Error = true
					break
				}
			}

			if tt.expectError && !hasBdl1Error {
				t.Errorf("Expected bdl-1 violation for %s with total=%v", tt.bundleType, tt.hasTotal)
			}
			if !tt.expectError && hasBdl1Error {
				t.Errorf("Unexpected bdl-1 violation for %s with total=%v", tt.bundleType, tt.hasTotal)
			}
		})
	}
}

// ============================================================================
// bdl-2: entry.search only when searchset
// ============================================================================

func TestValidateBdl2SearchConstraint(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		bundleType  string
		hasSearch   bool
		expectError bool
	}{
		// Search allowed only in searchset
		{"searchset-with-search", BundleTypeSearchset, true, false},
		{"searchset-without-search", BundleTypeSearchset, false, false},
		// Search NOT allowed in other types
		{"document-with-search", BundleTypeDocument, true, true},
		{"transaction-with-search", BundleTypeTransaction, true, true},
		{"collection-with-search", BundleTypeCollection, true, true},
		{"history-with-search", BundleTypeHistory, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchField := ""
			if tt.hasSearch {
				searchField = `"search": {"mode": "match"},`
			}

			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-bdl2",
				"type": "` + tt.bundleType + `",
				"entry": [{
					` + searchField + `
					"resource": {
						"resourceType": "Patient",
						"id": "pat1"
					}
				}]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasBdl2Error := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "bdl-2") {
					hasBdl2Error = true
					break
				}
			}

			if tt.expectError && !hasBdl2Error {
				t.Errorf("Expected bdl-2 violation for %s with search=%v", tt.bundleType, tt.hasSearch)
			}
			if !tt.expectError && hasBdl2Error {
				t.Errorf("Unexpected bdl-2 violation for %s with search=%v", tt.bundleType, tt.hasSearch)
			}
		})
	}
}

// ============================================================================
// bdl-3: entry.request mandatory for batch/transaction/history
// ============================================================================

func TestValidateBdl3RequestConstraint(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		bundleType  string
		hasRequest  bool
		expectError bool
	}{
		// Request REQUIRED for transaction, batch, history
		{"transaction-with-request", BundleTypeTransaction, true, false},
		{"transaction-without-request", BundleTypeTransaction, false, true},
		{"batch-with-request", BundleTypeBatch, true, false},
		{"batch-without-request", BundleTypeBatch, false, true},
		{"history-with-request", BundleTypeHistory, true, false},
		{"history-without-request", BundleTypeHistory, false, true},
		// Request NOT allowed in other types (except history)
		{"document-with-request", BundleTypeDocument, true, true},
		{"document-without-request", BundleTypeDocument, false, false},
		{"collection-with-request", BundleTypeCollection, true, true},
		{"collection-without-request", BundleTypeCollection, false, false},
		{"searchset-with-request", BundleTypeSearchset, true, true},
		{"searchset-without-request", BundleTypeSearchset, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestField := ""
			if tt.hasRequest {
				requestField = `"request": {"method": "POST", "url": "Patient"},`
			}

			// For history bundles, also need response
			responseField := ""
			if tt.bundleType == BundleTypeHistory {
				responseField = `"response": {"status": "200 OK"},`
			}

			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-bdl3",
				"type": "` + tt.bundleType + `",
				"entry": [{
					` + requestField + `
					` + responseField + `
					"resource": {
						"resourceType": "Patient",
						"id": "pat1"
					}
				}]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasBdl3Error := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "bdl-3") {
					hasBdl3Error = true
					break
				}
			}

			if tt.expectError && !hasBdl3Error {
				t.Errorf("Expected bdl-3 violation for %s with request=%v", tt.bundleType, tt.hasRequest)
			}
			if !tt.expectError && hasBdl3Error {
				t.Errorf("Unexpected bdl-3 violation for %s with request=%v", tt.bundleType, tt.hasRequest)
			}
		})
	}
}

// ============================================================================
// bdl-4: entry.response mandatory for batch-response/transaction-response/history
// ============================================================================

func TestValidateBdl4ResponseConstraint(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		bundleType  string
		hasResponse bool
		expectError bool
	}{
		// Response REQUIRED for transaction-response, batch-response, history
		{"transaction-response-with-response", BundleTypeTransactionResponse, true, false},
		{"transaction-response-without-response", BundleTypeTransactionResponse, false, true},
		{"batch-response-with-response", BundleTypeBatchResponse, true, false},
		{"batch-response-without-response", BundleTypeBatchResponse, false, true},
		{"history-with-response", BundleTypeHistory, true, false},
		{"history-without-response", BundleTypeHistory, false, true},
		// Response NOT allowed in other types
		{"document-with-response", BundleTypeDocument, true, true},
		{"document-without-response", BundleTypeDocument, false, false},
		{"collection-with-response", BundleTypeCollection, true, true},
		{"searchset-with-response", BundleTypeSearchset, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			responseField := ""
			if tt.hasResponse {
				responseField = `"response": {"status": "200 OK"},`
			}

			// For history bundles, also need request
			requestField := ""
			if tt.bundleType == BundleTypeHistory {
				requestField = `"request": {"method": "GET", "url": "Patient/1"},`
			}

			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-bdl4",
				"type": "` + tt.bundleType + `",
				"entry": [{
					` + requestField + `
					` + responseField + `
					"resource": {
						"resourceType": "Patient",
						"id": "pat1"
					}
				}]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasBdl4Error := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "bdl-4") {
					hasBdl4Error = true
					break
				}
			}

			if tt.expectError && !hasBdl4Error {
				t.Errorf("Expected bdl-4 violation for %s with response=%v", tt.bundleType, tt.hasResponse)
			}
			if !tt.expectError && hasBdl4Error {
				t.Errorf("Unexpected bdl-4 violation for %s with response=%v", tt.bundleType, tt.hasResponse)
			}
		})
	}
}

// ============================================================================
// bdl-5: must be a resource unless there's a request or response
// ============================================================================

func TestValidateBdl5ResourceOrRequestResponse(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		entry       string
		expectError bool
	}{
		{
			name: "entry-with-resource",
			entry: `{
				"resource": {"resourceType": "Patient", "id": "pat1"}
			}`,
			expectError: false,
		},
		{
			name: "entry-with-request-only",
			entry: `{
				"request": {"method": "DELETE", "url": "Patient/pat1"}
			}`,
			expectError: false,
		},
		{
			name: "entry-with-response-only",
			entry: `{
				"response": {"status": "200 OK"}
			}`,
			expectError: false,
		},
		{
			name:        "entry-empty",
			entry:       `{"fullUrl": "urn:uuid:test"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-bdl5",
				"type": "collection",
				"entry": [` + tt.entry + `]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasBdl5Error := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "bdl-5") {
					hasBdl5Error = true
					break
				}
			}

			if tt.expectError && !hasBdl5Error {
				t.Errorf("Expected bdl-5 violation for %s", tt.name)
			}
			if !tt.expectError && hasBdl5Error {
				t.Errorf("Unexpected bdl-5 violation for %s", tt.name)
			}
		})
	}
}

// ============================================================================
// bdl-7: fullUrl must be unique
// ============================================================================

func TestValidateBdl7FullUrlUniqueness(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		bundleType  string
		entries     string
		expectError bool
	}{
		{
			name:       "unique-fullurls",
			bundleType: BundleTypeCollection,
			entries: `[
				{"fullUrl": "urn:uuid:1", "resource": {"resourceType": "Patient", "id": "p1"}},
				{"fullUrl": "urn:uuid:2", "resource": {"resourceType": "Patient", "id": "p2"}}
			]`,
			expectError: false,
		},
		{
			name:       "duplicate-fullurls",
			bundleType: BundleTypeCollection,
			entries: `[
				{"fullUrl": "urn:uuid:same", "resource": {"resourceType": "Patient", "id": "p1"}},
				{"fullUrl": "urn:uuid:same", "resource": {"resourceType": "Patient", "id": "p2"}}
			]`,
			expectError: true,
		},
		{
			name:       "history-allows-duplicate-fullurls",
			bundleType: BundleTypeHistory,
			entries: `[
				{
					"fullUrl": "http://example.org/Patient/1",
					"resource": {"resourceType": "Patient", "id": "1", "meta": {"versionId": "1"}},
					"request": {"method": "POST", "url": "Patient"},
					"response": {"status": "201 Created"}
				},
				{
					"fullUrl": "http://example.org/Patient/1",
					"resource": {"resourceType": "Patient", "id": "1", "meta": {"versionId": "2"}},
					"request": {"method": "PUT", "url": "Patient/1"},
					"response": {"status": "200 OK"}
				}
			]`,
			expectError: false,
		},
		{
			name:       "same-fullurl-different-versionid",
			bundleType: BundleTypeCollection,
			entries: `[
				{
					"fullUrl": "http://example.org/Patient/1",
					"resource": {"resourceType": "Patient", "id": "1", "meta": {"versionId": "1"}}
				},
				{
					"fullUrl": "http://example.org/Patient/1",
					"resource": {"resourceType": "Patient", "id": "1", "meta": {"versionId": "2"}}
				}
			]`,
			expectError: false, // Different versionId makes them unique
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-bdl7",
				"type": "` + tt.bundleType + `",
				"entry": ` + tt.entries + `
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasBdl7Error := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "bdl-7") {
					hasBdl7Error = true
					break
				}
			}

			if tt.expectError && !hasBdl7Error {
				t.Errorf("Expected bdl-7 violation for %s", tt.name)
			}
			if !tt.expectError && hasBdl7Error {
				t.Errorf("Unexpected bdl-7 violation for %s", tt.name)
			}
		})
	}
}

// ============================================================================
// bdl-8: fullUrl cannot be version specific reference
// ============================================================================

func TestValidateBdl8FullUrlNoVersionSpecific(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		fullURL     string
		expectError bool
	}{
		{"valid-fullurl", "http://example.org/Patient/123", false},
		{"valid-urn-uuid", "urn:uuid:12345678-1234-1234-1234-123456789012", false},
		{"valid-urn-oid", "urn:oid:1.2.3.4.5", false},
		{"invalid-version-specific", "http://example.org/Patient/123/_history/1", true},
		{"invalid-version-specific-mid", "http://example.org/_history/1/Patient/123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-bdl8",
				"type": "collection",
				"entry": [{
					"fullUrl": "` + tt.fullURL + `",
					"resource": {"resourceType": "Patient", "id": "123"}
				}]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasBdl8Error := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "bdl-8") {
					hasBdl8Error = true
					break
				}
			}

			if tt.expectError && !hasBdl8Error {
				t.Errorf("Expected bdl-8 violation for fullUrl '%s'", tt.fullURL)
			}
			if !tt.expectError && hasBdl8Error {
				t.Errorf("Unexpected bdl-8 violation for fullUrl '%s'", tt.fullURL)
			}
		})
	}
}

// ============================================================================
// bdl-9: Document must have identifier with system and value
// ============================================================================

func TestValidateBdl9DocumentIdentifier(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		identifier    string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid-identifier",
			identifier:  `"identifier": {"system": "urn:ietf:rfc:3986", "value": "urn:uuid:12345"},`,
			expectError: false,
		},
		{
			name:          "missing-identifier",
			identifier:    "",
			expectError:   true,
			errorContains: "must have an identifier",
		},
		{
			name:          "missing-system",
			identifier:    `"identifier": {"value": "12345"},`,
			expectError:   true,
			errorContains: "bdl-9",
		},
		{
			name:          "missing-value",
			identifier:    `"identifier": {"system": "urn:ietf:rfc:3986"},`,
			expectError:   true,
			errorContains: "bdl-9",
		},
		{
			name:          "empty-system",
			identifier:    `"identifier": {"system": "", "value": "12345"},`,
			expectError:   true,
			errorContains: "must have a system",
		},
		{
			name:          "empty-value",
			identifier:    `"identifier": {"system": "urn:ietf:rfc:3986", "value": ""},`,
			expectError:   true,
			errorContains: "must have a value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-bdl9",
				"type": "document",
				` + tt.identifier + `
				"timestamp": "2024-01-15T10:00:00Z",
				"entry": [{
					"fullUrl": "urn:uuid:composition-1",
					"resource": {
						"resourceType": "Composition",
						"id": "comp1",
						"status": "final",
						"type": {"coding": [{"system": "http://loinc.org", "code": "11503-0"}]},
						"subject": {"reference": "Patient/pat1"},
						"date": "2024-01-15",
						"author": [{"reference": "Practitioner/prac1"}],
						"title": "Test Document"
					}
				}]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasBdl9Error := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "bdl-9") {
					hasBdl9Error = true
					if tt.errorContains != "" && !strings.Contains(issue.Diagnostics, tt.errorContains) {
						t.Errorf("Expected error containing '%s', got '%s'", tt.errorContains, issue.Diagnostics)
					}
					break
				}
			}

			if tt.expectError && !hasBdl9Error {
				t.Errorf("Expected bdl-9 violation for %s", tt.name)
			}
			if !tt.expectError && hasBdl9Error {
				t.Errorf("Unexpected bdl-9 violation for %s", tt.name)
			}
		})
	}
}

// ============================================================================
// bdl-10: Document must have timestamp
// ============================================================================

func TestValidateBdl10DocumentTimestamp(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		timestamp   string
		expectError bool
	}{
		{
			name:        "valid-timestamp",
			timestamp:   `"timestamp": "2024-01-15T10:00:00Z",`,
			expectError: false,
		},
		{
			name:        "missing-timestamp",
			timestamp:   "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-bdl10",
				"type": "document",
				"identifier": {"system": "urn:ietf:rfc:3986", "value": "urn:uuid:12345"},
				` + tt.timestamp + `
				"entry": [{
					"fullUrl": "urn:uuid:composition-1",
					"resource": {
						"resourceType": "Composition",
						"id": "comp1",
						"status": "final",
						"type": {"coding": [{"system": "http://loinc.org", "code": "11503-0"}]},
						"subject": {"reference": "Patient/pat1"},
						"date": "2024-01-15",
						"author": [{"reference": "Practitioner/prac1"}],
						"title": "Test Document"
					}
				}]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasBdl10Error := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "bdl-10") {
					hasBdl10Error = true
					break
				}
			}

			if tt.expectError && !hasBdl10Error {
				t.Errorf("Expected bdl-10 violation for %s", tt.name)
			}
			if !tt.expectError && hasBdl10Error {
				t.Errorf("Unexpected bdl-10 violation for %s", tt.name)
			}
		})
	}
}

// ============================================================================
// bdl-11: Document first entry must be Composition
// ============================================================================

func TestValidateBdl11DocumentFirstEntryComposition(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		firstResource string
		expectError   bool
	}{
		{
			name: "valid-composition-first",
			firstResource: `{
				"resourceType": "Composition",
				"id": "comp1",
				"status": "final",
				"type": {"coding": [{"system": "http://loinc.org", "code": "11503-0"}]},
				"subject": {"reference": "Patient/pat1"},
				"date": "2024-01-15",
				"author": [{"reference": "Practitioner/prac1"}],
				"title": "Test Document"
			}`,
			expectError: false,
		},
		{
			name: "invalid-patient-first",
			firstResource: `{
				"resourceType": "Patient",
				"id": "pat1"
			}`,
			expectError: true,
		},
		{
			name: "invalid-observation-first",
			firstResource: `{
				"resourceType": "Observation",
				"id": "obs1",
				"status": "final",
				"code": {"coding": [{"system": "http://loinc.org", "code": "1234-5"}]}
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-bdl11",
				"type": "document",
				"identifier": {"system": "urn:ietf:rfc:3986", "value": "urn:uuid:12345"},
				"timestamp": "2024-01-15T10:00:00Z",
				"entry": [{
					"fullUrl": "urn:uuid:first-entry",
					"resource": ` + tt.firstResource + `
				}]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasBdl11Error := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "bdl-11") {
					hasBdl11Error = true
					break
				}
			}

			if tt.expectError && !hasBdl11Error {
				t.Errorf("Expected bdl-11 violation for %s", tt.name)
			}
			if !tt.expectError && hasBdl11Error {
				t.Errorf("Unexpected bdl-11 violation for %s", tt.name)
			}
		})
	}
}

// ============================================================================
// bdl-12: Message first entry must be MessageHeader
// ============================================================================

func TestValidateBdl12MessageFirstEntryMessageHeader(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		firstResource string
		expectError   bool
	}{
		{
			name: "valid-messageheader-first",
			firstResource: `{
				"resourceType": "MessageHeader",
				"id": "mh1",
				"eventCoding": {"system": "http://example.org", "code": "admin-notify"},
				"source": {"endpoint": "http://example.org/source"}
			}`,
			expectError: false,
		},
		{
			name: "invalid-patient-first",
			firstResource: `{
				"resourceType": "Patient",
				"id": "pat1"
			}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-bdl12",
				"type": "message",
				"entry": [{
					"fullUrl": "urn:uuid:first-entry",
					"resource": ` + tt.firstResource + `
				}]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasBdl12Error := false
			for _, issue := range result.Issues {
				if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "bdl-12") {
					hasBdl12Error = true
					break
				}
			}

			if tt.expectError && !hasBdl12Error {
				t.Errorf("Expected bdl-12 violation for %s", tt.name)
			}
			if !tt.expectError && hasBdl12Error {
				t.Errorf("Unexpected bdl-12 violation for %s", tt.name)
			}
		})
	}
}

// ============================================================================
// Entry request/response validation
// ============================================================================

func TestValidateEntryRequestContent(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		request       string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid-request",
			request:     `{"method": "POST", "url": "Patient"}`,
			expectError: false,
		},
		{
			name:          "missing-method",
			request:       `{"url": "Patient"}`,
			expectError:   true,
			errorContains: "method is required",
		},
		{
			name:          "missing-url",
			request:       `{"method": "POST"}`,
			expectError:   true,
			errorContains: "url is required",
		},
		{
			name:          "invalid-method",
			request:       `{"method": "INVALID", "url": "Patient"}`,
			expectError:   true,
			errorContains: "Invalid request method",
		},
		{
			name:        "valid-get",
			request:     `{"method": "GET", "url": "Patient/123"}`,
			expectError: false,
		},
		{
			name:        "valid-put",
			request:     `{"method": "PUT", "url": "Patient/123"}`,
			expectError: false,
		},
		{
			name:        "valid-delete",
			request:     `{"method": "DELETE", "url": "Patient/123"}`,
			expectError: false,
		},
		{
			name:        "valid-patch",
			request:     `{"method": "PATCH", "url": "Patient/123"}`,
			expectError: false,
		},
		{
			name:        "valid-head",
			request:     `{"method": "HEAD", "url": "Patient/123"}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-request",
				"type": "transaction",
				"entry": [{
					"request": ` + tt.request + `,
					"resource": {"resourceType": "Patient", "id": "pat1"}
				}]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasExpectedError := false
			for _, issue := range result.Issues {
				if tt.errorContains != "" && strings.Contains(issue.Diagnostics, tt.errorContains) {
					hasExpectedError = true
					break
				}
				if tt.errorContains == "" && (issue.Code == IssueCodeRequired || issue.Code == IssueCodeCodeInvalid) {
					hasExpectedError = true
				}
			}

			if tt.expectError && !hasExpectedError {
				t.Errorf("Expected error containing '%s' for %s", tt.errorContains, tt.name)
			}
		})
	}
}

func TestValidateEntryResponseContent(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		response      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid-response",
			response:    `{"status": "200 OK"}`,
			expectError: false,
		},
		{
			name:        "valid-created",
			response:    `{"status": "201 Created"}`,
			expectError: false,
		},
		{
			name:          "missing-status",
			response:      `{"location": "Patient/123/_history/1"}`,
			expectError:   true,
			errorContains: "status is required",
		},
		{
			name:          "empty-status",
			response:      `{"status": ""}`,
			expectError:   true,
			errorContains: "status is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-response",
				"type": "transaction-response",
				"entry": [{
					"response": ` + tt.response + `
				}]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasExpectedError := false
			for _, issue := range result.Issues {
				if tt.errorContains != "" && strings.Contains(issue.Diagnostics, tt.errorContains) {
					hasExpectedError = true
					break
				}
			}

			if tt.expectError && !hasExpectedError {
				t.Errorf("Expected error containing '%s' for %s", tt.errorContains, tt.name)
			}
		})
	}
}

// ============================================================================
// Entry search validation
// ============================================================================

func TestValidateEntrySearchContent(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name          string
		search        string
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid-search-match",
			search:      `{"mode": "match"}`,
			expectError: false,
		},
		{
			name:        "valid-search-include",
			search:      `{"mode": "include"}`,
			expectError: false,
		},
		{
			name:        "valid-search-outcome",
			search:      `{"mode": "outcome"}`,
			expectError: false,
		},
		{
			name:          "invalid-search-mode",
			search:        `{"mode": "invalid"}`,
			expectError:   true,
			errorContains: "Invalid search mode",
		},
		{
			name:        "valid-search-score",
			search:      `{"mode": "match", "score": 0.85}`,
			expectError: false,
		},
		{
			name:        "valid-search-score-zero",
			search:      `{"mode": "match", "score": 0}`,
			expectError: false,
		},
		{
			name:        "valid-search-score-one",
			search:      `{"mode": "match", "score": 1}`,
			expectError: false,
		},
		{
			name:          "invalid-search-score-negative",
			search:        `{"mode": "match", "score": -0.5}`,
			expectError:   true,
			errorContains: "score must be between 0 and 1",
		},
		{
			name:          "invalid-search-score-above-one",
			search:        `{"mode": "match", "score": 1.5}`,
			expectError:   true,
			errorContains: "score must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-search",
				"type": "searchset",
				"total": 1,
				"entry": [{
					"search": ` + tt.search + `,
					"resource": {"resourceType": "Patient", "id": "pat1"}
				}]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasExpectedError := false
			for _, issue := range result.Issues {
				if tt.errorContains != "" && strings.Contains(issue.Diagnostics, tt.errorContains) {
					hasExpectedError = true
					break
				}
			}

			if tt.expectError && !hasExpectedError {
				t.Errorf("Expected error containing '%s' for %s", tt.errorContains, tt.name)
			}
			if !tt.expectError && hasExpectedError {
				t.Errorf("Unexpected error for %s", tt.name)
			}
		})
	}
}

// ============================================================================
// Nested Bundle validation (recursive)
// ============================================================================

func TestValidateNestedBundle(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Collection containing a nested transaction bundle
	bundle := []byte(`{
		"resourceType": "Bundle",
		"id": "outer-bundle",
		"type": "collection",
		"entry": [{
			"fullUrl": "urn:uuid:inner-bundle",
			"resource": {
				"resourceType": "Bundle",
				"id": "inner-bundle",
				"type": "transaction",
				"entry": [{
					"request": {"method": "POST", "url": "Patient"},
					"resource": {"resourceType": "Patient", "id": "pat1"}
				}]
			}
		}]
	}`)

	result, err := v.Validate(ctx, bundle)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	t.Logf("Nested bundle validation: valid=%v, errors=%d, warnings=%d", result.Valid, result.ErrorCount(), result.WarningCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}
}

func TestValidateNestedBundleWithErrors(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	// Collection containing a nested transaction bundle missing request
	bundle := []byte(`{
		"resourceType": "Bundle",
		"id": "outer-bundle",
		"type": "collection",
		"entry": [{
			"fullUrl": "urn:uuid:inner-bundle",
			"resource": {
				"resourceType": "Bundle",
				"id": "inner-bundle",
				"type": "transaction",
				"entry": [{
					"resource": {"resourceType": "Patient", "id": "pat1"}
				}]
			}
		}]
	}`)

	result, err := v.Validate(ctx, bundle)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	// Should have bdl-3 violation for nested bundle
	hasBdl3Error := false
	for _, issue := range result.Issues {
		if issue.Code == IssueCodeInvariant && strings.Contains(issue.Diagnostics, "bdl-3") {
			hasBdl3Error = true
			break
		}
	}

	if !hasBdl3Error {
		t.Error("Expected bdl-3 violation for nested transaction bundle missing request")
	}
}

// ============================================================================
// Entry resource validation
// ============================================================================

func TestValidateEntryResourceType(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	tests := []struct {
		name        string
		resource    string
		expectError bool
	}{
		{
			name:        "valid-patient",
			resource:    `{"resourceType": "Patient", "id": "pat1"}`,
			expectError: false,
		},
		{
			name:        "valid-observation",
			resource:    `{"resourceType": "Observation", "id": "obs1", "status": "final", "code": {"coding": [{"system": "http://loinc.org", "code": "1234-5"}]}}`,
			expectError: false,
		},
		{
			name:        "missing-resourceType",
			resource:    `{"id": "missing-type"}`,
			expectError: true,
		},
		{
			name:        "unknown-resourceType",
			resource:    `{"resourceType": "UnknownResource", "id": "unknown"}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := []byte(`{
				"resourceType": "Bundle",
				"id": "test-entry-resource",
				"type": "collection",
				"entry": [{
					"fullUrl": "urn:uuid:test-entry",
					"resource": ` + tt.resource + `
				}]
			}`)

			result, err := v.Validate(ctx, bundle)
			if err != nil {
				t.Fatalf("Validate returned error: %v", err)
			}

			hasError := result.HasErrors()

			if tt.expectError && !hasError {
				t.Errorf("Expected error for %s", tt.name)
			}
		})
	}
}

// ============================================================================
// Valid complete Bundle examples
// ============================================================================

func TestValidateCompleteDocumentBundle(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	document := []byte(`{
		"resourceType": "Bundle",
		"id": "complete-document",
		"type": "document",
		"identifier": {
			"system": "urn:ietf:rfc:3986",
			"value": "urn:uuid:0c3201bd-1c00-4b61-9b45-12345678"
		},
		"timestamp": "2024-01-15T10:00:00Z",
		"entry": [
			{
				"fullUrl": "urn:uuid:composition-1",
				"resource": {
					"resourceType": "Composition",
					"id": "comp1",
					"status": "final",
					"type": {
						"coding": [{
							"system": "http://loinc.org",
							"code": "11503-0",
							"display": "Medical records"
						}]
					},
					"subject": {"reference": "urn:uuid:patient-1"},
					"date": "2024-01-15",
					"author": [{"reference": "urn:uuid:practitioner-1"}],
					"title": "Patient Summary Document",
					"section": [{
						"title": "Active Problems",
						"entry": [{"reference": "urn:uuid:condition-1"}]
					}]
				}
			},
			{
				"fullUrl": "urn:uuid:patient-1",
				"resource": {
					"resourceType": "Patient",
					"id": "pat1",
					"name": [{"family": "Doe", "given": ["John"]}],
					"gender": "male",
					"birthDate": "1970-01-01"
				}
			},
			{
				"fullUrl": "urn:uuid:practitioner-1",
				"resource": {
					"resourceType": "Practitioner",
					"id": "prac1",
					"name": [{"family": "Smith", "given": ["Jane"]}]
				}
			},
			{
				"fullUrl": "urn:uuid:condition-1",
				"resource": {
					"resourceType": "Condition",
					"id": "cond1",
					"clinicalStatus": {
						"coding": [{
							"system": "http://terminology.hl7.org/CodeSystem/condition-clinical",
							"code": "active"
						}]
					},
					"code": {
						"coding": [{
							"system": "http://snomed.info/sct",
							"code": "38341003",
							"display": "Hypertension"
						}]
					},
					"subject": {"reference": "urn:uuid:patient-1"}
				}
			}
		]
	}`)

	result, err := v.Validate(ctx, document)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	t.Logf("Document bundle validation: valid=%v, errors=%d, warnings=%d", result.Valid, result.ErrorCount(), result.WarningCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	// Should be valid
	if result.HasErrors() {
		t.Error("Complete valid document bundle should not have errors")
	}
}

func TestValidateCompleteTransactionBundle(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	transaction := []byte(`{
		"resourceType": "Bundle",
		"id": "complete-transaction",
		"type": "transaction",
		"entry": [
			{
				"fullUrl": "urn:uuid:new-patient",
				"resource": {
					"resourceType": "Patient",
					"name": [{"family": "Johnson", "given": ["Alice"]}],
					"gender": "female",
					"birthDate": "1985-06-15"
				},
				"request": {
					"method": "POST",
					"url": "Patient"
				}
			},
			{
				"fullUrl": "urn:uuid:new-observation",
				"resource": {
					"resourceType": "Observation",
					"status": "final",
					"code": {
						"coding": [{
							"system": "http://loinc.org",
							"code": "29463-7",
							"display": "Body Weight"
						}]
					},
					"subject": {"reference": "urn:uuid:new-patient"},
					"valueQuantity": {
						"value": 65.5,
						"unit": "kg",
						"system": "http://unitsofmeasure.org",
						"code": "kg"
					}
				},
				"request": {
					"method": "POST",
					"url": "Observation"
				}
			},
			{
				"request": {
					"method": "DELETE",
					"url": "Observation/old-obs-123"
				}
			}
		]
	}`)

	result, err := v.Validate(ctx, transaction)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	t.Logf("Transaction bundle validation: valid=%v, errors=%d, warnings=%d", result.Valid, result.ErrorCount(), result.WarningCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}
}

func TestValidateCompleteSearchsetBundle(t *testing.T) {
	v := setupTestValidator(t)
	ctx := context.Background()

	searchset := []byte(`{
		"resourceType": "Bundle",
		"id": "complete-searchset",
		"type": "searchset",
		"total": 2,
		"link": [
			{"relation": "self", "url": "http://example.org/Patient?name=doe"},
			{"relation": "next", "url": "http://example.org/Patient?name=doe&_page=2"}
		],
		"entry": [
			{
				"fullUrl": "http://example.org/Patient/pat1",
				"resource": {
					"resourceType": "Patient",
					"id": "pat1",
					"name": [{"family": "Doe", "given": ["John"]}]
				},
				"search": {
					"mode": "match",
					"score": 0.95
				}
			},
			{
				"fullUrl": "http://example.org/Patient/pat2",
				"resource": {
					"resourceType": "Patient",
					"id": "pat2",
					"name": [{"family": "Doe", "given": ["Jane"]}]
				},
				"search": {
					"mode": "match",
					"score": 0.90
				}
			}
		]
	}`)

	result, err := v.Validate(ctx, searchset)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	t.Logf("Searchset bundle validation: valid=%v, errors=%d, warnings=%d", result.Valid, result.ErrorCount(), result.WarningCount())
	for _, issue := range result.Issues {
		t.Logf("Issue: [%s] %s - %s (path: %v)", issue.Severity, issue.Code, issue.Diagnostics, issue.Expression)
	}

	// Should be valid
	if result.HasErrors() {
		t.Error("Complete valid searchset bundle should not have errors")
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkValidateCollectionBundle(b *testing.B) {
	v := setupBenchmarkValidator(b)
	ctx := context.Background()

	bundle := []byte(`{
		"resourceType": "Bundle",
		"id": "bench-collection",
		"type": "collection",
		"entry": [
			{"fullUrl": "urn:uuid:1", "resource": {"resourceType": "Patient", "id": "p1", "name": [{"family": "Doe"}]}},
			{"fullUrl": "urn:uuid:2", "resource": {"resourceType": "Patient", "id": "p2", "name": [{"family": "Smith"}]}},
			{"fullUrl": "urn:uuid:3", "resource": {"resourceType": "Observation", "id": "o1", "status": "final", "code": {"coding": [{"system": "http://loinc.org", "code": "1234-5"}]}}}
		]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate(ctx, bundle)
	}
}

func BenchmarkValidateTransactionBundle(b *testing.B) {
	v := setupBenchmarkValidator(b)
	ctx := context.Background()

	bundle := []byte(`{
		"resourceType": "Bundle",
		"id": "bench-transaction",
		"type": "transaction",
		"entry": [
			{
				"fullUrl": "urn:uuid:1",
				"resource": {"resourceType": "Patient", "id": "p1", "name": [{"family": "Doe"}]},
				"request": {"method": "POST", "url": "Patient"}
			},
			{
				"fullUrl": "urn:uuid:2",
				"resource": {"resourceType": "Observation", "id": "o1", "status": "final", "code": {"coding": [{"system": "http://loinc.org", "code": "1234-5"}]}},
				"request": {"method": "POST", "url": "Observation"}
			}
		]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate(ctx, bundle)
	}
}

func BenchmarkValidateDocumentBundle(b *testing.B) {
	v := setupBenchmarkValidator(b)
	ctx := context.Background()

	bundle := []byte(`{
		"resourceType": "Bundle",
		"id": "bench-document",
		"type": "document",
		"identifier": {"system": "urn:ietf:rfc:3986", "value": "urn:uuid:12345"},
		"timestamp": "2024-01-15T10:00:00Z",
		"entry": [{
			"fullUrl": "urn:uuid:composition-1",
			"resource": {
				"resourceType": "Composition",
				"id": "comp1",
				"status": "final",
				"type": {"coding": [{"system": "http://loinc.org", "code": "11503-0"}]},
				"subject": {"reference": "Patient/pat1"},
				"date": "2024-01-15",
				"author": [{"reference": "Practitioner/prac1"}],
				"title": "Test Document"
			}
		}]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Validate(ctx, bundle)
	}
}

// Helper function for benchmarks
func setupBenchmarkValidator(b *testing.B) *Validator {
	b.Helper()
	reg := NewRegistry(FHIRVersionR4)
	resourcesPath := "../../specs/r4/profiles-resources.json"
	if _, err := reg.LoadFromFile(resourcesPath); err != nil {
		b.Skip("Specs not found, skipping benchmark")
	}
	typesPath := "../../specs/r4/profiles-types.json"
	reg.LoadFromFile(typesPath)
	return NewValidator(reg, DefaultValidatorOptions())
}
