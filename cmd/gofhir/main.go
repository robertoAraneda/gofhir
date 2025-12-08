package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
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

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("gofhir version %s\n", version)
	},
}

var validateCmd = &cobra.Command{
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

var fhirpathCmd = &cobra.Command{
	Use:   "fhirpath [expression] [file]",
	Short: "Evaluate a FHIRPath expression",
	Long:  `Evaluate a FHIRPath expression against a FHIR resource.`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement FHIRPath evaluation
		fmt.Printf("Expression: %s\n", args[0])
		fmt.Printf("File: %s\n", args[1])
		fmt.Println("FHIRPath evaluation not yet implemented")
		return nil
	},
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Go types from FHIR specifications",
	Long:  `Generate Go types, builders, and utilities from FHIR StructureDefinitions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		specsDir, _ := cmd.Flags().GetString("specs")
		outputDir, _ := cmd.Flags().GetString("output")
		fhirVersion, _ := cmd.Flags().GetString("version")

		fmt.Printf("Generating FHIR types...\n")
		fmt.Printf("  Specs:   %s\n", specsDir)
		fmt.Printf("  Output:  %s\n", outputDir)
		fmt.Printf("  Version: %s\n", fhirVersion)
		fmt.Println("Code generation not yet implemented")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(fhirpathCmd)
	rootCmd.AddCommand(generateCmd)

	// Validate flags
	validateCmd.Flags().StringP("version", "v", "R4", "FHIR version (R4, R4B, R5)")
	validateCmd.Flags().Bool("constraints", true, "Validate FHIRPath constraints")
	validateCmd.Flags().Bool("terminology", false, "Validate terminology bindings")
	validateCmd.Flags().StringP("output", "o", "text", "Output format (text, json)")

	// Generate flags
	generateCmd.Flags().String("specs", "./specs", "Path to FHIR specifications")
	generateCmd.Flags().String("output", "./pkg/fhir", "Output directory")
	generateCmd.Flags().String("version", "R4", "FHIR version to generate (R4, R4B, R5, all)")
}
