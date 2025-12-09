package types

import (
	"testing"
)

func TestBoolean(t *testing.T) {
	t.Run("creation and value", func(t *testing.T) {
		b := NewBoolean(true)
		if !b.Bool() {
			t.Error("expected true")
		}
		if b.Type() != "Boolean" {
			t.Errorf("expected Boolean, got %s", b.Type())
		}
	})

	t.Run("equality", func(t *testing.T) {
		b1 := NewBoolean(true)
		b2 := NewBoolean(true)
		b3 := NewBoolean(false)

		if !b1.Equal(b2) {
			t.Error("expected true == true")
		}
		if b1.Equal(b3) {
			t.Error("expected true != false")
		}
	})

	t.Run("not", func(t *testing.T) {
		b := NewBoolean(true)
		if b.Not().Bool() {
			t.Error("expected !true = false")
		}
	})
}

func TestString(t *testing.T) {
	t.Run("creation and value", func(t *testing.T) {
		s := NewString("hello")
		if s.Value() != "hello" {
			t.Errorf("expected hello, got %s", s.Value())
		}
		if s.Type() != "String" {
			t.Errorf("expected String, got %s", s.Type())
		}
	})

	t.Run("equality", func(t *testing.T) {
		s1 := NewString("hello")
		s2 := NewString("hello")
		s3 := NewString("world")

		if !s1.Equal(s2) {
			t.Error("expected hello == hello")
		}
		if s1.Equal(s3) {
			t.Error("expected hello != world")
		}
	})

	t.Run("equivalence", func(t *testing.T) {
		s1 := NewString("HELLO")
		s2 := NewString("hello")
		s3 := NewString("  hello  ")

		if !s1.Equivalent(s2) {
			t.Error("expected HELLO ~ hello")
		}
		if !s2.Equivalent(s3) {
			t.Error("expected hello ~ '  hello  '")
		}
	})

	t.Run("methods", func(t *testing.T) {
		s := NewString("Hello World")

		if s.Length() != 11 {
			t.Errorf("expected length 11, got %d", s.Length())
		}
		if !s.Contains("World") {
			t.Error("expected contains World")
		}
		if !s.StartsWith("Hello") {
			t.Error("expected starts with Hello")
		}
		if !s.EndsWith("World") {
			t.Error("expected ends with World")
		}
		if s.Upper().Value() != "HELLO WORLD" {
			t.Errorf("expected HELLO WORLD, got %s", s.Upper().Value())
		}
		if s.Lower().Value() != "hello world" {
			t.Errorf("expected hello world, got %s", s.Lower().Value())
		}
	})
}

func TestInteger(t *testing.T) {
	t.Run("creation and value", func(t *testing.T) {
		i := NewInteger(42)
		if i.Value() != 42 {
			t.Errorf("expected 42, got %d", i.Value())
		}
		if i.Type() != "Integer" {
			t.Errorf("expected Integer, got %s", i.Type())
		}
	})

	t.Run("equality", func(t *testing.T) {
		i1 := NewInteger(42)
		i2 := NewInteger(42)
		i3 := NewInteger(100)

		if !i1.Equal(i2) {
			t.Error("expected 42 == 42")
		}
		if i1.Equal(i3) {
			t.Error("expected 42 != 100")
		}
	})

	t.Run("arithmetic", func(t *testing.T) {
		i1 := NewInteger(10)
		i2 := NewInteger(3)

		if i1.Add(i2).Value() != 13 {
			t.Errorf("expected 10+3=13, got %d", i1.Add(i2).Value())
		}
		if i1.Subtract(i2).Value() != 7 {
			t.Errorf("expected 10-3=7, got %d", i1.Subtract(i2).Value())
		}
		if i1.Multiply(i2).Value() != 30 {
			t.Errorf("expected 10*3=30, got %d", i1.Multiply(i2).Value())
		}

		div, err := i1.Div(i2)
		if err != nil || div.Value() != 3 {
			t.Errorf("expected 10 div 3=3, got %d", div.Value())
		}

		mod, err := i1.Mod(i2)
		if err != nil || mod.Value() != 1 {
			t.Errorf("expected 10 mod 3=1, got %d", mod.Value())
		}
	})

	t.Run("comparison", func(t *testing.T) {
		i1 := NewInteger(10)
		i2 := NewInteger(20)

		cmp, _ := i1.Compare(i2)
		if cmp != -1 {
			t.Errorf("expected 10 < 20, got %d", cmp)
		}
	})
}

