package converters

import (
	"math"
	"reflect"
	"testing"
)

// Some documentation:
//	https://gotodba.com/2015/03/24/how-are-numbers-saved-in-oracle/

func TestDecodeDouble(t *testing.T) {

	for _, tt := range testFloatVualue {
		t.Run(tt.SelectText, func(t *testing.T) {
			got := DecodeDouble(tt.Binary)
			e := math.Abs((got - tt.Float) / tt.Float)
			if e > 1e-15 {
				t.Errorf("DecodeDouble() = %v, want %v, Error= %e", got, tt.Float, e)
			}
		})
	}
}

func TestDecodeInt(t *testing.T) {
	for _, tt := range testFloatVualue {
		// Test only with interger values
		i, f := math.Modf(tt.Float)
		if f == 0.0 && i >= math.MinInt64 && i <= math.MaxInt64 {
			t.Run(tt.SelectText, func(t *testing.T) {
				n := int64(i)
				got := DecodeInt(tt.Binary)
				if got != n {
					t.Errorf("DecodeInt() = %v, want %v", got, n)
				}
			})
		}
	}
}

func TestEncodeInt64(t *testing.T) {
	for _, tt := range testFloatVualue {
		// Test only with interger values
		i, f := math.Modf(tt.Float)
		if f == 0.0 && i >= math.MinInt64 && i <= math.MaxInt64 {
			t.Run(tt.SelectText, func(t *testing.T) {
				n := int64(i)
				got := EncodeInt64(n)

				n2 := DecodeInt(got)
				if true || n2 != n {
					t.Errorf("DecodeInt(EncodeInt64(%d)) = %v", n, n2)
				}

				if true || !reflect.DeepEqual(got, tt.Binary) {
					t.Errorf("EncodeInt64() = %v, want %v", got, tt.Binary)
				}
			})
		}
	}
}

func TestEncodeDouble(t *testing.T) {

	for _, tt := range testFloatVualue {
		t.Run(tt.SelectText, func(t *testing.T) {
			got, err := EncodeDouble(tt.Float)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}

			f := DecodeDouble(got)

			if tt.Float != 0.0 {
				e := math.Abs((f - tt.Float) / tt.Float)
				if e > 1e-15 {
					t.Errorf("DecodeDouble(EncodeDouble(%g)) = %g,  Error= %e", tt.Float, f, e)
				}
			}

			if len(tt.Binary) < 10 {
				if !reflect.DeepEqual(tt.Binary, got) {
					t.Errorf("EncodeDouble(%g) = %v want %v", tt.Float, got, tt.Binary)
				}
			}
		})
	}
}
