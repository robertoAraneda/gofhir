// bundle_validation.go demonstrates FHIR Bundle validation.
// This includes validation of:
// - Bundle types (document, message, transaction, searchset, etc.)
// - Bundle-specific constraints (bdl-1 through bdl-12)
// - Entry validation (request, response, search)
// - Nested resource validation within Bundle entries
package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/robertoaraneda/gofhir/pkg/validator"
)

// RunBundleValidationExamples demonstrates Bundle validation features
func RunBundleValidationExamples(ctx context.Context) {
	fmt.Println("\n" + separator)
	fmt.Println("BUNDLE VALIDATION EXAMPLES")
	fmt.Println(separator)

	v, err := createBundleValidator()
	if err != nil {
		log.Printf("Failed to create validator: %v", err)
		return
	}

	// Example 1: Valid Document Bundle
	fmt.Println("\n--- 1. Valid Document Bundle ---")
	validateDocumentBundle(ctx, v)

	// Example 2: Invalid Document Bundle (missing timestamp - bdl-10)
	fmt.Println("\n--- 2. Invalid Document Bundle (missing timestamp) ---")
	validateDocumentBundleMissingTimestamp(ctx, v)

	// Example 3: Invalid Document Bundle (missing identifier - bdl-9)
	fmt.Println("\n--- 3. Invalid Document Bundle (missing identifier) ---")
	validateDocumentBundleMissingIdentifier(ctx, v)

	// Example 4: Invalid Document Bundle (first entry not Composition - bdl-11)
	fmt.Println("\n--- 4. Invalid Document Bundle (first entry not Composition) ---")
	validateDocumentBundleWrongFirstEntry(ctx, v)

	// Example 5: Valid Transaction Bundle
	fmt.Println("\n--- 5. Valid Transaction Bundle ---")
	validateTransactionBundle(ctx, v)

	// Example 6: Invalid Transaction Bundle (missing request - bdl-3)
	fmt.Println("\n--- 6. Invalid Transaction Bundle (missing request) ---")
	validateTransactionBundleMissingRequest(ctx, v)

	// Example 7: Valid Searchset Bundle
	fmt.Println("\n--- 7. Valid Searchset Bundle ---")
	validateSearchsetBundle(ctx, v)

	// Example 8: Invalid Bundle (total in wrong type - bdl-1)
	fmt.Println("\n--- 8. Invalid Bundle (total in wrong type) ---")
	validateBundleWithWrongTotal(ctx, v)

	// Example 9: Valid Message Bundle
	fmt.Println("\n--- 9. Valid Message Bundle ---")
	validateMessageBundle(ctx, v)

	// Example 10: Invalid Message Bundle (first entry not MessageHeader - bdl-12)
	fmt.Println("\n--- 10. Invalid Message Bundle (first entry not MessageHeader) ---")
	validateMessageBundleWrongFirstEntry(ctx, v)

	// Example 11: Duplicate fullUrl (bdl-7)
	fmt.Println("\n--- 11. Invalid Bundle (duplicate fullUrl) ---")
	validateBundleDuplicateFullUrl(ctx, v)

	// Example 12: Version-specific fullUrl (bdl-8)
	fmt.Println("\n--- 12. Invalid Bundle (version-specific fullUrl) ---")
	validateBundleVersionSpecificFullUrl(ctx, v)

	// Example 13: Valid Batch Bundle
	fmt.Println("\n--- 13. Valid Batch Bundle ---")
	validateBatchBundle(ctx, v)

	// Example 14: Valid Collection Bundle
	fmt.Println("\n--- 14. Valid Collection Bundle ---")
	validateCollectionBundle(ctx, v)
}