func TestDecimal(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		d, err := NewDecimal("3.14")
		if err != nil {
			t.Fatal(err)
		}
		if d.Type() != "Decimal" {
			t.Errorf("expected Decimal, got %s", d.Type())
		}
	})

	t.Run("precision", func(t *testing.T) {
		d1 := MustDecimal("0.1")
		d2 := MustDecimal("0.2")
		sum := d1.Add(d2)

		expected := MustDecimal("0.3")
		if !sum.Equal(expected) {
			t.Errorf("expected 0.1+0.2=0.3, got %s", sum.String())
		}
	})

	t.Run("arithmetic", func(t *testing.T) {
		d1 := MustDecimal("10.5")
		d2 := MustDecimal("3.5")

		if d1.Add(d2).String() != "14" {
			t.Errorf("expected 14, got %s", d1.Add(d2).String())
		}
		if d1.Subtract(d2).String() != "7" {
			t.Errorf("expected 7, got %s", d1.Subtract(d2).String())
		}
	})

	t.Run("rounding", func(t *testing.T) {
		d := MustDecimal("3.7")

		if d.Ceiling().Value() != 4 {
			t.Errorf("expected ceiling 4, got %d", d.Ceiling().Value())
		}
		if d.Floor().Value() != 3 {
			t.Errorf("expected floor 3, got %d", d.Floor().Value())
		}
	})

	t.Run("cross-type equality", func(t *testing.T) {
		d := MustDecimal("42")
		i := NewInteger(42)

		if !d.Equal(i) {
			t.Error("expected 42.0 == 42")
		}
		if !i.Equal(d) {
			t.Error("expected 42 == 42.0")
		}
	})
}

func TestCollection(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		c := Collection{}
		if !c.Empty() {
			t.Error("expected empty collection")
		}
		if c.Count() != 0 {
			t.Error("expected count 0")
		}
	})

	t.Run("first and last", func(t *testing.T) {
		c := Collection{NewInteger(1), NewInteger(2), NewInteger(3)}

		first, ok := c.First()
		if !ok || first.(Integer).Value() != 1 {
			t.Error("expected first = 1")
		}

		last, ok := c.Last()
		if !ok || last.(Integer).Value() != 3 {
			t.Error("expected last = 3")
		}
	})

	t.Run("single", func(t *testing.T) {
		c1 := Collection{NewInteger(42)}
		single, err := c1.Single()
		if err != nil || single.(Integer).Value() != 42 {
			t.Error("expected single = 42")
		}

		c2 := Collection{}
		_, err = c2.Single()
		if err == nil {
			t.Error("expected error for empty collection")
		}

		c3 := Collection{NewInteger(1), NewInteger(2)}
		_, err = c3.Single()
		if err == nil {
			t.Error("expected error for multiple elements")
		}
	})

	t.Run("skip and take", func(t *testing.T) {
		c := Collection{NewInteger(1), NewInteger(2), NewInteger(3), NewInteger(4), NewInteger(5)}

		skipped := c.Skip(2)
		if skipped.Count() != 3 {
			t.Errorf("expected 3 after skip, got %d", skipped.Count())
		}

		taken := c.Take(3)
		if taken.Count() != 3 {
			t.Errorf("expected 3 after take, got %d", taken.Count())
		}
	})

	t.Run("distinct", func(t *testing.T) {
		c := Collection{NewInteger(1), NewInteger(2), NewInteger(1), NewInteger(3), NewInteger(2)}
		distinct := c.Distinct()

		if distinct.Count() != 3 {
			t.Errorf("expected 3 distinct, got %d", distinct.Count())
		}
	})

	t.Run("union and intersect", func(t *testing.T) {
		c1 := Collection{NewInteger(1), NewInteger(2), NewInteger(3)}
		c2 := Collection{NewInteger(2), NewInteger(3), NewInteger(4)}

		union := c1.Union(c2)
		if union.Count() != 4 {
			t.Errorf("expected 4 in union, got %d", union.Count())
		}

		intersect := c1.Intersect(c2)
		if intersect.Count() != 2 {
			t.Errorf("expected 2 in intersect, got %d", intersect.Count())
		}
	})

	t.Run("boolean aggregation", func(t *testing.T) {
		c := Collection{NewBoolean(true), NewBoolean(true), NewBoolean(true)}
		if !c.AllTrue() {
			t.Error("expected all true")
		}

		c2 := Collection{NewBoolean(false), NewBoolean(true)}
		if !c2.AnyTrue() {
			t.Error("expected any true")
		}
		if !c2.AnyFalse() {
			t.Error("expected any false")
		}
	})
}

