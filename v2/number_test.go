package go_ora

import (
	"strings"
	"testing"
)

func TestNumberStringNearZero(t *testing.T) {
	// The exponent byte 0x80 is shared by zero, which is that single byte, and by positive numbers from 1E-130 up to
	// 1E-128, which have mantissa bytes after it.
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"zero", []byte{0x80}, "0"},
		{"1E-130", []byte{0x80, 0x02}, "0." + strings.Repeat("0", 129) + "1"},
		{"1.5E-129", []byte{0x80, 0x10}, "0." + strings.Repeat("0", 128) + "15"},
		{"1E-128", []byte{0x81, 0x02}, "0." + strings.Repeat("0", 127) + "1"},
	}
	for _, tt := range tests {
		num := Number{data: tt.data}
		got, err := num.String()
		if err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		if got != tt.want {
			t.Errorf("%s: got %s, want %s", tt.name, got, tt.want)
		}
		if isZero := num.isZero(); isZero != (tt.name == "zero") {
			t.Errorf("%s: isZero() = %v", tt.name, isZero)
		}
	}
}
