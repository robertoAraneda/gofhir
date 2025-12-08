// Package types defines the FHIRPath type system.
package types

// Value is the base interface for all FHIRPath values.
type Value interface {
	// Type returns the FHIRPath type name.
	Type() string

	// Equal compares exact equality (= operator).
	Equal(other Value) bool

	// Equivalent compares equivalence (~ operator).
	// For strings: case-insensitive, ignores leading/trailing whitespace.
	Equivalent(other Value) bool

	// String returns a string representation of the value.
	String() string

	// IsEmpty indicates if this value represents empty.
	IsEmpty() bool
}

// Comparable is implemented by types that support ordering.
type Comparable interface {
	Value
	// Compare returns -1 if less than, 0 if equal, 1 if greater than.
	// Returns error if types are incompatible.
	Compare(other Value) (int, error)
}

// Numeric is implemented by numeric types (Integer, Decimal).
type Numeric interface {
	Value
	// ToDecimal converts the numeric to a Decimal.
	ToDecimal() Decimal
}
