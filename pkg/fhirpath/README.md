# FHIRPath

A complete FHIRPath 2.0 expression evaluator for FHIR resources in Go.

## Overview

This package implements the [FHIRPath specification](https://hl7.org/fhirpath/) for evaluating path expressions against FHIR resources. It supports parsing, compiling, and evaluating FHIRPath expressions with full type safety and UCUM unit normalization.

## Installation

```go
import "github.com/robertoaraneda/gofhir/pkg/fhirpath"
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/robertoaraneda/gofhir/pkg/fhirpath"
)

func main() {
    patient := []byte(`{
        "resourceType": "Patient",
        "id": "123",
        "name": [{"family": "Doe", "given": ["John"]}],
        "birthDate": "1990-05-15"
    }`)

    // Simple evaluation
    result, err := fhirpath.Evaluate(patient, "Patient.name.family")
    if err != nil {
        panic(err)
    }
    fmt.Println(result) // ["Doe"]

    // Compile once, evaluate many times
    expr := fhirpath.MustCompile("birthDate > @1980-01-01")
    result = expr.Evaluate(patient)
    fmt.Println(result) // [true]
}
```

## API Reference

### Main Functions

| Function | Description |
|----------|-------------|
| `Evaluate(resource []byte, expr string) (Collection, error)` | Evaluate expression on resource |
| `MustEvaluate(resource []byte, expr string) Collection` | Evaluate, panic on error |
| `Compile(expr string) (*Expression, error)` | Compile expression for reuse |
| `MustCompile(expr string) *Expression` | Compile, panic on error |

### Expression Methods

```go
expr := fhirpath.MustCompile("Patient.name.given")
result := expr.Evaluate(patientJSON)
```

## Type System

FHIRPath defines a type system that this implementation fully supports:

| Type | Go Type | Example |
|------|---------|---------|
| Boolean | `types.Boolean` | `true`, `false` |
| Integer | `types.Integer` | `42`, `-17` |
| Decimal | `types.Decimal` | `3.14159` |
| String | `types.String` | `'hello'` |
| Date | `types.Date` | `@2024-01-15` |
| DateTime | `types.DateTime` | `@2024-01-15T10:30:00Z` |
| Time | `types.Time` | `@T14:30:00` |
| Quantity | `types.Quantity` | `10 'mg'`, `100 'cm'` |

### Quantity with UCUM Normalization

Quantities support UCUM unit normalization for comparison:

```go
// These evaluate to true
fhirpath.MustEvaluate(resource, "1000 'mg' = 1 'g'")
fhirpath.MustEvaluate(resource, "100 'cm' = 1 'm'")
fhirpath.MustEvaluate(resource, "60 'min' = 1 'h'")
fhirpath.MustEvaluate(resource, "1 'kg' ~ 1000 'g'")
```

## Operators

### Arithmetic Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `+` | Addition | `2 + 3` → `5` |
| `-` | Subtraction | `5 - 2` → `3` |
| `*` | Multiplication | `3 * 4` → `12` |
| `/` | Division | `10 / 4` → `2.5` |
| `div` | Integer division | `10 div 4` → `2` |
| `mod` | Modulo | `10 mod 3` → `1` |

### Comparison Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `=` | Equals | `name = 'John'` |
| `!=` | Not equals | `status != 'active'` |
| `<` | Less than | `age < 18` |
| `>` | Greater than | `value > 100` |
| `<=` | Less or equal | `count <= 10` |
| `>=` | Greater or equal | `priority >= 1` |
| `~` | Equivalent | `'Hello' ~ 'hello'` |
| `!~` | Not equivalent | `code !~ 'ABC'` |

### Boolean Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `and` | Logical AND | `active and verified` |
| `or` | Logical OR | `draft or pending` |
| `not` | Logical NOT | `not deceased` |
| `implies` | Implication | `a implies b` |
| `xor` | Exclusive OR | `a xor b` |

### Collection Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `\|` | Union | `name \| alias` |
| `in` | Membership | `'active' in status` |
| `contains` | Contains | `codes contains 'ABC'` |

### Type Operators

| Operator | Description | Example |
|----------|-------------|---------|
| `is` | Type check | `value is Quantity` |
| `as` | Type cast | `value as String` |

## Functions

### String Functions

| Function | Description | Example |
|----------|-------------|---------|
| `startsWith(prefix)` | Check prefix | `name.startsWith('Dr')` |
| `endsWith(suffix)` | Check suffix | `file.endsWith('.pdf')` |
| `contains(substring)` | Check contains | `text.contains('error')` |
| `replace(old, new)` | Replace text | `name.replace('-', '_')` |
| `matches(regex)` | Regex match | `code.matches('[A-Z]{3}')` |
| `replaceMatches(regex, sub)` | Regex replace | `text.replaceMatches('\\s+', ' ')` |
| `indexOf(substring)` | Find position | `text.indexOf(':')` |
| `substring(start[, length])` | Extract substring | `code.substring(0, 3)` |
| `lower()` | Lowercase | `name.lower()` |
| `upper()` | Uppercase | `code.upper()` |
| `length()` | String length | `name.length()` |
| `toChars()` | Split to chars | `'abc'.toChars()` |
| `trim()` | Trim whitespace | `input.trim()` |
| `split(separator)` | Split string | `csv.split(',')` |
| `join([separator])` | Join collection | `names.join(', ')` |
| `encode(encoding)` | Encode string | `text.encode('base64')` |
| `decode(encoding)` | Decode string | `data.decode('base64')` |

### Math Functions

| Function | Description | Example |
|----------|-------------|---------|
| `abs()` | Absolute value | `(-5).abs()` → `5` |
| `ceiling()` | Round up | `(3.2).ceiling()` → `4` |
| `floor()` | Round down | `(3.8).floor()` → `3` |
| `truncate()` | Remove decimals | `(3.9).truncate()` → `3` |
| `round([precision])` | Round | `(3.456).round(2)` → `3.46` |
| `exp()` | Exponential | `(2).exp()` |
| `ln()` | Natural log | `(10).ln()` |
| `log(base)` | Logarithm | `(100).log(10)` → `2` |
| `power(exp)` | Power | `(2).power(3)` → `8` |
| `sqrt()` | Square root | `(16).sqrt()` → `4` |

### Existence Functions

| Function | Description | Example |
|----------|-------------|---------|
| `empty()` | Collection empty | `name.empty()` |
| `exists([criteria])` | Any exist | `telecom.exists(system='email')` |
| `all(criteria)` | All match | `name.all(family.exists())` |
| `allTrue()` | All true | `flags.allTrue()` |
| `anyTrue()` | Any true | `conditions.anyTrue()` |
| `allFalse()` | All false | `errors.allFalse()` |
| `anyFalse()` | Any false | `checks.anyFalse()` |
| `count()` | Count items | `name.count()` |
| `distinct()` | Remove duplicates | `codes.distinct()` |
| `isDistinct()` | All unique | `ids.isDistinct()` |

### Filtering Functions

| Function | Description | Example |
|----------|-------------|---------|
| `where(criteria)` | Filter items | `name.where(use='official')` |
| `select(projection)` | Transform items | `telecom.select(value)` |
| `repeat(expression)` | Recursive navigation | `contained.repeat(children())` |
| `ofType(type)` | Filter by type | `value.ofType(Quantity)` |

### Subsetting Functions

| Function | Description | Example |
|----------|-------------|---------|
| `first()` | First item | `name.first()` |
| `last()` | Last item | `entry.last()` |
| `tail()` | All except first | `items.tail()` |
| `take(n)` | First n items | `results.take(5)` |
| `skip(n)` | Skip n items | `entries.skip(10)` |
| `single()` | Exactly one | `identifier.single()` |
| `intersect(other)` | Intersection | `a.intersect(b)` |
| `exclude(other)` | Exclusion | `all.exclude(removed)` |

### Combining Functions

| Function | Description | Example |
|----------|-------------|---------|
| `union(other)` | Set union | `names.union(aliases)` |
| `combine(other)` | Concatenate | `first.combine(second)` |

### Aggregate Functions

| Function | Description | Example |
|----------|-------------|---------|
| `aggregate(init, accumulator)` | Reduce collection | `values.aggregate(0, $total + $this)` |

### Conversion Functions

| Function | Description | Example |
|----------|-------------|---------|
| `iif(cond, true[, false])` | Conditional | `iif(active, 'Yes', 'No')` |
| `toBoolean()` | Convert to boolean | `'true'.toBoolean()` |
| `convertsToBoolean()` | Can convert | `value.convertsToBoolean()` |
| `toInteger()` | Convert to integer | `'42'.toInteger()` |
| `convertsToInteger()` | Can convert | `value.convertsToInteger()` |
| `toDecimal()` | Convert to decimal | `'3.14'.toDecimal()` |
| `convertsToDecimal()` | Can convert | `value.convertsToDecimal()` |
| `toString()` | Convert to string | `(42).toString()` |
| `convertsToString()` | Can convert | `value.convertsToString()` |
| `toDate()` | Convert to date | `'2024-01-15'.toDate()` |
| `convertsToDate()` | Can convert | `value.convertsToDate()` |
| `toDateTime()` | Convert to datetime | `date.toDateTime()` |
| `convertsToDateTime()` | Can convert | `value.convertsToDateTime()` |
| `toTime()` | Convert to time | `'14:30:00'.toTime()` |
| `convertsToTime()` | Can convert | `value.convertsToTime()` |
| `toQuantity([unit])` | Convert to quantity | `value.toQuantity('mg')` |
| `convertsToQuantity([unit])` | Can convert | `value.convertsToQuantity('kg')` |

### Temporal Functions

| Function | Description | Example |
|----------|-------------|---------|
| `now()` | Current datetime | `now()` |
| `today()` | Current date | `today()` |
| `timeOfDay()` | Current time | `timeOfDay()` |
| `year()` | Extract year | `birthDate.year()` |
| `month()` | Extract month | `birthDate.month()` |
| `day()` | Extract day | `birthDate.day()` |
| `hour()` | Extract hour | `time.hour()` |
| `minute()` | Extract minute | `time.minute()` |
| `second()` | Extract second | `time.second()` |
| `millisecond()` | Extract ms | `time.millisecond()` |

### Utility Functions

| Function | Description | Example |
|----------|-------------|---------|
| `trace([name])` | Debug output | `value.trace('debug')` |
| `children()` | Child elements | `element.children()` |
| `descendants()` | All descendants | `resource.descendants()` |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `%resource` | Root resource being evaluated |
| `%context` | Current evaluation context |
| `%ucum` | UCUM unit system URL |

```go
// Access environment variables
fhirpath.Evaluate(patient, "%resource.id")
fhirpath.Evaluate(patient, "%context.resourceType")
```

## Special Identifiers

### Backtick-Delimited Identifiers

For field names with special characters:

```go
// JSON: {"PID-1": "12345"}
fhirpath.Evaluate(resource, "Patient.`PID-1`")
```

### Polymorphic Elements (value[x])

Automatic resolution of FHIR polymorphic elements:

```go
// Observation.valueQuantity, Observation.valueString, etc.
fhirpath.Evaluate(observation, "Observation.value")  // Resolves automatically
fhirpath.Evaluate(observation, "Observation.value.ofType(Quantity)")
```

## Lazy Evaluation

The `iif()` function uses lazy evaluation - only the matching branch is evaluated:

```go
// Only evaluates the true branch, avoiding potential errors in false branch
fhirpath.Evaluate(resource, "iif(value.exists(), value.first(), 'default')")
```

## Performance

### Expression Caching

Compiled expressions are cached internally for performance:

```go
// Recommended: compile once, reuse
expr := fhirpath.MustCompile("Patient.name.where(use='official').given")
for _, patient := range patients {
    result := expr.Evaluate(patient)
}
```

### Best Practices

1. **Compile expressions** that will be used multiple times
2. **Use specific paths** rather than wildcards when possible
3. **Filter early** with `where()` to reduce collection sizes
4. **Avoid unnecessary conversions** - work with native types

## Error Handling

```go
result, err := fhirpath.Evaluate(resource, expr)
if err != nil {
    // Parse error or evaluation error
    log.Printf("FHIRPath error: %v", err)
}

if result.Empty() {
    // Expression evaluated but no results
}
```

## Specification Compliance

This implementation follows **FHIRPath Normative Release 2.0.0**:

- [x] Full type system (Boolean, Integer, Decimal, String, Date, DateTime, Time, Quantity)
- [x] All operators (arithmetic, comparison, boolean, collection, type)
- [x] All standard functions (40+ functions)
- [x] UCUM unit normalization for Quantity comparisons
- [x] Three-valued logic (empty propagation)
- [x] Lazy evaluation for `iif()`
- [x] Polymorphic element resolution (value[x])
- [x] Environment variables (%resource, %context)
- [x] Delimited identifiers (backticks)

## License

See repository root for license information.
