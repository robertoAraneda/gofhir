package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robertoaraneda/gofhir/internal/codegen/parser"
)

var samplePatientSD = []byte(`{
	"resourceType": "StructureDefinition",
	"id": "Patient",
	"url": "http://hl7.org/fhir/StructureDefinition/Patient",
	"name": "Patient",
	"title": "Patient Resource",
	"status": "active",
	"kind": "resource",
	"abstract": false,
	"type": "Patient",
	"baseDefinition": "http://hl7.org/fhir/StructureDefinition/DomainResource",
	"snapshot": {
		"element": [
			{
				"id": "Patient",
				"path": "Patient",
				"short": "Information about an individual",
				"min": 0,
				"max": "*",
				"constraint": [
					{
						"key": "pat-1",
						"severity": "error",
						"human": "Contact should have a name or organization",
						"expression": "contact.exists() implies (contact.all(name.exists() or organization.exists()))"
					}
				]
			},
			{
				"id": "Patient.id",
				"path": "Patient.id",
				"short": "Logical id",
				"min": 0,
				"max": "1",
				"type": [{"code": "id"}]
			},
			{
				"id": "Patient.active",
				"path": "Patient.active",
				"short": "Whether record is active",
				"min": 0,
				"max": "1",
				"type": [{"code": "boolean"}]
			},
			{
				"id": "Patient.name",
				"path": "Patient.name",
				"short": "A name for the patient",
				"min": 0,
				"max": "*",
				"type": [{"code": "HumanName"}]
			},
			{
				"id": "Patient.birthDate",
				"path": "Patient.birthDate",
				"short": "Date of birth",
				"min": 0,
				"max": "1",
				"type": [{"code": "date"}]
			},
			{
				"id": "Patient.deceased[x]",
				"path": "Patient.deceased[x]",
				"short": "Indicates if the patient is deceased",
				"min": 0,
				"max": "1",
				"type": [
					{"code": "boolean"},
					{"code": "dateTime"}
				]
			},
			{
				"id": "Patient.gender",
				"path": "Patient.gender",
				"short": "male | female | other | unknown",
				"min": 0,
				"max": "1",
				"type": [{"code": "code"}],
				"binding": {
					"strength": "required",
					"valueSet": "http://hl7.org/fhir/ValueSet/administrative-gender"
				}
			}
		]
	}
}`)

func TestAnalyzer_Analyze(t *testing.T) {
	sd, err := parser.ParseStructureDefinition(samplePatientSD)
	require.NoError(t, err)

	analyzer := NewAnalyzer([]*parser.StructureDefinition{sd}, nil)

	t.Run("analyze patient", func(t *testing.T) {
		result, err := analyzer.Analyze(sd)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Patient", result.Name)
		assert.Equal(t, "resource", result.Kind)
		assert.Equal(t, "Patient Resource", result.Description)
		assert.False(t, result.IsAbstract)

		// Should have constraints
		require.Len(t, result.Constraints, 1)
		assert.Equal(t, "pat-1", result.Constraints[0].Key)
	})

	t.Run("properties exist", func(t *testing.T) {
		result, err := analyzer.Analyze(sd)
		require.NoError(t, err)

		propMap := make(map[string]AnalyzedProperty)
		for _, p := range result.Properties {
			propMap[p.Name] = p
		}

		// Check id field
		idProp, ok := propMap["Id"]
		require.True(t, ok, "should have Id property")
		assert.Equal(t, "id", idProp.JSONName)
		assert.Equal(t, "*string", idProp.GoType)
		assert.True(t, idProp.IsPrimitive)
		assert.True(t, idProp.HasExtension)

		// Check active field
		activeProp, ok := propMap["Active"]
		require.True(t, ok, "should have Active property")
		assert.Equal(t, "active", activeProp.JSONName)
		assert.Equal(t, "*bool", activeProp.GoType)
		assert.True(t, activeProp.IsPrimitive)

		// Check name field (array)
		nameProp, ok := propMap["Name"]
		require.True(t, ok, "should have Name property")
		assert.Equal(t, "name", nameProp.JSONName)
		assert.Equal(t, "[]HumanName", nameProp.GoType)
		assert.True(t, nameProp.IsArray)
		assert.False(t, nameProp.IsPrimitive)

		// Check gender field with binding
		genderProp, ok := propMap["Gender"]
		require.True(t, ok, "should have Gender property")
		require.NotNil(t, genderProp.Binding)
		assert.Equal(t, "required", genderProp.Binding.Strength)
	})

	t.Run("choice type expansion", func(t *testing.T) {
		result, err := analyzer.Analyze(sd)
		require.NoError(t, err)

		propMap := make(map[string]AnalyzedProperty)
		for _, p := range result.Properties {
			propMap[p.Name] = p
		}

		// Choice type should be expanded to multiple fields
		deceasedBool, ok := propMap["DeceasedBoolean"]
		require.True(t, ok, "should have DeceasedBoolean property")
		assert.Equal(t, "deceasedBoolean", deceasedBool.JSONName)
		assert.Equal(t, "*bool", deceasedBool.GoType)
		assert.True(t, deceasedBool.IsChoice)
		assert.Contains(t, deceasedBool.ChoiceTypes, "boolean")
		assert.Contains(t, deceasedBool.ChoiceTypes, "dateTime")

		deceasedDateTime, ok := propMap["DeceasedDateTime"]
		require.True(t, ok, "should have DeceasedDateTime property")
		assert.Equal(t, "deceasedDateTime", deceasedDateTime.JSONName)
		assert.Equal(t, "*string", deceasedDateTime.GoType)
		assert.True(t, deceasedDateTime.IsChoice)

		// Should have extension fields for primitive choice types
		_, hasExtBool := propMap["DeceasedBooleanExt"]
		assert.True(t, hasExtBool, "should have DeceasedBooleanExt for primitive")

		_, hasExtDateTime := propMap["DeceasedDateTimeExt"]
		assert.True(t, hasExtDateTime, "should have DeceasedDateTimeExt for primitive")
	})

	t.Run("primitive extension fields", func(t *testing.T) {
		result, err := analyzer.Analyze(sd)
		require.NoError(t, err)

		propMap := make(map[string]AnalyzedProperty)
		for _, p := range result.Properties {
			propMap[p.Name] = p
		}

		// Primitives should have HasExtension = true
		idProp := propMap["Id"]
		assert.True(t, idProp.HasExtension)

		// Complex types should not
		nameProp := propMap["Name"]
		assert.False(t, nameProp.HasExtension)
	})
}

