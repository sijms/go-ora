package network

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// clrValue encodes value as a CLR: one length byte for a short value, or 0xFE then chunks of at most chunk bytes, each
// with its length in one byte or, with big, in a length-prefixed int, ending with a zero length.
func clrValue(value []byte, chunk int, big bool) []byte {
	if len(value) < 0xFE && chunk == 0 {
		return append([]byte{byte(len(value))}, value...)
	}
	out := []byte{0xFE}
	size := func(n int) {
		if !big {
			out = append(out, byte(n))
			return
		}
		var b [4]byte
		binary.BigEndian.PutUint32(b[:], uint32(n))
		trimmed := bytes.TrimLeft(b[:], "\x00")
		out = append(out, byte(len(trimmed)))
		out = append(out, trimmed...)
	}
	for len(value) > 0 {
		n := chunk
		if n > len(value) {
			n = len(value)
		}
		size(n)
		out = append(out, value[:n]...)
		value = value[n:]
	}
	size(0)
	return out
}

func TestSessionAppendClr(t *testing.T) {
	long := make([]byte, 70000)
	for i := range long {
		long[i] = byte(i * 7)
	}
	var input []byte
	input = append(input, 0) // NULL
	input = append(input, clrValue([]byte("abc"), 0, false)...)
	input = append(input, clrValue(long[:4000], 64, false)...)
	session := NewSessionWithInputBufferForDebug(input)
	dst := []byte("prefix")
	var err error
	for i := 0; i < 3; i++ {
		if dst, err = session.AppendClr(dst); err != nil {
			t.Fatalf("value %d: %v", i, err)
		}
	}
	want := append(append([]byte("prefix"), "abc"...), long[:4000]...)
	if !bytes.Equal(dst, want) {
		t.Fatal("wrong bytes")
	}

	// Big chunks, into a buffer large enough, which isn't reallocated.
	session = NewSessionWithInputBufferForDebug(clrValue(long, 32767, true))
	session.UseBigClrChunks = true
	buffer := make([]byte, 0, len(long))
	got, err := session.AppendClr(buffer)
	if err != nil || !bytes.Equal(got, long) || &got[0] != &buffer[:1][0] {
		t.Fatalf("big chunks: %v", err)
	}
}
