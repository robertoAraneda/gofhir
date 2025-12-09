package types

import (
	"testing"
	"time"
)

func TestDate(t *testing.T) {
	t.Run("full date", func(t *testing.T) {
		d, err := NewDate("2024-01-15")
		if err != nil {
			t.Fatal(err)
		}
		if d.Year() != 2024 {
			t.Errorf("expected year 2024, got %d", d.Year())
		}
		if d.Month() != 1 {
			t.Errorf("expected month 1, got %d", d.Month())
		}
		if d.Day() != 15 {
			t.Errorf("expected day 15, got %d", d.Day())
		}
		if d.Type() != "Date" {
			t.Errorf("expected Date, got %s", d.Type())
		}
		if d.String() != "2024-01-15" {
			t.Errorf("expected 2024-01-15, got %s", d.String())
		}
	})

	t.Run("year-month only", func(t *testing.T) {
		d, err := NewDate("2024-06")
		if err != nil {
			t.Fatal(err)
		}
		if d.Year() != 2024 || d.Month() != 6 || d.Day() != 0 {
			t.Errorf("unexpected values: %d-%d-%d", d.Year(), d.Month(), d.Day())
		}
		if d.Precision() != MonthPrecision {
			t.Error("expected month precision")
		}
		if d.String() != "2024-06" {
			t.Errorf("expected 2024-06, got %s", d.String())
		}
	})

	t.Run("year only", func(t *testing.T) {
		d, err := NewDate("2024")
		if err != nil {
			t.Fatal(err)
		}
		if d.Year() != 2024 {
			t.Errorf("expected year 2024, got %d", d.Year())
		}
		if d.Precision() != YearPrecision {
			t.Error("expected year precision")
		}
		if d.String() != "2024" {
			t.Errorf("expected 2024, got %s", d.String())
		}
	})

	t.Run("invalid date", func(t *testing.T) {
		_, err := NewDate("invalid")
		if err == nil {
			t.Error("expected error for invalid date")
		}
	})

	t.Run("equality", func(t *testing.T) {
		d1, _ := NewDate("2024-01-15")
		d2, _ := NewDate("2024-01-15")
		d3, _ := NewDate("2024-01-16")

		if !d1.Equal(d2) {
			t.Error("expected equal dates")
		}
		if d1.Equal(d3) {
			t.Error("expected different dates")
		}
	})

	t.Run("compare", func(t *testing.T) {
		d1, _ := NewDate("2024-01-15")
		d2, _ := NewDate("2024-01-20")

		cmp, err := d1.Compare(d2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected d1 < d2")
		}
		cmp, err = d2.Compare(d1)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != 1 {
			t.Error("expected d2 > d1")
		}
		d1Copy, _ := NewDate("2024-01-15")
		cmp, err = d1.Compare(d1Copy)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != 0 {
			t.Error("expected d1 = d1Copy")
		}
	})

	t.Run("compare same precision - year", func(t *testing.T) {
		d1, _ := NewDate("2024")
		d2, _ := NewDate("2025")

		cmp, err := d1.Compare(d2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected 2024 < 2025")
		}
	})

	t.Run("compare same precision - month", func(t *testing.T) {
		d1, _ := NewDate("2024-01")
		d2, _ := NewDate("2024-06")

		cmp, err := d1.Compare(d2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected 2024-01 < 2024-06")
		}
	})

	t.Run("compare different precision - different years", func(t *testing.T) {
		d1, _ := NewDate("2024")
		d2, _ := NewDate("2025-06-15")

		cmp, err := d1.Compare(d2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected 2024 < 2025-06-15")
		}
	})

	t.Run("compare different precision - same year ambiguous", func(t *testing.T) {
		d1, _ := NewDate("2024")
		d2, _ := NewDate("2024-06-15")

		_, err := d1.Compare(d2)
		if err == nil {
			t.Error("expected ambiguous comparison error")
		}
	})

	t.Run("compare different precision - month vs day ambiguous", func(t *testing.T) {
		d1, _ := NewDate("2024-06")
		d2, _ := NewDate("2024-06-15")

		_, err := d1.Compare(d2)
		if err == nil {
			t.Error("expected ambiguous comparison error")
		}
	})

	t.Run("compare different precision - different months", func(t *testing.T) {
		d1, _ := NewDate("2024-05")
		d2, _ := NewDate("2024-06-15")

		cmp, err := d1.Compare(d2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected 2024-05 < 2024-06-15")
		}
	})

	t.Run("compare with non-Date type", func(t *testing.T) {
		d1, _ := NewDate("2024-01-15")

		_, err := d1.Compare(NewInteger(42))
		if err == nil {
			t.Error("expected error when comparing Date with Integer")
		}
	})

	t.Run("from time.Time", func(t *testing.T) {
		tm := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
		d := NewDateFromTime(tm)

		if d.Year() != 2024 || d.Month() != 3 || d.Day() != 15 {
			t.Errorf("unexpected values: %d-%d-%d", d.Year(), d.Month(), d.Day())
		}
	})

	t.Run("toTime", func(t *testing.T) {
		d, _ := NewDate("2024-01-15")
		tm := d.ToTime()

		if tm.Year() != 2024 || tm.Month() != 1 || tm.Day() != 15 {
			t.Errorf("unexpected time: %v", tm)
		}
	})
}

