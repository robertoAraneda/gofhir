package eval

import (
	"testing"

	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

func TestContext(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		json := []byte(`{"name": "test"}`)
		ctx := NewContext(json)

		if ctx.Root().Empty() {
			t.Error("expected non-empty root")
		}
		if ctx.This().Empty() {
			t.Error("expected non-empty this")
		}
	})

	t.Run("variables", func(t *testing.T) {
		ctx := NewContext([]byte(`{}`))

		ctx.SetVariable("myVar", types.Collection{types.NewString("test")})

		v, ok := ctx.GetVariable("myVar")
		if !ok {
			t.Error("expected variable to exist")
		}
		if v.Empty() || v[0].(types.String).Value() != "test" {
			t.Error("expected variable value 'test'")
		}

		_, ok = ctx.GetVariable("nonexistent")
		if ok {
			t.Error("expected variable to not exist")
		}
	})
}

func TestErrors(t *testing.T) {
	t.Run("error types", func(t *testing.T) {
		tests := []struct {
			errType  ErrorType
			expected string
		}{
			{ErrParse, "ParseError"},
			{ErrType, "TypeError"},
			{ErrSingletonExpected, "SingletonExpectedError"},
			{ErrFunctionNotFound, "FunctionNotFoundError"},
			{ErrInvalidArguments, "InvalidArgumentsError"},
			{ErrDivisionByZero, "DivisionByZeroError"},
			{ErrInvalidPath, "InvalidPathError"},
			{ErrTimeout, "TimeoutError"},
			{ErrInvalidOperation, "InvalidOperationError"},
			{ErrInvalidExpression, "InvalidExpressionError"},
		}

		for _, tt := range tests {
			if tt.errType.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.errType.String())
			}
		}
	})

	t.Run("error constructors", func(t *testing.T) {
		err := ParseError("test message")
		if err.Type != ErrParse {
			t.Error("expected parse error")
		}

		err = TypeError("String", "Integer", "add")
		if err.Type != ErrType {
			t.Error("expected type error")
		}

		err = SingletonError(5)
		if err.Type != ErrSingletonExpected {
			t.Error("expected singleton error")
		}

		err = FunctionNotFoundError("myFunc")
		if err.Type != ErrFunctionNotFound {
			t.Error("expected function not found error")
		}

		err = InvalidArgumentsError("myFunc", 2, 1)
		if err.Type != ErrInvalidArguments {
			t.Error("expected invalid arguments error")
		}

		err = DivisionByZeroError()
		if err.Type != ErrDivisionByZero {
			t.Error("expected division by zero error")
		}

		err = InvalidPathError("/invalid")
		if err.Type != ErrInvalidPath {
			t.Error("expected invalid path error")
		}

		err = InvalidOperationError("+", "String", "Boolean")
		if err.Type != ErrInvalidOperation {
			t.Error("expected invalid operation error")
		}
	})

	t.Run("error message formatting", func(t *testing.T) {
		err := NewEvalError(ErrType, "test message")
		if err.Error() != "TypeError: test message" {
			t.Errorf("unexpected error message: %s", err.Error())
		}

		err = err.WithPath("Patient.name")
		if err.Path != "Patient.name" {
			t.Error("expected path to be set")
		}

		err = err.WithPosition(10, 5)
		if err.Position.Line != 10 || err.Position.Column != 5 {
			t.Error("expected position to be set")
		}
	})
}

