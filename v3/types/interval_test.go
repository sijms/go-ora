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
		binary.BigEndian.PutUint32(b[7:], uint32(0x80000000+0))
		interval := &Interval{}
		interval.SetDataType(INTERVALDS_DTY)
		interval.SetBytes(b)
		val, err := interval.Value()
		if err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if val != "+04 02:30:15.000000" {
			t.Errorf("expected +04 02:30:15.000000, got %v", val)
		}
	})
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
