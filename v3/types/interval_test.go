package types

import (
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
		result := val.(time.Time)
		if result.Year() != 2026 {
			t.Errorf("expected year=2026, got %d", result.Year())
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
		result := val.(time.Time)
		if result.Hour() != 2 || result.Minute() != 30 || result.Second() != 15 {
			t.Errorf("expected 2:30:15, got %d:%d:%d",
				result.Hour(), result.Minute(), result.Second())
		}
	})
}