func TestOperators(t *testing.T) {
	t.Run("add integers", func(t *testing.T) {
		result, err := Add(types.NewInteger(5), types.NewInteger(3))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Integer).Value() != 8 {
			t.Errorf("expected 8, got %v", result)
		}
	})

	t.Run("add strings", func(t *testing.T) {
		result, err := Add(types.NewString("Hello"), types.NewString(" World"))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.String).Value() != "Hello World" {
			t.Errorf("expected 'Hello World', got %v", result)
		}
	})

	t.Run("subtract", func(t *testing.T) {
		result, err := Subtract(types.NewInteger(10), types.NewInteger(3))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Integer).Value() != 7 {
			t.Errorf("expected 7, got %v", result)
		}
	})

	t.Run("multiply", func(t *testing.T) {
		result, err := Multiply(types.NewInteger(4), types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Integer).Value() != 20 {
			t.Errorf("expected 20, got %v", result)
		}
	})

	t.Run("divide", func(t *testing.T) {
		result, err := Divide(types.NewInteger(10), types.NewInteger(4))
		if err != nil {
			t.Fatal(err)
		}
		// Division returns Decimal
		if result.Type() != "Decimal" {
			t.Errorf("expected Decimal, got %s", result.Type())
		}
	})

	t.Run("divide by zero", func(t *testing.T) {
		_, err := Divide(types.NewInteger(10), types.NewInteger(0))
		if err == nil {
			t.Error("expected division by zero error")
		}
	})

	t.Run("integer divide", func(t *testing.T) {
		result, err := IntegerDivide(types.NewInteger(10), types.NewInteger(3))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Integer).Value() != 3 {
			t.Errorf("expected 3, got %v", result)
		}
	})

	t.Run("modulo", func(t *testing.T) {
		result, err := Modulo(types.NewInteger(10), types.NewInteger(3))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Integer).Value() != 1 {
			t.Errorf("expected 1, got %v", result)
		}
	})

	t.Run("comparison", func(t *testing.T) {
		result, err := LessThan(types.NewInteger(5), types.NewInteger(10))
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 5 < 10 to be true")
		}

		result, err = GreaterThan(types.NewInteger(10), types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 10 > 5 to be true")
		}
	})

	t.Run("equality", func(t *testing.T) {
		result := Equal(types.Collection{types.NewInteger(5)}, types.Collection{types.NewInteger(5)})
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 5 = 5 to be true")
		}

		result = NotEqual(types.Collection{types.NewInteger(5)}, types.Collection{types.NewInteger(10)})
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 5 != 10 to be true")
		}
	})

	t.Run("boolean operators", func(t *testing.T) {
		result := And(types.Collection{types.NewBoolean(true)}, types.Collection{types.NewBoolean(true)})
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true and true to be true")
		}

		result = Or(types.Collection{types.NewBoolean(false)}, types.Collection{types.NewBoolean(true)})
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected false or true to be true")
		}

		result = Xor(types.Collection{types.NewBoolean(true)}, types.Collection{types.NewBoolean(false)})
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true xor false to be true")
		}
	})

	t.Run("collection operators", func(t *testing.T) {
		c1 := types.Collection{types.NewInteger(1), types.NewInteger(2)}
		c2 := types.Collection{types.NewInteger(3)}

		result := Union(c1, c2)
		if result.Count() != 3 {
			t.Errorf("expected 3 elements, got %d", result.Count())
		}

		resultCol := In(types.Collection{types.NewInteger(2)}, c1)
		if !resultCol[0].(types.Boolean).Bool() {
			t.Error("expected 2 in [1,2] to be true")
		}

		resultCol = Contains(c1, types.Collection{types.NewInteger(1)})
		if !resultCol[0].(types.Boolean).Bool() {
			t.Error("expected [1,2] contains 1 to be true")
		}
	})

	t.Run("string concatenation", func(t *testing.T) {
		result := Concatenate(types.Collection{types.NewString("Hello")}, types.Collection{types.NewString(" World")})
		if result[0].(types.String).Value() != "Hello World" {
			t.Errorf("expected 'Hello World', got %v", result[0])
		}
	})

	t.Run("equivalence", func(t *testing.T) {
		result := Equivalent(types.Collection{types.NewString("HELLO")}, types.Collection{types.NewString("hello")})
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected HELLO ~ hello to be true")
		}
	})

	t.Run("not", func(t *testing.T) {
		result := Not(types.Collection{types.NewBoolean(true)})
		if result[0].(types.Boolean).Bool() {
			t.Error("expected not true to be false")
		}
	})

	t.Run("implies", func(t *testing.T) {
		// false implies X = true
		result := Implies(types.Collection{types.NewBoolean(false)}, types.Collection{types.NewBoolean(false)})
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected false implies false to be true")
		}

		// true implies true = true
		result = Implies(types.Collection{types.NewBoolean(true)}, types.Collection{types.NewBoolean(true)})
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true implies true to be true")
		}

		// true implies false = false
		result = Implies(types.Collection{types.NewBoolean(true)}, types.Collection{types.NewBoolean(false)})
		if result[0].(types.Boolean).Bool() {
			t.Error("expected true implies false to be false")
		}
	})

	t.Run("less or equal", func(t *testing.T) {
		result, err := LessOrEqual(types.NewInteger(5), types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 5 <= 5 to be true")
		}

		result, err = LessOrEqual(types.NewInteger(4), types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 4 <= 5 to be true")
		}

		result, err = LessOrEqual(types.NewInteger(6), types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 6 <= 5 to be false")
		}
	})

	t.Run("greater or equal", func(t *testing.T) {
		result, err := GreaterOrEqual(types.NewInteger(5), types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 5 >= 5 to be true")
		}

		result, err = GreaterOrEqual(types.NewInteger(6), types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 6 >= 5 to be true")
		}

		result, err = GreaterOrEqual(types.NewInteger(4), types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 4 >= 5 to be false")
		}
	})

	t.Run("not equivalent", func(t *testing.T) {
		result := NotEquivalent(types.Collection{types.NewString("HELLO")}, types.Collection{types.NewString("world")})
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected HELLO !~ world to be true")
		}

		result = NotEquivalent(types.Collection{types.NewString("hello")}, types.Collection{types.NewString("HELLO")})
		if result[0].(types.Boolean).Bool() {
			t.Error("expected hello !~ HELLO to be false")
		}
	})

	t.Run("negate", func(t *testing.T) {
		result, err := Negate(types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Integer).Value() != -5 {
			t.Errorf("expected -5, got %v", result)
		}

		result, err = Negate(types.NewDecimalFromFloat(3.14))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Decimal).Value().InexactFloat64() != -3.14 {
			t.Errorf("expected -3.14, got %v", result)
		}

		_, err = Negate(types.NewString("test"))
		if err == nil {
			t.Error("expected error for negating string")
		}
	})

	t.Run("empty collection handling", func(t *testing.T) {
		empty := types.EmptyCollection

		// Equal with empty
		result := Equal(empty, types.Collection{types.NewInteger(5)})
		if !result.Empty() {
			t.Error("expected empty for equal with empty")
		}

		// Not with empty
		result = Not(empty)
		if !result.Empty() {
			t.Error("expected empty for not empty")
		}

		// And with empty and false
		result = And(empty, types.Collection{types.NewBoolean(false)})
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false for empty and false")
		}

		// Or with empty and true
		result = Or(empty, types.Collection{types.NewBoolean(true)})
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for empty or true")
		}

		// Xor with empty
		result = Xor(empty, types.Collection{types.NewBoolean(true)})
		if !result.Empty() {
			t.Error("expected empty for xor with empty")
		}

		// Implies with empty
		result = Implies(empty, types.Collection{types.NewBoolean(true)})
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true for empty implies true")
		}

		// In with empty
		result = In(empty, types.Collection{types.NewInteger(1)})
		if !result.Empty() {
			t.Error("expected empty for in with empty left")
		}

		// Contains with empty
		result = Contains(types.Collection{types.NewInteger(1)}, empty)
		if !result.Empty() {
			t.Error("expected empty for contains with empty right")
		}
	})

	t.Run("equivalence edge cases", func(t *testing.T) {
		// Both empty
		result := Equivalent(types.EmptyCollection, types.EmptyCollection)
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected empty ~ empty to be true")
		}

		// One empty
		result = Equivalent(types.EmptyCollection, types.Collection{types.NewInteger(1)})
		if result[0].(types.Boolean).Bool() {
			t.Error("expected empty ~ 1 to be false")
		}

		// Multiple elements
		result = Equivalent(
			types.Collection{types.NewInteger(1), types.NewInteger(2)},
			types.Collection{types.NewInteger(1)},
		)
		if result[0].(types.Boolean).Bool() {
			t.Error("expected [1,2] ~ [1] to be false")
		}
	})
}

