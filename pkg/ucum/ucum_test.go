package ucum

import (
	"math"
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		code      string
		wantValue float64
		wantCode  string
		tolerance float64
	}{
		// Mass conversions
		{"kg to g", 1, "kg", 1000, "g", 0.0001},
		{"mg to g", 100, "mg", 0.1, "g", 0.0001},
		{"ug to g", 1000, "ug", 0.001, "g", 0.0001},
		{"g unchanged", 5, "g", 5, "g", 0.0001},
		{"lb to g", 1, "lb", 453.59237, "g", 0.0001},

		// Length conversions
		{"km to m", 1, "km", 1000, "m", 0.0001},
		{"cm to m", 100, "cm", 1, "m", 0.0001},
		{"mm to m", 1000, "mm", 1, "m", 0.0001},
		{"inch to m", 1, "[in_i]", 0.0254, "m", 0.0001},
		{"foot to m", 1, "[ft_i]", 0.3048, "m", 0.0001},

		// Volume conversions
		{"mL to L", 1000, "mL", 1, "L", 0.0001},
		{"dL to L", 10, "dL", 1, "L", 0.0001},
		{"L unchanged", 5, "L", 5, "L", 0.0001},
		{"lowercase l", 5, "l", 5, "L", 0.0001},

		// Time conversions
		{"min to s", 1, "min", 60, "s", 0.0001},
		{"h to s", 1, "h", 3600, "s", 0.0001},
		{"d to s", 1, "d", 86400, "s", 0.0001},
		{"ms to s", 1000, "ms", 1, "s", 0.0001},

		// Concentration conversions
		{"mg/dL to g/L", 100, "mg/dL", 1, "g/L", 0.0001},
		{"g/dL to g/L", 1, "g/dL", 10, "g/L", 0.0001},
		{"mg/mL to g/L", 1, "mg/mL", 1, "g/L", 0.0001},

		// Molar concentration
		{"mmol/L to mol/L", 1, "mmol/L", 0.001, "mol/L", 0.0001},
		{"umol/L to mol/L", 1000, "umol/L", 0.001, "mol/L", 0.0001},

		// Pressure
		{"mmHg to Pa", 1, "mm[Hg]", 133.322387415, "Pa", 0.0001},
		{"kPa to Pa", 1, "kPa", 1000, "Pa", 0.0001},

		// Cell counts
		{"10*12/L to 10*9/L", 1, "10*12/L", 1000, "10*9/L", 0.0001},
		{"10*3/uL to 10*9/L", 5, "10*3/uL", 5, "10*9/L", 0.0001},

		// Energy
		{"kcal to J", 1, "kcal", 4184, "J", 0.0001},
		{"cal to J", 1, "cal", 4.184, "J", 0.0001},

		// Unknown unit - should return as-is
		{"unknown unit", 42, "unknownUnit", 42, "unknownUnit", 0.0001},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.value, tt.code)
			if math.Abs(got.Value-tt.wantValue) > tt.tolerance {
				t.Errorf("Normalize(%v, %q).Value = %v, want %v", tt.value, tt.code, got.Value, tt.wantValue)
			}
			if got.Code != tt.wantCode {
				t.Errorf("Normalize(%v, %q).Code = %q, want %q", tt.value, tt.code, got.Code, tt.wantCode)
			}
		})
	}
}

func TestNormalize_CaseInsensitive(t *testing.T) {
	tests := []struct {
		code     string
		wantCode string
	}{
		{"ML", "L"},
		{"Ml", "L"},
		{"MG", "g"},
		{"Mg", "g"},
		{"KG", "g"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := Normalize(1, tt.code)
			if got.Code != tt.wantCode {
				t.Errorf("Normalize(1, %q).Code = %q, want %q", tt.code, got.Code, tt.wantCode)
			}
		})
	}
}

func TestNormalizeWithSystem(t *testing.T) {
	tests := []struct {
		name      string
		value     float64
		system    string
		code      string
		wantValue float64
		wantCode  string
	}{
		{
			name:      "UCUM system normalizes",
			value:     100,
			system:    "http://unitsofmeasure.org",
			code:      "mg",
			wantValue: 0.1,
			wantCode:  "g",
		},
		{
			name:      "empty system normalizes",
			value:     100,
			system:    "",
			code:      "mg",
			wantValue: 0.1,
			wantCode:  "g",
		},
		{
			name:      "non-UCUM system unchanged",
			value:     100,
			system:    "http://example.org/units",
			code:      "mg",
			wantValue: 100,
			wantCode:  "mg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeWithSystem(tt.value, tt.system, tt.code)
			if math.Abs(got.Value-tt.wantValue) > 0.0001 {
				t.Errorf("NormalizeWithSystem().Value = %v, want %v", got.Value, tt.wantValue)
			}
			if got.Code != tt.wantCode {
				t.Errorf("NormalizeWithSystem().Code = %q, want %q", got.Code, tt.wantCode)
			}
		})
	}
}

func TestIsKnownUnit(t *testing.T) {
	tests := []struct {
		code string
		want bool
	}{
		{"g", true},
		{"mg", true},
		{"kg", true},
		{"L", true},
		{"mL", true},
		{"ml", true}, // case insensitive
		{"ML", true}, // case insensitive
		{"mmol/L", true},
		{"mm[Hg]", true},
		{"%", true},
		{"unknownUnit", false},
		{"xyz", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := IsKnownUnit(tt.code); got != tt.want {
				t.Errorf("IsKnownUnit(%q) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestGetCanonicalUnit(t *testing.T) {
	tests := []struct {
		code string
		want string
	}{
		{"mg", "g"},
		{"kg", "g"},
		{"g", "g"},
		{"mL", "L"},
		{"dL", "L"},
		{"L", "L"},
		{"cm", "m"},
		{"km", "m"},
		{"min", "s"},
		{"h", "s"},
		{"mmol/L", "mol/L"},
		{"mg/dL", "g/L"},
		{"unknownUnit", "unknownUnit"}, // returns original if not found
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			if got := GetCanonicalUnit(tt.code); got != tt.want {
				t.Errorf("GetCanonicalUnit(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}

func TestNormalize_RealWorldExamples(t *testing.T) {
	// Test real-world clinical values
	tests := []struct {
		name        string
		value       float64
		code        string
		description string
	}{
		{"glucose mg/dL", 100, "mg/dL", "Normal fasting glucose ~100 mg/dL = 1 g/L"},
		{"hemoglobin g/dL", 14, "g/dL", "Normal hemoglobin ~14 g/dL = 140 g/L"},
		{"potassium mmol/L", 4.5, "mmol/L", "Normal potassium ~4.5 mmol/L = 0.0045 mol/L"},
		{"blood pressure mmHg", 120, "mm[Hg]", "Systolic BP ~120 mmHg"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Normalize(tt.value, tt.code)
			// Just verify it normalizes without error
			if got.Code == "" {
				t.Errorf("Normalize(%v, %q) returned empty code", tt.value, tt.code)
			}
			t.Logf("%s: %v %s -> %v %s", tt.description, tt.value, tt.code, got.Value, got.Code)
		})
	}
}