func TestAnalyzer_NilInput(t *testing.T) {
	analyzer := NewAnalyzer(nil, nil)
	result, err := analyzer.Analyze(nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFHIRToGoType(t *testing.T) {
	tests := []struct {
		fhirType string
		expected string
	}{
		// Primitives
		{"boolean", "bool"},
		{"integer", "int"},
		{"integer64", "int64"},
		{"decimal", "float64"},
		{"string", "string"},
		{"uri", "string"},
		{"date", "string"},
		{"dateTime", "string"},
		{"code", "string"},
		{"id", "string"},
		{"unsignedInt", "uint32"},
		{"positiveInt", "uint32"},

		// Complex types
		{"Coding", "Coding"},
		{"CodeableConcept", "CodeableConcept"},
		{"Reference", "Reference"},
		{"HumanName", "HumanName"},
		{"Extension", "Extension"},

		// Resources (pass through)
		{"Patient", "Patient"},
		{"Observation", "Observation"},
	}

	for _, tt := range tests {
		t.Run(tt.fhirType, func(t *testing.T) {
			result := FHIRToGoType(tt.fhirType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPrimitiveType(t *testing.T) {
	primitives := []string{"boolean", "integer", "string", "decimal", "date", "dateTime", "code", "uri"}
	for _, p := range primitives {
		assert.True(t, IsPrimitiveType(p), "%s should be primitive", p)
	}

	nonPrimitives := []string{"Coding", "Patient", "HumanName", "Reference"}
	for _, np := range nonPrimitives {
		assert.False(t, IsPrimitiveType(np), "%s should not be primitive", np)
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("toPascalCase", func(t *testing.T) {
		assert.Equal(t, "Name", toPascalCase("name"))
		assert.Equal(t, "BirthDate", toPascalCase("birthDate"))
		assert.Equal(t, "", toPascalCase(""))
	})

	t.Run("toLowerFirst", func(t *testing.T) {
		assert.Equal(t, "name", toLowerFirst("Name"))
		assert.Equal(t, "birthDate", toLowerFirst("BirthDate"))
		assert.Equal(t, "", toLowerFirst(""))
	})

	t.Run("toGoFieldName reserved words", func(t *testing.T) {
		assert.Equal(t, "Class", toGoFieldName("class"))
		assert.Equal(t, "Type", toGoFieldName("type"))
		assert.Equal(t, "Import", toGoFieldName("import"))
	})
}
