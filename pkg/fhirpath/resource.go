package fhirpath

import (
	"encoding/json"
	"fmt"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

// Collection is an alias for types.Collection for easier external use.
type Collection = types.Collection

// Value is an alias for types.Value for easier external use.
type Value = types.Value

// Resource represents any FHIR resource that can be evaluated.
type Resource interface {
	GetResourceType() string
}

// EvaluateResource evaluates a FHIRPath expression against a Go struct.
// The resource is serialized to JSON first, then evaluated.
// For better performance with multiple evaluations, cache the JSON bytes.
func EvaluateResource(resource Resource, expr string) (Collection, error) {
	jsonBytes, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource: %w", err)
	}
	return Evaluate(jsonBytes, expr)
}

// EvaluateResourceCached is like EvaluateResource but uses the expression cache.
func EvaluateResourceCached(resource Resource, expr string) (Collection, error) {
	jsonBytes, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource: %w", err)
	}
	return EvaluateCached(jsonBytes, expr)
}

// ResourceJSON wraps a resource with its pre-serialized JSON for efficient repeated evaluation.
type ResourceJSON struct {
	resource Resource
	json     []byte
}

// NewResourceJSON creates a ResourceJSON from a Go resource.
func NewResourceJSON(resource Resource) (*ResourceJSON, error) {
	jsonBytes, err := json.Marshal(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource: %w", err)
	}
	return &ResourceJSON{
		resource: resource,
		json:     jsonBytes,
	}, nil
}

// MustNewResourceJSON is like NewResourceJSON but panics on error.
func MustNewResourceJSON(resource Resource) *ResourceJSON {
	rj, err := NewResourceJSON(resource)
	if err != nil {
		panic(err)
	}
	return rj
}

// Evaluate evaluates a FHIRPath expression against this resource.
func (r *ResourceJSON) Evaluate(expr string) (Collection, error) {
	return Evaluate(r.json, expr)
}

// EvaluateCached evaluates using the expression cache.
func (r *ResourceJSON) EvaluateCached(expr string) (Collection, error) {
	return EvaluateCached(r.json, expr)
}

// JSON returns the pre-serialized JSON bytes.
func (r *ResourceJSON) JSON() []byte {
	return r.json
}

// Resource returns the original Go resource.
func (r *ResourceJSON) Resource() Resource {
	return r.resource
}

// EvaluateToBoolean evaluates an expression and returns a boolean result.
// Returns false if the result is empty or not a boolean.
func EvaluateToBoolean(resource []byte, expr string) (bool, error) {
	result, err := EvaluateCached(resource, expr)
	if err != nil {
		return false, err
	}
	if result.Empty() {
		return false, nil
	}
	if len(result) != 1 {
		return false, fmt.Errorf("expected single value, got %d", len(result))
	}
	if b, ok := result[0].(types.Boolean); ok {
		return b.Bool(), nil
	}
	return false, fmt.Errorf("expected Boolean, got %s", result[0].Type())
}

// EvaluateToString evaluates an expression and returns a string result.
func EvaluateToString(resource []byte, expr string) (string, error) {
	result, err := EvaluateCached(resource, expr)
	if err != nil {
		return "", err
	}
	if result.Empty() {
		return "", nil
	}
	if len(result) != 1 {
		return "", fmt.Errorf("expected single value, got %d", len(result))
	}
	if s, ok := result[0].(types.String); ok {
		return s.Value(), nil
	}
	// Try to convert to string
	return result[0].String(), nil
}

// EvaluateToStrings evaluates an expression and returns all results as strings.
func EvaluateToStrings(resource []byte, expr string) ([]string, error) {
	result, err := EvaluateCached(resource, expr)
	if err != nil {
		return nil, err
	}
	strings := make([]string, len(result))
	for i, v := range result {
		if s, ok := v.(types.String); ok {
			strings[i] = s.Value()
		} else {
			strings[i] = v.String()
		}
	}
	return strings, nil
}

// Exists evaluates an expression and returns true if any results exist.
func Exists(resource []byte, expr string) (bool, error) {
	result, err := EvaluateCached(resource, expr)
	if err != nil {
		return false, err
	}
	return !result.Empty(), nil
}

// Count evaluates an expression and returns the number of results.
func Count(resource []byte, expr string) (int, error) {
	result, err := EvaluateCached(resource, expr)
	if err != nil {
		return 0, err
	}
	return len(result), nil
}
