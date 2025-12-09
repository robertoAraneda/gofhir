package funcs

import (
	"math"

	"github.com/shopspring/decimal"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/eval"
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
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
		// Limit precision to reasonable bounds to avoid overflow
		if p > math.MaxInt32 {
			p = math.MaxInt32
		} else if p < math.MinInt32 {
			p = math.MinInt32
		}
		precision = int32(p) //nolint:gosec // bounds checked above
	}

	switch v := input[0].(type) {
	case types.Integer:
		return types.Collection{v}, nil
	case types.Decimal:
		rounded := v.Value().Round(precision)
		d, err := types.NewDecimal(rounded.String())
		if err != nil {
			return types.Collection{}, nil
		}
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
		d, err := types.NewDecimal(sum.String())
		if err != nil {
			return types.Collection{}, nil
		}
		return types.Collection{d}, nil
	}
	return types.Collection{types.NewInteger(sum.IntPart())}, nil
}

// findExtreme finds either the minimum or maximum value in a collection.
// When findMin is true, finds minimum; otherwise finds maximum.
func findExtreme(ctx *eval.Context, input types.Collection, findMin bool) (types.Collection, error) {
	if input.Empty() {
		return types.Collection{}, nil
	}

	// Check for cancellation
	if err := ctx.CheckCancellation(); err != nil {
		return nil, err
	}

	var extremeVal types.Value
	var extremeFloat float64
	first := true
	isNumeric := false

	// compareFn returns true if newVal should replace current extreme
	compareFn := func(newVal, currentVal float64) bool {
		if findMin {
			return newVal < currentVal
		}
		return newVal > currentVal
	}

	// compareStrFn returns true if newStr should replace current extreme
	compareStrFn := func(newStr, currentStr string) bool {
		if findMin {
			return newStr < currentStr
		}
		return newStr > currentStr
	}

	// compareCmpFn returns true if cmp result indicates newVal should replace
	compareCmpFn := func(cmp int) bool {
		if findMin {
			return cmp < 0
		}
		return cmp > 0
	}

	for _, item := range input {
		switch v := item.(type) {
		case types.Integer:
			val := float64(v.Value())
			if first {
				extremeFloat = val
				extremeVal = item
				first = false
				isNumeric = true
			} else if isNumeric && compareFn(val, extremeFloat) {
				extremeFloat = val
				extremeVal = item
			}
		case types.Decimal:
			val := v.Value().InexactFloat64()
			if first {
				extremeFloat = val
				extremeVal = item
				first = false
				isNumeric = true
			} else if isNumeric && compareFn(val, extremeFloat) {
				extremeFloat = val
				extremeVal = item
			}
		case types.String:
			if first {
				extremeVal = v
				first = false
			} else if extremeStr, ok := extremeVal.(types.String); ok {
				if compareStrFn(v.Value(), extremeStr.Value()) {
					extremeVal = v
				}
			}
		case types.Date:
			if first {
				extremeVal = v
				first = false
			} else if extremeDate, ok := extremeVal.(types.Date); ok {
				cmp, err := v.Compare(extremeDate)
				if err == nil && compareCmpFn(cmp) {
					extremeVal = v
				}
			}
		case types.DateTime:
			if first {
				extremeVal = v
				first = false
			} else if extremeDT, ok := extremeVal.(types.DateTime); ok {
				cmp, err := v.Compare(extremeDT)
				if err == nil && compareCmpFn(cmp) {
					extremeVal = v
				}
			}
		case types.Time:
			if first {
				extremeVal = v
				first = false
			} else if extremeTime, ok := extremeVal.(types.Time); ok {
				cmp, err := v.Compare(extremeTime)
				if err == nil && compareCmpFn(cmp) {
					extremeVal = v
				}
			}
		default:
			return types.Collection{}, nil
		}
	}

	if extremeVal == nil {
		return types.Collection{}, nil
	}
	return types.Collection{extremeVal}, nil
}

// fnMin returns the minimum value in the collection.
// Returns empty if the collection is empty.
func fnMin(ctx *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	return findExtreme(ctx, input, true)
}

// fnMax returns the maximum value in the collection.
// Returns empty if the collection is empty.
func fnMax(ctx *eval.Context, input types.Collection, _ []interface{}) (types.Collection, error) {
	return findExtreme(ctx, input, false)
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
	d, err := types.NewDecimal(avg.String())
	if err != nil {
		return types.Collection{}, nil
	}
	return types.Collection{d}, nil
}
