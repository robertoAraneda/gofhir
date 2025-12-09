package types

import (
	"fmt"
	"math"

	"github.com/shopspring/decimal"
)

// TypeNameDecimal is the FHIRPath type name for decimal values.
const TypeNameDecimal = "Decimal"

// Decimal represents a FHIRPath decimal value with arbitrary precision.
type Decimal struct {
	value decimal.Decimal
}

// NewDecimal creates a new Decimal from a string.
func NewDecimal(s string) (Decimal, error) {
	d, err := decimal.NewFromString(s)
	if err != nil {
		return Decimal{}, fmt.Errorf("invalid decimal: %s", s)
	}
	return Decimal{value: d}, nil
}

// NewDecimalFromInt creates a new Decimal from an int64.
func NewDecimalFromInt(v int64) Decimal {
	return Decimal{value: decimal.NewFromInt(v)}
}

// NewDecimalFromFloat creates a new Decimal from a float64.
func NewDecimalFromFloat(v float64) Decimal {
	return Decimal{value: decimal.NewFromFloat(v)}
}

// MustDecimal creates a new Decimal, panicking on error.
func MustDecimal(s string) Decimal {
	d, err := NewDecimal(s)
	if err != nil {
		panic(err)
	}
	return d
}

// Value returns the underlying decimal.Decimal value.
func (d Decimal) Value() decimal.Decimal {
	return d.value
}

// Type returns "Decimal".
func (d Decimal) Type() string {
	return TypeNameDecimal
}

// Equal returns true if other is numerically equal.
func (d Decimal) Equal(other Value) bool {
	switch o := other.(type) {
	case Decimal:
		return d.value.Equal(o.value)
	case Integer:
		return d.value.Equal(decimal.NewFromInt(o.value))
	}
	return false
}

// Equivalent is the same as Equal for decimals.
func (d Decimal) Equivalent(other Value) bool {
	return d.Equal(other)
}

// String returns the decimal string representation.
func (d Decimal) String() string {
	return d.value.String()
}

// IsEmpty returns false for decimal values.
func (d Decimal) IsEmpty() bool {
	return false
}

// ToDecimal returns itself (implements Numeric interface).
func (d Decimal) ToDecimal() Decimal {
	return d
}

// Compare compares two numeric values.
func (d Decimal) Compare(other Value) (int, error) {
	switch o := other.(type) {
	case Decimal:
		return d.value.Cmp(o.value), nil
	case Integer:
		return d.value.Cmp(decimal.NewFromInt(o.value)), nil
	}
	return 0, NewTypeError(TypeNameDecimal, other.Type(), "comparison")
}

// Add returns the sum of two decimals.
func (d Decimal) Add(other Decimal) Decimal {
	return Decimal{value: d.value.Add(other.value)}
}

// Subtract returns the difference of two decimals.
func (d Decimal) Subtract(other Decimal) Decimal {
	return Decimal{value: d.value.Sub(other.value)}
}

// Multiply returns the product of two decimals.
func (d Decimal) Multiply(other Decimal) Decimal {
	return Decimal{value: d.value.Mul(other.value)}
}

// Divide returns the result of division.
func (d Decimal) Divide(other Decimal) (Decimal, error) {
	if other.value.IsZero() {
		return Decimal{}, fmt.Errorf("division by zero")
	}
	// Use 16 decimal places of precision
	return Decimal{value: d.value.DivRound(other.value, 16)}, nil
}

// Negate returns the negation of the decimal.
func (d Decimal) Negate() Decimal {
	return Decimal{value: d.value.Neg()}
}

// Abs returns the absolute value.
func (d Decimal) Abs() Decimal {
	return Decimal{value: d.value.Abs()}
}

// Ceiling returns the smallest integer >= d.
func (d Decimal) Ceiling() Integer {
	return NewInteger(d.value.Ceil().IntPart())
}

// Floor returns the largest integer <= d.
func (d Decimal) Floor() Integer {
	return NewInteger(d.value.Floor().IntPart())
}

// Truncate returns the integer part.
func (d Decimal) Truncate() Integer {
	return NewInteger(d.value.Truncate(0).IntPart())
}

// Round rounds to the given precision.
func (d Decimal) Round(precision int32) Decimal {
	return Decimal{value: d.value.Round(precision)}
}

// Power returns d raised to the given power.
func (d Decimal) Power(exp Decimal) Decimal {
	// Convert to float64 for power operation
	base, _ := d.value.Float64()
	exponent, _ := exp.value.Float64()
	result := math.Pow(base, exponent)
	return NewDecimalFromFloat(result)
}

// Sqrt returns the square root.
func (d Decimal) Sqrt() (Decimal, error) {
	if d.value.IsNegative() {
		return Decimal{}, fmt.Errorf("cannot take square root of negative number")
	}
	f, _ := d.value.Float64()
	return NewDecimalFromFloat(math.Sqrt(f)), nil
}

// Exp returns e^d.
func (d Decimal) Exp() Decimal {
	f, _ := d.value.Float64()
	return NewDecimalFromFloat(math.Exp(f))
}

// Ln returns the natural logarithm.
func (d Decimal) Ln() (Decimal, error) {
	if !d.value.IsPositive() {
		return Decimal{}, fmt.Errorf("cannot take logarithm of non-positive number")
	}
	f, _ := d.value.Float64()
	return NewDecimalFromFloat(math.Log(f)), nil
}

// Log returns the logarithm with the given base.
func (d Decimal) Log(base Decimal) (Decimal, error) {
	if !d.value.IsPositive() {
		return Decimal{}, fmt.Errorf("cannot take logarithm of non-positive number")
	}
	if !base.value.IsPositive() || base.value.Equal(decimal.NewFromInt(1)) {
		return Decimal{}, fmt.Errorf("invalid logarithm base")
	}
	f, _ := d.value.Float64()
	b, _ := base.value.Float64()
	return NewDecimalFromFloat(math.Log(f) / math.Log(b)), nil
}

// IsInteger returns true if the decimal has no fractional part.
func (d Decimal) IsInteger() bool {
	return d.value.Equal(d.value.Truncate(0))
}

// ToInteger converts to Integer if it's a whole number.
func (d Decimal) ToInteger() (Integer, bool) {
	if d.IsInteger() {
		return NewInteger(d.value.IntPart()), true
	}
	return Integer{}, false
}