func createBundleValidator() (*validator.Validator, error) {
	registry := validator.NewRegistry(validator.FHIRVersionR4)
	specsDir := filepath.Join("..", "..", "specs", "r4")
	_, err := registry.LoadR4Specs(specsDir)
	if err != nil {
		return nil, err
	}

	opts := validator.ValidatorOptions{
		ValidateConstraints: true,
		ValidateExtensions:  true,
		ValidateReferences:  false,
		ValidateTerminology: false,
		StrictMode:          false,
	}
	return validator.NewValidator(registry, opts), nil
}

func validateDocumentBundle(ctx context.Context, v *validator.Validator) {
	document := []byte(`{
		"resourceType": "Bundle",
		"id": "document-example",
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
					"title": "Patient Summary"
				}
			},
			{
				"fullUrl": "urn:uuid:patient-1",
				"resource": {
					"resourceType": "Patient",
					"id": "pat1",
					"name": [{"family": "Doe", "given": ["John"]}],
					"gender": "male"
				}
			}
		]
	}`)

	result, err := v.Validate(ctx, document)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Valid Document Bundle", result)
	fmt.Println("  -> Document has identifier (bdl-9), timestamp (bdl-10), and Composition first (bdl-11)")
}

func validateDocumentBundleMissingTimestamp(ctx context.Context, v *validator.Validator) {
	document := []byte(`{
		"resourceType": "Bundle",
		"id": "document-no-timestamp",
		"type": "document",
		"identifier": {
			"system": "urn:ietf:rfc:3986",
			"value": "urn:uuid:12345"
		},
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
				"title": "Test"
			}
		}]
	}`)

	result, err := v.Validate(ctx, document)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Document Bundle Missing Timestamp", result)
	printBundleConstraintViolations(result, "bdl-10")
	fmt.Println("\n  -> Constraint bdl-10: Document bundles must have a timestamp")
}

func validateDocumentBundleMissingIdentifier(ctx context.Context, v *validator.Validator) {
	document := []byte(`{
		"resourceType": "Bundle",
		"id": "document-no-identifier",
		"type": "document",
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
				"title": "Test"
			}
		}]
	}`)

	result, err := v.Validate(ctx, document)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Document Bundle Missing Identifier", result)
	printBundleConstraintViolations(result, "bdl-9")
	fmt.Println("\n  -> Constraint bdl-9: Document bundles must have identifier with system and value")
}

func validateDocumentBundleWrongFirstEntry(ctx context.Context, v *validator.Validator) {
	document := []byte(`{
		"resourceType": "Bundle",
		"id": "document-wrong-first",
		"type": "document",
		"identifier": {"system": "urn:ietf:rfc:3986", "value": "urn:uuid:12345"},
		"timestamp": "2024-01-15T10:00:00Z",
		"entry": [{
			"fullUrl": "urn:uuid:patient-1",
			"resource": {
				"resourceType": "Patient",
				"id": "pat1",
				"name": [{"family": "Doe"}]
			}
		}]
	}`)

	result, err := v.Validate(ctx, document)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Document Bundle Wrong First Entry", result)
	printBundleConstraintViolations(result, "bdl-11")
	fmt.Println("\n  -> Constraint bdl-11: Document first entry must be a Composition")
}

