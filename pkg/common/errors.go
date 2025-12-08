package common

import (
	"errors"
	"fmt"
)

// PathError wraps an error with path context.
// Used internally to add location information when errors occur during
// parsing, serialization, or other internal operations.
//
// Note: This is NOT for FHIR validation errors. Validation errors are
// reported as OperationOutcome resources via the validator package.
type PathError struct {
	Path string
	Err  error
}

// Error implements the error interface.
func (e *PathError) Error() string {
	if e.Path == "" {
		return e.Err.Error()
	}
	return fmt.Sprintf("at %s: %v", e.Path, e.Err)
}

// Unwrap returns the wrapped error for errors.Is/As support.
func (e *PathError) Unwrap() error {
	return e.Err
}

// WrapPath wraps an error with path context.
// Returns nil if err is nil.
func WrapPath(path string, err error) error {
	if err == nil {
		return nil
	}
	return &PathError{Path: path, Err: err}
}

// WrapPathf wraps an error with path context and a formatted message.
func WrapPathf(path string, format string, args ...any) error {
	return &PathError{Path: path, Err: fmt.Errorf(format, args...)}
}

// Sentinel errors for common internal error conditions.
// These are programming/system errors, not FHIR validation errors.
var (
	// Resource handling
	ErrNilResource = errors.New("resource is nil")
	ErrUnknownType = errors.New("unknown resource type")

	// JSON/Serialization
	ErrInvalidJSON    = errors.New("invalid JSON")
	ErrMarshalFailed  = errors.New("marshal failed")
	ErrUnmarshalFailed = errors.New("unmarshal failed")

	// Code generation
	ErrInvalidSpec     = errors.New("invalid specification")
	ErrMissingRequired = errors.New("missing required field in spec")

	// FHIRPath
	ErrInvalidExpression = errors.New("invalid FHIRPath expression")
	ErrEvaluationFailed  = errors.New("FHIRPath evaluation failed")
)

// IsPathError checks if an error is or wraps a PathError.
func IsPathError(err error) bool {
	var pathErr *PathError
	return errors.As(err, &pathErr)
}

// GetPath extracts the path from a PathError, or returns empty string.
func GetPath(err error) string {
	var pathErr *PathError
	if errors.As(err, &pathErr) {
		return pathErr.Path
	}
	return ""
}
