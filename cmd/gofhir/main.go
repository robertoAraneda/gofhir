package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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
	return &cobra.Command{
		Use:   "fhirpath [expression] [file]",
		Short: "Evaluate a FHIRPath expression",
		Long:  `Evaluate a FHIRPath expression against a FHIR resource.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			// TODO: Implement FHIRPath evaluation
			fmt.Printf("Expression: %s\n", args[0])
			fmt.Printf("File: %s\n", args[1])
			fmt.Println("FHIRPath evaluation not yet implemented")
			return nil
		},
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

			fmt.Printf("Generating FHIR types...\n")
			fmt.Printf("  Specs:   %s\n", specsDir)
			fmt.Printf("  Output:  %s\n", outputDir)
			fmt.Printf("  Version: %s\n", fhirVersion)
			fmt.Println("Code generation not yet implemented")
			return nil
		},
	}

	cmd.Flags().String("specs", "./specs", "Path to FHIR specifications")
	cmd.Flags().String("output", "./pkg/fhir", "Output directory")
	cmd.Flags().String("version", "R4", "FHIR version to generate (R4, R4B, R5, all)")

	return cmd
}
