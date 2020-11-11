package converters

import (
	"math"
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
				e = math.Abs(got-tt.Float) / tt.Float
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
