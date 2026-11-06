package converters

import (
	"bytes"
	"testing"
)

func TestUnrepresentableCharacterReplacement(t *testing.T) {
	tests := []struct {
		name string
		id   int
		want []byte
	}{
		{"US7ASCII", 1, []byte{0x3f}},
		{"WE8MSWIN1252", 178, []byte{0xbf}},
		{"JA16EUC", 830, []byte{0xa1, 0xa9}},
		{"JA16SJIS", 832, []byte{0x81, 0x48}},
		{"JA16DBCS", 833, []byte{0x0e, 0x42, 0x6f, 0x0f}},
		{"KO16KSC5601", 840, []byte{0xa3, 0xbf}},
		{"KO16KSCCS", 845, []byte{0xd9, 0x39}},
		{"KO16MSWIN949", 846, []byte{0xa3, 0xbf}},
		{"ZHS16GBK", 852, []byte{0xa3, 0xbf}},
		{"ZHT16BIG5", 865, []byte{0xa1, 0x48}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewStringConverter(tt.id)
			for _, input := range []string{"☃", "😀"} {
				if got := conv.Encode(input); !bytes.Equal(got, tt.want) {
					t.Errorf("Encode(%q) = %x, want %x", input, got, tt.want)
				}
			}
		})
	}
}

func TestKO16KSC5601ReplacementPreservesAdjacentCharacters(t *testing.T) {
	conv := NewStringConverter(840)
	got := conv.Encode("가나다힣₩😀")
	want := []byte{0xb0, 0xa1, 0xb3, 0xaa, 0xb4, 0xd9, 0xa3, 0xbf, 0xa3, 0xbf, 0xa3, 0xbf}
	if !bytes.Equal(got, want) {
		t.Fatalf("Encode() = %x, want %x", got, want)
	}
	if decoded := conv.Decode(got); decoded != "가나다？？？" {
		t.Fatalf("Decode() = %q, want %q", decoded, "가나다？？？")
	}
}
