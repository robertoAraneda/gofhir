package funcs

import (
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func init() {
	// Register subsetting functions
	Register(FuncDef{
		Name:        "first",
		MinArgs:     0,
		MaxArgs:     0,
		Fn:          fnFirst,
		
	})

	Register(FuncDef{
		Name:        "last",
		MinArgs:     0,
		MaxArgs:     0,
		Fn:          fnLast,
		
	})

	Register(FuncDef{
		Name:        "tail",
		MinArgs:     0,
		MaxArgs:     0,
		Fn:          fnTail,
		
	})

	Register(FuncDef{
		Name:        "skip",
		MinArgs:     1,
		MaxArgs:     1,
		Fn:          fnSkip,
		
	})

	Register(FuncDef{
		Name:        "take",
		MinArgs:     1,
		MaxArgs:     1,
		Fn:          fnTake,
		
	})

	Register(FuncDef{
		Name:        "single",
		MinArgs:     0,
		MaxArgs:     0,
		Fn:          fnSingle,
		
	})

	Register(FuncDef{
		Name:        "intersect",
		MinArgs:     1,
		MaxArgs:     1,
		Fn:          fnIntersect,
		
	})

	Register(FuncDef{
		Name:        "exclude",
		MinArgs:     1,
		MaxArgs:     1,
		Fn:          fnExclude,
		
	})
}

// fnFirst returns the first element of the collection.
func fnFirst(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if first, ok := input.First(); ok {
		return types.Collection{first}, nil
	}
	return types.Collection{}, nil
}

// fnLast returns the last element of the collection.
func fnLast(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if last, ok := input.Last(); ok {
		return types.Collection{last}, nil
	}
	return types.Collection{}, nil
}

// fnTail returns all elements except the first.
func fnTail(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	return input.Tail(), nil
}

// fnSkip returns elements after skipping the first n.
func fnSkip(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) == 0 {
		return nil, eval.InvalidArgumentsError("skip", 1, 0)
	}

	n, err := toInteger(args[0])
	if err != nil {
		return nil, err
	}

	return input.Skip(int(n)), nil
}

// fnTake returns the first n elements.
func fnTake(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) == 0 {
		return nil, eval.InvalidArgumentsError("take", 1, 0)
	}

	n, err := toInteger(args[0])
	if err != nil {
		return nil, err
	}

	return input.Take(int(n)), nil
}

// fnSingle returns the single element or errors if not exactly one.
func fnSingle(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	single, err := input.Single()
	if err != nil {
		return nil, eval.NewEvalError(eval.ErrSingletonExpected, err.Error())
	}
	return types.Collection{single}, nil
}

// fnIntersect returns elements that are in both collections.
func fnIntersect(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) == 0 {
		return nil, eval.InvalidArgumentsError("intersect", 1, 0)
	}

	other, ok := args[0].(types.Collection)
	if !ok {
		return nil, eval.TypeError("Collection", "unknown", "intersect")
	}

	return input.Intersect(other), nil
}

// fnExclude returns elements not in the other collection.
func fnExclude(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) == 0 {
		return nil, eval.InvalidArgumentsError("exclude", 1, 0)
	}

	other, ok := args[0].(types.Collection)
	if !ok {
		return nil, eval.TypeError("Collection", "unknown", "exclude")
	}

	return input.Exclude(other), nil
}

// toInteger converts an argument to int64.
func toInteger(arg interface{}) (int64, error) {
	switch v := arg.(type) {
	case types.Collection:
		if v.Empty() {
			return 0, eval.NewEvalError(eval.ErrType, "expected integer, got empty collection")
		}
		if i, ok := v[0].(types.Integer); ok {
			return i.Value(), nil
		}
		return 0, eval.TypeError("Integer", v[0].Type(), "argument")
	case types.Integer:
		return v.Value(), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	default:
		return 0, eval.NewEvalError(eval.ErrType, "expected integer")
	}
}