func TestDateTime(t *testing.T) {
	t.Run("full datetime with timezone", func(t *testing.T) {
		dt, err := NewDateTime("2024-01-15T10:30:45.123Z")
		if err != nil {
			t.Fatal(err)
		}
		if dt.Year() != 2024 {
			t.Errorf("expected year 2024, got %d", dt.Year())
		}
		if dt.Month() != 1 {
			t.Errorf("expected month 1, got %d", dt.Month())
		}
		if dt.Day() != 15 {
			t.Errorf("expected day 15, got %d", dt.Day())
		}
		if dt.Hour() != 10 {
			t.Errorf("expected hour 10, got %d", dt.Hour())
		}
		if dt.Minute() != 30 {
			t.Errorf("expected minute 30, got %d", dt.Minute())
		}
		if dt.Second() != 45 {
			t.Errorf("expected second 45, got %d", dt.Second())
		}
		if dt.Millisecond() != 123 {
			t.Errorf("expected millisecond 123, got %d", dt.Millisecond())
		}
		if dt.Type() != "DateTime" {
			t.Errorf("expected DateTime, got %s", dt.Type())
		}
	})

	t.Run("with offset", func(t *testing.T) {
		dt, err := NewDateTime("2024-01-15T10:30:00+05:30")
		if err != nil {
			t.Fatal(err)
		}
		if dt.Hour() != 10 || dt.Minute() != 30 {
			t.Errorf("unexpected time: %d:%d", dt.Hour(), dt.Minute())
		}
	})

	t.Run("date only", func(t *testing.T) {
		dt, err := NewDateTime("2024-01-15")
		if err != nil {
			t.Fatal(err)
		}
		if dt.Year() != 2024 || dt.Month() != 1 || dt.Day() != 15 {
			t.Errorf("unexpected date: %d-%d-%d", dt.Year(), dt.Month(), dt.Day())
		}
	})

	t.Run("invalid datetime", func(t *testing.T) {
		_, err := NewDateTime("invalid")
		if err == nil {
			t.Error("expected error for invalid datetime")
		}
	})

	t.Run("equality", func(t *testing.T) {
		dt1, _ := NewDateTime("2024-01-15T10:30:00Z")
		dt2, _ := NewDateTime("2024-01-15T10:30:00Z")
		dt3, _ := NewDateTime("2024-01-15T10:31:00Z")

		if !dt1.Equal(dt2) {
			t.Error("expected equal datetimes")
		}
		if dt1.Equal(dt3) {
			t.Error("expected different datetimes")
		}
	})

	t.Run("from time.Time", func(t *testing.T) {
		tm := time.Date(2024, 3, 15, 10, 30, 45, 123000000, time.UTC)
		dt := NewDateTimeFromTime(tm)

		if dt.Year() != 2024 || dt.Hour() != 10 || dt.Millisecond() != 123 {
			t.Errorf("unexpected datetime: %v", dt)
		}
	})

	t.Run("compare same precision", func(t *testing.T) {
		dt1, _ := NewDateTime("2024-01-15T10:30:00Z")
		dt2, _ := NewDateTime("2024-01-15T10:31:00Z")

		cmp, err := dt1.Compare(dt2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected dt1 < dt2")
		}

		cmp, err = dt2.Compare(dt1)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != 1 {
			t.Error("expected dt2 > dt1")
		}

		dt1Copy, _ := NewDateTime("2024-01-15T10:30:00Z")
		cmp, err = dt1.Compare(dt1Copy)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != 0 {
			t.Error("expected dt1 = dt1Copy")
		}
	})

	t.Run("compare same precision - year only", func(t *testing.T) {
		dt1, _ := NewDateTime("2024")
		dt2, _ := NewDateTime("2025")

		cmp, err := dt1.Compare(dt2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected 2024 < 2025")
		}
	})

	t.Run("compare same precision - with milliseconds", func(t *testing.T) {
		dt1, _ := NewDateTime("2024-01-15T10:30:45.100Z")
		dt2, _ := NewDateTime("2024-01-15T10:30:45.200Z")

		cmp, err := dt1.Compare(dt2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected .100 < .200")
		}
	})

	t.Run("compare different precision - different years", func(t *testing.T) {
		dt1, _ := NewDateTime("2024")
		dt2, _ := NewDateTime("2025-06-15T10:30:00Z")

		cmp, err := dt1.Compare(dt2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected 2024 < 2025-06-15T10:30:00Z")
		}
	})

	t.Run("compare different precision - same year ambiguous", func(t *testing.T) {
		dt1, _ := NewDateTime("2024")
		dt2, _ := NewDateTime("2024-06-15T10:30:00Z")

		_, err := dt1.Compare(dt2)
		if err == nil {
			t.Error("expected ambiguous comparison error")
		}
	})

	t.Run("compare different precision - different months", func(t *testing.T) {
		dt1, _ := NewDateTime("2024-05")
		dt2, _ := NewDateTime("2024-06-15T10:30:00Z")

		cmp, err := dt1.Compare(dt2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected 2024-05 < 2024-06-15T10:30:00Z")
		}
	})

	t.Run("compare different precision - same month ambiguous", func(t *testing.T) {
		dt1, _ := NewDateTime("2024-06")
		dt2, _ := NewDateTime("2024-06-15T10:30:00Z")

		_, err := dt1.Compare(dt2)
		if err == nil {
			t.Error("expected ambiguous comparison error")
		}
	})

	t.Run("compare different precision - different days", func(t *testing.T) {
		dt1, _ := NewDateTime("2024-06-10")
		dt2, _ := NewDateTime("2024-06-15T10:30:00Z")

		cmp, err := dt1.Compare(dt2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected 2024-06-10 < 2024-06-15T10:30:00Z")
		}
	})

	t.Run("compare different precision - same day ambiguous", func(t *testing.T) {
		dt1, _ := NewDateTime("2024-06-15")
		dt2, _ := NewDateTime("2024-06-15T10:30:00Z")

		_, err := dt1.Compare(dt2)
		if err == nil {
			t.Error("expected ambiguous comparison error")
		}
	})

	t.Run("compare with non-DateTime type", func(t *testing.T) {
		dt1, _ := NewDateTime("2024-01-15T10:30:00Z")

		_, err := dt1.Compare(NewInteger(42))
		if err == nil {
			t.Error("expected error when comparing DateTime with Integer")
		}
	})

	t.Run("compare with timezone handling", func(t *testing.T) {
		// Same instant in different timezones
		dt1, _ := NewDateTime("2024-01-15T10:00:00Z")
		dt2, _ := NewDateTime("2024-01-15T15:00:00+05:00")

		cmp, err := dt1.Compare(dt2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != 0 {
			t.Error("expected equal times in different timezones")
		}
	})
}

func TestTime(t *testing.T) {
	t.Run("full time", func(t *testing.T) {
		tm, err := NewTime("10:30:45.123")
		if err != nil {
			t.Fatal(err)
		}
		if tm.Hour() != 10 {
			t.Errorf("expected hour 10, got %d", tm.Hour())
		}
		if tm.Minute() != 30 {
			t.Errorf("expected minute 30, got %d", tm.Minute())
		}
		if tm.Second() != 45 {
			t.Errorf("expected second 45, got %d", tm.Second())
		}
		if tm.Millisecond() != 123 {
			t.Errorf("expected millisecond 123, got %d", tm.Millisecond())
		}
		if tm.Type() != "Time" {
			t.Errorf("expected Time, got %s", tm.Type())
		}
	})

	t.Run("with T prefix", func(t *testing.T) {
		tm, err := NewTime("T14:30:00")
		if err != nil {
			t.Fatal(err)
		}
		if tm.Hour() != 14 {
			t.Errorf("expected hour 14, got %d", tm.Hour())
		}
	})

	t.Run("hour and minute only", func(t *testing.T) {
		tm, err := NewTime("10:30")
		if err != nil {
			t.Fatal(err)
		}
		if tm.Hour() != 10 || tm.Minute() != 30 {
			t.Errorf("unexpected time: %d:%d", tm.Hour(), tm.Minute())
		}
	})

	t.Run("invalid time", func(t *testing.T) {
		_, err := NewTime("invalid")
		if err == nil {
			t.Error("expected error for invalid time")
		}
	})

	t.Run("equality", func(t *testing.T) {
		t1, _ := NewTime("10:30:45")
		t2, _ := NewTime("10:30:45")
		t3, _ := NewTime("10:30:46")

		if !t1.Equal(t2) {
			t.Error("expected equal times")
		}
		if t1.Equal(t3) {
			t.Error("expected different times")
		}
	})

	t.Run("compare", func(t *testing.T) {
		t1, _ := NewTime("10:30:00")
		t2, _ := NewTime("10:31:00")

		cmp, err := t1.Compare(t2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected t1 < t2")
		}
		cmp, err = t2.Compare(t1)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != 1 {
			t.Error("expected t2 > t1")
		}
		t1Copy, _ := NewTime("10:30:00")
		cmp, err = t1.Compare(t1Copy)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != 0 {
			t.Error("expected t1 = t1Copy")
		}
	})

	t.Run("compare same precision - hour", func(t *testing.T) {
		t1, _ := NewTime("10")
		t2, _ := NewTime("14")

		cmp, err := t1.Compare(t2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected 10 < 14")
		}
	})

	t.Run("compare same precision - minute", func(t *testing.T) {
		t1, _ := NewTime("10:30")
		t2, _ := NewTime("10:45")

		cmp, err := t1.Compare(t2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected 10:30 < 10:45")
		}
	})

	t.Run("compare same precision - milliseconds", func(t *testing.T) {
		t1, _ := NewTime("10:30:45.100")
		t2, _ := NewTime("10:30:45.200")

		cmp, err := t1.Compare(t2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected .100 < .200")
		}
	})

	t.Run("compare different precision - different hours", func(t *testing.T) {
		t1, _ := NewTime("10")
		t2, _ := NewTime("14:30:45")

		cmp, err := t1.Compare(t2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected 10 < 14:30:45")
		}
	})

	t.Run("compare different precision - same hour ambiguous", func(t *testing.T) {
		t1, _ := NewTime("10")
		t2, _ := NewTime("10:30:45")

		_, err := t1.Compare(t2)
		if err == nil {
			t.Error("expected ambiguous comparison error")
		}
	})

	t.Run("compare different precision - different minutes", func(t *testing.T) {
		t1, _ := NewTime("10:30")
		t2, _ := NewTime("10:45:30")

		cmp, err := t1.Compare(t2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected 10:30 < 10:45:30")
		}
	})

	t.Run("compare different precision - same minute ambiguous", func(t *testing.T) {
		t1, _ := NewTime("10:30")
		t2, _ := NewTime("10:30:45")

		_, err := t1.Compare(t2)
		if err == nil {
			t.Error("expected ambiguous comparison error")
		}
	})

	t.Run("compare different precision - second vs millisecond ambiguous", func(t *testing.T) {
		t1, _ := NewTime("10:30:45")
		t2, _ := NewTime("10:30:45.100")

		_, err := t1.Compare(t2)
		if err == nil {
			t.Error("expected ambiguous comparison error")
		}
	})

	t.Run("compare with non-Time type", func(t *testing.T) {
		t1, _ := NewTime("10:30:00")

		_, err := t1.Compare(NewInteger(42))
		if err == nil {
			t.Error("expected error when comparing Time with Integer")
		}
	})

	t.Run("from time.Time", func(t *testing.T) {
		tm := time.Date(2024, 1, 1, 10, 30, 45, 123000000, time.UTC)
		ft := NewTimeFromGoTime(tm)

		if ft.Hour() != 10 || ft.Minute() != 30 || ft.Second() != 45 {
			t.Errorf("unexpected time: %v", ft)
		}
	})
}

func TestQuantity(t *testing.T) {
	t.Run("with unit", func(t *testing.T) {
		q, err := NewQuantity("10 kg")
		if err != nil {
			t.Fatal(err)
		}
		if q.Value().String() != "10" {
			t.Errorf("expected value 10, got %s", q.Value().String())
		}
		if q.Unit() != "kg" {
			t.Errorf("expected unit kg, got %s", q.Unit())
		}
		if q.Type() != "Quantity" {
			t.Errorf("expected Quantity, got %s", q.Type())
		}
	})

	t.Run("with quoted unit", func(t *testing.T) {
		q, err := NewQuantity("5.5 'kg/m2'")
		if err != nil {
			t.Fatal(err)
		}
		if q.Unit() != "kg/m2" {
			t.Errorf("expected unit kg/m2, got %s", q.Unit())
		}
	})

	t.Run("without unit", func(t *testing.T) {
		q, err := NewQuantity("42")
		if err != nil {
			t.Fatal(err)
		}
		if q.Value().String() != "42" {
			t.Errorf("expected value 42, got %s", q.Value().String())
		}
		if q.Unit() != "" {
			t.Errorf("expected empty unit, got %s", q.Unit())
		}
	})

	t.Run("decimal value", func(t *testing.T) {
		q, err := NewQuantity("3.14159 rad")
		if err != nil {
			t.Fatal(err)
		}
		if q.Value().String() != "3.14159" {
			t.Errorf("expected 3.14159, got %s", q.Value().String())
		}
	})

	t.Run("invalid quantity", func(t *testing.T) {
		_, err := NewQuantity("invalid")
		if err == nil {
			t.Error("expected error for invalid quantity")
		}
	})

	t.Run("equality", func(t *testing.T) {
		q1, _ := NewQuantity("10 kg")
		q2, _ := NewQuantity("10 kg")
		q3, _ := NewQuantity("10 lb")

		if !q1.Equal(q2) {
			t.Error("expected equal quantities")
		}
		if q1.Equal(q3) {
			t.Error("expected different quantities")
		}
	})

	t.Run("equivalence", func(t *testing.T) {
		q1, _ := NewQuantity("10 kg")
		q2, _ := NewQuantity("10 KG")
		q3, _ := NewQuantity("10")

		if !q1.Equivalent(q2) {
			t.Error("expected equivalent quantities (case insensitive)")
		}
		if !q1.Equivalent(q3) {
			t.Error("expected equivalent with empty unit")
		}
	})

	t.Run("arithmetic", func(t *testing.T) {
		q1, _ := NewQuantity("10 kg")
		q2, _ := NewQuantity("5 kg")

		sum, err := q1.Add(q2)
		if err != nil {
			t.Fatal(err)
		}
		if sum.Value().String() != "15" {
			t.Errorf("expected 15, got %s", sum.Value().String())
		}

		diff, err := q1.Subtract(q2)
		if err != nil {
			t.Fatal(err)
		}
		if diff.Value().String() != "5" {
			t.Errorf("expected 5, got %s", diff.Value().String())
		}
	})

	t.Run("incompatible units", func(t *testing.T) {
		q1, _ := NewQuantity("10 kg")
		q2, _ := NewQuantity("5 m")

		_, err := q1.Add(q2)
		if err == nil {
			t.Error("expected error for incompatible units")
		}
	})

	t.Run("compare", func(t *testing.T) {
		q1, _ := NewQuantity("10 kg")
		q2, _ := NewQuantity("20 kg")

		cmp, err := q1.Compare(q2)
		if err != nil {
			t.Fatal(err)
		}
		if cmp != -1 {
			t.Error("expected q1 < q2")
		}
	})

	t.Run("string representation", func(t *testing.T) {
		q1, _ := NewQuantity("10 kg")
		if q1.String() != "10 kg" {
			t.Errorf("expected '10 kg', got '%s'", q1.String())
		}

		q2, _ := NewQuantity("5")
		if q2.String() != "5" {
			t.Errorf("expected '5', got '%s'", q2.String())
		}
	})
}
