package converters

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"
)

func encodeSign(input []byte, neg bool) []byte {
	if !neg {
		input[0] = uint8(int(input[0]) + 0x80 + 0x40 + 1)
		for x := 1; x < len(input); x++ {
			input[x] = input[x] + 1
		}
	} else {
		input[0] = 0xFF - uint8(int(input[0])+0x80+0x40+1)
		for x := 1; x < len(input); x++ {
			input[x] = uint8(101 - input[x])
		}
		if len(input) <= 20 {
			input = append(input, 102)
		}
	}
	return input
}

func EncodeInt64(val int64) []byte {
	if val == 0 {
		return []byte{0x80}
	}
	output := make([]byte, 0, 20)
	neg := val < 0
	for val != 0 {
		output = append(output, uint8(math.Abs(float64(val%100))))
		val = val / 100
	}
	for i, j := 0, len(output)-1; i < j; i, j = i+1, j-1 {
		output[i], output[j] = output[j], output[i]
	}
	byteLen := uint8(len(output)) - 1
	output = bytes.TrimRight(output, "\x00")

	output = append([]byte{byteLen}, output...)
	return encodeSign(output, neg)
}
func EncodeInt(val int) []byte {
	if val == 0 {
		return []byte{0x80}
	}
	output := make([]byte, 0, 20)
	neg := val < 0
	for val != 0 {
		output = append(output, uint8(math.Abs(float64(val%100))))
		val = val / 100
	}
	for i, j := 0, len(output)-1; i < j; i, j = i+1, j-1 {
		output[i], output[j] = output[j], output[i]
	}
	byteLen := uint8(len(output)) - 1
	output = bytes.TrimRight(output, "\x00")

	output = append([]byte{byteLen}, output...)
	return encodeSign(output, neg)
}

// EncodeDate convert time.Time into oracle representation
func EncodeDate(ti time.Time) []byte {
	ret := make([]byte, 7)
	ret[0] = uint8(ti.Year()/100 + 100)
	ret[1] = uint8(ti.Year()%100 + 100)
	ret[2] = uint8(ti.Month())
	ret[3] = uint8(ti.Day())
	ret[4] = uint8(ti.Hour() + 1)
	ret[5] = uint8(ti.Minute() + 1)
	ret[6] = uint8(ti.Second() + 1)
	return ret
}

func DecodeDate(data []byte) (time.Time, error) {
	if len(data) < 7 {
		return time.Now(), errors.New("abnormal data representation for date")
	}
	year := (int(data[0]) - 100) * 100
	year += int(data[1]) - 100
	nanoSec := 0
	if len(data) > 7 {
		nanoSec = int(binary.BigEndian.Uint32(data[7:11]))
	}
	tzHour := 0
	tzMin := 0
	if len(data) > 11 {
		tzHour = int(data[11]) - 20
		tzMin = int(data[12]) - 60
	}

	return time.Date(year, time.Month(data[2]), int(data[3]),
		int(data[4]-1)+tzHour, int(data[5]-1)+tzMin, int(data[6]-1), nanoSec, time.UTC), nil
}

// protectAddFigure check if adding digit d overflows the int64 capacity.
// Return true when overflow
func protectAddFigure(m *int64, d int64) bool {
	r := *m * 10
	if r < 0 {
		return true
	}
	r += d
	if r < 0 {
		return true
	}
	*m = r
	return false
}

