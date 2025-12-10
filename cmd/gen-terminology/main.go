// Command gen-terminology generates embedded ValueSet code from FHIR specifications.
//
// Usage:
//
//	go run ./cmd/gen-terminology -version r4
//	go run ./cmd/gen-terminology -version r4b
//	go run ./cmd/gen-terminology -version r5
//	go run ./cmd/gen-terminology -version all
//
// This generates pkg/validator/terminology_embedded_{version}.go files
// containing pre-loaded ValueSets for efficient terminology validation.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/robertoaraneda/gofhir/internal/codegen/generator"
)

// Required ValueSet URLs for terminology validation.
// These are the ValueSets with "required" binding strength that must be validated.
// Generated from analysis of FHIR R4 StructureDefinitions.
var requiredValueSetURLs = []string{
	// Most commonly used (high priority)
	"http://hl7.org/fhir/ValueSet/publication-status",
	"http://hl7.org/fhir/ValueSet/administrative-gender",
	"http://hl7.org/fhir/ValueSet/observation-status",
	"http://hl7.org/fhir/ValueSet/request-status",
	"http://hl7.org/fhir/ValueSet/request-priority",
	"http://hl7.org/fhir/ValueSet/event-status",
	"http://hl7.org/fhir/ValueSet/diagnostic-report-status",
	"http://hl7.org/fhir/ValueSet/medication-request-status",
	"http://hl7.org/fhir/ValueSet/medicationrequest-intent",
	"http://hl7.org/fhir/ValueSet/fm-status",

	// Resource types and references
	"http://hl7.org/fhir/ValueSet/resource-types",
	"http://hl7.org/fhir/ValueSet/all-types",
	"http://hl7.org/fhir/ValueSet/defined-types",

	// Patient related
	"http://hl7.org/fhir/ValueSet/link-type",
	"http://hl7.org/fhir/ValueSet/identifier-use",
	"http://hl7.org/fhir/ValueSet/name-use",
	"http://hl7.org/fhir/ValueSet/contact-point-system",
	"http://hl7.org/fhir/ValueSet/contact-point-use",
	"http://hl7.org/fhir/ValueSet/address-use",
	"http://hl7.org/fhir/ValueSet/address-type",

	// Encounter and Episode
	"http://hl7.org/fhir/ValueSet/encounter-status",
	"http://hl7.org/fhir/ValueSet/encounter-location-status",
	"http://hl7.org/fhir/ValueSet/episode-of-care-status",

	// Clinical
	"http://hl7.org/fhir/ValueSet/condition-clinical",
	"http://hl7.org/fhir/ValueSet/condition-ver-status",
	"http://hl7.org/fhir/ValueSet/allergy-intolerance-type",
	"http://hl7.org/fhir/ValueSet/allergy-intolerance-category",
	"http://hl7.org/fhir/ValueSet/allergy-intolerance-criticality",
	"http://hl7.org/fhir/ValueSet/allergyintolerance-clinical",
	"http://hl7.org/fhir/ValueSet/allergyintolerance-verification",

	// Medication
	"http://hl7.org/fhir/ValueSet/medication-status",
	"http://hl7.org/fhir/ValueSet/medication-admin-status",
	"http://hl7.org/fhir/ValueSet/medication-statement-status",
	"http://hl7.org/fhir/ValueSet/medicationdispense-status",

	// Procedure and ServiceRequest
	"http://hl7.org/fhir/ValueSet/procedure-status",
	"http://hl7.org/fhir/ValueSet/request-intent",

	// Documents and Composition
	"http://hl7.org/fhir/ValueSet/composition-status",
	"http://hl7.org/fhir/ValueSet/document-reference-status",
	"http://hl7.org/fhir/ValueSet/list-status",
	"http://hl7.org/fhir/ValueSet/list-mode",

	// Care Plan and Goals
	"http://hl7.org/fhir/ValueSet/care-plan-status",
	"http://hl7.org/fhir/ValueSet/care-plan-intent",
	"http://hl7.org/fhir/ValueSet/care-plan-activity-status",
	"http://hl7.org/fhir/ValueSet/goal-status",

	// Immunization
	"http://hl7.org/fhir/ValueSet/immunization-status",

	// Appointment and Schedule
	"http://hl7.org/fhir/ValueSet/appointmentstatus",
	"http://hl7.org/fhir/ValueSet/participationstatus",
	"http://hl7.org/fhir/ValueSet/slotstatus",

	// Communication
	"http://hl7.org/fhir/ValueSet/communication-status",

	// Task
	"http://hl7.org/fhir/ValueSet/task-status",
	"http://hl7.org/fhir/ValueSet/task-intent",

	// Device
	"http://hl7.org/fhir/ValueSet/device-status",

	// Coverage and Claims
	"http://hl7.org/fhir/ValueSet/coverage-status",
	"http://hl7.org/fhir/ValueSet/claim-status",
	"http://hl7.org/fhir/ValueSet/explanationofbenefit-status",

	// Consent
	"http://hl7.org/fhir/ValueSet/consent-state-codes",

	// Narrative
	"http://hl7.org/fhir/ValueSet/narrative-status",

	// Quantity
	"http://hl7.org/fhir/ValueSet/quantity-comparator",

	// Bundle
	"http://hl7.org/fhir/ValueSet/bundle-type",
	"http://hl7.org/fhir/ValueSet/http-verb",
	"http://hl7.org/fhir/ValueSet/search-entry-mode",

	// StructureDefinition
	"http://hl7.org/fhir/ValueSet/type-derivation-rule",
	"http://hl7.org/fhir/ValueSet/structure-definition-kind",
	"http://hl7.org/fhir/ValueSet/extension-context-type",

	// CapabilityStatement
	"http://hl7.org/fhir/ValueSet/restful-capability-mode",
	"http://hl7.org/fhir/ValueSet/type-restful-interaction",
	"http://hl7.org/fhir/ValueSet/system-restful-interaction",
	"http://hl7.org/fhir/ValueSet/resource-version-policy",
	"http://hl7.org/fhir/ValueSet/conditional-read-status",
	"http://hl7.org/fhir/ValueSet/conditional-delete-status",
	"http://hl7.org/fhir/ValueSet/reference-handling-policy",
	"http://hl7.org/fhir/ValueSet/search-param-type",

	// CodeSystem and ValueSet
	"http://hl7.org/fhir/ValueSet/codesystem-content-mode",
	"http://hl7.org/fhir/ValueSet/filter-operator",

	// OperationOutcome
	"http://hl7.org/fhir/ValueSet/issue-severity",
	"http://hl7.org/fhir/ValueSet/issue-type",

	// Binding strength
	"http://hl7.org/fhir/ValueSet/binding-strength",

	// Days of week
	"http://hl7.org/fhir/ValueSet/days-of-week",

	// Units of time
	"http://hl7.org/fhir/ValueSet/units-of-time",

	// Event timing
	"http://hl7.org/fhir/ValueSet/event-timing",

	// Subscription
	"http://hl7.org/fhir/ValueSet/subscription-status",
	"http://hl7.org/fhir/ValueSet/subscription-channel-type",

	// AuditEvent
	"http://hl7.org/fhir/ValueSet/audit-event-action",
	"http://hl7.org/fhir/ValueSet/audit-event-outcome",

	// Flag
	"http://hl7.org/fhir/ValueSet/flag-status",

	// Specimen
	"http://hl7.org/fhir/ValueSet/specimen-status",

	// Location
	"http://hl7.org/fhir/ValueSet/location-status",
	"http://hl7.org/fhir/ValueSet/location-mode",

	// Organization
	"http://hl7.org/fhir/ValueSet/organization-type",

	// Contract
	"http://hl7.org/fhir/ValueSet/contract-status",

	// ChargeItem
	"http://hl7.org/fhir/ValueSet/chargeitem-status",

	// Invoice
	"http://hl7.org/fhir/ValueSet/invoice-status",

	// Research
	"http://hl7.org/fhir/ValueSet/research-study-status",
	"http://hl7.org/fhir/ValueSet/research-subject-status",

	// Supply
	"http://hl7.org/fhir/ValueSet/supplyrequest-status",
	"http://hl7.org/fhir/ValueSet/supplydelivery-status",

	// Vision
	"http://hl7.org/fhir/ValueSet/vision-eye-codes",
	"http://hl7.org/fhir/ValueSet/vision-base-codes",

	// Questionnaire
	"http://hl7.org/fhir/ValueSet/questionnaire-answers-status",
	"http://hl7.org/fhir/ValueSet/item-type",
	"http://hl7.org/fhir/ValueSet/questionnaire-enable-operator",
	"http://hl7.org/fhir/ValueSet/questionnaire-enable-behavior",

	// Report
	"http://hl7.org/fhir/ValueSet/report-status-codes",
	"http://hl7.org/fhir/ValueSet/report-result-codes",
	"http://hl7.org/fhir/ValueSet/report-action-result-codes",
	"http://hl7.org/fhir/ValueSet/report-participant-type",

	// Additional status codes
	"http://hl7.org/fhir/ValueSet/operation-kind",
	"http://hl7.org/fhir/ValueSet/message-significance-category",
	"http://hl7.org/fhir/ValueSet/response-code",
	"http://hl7.org/fhir/ValueSet/graph-compartment-use",
	"http://hl7.org/fhir/ValueSet/graph-compartment-rule",
	"http://hl7.org/fhir/ValueSet/compartment-type",
	"http://hl7.org/fhir/ValueSet/assert-direction-codes",
	"http://hl7.org/fhir/ValueSet/assert-operator-codes",
	"http://hl7.org/fhir/ValueSet/assert-response-code-types",
	"http://hl7.org/fhir/ValueSet/action-participant-type",
	"http://hl7.org/fhir/ValueSet/action-grouping-behavior",
	"http://hl7.org/fhir/ValueSet/action-selection-behavior",
	"http://hl7.org/fhir/ValueSet/action-required-behavior",
	"http://hl7.org/fhir/ValueSet/action-precheck-behavior",
	"http://hl7.org/fhir/ValueSet/action-cardinality-behavior",
	"http://hl7.org/fhir/ValueSet/action-relationship-type",
	"http://hl7.org/fhir/ValueSet/action-condition-kind",
	"http://hl7.org/fhir/ValueSet/trigger-type",
	"http://hl7.org/fhir/ValueSet/sort-direction",
	"http://hl7.org/fhir/ValueSet/expression-language",
	"http://hl7.org/fhir/ValueSet/related-artifact-type",
	"http://hl7.org/fhir/ValueSet/contributor-type",
	"http://hl7.org/fhir/ValueSet/guidance-response-status",
	"http://hl7.org/fhir/ValueSet/detectedissue-severity",
}