func TestMixedTypeArithmetic(t *testing.T) {
	t.Run("integer and decimal addition", func(t *testing.T) {
		result, err := Add(types.NewInteger(5), types.NewDecimalFromFloat(3.5))
		if err != nil {
			t.Fatal(err)
		}
		if result.Type() != "Decimal" {
			t.Errorf("expected Decimal, got %s", result.Type())
		}
		if result.(types.Decimal).Value().InexactFloat64() != 8.5 {
			t.Errorf("expected 8.5, got %v", result)
		}
	})

	t.Run("decimal and integer addition", func(t *testing.T) {
		result, err := Add(types.NewDecimalFromFloat(3.5), types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Decimal).Value().InexactFloat64() != 8.5 {
			t.Errorf("expected 8.5, got %v", result)
		}
	})

	t.Run("decimal subtraction", func(t *testing.T) {
		result, err := Subtract(types.NewDecimalFromFloat(10.5), types.NewDecimalFromFloat(3.5))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Decimal).Value().InexactFloat64() != 7.0 {
			t.Errorf("expected 7.0, got %v", result)
		}
	})

	t.Run("integer and decimal subtraction", func(t *testing.T) {
		result, err := Subtract(types.NewInteger(10), types.NewDecimalFromFloat(3.5))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Decimal).Value().InexactFloat64() != 6.5 {
			t.Errorf("expected 6.5, got %v", result)
		}
	})

	t.Run("decimal and integer subtraction", func(t *testing.T) {
		result, err := Subtract(types.NewDecimalFromFloat(10.5), types.NewInteger(3))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Decimal).Value().InexactFloat64() != 7.5 {
			t.Errorf("expected 7.5, got %v", result)
		}
	})

	t.Run("decimal multiplication", func(t *testing.T) {
		result, err := Multiply(types.NewDecimalFromFloat(3.0), types.NewDecimalFromFloat(4.0))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Decimal).Value().InexactFloat64() != 12.0 {
			t.Errorf("expected 12.0, got %v", result)
		}
	})

	t.Run("integer and decimal multiplication", func(t *testing.T) {
		result, err := Multiply(types.NewInteger(3), types.NewDecimalFromFloat(4.5))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Decimal).Value().InexactFloat64() != 13.5 {
			t.Errorf("expected 13.5, got %v", result)
		}
	})

	t.Run("decimal and integer multiplication", func(t *testing.T) {
		result, err := Multiply(types.NewDecimalFromFloat(3.5), types.NewInteger(4))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Decimal).Value().InexactFloat64() != 14.0 {
			t.Errorf("expected 14.0, got %v", result)
		}
	})

	t.Run("decimal division", func(t *testing.T) {
		result, err := Divide(types.NewDecimalFromFloat(10.0), types.NewDecimalFromFloat(4.0))
		if err != nil {
			t.Fatal(err)
		}
		if result.(types.Decimal).Value().InexactFloat64() != 2.5 {
			t.Errorf("expected 2.5, got %v", result)
		}
	})

	t.Run("decimal division by zero", func(t *testing.T) {
		_, err := Divide(types.NewDecimalFromFloat(10.0), types.NewDecimalFromFloat(0.0))
		if err == nil {
			t.Error("expected division by zero error")
		}
	})
}

func TestOperatorErrors(t *testing.T) {
	t.Run("add type errors", func(t *testing.T) {
		_, err := Add(types.NewBoolean(true), types.NewInteger(5))
		if err == nil {
			t.Error("expected error for boolean + integer")
		}

		_, err = Add(types.NewString("test"), types.NewInteger(5))
		if err == nil {
			t.Error("expected error for string + integer")
		}
	})

	t.Run("subtract type errors", func(t *testing.T) {
		_, err := Subtract(types.NewString("test"), types.NewInteger(5))
		if err == nil {
			t.Error("expected error for string - integer")
		}

		_, err = Subtract(types.NewBoolean(true), types.NewBoolean(false))
		if err == nil {
			t.Error("expected error for boolean - boolean")
		}
	})

	t.Run("multiply type errors", func(t *testing.T) {
		_, err := Multiply(types.NewString("test"), types.NewInteger(5))
		if err == nil {
			t.Error("expected error for string * integer")
		}
	})

	t.Run("divide type errors", func(t *testing.T) {
		_, err := Divide(types.NewString("test"), types.NewInteger(5))
		if err == nil {
			t.Error("expected error for string / integer")
		}

		_, err = Divide(types.NewInteger(5), types.NewString("test"))
		if err == nil {
			t.Error("expected error for integer / string")
		}
	})

	t.Run("integer divide type errors", func(t *testing.T) {
		_, err := IntegerDivide(types.NewDecimalFromFloat(10.5), types.NewInteger(3))
		if err == nil {
			t.Error("expected error for decimal div integer")
		}

		_, err = IntegerDivide(types.NewInteger(10), types.NewDecimalFromFloat(3.5))
		if err == nil {
			t.Error("expected error for integer div decimal")
		}
	})

	t.Run("modulo type errors", func(t *testing.T) {
		_, err := Modulo(types.NewDecimalFromFloat(10.5), types.NewInteger(3))
		if err == nil {
			t.Error("expected error for decimal mod integer")
		}

		_, err = Modulo(types.NewInteger(10), types.NewDecimalFromFloat(3.5))
		if err == nil {
			t.Error("expected error for integer mod decimal")
		}
	})

	t.Run("comparison type errors", func(t *testing.T) {
		_, err := Compare(types.NewBoolean(true), types.NewInteger(5))
		if err == nil {
			t.Error("expected error for comparing boolean to integer")
		}
	})
}

