// 2022/9/13 Bin Liu <bin.liu@enmotech.com>

package converters

import (
	"testing"
)

func TestNewNumber(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "0",
			args: args{
				b: []byte{128},
			},
			want: "0",
		},
		{
			name: "999999999999.9999",
			args: args{
				b: []byte{198, 100, 100, 100, 100, 100, 100, 100, 100},
			},
			want: "999999999999.9999",
		},
		{
			name: "9999999999999.999",
			args: args{
				b: []byte{199, 10, 100, 100, 100, 100, 100, 100, 100, 91},
			},
			want: "9999999999999.999",
		},
		{
			name: "99999999999999.99",
			args: args{
				b: []byte{199, 100, 100, 100, 100, 100, 100, 100, 100},
			},
			want: "99999999999999.99",
		},
		{
			name: "999999999999999.9",
			args: args{
				b: []byte{200, 10, 100, 100, 100, 100, 100, 100, 100, 91},
			},
			want: "999999999999999.9",
		},
		{
			name: "9999999999999999",
			args: args{
				b: []byte{200, 100, 100, 100, 100, 100, 100, 100, 100},
			},
			want: "9999999999999999",
		},
		{
			name: "999999999999.9998",
			args: args{
				b: []byte{198, 100, 100, 100, 100, 100, 100, 100, 99},
			},
			want: "999999999999.9998",
		},
		{
			name: "99999999999999.998",
			args: args{
				b: []byte{199, 100, 100, 100, 100, 100, 100, 100, 100, 81},
			},
			want: "99999999999999.998",
		},
		{
			name: "99999999999999.98",
			args: args{
				b: []byte{199, 100, 100, 100, 100, 100, 100, 100, 99},
			},
			want: "99999999999999.98",
		},
		{
			name: "100",
			args: args{
				b: []byte{194, 2},
			},
			want: "100",
		},
		{
			name: "0.1",
			args: args{
				b: []byte{192, 11},
			},
			want: "0.1",
		},
		{
			name: "0.001",
			args: args{
				b: []byte{191, 11},
			},
			want: "0.001",
		},
		{
			name: "-1.234",
			args: args{
				b: []byte{62, 100, 78, 61, 102},
			},
			want: "-1.234",
		},
		{name: "NULL", args: args{b: []byte{255}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewNumber(tt.args.b)
			got, err := n.String()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCurrentPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("String() want %v got %v ", tt.want, got)
			}
		})
	}
}

