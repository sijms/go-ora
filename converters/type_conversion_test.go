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
func TestDecodeDouble(t *testing.T) {
	type args struct {
		inputData []byte
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "0",
			args: args{
				[]byte{128},
			},
			want: float64(0),
		},
		{
			name: "1008",
			args: args{
				[]byte{194, 11, 9},
			},
			want: float64(1008),
		},
		{
			name: "10.08",
			args: args{
				[]byte{193, 11, 9},
			},
			want: float64(10.08),
		},
		{
			name: "-12300",
			args: args{
				[]byte{60, 100, 78, 102},
			},
			want: float64(-12300),
		},
		{
			name: "69",
			args: args{
				[]byte{193, 70},
			},
			want: float64(69),
		},
		{
			name: "-10",
			args: args{
				[]byte{62, 91, 102},
			},
			want: float64(-10),
		},

		{
			name: "-1",
			args: args{
				[]byte{62, 100, 102},
			},
			want: float64(-1),
		},
		{
			name: "-1008",
			args: args{
				[]byte{61, 91, 93, 102},
			},
			want: float64(-1008),
		},
		{
			name: "123456.78",
			args: args{
				[]byte{195, 13, 35, 57, 79},
			},
			want: float64(123456.78),
		},
		{
			name: "-123456.78",
			args: args{
				[]byte{60, 89, 67, 45, 23, 102},
			},
			want: float64(-123456.78),
		},
		{
			name: "2*1/3",
			args: args{
				[]byte{192, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 67, 68},
			},
			want: float64(2.0 * 1.0 / 3.0),
		},
		{
			name: "sqrt(2)",
			args: args{
				[]byte{193, 2, 42, 43, 14, 57, 24, 74, 10, 51, 49, 81, 17, 89, 73, 43, 10, 70, 81, 79, 58},
			},
			want: math.Sqrt2,
		},
		{
			name: "2*asin(1)",
			args: args{
				[]byte{193, 4, 15, 16, 93, 66, 36, 90, 80, 33, 39, 47, 27, 44, 39, 33, 80, 51, 29, 85, 21},
			},
			want: 2.0 * math.Asin(1.0),
		},
		{
			name: "1e125",
			args: args{
				[]byte{255, 11},
			},
			want: 1e125,
		},
		{
			name: "1e-125",
			args: args{
				[]byte{130, 11},
			},
			want: 1e-125,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecodeDouble(tt.args.inputData)
			if math.Abs((got-tt.want)/tt.want) > 1e-15 {
				t.Errorf("DecodeDouble() = %v, want %v, Error= %e", got, tt.want, math.Abs((got-tt.want)/tt.want))
			}
		})
	}
}
