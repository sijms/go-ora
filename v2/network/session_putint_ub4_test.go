package network

import (
	"bytes"
	"testing"
)

func putIntBytes(n int, size uint8) []byte {
	s := NewSessionWithInputBufferForDebug(nil)
	s.PutInt(n, size, true, true)
	return append([]byte(nil), s.outBuffer.Bytes()...)
}

// checkUB4 checks that a compressed UB4 has a length byte equal to the number
// of bytes that follow, reads back as n, and is byte-identical to the 2-byte
// cap encoding for values up to 0xFFFF.
func checkUB4(t *testing.T, n int, got []byte) {
	t.Helper()
	if len(got) == 0 || int(got[0]) != len(got)-1 {
		t.Fatalf("PutInt(%d, 4) = % x: length byte does not match the bytes that follow", n, got)
	}
	if n <= 0xFFFF {
		if old := putIntBytes(n, 2); !bytes.Equal(got, old) {
			t.Fatalf("PutInt(%d): cap 4 = % x, cap 2 = % x, want equal", n, got, old)
		}
	}
	back, err := NewSessionWithInputBufferForDebug(got).GetInt(4, true, true)
	if err != nil || back != n {
		t.Fatalf("GetInt(PutInt(%d, 4)) = %d, %v", n, back, err)
	}
}

func TestPutIntCompressedUB4(t *testing.T) {
	cases := []struct {
		n    int
		want []byte
	}{
		{0, []byte{0}},
		{1, []byte{1, 1}},
		{25, []byte{1, 25}},
		{0x100, []byte{2, 1, 0}},
		{43691, []byte{2, 0xaa, 0xab}},
		{0xFFFF, []byte{2, 0xFF, 0xFF}},
		{0x10000, []byte{3, 1, 0, 0}},
		{65537, []byte{3, 1, 0, 1}},
		{131073, []byte{3, 2, 0, 1}},
		{0x7FFFFFFF, []byte{4, 0x7F, 0xFF, 0xFF, 0xFF}},
	}
	for _, c := range cases {
		got := putIntBytes(c.n, 4)
		if !bytes.Equal(got, c.want) {
			t.Errorf("PutInt(%d, 4) = % x, want % x", c.n, got, c.want)
			continue
		}
		checkUB4(t, c.n, got)
	}
}

func FuzzPutIntCompressedUB4(f *testing.F) {
	for _, n := range []uint32{0, 1, 25, 0xFFFF, 0x10000, 65537, 131073, 0x7FFFFFFF} {
		f.Add(n)
	}
	f.Fuzz(func(t *testing.T, u uint32) {
		n := int(u & 0x7FFFFFFF) // stay non-negative where int is 32 bits
		checkUB4(t, n, putIntBytes(n, 4))
	})
}