var TestOraNumberValue = []struct {
	SelectText    string
	OracleText    string
	Float         float64
	Integer       int64
	Uint64        uint64
	IsUint64      bool
	IsInteger     bool
	Binary        []byte
	wantErr       bool
	wantInt64Err  bool
	wantUInt64Err bool
}{
	{
		SelectText: "0",
		OracleText: "0",
		IsInteger:  true,
		Binary:     []byte{128},
	}, // 0.000000e+00
	{
		SelectText: "1",
		OracleText: "1",
		Float:      1,
		Integer:    1,
		IsInteger:  true,
		Binary:     []byte{193, 2},
	}, // 1.000000e+00
	{SelectText: "10", OracleText: "10", Float: 10, Integer: 10, IsInteger: true, Binary: []byte{193, 11}}, // 1.000000e+01
	{
		SelectText: "100",
		OracleText: "100",
		Float:      100,
		Integer:    100,
		IsInteger:  true,
		Binary:     []byte{194, 2},
	}, // 1.000000e+02
	{
		SelectText: "1000",
		OracleText: "1000",
		Float:      1000,
		Integer:    1000,
		IsInteger:  true,
		Binary:     []byte{194, 11},
	}, // 1.000000e+03
	{
		SelectText: "10000000",
		OracleText: "10000000",
		Float:      1e+07,
		Integer:    10000000,
		IsInteger:  true,
		Binary:     []byte{196, 11},
	}, // 1.000000e+07
	{
		SelectText:   "1E+30",
		OracleText:   "1000000000000000000000000000000",
		Float:        1e+30,
		Binary:       []byte{208, 2},
		wantInt64Err: true,
	}, // 1.000000e+30
	{
		SelectText: "0.1",
		OracleText: "0.1",
		Float:      0.1,
		Binary:     []byte{192, 11},
	}, // 1.000000e-01
	{
		SelectText: "0.01",
		OracleText: "0.01",
		Float:      0.01,
		Binary:     []byte{192, 2},
	}, // 1.000000e-02
	{
		SelectText: "0.001",
		OracleText: "0.001",
		Float:      0.001,
		Binary:     []byte{191, 11},
	}, // 1.000000e-03
	{
		SelectText: "0.0001",
		OracleText: "0.0001",
		Float:      0.0001,
		Binary:     []byte{191, 2},
	}, // 1.000000e-04
	{
		SelectText: "0.00001",
		OracleText: "0.00001",
		Float:      1e-05,
		Binary:     []byte{190, 11},
	}, // 1.000000e-05
	{
		SelectText: "0.000001",
		OracleText: "0.000001",
		Float:      1e-06,
		Binary:     []byte{190, 2},
	}, // 1.000000e-06
	{
		SelectText:   "1E+125",
		OracleText:   "100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		Float:        1e+125,
		Binary:       []byte{255, 11},
		wantInt64Err: true,
	}, // 1.000000e+125
	{
		SelectText: "1E-125",
		OracleText: "0.00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001",
		Float:      1e-125, Binary: []byte{130, 11},
	}, // 1.000000e-125
	{
		SelectText: "-1E+125",
		OracleText: "-100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
		Float:      -1e+125, Binary: []byte{0, 91, 102},
		wantInt64Err: true,
	}, // -1.000000e+125
	{
		SelectText: "-1E-125",
		OracleText: "-0.00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001",
		Float:      -1e-125, Binary: []byte{125, 91, 102},
	}, // -1.000000e-125
	{SelectText: "1.23456789e15", OracleText: "1234567890000000", Float: 1.23456789e+15, Integer: 1234567890000000, IsInteger: true, Binary: []byte{200, 13, 35, 57, 79, 91}}, // 1.234568e+15
	{SelectText: "1.23456789e-15", OracleText: "0.00000000000000123456789", Float: 1.23456789e-15, Binary: []byte{185, 13, 35, 57, 79, 91}},                                   // 1.234568e-15
	{
		SelectText: "1.234",
		OracleText: "1.234",
		Float:      1.234,
		Integer:    1,
		Binary:     []byte{193, 2, 24, 41},
	}, // 1.234000e+00
	{
		SelectText: "12.34",
		OracleText: "12.34",
		Float:      12.34,
		Integer:    12,
		Binary:     []byte{193, 13, 35},
	}, // 1.234000e+01
	{
		SelectText: "123.4",
		OracleText: "123.4",
		Integer:    123,
		Float:      123.4,
		Binary:     []byte{194, 2, 24, 41},
	}, // 1.234000e+02
	{SelectText: "1234", OracleText: "1234", Float: 1234, Integer: 1234, IsInteger: true, Binary: []byte{194, 13, 35}},                  // 1.234000e+03
	{SelectText: "12340", OracleText: "12340", Float: 12340, Integer: 12340, IsInteger: true, Binary: []byte{195, 2, 24, 41}},           // 1.234000e+04
	{SelectText: "123400", OracleText: "123400", Float: 123400, Integer: 123400, IsInteger: true, Binary: []byte{195, 13, 35}},          // 1.234000e+05
	{SelectText: "1234000", OracleText: "1234000", Float: 1.234e+06, Integer: 1234000, IsInteger: true, Binary: []byte{196, 2, 24, 41}}, // 1.234000e+06
	{SelectText: "12340000", OracleText: "12340000", Float: 1.234e+07, Integer: 12340000, IsInteger: true, Binary: []byte{196, 13, 35}}, // 1.234000e+07
	{SelectText: "0.1234", OracleText: "0.1234", Float: 0.1234, Binary: []byte{192, 13, 35}},                                            // 1.234000e-01
	{SelectText: "0.01234", OracleText: "0.01234", Float: 0.01234, Binary: []byte{192, 2, 24, 41}},                                      // 1.234000e-02
	{SelectText: "0.001234", OracleText: "0.001234", Float: 0.001234, Binary: []byte{191, 13, 35}},                                      // 1.234000e-03
	{SelectText: "0.0001234", OracleText: "0.0001234", Float: 0.0001234, Binary: []byte{191, 2, 24, 41}},                                // 1.234000e-04
	{SelectText: "0.00001234", OracleText: "0.00001234", Float: 1.234e-05, Binary: []byte{190, 13, 35}},                                 // 1.234000e-05
	{SelectText: "0.000001234", OracleText: "0.000001234", Float: 1.234e-06, Binary: []byte{190, 2, 24, 41}},                            // 1.234000e-06
	{
		SelectText: "-1.234",
		OracleText: "-1.234",
		Float:      -1.234,
		Integer:    -1,
		Binary:     []byte{62, 100, 78, 61, 102},
	}, // -1.234000e+00
	{
		SelectText: "-12.34",
		OracleText: "-12.34",
		Float:      -12.34,
		Integer:    -12,
		Binary:     []byte{62, 89, 67, 102},
	}, // -1.234000e+01
	{
		SelectText: "-123.4",
		OracleText: "-123.4",
		Float:      -123.4,
		Integer:    -123,
		Binary:     []byte{61, 100, 78, 61, 102},
	}, // -1.234000e+02
	{SelectText: "-1234", OracleText: "-1234", Float: -1234, Integer: -1234, IsInteger: true, Binary: []byte{61, 89, 67, 102}},                    // -1.234000e+03
	{SelectText: "-12340", OracleText: "-12340", Float: -12340, Integer: -12340, IsInteger: true, Binary: []byte{60, 100, 78, 61, 102}},           // -1.234000e+04
	{SelectText: "-123400", OracleText: "-123400", Float: -123400, Integer: -123400, IsInteger: true, Binary: []byte{60, 89, 67, 102}},            // -1.234000e+05
	{SelectText: "-1234000", OracleText: "-1234000", Float: -1.234e+06, Integer: -1234000, IsInteger: true, Binary: []byte{59, 100, 78, 61, 102}}, // -1.234000e+06
	{SelectText: "-12340000", OracleText: "-12340000", Float: -1.234e+07, Integer: -12340000, IsInteger: true, Binary: []byte{59, 89, 67, 102}},   // -1.234000e+07 	// -1.234000e+00
	{
		SelectText: "9.8765",
		OracleText: "9.8765",
		Float:      9.8765,
		Integer:    9,
		Binary:     []byte{193, 10, 88, 66},
	}, // 9.876500e+00
	{
		SelectText: "98.765",
		OracleText: "98.765",
		Float:      98.765,
		Integer:    98,
		Binary:     []byte{193, 99, 77, 51},
	}, // 9.876500e+01
	{
		SelectText: "987.65",
		OracleText: "987.65",
		Float:      987.65,
		Integer:    987,
		Binary:     []byte{194, 10, 88, 66},
	}, // 9.876500e+02
	{
		SelectText: "9876.5",
		OracleText: "9876.5",
		Float:      9876.5,
		Integer:    9876,
		Binary:     []byte{194, 99, 77, 51},
	}, // 9.876500e+03
	{SelectText: "98765", OracleText: "98765", Float: 98765, Integer: 98765, IsInteger: true, Binary: []byte{195, 10, 88, 66}},            // 9.876500e+04
	{SelectText: "987650", OracleText: "987650", Float: 987650, Integer: 987650, IsInteger: true, Binary: []byte{195, 99, 77, 51}},        // 9.876500e+05
	{SelectText: "9876500", OracleText: "9876500", Float: 9.8765e+06, Integer: 9876500, IsInteger: true, Binary: []byte{196, 10, 88, 66}}, // 9.876500e+06
	{SelectText: "0.98765", OracleText: "0.98765", Float: 0.98765, Binary: []byte{192, 99, 77, 51}},                                       // 9.876500e-01
	{SelectText: "0.098765", OracleText: "0.098765", Float: 0.098765, Binary: []byte{192, 10, 88, 66}},                                    // 9.876500e-02
	{SelectText: "0.0098765", OracleText: "0.0098765", Float: 0.0098765, Binary: []byte{191, 99, 77, 51}},                                 // 9.876500e-03
	{SelectText: "0.00098765", OracleText: "0.00098765", Float: 0.00098765, Binary: []byte{191, 10, 88, 66}},                              // 9.876500e-04
	{SelectText: "0.000098765", OracleText: "0.000098765", Float: 9.8765e-05, Binary: []byte{190, 99, 77, 51}},                            // 9.876500e-05
	{SelectText: "0.0000098765", OracleText: "0.0000098765", Float: 9.8765e-06, Binary: []byte{190, 10, 88, 66}},                          // 9.876500e-06
	{SelectText: "0.00000098765", OracleText: "0.00000098765", Float: 9.8765e-07, Binary: []byte{189, 99, 77, 51}},                        // 9.876500e-07
	{
		SelectText: "-9.8765",
		OracleText: "-9.8765",
		Float:      -9.8765,
		Integer:    -9,
		Binary:     []byte{62, 92, 14, 36, 102},
	}, // -9.876500e+00
	{
		SelectText: "-98.765",
		OracleText: "-98.765",
		Float:      -98.765,
		Integer:    -98,
		Binary:     []byte{62, 3, 25, 51, 102},
	}, // -9.876500e+01
	{
		SelectText: "-987.65",
		OracleText: "-987.65",
		Float:      -987.65,
		Integer:    -987,
		Binary:     []byte{61, 92, 14, 36, 102},
	}, // -9.876500e+02
	{
		SelectText: "-9876.5",
		OracleText: "-9876.5",
		Float:      -9876.5,
		Integer:    -9876,
		Binary:     []byte{61, 3, 25, 51, 102},
	}, // -9.876500e+03
	{SelectText: "-98765", OracleText: "-98765", Float: -98765, Integer: -98765, IsInteger: true, Binary: []byte{60, 92, 14, 36, 102}},            // -9.876500e+04
	{SelectText: "-987650", OracleText: "-987650", Float: -987650, Integer: -987650, IsInteger: true, Binary: []byte{60, 3, 25, 51, 102}},         // -9.876500e+05
	{SelectText: "-9876500", OracleText: "-9876500", Float: -9.8765e+06, Integer: -9876500, IsInteger: true, Binary: []byte{59, 92, 14, 36, 102}}, // -9.876500e+06
	{
		SelectText: "-98765000",
		OracleText: "-98765000",
		Float:      -9.8765e+07,
		Integer:    -98765000,
		IsInteger:  true,
		Binary:     []byte{59, 3, 25, 51, 102},
	}, // -9.876500e+07
	{SelectText: "-0.98765", OracleText: "-0.98765", Float: -0.98765, Binary: []byte{63, 3, 25, 51, 102}}, // -9.876500e-01
	{
		SelectText: "-0.098765",
		OracleText: "-0.098765",
		Float:      -0.098765,
		Binary:     []byte{63, 92, 14, 36, 102},
	}, // -9.876500e-02
	{
		SelectText: "-0.0098765",
		OracleText: "-0.0098765",
		Float:      -0.0098765,
		Binary:     []byte{64, 3, 25, 51, 102},
	}, // -9.876500e-03
	{SelectText: "-0.00098765", OracleText: "-0.00098765", Float: -0.00098765, Binary: []byte{64, 92, 14, 36, 102}},      // -9.876500e-04
	{SelectText: "-0.000098765", OracleText: "-0.000098765", Float: -9.8765e-05, Binary: []byte{65, 3, 25, 51, 102}},     // -9.876500e-05
	{SelectText: "-0.0000098765", OracleText: "-0.0000098765", Float: -9.8765e-06, Binary: []byte{65, 92, 14, 36, 102}},  // -9.876500e-06
	{SelectText: "-0.00000098765", OracleText: "-0.00000098765", Float: -9.8765e-07, Binary: []byte{66, 3, 25, 51, 102}}, // -9.876500e-07
	{
		SelectText: "2*asin(1)",
		OracleText: "3.1415926535897932384626433832795028842",
		Float:      3.141592653589793,
		Integer:    3,
		Binary:     []byte{193, 4, 15, 16, 93, 66, 36, 90, 80, 33, 39, 47, 27, 44, 39, 33, 80, 51, 29, 85, 21},
	}, // 3.141593e+00
	{SelectText: "1/3", OracleText: "0.3333333333333333333333333333333333333333", Float: 0.3333333333333333, Binary: []byte{192, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34, 34}},   // 3.333333e-01
	{SelectText: "-1/3", OracleText: "-0.3333333333333333333333333333333333333333", Float: -0.3333333333333333, Binary: []byte{63, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68, 68}}, // -3.333333e-01
	{SelectText: "9000000000000000000", OracleText: "9000000000000000000", Float: 9e+18, Integer: 9000000000000000000, IsInteger: true, Binary: []byte{202, 10}},                                                    // 9.000000e+18
	{SelectText: "-9000000000000000000", OracleText: "-9000000000000000000", Float: -9e+18, Integer: -9000000000000000000, IsInteger: true, Binary: []byte{53, 92, 102}},                                            // -9.000000e+18
	{SelectText: "9223372036854775807", OracleText: "9223372036854775807", Float: 9.223372036854776e+18, Integer: 9223372036854775807, IsInteger: true, Binary: []byte{202, 10, 23, 34, 73, 4, 69, 55, 78, 59, 8}},
	{SelectText: "-9223372036854775808", OracleText: "-9223372036854775808", Float: -9.223372036854776e+18, Integer: -9223372036854775808, IsInteger: true, Binary: []byte{53, 92, 79, 68, 29, 98, 33, 47, 24, 43, 93, 102}},
	{
		SelectText: "999999999999.9999",
		OracleText: "999999999999.9999",
		Float:      999999999999.9999,
		Integer:    999999999999,
		Binary:     []byte{198, 100, 100, 100, 100, 100, 100, 100, 100},
	},
	{
		SelectText: "9999999999999.999",
		OracleText: "9999999999999.999",
		Float:      9999999999999.999,
		Integer:    9999999999999,
		Binary:     []byte{199, 10, 100, 100, 100, 100, 100, 100, 100, 91},
	},
	{
		SelectText: "99999999999999.99",
		OracleText: "99999999999999.99",
		Float:      99999999999999.99,
		Integer:    99999999999999,
		Binary:     []byte{199, 100, 100, 100, 100, 100, 100, 100, 100},
	},
	{
		SelectText: "999999999999999.9",
		OracleText: "999999999999999.9",
		Float:      999999999999999.9,
		Integer:    999999999999999,
		Binary:     []byte{200, 10, 100, 100, 100, 100, 100, 100, 100, 91},
	},
	{SelectText: "9999999999999999", OracleText: "9999999999999999", Float: 9999999999999999, Integer: 9999999999999999, IsInteger: true, Binary: []byte{200, 100, 100, 100, 100, 100, 100, 100, 100}},
	{
		SelectText: "999999999999.9998",
		OracleText: "999999999999.9998",
		Float:      999999999999.9998,
		Integer:    999999999999,
		Binary:     []byte{198, 100, 100, 100, 100, 100, 100, 100, 99},
	},
	{
		SelectText: "99999999999999.998",
		OracleText: "99999999999999.998",
		Float:      99999999999999.998,
		Integer:    99999999999999,
		Binary:     []byte{199, 100, 100, 100, 100, 100, 100, 100, 100, 81},
	},
	{
		SelectText: "99999999999999.98",
		OracleText: "99999999999999.98",
		Float:      99999999999999.98,
		Integer:    99999999999999,
		Binary:     []byte{199, 100, 100, 100, 100, 100, 100, 100, 99},
	},
	{
		SelectText: "9.9",
		OracleText: "9.9",
		Float:      9.9,
		Integer:    9,
		Binary:     []byte{193, 10, 91},
	},
	{SelectText: "Infinity", OracleText: "Infinity", Binary: []byte{255, 101}, wantInt64Err: true},
	{SelectText: "NULL", OracleText: "", Binary: []byte{255}, wantErr: true, wantInt64Err: true},
	{
		SelectText:   "18446744073709550615",
		OracleText:   "18446744073709550615",
		Binary:       []byte{202, 19, 45, 68, 45, 8, 38, 10, 56, 7, 16},
		wantErr:      false,
		wantInt64Err: true,
		Uint64:       18446744073709550615,
		IsUint64:     true,
	},
	{
		SelectText:    "-1",
		OracleText:    "-1",
		Binary:        []byte{62, 100, 102},
		Integer:       -1,
		wantErr:       false,
		wantInt64Err:  false,
		Uint64:        0,
		IsUint64:      true,
		wantUInt64Err: true,
	},
}

