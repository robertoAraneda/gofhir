// Package eval provides the FHIRPath expression evaluator.
package eval

import "fmt"

// ErrorType represents the category of evaluation error.
type ErrorType int

const (
	// ErrParse indicates a parsing error.
	ErrParse ErrorType = iota
	// ErrType indicates a type mismatch error.
	ErrType
	// ErrSingletonExpected indicates multiple values where one was expected.
	ErrSingletonExpected
	// ErrFunctionNotFound indicates an unknown function.
	ErrFunctionNotFound
	// ErrInvalidArguments indicates invalid function arguments.
	ErrInvalidArguments
	// ErrDivisionByZero indicates division by zero.
	ErrDivisionByZero
	// ErrInvalidPath indicates an invalid path expression.
	ErrInvalidPath
	// ErrTimeout indicates evaluation timeout.
	ErrTimeout
	// ErrInvalidOperation indicates an unsupported operation.
	ErrInvalidOperation
	// ErrInvalidExpression indicates an invalid expression.
	ErrInvalidExpression
)

// String returns the string representation of the error type.
func (t ErrorType) String() string {
	switch t {
	case ErrParse:
		return "ParseError"
	case ErrType:
		return "TypeError"
	case ErrSingletonExpected:
		return "SingletonExpectedError"
	case ErrFunctionNotFound:
		return "FunctionNotFoundError"
	case ErrInvalidArguments:
		return "InvalidArgumentsError"
	case ErrDivisionByZero:
		return "DivisionByZeroError"
	case ErrInvalidPath:
		return "InvalidPathError"
	case ErrTimeout:
		return "TimeoutError"
	case ErrInvalidOperation:
		return "InvalidOperationError"
	case ErrInvalidExpression:
		return "InvalidExpressionError"
	default:
		return "UnknownError"
	}
}

// Position represents a location in the source expression.
type Position struct {
	Line   int
	Column int
}

// EvalError represents an error that occurred during evaluation.
//
//nolint:revive // Keeping EvalError name for API compatibility
type EvalError struct {
	Type       ErrorType
	Message    string
	Path       string   // Expression path where error occurred
	Position   Position // Position in source expression
	Underlying error    // Original error if wrapping
}

// Error implements the error interface.
func (e *EvalError) Error() string {
	if e.Position.Line > 0 {
		return fmt.Sprintf("%s at %d:%d: %s", e.Type, e.Position.Line, e.Position.Column, e.Message)
	}
	if e.Path != "" {
		return fmt.Sprintf("%s in '%s': %s", e.Type, e.Path, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap returns the underlying error.
func (e *EvalError) Unwrap() error {
	return e.Underlying
}

// NewEvalError creates a new evaluation error.
// Supports format strings like fmt.Sprintf.
func NewEvalError(errType ErrorType, format string, args ...interface{}) *EvalError {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args...)
	}
	return &EvalError{
		Type:    errType,
		Message: message,
	}
}

// WithPath adds path information to the error.
func (e *EvalError) WithPath(path string) *EvalError {
	e.Path = path
	return e
}

// WithPosition adds position information to the error.
func (e *EvalError) WithPosition(line, column int) *EvalError {
	e.Position = Position{Line: line, Column: column}
	return e
}

// WithUnderlying adds an underlying error.
func (e *EvalError) WithUnderlying(err error) *EvalError {
	e.Underlying = err
	return e
}

// Helper functions for common errors

// ParseError creates a parsing error.
func ParseError(message string) *EvalError {
	return NewEvalError(ErrParse, message)
}

// TypeError creates a type mismatch error.
func TypeError(expected, actual, operation string) *EvalError {
	return NewEvalError(ErrType, fmt.Sprintf("expected %s, got %s in %s", expected, actual, operation))
}

// SingletonError creates a singleton expected error.
func SingletonError(count int) *EvalError {
	return NewEvalError(ErrSingletonExpected, fmt.Sprintf("expected single value, got %d elements", count))
}

// FunctionNotFoundError creates a function not found error.
func FunctionNotFoundError(name string) *EvalError {
	return NewEvalError(ErrFunctionNotFound, fmt.Sprintf("unknown function '%s'", name))
}

// InvalidArgumentsError creates an invalid arguments error.
func InvalidArgumentsError(funcName string, expected, actual int) *EvalError {
	return NewEvalError(ErrInvalidArguments, fmt.Sprintf("function '%s' expects %d arguments, got %d", funcName, expected, actual))
}

// DivisionByZeroError creates a division by zero error.
func DivisionByZeroError() *EvalError {
	return NewEvalError(ErrDivisionByZero, "division by zero")
}

// InvalidPathError creates an invalid path error.
func InvalidPathError(path string) *EvalError {
	return NewEvalError(ErrInvalidPath, fmt.Sprintf("invalid path '%s'", path))
}

// InvalidOperationError creates an invalid operation error.
func InvalidOperationError(op, leftType, rightType string) *EvalError {
	return NewEvalError(ErrInvalidOperation, fmt.Sprintf("cannot apply '%s' to %s and %s", op, leftType, rightType))
}
