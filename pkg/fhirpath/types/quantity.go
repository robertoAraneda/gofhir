package types

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"
)

// Quantity represents a FHIRPath quantity value with a numeric value and unit.
type Quantity struct {
	value decimal.Decimal
	unit  string
}

// Quantity regex pattern: number followed by optional unit
var quantityPattern = regexp.MustCompile(`^([+-]?\d+\.?\d*)\s*(?:'([^']+)'|(\S+))?$`)

// NewQuantity creates a Quantity from a string.
func NewQuantity(s string) (Quantity, error) {
	matches := quantityPattern.FindStringSubmatch(strings.TrimSpace(s))
	if matches == nil {
		return Quantity{}, fmt.Errorf("invalid quantity format: %s", s)
	}

	val, err := decimal.NewFromString(matches[1])
	if err != nil {
		return Quantity{}, fmt.Errorf("invalid quantity value: %s", matches[1])
	}

	unit := ""
	if matches[2] != "" {
		unit = matches[2] // Quoted unit
	} else if matches[3] != "" {
		unit = matches[3] // Unquoted unit
	}

	return Quantity{value: val, unit: unit}, nil
}

// NewQuantityFromDecimal creates a Quantity from a decimal value and unit.
func NewQuantityFromDecimal(value decimal.Decimal, unit string) Quantity {
	return Quantity{value: value, unit: unit}
}

// Type returns the type name.
func (q Quantity) Type() string {
	return "Quantity"
}

// Equal checks equality with another value.
func (q Quantity) Equal(other Value) bool {
	if o, ok := other.(Quantity); ok {
		return q.value.Equal(o.value) && q.unit == o.unit
	}
	return false
}

// Equivalent checks equivalence with another value.
// For quantities, this considers unit compatibility.
func (q Quantity) Equivalent(other Value) bool {
	if o, ok := other.(Quantity); ok {
		// Same value and compatible units
		if q.value.Equal(o.value) {
			// Empty units are compatible with anything
			if q.unit == "" || o.unit == "" {
				return true
			}
			return strings.EqualFold(q.unit, o.unit)
		}
	}
	return false
}

// String returns the string representation.
func (q Quantity) String() string {
	if q.unit == "" {
		return q.value.String()
	}
	// Use quotes if unit contains spaces
	if strings.Contains(q.unit, " ") {
		return fmt.Sprintf("%s '%s'", q.value.String(), q.unit)
	}
	return fmt.Sprintf("%s %s", q.value.String(), q.unit)
}

// IsEmpty returns false for Quantity.
func (q Quantity) IsEmpty() bool {
	return false
}

// Value returns the numeric value.
func (q Quantity) Value() decimal.Decimal {
	return q.value
}

// Unit returns the unit string.
func (q Quantity) Unit() string {
	return q.unit
}

// Compare compares two quantities.
// Returns -1, 0, or 1 if units are compatible, or error if not.
// Implements the Comparable interface.
func (q Quantity) Compare(other Value) (int, error) {
	otherQ, ok := other.(Quantity)
	if !ok {
		return 0, fmt.Errorf("cannot compare Quantity with %s", other.Type())
	}
	if q.unit != otherQ.unit && q.unit != "" && otherQ.unit != "" {
		return 0, fmt.Errorf("incompatible units: %s and %s", q.unit, otherQ.unit)
	}
	return q.value.Cmp(otherQ.value), nil
}

// Add adds two quantities.
func (q Quantity) Add(other Quantity) (Quantity, error) {
	if q.unit != other.unit && q.unit != "" && other.unit != "" {
		return Quantity{}, fmt.Errorf("incompatible units: %s and %s", q.unit, other.unit)
	}
	unit := q.unit
	if unit == "" {
		unit = other.unit
	}
	return Quantity{value: q.value.Add(other.value), unit: unit}, nil
}

// Subtract subtracts two quantities.
func (q Quantity) Subtract(other Quantity) (Quantity, error) {
	if q.unit != other.unit && q.unit != "" && other.unit != "" {
		return Quantity{}, fmt.Errorf("incompatible units: %s and %s", q.unit, other.unit)
	}
	unit := q.unit
	if unit == "" {
		unit = other.unit
	}
	return Quantity{value: q.value.Sub(other.value), unit: unit}, nil
}

// Multiply multiplies the quantity by a number.
func (q Quantity) Multiply(factor decimal.Decimal) Quantity {
	return Quantity{value: q.value.Mul(factor), unit: q.unit}
}

// Divide divides the quantity by a number.
func (q Quantity) Divide(divisor decimal.Decimal) (Quantity, error) {
	if divisor.IsZero() {
		return Quantity{}, fmt.Errorf("division by zero")
	}
	return Quantity{value: q.value.Div(divisor), unit: q.unit}, nil
}
