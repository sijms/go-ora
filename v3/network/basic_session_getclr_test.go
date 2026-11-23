package network

import (
	"bytes"
	"testing"
)

// clrReader is what the tests read through.
type clrReader interface {
	GetClr() ([]byte, error)
	GetByte() (uint8, error)
}

// clrSessions are the two readers a value can come through: a network session reads
// with Session.read, a memory session with MemorySession.read.
var clrSessions = []struct {
	name string
	make func(input []byte, bigChunks bool) clrReader
}{
	{"Session", func(input []byte, bigChunks bool) clrReader {
		s := NewSessionWithInputBufferForDebug(input)
		s.terminal = s // NewSessionWithInputBufferForDebug may leave it nil
		s.UseBigClrChunks = bigChunks
		return s
	}},
	{"MemorySession", func(input []byte, bigChunks bool) clrReader {
		return NewMemorySession(input, nil, SessionProperties{UseBigClrChunks: bigChunks})
	}},
}

// GetClr reads a value of one chunk (a length byte, then the bytes), of several chunks
// (0xFE, then chunks of a length and bytes, ended by a zero length), or NULL (a length
// of 0, 0xFF or 0xFD), and stops at the value's end.
func TestGetClr(t *testing.T) {
	longest := append([]byte{0xFC}, bytes.Repeat([]byte{'x'}, 0xFC)...)
	testScenarios := []struct {
		description string
		input       []byte
		bigChunks   bool
		expected    []byte
		expectError bool
	}{
		{description: "one chunk", input: []byte{3, 'a', 'b', 'c'}, expected: []byte("abc")},
		{description: "the longest one chunk", input: longest, expected: longest[1:]},
		{description: "chunks", input: []byte{0xFE, 2, 'a', 'b', 1, 'c', 0}, expected: []byte("abc")},
		{description: "chunks with 4-byte lengths", input: []byte{0xFE, 1, 2, 'a', 'b', 1, 1, 'c', 0}, bigChunks: true, expected: []byte("abc")},
		{description: "NULL 0", input: []byte{0}},
		{description: "NULL 0xFF", input: []byte{0xFF}},
		{description: "NULL 0xFD", input: []byte{0xFD}},
		{description: "a value cut short", input: []byte{5, 'a', 'b'}, expectError: true},
	}
	for _, reader := range clrSessions {
		for _, scenario := range testScenarios {
			t.Run(reader.name+"/"+scenario.description, func(t *testing.T) {
				input := append(append([]byte{}, scenario.input...), 9) // 9: the next field
				session := reader.make(input, scenario.bigChunks)
				got, err := session.GetClr()
				if scenario.expectError {
					if err == nil {
						t.Errorf("GetClr = %q, want an error", got)
					}
					return
				}
				if err != nil || !bytes.Equal(got, scenario.expected) || (got == nil) != (scenario.expected == nil) {
					t.Fatalf("GetClr = %q, %v; want %q", got, err, scenario.expected)
				}
				if next, err := session.GetByte(); err != nil || next != 9 {
					t.Errorf("the next byte is %d, %v; want 9", next, err)
				}
			})
		}
	}
}

// A value of one chunk is its own bytes, with no spare capacity: a 7-byte DATE does
// not keep a 64-byte buffer alive, and no caller can read past the value into it.
func TestGetClrOneChunkHasNoSpareCapacity(t *testing.T) {
	for _, reader := range clrSessions {
		got, err := reader.make([]byte{7, 1, 2, 3, 4, 5, 6, 7}, false).GetClr()
		if err != nil || len(got) != 7 || cap(got) != 7 {
			t.Errorf("%s: GetClr = %v (cap %d), %v; want 7 bytes, cap 7", reader.name, got, cap(got), err)
		}
	}
}

// The value GetClr returns is the caller's: the bytes it was read from can change
// afterwards (a session reuses its input buffer) without changing the value.
func TestGetClrValueIsTheCallers(t *testing.T) {
	for _, reader := range clrSessions {
		input := []byte{3, 'a', 'b', 'c', 0xFE, 2, 'd', 'e', 0}
		session := reader.make(input, false)
		one, err1 := session.GetClr()
		chunks, err2 := session.GetClr()
		for i := range input {
			input[i] = 0
		}
		if err1 != nil || err2 != nil || string(one) != "abc" || string(chunks) != "de" {
			t.Errorf("%s, after the input changed: %q, %v and %q, %v; want \"abc\" and \"de\"", reader.name, one, err1, chunks, err2)
		}
	}
}

// BenchmarkGetClrShort reads 7-byte values, the size of a DATE.
func BenchmarkGetClrShort(b *testing.B) {
	for _, reader := range clrSessions {
		b.Run(reader.name, func(b *testing.B) {
			session := reader.make(bytes.Repeat([]byte{7, 1, 2, 3, 4, 5, 6, 7}, b.N), false)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := session.GetClr(); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
