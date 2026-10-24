package converters

import (
	"bytes"
	"testing"
)

func TestJA16EUCReplacement(t *testing.T) {
	// Oracle JDBC 19.3 CharacterSet.convertWithReplacement uses 0x3F for
	// these unsupported BMP characters in each of the JA16EUC variants.
	for _, charset := range []struct {
		name string
		id   int
	}{
		{"JA16EUC", 830},
		{"JA16EUCYEN", 831},
		{"JA16EUCTILDE", 837},
	} {
		t.Run(charset.name, func(t *testing.T) {
			conv := NewStringConverter(charset.id)
			for _, tc := range []struct {
				input string
				want  []byte
			}{
				{"Ⅰ", []byte{0x3f}},
				{"Ⅱ", []byte{0x3f}},
				{"ⅠⅡ", []byte{0x3f, 0x3f}},
				{"AⅡB", []byte{'A', '?', 'B'}},
				{"テストハウスⅡ", []byte{0xa5, 0xc6, 0xa5, 0xb9, 0xa5, 0xc8, 0xa5, 0xcf, 0xa5, 0xa6, 0xa5, 0xb9, 0x3f}},
				{"テストハウス", []byte{0xa5, 0xc6, 0xa5, 0xb9, 0xa5, 0xc8, 0xa5, 0xcf, 0xa5, 0xa6, 0xa5, 0xb9}},
				{"ABC?", []byte("ABC?")},
				{"", nil},
			} {
				for _, encoder := range []IStringConverter{conv, conv.Clone()} {
					if got := encoder.Encode(tc.input); !bytes.Equal(got, tc.want) {
						t.Errorf("Encode(%q) = % X; want % X", tc.input, got, tc.want)
					}
				}
			}
		})
	}
}

func TestRomanNumeralSupportedCharsets(t *testing.T) {
	for _, tc := range []struct {
		name string
		id   int
		want []byte
	}{
		{"JA16SJIS", 832, []byte{0x87, 0x55}},
		{"JA16SJISYEN", 834, []byte{0xfa, 0x4b}},
		{"JA16SJISTILDE", 838, []byte{0x87, 0x55}},
		{"AL32UTF8", 873, []byte{0xe2, 0x85, 0xa1}},
		{"AL16UTF16", 2000, []byte{0x21, 0x61}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conv := NewStringConverter(tc.id)
			got := conv.Encode("Ⅱ")
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("Encode = % X; want % X", got, tc.want)
			}
			if back := conv.Decode(got); back != "Ⅱ" {
				t.Errorf("Decode = %q; want Ⅱ", back)
			}
		})
	}
}
