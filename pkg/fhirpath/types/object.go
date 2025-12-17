package types

import (
	"bytes"
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

// FHIR type constants for type inference.
const (
	typeQuantity        = "Quantity"
	typeCoding          = "Coding"
	typeCodeableConcept = "CodeableConcept"
	typeReference       = "Reference"
	typePeriod          = "Period"
	typeIdentifier      = "Identifier"
	typeRange           = "Range"
	typeRatio           = "Ratio"
	typeAttachment      = "Attachment"
	typeHumanName       = "HumanName"
	typeAddress         = "Address"
	typeContactPoint    = "ContactPoint"
	typeAnnotation      = "Annotation"
	typeObject          = "Object"
)

// Type returns the FHIR type of this object.
// First checks resourceType, then attempts to infer common FHIR types from structure.
func (o *ObjectValue) Type() string {
	// First, check for explicit resourceType (FHIR resources)
	if rt, err := jsonparser.GetString(o.data, "resourceType"); err == nil {
		return rt
	}

	// Try to infer type from structure for common FHIR complex types
	return o.inferType()
}

// inferType attempts to infer the FHIR type from the object's structure.
// Uses a series of helper methods to reduce cyclomatic complexity.
func (o *ObjectValue) inferType() string {
	if t := o.inferQuantityType(); t != "" {
		return t
	}
	if t := o.inferCodingType(); t != "" {
		return t
	}
	if t := o.inferComplexTypes(); t != "" {
		return t
	}
	return typeObject
}

// inferQuantityType checks if the object is a Quantity type.
func (o *ObjectValue) inferQuantityType() string {
	if o.hasField("value") {
		if o.hasField("unit") || o.hasField("code") || o.hasField("system") {
			return typeQuantity
		}
	}
	return ""
}

// inferCodingType checks if the object is a Coding type.
func (o *ObjectValue) inferCodingType() string {
	if o.hasField("system") && o.hasField("code") && !o.hasField("value") {
		return typeCoding
	}
	return ""
}

// inferComplexTypes checks for various FHIR complex types.
func (o *ObjectValue) inferComplexTypes() string {
	// CodeableConcept
	if o.hasArrayField("coding") {
		return typeCodeableConcept
	}

	// Reference
	if o.hasField("reference") {
		return typeReference
	}

	// Period
	if o.hasPeriodFields() {
		return typePeriod
	}

	// Identifier
	if o.hasIdentifierFields() {
		return typeIdentifier
	}

	// Range
	if o.hasField("low") || o.hasField("high") {
		return typeRange
	}

	// Ratio
	if o.hasField("numerator") || o.hasField("denominator") {
		return typeRatio
	}

	// Attachment
	if o.hasField("contentType") {
		return typeAttachment
	}

	// HumanName
	if o.hasHumanNameFields() {
		return typeHumanName
	}

	// Address
	if o.hasAddressFields() {
		return typeAddress
	}

	// ContactPoint
	if o.hasContactPointFields() {
		return typeContactPoint
	}

	// Annotation
	if o.hasAnnotationFields() {
		return typeAnnotation
	}

	return ""
}

// hasArrayField checks if a field exists and is an array.
func (o *ObjectValue) hasArrayField(name string) bool {
	_, dataType, _, err := jsonparser.Get(o.data, name)
	return err == nil && dataType == jsonparser.Array
}

func (o *ObjectValue) hasPeriodFields() bool {
	hasStart := o.hasField("start")
	hasEnd := o.hasField("end")
	return hasStart || hasEnd
}

// hasField checks if a field exists in the object.
func (o *ObjectValue) hasField(name string) bool {
	//nolint:dogsled // jsonparser.Get returns 4 values, we only need the error
	_, _, _, err := jsonparser.Get(o.data, name)
	return err == nil
}

func (o *ObjectValue) hasIdentifierFields() bool {
	return o.hasField("system") && o.hasStringField("value")
}

// hasStringField checks if a field exists and is a string.
func (o *ObjectValue) hasStringField(name string) bool {
	_, dataType, _, err := jsonparser.Get(o.data, name)
	return err == nil && dataType == jsonparser.String
}

func (o *ObjectValue) hasHumanNameFields() bool {
	return o.hasField("family") || o.hasArrayField("given")
}

func (o *ObjectValue) hasAddressFields() bool {
	return o.hasField("city") || o.hasField("postalCode")
}

func (o *ObjectValue) hasContactPointFields() bool {
	return o.hasField("system") && o.hasField("use")
}

func (o *ObjectValue) hasAnnotationFields() bool {
	if !o.hasField("text") {
		return false
	}
	return o.hasField("time") || o.hasField("authorReference") || o.hasField("authorString")
}

// Equal returns true if the JSON data is identical.
func (o *ObjectValue) Equal(other Value) bool {
	if ov, ok := other.(*ObjectValue); ok {
		return bytes.Equal(o.data, ov.data)
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
	//nolint:errcheck // ObjectEach only returns errors for non-objects; o.data is always a valid object
	jsonparser.ObjectEach(o.data, func(key []byte, _ []byte, _ jsonparser.ValueType, _ int) error {
		keys = append(keys, string(key))
		return nil
	})
	return keys
}

// Children returns a collection of all child values.
func (o *ObjectValue) Children() Collection {
	var result Collection
	//nolint:errcheck // ObjectEach only returns errors for non-objects; o.data is always a valid object
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
			if i, err := jsonparser.ParseInt(data); err == nil {
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
		b, err := jsonparser.ParseBoolean(data)
		if err != nil {
			return nil
		}
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
	//nolint:errcheck // ArrayEach only returns errors for non-arrays; data is already validated as array
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
