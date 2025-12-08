package common

// Pointer helpers for creating pointers to primitive values.
// These are useful when constructing FHIR resources where most fields are optional pointers.

// String returns a pointer to the given string value.
func String(s string) *string {
	return &s
}

// StringVal returns the value of a string pointer, or empty string if nil.
func StringVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Bool returns a pointer to the given bool value.
func Bool(b bool) *bool {
	return &b
}

// BoolVal returns the value of a bool pointer, or false if nil.
func BoolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

// Int returns a pointer to the given int value.
func Int(i int) *int {
	return &i
}

// IntVal returns the value of an int pointer, or 0 if nil.
func IntVal(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

// Int64 returns a pointer to the given int64 value.
func Int64(i int64) *int64 {
	return &i
}

// Int64Val returns the value of an int64 pointer, or 0 if nil.
func Int64Val(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

// Uint32 returns a pointer to the given uint32 value.
func Uint32(i uint32) *uint32 {
	return &i
}

// Uint32Val returns the value of a uint32 pointer, or 0 if nil.
func Uint32Val(i *uint32) uint32 {
	if i == nil {
		return 0
	}
	return *i
}

// Float64 returns a pointer to the given float64 value.
func Float64(f float64) *float64 {
	return &f
}

// Float64Val returns the value of a float64 pointer, or 0 if nil.
func Float64Val(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}