func TestObjectValue(t *testing.T) {
	t.Run("creation", func(t *testing.T) {
		json := []byte(`{"name": "John", "age": 30}`)
		obj := NewObjectValue(json)

		if obj.Type() != "Object" {
			t.Errorf("expected Object, got %s", obj.Type())
		}
	})

	t.Run("get field", func(t *testing.T) {
		json := []byte(`{"name": "John", "age": 30, "active": true}`)
		obj := NewObjectValue(json)

		name, ok := obj.Get("name")
		if !ok || name.(String).Value() != "John" {
			t.Error("expected name = John")
		}

		age, ok := obj.Get("age")
		if !ok || age.(Integer).Value() != 30 {
			t.Error("expected age = 30")
		}

		active, ok := obj.Get("active")
		if !ok || !active.(Boolean).Bool() {
			t.Error("expected active = true")
		}
	})

	t.Run("get collection", func(t *testing.T) {
		json := []byte(`{"items": [1, 2, 3]}`)
		obj := NewObjectValue(json)

		items := obj.GetCollection("items")
		if items.Count() != 3 {
			t.Errorf("expected 3 items, got %d", items.Count())
		}
	})

	t.Run("resourceType", func(t *testing.T) {
		json := []byte(`{"resourceType": "Patient", "id": "123"}`)
		obj := NewObjectValue(json)

		if obj.Type() != "Patient" {
			t.Errorf("expected Patient, got %s", obj.Type())
		}
	})

	t.Run("toQuantity with unit field", func(t *testing.T) {
		json := []byte(`{"value": 120, "unit": "mm[Hg]"}`)
		obj := NewObjectValue(json)

		q, ok := obj.ToQuantity()
		if !ok {
			t.Fatal("expected ToQuantity to succeed")
		}
		if q.Value().String() != "120" {
			t.Errorf("expected value 120, got %s", q.Value().String())
		}
		if q.Unit() != "mm[Hg]" {
			t.Errorf("expected unit mm[Hg], got %s", q.Unit())
		}
	})

	t.Run("toQuantity with code field", func(t *testing.T) {
		json := []byte(`{"value": 75.5, "code": "kg"}`)
		obj := NewObjectValue(json)

		q, ok := obj.ToQuantity()
		if !ok {
			t.Fatal("expected ToQuantity to succeed")
		}
		if q.Value().String() != "75.5" {
			t.Errorf("expected value 75.5, got %s", q.Value().String())
		}
		if q.Unit() != "kg" {
			t.Errorf("expected unit kg, got %s", q.Unit())
		}
	})

	t.Run("toQuantity with both unit and code", func(t *testing.T) {
		// unit takes precedence
		json := []byte(`{"value": 100, "unit": "mg", "code": "mg"}`)
		obj := NewObjectValue(json)

		q, ok := obj.ToQuantity()
		if !ok {
			t.Fatal("expected ToQuantity to succeed")
		}
		if q.Unit() != "mg" {
			t.Errorf("expected unit mg, got %s", q.Unit())
		}
	})

	t.Run("toQuantity without unit", func(t *testing.T) {
		json := []byte(`{"value": 42}`)
		obj := NewObjectValue(json)

		q, ok := obj.ToQuantity()
		if !ok {
			t.Fatal("expected ToQuantity to succeed")
		}
		if q.Value().String() != "42" {
			t.Errorf("expected value 42, got %s", q.Value().String())
		}
		if q.Unit() != "" {
			t.Errorf("expected empty unit, got %s", q.Unit())
		}
	})

	t.Run("toQuantity with decimal value", func(t *testing.T) {
		json := []byte(`{"value": 3.14159, "unit": "rad"}`)
		obj := NewObjectValue(json)

		q, ok := obj.ToQuantity()
		if !ok {
			t.Fatal("expected ToQuantity to succeed")
		}
		if q.Value().String() != "3.14159" {
			t.Errorf("expected value 3.14159, got %s", q.Value().String())
		}
	})

	t.Run("toQuantity fails without value field", func(t *testing.T) {
		json := []byte(`{"unit": "kg"}`)
		obj := NewObjectValue(json)

		_, ok := obj.ToQuantity()
		if ok {
			t.Error("expected ToQuantity to fail without value field")
		}
	})

	t.Run("toQuantity fails with non-numeric value", func(t *testing.T) {
		json := []byte(`{"value": "not a number", "unit": "kg"}`)
		obj := NewObjectValue(json)

		_, ok := obj.ToQuantity()
		if ok {
			t.Error("expected ToQuantity to fail with non-numeric value")
		}
	})

	t.Run("toQuantity fails with null value", func(t *testing.T) {
		json := []byte(`{"value": null, "unit": "kg"}`)
		obj := NewObjectValue(json)

		_, ok := obj.ToQuantity()
		if ok {
			t.Error("expected ToQuantity to fail with null value")
		}
	})

	t.Run("toQuantity FHIR Quantity example", func(t *testing.T) {
		// Full FHIR Quantity structure
		json := []byte(`{
			"value": 6.3,
			"unit": "mmol/l",
			"system": "http://unitsofmeasure.org",
			"code": "mmol/L"
		}`)
		obj := NewObjectValue(json)

		q, ok := obj.ToQuantity()
		if !ok {
			t.Fatal("expected ToQuantity to succeed")
		}
		if q.Value().String() != "6.3" {
			t.Errorf("expected value 6.3, got %s", q.Value().String())
		}
		// unit field takes precedence over code
		if q.Unit() != "mmol/l" {
			t.Errorf("expected unit mmol/l, got %s", q.Unit())
		}
	})

	t.Run("toQuantity comparison", func(t *testing.T) {
		json := []byte(`{"value": 120, "unit": "mm[Hg]"}`)
		obj := NewObjectValue(json)

		q, ok := obj.ToQuantity()
		if !ok {
			t.Fatal("expected ToQuantity to succeed")
		}

		// Compare with a FHIRPath Quantity
		other, _ := NewQuantity("90 mm[Hg]")
		cmp, err := q.Compare(other)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != 1 {
			t.Error("expected 120 mm[Hg] > 90 mm[Hg]")
		}
	})
}

