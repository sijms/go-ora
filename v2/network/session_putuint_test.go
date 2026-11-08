package network

import (
	"bytes"
	"testing"
)

// A compressed integer is a length byte followed by exactly that many bytes.
// Stmt.write sends the re-execute iteration count this way; an array bind of
// 65,536 rows or more needs three bytes, so the field must be written with a
// 4-byte cap (UB4), as the other execute fields are.
func TestPutUintCompressedUB4(t *testing.T) {
	cases := []struct {
		n    int
		want []byte
	}{
		{0, []byte{0}},
		{1, []byte{1, 1}},
		{0xFFFF, []byte{2, 0xFF, 0xFF}},
		{0x10000, []byte{3, 1, 0, 0}},
		{0x40020, []byte{3, 4, 0, 0x20}},
		{100000, []byte{3, 0x01, 0x86, 0xA0}},
	}
	for _, c := range cases {
		s := NewSessionWithInputBufferForDebug(nil)
		s.PutUint(c.n, 4, true, true)
		got := s.outBuffer.Bytes()
		if !bytes.Equal(got, c.want) {
			t.Errorf("PutUint(%d, 4) = % x, want % x", c.n, got, c.want)
		}
		if int(got[0]) != len(got)-1 {
			t.Errorf("PutUint(%d, 4): length byte %d, %d bytes follow", c.n, got[0], len(got)-1)
		}
	}
}
