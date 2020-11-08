package converters

import (
	"bytes"
	"encoding/binary"
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
var factorTable = [][]float64{
	{15.0, 1e+30, 1e-30},
	{14.0, 1e+28, 1e-28},
	{13.0, 1e+26, 1e-26},
	{12.0, 1e+24, 1e-24},
	{11.0, 1e+22, 1e-22},
	{10.0, 1e+20, 1e-20},
	{9.0, 1e+18, 1e-18},
	{8.0, 1e+16, 1e-16},
	{7.0, 100000000000000.0, 1e-14},
	{6.0, 1000000000000.0, 1e-12},
	{5.0, 10000000000.0, 1e-10},
	{4.0, 100000000.0, 1e-08},
	{3.0, 1000000.0, 1e-06},
	{2.0, 10000.0, 0.0001},
	{1.0, 100.0, 0.01},
	{0.0, 1.0, 1.0},
	{-1.0, 0.01, 100.0},
	{-2.0, 0.0001, 10000.0},
	{-3.0, 1e-06, 1000000.0},
	{-4.0, 1e-08, 100000000.0},
	{-5.0, 1e-10, 10000000000.0},
	{-6.0, 1e-12, 1000000000000.0},
	{-7.0, 1e-14, 100000000000000.0},
	{-8.0, 1e-16, 1e+16},
	{-9.0, 1e-18, 1e+18},
	{-10.0, 1e-20, 1e+20},
	{-11.0, 1e-22, 1e+22},
	{-12.0, 1e-24, 1e+24},
	{-13.0, 1e-26, 1e+26},
	{-14.0, 1e-28, 1e+28},
	{-15.0, 1e-30, 1e+30},
	{-16.0, 1e-32, 1e+32},
	{-17.0, 1e-34, 1e+34},
	{-18.0, 1e-36, 1e+36},
	{-19.0, 1e-38, 1e+38},
	{-20.0, 1e-40, 1e+40},
	{-21.0, 1e-42, 1e+42},
	{-22.0, 1e-44, 1e+44},
	{-23.0, 1e-46, 1e+46},
	{-24.0, 1e-48, 1e+48},
	{-25.0, 1e-50, 1e+50},
	{-26.0, 1e-52, 1e+52},
	{-27.0, 1e-54, 1e+54},
	{-28.0, 1e-56, 1e+56},
	{-29.0, 1e-58, 1e+58},
	{-30.0, 1e-60, 1e+60},
	{-31.0, 1e-62, 1e+62},
	{-32.0, 1e-64, 1e+64},
	{-33.0, 1e-66, 1e+66},
	{-34.0, 1e-68, 1e+68},
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

//func decodeSign(input []byte) (length int, neg bool) {
//	if input[0] > 0x80 {
//		length = int(input[0]) - 0x80 - 0x40
//		for x := 1; x < len(input); x++ {
//			input[x] = input[x] - 1
//		}
//		neg = false
//	} else {
//		length = 0xFF - int(input[0]) - 0x80 - 0x40
//		if len(input) <= 20 && input[len(input)-1] == 102 {
//			input = input[:len(input)-1]
//		}
//		for x := 1; x < len(input); x++ {
//			input[x] = uint8(101 - input[x])
//		}
//		neg = true
//	}
//	return
//}

func DecodeInt(inputData []byte) int64 {
	// take a copy of input
	input := make([]byte, len(inputData))
	copy(input, inputData)
	if input[0] == 0x80 {
		return 0
	}
	data, neg := decodeSign(input)
	length := int(data[0])
	if length > len(data[1:]) {
		data = append(data, make([]int64, length-len(data[1:]))...)
	}
	data = data[1 : 1+length]
	var ret int64 = 0
	for x := 0; x < len(data); x++ {
		ret = (ret * 100) + data[x]
	}
	if neg {
		ret = ret * -1
	}
	return ret
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

// DecodeDouble decode Oracle binary representation of numbers into float64
//
// Some documentation:
//	https://gotodba.com/2015/03/24/how-are-numbers-saved-in-oracle/
//  https://www.orafaq.com/wiki/Number

//func DecodeDouble(inputData []byte) float64 {
//
//	if len(inputData) == 0 {
//		return math.NaN()
//	}
//	if inputData[0] == 0x80 {
//		return 0
//	}
//	var (
//		negative bool
//		exponent int
//		mantissa int64
//	)
//
//	negative = inputData[0]&0x80 == 0
//	if negative {
//		exponent = int(inputData[0]^0x7f) - 64
//	} else {
//		exponent = int(inputData[0]&0x7f) - 64
//	}
//
//	buf := inputData[1:]
//	// When negative, strip the last byte if equal 0x66
//	if negative && inputData[len(inputData)-1] == 0x66 {
//		buf = inputData[1 : len(inputData)-1]
//	}
//
//	for _, digit100 := range buf {
//		digit100--
//		if negative {
//			digit100 = 100 - digit100
//		}
//		mantissa *= 10
//		mantissa += int64(digit100 / 10)
//		mantissa *= 10
//		mantissa += int64(digit100 % 10)
//	}
//
//	digits := 0
//	temp64 := mantissa
//	for temp64 > 0 {
//		digits++
//		temp64 /= 100
//	}
//	exponent = (exponent - digits) * 2
//	if negative {
//		mantissa = -mantissa
//	}
//
//	ret := float64(mantissa) * math.Pow10(exponent)
//	return ret
//}

//func DecodeDouble(inputData []byte) float64 {
//	input := make([]byte, len(inputData))
//	copy(input, inputData)
//	if input[0] == 0x80 {
//		return 0
//	}
//	length, neg := decodeSign(input)
//	length -= 1
//	data := input[1:]
//	var ret float64
//	ret = float64(data[0])
//	for x := 1; x < len(data); x++ {
//		ret = ret + (float64(data[x]) / math.Pow(100.0, float64(x)))
//	}
//	if length < 0 {
//		for x := 0; x < 8; x++ {
//			if length&int(powerTable[x][0]) == 0 {
//				ret = ret / powerTable[x][1]
//				length += int(powerTable[x][0])
//			}
//		}
//		if length < 0 {
//			ret = ret / 100.0
//			length += 1
//		}
//	} else if length > 0 {
//		for x := 0; x < 8; x++ {
//			if length&int(powerTable[x][0]) > 0 {
//				ret = ret / powerTable[x][2]
//			}
//		}
//	}
//	if neg {
//		ret = ret * -1
//	}
//	fmt.Println("data: ", inputData, "\t result: ", ret)
//	return ret
//
//}

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

func DecodeDouble(inputData []byte) float64 {
	//fmt.Println(inputData)
	input := make([]byte, len(inputData))
	copy(input, inputData)
	if input[0] == 0x80 {
		return 0
	}
	data, neg := decodeSign(input)
	data[0] -= 1
	// data := append([]byte{uint8(length)}, input[1:]...)
	flag2 := data[1] < 10
	num2 := 15  //  factorTable[0][0]; // 15
	num3 := -15 //  factorTable[0][0] - float64(50 - 20); // -15
	index2 := 0
	num4 := 0
	index1 := 1
	length := int(data[0])
	if length > num2 || length < num3 {
		if length > num2 {
			index2 = -1
			num4 = length - num2
		} else {
			index2 = 50 - 20 - 1
			num4 = length - num3
		}

	} else {
		index2 = int(num2-length) - 1
		num4 = 0
	}
	num5 := len(data) - 1
	flag1 := false
	if data[1] < 10 && num5 > 8 || data[1] >= 10 && num5 >= 8 {
		num5 = 8
		flag1 = true
	}
	num6 := 0.0
	switch num5 % 4 {
	case 1:
		index2++
		if factorTable[index2][1] >= 1.0 {
			num6 = float64(data[1]) * factorTable[index2][1]
		} else {
			num6 = float64(data[1]) * factorTable[index2][2]
		}
		index1++
		num5--
	case 2:
		num8 := data[1]*100 + data[2]
		index2 += 2
		if factorTable[index2][1] >= 1.0 {
			num6 = float64(num8) * factorTable[index2][1]
		} else {
			num6 = float64(num8) / factorTable[index2][2]
		}
		index1 += 2
		num5 -= 2
	case 3:
		num9 := (data[1]*100+data[2])*100 + data[3]
		index2 += 3
		if factorTable[index2][1] >= 1.0 {
			num6 = float64(num9) * factorTable[index2][1]
		} else {
			num6 = float64(num9) * factorTable[index2][2]
		}
		index1 += 3
		num5 -= 3
	default:
		num6 = 0.0
	}
	for num5 > 0 {
		num10 := ((data[index1]*100+data[index1+1])*100+data[index1+2])*100 + data[index1+3]
		index2 += 4
		if factorTable[index2][1] < 1.0 {
			num6 += float64(num10) / factorTable[index2][2]
		} else {
			num6 += float64(num10) * factorTable[index2][1]
		}
		index1 += 4
		num5 -= 4
	}
	if flag1 {
		if flag2 {
			if data[index1] > 50 {
				num10 := 1
				num6 += float64(num10) * factorTable[index2][1]
			}
		} else {
			index3 := index1 - 1
			var num10 int64 = 0
			if data[index3]%10 < 5 {
				num10 = ((data[index3] / 10) * 10) + data[index3]
			} else {
				num10 = ((data[index3]/10 + 1) * 10) - data[index3]
			}
			num6 += float64(num10) * factorTable[index2][1]
		}
	}

	if num4 != 0 {
		index3 := 0
		for num4 > 0 {
			if int(powerTable[index3][0]) <= num4 {
				num4 -= int(powerTable[index3][0])
				num6 *= powerTable[index3][1]
			}
			index3++
		}
		for num4 < 0 {
			if int(powerTable[index3][0]) <= -num4 {
				num4 += int(powerTable[index3][0])
				num6 *= powerTable[index3][2]
			}
			index3++
		}
	}
	if neg {
		num6 = num6 * -1
	}
	return num6
}
