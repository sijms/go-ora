package converters

import (
	"bytes"
	"testing"
)

func TestClobCharsetRoundTrip(t *testing.T) {
	const input = "prefix این یک متن تستی است، و پایان العربية"

	tests := []struct {
		name      string
		charsetID int
	}{
		{name: "AL32UTF8", charsetID: 873},
		{name: "UTF16BE", charsetID: 2000},
		{name: "UTF16LE", charsetID: 2002},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			converter := NewStringConverter(test.charsetID)
			encoded := converter.Encode(input)
			if len(encoded) == 0 {
				t.Fatal("encoding returned no data")
			}
			if test.charsetID == 873 && !bytes.Equal(encoded, []byte(input)) {
				t.Fatal("AL32UTF8 encoding differs from UTF-8 bytes")
			}
			if decoded := converter.Decode(encoded); decoded != input {
				t.Fatalf("decoded value differs: %q", decoded)
			}
		})
	}
}
