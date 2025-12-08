// Package fhirpath provides a FHIRPath engine for evaluating expressions on FHIR resources.
package fhirpath

import (
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

// Evaluate parses and evaluates a FHIRPath expression against a JSON resource.
// This is a convenience function that compiles and evaluates in one step.
func Evaluate(resource []byte, expr string) (types.Collection, error) {
	compiled, err := Compile(expr)
	if err != nil {
		return nil, err
	}
	return compiled.Evaluate(resource)
}

// MustEvaluate is like Evaluate but panics on error.
func MustEvaluate(resource []byte, expr string) types.Collection {
	result, err := Evaluate(resource, expr)
	if err != nil {
		panic(err)
	}
	return result
}

// Compile parses a FHIRPath expression and returns a compiled Expression.
// The compiled expression can be evaluated multiple times against different resources.
func Compile(expr string) (*Expression, error) {
	return compile(expr)
}

// MustCompile is like Compile but panics on error.
func MustCompile(expr string) *Expression {
	compiled, err := Compile(expr)
	if err != nil {
		panic(err)
	}
	return compiled
}