func TestContextAdvanced(t *testing.T) {
	t.Run("multiple variables", func(t *testing.T) {
		ctx := NewContext([]byte(`{}`))

		ctx.SetVariable("var1", types.Collection{types.NewString("one")})
		ctx.SetVariable("var2", types.Collection{types.NewInteger(2)})
		ctx.SetVariable("var3", types.Collection{types.NewBoolean(true)})

		v1, ok := ctx.GetVariable("var1")
		if !ok || v1[0].(types.String).Value() != "one" {
			t.Error("expected var1 = 'one'")
		}

		v2, ok := ctx.GetVariable("var2")
		if !ok || v2[0].(types.Integer).Value() != 2 {
			t.Error("expected var2 = 2")
		}

		v3, ok := ctx.GetVariable("var3")
		if !ok || !v3[0].(types.Boolean).Bool() {
			t.Error("expected var3 = true")
		}
	})

	t.Run("overwrite variable", func(t *testing.T) {
		ctx := NewContext([]byte(`{}`))

		ctx.SetVariable("myVar", types.Collection{types.NewString("first")})
		ctx.SetVariable("myVar", types.Collection{types.NewString("second")})

		v, ok := ctx.GetVariable("myVar")
		if !ok || v[0].(types.String).Value() != "second" {
			t.Error("expected myVar = 'second'")
		}
	})
}

func TestErrorMethods(t *testing.T) {
	t.Run("error with underlying", func(t *testing.T) {
		underlying := NewEvalError(ErrParse, "parse failed")
		err := NewEvalError(ErrType, "type error").WithUnderlying(underlying)
		if err.Underlying != underlying {
			t.Error("expected underlying error to be set")
		}
	})

	t.Run("error string formatting", func(t *testing.T) {
		err := NewEvalError(ErrDivisionByZero, "cannot divide by zero")
		expected := "DivisionByZeroError: cannot divide by zero"
		if err.Error() != expected {
			t.Errorf("expected %s, got %s", expected, err.Error())
		}
	})

	t.Run("unknown error type", func(t *testing.T) {
		err := ErrorType(999)
		if err.String() != "UnknownError" {
			t.Errorf("expected UnknownError, got %s", err.String())
		}
	})
}

func TestBooleanLogicEdgeCases(t *testing.T) {
	t.Run("and with non-boolean", func(t *testing.T) {
		result := And(
			types.Collection{types.NewInteger(1)},
			types.Collection{types.NewBoolean(true)},
		)
		if !result.Empty() {
			t.Error("expected empty for and with non-boolean")
		}
	})

	t.Run("or with non-boolean", func(t *testing.T) {
		result := Or(
			types.Collection{types.NewBoolean(false)},
			types.Collection{types.NewInteger(1)},
		)
		if !result.Empty() {
			t.Error("expected empty for or with non-boolean")
		}
	})

	t.Run("xor with non-boolean", func(t *testing.T) {
		result := Xor(
			types.Collection{types.NewInteger(1)},
			types.Collection{types.NewBoolean(true)},
		)
		if !result.Empty() {
			t.Error("expected empty for xor with non-boolean")
		}
	})

	t.Run("not with multiple values", func(t *testing.T) {
		result := Not(types.Collection{types.NewBoolean(true), types.NewBoolean(false)})
		if !result.Empty() {
			t.Error("expected empty for not with multiple values")
		}
	})

	t.Run("not with non-boolean", func(t *testing.T) {
		result := Not(types.Collection{types.NewInteger(1)})
		if !result.Empty() {
			t.Error("expected empty for not with non-boolean")
		}
	})
}

func TestCollectionOperatorEdgeCases(t *testing.T) {
	t.Run("in with multiple left values", func(t *testing.T) {
		result := In(
			types.Collection{types.NewInteger(1), types.NewInteger(2)},
			types.Collection{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)},
		)
		if !result.Empty() {
			t.Error("expected empty for in with multiple left values")
		}
	})

	t.Run("contains with multiple right values", func(t *testing.T) {
		result := Contains(
			types.Collection{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)},
			types.Collection{types.NewInteger(1), types.NewInteger(2)},
		)
		if !result.Empty() {
			t.Error("expected empty for contains with multiple right values")
		}
	})

	t.Run("concatenate with empty", func(t *testing.T) {
		result := Concatenate(types.EmptyCollection, types.Collection{types.NewString("world")})
		if result[0].(types.String).Value() != "world" {
			t.Errorf("expected 'world', got %v", result[0])
		}

		result = Concatenate(types.Collection{types.NewString("hello")}, types.EmptyCollection)
		if result[0].(types.String).Value() != "hello" {
			t.Errorf("expected 'hello', got %v", result[0])
		}
	})

	t.Run("equal with multiple elements", func(t *testing.T) {
		result := Equal(
			types.Collection{types.NewInteger(1), types.NewInteger(2)},
			types.Collection{types.NewInteger(1)},
		)
		if !result.Empty() {
			t.Error("expected empty for equal with multiple elements on left")
		}

		result = Equal(
			types.Collection{types.NewInteger(1)},
			types.Collection{types.NewInteger(1), types.NewInteger(2)},
		)
		if !result.Empty() {
			t.Error("expected empty for equal with multiple elements on right")
		}
	})

	t.Run("not equal with empty", func(t *testing.T) {
		result := NotEqual(types.EmptyCollection, types.Collection{types.NewInteger(1)})
		if !result.Empty() {
			t.Error("expected empty for not equal with empty left")
		}
	})
}