func main() {
	version := flag.String("version", "r4", "FHIR version: r4, r4b, r5, or all")
	specsDir := flag.String("specs", "specs", "Directory containing FHIR specs")
	outputDir := flag.String("output", "pkg/validator", "Output directory for generated code")
	flag.Parse()

	versions := []string{*version}
	if *version == "all" {
		versions = []string{"r4", "r4b", "r5"}
	}

	for _, v := range versions {
		if err := generateForVersion(*specsDir, *outputDir, v); err != nil {
			log.Printf("Warning: Failed to generate for %s: %v", v, err)
			// Continue with other versions
		}
	}
}

func generateForVersion(specsDir, outputDir, version string) error {
	valueSetsPath := filepath.Join(specsDir, version, "valuesets.json")

	// Check if file exists
	if _, err := os.Stat(valueSetsPath); os.IsNotExist(err) {
		return fmt.Errorf("valuesets.json not found at %s", valueSetsPath)
	}

	fmt.Printf("Loading %s valuesets from %s...\n", strings.ToUpper(version), valueSetsPath)

	gen := generator.NewTerminologyCodegen()
	if err := gen.LoadFromFile(valueSetsPath); err != nil {
		return fmt.Errorf("failed to load valuesets: %w", err)
	}

	cs, vs, codes := gen.Stats()
	fmt.Printf("Loaded %d CodeSystems, %d ValueSets, %d total codes\n", cs, vs, codes)

	// Generate embedded ValueSets
	outputPath := filepath.Join(outputDir, fmt.Sprintf("terminology_embedded_%s.go", version))
	fhirVersion := fhirVersionString(version)

	fmt.Printf("Generating %s with %d required ValueSets...\n", outputPath, len(requiredValueSetURLs))

	if err := gen.WriteToFile(outputPath, "validator", fhirVersion, requiredValueSetURLs); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("Generated %s successfully!\n\n", outputPath)
	return nil
}

func fhirVersionString(version string) string {
	switch version {
	case "r4":
		return "4.0.1"
	case "r4b":
		return "4.3.0"
	case "r5":
		return "5.0.0"
	default:
		return version
	}
}
