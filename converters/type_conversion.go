package converters

import (
	"bytes"
	"errors"
	"math"
	"time"
)

var powerTable = [][]float64{
	{128.0, 1e+256, 1e-256},
	{64.0, 1e+128, 1e-128},
	{32.0, 1e+64, 1e-64},
	{16.0, 1e+32, 1e-32},
	{8.0, 1e+16, 1e-16},
	{4.0, 100000000.0, 1e-08},
	{2.0, 10000.0, 0.0001},
	{1.0, 100.0, 0.01},
}

// note that int64 is different in size > 22 byte size and when make clear type you will path 1 instead of 0
//case OraType.ORA_NUMBER:
//case OraType.ORA_VARNUM:
//if (numArray.Length >= 22)
//{
//mEngine.MarshalCLR(numArray, 1, (int) numArray[0]);
//break;
//}
//mEngine.MarshalCLR(numArray, 0, numArray.Length);
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
func decodeSign(input []byte) (length int, neg bool) {
	if input[0] > 0x80 {
		length = int(input[0]) - 0x80 - 0x40
		for x := 1; x < len(input); x++ {
			input[x] = input[x] - 1
		}
		neg = false
	} else {
		length = 0xFF - int(input[0]) - 0x80 - 0x40
		if len(input) <= 20 && input[len(input)-1] == 102 {
			input = input[:len(input)-1]
		}
		for x := 1; x < len(input); x++ {
			input[x] = uint8(101 - input[x])
		}
		neg = true
	}
	return
}
func DecodeInt(inputData []byte) int {
	// take a copy of input
	input := make([]byte, len(inputData))
	copy(input, inputData)
	if input[0] == 0x80 {
		return 0
	}
	length, neg := decodeSign(input)
	if length > len(input[1:]) {
		input = append(input, make([]byte, length-len(input[1:]))...)
	}
	data := input[1 : 1+length]
	ret := 0
	for x := 0; x < len(data); x++ {
		ret = (ret * 100) + int(data[x])
	}
	if neg {
		return ret * -1
	} else {
		return ret
	}
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
		dateTmp := data[7:10]
		switch len(dateTmp) {
		case 0:
			nanoSec = 0
		case 1:
			nanoSec = int(uint32(dateTmp[0]))
		case 2:
			nanoSec = int(uint32(dateTmp[0]) | uint32(dateTmp[1])<<8)
		case 3:
			nanoSec = int(uint32(dateTmp[0]) | uint32(dateTmp[1])<<8 | uint32(dateTmp[2])<<16)
		default:
			nanoSec = int(uint32(dateTmp[0]) | uint32(dateTmp[1])<<8 | uint32(dateTmp[2])<<16 | uint32(dateTmp[3])<<24)
		}
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

func DecodeDouble(inputData []byte) float64 {
	input := make([]byte, len(inputData))
	copy(input, inputData)
	if input[0] == 0x80 {
		return 0
	}
	length, neg := decodeSign(input)
	length -= 1
	data := input[1:]
	var ret float64
	ret = float64(data[0])
	for x := 1; x < len(data); x++ {
		ret = ret + (float64(data[x]) / math.Pow(100.0, float64(x)))
	}
	if length < 0 {
		for x := 0; x < 8; x++ {
			if length&int(powerTable[x][0]) == 0 {
				ret = ret / powerTable[x][1]
				length += int(powerTable[x][0])
			}
		}
		if length < 0 {
			ret = ret / 100.0
			length += 1
		}
	} else if length > 0 {
		for x := 0; x < 8; x++ {
			if length&int(powerTable[x][0]) > 0 {
				ret = ret / powerTable[x][2]
			}
		}
	}
	if neg {
		ret = ret * -1
	}
	return ret

}

func EncodeDouble(num float64) ([]byte, error) {
	//byte[] numArray = new byte[20];
	num1 := 0
	neg := num < 0.0
	num = math.Abs(num)
	if num < 1.0 {
		for x := 0; x < 8; x++ {
			if powerTable[x][2] >= num {
				num1 -= int(powerTable[x][0])
				num *= powerTable[x][1]
			}
		}
		if num < 1.0 {
			num1--
			num *= 100.0
		}
	} else {
		for x := 0; x < 8; x++ {
			if powerTable[x][1] <= num {
				num1 += int(powerTable[x][0])
				num *= powerTable[x][2]
			}
		}
	}
	if num1 > 62 || num1 < -65 {
		return nil, errors.New("overflow occur")
	}
	flag := num < 10.0
	ret := make([]byte, 20)
	num3 := uint8(num)
	for x := 0; x < 8; x++ {
		ret[x] = num3
		num = (num - float64(num3)) * 100.0
		num3 = uint8(num)
	}
	if flag {
		if int(num3) >= 50 {
			ret[7]++
		}
	} else {
		ret[7] = uint8(((int(ret[7]) + 5) / 10) * 10)
		if num1 == 62 && ((int(ret[7])+5)/10)*10 == 100 {
			ret[7] = uint8(((int(ret[7]) - 5) / 10) * 10)
		}
	}
	x := 7
	for ret[x] == 100 {
		if x == 0 {
			num1++
			ret[x] = 1
			break
		}
		ret[x] = 0
		x--
		ret[x]++
	}
	ret = bytes.TrimRight(ret, "\x00")
	ret = append([]byte{uint8(num1)}, ret...)
	return encodeSign(ret, neg), nil
	// 192, 2, 79
}
