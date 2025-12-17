package types

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/robertoaraneda/gofhir/pkg/ucum"
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
// For quantities with different units, uses UCUM normalization per FHIRPath spec.
func (q Quantity) Equal(other Value) bool {
	o, ok := other.(Quantity)
	if !ok {
		return false
	}

	// Same unit - compare values directly
	if q.unit == o.unit {
		return q.value.Equal(o.value)
	}

	// Empty units - compare values directly
	if q.unit == "" || o.unit == "" {
		return q.value.Equal(o.value)
	}

	// Different units - use UCUM normalization
	norm1 := q.Normalize()
	norm2 := o.Normalize()

	// Must have same canonical unit
	if norm1.Code != norm2.Code {
		return false
	}

	// Compare normalized values
	val1 := decimal.NewFromFloat(norm1.Value)
	val2 := decimal.NewFromFloat(norm2.Value)
	return val1.Equal(val2)
}

// Equivalent checks equivalence with another value.
// For quantities, this uses UCUM normalization to compare values with different units.
// Per FHIRPath spec: quantities are equivalent if their canonical normalized forms are equal.
func (q Quantity) Equivalent(other Value) bool {
	o, ok := other.(Quantity)
	if !ok {
		return false
	}

	// Empty units are compatible with anything - compare values directly
	if q.unit == "" || o.unit == "" {
		return q.value.Equal(o.value)
	}

	// Same unit - compare values directly
	if strings.EqualFold(q.unit, o.unit) {
		return q.value.Equal(o.value)
	}

	// Different units - try UCUM normalization
	norm1 := q.Normalize()
	norm2 := o.Normalize()

	// Must have same canonical unit
	if norm1.Code != norm2.Code {
		return false
	}

	// Compare normalized values with tolerance for floating point
	diff := norm1.Value - norm2.Value
	if diff < 0 {
		diff = -diff
	}
	// Use relative tolerance for comparison
	maxVal := norm1.Value
	if norm2.Value > maxVal {
		maxVal = norm2.Value
	}
	if maxVal == 0 {
		return diff == 0
	}
	return diff/maxVal < 1e-10
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
// Uses UCUM normalization to compare quantities with different but compatible units.
// Implements the Comparable interface.
func (q Quantity) Compare(other Value) (int, error) {
	otherQ, ok := other.(Quantity)
	if !ok {
		return 0, fmt.Errorf("cannot compare Quantity with %s", other.Type())
	}

	// If units are the same (or one is empty), compare directly
	if q.unit == otherQ.unit || q.unit == "" || otherQ.unit == "" {
		return q.value.Cmp(otherQ.value), nil
	}

	// Try UCUM normalization for different units
	norm1 := q.Normalize()
	norm2 := otherQ.Normalize()

	// Check if units are compatible after normalization
	if norm1.Code != norm2.Code {
		return 0, fmt.Errorf("incompatible units: %s and %s", q.unit, otherQ.unit)
	}

	// Compare normalized values
	val1 := decimal.NewFromFloat(norm1.Value)
	val2 := decimal.NewFromFloat(norm2.Value)
	return val1.Cmp(val2), nil
}

// Normalize returns the UCUM-normalized form of this quantity.
func (q Quantity) Normalize() ucum.NormalizedQuantity {
	val, _ := q.value.Float64()
	return ucum.Normalize(val, q.unit)
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