// DecodeNumber decode Oracle binary representation of numbers
// and returns mantissa and exponent as int64
// Some documentation:
//	https://gotodba.com/2015/03/24/how-are-numbers-saved-in-oracle/
//  https://www.orafaq.com/wiki/Number
func DecodeNumber(inputData []byte) (int64, int, error) {
	if len(inputData) == 0 {
		return 0, 0, fmt.Errorf("Invalid NUMBER")
	}
	if inputData[0] == 0x80 {
		return 0, 0, nil
	}
	var (
		negative bool
		exponent int
		mantissa int64
	)

	negative = inputData[0]&0x80 == 0
	if negative {
		exponent = int(inputData[0]^0x7f) - 64
	} else {
		exponent = int(inputData[0]&0x7f) - 64
	}

	buf := inputData[1:]
	// When negative, strip the last byte if equal 0x66
	if negative && inputData[len(inputData)-1] == 0x66 {
		buf = inputData[1 : len(inputData)-1]
	}

	// Loop on mantissa digits, stop with the capacity of int64 is reached
	mantissaDigits := 0
	for _, digit100 := range buf {
		digit100--
		if negative {
			digit100 = 100 - digit100
		}
		if protectAddFigure(&mantissa, int64(digit100/10)) {
			break
		}
		mantissaDigits++
		if protectAddFigure(&mantissa, int64(digit100%10)) {
			break
		}
		mantissaDigits++
	}

	exponent = exponent*2 - mantissaDigits // Adjust exponent to the retrieved mantissa
	if negative {
		mantissa = -mantissa
	}
	return mantissa, exponent, nil
}

// DecodeDouble decode NUMEBER as a float64
func DecodeDouble(inputData []byte) float64 {
	mantissa, exponent, err := DecodeNumber(inputData)
	if err != nil {
		return math.NaN()
	}
	return float64(mantissa) * math.Pow10(exponent)
}

// DecodeInt convert NUMBER to int64
func DecodeInt(inputData []byte) int64 {
	mantissa, exponent, err := DecodeNumber(inputData)
	if err != nil || exponent < 0 {
		return 0
	}

	for exponent > 0 {
		mantissa *= 10
		exponent--
	}
	return mantissa
}

func decodeSign(input []byte) (ret []int64, neg bool) {
	if input[0] > 0x80 {
		length := int(input[0]) - 0x80 - 0x40
		for x := 1; x < len(input); x++ {
			input[x] = input[x] - 1
		}
		input[0] = uint8(length)
		neg = false
	} else {
		length := 0xFF - int(input[0]) - 0x80 - 0x40
		if len(input) <= 20 && input[len(input)-1] == 102 {
			// fmt.Println("inside neg: ", input[:len(input)-1])
			input = input[:len(input)-1]
		}
		for x := 1; x < len(input); x++ {
			input[x] = uint8(101 - input[x])
		}
		input[0] = uint8(length)
		neg = true
	}
	ret = make([]int64, len(input))
	for x := 0; x < len(input); x++ {
		ret[x] = int64(int8(input[x]))
	}
	return
}

func EncodeDouble(num float64) ([]byte, error) {
	if num == 0.0 {
		return []byte{128}, nil
	}

	var (
		err      error
		negative bool
		exponent int
		mantissa int
	)

	// Let's the standard library doing the delicate work of converting binary float to decimal figures
	s := []byte(strconv.FormatFloat(num, 'e', -1, 64))

	if s[0] == '-' {
		negative = true
		s = s[1:]
	}

	if i := bytes.Index(s, []byte{'e'}); i >= 0 {
		exponent, err = strconv.Atoi(string(s[i+1:]))
		if err != nil {
			return nil, err
		}
		if exponent%2 != 0 {
			s = s[:i]
		} else {
			s = append([]byte("0"), s[:i]...)
		}
		if exponent < 0 {
			exponent--
		}
	}

	if i := bytes.Index(s, []byte{'.'}); i >= 0 {
		s = append(s[:i], s[i+1:]...)
	}

	mantissa = len(s)
	size := 1 + (mantissa+1)/2
	if negative && mantissa < 21 {
		size++
	}
	buf := make([]byte, size, size)

	for i := 0; i < mantissa; i += 2 {
		b := 10 * (s[i] - '0')
		if i < mantissa-1 {
			b += s[i+1] - '0'
		}
		if negative {
			b = 100 - b
		}
		buf[1+i/2] = b + 1
	}

	if negative && mantissa < 21 {
		buf[len(buf)-1] = 0x66
	}

	exponent = (exponent / 2) + 1
	if negative {
		buf[0] = byte(exponent+64) ^ 0x7f
	} else {
		buf[0] = byte(exponent+64) | 0x80
	}
	return buf, nil
}
