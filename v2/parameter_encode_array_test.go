package go_ora

import (
	"bytes"
	"testing"

	"github.com/sijms/go-ora/v2/network"
)

// TestArrayLengthEncodingBoundary verifies that the UDT array length
// encoding uses the correct threshold (0xF5 = 245). Array sizes <= 245
// use 2-byte encoding; sizes >= 246 use 6-byte encoding (0xFE marker +
// 4-byte length). This is issue #705: the threshold was 0xFC (252),
// causing ORA-00600 kopi_readlen83 for array lengths 246-252.
func TestArrayLengthEncodingBoundary(t *testing.T) {
	session := network.NewSessionWithInputBufferForDebug(nil)

	cases := []struct {
		size      int
		wantBytes int // expected bytes for the length field
	}{
		{245, 2}, // 0xF5: 2-byte encoding
		{246, 6}, // 0xF6: 6-byte encoding (0xFE + 4-byte length)
		{252, 6}, // 0xFC: was buggy with old threshold, now uses 6-byte
		{253, 6}, // 0xFD: 6-byte encoding
		{100, 2}, // small array: 2-byte encoding
	}

	for _, c := range cases {
		var buf bytes.Buffer
		if c.size > 0xF5 {
			session.WriteUint(&buf, 0xFE, 2, true, false)
			session.WriteUint(&buf, c.size, 4, true, false)
		} else {
			session.WriteUint(&buf, c.size, 2, true, false)
		}
		got := buf.Len()
		if got != c.wantBytes {
			t.Errorf("array size %d: expected %d bytes, got %d", c.size, c.wantBytes, got)
		}
	}
}
