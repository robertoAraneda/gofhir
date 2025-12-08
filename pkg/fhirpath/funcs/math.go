package funcs

import (
	"math"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
	"github.com/shopspring/decimal"
)

func init() {
	// Register math functions
	Register(FuncDef{
		Name:    "abs",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnAbs,
	})

	Register(FuncDef{
		Name:    "ceiling",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnCeiling,
	})

	Register(FuncDef{
		Name:    "exp",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnExp,
	})

	Register(FuncDef{
		Name:    "floor",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnFloor,
	})

	Register(FuncDef{
		Name:    "ln",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnLn,
	})

	Register(FuncDef{
		Name:    "log",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnLog,
	})

	Register(FuncDef{
		Name:    "power",
		MinArgs: 1,
		MaxArgs: 1,
		Fn:      fnPower,
	})

	Register(FuncDef{
		Name:    "round",
		MinArgs: 0,
		MaxArgs: 1,
		Fn:      fnRound,
	})

	Register(FuncDef{
		Name:    "sqrt",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnSqrt,
	})

	Register(FuncDef{
		Name:    "truncate",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnTruncate,
	})

	// Aggregate functions
	Register(FuncDef{
		Name:    "sum",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnSum,
	})

	Register(FuncDef{
		Name:    "min",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnMin,
	})

	Register(FuncDef{
		Name:    "max",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnMax,
	})

	Register(FuncDef{
		Name:    "avg",
		MinArgs: 0,
		MaxArgs: 0,
		Fn:      fnAvg,
	})
}