func validateTransactionBundle(ctx context.Context, v *validator.Validator) {
	transaction := []byte(`{
		"resourceType": "Bundle",
		"id": "transaction-example",
		"type": "transaction",
		"entry": [
			{
				"fullUrl": "urn:uuid:new-patient",
				"resource": {
					"resourceType": "Patient",
					"name": [{"family": "Johnson", "given": ["Alice"]}],
					"gender": "female"
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
					"code": {"coding": [{"system": "http://loinc.org", "code": "29463-7"}]},
					"subject": {"reference": "urn:uuid:new-patient"},
					"valueQuantity": {"value": 65.5, "unit": "kg"}
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
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Valid Transaction Bundle", result)
	fmt.Println("  -> Transaction entries all have request with method and url")
	fmt.Println("  -> DELETE entry has no resource (allowed)")
}

func validateTransactionBundleMissingRequest(ctx context.Context, v *validator.Validator) {
	transaction := []byte(`{
		"resourceType": "Bundle",
		"id": "transaction-no-request",
		"type": "transaction",
		"entry": [{
			"fullUrl": "urn:uuid:new-patient",
			"resource": {
				"resourceType": "Patient",
				"name": [{"family": "Johnson"}]
			}
		}]
	}`)

	result, err := v.Validate(ctx, transaction)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Transaction Bundle Missing Request", result)
	printBundleConstraintViolations(result, "bdl-3")
	fmt.Println("\n  -> Constraint bdl-3: Transaction entries must have entry.request")
}

func validateSearchsetBundle(ctx context.Context, v *validator.Validator) {
	searchset := []byte(`{
		"resourceType": "Bundle",
		"id": "searchset-example",
		"type": "searchset",
		"total": 2,
		"link": [
			{"relation": "self", "url": "http://example.org/Patient?name=doe"}
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
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Valid Searchset Bundle", result)
	fmt.Println("  -> Searchset allows total and entry.search (bdl-1, bdl-2)")
	fmt.Println("  -> search.mode and search.score validated")
}

func validateBundleWithWrongTotal(ctx context.Context, v *validator.Validator) {
	document := []byte(`{
		"resourceType": "Bundle",
		"id": "document-with-total",
		"type": "document",
		"total": 5,
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
				"title": "Test"
			}
		}]
	}`)

	result, err := v.Validate(ctx, document)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Document Bundle with Total", result)
	printBundleConstraintViolations(result, "bdl-1")
	fmt.Println("\n  -> Constraint bdl-1: total only allowed for searchset and history")
}

func validateMessageBundle(ctx context.Context, v *validator.Validator) {
	message := []byte(`{
		"resourceType": "Bundle",
		"id": "message-example",
		"type": "message",
		"entry": [
			{
				"fullUrl": "urn:uuid:messageheader-1",
				"resource": {
					"resourceType": "MessageHeader",
					"id": "mh1",
					"eventCoding": {
						"system": "http://example.org/events",
						"code": "admin-notify"
					},
					"source": {
						"endpoint": "http://example.org/source"
					}
				}
			},
			{
				"fullUrl": "urn:uuid:patient-1",
				"resource": {
					"resourceType": "Patient",
					"id": "pat1",
					"name": [{"family": "Doe"}]
				}
			}
		]
	}`)

	result, err := v.Validate(ctx, message)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Valid Message Bundle", result)
	fmt.Println("  -> Message first entry is MessageHeader (bdl-12)")
}

func validateMessageBundleWrongFirstEntry(ctx context.Context, v *validator.Validator) {
	message := []byte(`{
		"resourceType": "Bundle",
		"id": "message-wrong-first",
		"type": "message",
		"entry": [{
			"fullUrl": "urn:uuid:patient-1",
			"resource": {
				"resourceType": "Patient",
				"id": "pat1",
				"name": [{"family": "Doe"}]
			}
		}]
	}`)

	result, err := v.Validate(ctx, message)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Message Bundle Wrong First Entry", result)
	printBundleConstraintViolations(result, "bdl-12")
	fmt.Println("\n  -> Constraint bdl-12: Message first entry must be MessageHeader")
}

func validateBundleDuplicateFullUrl(ctx context.Context, v *validator.Validator) {
	collection := []byte(`{
		"resourceType": "Bundle",
		"id": "collection-duplicate-fullurl",
		"type": "collection",
		"entry": [
			{
				"fullUrl": "urn:uuid:same-url",
				"resource": {"resourceType": "Patient", "id": "p1", "name": [{"family": "Doe"}]}
			},
			{
				"fullUrl": "urn:uuid:same-url",
				"resource": {"resourceType": "Patient", "id": "p2", "name": [{"family": "Smith"}]}
			}
		]
	}`)

	result, err := v.Validate(ctx, collection)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Bundle with Duplicate fullUrl", result)
	printBundleConstraintViolations(result, "bdl-7")
	fmt.Println("\n  -> Constraint bdl-7: fullUrl must be unique within Bundle")
}

func validateBundleVersionSpecificFullUrl(ctx context.Context, v *validator.Validator) {
	collection := []byte(`{
		"resourceType": "Bundle",
		"id": "collection-version-fullurl",
		"type": "collection",
		"entry": [{
			"fullUrl": "http://example.org/Patient/123/_history/1",
			"resource": {"resourceType": "Patient", "id": "123", "name": [{"family": "Doe"}]}
		}]
	}`)

	result, err := v.Validate(ctx, collection)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Bundle with Version-Specific fullUrl", result)
	printBundleConstraintViolations(result, "bdl-8")
	fmt.Println("\n  -> Constraint bdl-8: fullUrl cannot be a version-specific reference")
}

func validateBatchBundle(ctx context.Context, v *validator.Validator) {
	batch := []byte(`{
		"resourceType": "Bundle",
		"id": "batch-example",
		"type": "batch",
		"entry": [
			{
				"request": {
					"method": "GET",
					"url": "Patient/123"
				}
			},
			{
				"request": {
					"method": "GET",
					"url": "Observation?patient=123"
				}
			},
			{
				"fullUrl": "urn:uuid:new-patient",
				"resource": {
					"resourceType": "Patient",
					"name": [{"family": "New"}]
				},
				"request": {
					"method": "POST",
					"url": "Patient"
				}
			}
		]
	}`)

	result, err := v.Validate(ctx, batch)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Valid Batch Bundle", result)
	fmt.Println("  -> Batch entries all have request (bdl-3)")
	fmt.Println("  -> Mix of GET (no resource) and POST (with resource)")
}

func validateCollectionBundle(ctx context.Context, v *validator.Validator) {
	collection := []byte(`{
		"resourceType": "Bundle",
		"id": "collection-example",
		"type": "collection",
		"entry": [
			{
				"fullUrl": "http://example.org/Patient/pat1",
				"resource": {
					"resourceType": "Patient",
					"id": "pat1",
					"name": [{"family": "Doe", "given": ["John"]}],
					"gender": "male"
				}
			},
			{
				"fullUrl": "http://example.org/Observation/obs1",
				"resource": {
					"resourceType": "Observation",
					"id": "obs1",
					"status": "final",
					"code": {"coding": [{"system": "http://loinc.org", "code": "29463-7"}]},
					"subject": {"reference": "Patient/pat1"},
					"valueQuantity": {"value": 70.5, "unit": "kg"}
				}
			},
			{
				"fullUrl": "http://example.org/Condition/cond1",
				"resource": {
					"resourceType": "Condition",
					"id": "cond1",
					"clinicalStatus": {
						"coding": [{"system": "http://terminology.hl7.org/CodeSystem/condition-clinical", "code": "active"}]
					},
					"code": {"coding": [{"system": "http://snomed.info/sct", "code": "38341003"}]},
					"subject": {"reference": "Patient/pat1"}
				}
			}
		]
	}`)

	result, err := v.Validate(ctx, collection)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	printResult("Valid Collection Bundle", result)
	fmt.Println("  -> Collection is the simplest bundle type")
	fmt.Println("  -> No request/response required, no special constraints")
	fmt.Println("  -> Each entry.resource is validated individually")
}

// printBundleConstraintViolations prints specific Bundle constraint violations
func printBundleConstraintViolations(result *validator.ValidationResult, constraintKey string) {
	for _, issue := range result.Issues {
		if issue.Code == validator.IssueCodeInvariant {
			fmt.Printf("  Error: %s\n", issue.Diagnostics)
			if len(issue.Expression) > 0 {
				fmt.Printf("         at %s\n", issue.Expression[0])
			}
		}
	}
}
