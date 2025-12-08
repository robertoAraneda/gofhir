package types

import (
	"encoding/json"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/shopspring/decimal"
)

// ObjectValue represents a FHIR resource or complex type as a JSON object.
type ObjectValue struct {
	data   []byte
	fields map[string]Value // Cache of accessed fields
}

// NewObjectValue creates a new ObjectValue from JSON bytes.
func NewObjectValue(data []byte) *ObjectValue {
	return &ObjectValue{
		data:   data,
		fields: make(map[string]Value),
	}
}

// Type returns "Object" or the resourceType if available.
func (o *ObjectValue) Type() string {
	if rt, err := jsonparser.GetString(o.data, "resourceType"); err == nil {
		return rt
	}
	return "Object"
}

// Equal returns true if the JSON data is identical.
func (o *ObjectValue) Equal(other Value) bool {
	if ov, ok := other.(*ObjectValue); ok {
		return string(o.data) == string(ov.data)
	}
	return false
}

// Equivalent is the same as Equal for objects.
func (o *ObjectValue) Equivalent(other Value) bool {
	return o.Equal(other)
}

// String returns the JSON representation.
func (o *ObjectValue) String() string {
	return string(o.data)
}

// IsEmpty returns false for object values.
func (o *ObjectValue) IsEmpty() bool {
	return false
}

// Data returns the raw JSON data.
func (o *ObjectValue) Data() []byte {
	return o.data
}

// Get retrieves a field value, caching the result.
func (o *ObjectValue) Get(field string) (Value, bool) {
	// Check cache first
	if v, ok := o.fields[field]; ok {
		return v, true
	}

	// Parse from JSON
	value, dataType, _, err := jsonparser.Get(o.data, field)
	if err != nil {
		return nil, false
	}

	// Convert to Value and cache
	v := jsonValueToFHIRValue(value, dataType)
	o.fields[field] = v

	return v, true
}

// GetCollection retrieves a field as a Collection.
// If the field is an array, returns all elements.
// If the field is a single value, returns a singleton collection.
func (o *ObjectValue) GetCollection(field string) Collection {
	value, dataType, _, err := jsonparser.Get(o.data, field)
	if err != nil {
		return Collection{}
	}

	if dataType == jsonparser.Array {
		return jsonArrayToCollection(value)
	}

	v := jsonValueToFHIRValue(value, dataType)
	if v == nil {
		return Collection{}
	}
	return Collection{v}
}

// Keys returns all field names in the object.
func (o *ObjectValue) Keys() []string {
	var keys []string
	jsonparser.ObjectEach(o.data, func(key []byte, _ []byte, _ jsonparser.ValueType, _ int) error {
		keys = append(keys, string(key))
		return nil
	})
	return keys
}

// Children returns a collection of all child values.
func (o *ObjectValue) Children() Collection {
	var result Collection
	jsonparser.ObjectEach(o.data, func(_ []byte, value []byte, dataType jsonparser.ValueType, _ int) error {
		if dataType == jsonparser.Array {
			result = append(result, jsonArrayToCollection(value)...)
		} else {
			v := jsonValueToFHIRValue(value, dataType)
			if v != nil {
				result = append(result, v)
			}
		}
		return nil
	})
	return result
}

// jsonValueToFHIRValue converts a JSON value to a FHIRPath Value.
func jsonValueToFHIRValue(data []byte, dataType jsonparser.ValueType) Value {
	switch dataType {
	case jsonparser.String:
		// Remove quotes and unescape
		var s string
		if err := json.Unmarshal(append([]byte{'"'}, append(data, '"')...), &s); err != nil {
			s = string(data)
		}
		return NewString(s)

	case jsonparser.Number:
		s := string(data)
		// Check if it's an integer
		if !strings.Contains(s, ".") && !strings.Contains(s, "e") && !strings.Contains(s, "E") {
			var i int64
			if _, err := jsonparser.ParseInt(data); err == nil {
				i, _ = jsonparser.ParseInt(data)
				return NewInteger(i)
			}
		}
		// Parse as decimal
		d, err := NewDecimal(s)
		if err != nil {
			return nil
		}
		return d

	case jsonparser.Boolean:
		b, _ := jsonparser.ParseBoolean(data)
		return NewBoolean(b)

	case jsonparser.Object:
		return NewObjectValue(data)

	case jsonparser.Array:
		// Arrays should be handled separately as collections
		return nil

	case jsonparser.Null:
		return nil
	}

	return nil
}

// jsonArrayToCollection converts a JSON array to a Collection.
func jsonArrayToCollection(data []byte) Collection {
	var result Collection
	jsonparser.ArrayEach(data, func(value []byte, dataType jsonparser.ValueType, _ int, _ error) {
		v := jsonValueToFHIRValue(value, dataType)
		if v != nil {
			result = append(result, v)
		}
	})
	return result
}

// JSONToCollection converts JSON bytes to a Collection.
func JSONToCollection(data []byte) (Collection, error) {
	// Detect JSON type
	value, dataType, _, err := jsonparser.Get(data)
	if err != nil {
		return nil, err
	}

	switch dataType {
	case jsonparser.Object:
		return Collection{NewObjectValue(value)}, nil
	case jsonparser.Array:
		return jsonArrayToCollection(value), nil
	case jsonparser.Null:
		return Collection{}, nil
	default:
		v := jsonValueToFHIRValue(value, dataType)
		if v == nil {
			return Collection{}, nil
		}
		return Collection{v}, nil
	}
}

// ToQuantity attempts to convert an ObjectValue to a Quantity.
// This is used when the object represents a FHIR Quantity type
// (with fields like "value", "unit", "code", "system").
// Returns the Quantity and true if successful, or zero Quantity and false if not.
func (o *ObjectValue) ToQuantity() (Quantity, bool) {
	// Try to get the "value" field (required for Quantity)
	valueBytes, dataType, _, err := jsonparser.Get(o.data, "value")
	if err != nil || dataType == jsonparser.NotExist {
		return Quantity{}, false
	}

	// Parse the numeric value
	var val decimal.Decimal
	if dataType == jsonparser.Number {
		s := string(valueBytes)
		val, err = decimal.NewFromString(s)
		if err != nil {
			return Quantity{}, false
		}
	} else {
		return Quantity{}, false
	}

	// Try to get the unit - can be "unit" or "code" field
	unit := ""
	if unitBytes, _, _, err := jsonparser.Get(o.data, "unit"); err == nil {
		unit = string(unitBytes)
	} else if codeBytes, _, _, err := jsonparser.Get(o.data, "code"); err == nil {
		unit = string(codeBytes)
	}

	return NewQuantityFromDecimal(val, unit), true
}