// fnAbs returns the absolute value.
func fnAbs(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	switch v := input[0].(type) {
	case types.Integer:
		val := v.Value()
		if val < 0 {
			val = -val
		}
		return types.Collection{types.NewInteger(val)}, nil
	case types.Decimal:
		return types.Collection{types.NewDecimalFromFloat(math.Abs(v.Value().InexactFloat64()))}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnCeiling returns the smallest integer >= input.
func fnCeiling(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	switch v := input[0].(type) {
	case types.Integer:
		return types.Collection{v}, nil
	case types.Decimal:
		ceil := math.Ceil(v.Value().InexactFloat64())
		return types.Collection{types.NewInteger(int64(ceil))}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnExp returns e raised to the power of input.
func fnExp(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	var val float64
	switch v := input[0].(type) {
	case types.Integer:
		val = float64(v.Value())
	case types.Decimal:
		val = v.Value().InexactFloat64()
	default:
		return types.Collection{}, nil
	}

	return types.Collection{types.NewDecimalFromFloat(math.Exp(val))}, nil
}

// fnFloor returns the largest integer <= input.
func fnFloor(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	switch v := input[0].(type) {
	case types.Integer:
		return types.Collection{v}, nil
	case types.Decimal:
		floor := math.Floor(v.Value().InexactFloat64())
		return types.Collection{types.NewInteger(int64(floor))}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnLn returns the natural logarithm.
func fnLn(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	var val float64
	switch v := input[0].(type) {
	case types.Integer:
		val = float64(v.Value())
	case types.Decimal:
		val = v.Value().InexactFloat64()
	default:
		return types.Collection{}, nil
	}

	if val <= 0 {
		return types.Collection{}, nil
	}

	return types.Collection{types.NewDecimalFromFloat(math.Log(val))}, nil
}

// fnLog returns the logarithm with the given base.
func fnLog(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() || len(args) == 0 {
		return types.Collection{}, nil
	}

	var val float64
	switch v := input[0].(type) {
	case types.Integer:
		val = float64(v.Value())
	case types.Decimal:
		val = v.Value().InexactFloat64()
	default:
		return types.Collection{}, nil
	}

	base, err := toFloat(args[0])
	if err != nil {
		return types.Collection{}, nil
	}

	if val <= 0 || base <= 0 || base == 1 {
		return types.Collection{}, nil
	}

	result := math.Log(val) / math.Log(base)
	return types.Collection{types.NewDecimalFromFloat(result)}, nil
}

// fnPower returns input raised to the given power.
func fnPower(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() || len(args) == 0 {
		return types.Collection{}, nil
	}

	var base float64
	switch v := input[0].(type) {
	case types.Integer:
		base = float64(v.Value())
	case types.Decimal:
		base = v.Value().InexactFloat64()
	default:
		return types.Collection{}, nil
	}

	exp, err := toFloat(args[0])
	if err != nil {
		return types.Collection{}, nil
	}

	result := math.Pow(base, exp)

	// Check for invalid results
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return types.Collection{}, nil
	}

	return types.Collection{types.NewDecimalFromFloat(result)}, nil
}

// fnRound rounds to the specified number of decimal places.
func fnRound(_ *eval.Context, input types.Collection, args []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	precision := int32(0)
	if len(args) > 0 {
		p, err := toInteger(args[0])
		if err != nil {
			return types.Collection{}, nil
		}
		precision = int32(p)
	}

	switch v := input[0].(type) {
	case types.Integer:
		return types.Collection{v}, nil
	case types.Decimal:
		rounded := v.Value().Round(precision)
		d, _ := types.NewDecimal(rounded.String())
		return types.Collection{d}, nil
	default:
		return types.Collection{}, nil
	}
}

// fnSqrt returns the square root.
func fnSqrt(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	var val float64
	switch v := input[0].(type) {
	case types.Integer:
		val = float64(v.Value())
	case types.Decimal:
		val = v.Value().InexactFloat64()
	default:
		return types.Collection{}, nil
	}

	if val < 0 {
		return types.Collection{}, nil
	}

	return types.Collection{types.NewDecimalFromFloat(math.Sqrt(val))}, nil
}

// fnTruncate returns the integer part (truncates toward zero).
func fnTruncate(_ *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	switch v := input[0].(type) {
	case types.Integer:
		return types.Collection{v}, nil
	case types.Decimal:
		trunc := math.Trunc(v.Value().InexactFloat64())
		return types.Collection{types.NewInteger(int64(trunc))}, nil
	default:
		return types.Collection{}, nil
	}
}

// toFloat converts an argument to float64.
func toFloat(arg interface{}) (float64, error) {
	switch v := arg.(type) {
	case types.Collection:
		if v.Empty() {
			return 0, eval.NewEvalError(eval.ErrType, "expected number, got empty collection")
		}
		return toFloat(v[0])
	case types.Integer:
		return float64(v.Value()), nil
	case types.Decimal:
		return v.Value().InexactFloat64(), nil
	case int64:
		return float64(v), nil
	case float64:
		return v, nil
	case decimal.Decimal:
		return v.InexactFloat64(), nil
	default:
		return 0, eval.NewEvalError(eval.ErrType, "expected number")
	}
}

// fnSum returns the sum of all numeric values in the collection.
// Returns empty if the collection is empty or contains non-numeric values.
func fnSum(ctx *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{types.NewInteger(0)}, nil
	}

	// Check for cancellation in large collections
	if err := ctx.CheckCancellation(); err != nil {
		return nil, err
	}

	var sum decimal.Decimal
	hasDecimal := false

	for _, item := range input {
		switch v := item.(type) {
		case types.Integer:
			sum = sum.Add(decimal.NewFromInt(v.Value()))
		case types.Decimal:
			sum = sum.Add(v.Value())
			hasDecimal = true
		default:
			// Non-numeric value - return empty per FHIRPath spec
			return types.Collection{}, nil
		}
	}

	// Return Integer if all inputs were Integer, otherwise Decimal
	if hasDecimal {
		d, _ := types.NewDecimal(sum.String())
		return types.Collection{d}, nil
	}
	return types.Collection{types.NewInteger(sum.IntPart())}, nil
}

// fnMin returns the minimum value in the collection.
// Returns empty if the collection is empty.
func fnMin(ctx *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	// Check for cancellation
	if err := ctx.CheckCancellation(); err != nil {
		return nil, err
	}

	var minVal types.Value
	var minFloat float64
	first := true
	isNumeric := false

	for _, item := range input {
		switch v := item.(type) {
		case types.Integer:
			val := float64(v.Value())
			if first {
				minFloat = val
				minVal = item
				first = false
				isNumeric = true
			} else if isNumeric && val < minFloat {
				minFloat = val
				minVal = item
			}
		case types.Decimal:
			val := v.Value().InexactFloat64()
			if first {
				minFloat = val
				minVal = item
				first = false
				isNumeric = true
			} else if isNumeric && val < minFloat {
				minFloat = val
				minVal = item
			}
		case types.String:
			// String comparison
			if first {
				minVal = v
				first = false
			} else if minStr, ok := minVal.(types.String); ok {
				if v.Value() < minStr.Value() {
					minVal = v
				}
			}
		case types.Date:
			// Date comparison using Compare method
			if first {
				minVal = v
				first = false
			} else if minDate, ok := minVal.(types.Date); ok {
				cmp, _ := v.Compare(minDate)
				if cmp < 0 {
					minVal = v
				}
			}
		case types.DateTime:
			// DateTime comparison using Compare method
			if first {
				minVal = v
				first = false
			} else if minDT, ok := minVal.(types.DateTime); ok {
				cmp, _ := v.Compare(minDT)
				if cmp < 0 {
					minVal = v
				}
			}
		case types.Time:
			// Time comparison using Compare method
			if first {
				minVal = v
				first = false
			} else if minTime, ok := minVal.(types.Time); ok {
				cmp, _ := v.Compare(minTime)
				if cmp < 0 {
					minVal = v
				}
			}
		default:
			return types.Collection{}, nil
		}
	}

	if minVal == nil {
		return types.Collection{}, nil
	}
	return types.Collection{minVal}, nil
}

// fnMax returns the maximum value in the collection.
// Returns empty if the collection is empty.
func fnMax(ctx *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	// Check for cancellation
	if err := ctx.CheckCancellation(); err != nil {
		return nil, err
	}

	var maxVal types.Value
	var maxFloat float64
	first := true
	isNumeric := false

	for _, item := range input {
		switch v := item.(type) {
		case types.Integer:
			val := float64(v.Value())
			if first {
				maxFloat = val
				maxVal = item
				first = false
				isNumeric = true
			} else if isNumeric && val > maxFloat {
				maxFloat = val
				maxVal = item
			}
		case types.Decimal:
			val := v.Value().InexactFloat64()
			if first {
				maxFloat = val
				maxVal = item
				first = false
				isNumeric = true
			} else if isNumeric && val > maxFloat {
				maxFloat = val
				maxVal = item
			}
		case types.String:
			// String comparison
			if first {
				maxVal = v
				first = false
			} else if maxStr, ok := maxVal.(types.String); ok {
				if v.Value() > maxStr.Value() {
					maxVal = v
				}
			}
		case types.Date:
			// Date comparison using Compare method
			if first {
				maxVal = v
				first = false
			} else if maxDate, ok := maxVal.(types.Date); ok {
				cmp, _ := v.Compare(maxDate)
				if cmp > 0 {
					maxVal = v
				}
			}
		case types.DateTime:
			// DateTime comparison using Compare method
			if first {
				maxVal = v
				first = false
			} else if maxDT, ok := maxVal.(types.DateTime); ok {
				cmp, _ := v.Compare(maxDT)
				if cmp > 0 {
					maxVal = v
				}
			}
		case types.Time:
			// Time comparison using Compare method
			if first {
				maxVal = v
				first = false
			} else if maxTime, ok := maxVal.(types.Time); ok {
				cmp, _ := v.Compare(maxTime)
				if cmp > 0 {
					maxVal = v
				}
			}
		default:
			return types.Collection{}, nil
		}
	}

	if maxVal == nil {
		return types.Collection{}, nil
	}
	return types.Collection{maxVal}, nil
}

// fnAvg returns the average of all numeric values in the collection.
// Returns empty if the collection is empty or contains non-numeric values.
func fnAvg(ctx *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	// Check for cancellation
	if err := ctx.CheckCancellation(); err != nil {
		return nil, err
	}

	var sum decimal.Decimal
	count := 0

	for _, item := range input {
		switch v := item.(type) {
		case types.Integer:
			sum = sum.Add(decimal.NewFromInt(v.Value()))
			count++
		case types.Decimal:
			sum = sum.Add(v.Value())
			count++
		default:
			// Non-numeric value - return empty per FHIRPath spec
			return types.Collection{}, nil
		}
	}

	if count == 0 {
		return types.Collection{}, nil
	}

	avg := sum.Div(decimal.NewFromInt(int64(count)))
	d, _ := types.NewDecimal(avg.String())
	return types.Collection{d}, nil
}
