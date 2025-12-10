// Package funcs provides FHIRPath function implementations.
// This file contains type checking functions: is() and as()
//
// According to FHIRPath specification:
// - is(type): Returns true if the input is of the specified type
// - as(type): Returns the input if it is of the specified type, otherwise empty
//
// These functions are equivalent to the 'is' and 'as' operators but in function form.
// Example: Patient.name.first().is(HumanName) is equivalent to Patient.name.first() is HumanName
package funcs

import (
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func init() {
	// Register type checking functions
	// Note: These are handled specially in the evaluator to extract type names
	// directly from the expression AST, rather than evaluating them as expressions.
	// This is necessary because type names like "Composition" or "Patient" would
	// otherwise be interpreted as path expressions.
	Register(FuncDef{
		Name:    "is",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnIsType,
	})

	// Note: as() with function syntax is also handled specially in the evaluator.
	// The fnAs in aggregate.go handles evaluated string arguments,
	// but the evaluator intercepts as(TypeName) calls directly.
}

// fnIsType is the function implementation for is().
// Note: This is typically not called directly - the evaluator handles is() specially
// to extract type names from the AST. This stub exists for completeness.
func fnIsType(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) == 0 {
		return nil, eval.InvalidArgumentsError("is", 1, 0)
	}

	// Empty input returns empty
	if input.Empty() {
		return types.Collection{}, nil
	}

	// is() requires singleton input
	if len(input) != 1 {
		return nil, eval.SingletonError(len(input))
	}

	// Try to extract type name from argument
	typeName := extractTypeName(args[0])
	if typeName == "" {
		return types.Collection{}, nil
	}

	// Get actual type
	actualType := input[0].Type()

	// Use the exported TypeMatches function from eval package
	matches := eval.TypeMatches(actualType, typeName)
	return types.Collection{types.NewBoolean(matches)}, nil
}

// extractTypeName extracts a type name from a function argument.
func extractTypeName(arg interface{}) string {
	switch v := arg.(type) {
	case string:
		return v
	case types.String:
		return v.Value()
	case types.Collection:
		if len(v) > 0 {
			if s, ok := v[0].(types.String); ok {
				return s.Value()
			}
		}
	}
	return ""
}
