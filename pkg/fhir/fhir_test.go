package fhir_test

import (
	"fmt"
	"testing"

	"github.com/robertoaraneda/gofhir/pkg/fhir"
	_ "github.com/robertoaraneda/gofhir/pkg/fhir/r4"
	_ "github.com/robertoaraneda/gofhir/pkg/fhir/r4b"
	_ "github.com/robertoaraneda/gofhir/pkg/fhir/r5"
)

func TestSupportedVersions(t *testing.T) {
	versions := fhir.SupportedVersions()
	if len(versions) != 3 {
		t.Errorf("Expected 3 versions, got %d", len(versions))
	}

	expectedVersions := map[fhir.Version]bool{
		fhir.R4:  true,
		fhir.R4B: true,
		fhir.R5:  true,
	}

	for _, v := range versions {
		if !expectedVersions[v] {
			t.Errorf("Unexpected version: %s", v)
		}
		delete(expectedVersions, v)
	}

	if len(expectedVersions) > 0 {
		t.Errorf("Missing versions: %v", expectedVersions)
	}
}

func TestGetFactory(t *testing.T) {
	testCases := []struct {
		version fhir.Version
		wantErr bool
	}{
		{fhir.R4, false},
		{fhir.R4B, false},
		{fhir.R5, false},
		{fhir.R6, true}, // R6 not implemented yet
	}

	for _, tc := range testCases {
		t.Run(string(tc.version), func(t *testing.T) {
			factory, err := fhir.GetFactory(tc.version)
			if tc.wantErr {
				if err == nil {
					t.Errorf("Expected error for version %s, got nil", tc.version)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for version %s: %v", tc.version, err)
				}
				if factory == nil {
					t.Errorf("Expected factory for version %s, got nil", tc.version)
				}
				if factory.Version() != tc.version {
					t.Errorf("Factory version mismatch: expected %s, got %s", tc.version, factory.Version())
				}
			}
		})
	}
}

func TestFactoryUnmarshalResource(t *testing.T) {
	patientJSON := []byte(`{"resourceType":"Patient","id":"123","name":[{"family":"Doe"}]}`)

	for _, version := range []fhir.Version{fhir.R4, fhir.R4B, fhir.R5} {
		t.Run(string(version), func(t *testing.T) {
			factory, err := fhir.GetFactory(version)
			if err != nil {
				t.Fatalf("Failed to get factory: %v", err)
			}

			resource, err := factory.UnmarshalResource(patientJSON)
			if err != nil {
				t.Fatalf("Failed to unmarshal resource: %v", err)
			}

			if resource.GetResourceType() != "Patient" {
				t.Errorf("Expected resourceType 'Patient', got '%s'", resource.GetResourceType())
			}

			// Test SetID
			resource.SetID("456")
			if *resource.GetID() != "456" {
				t.Errorf("Expected id '456', got '%s'", *resource.GetID())
			}

			// Test Meta
			meta := factory.NewMeta()
			meta.SetVersionID("1")
			meta.SetLastUpdated("2025-01-01T00:00:00Z")
			resource.SetMeta(meta)

			gotMeta := resource.GetMeta()
			if gotMeta == nil {
				t.Fatal("Expected meta, got nil")
			}
			if *gotMeta.GetVersionID() != "1" {
				t.Errorf("Expected versionId '1', got '%s'", *gotMeta.GetVersionID())
			}
		})
	}
}

func ExampleGetFactory() {
	factory, _ := fhir.GetFactory(fhir.R4)
	fmt.Printf("Factory version: %s\n", factory.Version())
	// Output: Factory version: R4
}
