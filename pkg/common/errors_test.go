package common

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathError(t *testing.T) {
	t.Run("with path", func(t *testing.T) {
		err := &PathError{
			Path: "Patient.name[0].family",
			Err:  errors.New("invalid value"),
		}

		assert.Equal(t, "at Patient.name[0].family: invalid value", err.Error())
	})

	t.Run("empty path", func(t *testing.T) {
		err := &PathError{
			Path: "",
			Err:  errors.New("some error"),
		}

		assert.Equal(t, "some error", err.Error())
	})

	t.Run("unwrap", func(t *testing.T) {
		innerErr := errors.New("inner error")
		err := &PathError{
			Path: "Patient.id",
			Err:  innerErr,
		}

		assert.Equal(t, innerErr, err.Unwrap())
		assert.True(t, errors.Is(err, innerErr))
	})
}

func TestWrapPath(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		result := WrapPath("some.path", nil)
		assert.Nil(t, result)
	})

	t.Run("wraps error", func(t *testing.T) {
		innerErr := errors.New("something failed")
		result := WrapPath("Patient.birthDate", innerErr)

		assert.NotNil(t, result)
		assert.Contains(t, result.Error(), "Patient.birthDate")
		assert.Contains(t, result.Error(), "something failed")
	})
}

func TestWrapPathf(t *testing.T) {
	err := WrapPathf("Observation.value", "expected %s, got %s", "Quantity", "string")

	assert.Contains(t, err.Error(), "Observation.value")
	assert.Contains(t, err.Error(), "expected Quantity, got string")
}

func TestIsPathError(t *testing.T) {
	t.Run("is PathError", func(t *testing.T) {
		err := WrapPath("some.path", errors.New("error"))
		assert.True(t, IsPathError(err))
	})

	t.Run("wrapped PathError", func(t *testing.T) {
		inner := WrapPath("some.path", errors.New("error"))
		wrapped := fmt.Errorf("outer: %w", inner)
		assert.True(t, IsPathError(wrapped))
	})

	t.Run("not PathError", func(t *testing.T) {
		err := errors.New("plain error")
		assert.False(t, IsPathError(err))
	})

	t.Run("nil error", func(t *testing.T) {
		assert.False(t, IsPathError(nil))
	})
}

func TestGetPath(t *testing.T) {
	t.Run("from PathError", func(t *testing.T) {
		err := WrapPath("Patient.name", errors.New("error"))
		assert.Equal(t, "Patient.name", GetPath(err))
	})

	t.Run("from wrapped PathError", func(t *testing.T) {
		inner := WrapPath("Observation.code", errors.New("error"))
		wrapped := fmt.Errorf("outer: %w", inner)
		assert.Equal(t, "Observation.code", GetPath(wrapped))
	})

	t.Run("not PathError", func(t *testing.T) {
		err := errors.New("plain error")
		assert.Equal(t, "", GetPath(err))
	})

	t.Run("nil error", func(t *testing.T) {
		assert.Equal(t, "", GetPath(nil))
	})
}

func TestSentinelErrors(t *testing.T) {
	// Test that sentinel errors can be checked with errors.Is
	testCases := []struct {
		name string
		err  error
	}{
		{"ErrNilResource", ErrNilResource},
		{"ErrUnknownType", ErrUnknownType},
		{"ErrInvalidJSON", ErrInvalidJSON},
		{"ErrMarshalFailed", ErrMarshalFailed},
		{"ErrUnmarshalFailed", ErrUnmarshalFailed},
		{"ErrInvalidSpec", ErrInvalidSpec},
		{"ErrMissingRequired", ErrMissingRequired},
		{"ErrInvalidExpression", ErrInvalidExpression},
		{"ErrEvaluationFailed", ErrEvaluationFailed},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := fmt.Errorf("wrapped: %w", tc.err)
			assert.True(t, errors.Is(wrapped, tc.err))
		})
	}
}
