package common

import "encoding/json"

// Clone creates a deep copy of any value using JSON marshaling/unmarshaling.
// This is a simple and reliable way to deep copy complex structs with nested pointers.
//
// Usage:
//
//	patient2 := common.Clone(patient)
//	patient2.ID = common.String("new-id") // doesn't affect original
func Clone[T any](v *T) *T {
	if v == nil {
		return nil
	}
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var clone T
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil
	}
	return &clone
}

// CloneSlice creates a deep copy of a slice of values.
func CloneSlice[T any](slice []T) []T {
	if slice == nil {
		return nil
	}
	if len(slice) == 0 {
		return []T{}
	}
	data, err := json.Marshal(slice)
	if err != nil {
		return nil
	}
	var clone []T
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil
	}
	return clone
}

// CloneMap creates a deep copy of a map.
func CloneMap[K comparable, V any](m map[K]V) map[K]V {
	if m == nil {
		return nil
	}
	if len(m) == 0 {
		return make(map[K]V)
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	var clone map[K]V
	if err := json.Unmarshal(data, &clone); err != nil {
		return nil
	}
	return clone
}