func TestContextMethods(t *testing.T) {
	t.Run("WithThis", func(t *testing.T) {
		ctx := NewContext([]byte(`{"name": "original"}`))
		newThis := types.Collection{types.NewString("modified")}

		newCtx := ctx.WithThis(newThis)

		// New context should have new this
		if newCtx.This()[0].(types.String).Value() != "modified" {
			t.Error("expected new context to have modified this")
		}

		// Original context should be unchanged
		if ctx.This()[0].(*types.ObjectValue) == nil {
			t.Error("expected original context to remain unchanged")
		}
	})

	t.Run("WithIndex", func(t *testing.T) {
		ctx := NewContext([]byte(`{}`))

		newCtx := ctx.WithIndex(42)

		// We can't directly access index, but we can test through evaluator
		// This at least exercises the WithIndex method
		if newCtx == nil {
			t.Error("expected non-nil context")
		}
	})

	t.Run("root and this initial values", func(t *testing.T) {
		json := []byte(`{"value": 123}`)
		ctx := NewContext(json)

		// Root and This should be the same initially
		if ctx.Root().Count() != ctx.This().Count() {
			t.Error("expected root and this to have same count")
		}
	})
}

func TestComparisonOperators(t *testing.T) {
	t.Run("compare integers", func(t *testing.T) {
		cmp, err := Compare(types.NewInteger(5), types.NewInteger(10))
		if err != nil {
			t.Fatal(err)
		}
		if cmp >= 0 {
			t.Error("expected 5 < 10")
		}

		cmp, err = Compare(types.NewInteger(10), types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if cmp <= 0 {
			t.Error("expected 10 > 5")
		}

		cmp, err = Compare(types.NewInteger(5), types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if cmp != 0 {
			t.Error("expected 5 = 5")
		}
	})

	t.Run("compare strings", func(t *testing.T) {
		cmp, err := Compare(types.NewString("apple"), types.NewString("banana"))
		if err != nil {
			t.Fatal(err)
		}
		if cmp >= 0 {
			t.Error("expected apple < banana")
		}
	})

	t.Run("compare decimals", func(t *testing.T) {
		cmp, err := Compare(types.NewDecimalFromFloat(3.14), types.NewDecimalFromFloat(2.71))
		if err != nil {
			t.Fatal(err)
		}
		if cmp <= 0 {
			t.Error("expected 3.14 > 2.71")
		}
	})

	t.Run("less than false case", func(t *testing.T) {
		result, err := LessThan(types.NewInteger(10), types.NewInteger(5))
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 10 < 5 to be false")
		}
	})

	t.Run("greater than false case", func(t *testing.T) {
		result, err := GreaterThan(types.NewInteger(5), types.NewInteger(10))
		if err != nil {
			t.Fatal(err)
		}
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 5 > 10 to be false")
		}
	})
}

func TestBooleanThreeValuedLogic(t *testing.T) {
	t.Run("and truth table", func(t *testing.T) {
		// true and true = true
		result := And(types.TrueCollection, types.TrueCollection)
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true and true = true")
		}

		// true and false = false
		result = And(types.TrueCollection, types.FalseCollection)
		if result[0].(types.Boolean).Bool() {
			t.Error("expected true and false = false")
		}

		// false and true = false
		result = And(types.FalseCollection, types.TrueCollection)
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false and true = false")
		}

		// false and false = false
		result = And(types.FalseCollection, types.FalseCollection)
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false and false = false")
		}

		// empty and true = empty
		result = And(types.EmptyCollection, types.TrueCollection)
		if !result.Empty() {
			t.Error("expected empty and true = empty")
		}

		// true and empty = empty
		result = And(types.TrueCollection, types.EmptyCollection)
		if !result.Empty() {
			t.Error("expected true and empty = empty")
		}
	})

	t.Run("or truth table", func(t *testing.T) {
		// true or false = true
		result := Or(types.TrueCollection, types.FalseCollection)
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true or false = true")
		}

		// false or true = true
		result = Or(types.FalseCollection, types.TrueCollection)
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected false or true = true")
		}

		// false or false = false
		result = Or(types.FalseCollection, types.FalseCollection)
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false or false = false")
		}

		// empty or false = empty
		result = Or(types.EmptyCollection, types.FalseCollection)
		if !result.Empty() {
			t.Error("expected empty or false = empty")
		}

		// false or empty = empty
		result = Or(types.FalseCollection, types.EmptyCollection)
		if !result.Empty() {
			t.Error("expected false or empty = empty")
		}
	})

	t.Run("xor truth table", func(t *testing.T) {
		// true xor true = false
		result := Xor(types.TrueCollection, types.TrueCollection)
		if result[0].(types.Boolean).Bool() {
			t.Error("expected true xor true = false")
		}

		// false xor false = false
		result = Xor(types.FalseCollection, types.FalseCollection)
		if result[0].(types.Boolean).Bool() {
			t.Error("expected false xor false = false")
		}

		// true xor false = true
		result = Xor(types.TrueCollection, types.FalseCollection)
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true xor false = true")
		}

		// false xor true = true
		result = Xor(types.FalseCollection, types.TrueCollection)
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected false xor true = true")
		}
	})

	t.Run("implies truth table", func(t *testing.T) {
		// false implies anything = true
		result := Implies(types.FalseCollection, types.TrueCollection)
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected false implies true = true")
		}

		result = Implies(types.FalseCollection, types.FalseCollection)
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected false implies false = true")
		}

		// anything implies true = true
		result = Implies(types.TrueCollection, types.TrueCollection)
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected true implies true = true")
		}

		// true implies false = false
		result = Implies(types.TrueCollection, types.FalseCollection)
		if result[0].(types.Boolean).Bool() {
			t.Error("expected true implies false = false")
		}

		// empty implies true = true
		result = Implies(types.EmptyCollection, types.TrueCollection)
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected empty implies true = true")
		}

		// empty implies false = empty
		result = Implies(types.EmptyCollection, types.FalseCollection)
		if !result.Empty() {
			t.Error("expected empty implies false = empty")
		}
	})

	t.Run("not truth table", func(t *testing.T) {
		result := Not(types.TrueCollection)
		if result[0].(types.Boolean).Bool() {
			t.Error("expected not true = false")
		}

		result = Not(types.FalseCollection)
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected not false = true")
		}

		result = Not(types.EmptyCollection)
		if !result.Empty() {
			t.Error("expected not empty = empty")
		}
	})
}