func TestNumberString(t *testing.T) {
	for _, tt := range TestOraNumberValue {
		t.Run(tt.SelectText, func(t *testing.T) {
			n := NewNumber(tt.Binary)
			got, err := n.String()
			if (err != nil) != tt.wantErr {
				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.OracleText {
				t.Errorf("String() want %v got %v ", tt.OracleText, got)
			}
		})
	}
}

func TestNumberToInt64(t *testing.T) {
	for _, tt := range TestOraNumberValue {
		t.Run(tt.SelectText, func(t *testing.T) {
			n := NewNumber(tt.Binary)
			got, err := n.Int64()
			if (err != nil) != tt.wantInt64Err {
				t.Errorf("String() error = %v, wantInt64Err %v", err, tt.wantErr)
				return
			}
			if got != tt.Integer {
				t.Errorf("Integer() want %v got %v ", tt.OracleText, got)
			}
		})
	}
}

func TestNumberToUInt64(t *testing.T) {
	for _, tt := range TestOraNumberValue {
		t.Run(tt.SelectText, func(t *testing.T) {
			if !tt.IsUint64 {
				t.Skip()
			}
			n := NewNumber(tt.Binary)
			got, err := n.UInt64()
			if (err != nil) != tt.wantUInt64Err {
				t.Errorf("String() error = %v, wantUInt64Err %v", err, tt.wantErr)
				return
			}
			if got != tt.Uint64 {
				t.Errorf("Uint64() want %v got %v ", tt.OracleText, got)
			}
		})
	}
}

// func TestNumber_toLnxFmt(t *testing.T) {
// 	for _, tt := range TestOraNumberValue {
// 		t.Run(tt.SelectText, func(t *testing.T) {
// 			b, err := toBytes(tt.Binary)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("String() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			fmt.Println(string(b))
// 			got, _ := ByteToNumber(b)
// 			if !reflect.DeepEqual(got, tt.Binary) {
// 				t.Errorf("ByteToNumber() got = %v, want %v", got, tt.Binary)
// 			}
// 		})
// 	}
// }

func Benchmark_Number_To_String(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// n := NewNumber([]byte{62, 3, 25, 51, 102})
		NumberToString([]byte{62, 3, 25, 51, 102})
	}
}

func Benchmark_Number_To_Byte(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// n := NewNumber([]byte{62, 3, 25, 51, 102})
		toBytes([]byte{62, 3, 25, 51, 102})
	}
}

func Benchmark_Number_Int64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		n := NewNumber([]byte{62, 3, 25, 51, 102})
		n.Int64()
	}
}

func Benchmark_DecodeNumber(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = DecodeNumber([]byte{62, 3, 25, 51, 102})
	}
}
