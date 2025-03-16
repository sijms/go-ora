package converters

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

/*
BINARY_FLOAT and BINARY_DOUBLE encoding observed using
SELECT dump(cast(xxx as binary_yyy) FROM dual;
*/
func EncodeBool(val bool) []byte {
	if val {
		return []byte{1, 1}
	} else {
		return []byte{1, 0}
	}
}

func DecodeBool(data []byte) bool {
	return bytes.Compare(data, []byte{1, 1}) == 0
}

func EncodeFloat32(number float32) []byte {
	b := make([]byte, 4)
	temp := math.Float32bits(number)
	binary.BigEndian.PutUint32(b, uint32(temp))
	if number > 0 {
		b[0] = b[0] | 128
	} else {
		b[0] = ^b[0]
		b[1] = ^b[1]
		b[2] = ^b[2]
		b[3] = ^b[3]
	}
	return b

}
func ConvertBinaryFloat(bytes []byte) float32 {
	if bytes[0]&128 != 0 {
		bytes[0] = bytes[0] & 127
	} else {
		bytes[0] = ^bytes[0]
		bytes[1] = ^bytes[1]
		bytes[2] = ^bytes[2]
		bytes[3] = ^bytes[3]
	}
	u := binary.BigEndian.Uint32(bytes)
	// user can cast to float64 on their side if necessary (pass float64 field as parameter for row.Scan method)
	return math.Float32frombits(u)
	//test2 := float64(test)
	//return test2
	//if u > (1 << 31) {
	//	return -math.Float32frombits(u)
	//}
	//return math.Float32frombits(^u)
}

func EncodeFloat64(number float64) []byte {
	b := make([]byte, 8)
	temp := math.Float64bits(number)
	binary.BigEndian.PutUint64(b, uint64(temp))
	if number > 0 {
		b[0] = b[0] | 128
	} else {
		b[0] = ^b[0]
		b[1] = ^b[1]
		b[2] = ^b[2]
		b[3] = ^b[3]
		b[4] = ^b[4]
		b[5] = ^b[5]
		b[6] = ^b[6]
		b[7] = ^b[7]
	}
	return b
}
func ConvertBinaryDouble(bytes []byte) float64 {
	if bytes[0]&128 != 0 {
		bytes[0] = bytes[0] & 127
	} else {
		bytes[0] = ^bytes[0]
		bytes[1] = ^bytes[1]
		bytes[2] = ^bytes[2]
		bytes[3] = ^bytes[3]
		bytes[4] = ^bytes[4]
		bytes[5] = ^bytes[5]
		bytes[6] = ^bytes[6]
		bytes[7] = ^bytes[7]
	}
	u := binary.BigEndian.Uint64(bytes)
	return math.Float64frombits(u)
}

/*
INTERVAL_xxx encoding described at https://www.orafaq.com/wiki/Interval
*/

func ConvertIntervalYM_DTY(val []byte) string {
	/*
	   The first 4 bytes gives the number of years, the fifth byte gives the number of months in the following format:
	   years + 0x80000000
	   months + 60
	*/
	uyears := binary.BigEndian.Uint32(val)
	years := int64(uyears) - 0x80000000
	if years >= 0 && val[4] >= 60 {
		months := val[4] - 60
		return fmt.Sprintf("+%02d-%02d", years, months)
	}

	years = -years
	months := 60 - val[4]
	return fmt.Sprintf("-%02d-%02d", years, months)
}

func ConvertIntervalDS_DTY(val []byte) string {
	/*
	   The first 4 bytes gives the number of days, the last 4 ones the number of nanoseconds and the 3 in the middle the number of hours, minutes and seconds in the following format:

	   days + 0x80000000
	   hours + 60
	   minutes + 60
	   seconds + 60
	   nanoseconds
	*/
	udays := binary.BigEndian.Uint32(val)
	uns := binary.BigEndian.Uint32(val[7:])

	days := int64(udays) - 0x80000000
	ns := (int64(uns) - 0x80000000) / 1000

	if days >= 0 && ns >= 0 && val[4] >= 60 && val[5] >= 60 && val[6] >= 60 {
		hours := val[4] - 60
		mins := val[5] - 60
		secs := val[6] - 60

		return fmt.Sprintf("+%02d %02d:%02d:%02d.%06d", days, hours, mins, secs, ns)
	}
	days = -days
	hours := 60 - val[4]
	mins := 60 - val[5]
	secs := 60 - val[6]
	ns = -ns
	return fmt.Sprintf("-%02d %02d:%02d:%02d.%06d", days, hours, mins, secs, ns)
}