func TestInContainsOperators(t *testing.T) {
	t.Run("in true cases", func(t *testing.T) {
		collection := types.Collection{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)}

		result := In(types.Collection{types.NewInteger(2)}, collection)
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected 2 in [1,2,3] = true")
		}
	})

	t.Run("in false cases", func(t *testing.T) {
		collection := types.Collection{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)}

		result := In(types.Collection{types.NewInteger(5)}, collection)
		if result[0].(types.Boolean).Bool() {
			t.Error("expected 5 in [1,2,3] = false")
		}
	})

	t.Run("contains true cases", func(t *testing.T) {
		collection := types.Collection{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)}

		result := Contains(collection, types.Collection{types.NewInteger(2)})
		if !result[0].(types.Boolean).Bool() {
			t.Error("expected [1,2,3] contains 2 = true")
		}
	})

	t.Run("contains false cases", func(t *testing.T) {
		collection := types.Collection{types.NewInteger(1), types.NewInteger(2), types.NewInteger(3)}

		result := Contains(collection, types.Collection{types.NewInteger(5)})
		if result[0].(types.Boolean).Bool() {
			t.Error("expected [1,2,3] contains 5 = false")
		}
	})
}

func TestTypeMatches(t *testing.T) {
	tests := []struct {
		name       string
		actualType string
		typeName   string
		expected   bool
	}{
		// Direct match
		{"direct match Boolean", "Boolean", "Boolean", true},
		{"direct match String", "String", "String", true},
		{"direct match Integer", "Integer", "Integer", true},
		{"direct match Decimal", "Decimal", "Decimal", true},
		{"direct match Date", "Date", "Date", true},
		{"direct match DateTime", "DateTime", "DateTime", true},
		{"direct match Time", "Time", "Time", true},
		{"direct match Quantity", "Quantity", "Quantity", true},

		// Case-insensitive match
		{"case insensitive boolean", "Boolean", "boolean", true},
		{"case insensitive string", "String", "string", true},
		{"case insensitive integer", "Integer", "integer", true},
		{"case insensitive decimal", "Decimal", "decimal", true},

		// FHIR primitive type mappings
		{"FHIR uri to String", "String", "uri", true},
		{"FHIR url to String", "String", "url", true},
		{"FHIR code to String", "String", "code", true},
		{"FHIR id to String", "String", "id", true},
		{"FHIR markdown to String", "String", "markdown", true},
		{"FHIR base64Binary to String", "String", "base64Binary", true},
		{"FHIR canonical to String", "String", "canonical", true},
		{"FHIR oid to String", "String", "oid", true},
		{"FHIR uuid to String", "String", "uuid", true},

		// FHIR integer variants
		{"FHIR positiveInt to Integer", "Integer", "positiveInt", true},
		{"FHIR unsignedInt to Integer", "Integer", "unsignedInt", true},
		{"FHIR integer64 to Integer", "Integer", "integer64", true},

		// FHIR DateTime variants
		{"FHIR instant to DateTime", "DateTime", "instant", true},

		// FHIR Quantity variants
		{"FHIR SimpleQuantity to Quantity", "Quantity", "SimpleQuantity", true},
		{"FHIR Age to Quantity", "Quantity", "Age", true},
		{"FHIR Count to Quantity", "Quantity", "Count", true},
		{"FHIR Distance to Quantity", "Quantity", "Distance", true},
		{"FHIR Duration to Quantity", "Quantity", "Duration", true},
		{"FHIR Money to Quantity", "Quantity", "Money", true},

		// System namespace
		{"System.Boolean", "Boolean", "System.Boolean", true},
		{"System.String", "String", "System.String", true},
		{"System.Integer", "Integer", "System.Integer", true},
		{"System.Decimal", "Decimal", "System.Decimal", true},

		// FHIR namespace
		{"FHIR.boolean", "Boolean", "FHIR.boolean", true},
		{"FHIR.string", "String", "FHIR.string", true},

		// Non-matches
		{"different types", "String", "Integer", false},
		{"different types 2", "Boolean", "Decimal", false},
		{"no match uri for Integer", "Integer", "uri", false},
		{"no match Date for String", "Date", "String", false},

		// Resource types
		{"Patient resource", "Patient", "Patient", true},
		{"Observation resource", "Observation", "Observation", true},

		// Resource and DomainResource base type inheritance
		{"Patient is Resource", "Patient", "Resource", true},
		{"Observation is Resource", "Observation", "Resource", true},
		{"Bundle is Resource", "Bundle", "Resource", true},
		{"Binary is Resource", "Binary", "Resource", true},
		{"Parameters is Resource", "Parameters", "Resource", true},

		{"Patient is DomainResource", "Patient", "DomainResource", true},
		{"Observation is DomainResource", "Observation", "DomainResource", true},
		{"MedicationRequest is DomainResource", "MedicationRequest", "DomainResource", true},

		// Bundle, Binary, Parameters inherit directly from Resource, NOT DomainResource
		{"Bundle is NOT DomainResource", "Bundle", "DomainResource", false},
		{"Binary is NOT DomainResource", "Binary", "DomainResource", false},
		{"Parameters is NOT DomainResource", "Parameters", "DomainResource", false},

		// Primitives are not resources
		{"String is not Resource", "String", "Resource", false},
		{"Boolean is not Resource", "Boolean", "Resource", false},
		{"Integer is not Resource", "Integer", "Resource", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TypeMatches(tt.actualType, tt.typeName)
			if result != tt.expected {
				t.Errorf("TypeMatches(%q, %q) = %v, expected %v",
					tt.actualType, tt.typeName, result, tt.expected)
			}
		})
	}
}

