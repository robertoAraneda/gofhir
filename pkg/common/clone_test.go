package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testStruct simulates a FHIR resource structure
type testStruct struct {
	ID       *string           `json:"id,omitempty"`
	Active   *bool             `json:"active,omitempty"`
	Name     string            `json:"name"`
	Count    int               `json:"count"`
	Tags     []string          `json:"tags,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Nested   *nestedStruct     `json:"nested,omitempty"`
}

type nestedStruct struct {
	Value *string `json:"value,omitempty"`
}

func TestClone(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		var input *testStruct
		result := Clone(input)
		assert.Nil(t, result)
	})

	t.Run("simple struct", func(t *testing.T) {
		original := &testStruct{
			ID:     String("123"),
			Active: Bool(true),
			Name:   "test",
			Count:  42,
		}

		cloned := Clone(original)
		require.NotNil(t, cloned)

		// Values should be equal
		assert.Equal(t, *original.ID, *cloned.ID)
		assert.Equal(t, *original.Active, *cloned.Active)
		assert.Equal(t, original.Name, cloned.Name)
		assert.Equal(t, original.Count, cloned.Count)

		// But pointers should be different (deep copy)
		assert.NotSame(t, original.ID, cloned.ID)
		assert.NotSame(t, original.Active, cloned.Active)

		// Modifying clone should not affect original
		*cloned.ID = "456"
		assert.Equal(t, "123", *original.ID)
	})

	t.Run("struct with slice", func(t *testing.T) {
		original := &testStruct{
			Tags: []string{"a", "b", "c"},
		}

		cloned := Clone(original)
		require.NotNil(t, cloned)

		assert.Equal(t, original.Tags, cloned.Tags)

		// Modifying clone's slice should not affect original
		cloned.Tags[0] = "modified"
		assert.Equal(t, "a", original.Tags[0])
	})

	t.Run("struct with map", func(t *testing.T) {
		original := &testStruct{
			Metadata: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		}

		cloned := Clone(original)
		require.NotNil(t, cloned)

		assert.Equal(t, original.Metadata, cloned.Metadata)

		// Modifying clone's map should not affect original
		cloned.Metadata["key1"] = "modified"
		assert.Equal(t, "value1", original.Metadata["key1"])
	})

	t.Run("struct with nested struct", func(t *testing.T) {
		original := &testStruct{
			Nested: &nestedStruct{
				Value: String("nested-value"),
			},
		}

		cloned := Clone(original)
		require.NotNil(t, cloned)
		require.NotNil(t, cloned.Nested)

		assert.Equal(t, *original.Nested.Value, *cloned.Nested.Value)
		assert.NotSame(t, original.Nested, cloned.Nested)
		assert.NotSame(t, original.Nested.Value, cloned.Nested.Value)

		// Modifying clone should not affect original
		*cloned.Nested.Value = "modified"
		assert.Equal(t, "nested-value", *original.Nested.Value)
	})
}

func TestCloneSlice(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		var input []string
		result := CloneSlice(input)
		assert.Nil(t, result)
	})

	t.Run("empty slice", func(t *testing.T) {
		input := []string{}
		result := CloneSlice(input)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("string slice", func(t *testing.T) {
		original := []string{"a", "b", "c"}
		cloned := CloneSlice(original)

		assert.Equal(t, original, cloned)

		// Modifying clone should not affect original
		cloned[0] = "modified"
		assert.Equal(t, "a", original[0])
	})

	t.Run("struct slice", func(t *testing.T) {
		original := []testStruct{
			{ID: String("1"), Name: "first"},
			{ID: String("2"), Name: "second"},
		}

		cloned := CloneSlice(original)
		require.Len(t, cloned, 2)

		// Modifying clone should not affect original
		*cloned[0].ID = "modified"
		assert.Equal(t, "1", *original[0].ID)
	})
}

func TestCloneMap(t *testing.T) {
	t.Run("nil map", func(t *testing.T) {
		var input map[string]string
		result := CloneMap(input)
		assert.Nil(t, result)
	})

	t.Run("empty map", func(t *testing.T) {
		input := map[string]string{}
		result := CloneMap(input)
		assert.NotNil(t, result)
		assert.Empty(t, result)
	})

	t.Run("string map", func(t *testing.T) {
		original := map[string]string{
			"key1": "value1",
			"key2": "value2",
		}

		cloned := CloneMap(original)
		assert.Equal(t, original, cloned)

		// Modifying clone should not affect original
		cloned["key1"] = "modified"
		assert.Equal(t, "value1", original["key1"])
	})

	t.Run("map with struct values", func(t *testing.T) {
		original := map[string]testStruct{
			"item1": {ID: String("1"), Name: "first"},
		}

		cloned := CloneMap(original)
		require.Contains(t, cloned, "item1")

		// Values are copied, so modifying requires getting the value first
		item := cloned["item1"]
		item.Name = "modified"
		cloned["item1"] = item

		assert.Equal(t, "first", original["item1"].Name)
	})
}
