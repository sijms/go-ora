package converters

import (
	"testing"
)

/*
SQL> SELECT dump(cast(134.45 as binary_float)) FROM dual;
Typ=100 Len=4: 195,6,115,51

SQL> SELECT dump(cast(-134.45 as binary_float)) FROM dual;
Typ=100 Len=4: 60,249,140,204
*/
func TestBinaryFloat(t *testing.T) {
	cases := []struct {
		raw      []byte
		expected float32
	}{
		{
			[]byte{195, 6, 115, 51},
			134.45,
		},
		{
			[]byte{60, 249, 140, 204},
			-134.45,
		},
	}

	for _, c := range cases {
		v := ConvertBinaryFloat(c.raw)
		if v != c.expected {
			t.Errorf("expected %v, got %v", c.expected, v)
		}
	}
}

/*
SQL> SELECT dump(cast(134.45 as binary_double)) FROM dual;
Typ=101 Len=8: 192,96,206,102,102,102,102,102

SQL> SELECT dump(cast(-134.45 as binary_double)) FROM dual;
Typ=101 Len=8: 63,159,49,153,153,153,153,153
*/
func TestBinaryDouble(t *testing.T) {
	cases := []struct {
		raw      []byte
		expected float64
	}{
		{
			[]byte{192, 96, 206, 102, 102, 102, 102, 102},
			134.45,
		},
		{
			[]byte{63, 159, 49, 153, 153, 153, 153, 153},
			-134.45,
		},
	}

	for _, c := range cases {
		v := ConvertBinaryDouble(c.raw)
		if v != c.expected {
			t.Errorf("expected %v, got %v", c.expected, v)
		}
	}
}

/*
SQL> SELECT dump(cast(TO_YMINTERVAL('2021-10') as INTERVAL YEAR TO MONTH)) FROM dual;
Typ=182 Len=5: 128,0,7,229,70
SQL> SELECT cast(TO_YMINTERVAL('2021-10') as INTERVAL YEAR TO MONTH) FROM dual;
+2021-10

SQL> SELECT dump(cast(TO_YMINTERVAL('-2021-10') as INTERVAL YEAR TO MONTH)) FROM dual;
Typ=182 Len=5: 127,255,248,27,50
SQL> SELECT cast(TO_YMINTERVAL('-2021-10') as INTERVAL YEAR TO MONTH) FROM dual;
-2021-10

SQL> SELECT dump(cast(TO_YMINTERVAL('-5-10') as INTERVAL YEAR TO MONTH)) FROM dual;
Typ=182 Len=5: 127,255,255,251,50
SQL> SELECT cast(TO_YMINTERVAL('-5-10') as INTERVAL YEAR TO MONTH) FROM dual;
-05-10

SQL> SELECT dump(cast(TO_YMINTERVAL('00-10') as INTERVAL YEAR TO MONTH)) FROM dual;
Typ=182 Len=5: 128,0,0,0,70
SQL> SELECT cast(TO_YMINTERVAL('00-10') as INTERVAL YEAR TO MONTH) FROM dual;
+00-10

Note that heading + is expected in the string representation
*/

type intervalCase struct {
	raw      []byte
	expected string
}

func TestIntervalYM(t *testing.T) {
	cases := []intervalCase{
		{
			[]byte{128, 0, 7, 229, 70},
			"+2021-10",
		},
		{
			[]byte{127, 255, 248, 27, 50},
			"-2021-10",
		},
		{
			[]byte{127, 255, 255, 251, 50},
			"-05-10",
		},
		{
			[]byte{128, 0, 0, 0, 70},
			"+00-10",
		},
	}

	for _, c := range cases {
		v := ConvertIntervalYM_DTY(c.raw)
		if v != c.expected {
			t.Errorf("expected %v, got %v", c.expected, v)
		}
	}
}

/*
SQL> SELECT dump(cast(TO_DSINTERVAL('2 12:23:34.456') as INTERVAL DAY TO SECOND)) FROM dual;
Typ=183 Len=11: 128,0,0,2,72,83,94,155,46,2,0
SQL> SELECT cast(TO_DSINTERVAL('2 12:23:34.456') as INTERVAL DAY TO SECOND) FROM dual;
+02 12:23:34.456000

SQL> SELECT dump(cast(TO_DSINTERVAL('-2 12:23:34.456789') as INTERVAL DAY TO SECOND)) FROM dual;
Typ=183 Len=11: 127,255,255,254,48,37,26,100,197,243,248
SQL> SELECT cast(TO_DSINTERVAL('-2 10:20:30.456') as INTERVAL DAY TO SECOND) FROM dual;
-02 12:23:34.456789

SQL> SELECT dump(cast(TO_DSINTERVAL('0 10:20:30.456789') as INTERVAL DAY TO SECOND)) FROM dual;
Typ=183 Len=11: 128,0,0,0,70,80,90,155,58,12,8
SQL> SELECT cast(TO_DSINTERVAL('0 10:20:30.456789') as INTERVAL DAY TO SECOND) FROM dual;
+00 10:20:30.456789

Note that heading + is expected in the string representation
*/
func TestIntervalDS(t *testing.T) {
	cases := []intervalCase{
		{
			[]byte{128, 0, 0, 2, 72, 83, 94, 155, 46, 2, 0},
			"+02 12:23:34.456000",
		},
		{
			[]byte{127, 255, 255, 254, 48, 37, 26, 100, 197, 243, 248},
			"-02 12:23:34.456789",
		},
		{
			[]byte{128, 0, 0, 0, 70, 80, 90, 155, 58, 12, 8},
			"+00 10:20:30.456789",
		},
	}

	for _, c := range cases {
		v := ConvertIntervalDS_DTY(c.raw)
		if v != c.expected {
			t.Errorf("expected %v, got %v", c.expected, v)
		}
	}
}
