package eval

import (
	"github.com/robertoaraneda/gofhir/pkg/fhirpath/types"
)

// Arithmetic operators

// Add performs addition on two values.
func Add(left, right types.Value) (types.Value, error) {
	switch l := left.(type) {
	case types.Integer:
		switch r := right.(type) {
		case types.Integer:
			return l.Add(r), nil
		case types.Decimal:
			return l.ToDecimal().Add(r), nil
		}
	case types.Decimal:
		switch r := right.(type) {
		case types.Integer:
			return l.Add(r.ToDecimal()), nil
		case types.Decimal:
			return l.Add(r), nil
		}
	case types.String:
		if r, ok := right.(types.String); ok {
			return types.NewString(l.Value() + r.Value()), nil
		}
	case types.Date:
		if q, ok := right.(types.Quantity); ok {
			// Date + Quantity (duration)
			value := int(q.Value().IntPart())
			return l.AddDuration(value, q.Unit()), nil
		}
	case types.DateTime:
		if q, ok := right.(types.Quantity); ok {
			// DateTime + Quantity (duration)
			value := int(q.Value().IntPart())
			return l.AddDuration(value, q.Unit()), nil
		}
	case types.Quantity:
		switch r := right.(type) {
		case types.Quantity:
			// Quantity + Quantity
			return l.Add(r)
		}
	}
	return nil, InvalidOperationError("+", left.Type(), right.Type())
}

// Subtract performs subtraction on two values.
func Subtract(left, right types.Value) (types.Value, error) {
	switch l := left.(type) {
	case types.Integer:
		switch r := right.(type) {
		case types.Integer:
			return l.Subtract(r), nil
		case types.Decimal:
			return l.ToDecimal().Subtract(r), nil
		}
	case types.Decimal:
		switch r := right.(type) {
		case types.Integer:
			return l.Subtract(r.ToDecimal()), nil
		case types.Decimal:
			return l.Subtract(r), nil
		}
	case types.Date:
		if q, ok := right.(types.Quantity); ok {
			// Date - Quantity (duration)
			value := int(q.Value().IntPart())
			return l.SubtractDuration(value, q.Unit()), nil
		}
	case types.DateTime:
		if q, ok := right.(types.Quantity); ok {
			// DateTime - Quantity (duration)
			value := int(q.Value().IntPart())
			return l.SubtractDuration(value, q.Unit()), nil
		}
	case types.Quantity:
		switch r := right.(type) {
		case types.Quantity:
			// Quantity - Quantity
			return l.Subtract(r)
		}
	}
	return nil, InvalidOperationError("-", left.Type(), right.Type())
}

// Multiply performs multiplication on two values.
func Multiply(left, right types.Value) (types.Value, error) {
	switch l := left.(type) {
	case types.Integer:
		switch r := right.(type) {
		case types.Integer:
			return l.Multiply(r), nil
		case types.Decimal:
			return l.ToDecimal().Multiply(r), nil
		}
	case types.Decimal:
		switch r := right.(type) {
		case types.Integer:
			return l.Multiply(r.ToDecimal()), nil
		case types.Decimal:
			return l.Multiply(r), nil
		}
	}
	return nil, InvalidOperationError("*", left.Type(), right.Type())
}

// Divide performs division on two values.
func Divide(left, right types.Value) (types.Value, error) {
	// Convert both to Decimal for division
	var lDec, rDec types.Decimal
	switch l := left.(type) {
	case types.Integer:
		lDec = l.ToDecimal()
	case types.Decimal:
		lDec = l
	default:
		return nil, InvalidOperationError("/", left.Type(), right.Type())
	}

	switch r := right.(type) {
	case types.Integer:
		rDec = r.ToDecimal()
	case types.Decimal:
		rDec = r
	default:
		return nil, InvalidOperationError("/", left.Type(), right.Type())
	}

	return lDec.Divide(rDec)
}

// IntegerDivide performs integer division (div operator).
func IntegerDivide(left, right types.Value) (types.Value, error) {
	l, ok := left.(types.Integer)
	if !ok {
		return nil, InvalidOperationError("div", left.Type(), right.Type())
	}
	r, ok := right.(types.Integer)
	if !ok {
		return nil, InvalidOperationError("div", left.Type(), right.Type())
	}
	return l.Div(r)
}

