package converters

import (
	"math"
	"reflect"
	"testing"
)

// Some documentation:
//	https://gotodba.com/2015/03/24/how-are-numbers-saved-in-oracle/
//  How to get test values:
//		 select N, vsize(N), dump(N) from (
//		 	select 0 N from dual union
//		 	select 1 N from dual union
//		 	select 69 N from dual union
//		 	select 1008 N from dual union
//		 	select -1 N from dual union
//		 	select -1008 N from dual
//		 	)
//		 	;
//
//		 -1008	4	Typ=2 Len=4: 61,91,93,102
//		 -1		3	Typ=2 Len=3: 62,100,102
//		 0		1	Typ=2 Len=1: 128
//		 1		2	Typ=2 Len=2: 193,2
//		 69		2	Typ=2 Len=2: 193,70
//		 1008	3	Typ=2 Len=3: 194,11,9
func TestDecodeDouble2(t *testing.T) {

	for _, tt := range testFloatVualue {
		t.Run(tt.SelectText, func(t *testing.T) {
			got := DecodeDouble2(tt.Binary)
			var e float64
			if tt.Float != 0 {
				e = math.Abs((got - tt.Float) / tt.Float)
			}
			if e > 1e-15 {
				t.Errorf("DecodeDouble2() = %g, want %g, Error= %e", got, tt.Float, e)
			}
		})
	}
}

func TestDecodeDouble(t *testing.T) {

	for _, tt := range testFloatVualue {
		t.Run(tt.SelectText, func(t *testing.T) {
			got := DecodeDouble(tt.Binary)
			e := math.Abs(got - tt.Float)
			if e > 1e-15 {
				t.Errorf("DecodeDouble() = %v, want %v, Error= %e", got, tt.Float, e)
			}
		})
	}
}
func TestEncodeDouble2(t *testing.T) {

	for _, tt := range testFloatVualue {
		t.Run(tt.SelectText, func(t *testing.T) {
			got, err := EncodeDouble2(tt.Float)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}

			f := DecodeDouble2(got)

			if tt.Float != 0.0 {
				e := math.Abs((f - tt.Float) / tt.Float)
				if e > 1e-15 {
					t.Errorf("DecodeDouble2(EncodeDouble2(%g)) = %g,  Error= %e", tt.Float, f, e)
				}
			}

			if len(tt.Binary) < 10 {
				if !reflect.DeepEqual(tt.Binary, got) {
					t.Errorf("EncodeDouble2(%g) = %v want %v", tt.Float, got, tt.Binary)
				}
			}
		})
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
