package funcs

import (
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func init() {
	// Register existence functions
	Register(FuncDef{
		Name:    "empty",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnEmpty,
	})

	Register(FuncDef{
		Name:    "exists",
		MinArgs: 0,
		MaxArgs: 1,
		Fn:      fnExists,
	})

	Register(FuncDef{
		Name:    "all",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnAll,
	})

	Register(FuncDef{
		Name:    "allTrue",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnAllTrue,
	})

	Register(FuncDef{
		Name:    "anyTrue",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnAnyTrue,
	})

	Register(FuncDef{
		Name:    "allFalse",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnAllFalse,
	})

	Register(FuncDef{
		Name:    "anyFalse",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnAnyFalse,
	})

	Register(FuncDef{
		Name:    "count",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnCount,
	})

	Register(FuncDef{
		Name:    "distinct",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnDistinct,
	})

	Register(FuncDef{
		Name:    "isDistinct",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnIsDistinct,
	})

	Register(FuncDef{
		Name:    "subsetOf",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnSubsetOf,
	})

	Register(FuncDef{
		Name:    "supersetOf",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnSupersetOf,
	})
}

// fnEmpty returns true if the collection is empty.
func fnEmpty(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.TrueCollection, nil
	}
	return types.FalseCollection, nil
}

// fnExists returns true if the collection is not empty.
// With criteria, the evaluation is handled by the evaluator.
func fnExists(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.FalseCollection, nil
	}
	return types.TrueCollection, nil
}

// fnAll returns true if all elements match the criteria.
// Criteria evaluation is handled by the evaluator.
// Empty collection returns true (vacuous truth).
func fnAll(ctx *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	return types.TrueCollection, nil
}

// fnAllTrue returns true if all items are boolean true.
func fnAllTrue(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() || input.AllTrue() {
		return types.TrueCollection, nil
	}
	return types.FalseCollection, nil
}

// fnAnyTrue returns true if any item is boolean true.
func fnAnyTrue(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if !input.Empty() && input.AnyTrue() {
		return types.TrueCollection, nil
	}
	return types.FalseCollection, nil
}

// fnAllFalse returns true if all items are boolean false.
func fnAllFalse(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() || input.AllFalse() {
		return types.TrueCollection, nil
	}
	return types.FalseCollection, nil
}

// fnAnyFalse returns true if any item is boolean false.
func fnAnyFalse(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if !input.Empty() && input.AnyFalse() {
		return types.TrueCollection, nil
	}
	return types.FalseCollection, nil
}

// fnCount returns the number of items in the collection.
func fnCount(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	return types.Collection{types.GetInteger(int64(input.Count()))}, nil
}

// fnDistinct returns a collection with duplicates removed.
func fnDistinct(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	return input.Distinct(), nil
}

// fnIsDistinct returns true if all items are distinct.
func fnIsDistinct(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.IsDistinct() {
		return types.TrueCollection, nil
	}
	return types.FalseCollection, nil
}

// fnSubsetOf returns true if all items in input are in other.
func fnSubsetOf(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) == 0 {
		return nil, eval.InvalidArgumentsError("subsetOf", 1, 0)
	}

	other, ok := args[0].(types.Collection)
	if !ok {
		return nil, eval.TypeError("Collection", "unknown", "subsetOf")
	}

	// All items in input must be in other
	for _, item := range input {
		if !other.Contains(item) {
			return types.FalseCollection, nil
		}
	}

	return types.TrueCollection, nil
}

// fnSupersetOf returns true if all items in other are in input.
func fnSupersetOf(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) == 0 {
		return nil, eval.InvalidArgumentsError("supersetOf", 1, 0)
	}

	other, ok := args[0].(types.Collection)
	if !ok {
		return nil, eval.TypeError("Collection", "unknown", "supersetOf")
	}

	// All items in other must be in input
	for _, item := range other {
		if !input.Contains(item) {
			return types.FalseCollection, nil
		}
	}

	return types.TrueCollection, nil
}
