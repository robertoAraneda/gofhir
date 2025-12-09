package types

import (
	"fmt"
	"math"

	"github.com/shopspring/decimal"
)

// Integer represents a FHIRPath integer value.
type Integer struct {
	value int64
}

// NewInteger creates a new Integer value.
func NewInteger(v int64) Integer {
	return Integer{value: v}
}

// Value returns the underlying int64 value.
func (i Integer) Value() int64 {
	return i.value
}

// Type returns "Integer".
func (i Integer) Type() string {
	return "Integer"
}

// Equal returns true if other is an Integer with the same value,
// or a Decimal with an equivalent integer value.
func (i Integer) Equal(other Value) bool {
	switch o := other.(type) {
	case Integer:
		return i.value == o.value
	case Decimal:
		// Compare as decimals for cross-type equality
		return i.ToDecimal().Equal(o)
	}
	return false
}

// Equivalent is the same as Equal for integers.
func (i Integer) Equivalent(other Value) bool {
	return i.Equal(other)
}

// String returns the decimal string representation.
func (i Integer) String() string {
	return fmt.Sprintf("%d", i.value)
}

// IsEmpty returns false for integer values.
func (i Integer) IsEmpty() bool {
	return false
}

// ToDecimal converts the integer to a Decimal.
func (i Integer) ToDecimal() Decimal {
	return Decimal{value: decimal.NewFromInt(i.value)}
}

// Compare compares two numeric values.
func (i Integer) Compare(other Value) (int, error) {
	switch o := other.(type) {
	case Integer:
		if i.value < o.value {
			return -1, nil
		}
		if i.value > o.value {
			return 1, nil
		}
		return 0, nil
	case Decimal:
		return i.ToDecimal().Compare(o)
	}
	return 0, NewTypeError("Integer", other.Type(), "comparison")
}

// Add returns the sum of two integers.
func (i Integer) Add(other Integer) Integer {
	return NewInteger(i.value + other.value)
}

// Subtract returns the difference of two integers.
func (i Integer) Subtract(other Integer) Integer {
	return NewInteger(i.value - other.value)
}

// Multiply returns the product of two integers.
func (i Integer) Multiply(other Integer) Integer {
	return NewInteger(i.value * other.value)
}

// Divide returns the result of division as a Decimal.
func (i Integer) Divide(other Integer) (Decimal, error) {
	if other.value == 0 {
		return Decimal{}, fmt.Errorf("division by zero")
	}
	return i.ToDecimal().Divide(other.ToDecimal())
}

// Div returns the integer division result.
func (i Integer) Div(other Integer) (Integer, error) {
	if other.value == 0 {
		return Integer{}, fmt.Errorf("division by zero")
	}
	return NewInteger(i.value / other.value), nil
}

// Mod returns the modulo result.
func (i Integer) Mod(other Integer) (Integer, error) {
	if other.value == 0 {
		return Integer{}, fmt.Errorf("division by zero")
	}
	return NewInteger(i.value % other.value), nil
}

// Negate returns the negation of the integer.
func (i Integer) Negate() Integer {
	return NewInteger(-i.value)
}

// Abs returns the absolute value.
func (i Integer) Abs() Integer {
	if i.value < 0 {
		return NewInteger(-i.value)
	}
	return i
}

// Power returns the integer raised to the given power.
func (i Integer) Power(exp Integer) Decimal {
	return i.ToDecimal().Power(exp.ToDecimal())
}

// Sqrt returns the square root as a Decimal.
func (i Integer) Sqrt() (Decimal, error) {
	if i.value < 0 {
		return Decimal{}, fmt.Errorf("cannot take square root of negative number")
	}
	return NewDecimalFromFloat(math.Sqrt(float64(i.value))), nil
}
