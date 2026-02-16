package types

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/sijms/go-ora/v3/utils"
)

type Number struct {
	Data []byte
}

func (number *Number) isZero() bool {
	return len(number.Data) > 0 && number.Data[0] == 0x80
}

func (number *Number) isPositive() bool {
	return len(number.Data) > 0 && number.Data[0]&0x80 != 0
}

func NewNumberFromString(val string) (*Number, error) {
	mantissa := []byte(val)
	negative := mantissa[0] == '-'
	if negative {
		mantissa = mantissa[1:]
	}
	var (
		exp int
		err error
	)
	isFloat := false
	if i := bytes.Index(mantissa, []byte{'e'}); i >= 0 {
		exp, err = strconv.Atoi(string(mantissa[i+1:]))
		if err != nil {
			return nil, err
		}
		mantissa = mantissa[:i]
	}
	if i := bytes.Index(mantissa, []byte{'.'}); i >= 0 {
		mantissa = append(mantissa[:i], mantissa[i+1:]...)
		exp += i - 1
		isFloat = true
	}
	if !isFloat {
		exp += len(mantissa) - 1
	}
	ret := new(Number)
	err = ret.encode(mantissa, exp, negative)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func NewNumberFromInt64(val int64) (*Number, error) {
	mantissa := []byte(strconv.FormatInt(val, 10))
	negative := mantissa[0] == '-'
	if negative {
		mantissa = mantissa[1:]
	}
	exp := len(mantissa) - 1
	ret := new(Number)
	err := ret.encode(mantissa, exp, negative)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func NewNumberFromUint64(val uint64) (*Number, error) {
	mantissa := []byte(strconv.FormatUint(val, 10))
	exponent := len(mantissa) - 1
	ret := new(Number)
	err := ret.encode(mantissa, exponent, false)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func NewNumberFromFloat(val float64) (*Number, error) {
	if val == 0.0 {
		return &Number{Data: []byte{128}}, nil
	}
	var (
		exponent int
		err      error
	)
	mantissa := []byte(strconv.FormatFloat(val, 'e', -1, 64))
	if i := bytes.Index(mantissa, []byte{'e'}); i >= 0 {
		exponent, err = strconv.Atoi(string(mantissa[i+1:]))
		if err != nil {
			return nil, err
		}
		mantissa = mantissa[:i]
	}

	negative := mantissa[0] == '-'
	if negative {
		mantissa = mantissa[1:]
	}

	if i := bytes.Index(mantissa, []byte{'.'}); i >= 0 {
		mantissa = append(mantissa[:i], mantissa[i+1:]...)
	}
	ret := new(Number)
	err = ret.encode(mantissa, exponent, negative)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (number *Number) encode(mantissa []byte, exp int, negative bool) error {
	trailingZeros := 0
	for i := len(mantissa) - 1; i >= 0 && mantissa[i] == '0'; i-- {
		trailingZeros++
	}
	mantissa = mantissa[:len(mantissa)-trailingZeros]
	if len(mantissa) == 0 {
		number.Data = []byte{0x80}
		return nil
	}
	if exp%2 == 0 {
		mantissa = append([]byte{'0'}, mantissa...)
	}
	mantissaLen := len(mantissa)
	size := 1 + (mantissaLen+1)/2
	if negative && mantissaLen < 21 {
		size++
	}
	number.Data = make([]byte, size)

	for i := 0; i < mantissaLen; i += 2 {
		b := 10 * (mantissa[i] - '0')
		if i < mantissaLen-1 {
			b += mantissa[i+1] - '0'
		}
		if negative {
			b = 100 - b
		}
		number.Data[1+i/2] = b + 1
	}

	if negative && mantissaLen < 21 {
		number.Data[len(number.Data)-1] = 0x66
	}

	if exp < 0 {
		exp--
	}
	exp = (exp / 2) + 1
	if negative {
		number.Data[0] = byte(exp+64) ^ 0x7f
	} else {
		number.Data[0] = byte(exp+64) | 0x80
	}
	return nil
}

func (number *Number) decode() (strNum string, exp int, negative bool, err error) {
	if len(number.Data) == 0 {
		err = fmt.Errorf("invalid NUMBER")
		return
	}
	if number.isZero() {
		strNum = "0"
		return
	}
	negative = number.Data[0]&0x80 == 0
	if negative {
		exp = int(number.Data[0]^0x7F) - 64
	} else {
		exp = int(number.Data[0]&0x7F) - 64
	}

	if _isPosInf(number.Data) || _isNegInf(number.Data) {
		strNum = "Infinity"
		exp = 0
		return
	}

	buf := number.Data[1:]
	if len(buf) == 0 {
		err = fmt.Errorf("invalid NUMBER")
		return
	}
	if negative && buf[len(buf)-1] == 0x66 {
		buf = buf[:len(buf)-1]
	}
	var output []byte
	for _, digit := range buf {
		digit--
		if negative {
			digit = 100 - digit
		}
		output = append(output, (digit/10)+'0')
		output = append(output, (digit%10)+'0')
	}
	exp = exp*2 - len(output)
	strNum = string(output)
	return
}

func _isNegInf(b []byte) bool {
	return b[0] == 0 && len(b) == 1
}

func _isPosInf(b []byte) bool {
	// -1 =255
	return len(b) == 2 && b[0] == 255 && b[1] == 101
}

func (number *Number) Int64() (int64, error) {
	strNum, exp, negative, err := number.decode()
	if err != nil {
		return 0, err
	}
	mantissa, err := strconv.ParseInt(strNum, 10, 64)
	if err != nil {
		return 0, err
	}
	for exp > 0 {
		mantissa *= 10
		exp--
	}
	if negative && (mantissa>>63) == 0 {
		return -mantissa, nil
	}
	return mantissa, nil
}

func (number *Number) Uint64() (uint64, error) {
	strNum, exp, _, err := number.decode()
	if err != nil {
		return 0, err
	}
	mantissa, err := strconv.ParseUint(strNum, 10, 64)
	if err != nil {
		return 0, err
	}
	for exp > 0 {
		mantissa *= 10
		exp--
	}
	return mantissa, nil
}

func (number *Number) Float64() (float64, error) {
	strNum, exp, negative, err := number.decode()
	if err != nil {
		return 0, err
	}
	mantissa, err := strconv.ParseFloat(strNum, 64)
	if err != nil {
		return 0, err
	}
	absExponent := int(math.Abs(float64(exp)))
	if negative {
		return -math.Round(mantissa*math.Pow10(exp)*math.Pow10(absExponent)) / math.Pow10(absExponent), nil
	}
	return math.Round(mantissa*math.Pow10(exp)*math.Pow10(absExponent)) / math.Pow10(absExponent), nil
}

func (number *Number) String() (string, error) {
	strNum, exp, negative, err := number.decode()
	if err != nil {
		return "", err
	}
	// remove zeros from beginning
	if len(strNum) > 1 {
		strNum = strings.TrimLeft(strNum, "0")
	}

	if exp > 0 {
		strNum += strings.Repeat("0", exp)
	} else if exp < 0 {
		pos := len(strNum) + exp // exp is negative
		if pos < 0 {
			pos = -pos
			strNum = strings.Repeat("0", pos) + strNum
			pos = 0
		}
		strNum = strNum[:pos] + "." + strNum[pos:]
		// remove zeros at right
		strNum = strings.TrimRight(strNum, "0")
	}
	if strNum[0] == '.' {
		strNum = "0" + strNum
	}
	if negative {
		strNum = "-" + strNum
	}
	return strNum, nil
}

func NewNumber(n interface{}) (*Number, error) {
	var err error
	n, err = utils.GetValue(n)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, nil
	}
	rType := reflect.TypeOf(n)
	rValue := reflect.ValueOf(n)
	if utils.IsSigned(rType) {
		return NewNumberFromInt64(rValue.Int())
	}
	if utils.IsUnsigned(rType) {
		return NewNumberFromUint64(rValue.Uint())
	}
	//if f32, ok := col.(float32); ok {
	//	return strconv.ParseFloat(fmt.Sprint(f32), 64)
	//}
	if utils.IsFloat(rType) {
		return NewNumberFromFloat(rValue.Float())
	}
	if rType == reflect.TypeOf((*Number)(nil)).Elem() {
		if num, ok := n.(Number); ok {
			return &num, nil
		}
		return nil, errors.New("conversion of unsupported type to number")
	}
	if rType == utils.TyBytes {
		return &Number{Data: rValue.Bytes()}, nil
	}
	switch rType.Kind() {
	case reflect.Bool:
		if rValue.Bool() {
			return NewNumberFromInt64(1)
		}

		return NewNumberFromInt64(0)
	case reflect.String:
		return NewNumberFromString(rValue.String())
	default:
		return nil, errors.New("conversion of unsupported type to number")
	}
}

//func (number *Number) Encode() ([]byte, error) {
//	return number.Data, nil
//}
//func (number *Number) Decode(data []byte, _ uint16) (interface{}, error) {
//	if data == nil {
//		return nil, nil
//	}
//	ret := Number{Data: data}
//	ret.SetTypeInfo(number.TypeInfo)
//	return ret.String()
//}
//func (number *Number) Read(session network.SessionReader, tnsType uint16, _ bool) (interface{}, error) {
//	bValue, err := number.basicRead(session)
//	if err != nil {
//		return nil, err
//	}
//	return number.Decode(bValue, tnsType)
//}
//func (number *Number) Write(session network.SessionWriter) error {
//	bValue, err := number.Encode()
//	if err != nil {
//		return err
//	}
//	session.PutClr(bValue)
//	return nil
//}

func (number *Number) Scan(value interface{}) error {
	temp, err := NewNumber(value)
	if err != nil {
		return err
	}
	if temp == nil {
		number.Data = nil
		return nil
	}
	number.Data = temp.Data
	return nil
}

//func (number *Number) Value() (driver.Value, error) {
//	return number.Int64()
//}
