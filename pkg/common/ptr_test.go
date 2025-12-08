package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	s := String("hello")
	assert.NotNil(t, s)
	assert.Equal(t, "hello", *s)
}

func TestStringVal(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{"nil", nil, ""},
		{"empty", String(""), ""},
		{"value", String("hello"), "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, StringVal(tt.input))
		})
	}
}

func TestBool(t *testing.T) {
	b := Bool(true)
	assert.NotNil(t, b)
	assert.True(t, *b)

	b = Bool(false)
	assert.NotNil(t, b)
	assert.False(t, *b)
}

func TestBoolVal(t *testing.T) {
	tests := []struct {
		name     string
		input    *bool
		expected bool
	}{
		{"nil", nil, false},
		{"true", Bool(true), true},
		{"false", Bool(false), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, BoolVal(tt.input))
		})
	}
}

func TestInt(t *testing.T) {
	i := Int(42)
	assert.NotNil(t, i)
	assert.Equal(t, 42, *i)

	i = Int(-1)
	assert.NotNil(t, i)
	assert.Equal(t, -1, *i)
}

func TestIntVal(t *testing.T) {
	tests := []struct {
		name     string
		input    *int
		expected int
	}{
		{"nil", nil, 0},
		{"zero", Int(0), 0},
		{"positive", Int(42), 42},
		{"negative", Int(-1), -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IntVal(tt.input))
		})
	}
}

func TestInt64(t *testing.T) {
	i := Int64(9223372036854775807)
	assert.NotNil(t, i)
	assert.Equal(t, int64(9223372036854775807), *i)
}

func TestInt64Val(t *testing.T) {
	tests := []struct {
		name     string
		input    *int64
		expected int64
	}{
		{"nil", nil, 0},
		{"value", Int64(123456789), 123456789},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Int64Val(tt.input))
		})
	}
}

func TestUint32(t *testing.T) {
	i := Uint32(4294967295)
	assert.NotNil(t, i)
	assert.Equal(t, uint32(4294967295), *i)
}

func TestUint32Val(t *testing.T) {
	tests := []struct {
		name     string
		input    *uint32
		expected uint32
	}{
		{"nil", nil, 0},
		{"value", Uint32(123), 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Uint32Val(tt.input))
		})
	}
}

func TestFloat64(t *testing.T) {
	f := Float64(3.14159)
	assert.NotNil(t, f)
	assert.Equal(t, 3.14159, *f)
}

func TestFloat64Val(t *testing.T) {
	tests := []struct {
		name     string
		input    *float64
		expected float64
	}{
		{"nil", nil, 0},
		{"zero", Float64(0), 0},
		{"positive", Float64(3.14), 3.14},
		{"negative", Float64(-2.5), -2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Float64Val(tt.input))
		})
	}
}