// Modulo performs modulo operation (mod operator).
func Modulo(left, right types.Value) (types.Value, error) {
	l, ok := left.(types.Integer)
	if !ok {
		return nil, InvalidOperationError("mod", left.Type(), right.Type())
	}
	r, ok := right.(types.Integer)
	if !ok {
		return nil, InvalidOperationError("mod", left.Type(), right.Type())
	}
	return l.Mod(r)
}

// Negate negates a numeric value.
func Negate(value types.Value) (types.Value, error) {
	switch v := value.(type) {
	case types.Integer:
		return v.Negate(), nil
	case types.Decimal:
		return v.Negate(), nil
	}
	return nil, NewEvalError(ErrType, "cannot negate "+value.Type())
}

// Comparison operators

// Compare compares two values and returns -1, 0, or 1.
func Compare(left, right types.Value) (int, error) {
	// Try to convert ObjectValue to Quantity if comparing with Quantity
	if obj, ok := left.(*types.ObjectValue); ok {
		if _, isRightQuantity := right.(types.Quantity); isRightQuantity {
			if q, ok := obj.ToQuantity(); ok {
				return q.Compare(right)
			}
		}
	}
	if obj, ok := right.(*types.ObjectValue); ok {
		if _, isLeftQuantity := left.(types.Quantity); isLeftQuantity {
			if q, ok := obj.ToQuantity(); ok {
				if comp, ok := left.(types.Comparable); ok {
					return comp.Compare(q)
				}
			}
		}
	}

	if comp, ok := left.(types.Comparable); ok {
		return comp.Compare(right)
	}
	return 0, InvalidOperationError("compare", left.Type(), right.Type())
}

// LessThan returns true if left < right.
func LessThan(left, right types.Value) (types.Collection, error) {
	cmp, err := Compare(left, right)
	if err != nil {
		return nil, err
	}
	if cmp < 0 {
		return types.TrueCollection, nil
	}
	return types.FalseCollection, nil
}

// LessOrEqual returns true if left <= right.
func LessOrEqual(left, right types.Value) (types.Collection, error) {
	cmp, err := Compare(left, right)
	if err != nil {
		return nil, err
	}
	if cmp <= 0 {
		return types.TrueCollection, nil
	}
	return types.FalseCollection, nil
}

// GreaterThan returns true if left > right.
func GreaterThan(left, right types.Value) (types.Collection, error) {
	cmp, err := Compare(left, right)
	if err != nil {
		return nil, err
	}
	if cmp > 0 {
		return types.TrueCollection, nil
	}
	return types.FalseCollection, nil
}

// GreaterOrEqual returns true if left >= right.
func GreaterOrEqual(left, right types.Value) (types.Collection, error) {
	cmp, err := Compare(left, right)
	if err != nil {
		return nil, err
	}
	if cmp >= 0 {
		return types.TrueCollection, nil
	}
	return types.FalseCollection, nil
}

// Equality operators

// Equal returns true if left = right.
func Equal(left, right types.Collection) types.Collection {
	// Empty propagation
	if left.Empty() || right.Empty() {
		return types.EmptyCollection
	}

	// Both must be singletons
	if len(left) != 1 || len(right) != 1 {
		return types.EmptyCollection
	}

	if left[0].Equal(right[0]) {
		return types.TrueCollection
	}
	return types.FalseCollection
}

// NotEqual returns true if left != right.
func NotEqual(left, right types.Collection) types.Collection {
	result := Equal(left, right)
	if result.Empty() {
		return result
	}
	if result[0].(types.Boolean).Bool() {
		return types.FalseCollection
	}
	return types.TrueCollection
}

// Equivalent returns true if left ~ right.
func Equivalent(left, right types.Collection) types.Collection {
	// For equivalence, empty collections are equivalent to each other
	if left.Empty() && right.Empty() {
		return types.TrueCollection
	}
	if left.Empty() || right.Empty() {
		return types.FalseCollection
	}

	// Both must be singletons
	if len(left) != 1 || len(right) != 1 {
		return types.FalseCollection
	}

	if left[0].Equivalent(right[0]) {
		return types.TrueCollection
	}
	return types.FalseCollection
}

// NotEquivalent returns true if left !~ right.
func NotEquivalent(left, right types.Collection) types.Collection {
	result := Equivalent(left, right)
	if result[0].(types.Boolean).Bool() {
		return types.FalseCollection
	}
	return types.TrueCollection
}

// Boolean operators (three-valued logic)

