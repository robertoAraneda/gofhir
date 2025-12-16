// Package ucum provides UCUM (Unified Code for Units of Measure) normalization
// for FHIR quantity search parameters.
//
// UCUM is the standard unit system used in FHIR for quantities.
// This package normalizes units to canonical base units to enable
// cross-unit search (e.g., 10mg = 0.01g).
//
// Reference: https://ucum.org/ucum.html
package ucum

import (
	"strings"
)

// NormalizedQuantity represents a quantity normalized to canonical UCUM units.
type NormalizedQuantity struct {
	Value float64 // Normalized value in canonical units
	Code  string  // Canonical unit code
}

// UnitConversion defines a conversion from a unit to its canonical form.
type UnitConversion struct {
	CanonicalCode string  // The canonical unit code (e.g., "g" for mass)
	Factor        float64 // Multiply original value by this to get canonical
}

// canonicalUnits maps UCUM codes to their canonical conversions.
// Organized by dimension (mass, length, volume, time, etc.)
var canonicalUnits = map[string]UnitConversion{
	// === MASS (canonical: g) ===
	"kg":      {CanonicalCode: "g", Factor: 1000},
	"g":       {CanonicalCode: "g", Factor: 1},
	"mg":      {CanonicalCode: "g", Factor: 0.001},
	"ug":      {CanonicalCode: "g", Factor: 0.000001},
	"ng":      {CanonicalCode: "g", Factor: 0.000000001},
	"pg":      {CanonicalCode: "g", Factor: 0.000000000001},
	"lb":      {CanonicalCode: "g", Factor: 453.59237},    // avoirdupois pound
	"oz":      {CanonicalCode: "g", Factor: 28.349523125}, // avoirdupois ounce
	"[lb_av]": {CanonicalCode: "g", Factor: 453.59237},
	"[oz_av]": {CanonicalCode: "g", Factor: 28.349523125},

	// === LENGTH (canonical: m) ===
	"km":     {CanonicalCode: "m", Factor: 1000},
	"m":      {CanonicalCode: "m", Factor: 1},
	"dm":     {CanonicalCode: "m", Factor: 0.1},
	"cm":     {CanonicalCode: "m", Factor: 0.01},
	"mm":     {CanonicalCode: "m", Factor: 0.001},
	"um":     {CanonicalCode: "m", Factor: 0.000001},
	"nm":     {CanonicalCode: "m", Factor: 0.000000001},
	"[in_i]": {CanonicalCode: "m", Factor: 0.0254},   // international inch
	"[ft_i]": {CanonicalCode: "m", Factor: 0.3048},   // international foot
	"[yd_i]": {CanonicalCode: "m", Factor: 0.9144},   // international yard
	"[mi_i]": {CanonicalCode: "m", Factor: 1609.344}, // international mile
	"in":     {CanonicalCode: "m", Factor: 0.0254},
	"ft":     {CanonicalCode: "m", Factor: 0.3048},

	// === VOLUME (canonical: L) ===
	"L":        {CanonicalCode: "L", Factor: 1},
	"l":        {CanonicalCode: "L", Factor: 1},
	"dL":       {CanonicalCode: "L", Factor: 0.1},
	"dl":       {CanonicalCode: "L", Factor: 0.1},
	"cL":       {CanonicalCode: "L", Factor: 0.01},
	"cl":       {CanonicalCode: "L", Factor: 0.01},
	"mL":       {CanonicalCode: "L", Factor: 0.001},
	"ml":       {CanonicalCode: "L", Factor: 0.001},
	"uL":       {CanonicalCode: "L", Factor: 0.000001},
	"ul":       {CanonicalCode: "L", Factor: 0.000001},
	"[gal_us]": {CanonicalCode: "L", Factor: 3.785411784},
	"[qt_us]":  {CanonicalCode: "L", Factor: 0.946352946},
	"[pt_us]":  {CanonicalCode: "L", Factor: 0.473176473},
	"[foz_us]": {CanonicalCode: "L", Factor: 0.0295735295625},

	// === TIME (canonical: s) ===
	"a":   {CanonicalCode: "s", Factor: 31557600},    // Julian year
	"mo":  {CanonicalCode: "s", Factor: 2629800},     // month (30.4375 days)
	"wk":  {CanonicalCode: "s", Factor: 604800},      // week
	"d":   {CanonicalCode: "s", Factor: 86400},       // day
	"h":   {CanonicalCode: "s", Factor: 3600},        // hour
	"min": {CanonicalCode: "s", Factor: 60},          // minute
	"s":   {CanonicalCode: "s", Factor: 1},           // second
	"ms":  {CanonicalCode: "s", Factor: 0.001},       // millisecond
	"us":  {CanonicalCode: "s", Factor: 0.000001},    // microsecond
	"ns":  {CanonicalCode: "s", Factor: 0.000000001}, // nanosecond

	// === TEMPERATURE (canonical: K) ===
	"K":      {CanonicalCode: "K", Factor: 1},   // Kelvin
	"Cel":    {CanonicalCode: "Cel", Factor: 1}, // Celsius (special handling needed)
	"[degF]": {CanonicalCode: "Cel", Factor: 1}, // Fahrenheit (special handling needed)

	// === CONCENTRATION (mass/volume) ===
	"g/L":   {CanonicalCode: "g/L", Factor: 1},
	"mg/L":  {CanonicalCode: "g/L", Factor: 0.001},
	"ug/L":  {CanonicalCode: "g/L", Factor: 0.000001},
	"ng/L":  {CanonicalCode: "g/L", Factor: 0.000000001},
	"g/dL":  {CanonicalCode: "g/L", Factor: 10},
	"mg/dL": {CanonicalCode: "g/L", Factor: 0.01},
	"ug/dL": {CanonicalCode: "g/L", Factor: 0.00001},
	"g/mL":  {CanonicalCode: "g/L", Factor: 1000},
	"mg/mL": {CanonicalCode: "g/L", Factor: 1},
	"ug/mL": {CanonicalCode: "g/L", Factor: 0.001},

	// === MOLAR CONCENTRATION (canonical: mol/L) ===
	"mol/L":  {CanonicalCode: "mol/L", Factor: 1},
	"mmol/L": {CanonicalCode: "mol/L", Factor: 0.001},
	"umol/L": {CanonicalCode: "mol/L", Factor: 0.000001},
	"nmol/L": {CanonicalCode: "mol/L", Factor: 0.000000001},
	"pmol/L": {CanonicalCode: "mol/L", Factor: 0.000000000001},

	// === PRESSURE (canonical: Pa) ===
	"Pa":     {CanonicalCode: "Pa", Factor: 1},
	"kPa":    {CanonicalCode: "Pa", Factor: 1000},
	"mm[Hg]": {CanonicalCode: "Pa", Factor: 133.322387415},
	"[psi]":  {CanonicalCode: "Pa", Factor: 6894.757293168},

	// === COUNT/CELLS ===
	"10*9/L":  {CanonicalCode: "10*9/L", Factor: 1},        // billions per liter (common for WBC)
	"10*12/L": {CanonicalCode: "10*9/L", Factor: 1000},     // trillions per liter (common for RBC)
	"10*6/L":  {CanonicalCode: "10*9/L", Factor: 0.001},    // millions per liter
	"10*3/uL": {CanonicalCode: "10*9/L", Factor: 1},        // thousands per microliter = billions per liter
	"/uL":     {CanonicalCode: "10*9/L", Factor: 0.000001}, // per microliter

	// === PERCENTAGE ===
	"%": {CanonicalCode: "%", Factor: 1},

	// === RATE ===
	"/min": {CanonicalCode: "/min", Factor: 1},          // per minute (heart rate, resp rate)
	"/h":   {CanonicalCode: "/min", Factor: 1.0 / 60.0}, // per hour

	// === INTERNATIONAL UNITS ===
	"[IU]":     {CanonicalCode: "[IU]", Factor: 1},
	"[IU]/L":   {CanonicalCode: "[IU]/L", Factor: 1},
	"[IU]/mL":  {CanonicalCode: "[IU]/L", Factor: 1000},
	"m[IU]/L":  {CanonicalCode: "[IU]/L", Factor: 0.001},
	"m[IU]/mL": {CanonicalCode: "[IU]/L", Factor: 1},
	"u[IU]/mL": {CanonicalCode: "[IU]/L", Factor: 0.001},

	// === ENERGY ===
	"J":     {CanonicalCode: "J", Factor: 1},
	"kJ":    {CanonicalCode: "J", Factor: 1000},
	"cal":   {CanonicalCode: "J", Factor: 4.184},
	"kcal":  {CanonicalCode: "J", Factor: 4184},
	"[Cal]": {CanonicalCode: "J", Factor: 4184},
}

