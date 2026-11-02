package types

import (
	"database/sql"
	"encoding/binary"
	"testing"
	"time"
)

// TestIntervalEncode verifies that Interval.encode produces the correct bytes
// with the 0x80000000 offset. This guards against the integer overflow on
// 32-bit platforms reported in issue #738.
func TestIntervalEncode(t *testing.T) {
	t.Run("YearMonth", func(t *testing.T) {
		interval := &Interval{}
		interval.SetDataType(INTERVALYM_DTY)
		input := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		if err := interval.SetValue(input); err != nil {
			t.Fatalf("encode failed: %v", err)
		}
		b := interval.Bytes()
		if len(b) != 5 {
			t.Fatalf("expected 5 bytes, got %d", len(b))
		}
		year := int64(binary.BigEndian.Uint32(b)) - 0x80000000
		if year != 2026 {
			t.Errorf("expected year=2026, got %d", year)
		}
		month := int(b[4]) - 60
		if month != 5 {
			t.Errorf("expected month=5, got %d", month)
		}
	})

	t.Run("DaySecond", func(t *testing.T) {
		interval := &Interval{}
		interval.SetDataType(INTERVALDS_DTY)
		input := time.Date(0, 0, 4, 2, 30, 15, 0, time.UTC)
		if err := interval.SetValue(input); err != nil {
			t.Fatalf("encode failed: %v", err)
		}
		b := interval.Bytes()
		if len(b) != 11 {
			t.Fatalf("expected 11 bytes, got %d", len(b))
		}
		day := int64(binary.BigEndian.Uint32(b)) - 0x80000000
		if day != 4 {
			t.Errorf("expected day=4, got %d", day)
		}
		if int(b[4])-60 != 2 {
			t.Errorf("expected hour=2, got %d", int(b[4])-60)
		}
		if int(b[5])-60 != 30 {
			t.Errorf("expected minute=30, got %d", int(b[5])-60)
		}
		if int(b[6])-60 != 15 {
			t.Errorf("expected second=15, got %d", int(b[6])-60)
		}
	})
}

// TestIntervalDecode verifies that Interval.Value decodes the 0x80000000
// offset correctly on all platforms (issue #738).
func TestIntervalDecode(t *testing.T) {
	t.Run("YearMonth", func(t *testing.T) {
		b := make([]byte, 5)
		binary.BigEndian.PutUint32(b, uint32(0x80000000+2026))
		b[4] = 60 + 5
		interval := &Interval{}
		interval.SetDataType(INTERVALYM_DTY)
		interval.SetBytes(b)
		val, err := interval.Value()
		if err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if val != "+2026-05" {
			t.Errorf("expected +2026-05, got %v", val)
		}
	})

	t.Run("DaySecond", func(t *testing.T) {
		b := make([]byte, 11)
		binary.BigEndian.PutUint32(b, uint32(0x80000000+4))
		b[4] = 60 + 2
		b[5] = 60 + 30
		b[6] = 60 + 15
		// The wire carries nanoseconds; INTERVAL DAY TO SECOND(9) must keep all nine digits.
		binary.BigEndian.PutUint32(b[7:], uint32(0x80000000+123456789))
		interval := &Interval{}
		interval.SetDataType(INTERVALDS_DTY)
		interval.SetBytes(b)
		val, err := interval.Value()
		if err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if val != "+04 02:30:15.123456789" {
			t.Errorf("expected +04 02:30:15.123456789, got %v", val)
		}
	})
}

