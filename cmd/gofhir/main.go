package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/robertoaraneda/gofhir/internal/codegen/generator"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath"
)

var version = "dev"

func main() {
	if err := execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func execute() error {
	rootCmd := newRootCmd()
	return rootCmd.Execute()
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "gofhir",
		Short: "GoFHIR - FHIR Toolkit for Go",
		Long: `GoFHIR is a production-grade FHIR toolkit for Go.

It provides:
  - Strongly-typed FHIR resources for R4, R4B, and R5
  - Fluent builders for resource construction
  - FHIRPath expression evaluation
  - Resource validation against StructureDefinitions

For more information, visit: https://github.com/robertoaraneda/gofhir`,
	}

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newValidateCmd())
	rootCmd.AddCommand(newFHIRPathCmd())
	rootCmd.AddCommand(newGenerateCmd())

	return rootCmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("gofhir version %s\n", version)
		},
	}
}

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [file]",
		Short: "Validate a FHIR resource",
		Long:  `Validate a FHIR resource against its StructureDefinition.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement validation
			fmt.Printf("Validating: %s\n", args[0])
			fmt.Println("Validation not yet implemented")
			return nil
		},
	}

	cmd.Flags().StringP("version", "v", "R4", "FHIR version (R4, R4B, R5)")
	cmd.Flags().Bool("constraints", true, "Validate FHIRPath constraints")
	cmd.Flags().Bool("terminology", false, "Validate terminology bindings")
	cmd.Flags().StringP("output", "o", "text", "Output format (text, json)")

	return cmd
}

func newFHIRPathCmd() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "fhirpath [expression] [file]",
		Short: "Evaluate a FHIRPath expression",
		Long: `Evaluate a FHIRPath expression against a FHIR resource.

Examples:
  gofhir fhirpath "Patient.name.given" patient.json
  gofhir fhirpath "Observation.value.ofType(Quantity).value" observation.json
  gofhir fhirpath "Bundle.entry.resource.ofType(Patient)" bundle.json --output json`,
		Args: cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			expression := args[0]
			filePath := args[1]

			// Read the FHIR resource file
			resourceData, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", filePath, err)
			}

			// Compile the expression (with caching for repeated use)
			compiled, err := fhirpath.Compile(expression)
			if err != nil {
				return fmt.Errorf("invalid FHIRPath expression: %w", err)
			}

			// Evaluate the expression
			result, err := compiled.Evaluate(resourceData)
			if err != nil {
				return fmt.Errorf("evaluation error: %w", err)
			}

			// Output the result
			switch outputFormat {
			case "json":
				return outputJSON(result)
			default:
				return outputText(result)
			}
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json)")

	return cmd
}

func outputText(result fhirpath.Collection) error {
	if result.Empty() {
		fmt.Println("(empty)")
		return nil
	}

	for i, value := range result {
		if len(result) > 1 {
			fmt.Printf("[%d] ", i)
		}
		fmt.Println(value.String())
	}
	return nil
}

func outputJSON(result fhirpath.Collection) error {
	if result.Empty() {
		fmt.Println("[]")
		return nil
	}

	// Convert to JSON-serializable format
	output := make([]interface{}, len(result))
	for i, value := range result {
		output[i] = valueToInterface(value)
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

func valueToInterface(v fhirpath.Value) interface{} {
	switch val := v.(type) {
	case interface{ Bool() bool }:
		return val.Bool()
	case interface{ Value() int64 }:
		return val.Value()
	case interface{ Value() string }:
		return val.Value()
	default:
		return v.String()
	}
}

func newGenerateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate Go types from FHIR specifications",
		Long:  `Generate Go types, builders, and utilities from FHIR StructureDefinitions.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			specsDir, err := cmd.Flags().GetString("specs")
			if err != nil {
				return fmt.Errorf("failed to get specs flag: %w", err)
			}
			outputDir, err := cmd.Flags().GetString("output")
			if err != nil {
				return fmt.Errorf("failed to get output flag: %w", err)
			}
			fhirVersion, err := cmd.Flags().GetString("version")
			if err != nil {
				return fmt.Errorf("failed to get version flag: %w", err)
			}

			// Normalize version to lowercase
			fhirVersion = strings.ToLower(fhirVersion)

			versions := []string{fhirVersion}
			if fhirVersion == "all" {
				versions = []string{"r4", "r4b", "r5"}
			}

			for _, v := range versions {
				fmt.Printf("Generating FHIR %s types...\n", strings.ToUpper(v))

				config := generator.Config{
					SpecsDir:    specsDir,
					OutputDir:   filepath.Join(outputDir, v),
					PackageName: v,
					Version:     v,
				}

				gen := generator.New(config)

				fmt.Printf("  Loading StructureDefinitions from %s/%s...\n", specsDir, v)
				if err := gen.LoadTypes(); err != nil {
					return fmt.Errorf("failed to load types for %s: %w", v, err)
				}

				fmt.Printf("  Generating code to %s...\n", config.OutputDir)
				if err := gen.Generate(); err != nil {
					return fmt.Errorf("failed to generate code for %s: %w", v, err)
				}

				fmt.Printf("  Done with %s\n\n", strings.ToUpper(v))
			}

			fmt.Println("Code generation complete!")
			return nil
		},
	}

	cmd.Flags().String("specs", "./specs", "Path to FHIR specifications")
	cmd.Flags().String("output", "./pkg/fhir", "Output directory")
	cmd.Flags().String("version", "r4", "FHIR version to generate (r4, r4b, r5, all)")

	return cmd
}