// Normalize converts a quantity to its canonical UCUM form.
// Returns the original values if the unit is not recognized.
func Normalize(value float64, code string) NormalizedQuantity {
	// Try exact match first
	if conv, ok := canonicalUnits[code]; ok {
		return NormalizedQuantity{
			Value: value * conv.Factor,
			Code:  conv.CanonicalCode,
		}
	}

	// Try case-insensitive match for common variations
	for ucumCode, conv := range canonicalUnits {
		if strings.EqualFold(ucumCode, code) {
			return NormalizedQuantity{
				Value: value * conv.Factor,
				Code:  conv.CanonicalCode,
			}
		}
	}

	// Unknown unit - return as-is
	return NormalizedQuantity{
		Value: value,
		Code:  code,
	}
}

// NormalizeWithSystem converts a quantity considering both system and code.
// For UCUM system (http://unitsofmeasure.org), it applies normalization.
// For other systems, it returns values unchanged.
func NormalizeWithSystem(value float64, system, code string) NormalizedQuantity {
	// Only normalize UCUM units
	if system != "" && system != "http://unitsofmeasure.org" {
		return NormalizedQuantity{
			Value: value,
			Code:  code,
		}
	}

	return Normalize(value, code)
}

// IsKnownUnit returns true if the unit code is recognized for normalization.
func IsKnownUnit(code string) bool {
	if _, ok := canonicalUnits[code]; ok {
		return true
	}

	for ucumCode := range canonicalUnits {
		if strings.EqualFold(ucumCode, code) {
			return true
		}
	}

	return false
}

// GetCanonicalUnit returns the canonical unit for a given code.
// Returns the original code if not found.
func GetCanonicalUnit(code string) string {
	if conv, ok := canonicalUnits[code]; ok {
		return conv.CanonicalCode
	}

	for ucumCode, conv := range canonicalUnits {
		if strings.EqualFold(ucumCode, code) {
			return conv.CanonicalCode
		}
	}

	return code
}
