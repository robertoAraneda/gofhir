package types

import "fmt"

// TypeError represents a type mismatch error.
type TypeError struct {
	Expected  string
	Actual    string
	Operation string
}

// NewTypeError creates a new TypeError.
func NewTypeError(expected, actual, operation string) *TypeError {
	return &TypeError{
		Expected:  expected,
		Actual:    actual,
		Operation: operation,
	}
}

// Error implements the error interface.
func (e *TypeError) Error() string {
	return fmt.Sprintf("type error in %s: expected %s, got %s", e.Operation, e.Expected, e.Actual)
}