// TestIntervalColumnPrecision checks that the leading and fractional precision of the
// column decide the width, matching TO_CHAR on Oracle 23:
//
//	CAST(INTERVAL '2 12:23:34.123456789' DAY TO SECOND(9) AS INTERVAL DAY(2) TO SECOND(0)) -> +02 12:23:34
//	... AS INTERVAL DAY(3) TO SECOND(3)  -> +002 12:23:34.123
//	... AS INTERVAL DAY(2) TO SECOND(9)  -> +02 12:23:34.123456789
//	CAST(INTERVAL '-2 12:23:34.5' DAY TO SECOND(9) AS INTERVAL DAY(9) TO SECOND(9)) -> -000000002 12:23:34.500000000
//	CAST(INTERVAL '21-10' YEAR TO MONTH AS INTERVAL YEAR(4) TO MONTH) -> +0021-10
func TestIntervalColumnPrecision(t *testing.T) {
	ds := func(days int32, hour, minute, second byte, ns int32) []byte {
		b := make([]byte, 11)
		binary.BigEndian.PutUint32(b, uint32(int64(days)+0x80000000))
		b[4], b[5], b[6] = hour, minute, second
		binary.BigEndian.PutUint32(b[7:], uint32(int64(ns)+0x80000000))
		return b
	}
	ym := make([]byte, 5)
	binary.BigEndian.PutUint32(ym, uint32(0x80000000+21))
	ym[4] = 60 + 10
	cases := []struct {
		name             string
		dataType         uint16
		precision, scale uint8
		raw              []byte
		want             string
	}{
		{"DAY(2) TO SECOND(0)", INTERVALDS_DTY, 2, 0, ds(2, 72, 83, 94, 0), "+02 12:23:34"},
		{"DAY(3) TO SECOND(3)", INTERVALDS_DTY, 3, 3, ds(2, 72, 83, 94, 123000000), "+002 12:23:34.123"},
		{"DAY(2) TO SECOND(9)", INTERVALDS_DTY, 2, 9, ds(2, 72, 83, 94, 123456789), "+02 12:23:34.123456789"},
		{"DAY(9) TO SECOND(9) negative", INTERVALDS_DTY, 9, 9, ds(-2, 48, 37, 26, -500000000), "-000000002 12:23:34.500000000"},
		{"YEAR(4) TO MONTH", INTERVALYM_DTY, 4, 0, ym, "+0021-10"},
		// An expression such as CAST(... AS INTERVAL DAY TO SECOND(6)) still carries nine digits; Oracle rounds them.
		{"rounds half up", INTERVALDS_DTY, 2, 6, ds(2, 72, 83, 94, 123456789), "+02 12:23:34.123457"},
		{"rounds negative into the second", INTERVALDS_DTY, 2, 0, ds(-2, 48, 37, 26, -500000000), "-02 12:23:35"},
		{"carries into the next day", INTERVALDS_DTY, 2, 0, ds(1, 83, 119, 119, 600000000), "+02 00:00:00"},
		// Without column metadata (output parameters, UDT attributes) nothing below a microsecond is dropped.
		{"unknown, nanoseconds", INTERVALDS_DTY, 0, 0, ds(2, 72, 83, 94, 123456789), "+02 12:23:34.123456789"},
		{"unknown, microseconds", INTERVALDS_DTY, 0, 0, ds(2, 72, 83, 94, 456000000), "+02 12:23:34.456000"},
		// parameter.go sets Scale to 0xFF when the describe reports -127; Oracle allows 0-9 fraction digits.
		{"scale 0xFF is unknown", INTERVALDS_DTY, 2, 0xFF, ds(2, 72, 83, 94, 123456789), "+02 12:23:34.123456789"},
		{"scale 12 is unknown", INTERVALDS_DTY, 2, 12, ds(2, 72, 83, 94, 456000000), "+02 12:23:34.456000"},
	}
	for _, c := range cases {
		interval := &Interval{}
		interval.SetDataType(c.dataType)
		interval.SetPrecisionScale(c.precision, c.scale)
		interval.SetBytes(c.raw)
		got, err := interval.Value()
		if err != nil || got != c.want {
			t.Errorf("%s: expected %q and got %v (%v)", c.name, c.want, got, err)
		}
	}
}

func TestIntervalCopyTo(t *testing.T) {
	interval := &Interval{}
	interval.SetDataType(INTERVALYM_DTY)
	interval.SetBytes([]byte{127, 255, 248, 27, 50})
	var str string
	if err := interval.CopyTo(&str); err != nil || str != "-2021-10" {
		t.Errorf("*string: expected -2021-10 and got %q (%v)", str, err)
	}
	var nullStr sql.NullString
	if err := interval.CopyTo(&nullStr); err != nil || nullStr != (sql.NullString{String: "-2021-10", Valid: true}) {
		t.Errorf("*sql.NullString: expected {-2021-10 true} and got %v (%v)", nullStr, err)
	}
	var tm time.Time
	if err := interval.CopyTo(&tm); err == nil {
		t.Errorf("*time.Time: expected an error because an interval is not a point in time, got %v", tm)
	}
	interval.SetBytes(nil)
	if err := interval.CopyTo(&nullStr); err != nil || nullStr.Valid {
		t.Errorf("null into *sql.NullString: expected invalid and got %v (%v)", nullStr, err)
	}
}
