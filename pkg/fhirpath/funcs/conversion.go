package funcs

import (
	"strconv"
	"strings"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
	"github.com/shopspring/decimal"
)

func init() {
	// Register conversion functions
	Register(FuncDef{
		Name:    "iif",
		MinArgs: 2,
		MaxArgs: 3,
		Fn:      fnIif,
	})

	Register(FuncDef{
		Name:    "toBoolean",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnToBoolean,
	})

	Register(FuncDef{
		Name:    "convertsToBoolean",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnConvertsToBoolean,
	})

	Register(FuncDef{
		Name:    "toInteger",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnToInteger,
	})

	Register(FuncDef{
		Name:    "convertsToInteger",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnConvertsToInteger,
	})

	Register(FuncDef{
		Name:    "toDecimal",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnToDecimal,
	})

	Register(FuncDef{
		Name:    "convertsToDecimal",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnConvertsToDecimal,
	})

	Register(FuncDef{
		Name:    "toString",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnToString,
	})

	Register(FuncDef{
		Name:    "convertsToString",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnConvertsToString,
	})

	Register(FuncDef{
		Name:    "toDate",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnToDate,
	})

	Register(FuncDef{
		Name:    "convertsToDate",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnConvertsToDate,
	})

	Register(FuncDef{
		Name:    "toDateTime",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnToDateTime,
	})

	Register(FuncDef{
		Name:    "convertsToDateTime",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnConvertsToDateTime,
	})

	Register(FuncDef{
		Name:    "toTime",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnToTime,
	})

	Register(FuncDef{
		Name:    "convertsToTime",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnConvertsToTime,
	})

	Register(FuncDef{
		Name:    "toQuantity",
		MinArgs: 0,
		MaxArgs: 1,
		Fn:      fnToQuantity,
	})

	Register(FuncDef{
		Name:    "convertsToQuantity",
		MinArgs: 0,
		MaxArgs: 1,
		Fn:      fnConvertsToQuantity,
	})
}

// fnIif returns the second argument if the first is true, otherwise the third.
func fnIif(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if len(args) < 2 {
		return nil, eval.InvalidArgumentsError("iif", 2, len(args))
	}

	// Evaluate the condition
	condition := false
	if cond, ok := args[0].(types.Collection); ok {
		if !cond.Empty() {
			if b, ok := cond[0].(types.Boolean); ok {
				condition = b.Bool()
			}
		}
	}

	if condition {
		// Return the true branch
		if result, ok := args[1].(types.Collection); ok {
			return result, nil
		}
		return types.Collection{}, nil
	}

	// Return the false branch (if provided)
	if len(args) > 2 {
		if result, ok := args[2].(types.Collection); ok {
			return result, nil
		}
	}

	return types.Collection{}, nil
}