func TestIsSubtypeOf(t *testing.T) {
	tests := []struct {
		name       string
		actualType string
		baseType   string
		expected   bool
	}{
		// Direct matches
		{"Patient equals Patient", "Patient", "Patient", true},
		{"Resource equals Resource", "Resource", "Resource", true},
		{"DomainResource equals DomainResource", "DomainResource", "DomainResource", true},

		// All resources inherit from Resource
		{"Patient is Resource", "Patient", "Resource", true},
		{"Observation is Resource", "Observation", "Resource", true},
		{"Encounter is Resource", "Encounter", "Resource", true},
		{"Bundle is Resource", "Bundle", "Resource", true},
		{"Binary is Resource", "Binary", "Resource", true},
		{"Parameters is Resource", "Parameters", "Resource", true},

		// Most resources inherit from DomainResource
		{"Patient is DomainResource", "Patient", "DomainResource", true},
		{"Observation is DomainResource", "Observation", "DomainResource", true},
		{"Condition is DomainResource", "Condition", "DomainResource", true},

		// Bundle, Binary, Parameters do NOT inherit from DomainResource
		{"Bundle is NOT DomainResource", "Bundle", "DomainResource", false},
		{"Binary is NOT DomainResource", "Binary", "DomainResource", false},
		{"Parameters is NOT DomainResource", "Parameters", "DomainResource", false},

		// Primitives are not resources
		{"String is not Resource", "String", "Resource", false},
		{"Boolean is not Resource", "Boolean", "Resource", false},
		{"Integer is not Resource", "Integer", "Resource", false},
		{"Quantity is not Resource", "Quantity", "Resource", false},

		// Case insensitive for base types
		{"Patient is resource (lowercase)", "Patient", "resource", true},
		{"Patient is domainresource (lowercase)", "Patient", "domainresource", true},

		// Different concrete types don't match
		{"Patient is not Observation", "Patient", "Observation", false},
		{"Bundle is not Patient", "Bundle", "Patient", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSubtypeOf(tt.actualType, tt.baseType)
			if result != tt.expected {
				t.Errorf("IsSubtypeOf(%q, %q) = %v, expected %v",
					tt.actualType, tt.baseType, result, tt.expected)
			}
		})
	}
}

func TestIsDomainResource(t *testing.T) {
	// Resources that are NOT DomainResources
	nonDomainResources := []string{"Bundle", "Binary", "Parameters"}
	for _, rt := range nonDomainResources {
		if IsDomainResource(rt) {
			t.Errorf("IsDomainResource(%q) = true, expected false", rt)
		}
	}

	// Resources that ARE DomainResources
	domainResources := []string{"Patient", "Observation", "Encounter", "Condition", "MedicationRequest"}
	for _, rt := range domainResources {
		if !IsDomainResource(rt) {
			t.Errorf("IsDomainResource(%q) = false, expected true", rt)
		}
	}
}

func TestDateArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		date     string
		value    int
		unit     string
		expected string
		subtract bool
	}{
		// Date + years
		{"date plus 1 year", "2020-01-01", 1, "year", "2021-01-01", false},
		{"date plus 2 years", "2020-01-01", 2, "years", "2022-01-01", false},
		{"date plus years quoted", "2020-01-01", 1, "'year'", "2021-01-01", false},

		// Date + months
		{"date plus 1 month", "2020-01-15", 1, "month", "2020-02-15", false},
		{"date plus 6 months", "2020-01-15", 6, "months", "2020-07-15", false},
		{"date plus months crossing year", "2020-11-15", 3, "months", "2021-02-15", false},

		// Date + weeks
		{"date plus 1 week", "2020-01-01", 1, "week", "2020-01-08", false},
		{"date plus 2 weeks", "2020-01-01", 2, "weeks", "2020-01-15", false},

		// Date + days
		{"date plus 1 day", "2020-01-01", 1, "day", "2020-01-02", false},
		{"date plus 30 days", "2020-01-01", 30, "days", "2020-01-31", false},
		{"date plus days crossing month", "2020-01-31", 1, "day", "2020-02-01", false},

		// Date - durations
		{"date minus 1 year", "2020-01-01", 1, "year", "2019-01-01", true},
		{"date minus 6 months", "2020-07-15", 6, "months", "2020-01-15", true},
		{"date minus 1 week", "2020-01-08", 1, "week", "2020-01-01", true},
		{"date minus 1 day", "2020-01-02", 1, "day", "2020-01-01", true},

		// Leap year handling
		{"leap year add day", "2020-02-28", 1, "day", "2020-02-29", false},
		{"non-leap year add day", "2019-02-28", 1, "day", "2019-03-01", false},

		// Year-only precision
		{"year precision plus year", "2020", 1, "year", "2021", false},

		// Year-month precision
		{"year-month precision plus month", "2020-06", 1, "month", "2020-07", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := types.NewDate(tt.date)
			if err != nil {
				t.Fatalf("failed to create date: %v", err)
			}

			quantity := types.NewQuantityFromDecimal(
				types.NewDecimalFromInt(int64(tt.value)).Value(),
				tt.unit,
			)

			var result types.Value
			if tt.subtract {
				result, err = Subtract(date, quantity)
			} else {
				result, err = Add(date, quantity)
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultDate, ok := result.(types.Date)
			if !ok {
				t.Fatalf("expected Date, got %T", result)
			}

			if resultDate.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, resultDate.String())
			}
		})
	}
}

func TestDateTimeArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		datetime string
		value    int
		unit     string
		expected string
		subtract bool
	}{
		// DateTime + years/months/days
		{"datetime plus 1 year", "2020-01-01T10:00:00", 1, "year", "2021-01-01T10:00:00", false},
		{"datetime plus 1 month", "2020-01-15T10:00:00", 1, "month", "2020-02-15T10:00:00", false},
		{"datetime plus 1 day", "2020-01-01T10:00:00", 1, "day", "2020-01-02T10:00:00", false},

		// DateTime + hours/minutes/seconds
		{"datetime plus 1 hour", "2020-01-01T10:00:00", 1, "hour", "2020-01-01T11:00:00", false},
		{"datetime plus 30 minutes", "2020-01-01T10:00:00", 30, "minutes", "2020-01-01T10:30:00", false},
		{"datetime plus 45 seconds", "2020-01-01T10:00:00", 45, "seconds", "2020-01-01T10:00:45", false},

		// DateTime - durations
		{"datetime minus 1 hour", "2020-01-01T10:00:00", 1, "hour", "2020-01-01T09:00:00", true},
		{"datetime minus 30 minutes", "2020-01-01T10:30:00", 30, "minutes", "2020-01-01T10:00:00", true},

		// Crossing boundaries
		{"datetime plus hours crossing day", "2020-01-01T23:00:00", 2, "hours", "2020-01-02T01:00:00", false},
		{"datetime minus hours crossing day", "2020-01-02T01:00:00", 2, "hours", "2020-01-01T23:00:00", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt, err := types.NewDateTime(tt.datetime)
			if err != nil {
				t.Fatalf("failed to create datetime: %v", err)
			}

			quantity := types.NewQuantityFromDecimal(
				types.NewDecimalFromInt(int64(tt.value)).Value(),
				tt.unit,
			)

			var result types.Value
			if tt.subtract {
				result, err = Subtract(dt, quantity)
			} else {
				result, err = Add(dt, quantity)
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultDT, ok := result.(types.DateTime)
			if !ok {
				t.Fatalf("expected DateTime, got %T", result)
			}

			if resultDT.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, resultDT.String())
			}
		})
	}
}

func TestQuantityArithmetic(t *testing.T) {
	tests := []struct {
		name      string
		q1Value   int64
		q1Unit    string
		q2Value   int64
		q2Unit    string
		expected  string
		subtract  bool
		expectErr bool
	}{
		// Quantity + Quantity same unit
		{"quantity plus quantity same unit", 5, "mg", 3, "mg", "8 mg", false, false},
		{"quantity minus quantity same unit", 10, "kg", 3, "kg", "7 kg", true, false},

		// Quantity with empty unit
		{"quantity plus quantity empty unit", 5, "", 3, "", "8", false, false},

		// Incompatible units
		{"quantity plus incompatible units", 5, "mg", 3, "kg", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q1 := types.NewQuantityFromDecimal(
				types.NewDecimalFromInt(tt.q1Value).Value(),
				tt.q1Unit,
			)
			q2 := types.NewQuantityFromDecimal(
				types.NewDecimalFromInt(tt.q2Value).Value(),
				tt.q2Unit,
			)

			var result types.Value
			var err error
			if tt.subtract {
				result, err = Subtract(q1, q2)
			} else {
				result, err = Add(q1, q2)
			}

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			resultQ, ok := result.(types.Quantity)
			if !ok {
				t.Fatalf("expected Quantity, got %T", result)
			}

			if resultQ.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, resultQ.String())
			}
		})
	}
}

func TestQuantityComparison(t *testing.T) {
	tests := []struct {
		name      string
		q1Value   int64
		q1Unit    string
		q2Value   int64
		q2Unit    string
		op        string
		expected  bool
		expectErr bool
	}{
		// Greater than
		{"10 kg > 5 kg", 10, "kg", 5, "kg", ">", true, false},
		{"5 kg > 10 kg", 5, "kg", 10, "kg", ">", false, false},

		// Less than
		{"5 kg < 10 kg", 5, "kg", 10, "kg", "<", true, false},
		{"10 kg < 5 kg", 10, "kg", 5, "kg", "<", false, false},

		// Greater or equal
		{"10 kg >= 10 kg", 10, "kg", 10, "kg", ">=", true, false},
		{"10 kg >= 5 kg", 10, "kg", 5, "kg", ">=", true, false},
		{"5 kg >= 10 kg", 5, "kg", 10, "kg", ">=", false, false},

		// Less or equal
		{"10 kg <= 10 kg", 10, "kg", 10, "kg", "<=", true, false},
		{"5 kg <= 10 kg", 5, "kg", 10, "kg", "<=", true, false},
		{"10 kg <= 5 kg", 10, "kg", 5, "kg", "<=", false, false},

		// Empty units (compatible)
		{"10 > 5 (no unit)", 10, "", 5, "", ">", true, false},

		// Mixed empty and non-empty units (compatible)
		{"10 kg > 5 (empty)", 10, "kg", 5, "", ">", true, false},

		// UCUM compatible units (kg and mg are both mass)
		{"10 kg > 5 mg (UCUM)", 10, "kg", 5, "mg", ">", true, false},
		{"5 mg > 10 kg (UCUM)", 5, "mg", 10, "kg", ">", false, false},

		// Truly incompatible units (mass vs length)
		{"incompatible units error", 10, "kg", 5, "m", ">", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q1 := types.NewQuantityFromDecimal(
				types.NewDecimalFromInt(tt.q1Value).Value(),
				tt.q1Unit,
			)
			q2 := types.NewQuantityFromDecimal(
				types.NewDecimalFromInt(tt.q2Value).Value(),
				tt.q2Unit,
			)

			var result types.Collection
			var err error

			switch tt.op {
			case ">":
				result, err = GreaterThan(q1, q2)
			case "<":
				result, err = LessThan(q1, q2)
			case ">=":
				result, err = GreaterOrEqual(q1, q2)
			case "<=":
				result, err = LessOrEqual(q1, q2)
			}

			if tt.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Empty() {
				t.Fatalf("expected result, got empty collection")
			}

			resultBool, ok := result[0].(types.Boolean)
			if !ok {
				t.Fatalf("expected Boolean, got %T", result[0])
			}

			if resultBool.Bool() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, resultBool.Bool())
			}
		})
	}
}
