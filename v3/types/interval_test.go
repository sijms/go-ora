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

// TestIntervalText verifies the text Oracle writes for each interval, from the
// bytes Oracle sends (#785), a negative one among them.
func TestIntervalText(t *testing.T) {
	cases := map[string]struct {
		dataType uint16
		raw      []byte
		want     string
	}{
		"year to month":            {INTERVALYM_DTY, []byte{128, 0, 7, 229, 70}, "+2021-10"},
		"negative year to month":   {INTERVALYM_DTY, []byte{127, 255, 248, 27, 50}, "-2021-10"},
		"negative months only":     {INTERVALYM_DTY, []byte{128, 0, 0, 0, 57}, "-00-03"},
		"day to second":            {INTERVALDS_DTY, []byte{128, 0, 0, 2, 72, 83, 94, 155, 46, 2, 0}, "+02 12:23:34.456000"},
		"negative day to second":   {INTERVALDS_DTY, []byte{127, 255, 255, 254, 48, 37, 26, 100, 197, 243, 248}, "-02 12:23:34.456789"},
		"negative hours only":      {INTERVALDS_DTY, []byte{128, 0, 0, 0, 50, 40, 30, 128, 0, 0, 0}, "-00 10:20:30.000000"},
		"nanoseconds":              {INTERVALDS_DTY, []byte{128, 0, 0, 0, 60, 60, 61, 128, 0, 0, 1}, "+00 00:00:01.000000001"},
		"day to second, untyped":   {0, []byte{128, 0, 0, 4, 65, 66, 67, 128, 0, 0, 0}, "+04 05:06:07.000000"},
		"year to month, untyped":   {0, []byte{128, 0, 0, 2, 63}, "+02-03"},
		"day to second, by its id": {IntervalDS, []byte{128, 0, 0, 4, 65, 66, 67, 128, 0, 0, 0}, "+04 05:06:07.000000"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			interval := &Interval{}
			interval.SetDataType(c.dataType)
			interval.SetBytes(c.raw)
			got, err := interval.Text()
			if err != nil {
				t.Fatal(err)
			}
			if got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
	null := &Interval{}
	if got, err := null.Text(); got != nil || err != nil {
		t.Errorf("NULL: got %v, %v", got, err)
	}
}