// fnToBoolean converts the input to a boolean.
func fnToBoolean(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	item := input[0]

	switch v := item.(type) {
	case types.Boolean:
		return types.Collection{v}, nil
	case types.String:
		str := strings.ToLower(v.Value())
		switch str {
		case "true", "t", "yes", "y", "1", "1.0":
			return types.Collection{types.NewBoolean(true)}, nil
		case "false", "f", "no", "n", "0", "0.0":
			return types.Collection{types.NewBoolean(false)}, nil
		default:
			return types.Collection{}, nil
		}
	case types.Integer:
		if v.Value() == 1 {
			return types.Collection{types.NewBoolean(true)}, nil
		} else if v.Value() == 0 {
			return types.Collection{types.NewBoolean(false)}, nil
		}
		return types.Collection{}, nil
	case types.Decimal:
		if v.Value().Equal(decimal.NewFromInt(1)) {
			return types.Collection{types.NewBoolean(true)}, nil
		} else if v.Value().Equal(decimal.NewFromInt(0)) {
			return types.Collection{types.NewBoolean(false)}, nil
		}
		return types.Collection{}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnConvertsToBoolean returns true if the input can be converted to boolean.
func fnConvertsToBoolean(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{types.NewBoolean(false)}, nil
	}

	item := input[0]

	switch v := item.(type) {
	case types.Boolean:
		return types.Collection{types.NewBoolean(true)}, nil
	case types.String:
		str := strings.ToLower(v.Value())
		switch str {
		case "true", "t", "yes", "y", "1", "1.0", "false", "f", "no", "n", "0", "0.0":
			return types.Collection{types.NewBoolean(true)}, nil
		default:
			return types.Collection{types.NewBoolean(false)}, nil
		}
	case types.Integer:
		if v.Value() == 0 || v.Value() == 1 {
			return types.Collection{types.NewBoolean(true)}, nil
		}
		return types.Collection{types.NewBoolean(false)}, nil
	case types.Decimal:
		if v.Value().Equal(decimal.NewFromInt(0)) || v.Value().Equal(decimal.NewFromInt(1)) {
			return types.Collection{types.NewBoolean(true)}, nil
		}
		return types.Collection{types.NewBoolean(false)}, nil
	default:
		return types.Collection{types.NewBoolean(false)}, nil
	}
}

// fnToInteger converts the input to an integer.
func fnToInteger(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	item := input[0]

	switch v := item.(type) {
	case types.Integer:
		return types.Collection{v}, nil
	case types.Boolean:
		if v.Bool() {
			return types.Collection{types.NewInteger(1)}, nil
		}
		return types.Collection{types.NewInteger(0)}, nil
	case types.String:
		i, err := strconv.ParseInt(v.Value(), 10, 64)
		if err != nil {
			return types.Collection{}, nil
		}
		return types.Collection{types.NewInteger(i)}, nil
	case types.Decimal:
		return types.Collection{types.NewInteger(v.Value().IntPart())}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnConvertsToInteger returns true if the input can be converted to integer.
func fnConvertsToInteger(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{types.NewBoolean(false)}, nil
	}

	item := input[0]

	switch v := item.(type) {
	case types.Integer:
		return types.Collection{types.NewBoolean(true)}, nil
	case types.Boolean:
		return types.Collection{types.NewBoolean(true)}, nil
	case types.String:
		_, err := strconv.ParseInt(v.Value(), 10, 64)
		return types.Collection{types.NewBoolean(err == nil)}, nil
	case types.Decimal:
		return types.Collection{types.NewBoolean(true)}, nil
	default:
		return types.Collection{types.NewBoolean(false)}, nil
	}
}

// fnToDecimal converts the input to a decimal.
func fnToDecimal(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	item := input[0]

	switch v := item.(type) {
	case types.Decimal:
		return types.Collection{v}, nil
	case types.Integer:
		return types.Collection{types.NewDecimalFromInt(v.Value())}, nil
	case types.Boolean:
		if v.Bool() {
			return types.Collection{types.NewDecimalFromInt(1)}, nil
		}
		return types.Collection{types.NewDecimalFromInt(0)}, nil
	case types.String:
		d, err := types.NewDecimal(v.Value())
		if err != nil {
			return types.Collection{}, nil
		}
		return types.Collection{d}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnConvertsToDecimal returns true if the input can be converted to decimal.
func fnConvertsToDecimal(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{types.NewBoolean(false)}, nil
	}

	item := input[0]

	switch v := item.(type) {
	case types.Decimal, types.Integer, types.Boolean:
		return types.Collection{types.NewBoolean(true)}, nil
	case types.String:
		_, err := decimal.NewFromString(v.Value())
		return types.Collection{types.NewBoolean(err == nil)}, nil
	default:
		return types.Collection{types.NewBoolean(false)}, nil
	}
}

// fnToString converts the input to a string.
func fnToString(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	return types.Collection{types.NewString(input[0].String())}, nil
}

// fnConvertsToString returns true if the input can be converted to string.
func fnConvertsToString(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{types.NewBoolean(false)}, nil
	}

	// All primitive types can be converted to string
	switch input[0].(type) {
	case types.String, types.Boolean, types.Integer, types.Decimal:
		return types.Collection{types.NewBoolean(true)}, nil
	default:
		return types.Collection{types.NewBoolean(false)}, nil
	}
}

// fnToDate converts the input to a date.
func fnToDate(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	switch v := input[0].(type) {
	case types.Date:
		return types.Collection{v}, nil
	case types.DateTime:
		// Extract date portion
		d, _ := types.NewDate(v.String()[:10])
		return types.Collection{d}, nil
	case types.String:
		d, err := types.NewDate(v.Value())
		if err != nil {
			return types.Collection{}, nil
		}
		return types.Collection{d}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnConvertsToDate returns true if the input can be converted to date.
func fnConvertsToDate(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{types.NewBoolean(false)}, nil
	}

	// Basic check - will be enhanced with temporal types
	if _, ok := input[0].(types.String); ok {
		return types.Collection{types.NewBoolean(true)}, nil
	}

	return types.Collection{types.NewBoolean(false)}, nil
}

// fnToDateTime converts the input to a datetime.
func fnToDateTime(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	if s, ok := input[0].(types.String); ok {
		return types.Collection{s}, nil
	}

	return types.Collection{}, nil
}

// fnConvertsToDateTime returns true if the input can be converted to datetime.
func fnConvertsToDateTime(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{types.NewBoolean(false)}, nil
	}

	if _, ok := input[0].(types.String); ok {
		return types.Collection{types.NewBoolean(true)}, nil
	}

	return types.Collection{types.NewBoolean(false)}, nil
}

// fnToTime converts the input to a time.
func fnToTime(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	if s, ok := input[0].(types.String); ok {
		return types.Collection{s}, nil
	}

	return types.Collection{}, nil
}

// fnConvertsToTime returns true if the input can be converted to time.
func fnConvertsToTime(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{types.NewBoolean(false)}, nil
	}

	if _, ok := input[0].(types.String); ok {
		return types.Collection{types.NewBoolean(true)}, nil
	}

	return types.Collection{types.NewBoolean(false)}, nil
}

// fnToQuantity converts the input to a quantity.
// Accepts an optional unit argument for Integer/Decimal inputs.
func fnToQuantity(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	// Get optional unit from arguments
	unit := ""
	if len(args) > 0 {
		if argCol, ok := args[0].(types.Collection); ok && !argCol.Empty() {
			if s, ok := argCol[0].(types.String); ok {
				unit = s.Value()
			}
		}
	}

	item := input[0]

	switch v := item.(type) {
	case types.Quantity:
		return types.Collection{v}, nil
	case types.Integer:
		q := types.NewQuantityFromDecimal(decimal.NewFromInt(v.Value()), unit)
		return types.Collection{q}, nil
	case types.Decimal:
		q := types.NewQuantityFromDecimal(v.Value(), unit)
		return types.Collection{q}, nil
	case types.String:
		// Try to parse as quantity string like "5.5 mg" or "10 'kg'"
		q, err := types.NewQuantity(v.Value())
		if err != nil {
			return types.Collection{}, nil
		}
		return types.Collection{q}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnConvertsToQuantity returns true if the input can be converted to quantity.
func fnConvertsToQuantity(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{types.NewBoolean(false)}, nil
	}

	item := input[0]

	switch v := item.(type) {
	case types.Quantity:
		return types.Collection{types.NewBoolean(true)}, nil
	case types.Integer, types.Decimal:
		return types.Collection{types.NewBoolean(true)}, nil
	case types.String:
		// Try to parse as quantity string
		_, err := types.NewQuantity(v.Value())
		return types.Collection{types.NewBoolean(err == nil)}, nil
	default:
		return types.Collection{types.NewBoolean(false)}, nil
	}
}