func TestJSONToCollection(t *testing.T) {
	t.Run("object", func(t *testing.T) {
		json := []byte(`{"name": "John"}`)
		c, err := JSONToCollection(json)
		if err != nil {
			t.Fatal(err)
		}
		if c.Count() != 1 {
			t.Errorf("expected 1 element, got %d", c.Count())
		}
	})

	t.Run("array", func(t *testing.T) {
		json := []byte(`[1, 2, 3]`)
		c, err := JSONToCollection(json)
		if err != nil {
			t.Fatal(err)
		}
		if c.Count() != 3 {
			t.Errorf("expected 3 elements, got %d", c.Count())
		}
	})

	t.Run("null", func(t *testing.T) {
		json := []byte(`null`)
		c, err := JSONToCollection(json)
		if err != nil {
			t.Fatal(err)
		}
		if !c.Empty() {
			t.Error("expected empty collection for null")
		}
	})

	t.Run("primitive", func(t *testing.T) {
		json := []byte(`42`)
		c, err := JSONToCollection(json)
		if err != nil {
			t.Fatal(err)
		}
		if c.Count() != 1 || c[0].(Integer).Value() != 42 {
			t.Error("expected single integer 42")
		}
	})
}

func TestPoolOptimizations(t *testing.T) {
	t.Run("GetBoolean cache", func(t *testing.T) {
		t1 := GetBoolean(true)
		t2 := GetBoolean(true)
		if t1 != t2 {
			t.Error("GetBoolean should return same instance")
		}

		f1 := GetBoolean(false)
		f2 := GetBoolean(false)
		if f1 != f2 {
			t.Error("GetBoolean should return same instance for false")
		}
	})

	t.Run("GetInteger cache range", func(t *testing.T) {
		// Cached range [-128, 127]
		i1 := GetInteger(42)
		i2 := GetInteger(42)
		if i1 != i2 {
			t.Error("GetInteger should return same instance for cached values")
		}

		// Negative cached
		n1 := GetInteger(-100)
		n2 := GetInteger(-100)
		if n1 != n2 {
			t.Error("GetInteger should cache negative values too")
		}

		// Outside cache
		big := GetInteger(1000)
		if big.Value() != 1000 {
			t.Errorf("expected 1000, got %d", big.Value())
		}
	})

	t.Run("cached collections", func(t *testing.T) {
		if TrueCollection.Empty() {
			t.Error("TrueCollection should not be empty")
		}
		if !TrueCollection[0].(Boolean).Bool() {
			t.Error("TrueCollection should contain true")
		}

		if FalseCollection.Empty() {
			t.Error("FalseCollection should not be empty")
		}
		if FalseCollection[0].(Boolean).Bool() {
			t.Error("FalseCollection should contain false")
		}

		if !EmptyCollection.Empty() {
			t.Error("EmptyCollection should be empty")
		}
	})

	t.Run("collection pool", func(t *testing.T) {
		c := GetCollection()
		if c == nil {
			t.Fatal("GetCollection should return non-nil")
		}
		*c = append(*c, NewInteger(1))
		PutCollection(c)

		// Get again
		c2 := GetCollection()
		if c2 == nil {
			t.Fatal("GetCollection should return non-nil")
		}
		// Should be empty after put
		if len(*c2) != 0 {
			t.Error("Collection from pool should be empty")
		}
	})

	t.Run("NewCollectionWithCap", func(t *testing.T) {
		c := NewCollectionWithCap(10)
		if cap(c) < 10 {
			t.Errorf("expected capacity >= 10, got %d", cap(c))
		}
	})

	t.Run("SingletonCollection", func(t *testing.T) {
		c := SingletonCollection(NewInteger(42))
		if c.Count() != 1 {
			t.Errorf("expected 1 element, got %d", c.Count())
		}
		if c[0].(Integer).Value() != 42 {
			t.Error("expected 42")
		}
	})
}

