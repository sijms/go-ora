package converters

import (
	"fmt"
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DecodeDouble(tt.args.inputData)
			if got != tt.want {
				t.Errorf("DecodeDouble() = %v, want %v", got, tt.want)
			}
			fmt.Println(got, tt.want)
		})
	}
}