// And performs logical AND with three-valued logic.
func And(left, right types.Collection) types.Collection {
	lEmpty := left.Empty()
	rEmpty := right.Empty()

	// If either is false, result is false
	if !lEmpty {
		if b, ok := left[0].(types.Boolean); ok && !b.Bool() {
			return types.FalseCollection
		}
	}
	if !rEmpty {
		if b, ok := right[0].(types.Boolean); ok && !b.Bool() {
			return types.FalseCollection
		}
	}

	// If either is empty, propagate empty
	if lEmpty || rEmpty {
		return types.EmptyCollection
	}

	// Both must be true
	lBool, lOk := left[0].(types.Boolean)
	rBool, rOk := right[0].(types.Boolean)
	if !lOk || !rOk {
		return types.EmptyCollection
	}

	if lBool.Bool() && rBool.Bool() {
		return types.TrueCollection
	}
	return types.FalseCollection
}

// Or performs logical OR with three-valued logic.
func Or(left, right types.Collection) types.Collection {
	lEmpty := left.Empty()
	rEmpty := right.Empty()

	// If either is true, result is true
	if !lEmpty {
		if b, ok := left[0].(types.Boolean); ok && b.Bool() {
			return types.TrueCollection
		}
	}
	if !rEmpty {
		if b, ok := right[0].(types.Boolean); ok && b.Bool() {
			return types.TrueCollection
		}
	}

	// If either is empty, propagate empty
	if lEmpty || rEmpty {
		return types.EmptyCollection
	}

	// Both must be false
	lBool, lOk := left[0].(types.Boolean)
	rBool, rOk := right[0].(types.Boolean)
	if !lOk || !rOk {
		return types.EmptyCollection
	}

	if lBool.Bool() || rBool.Bool() {
		return types.TrueCollection
	}
	return types.FalseCollection
}

// Xor performs logical XOR.
func Xor(left, right types.Collection) types.Collection {
	if left.Empty() || right.Empty() {
		return types.EmptyCollection
	}

	lBool, lOk := left[0].(types.Boolean)
	rBool, rOk := right[0].(types.Boolean)
	if !lOk || !rOk {
		return types.EmptyCollection
	}

	if lBool.Bool() != rBool.Bool() {
		return types.TrueCollection
	}
	return types.FalseCollection
}

// Implies performs logical implication.
func Implies(left, right types.Collection) types.Collection {
	lEmpty := left.Empty()
	rEmpty := right.Empty()

	// If left is false, result is true
	if !lEmpty {
		if b, ok := left[0].(types.Boolean); ok && !b.Bool() {
			return types.TrueCollection
		}
	}

	// If right is true, result is true
	if !rEmpty {
		if b, ok := right[0].(types.Boolean); ok && b.Bool() {
			return types.TrueCollection
		}
	}

	// If either is empty, propagate empty
	if lEmpty || rEmpty {
		return types.EmptyCollection
	}

	// left is true and right is false
	return types.FalseCollection
}

// Not performs logical NOT.
func Not(value types.Collection) types.Collection {
	if value.Empty() {
		return types.EmptyCollection
	}
	if len(value) != 1 {
		return types.EmptyCollection
	}
	if b, ok := value[0].(types.Boolean); ok {
		if b.Bool() {
			return types.FalseCollection
		}
		return types.TrueCollection
	}
	return types.EmptyCollection
}

// String operators

// Concatenate performs string concatenation (& operator).
// Unlike +, & treats empty as empty string.
func Concatenate(left, right types.Collection) types.Collection {
	var lStr, rStr string

	if !left.Empty() {
		if s, ok := left[0].(types.String); ok {
			lStr = s.Value()
		}
	}

	if !right.Empty() {
		if s, ok := right[0].(types.String); ok {
			rStr = s.Value()
		}
	}

	return types.Collection{types.NewString(lStr + rStr)}
}

// Collection operators

// Union returns the union of two collections.
func Union(left, right types.Collection) types.Collection {
	return left.Union(right)
}

// In checks if left is in right collection.
func In(left, right types.Collection) types.Collection {
	if left.Empty() {
		return types.EmptyCollection
	}
	if len(left) != 1 {
		return types.EmptyCollection
	}
	if right.Contains(left[0]) {
		return types.TrueCollection
	}
	return types.FalseCollection
}

// Contains checks if left collection contains right.
func Contains(left, right types.Collection) types.Collection {
	if right.Empty() {
		return types.EmptyCollection
	}
	if len(right) != 1 {
		return types.EmptyCollection
	}
	if left.Contains(right[0]) {
		return types.TrueCollection
	}
	return types.FalseCollection
}
