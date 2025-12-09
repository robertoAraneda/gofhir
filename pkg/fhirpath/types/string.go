package types

import (
	"strings"
	"unicode"
)

// String represents a FHIRPath string value.
type String struct {
	value string
}

// NewString creates a new String value.
func NewString(v string) String {
	return String{value: v}
}

// Value returns the underlying string value.
func (s String) Value() string {
	return s.value
}

// Type returns "String".
func (s String) Type() string {
	return "String"
}

// Equal returns true if other is a String with the same value.
func (s String) Equal(other Value) bool {
	if o, ok := other.(String); ok {
		return s.value == o.value
	}
	return false
}

// Equivalent compares strings case-insensitively with normalized whitespace.
func (s String) Equivalent(other Value) bool {
	if o, ok := other.(String); ok {
		return normalizeString(s.value) == normalizeString(o.value)
	}
	return false
}

// normalizeString converts to lowercase and normalizes whitespace.
func normalizeString(s string) string {
	// Trim leading/trailing whitespace
	s = strings.TrimSpace(s)
	// Convert to lowercase
	s = strings.ToLower(s)
	// Normalize internal whitespace to single spaces
	var result strings.Builder
	prevSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !prevSpace {
				result.WriteRune(' ')
				prevSpace = true
			}
		} else {
			result.WriteRune(r)
			prevSpace = false
		}
	}
	return result.String()
}

// String returns the string value.
func (s String) String() string {
	return s.value
}

// IsEmpty returns true if the string is empty.
func (s String) IsEmpty() bool {
	return s.value == ""
}

// Length returns the number of characters.
func (s String) Length() int {
	return len([]rune(s.value))
}

// Contains returns true if the string contains the substring.
func (s String) Contains(substr string) bool {
	return strings.Contains(s.value, substr)
}

// StartsWith returns true if the string starts with the prefix.
func (s String) StartsWith(prefix string) bool {
	return strings.HasPrefix(s.value, prefix)
}

// EndsWith returns true if the string ends with the suffix.
func (s String) EndsWith(suffix string) bool {
	return strings.HasSuffix(s.value, suffix)
}

// Upper returns a new String with all characters uppercase.
func (s String) Upper() String {
	return NewString(strings.ToUpper(s.value))
}

// Lower returns a new String with all characters lowercase.
func (s String) Lower() String {
	return NewString(strings.ToLower(s.value))
}

// Compare compares two strings lexicographically.
func (s String) Compare(other Value) (int, error) {
	if o, ok := other.(String); ok {
		return strings.Compare(s.value, o.value), nil
	}
	return 0, NewTypeError("String", other.Type(), "comparison")
}

// IndexOf returns the index of the first occurrence of substr, or -1.
func (s String) IndexOf(substr string) int {
	return strings.Index(s.value, substr)
}

// Substring returns a substring starting at start with the given length.
func (s String) Substring(start, length int) String {
	runes := []rune(s.value)
	if start < 0 || start >= len(runes) {
		return NewString("")
	}
	end := start + length
	if end > len(runes) {
		end = len(runes)
	}
	return NewString(string(runes[start:end]))
}

// Replace returns a new String with all occurrences of old replaced by replacement.
func (s String) Replace(old, replacement string) String {
	return NewString(strings.ReplaceAll(s.value, old, replacement))
}

// ToChars returns a collection of single-character strings.
func (s String) ToChars() Collection {
	runes := []rune(s.value)
	result := make(Collection, len(runes))
	for i, r := range runes {
		result[i] = NewString(string(r))
	}
	return result
}