func TestBooleanEdgeCases(t *testing.T) {
	t.Run("string representation", func(t *testing.T) {
		if NewBoolean(true).String() != "true" {
			t.Error("expected 'true'")
		}
		if NewBoolean(false).String() != "false" {
			t.Error("expected 'false'")
		}
	})

	t.Run("isEmpty", func(t *testing.T) {
		if NewBoolean(true).IsEmpty() {
			t.Error("boolean should not be empty")
		}
	})

	t.Run("equivalence", func(t *testing.T) {
		if !NewBoolean(true).Equivalent(NewBoolean(true)) {
			t.Error("true should be equivalent to true")
		}
		if NewBoolean(true).Equivalent(NewBoolean(false)) {
			t.Error("true should not be equivalent to false")
		}
	})
}

func TestStringEdgeCases(t *testing.T) {
	t.Run("isEmpty", func(t *testing.T) {
		if NewString("hello").IsEmpty() {
			t.Error("non-empty string should not be empty")
		}
	})

	t.Run("compare", func(t *testing.T) {
		s1 := NewString("apple")
		s2 := NewString("banana")

		cmp, err := s1.Compare(s2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp >= 0 {
			t.Error("apple should be less than banana")
		}

		cmp, err = s2.Compare(s1)
		if err != nil {
			t.Fatal(err)
		}
		if cmp <= 0 {
			t.Error("banana should be greater than apple")
		}
	})

	t.Run("string methods", func(t *testing.T) {
		replaced := NewString("hello").Replace("l", "L")
		if replaced.Value() != "heLLo" {
			t.Errorf("expected 'heLLo', got '%s'", replaced.Value())
		}

		sub := NewString("hello").Substring(1, 3)
		if sub.Value() != "ell" {
			t.Errorf("expected 'ell', got '%s'", sub.Value())
		}
	})
}

func TestIntegerEdgeCases(t *testing.T) {
	t.Run("isEmpty", func(t *testing.T) {
		if NewInteger(0).IsEmpty() {
			t.Error("integer should not be empty")
		}
	})

	t.Run("negate", func(t *testing.T) {
		i := NewInteger(42)
		neg := i.Negate()
		if neg.Value() != -42 {
			t.Errorf("expected -42, got %d", neg.Value())
		}

		// Negate negative
		neg2 := neg.Negate()
		if neg2.Value() != 42 {
			t.Errorf("expected 42, got %d", neg2.Value())
		}
	})

	t.Run("toDecimal", func(t *testing.T) {
		i := NewInteger(42)
		d := i.ToDecimal()
		if d.Type() != "Decimal" {
			t.Errorf("expected Decimal, got %s", d.Type())
		}
	})

	t.Run("equivalence", func(t *testing.T) {
		if !NewInteger(42).Equivalent(NewInteger(42)) {
			t.Error("42 should be equivalent to 42")
		}
	})
}

func TestDecimalEdgeCases(t *testing.T) {
	t.Run("isEmpty", func(t *testing.T) {
		d := NewDecimalFromFloat(3.14)
		if d.IsEmpty() {
			t.Error("decimal should not be empty")
		}
	})

	t.Run("negate", func(t *testing.T) {
		d := NewDecimalFromFloat(3.14)
		neg := d.Negate()
		if neg.Value().InexactFloat64() != -3.14 {
			t.Errorf("expected -3.14, got %v", neg.Value())
		}
	})

	t.Run("abs", func(t *testing.T) {
		d := NewDecimalFromFloat(-3.14)
		abs := d.Abs()
		if abs.Value().InexactFloat64() != 3.14 {
			t.Errorf("expected 3.14, got %v", abs.Value())
		}
	})

	t.Run("truncate", func(t *testing.T) {
		d := NewDecimalFromFloat(3.99)
		tr := d.Truncate()
		if tr.Value() != 3 {
			t.Errorf("expected 3, got %d", tr.Value())
		}
	})

	t.Run("equivalence", func(t *testing.T) {
		d1 := NewDecimalFromFloat(42.0)
		d2 := NewDecimalFromFloat(42.0)
		if !d1.Equivalent(d2) {
			t.Error("same decimals should be equivalent")
		}
	})
}

func TestCollectionEdgeCases(t *testing.T) {
	t.Run("tail of empty", func(t *testing.T) {
		c := Collection{}
		tail := c.Tail()
		if !tail.Empty() {
			t.Error("tail of empty should be empty")
		}
	})

	t.Run("skip edge cases", func(t *testing.T) {
		c := Collection{NewInteger(1), NewInteger(2)}

		// Skip more than count
		skipped := c.Skip(10)
		if !skipped.Empty() {
			t.Error("skip(10) on 2 elements should be empty")
		}

		// Skip 0
		skipped = c.Skip(0)
		if skipped.Count() != 2 {
			t.Errorf("skip(0) should return all elements")
		}
	})

	t.Run("take edge cases", func(t *testing.T) {
		c := Collection{NewInteger(1), NewInteger(2)}

		// Take more than count
		taken := c.Take(10)
		if taken.Count() != 2 {
			t.Errorf("take(10) on 2 elements should return 2")
		}

		// Take 0
		taken = c.Take(0)
		if !taken.Empty() {
			t.Error("take(0) should be empty")
		}
	})

	t.Run("isDistinct", func(t *testing.T) {
		distinct := Collection{NewInteger(1), NewInteger(2)}
		if !distinct.IsDistinct() {
			t.Error("expected distinct")
		}

		notDistinct := Collection{NewInteger(1), NewInteger(1)}
		if notDistinct.IsDistinct() {
			t.Error("expected not distinct")
		}
	})

	t.Run("exclude", func(t *testing.T) {
		c1 := Collection{NewInteger(1), NewInteger(2), NewInteger(3)}
		c2 := Collection{NewInteger(2)}
		excluded := c1.Exclude(c2)
		if excluded.Count() != 2 {
			t.Errorf("expected 2 after exclude, got %d", excluded.Count())
		}
	})

	t.Run("combine with duplicates", func(t *testing.T) {
		c1 := Collection{NewInteger(1)}
		c2 := Collection{NewInteger(1)}
		combined := c1.Combine(c2)
		if combined.Count() != 2 {
			t.Errorf("combine should keep duplicates, got %d", combined.Count())
		}
	})

	t.Run("allFalse/anyFalse", func(t *testing.T) {
		allFalse := Collection{NewBoolean(false), NewBoolean(false)}
		if !allFalse.AllFalse() {
			t.Error("expected allFalse")
		}

		mixed := Collection{NewBoolean(true), NewBoolean(false)}
		if !mixed.AnyFalse() {
			t.Error("expected anyFalse")
		}
	})

	t.Run("toBoolean errors", func(t *testing.T) {
		// Multiple elements
		_, err := Collection{NewBoolean(true), NewBoolean(true)}.ToBoolean()
		if err == nil {
			t.Error("expected error for multiple elements")
		}

		// Non-boolean
		_, err = Collection{NewInteger(1)}.ToBoolean()
		if err == nil {
			t.Error("expected error for non-boolean")
		}
	})
}
