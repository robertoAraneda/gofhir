package funcs

import (
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func init() {
	// Register filtering functions
	Register(FuncDef{
		Name:    "where",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnWhere,
	})

	Register(FuncDef{
		Name:    "select",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnSelect,
	})

	Register(FuncDef{
		Name:    "repeat",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnRepeat,
	})

	Register(FuncDef{
		Name:    "ofType",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnOfType,
	})
}

// fnWhere filters the collection based on a criteria expression.
// Returns elements where the criteria evaluates to true.
func fnWhere(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) == 0 {
		return nil, eval.InvalidArgumentsError("where", 1, 0)
	}

	// The argument should be an evaluated collection for each element
	// For now, we expect args[0] to be a function that evaluates the criteria
	// This is handled specially in the evaluator

	// If we receive pre-evaluated results (collection of booleans), filter based on them
	if criteria, ok := args[0].(types.Collection); ok {
		result := types.Collection{}
		for i, item := range input {
			if i < len(criteria) {
				if b, ok := criteria[i].(types.Boolean); ok && b.Bool() {
					result = append(result, item)
				}
			}
		}
		return result, nil
	}

	// Default: return input (criteria evaluation should be handled by evaluator)
	return input, nil
}

// fnSelect projects each element using an expression.
// Returns the flattened results of evaluating the expression on each element.
func fnSelect(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) == 0 {
		return nil, eval.InvalidArgumentsError("select", 1, 0)
	}

	// The argument should be evaluated for each element
	// This is handled specially in the evaluator
	if results, ok := args[0].(types.Collection); ok {
		return results, nil
	}

	return types.Collection{}, nil
}

// fnRepeat repeatedly applies an expression until no new results are found.
func fnRepeat(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) == 0 {
		return nil, eval.InvalidArgumentsError("repeat", 1, 0)
	}

	// This requires special handling in the evaluator for recursive evaluation
	// For now, return the input
	return input, nil
}

// fnOfType filters elements by type.
func fnOfType(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) == 0 {
		return nil, eval.InvalidArgumentsError("ofType", 1, 0)
	}

	// Get the type name
	typeName := ""
	switch v := args[0].(type) {
	case types.Collection:
		if len(v) > 0 {
			if s, ok := v[0].(types.String); ok {
				typeName = s.Value()
			}
		}
	case types.String:
		typeName = v.Value()
	case string:
		typeName = v
	}

	if typeName == "" {
		return types.Collection{}, nil
	}

	result := types.Collection{}
	for _, item := range input {
		if item.Type() == typeName {
			result = append(result, item)
		}
	}

	return result, nil
}
