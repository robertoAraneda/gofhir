package types

import (
	"fmt"
	"strings"
)

// Collection is an ordered sequence of FHIRPath values.
// It is the fundamental return type for all FHIRPath expressions.
type Collection []Value

// Empty returns true if the collection has no elements.
func (c Collection) Empty() bool {
	return len(c) == 0
}

// Count returns the number of elements in the collection.
func (c Collection) Count() int {
	return len(c)
}

// First returns the first element and true, or nil and false if empty.
func (c Collection) First() (Value, bool) {
	if len(c) == 0 {
		return nil, false
	}
	return c[0], true
}

// Last returns the last element and true, or nil and false if empty.
func (c Collection) Last() (Value, bool) {
	if len(c) == 0 {
		return nil, false
	}
	return c[len(c)-1], true
}

// Single returns the single element if the collection has exactly one element.
// Returns an error if empty or has more than one element.
func (c Collection) Single() (Value, error) {
	switch len(c) {
	case 0:
		return nil, fmt.Errorf("expected single value, got empty collection")
	case 1:
		return c[0], nil
	default:
		return nil, fmt.Errorf("expected single value, got %d elements", len(c))
	}
}

// Tail returns all elements except the first.
func (c Collection) Tail() Collection {
	if len(c) <= 1 {
		return Collection{}
	}
	return c[1:]
}

// Skip returns a collection with the first n elements removed.
func (c Collection) Skip(n int) Collection {
	if n >= len(c) {
		return Collection{}
	}
	if n <= 0 {
		return c
	}
	return c[n:]
}

// Take returns a collection with only the first n elements.
func (c Collection) Take(n int) Collection {
	if n <= 0 {
		return Collection{}
	}
	if n >= len(c) {
		return c
	}
	return c[:n]
}

// Contains returns true if the collection contains a value equal to v.
func (c Collection) Contains(v Value) bool {
	for _, item := range c {
		if item.Equal(v) {
			return true
		}
	}
	return false
}

// Distinct returns a new collection with duplicate values removed.
// Preserves the order of first occurrence.
func (c Collection) Distinct() Collection {
	if len(c) <= 1 {
		return c
	}
	result := make(Collection, 0, len(c))
	for _, item := range c {
		if !result.Contains(item) {
			result = append(result, item)
		}
	}
	return result
}

// IsDistinct returns true if all elements in the collection are unique.
func (c Collection) IsDistinct() bool {
	return len(c) == len(c.Distinct())
}

// Union returns a new collection that is the union of c and other.
// Duplicates are removed.
func (c Collection) Union(other Collection) Collection {
	result := make(Collection, 0, len(c)+len(other))
	result = append(result, c...)
	for _, item := range other {
		if !result.Contains(item) {
			result = append(result, item)
		}
	}
	return result
}

// Combine returns a new collection that combines c and other.
// Unlike Union, duplicates are preserved.
func (c Collection) Combine(other Collection) Collection {
	result := make(Collection, 0, len(c)+len(other))
	result = append(result, c...)
	result = append(result, other...)
	return result
}

// Intersect returns elements that are in both collections.
func (c Collection) Intersect(other Collection) Collection {
	result := make(Collection, 0)
	for _, item := range c {
		if other.Contains(item) && !result.Contains(item) {
			result = append(result, item)
		}
	}
	return result
}

// Exclude returns elements in c that are not in other.
func (c Collection) Exclude(other Collection) Collection {
	result := make(Collection, 0)
	for _, item := range c {
		if !other.Contains(item) {
			result = append(result, item)
		}
	}
	return result
}

// String returns a string representation of the collection.
func (c Collection) String() string {
	if len(c) == 0 {
		return "[]"
	}
	parts := make([]string, len(c))
	for i, v := range c {
		parts[i] = v.String()
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// ToBoolean converts singleton collection to boolean.
// Returns error if not a singleton or not a boolean value.
func (c Collection) ToBoolean() (bool, error) {
	if len(c) == 0 {
		return false, fmt.Errorf("cannot convert empty collection to boolean")
	}
	if len(c) > 1 {
		return false, fmt.Errorf("cannot convert collection with %d elements to boolean", len(c))
	}
	if b, ok := c[0].(Boolean); ok {
		return b.Bool(), nil
	}
	return false, fmt.Errorf("cannot convert %s to boolean", c[0].Type())
}

// AllTrue returns true if all items are boolean true.
func (c Collection) AllTrue() bool {
	for _, item := range c {
		if b, ok := item.(Boolean); !ok || !b.Bool() {
			return false
		}
	}
	return true
}

// AnyTrue returns true if any item is boolean true.
func (c Collection) AnyTrue() bool {
	for _, item := range c {
		if b, ok := item.(Boolean); ok && b.Bool() {
			return true
		}
	}
	return false
}

// AllFalse returns true if all items are boolean false.
func (c Collection) AllFalse() bool {
	for _, item := range c {
		if b, ok := item.(Boolean); !ok || b.Bool() {
			return false
		}
	}
	return true
}

// AnyFalse returns true if any item is boolean false.
func (c Collection) AnyFalse() bool {
	for _, item := range c {
		if b, ok := item.(Boolean); ok && !b.Bool() {
			return true
		}
	}
	return false
}
