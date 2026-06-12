package types

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/sijms/go-ora/v3/utils"
)

type Number struct {
	bValue []byte
}

func (number *Number) isZero() bool {
	return len(number.bValue) > 0 && number.bValue[0] == 0x80
}

func (number *Number) isPositive() bool {
	return len(number.bValue) > 0 && number.bValue[0]&0x80 != 0
}

//func NewNumberFromString(val string) (*Number, error) {
//	mantissa := []byte(val)
//	negative := mantissa[0] == '-'
//	if negative {
//		mantissa = mantissa[1:]
//	}
//	var (
//		exp int
//		err error
//	)
//	isFloat := false
//	if i := bytes.Index(mantissa, []byte{'e'}); i >= 0 {
//		exp, err = strconv.Atoi(string(mantissa[i+1:]))
//		if err != nil {
//			return nil, err
//		}
//		mantissa = mantissa[:i]
//	}
//	if i := bytes.Index(mantissa, []byte{'.'}); i >= 0 {
//		mantissa = append(mantissa[:i], mantissa[i+1:]...)
//		exp += i - 1
//		isFloat = true
//	}
//	if !isFloat {
//		exp += len(mantissa) - 1
//	}
//	ret := new(Number)
//	err = ret.encode(mantissa, exp, negative)
//	if err != nil {
//		return nil, err
//	}
//	return ret, nil
//}
//func NewNumberFromInt64(val int64) (*Number, error) {
//	mantissa := []byte(strconv.FormatInt(val, 10))
//	negative := mantissa[0] == '-'
//	if negative {
//		mantissa = mantissa[1:]
//	}
//	exp := len(mantissa) - 1
//	ret := new(Number)
//	err := ret.encode(mantissa, exp, negative)
//	if err != nil {
//		return nil, err
//	}
//	return ret, nil
//}
//func NewNumberFromUint64(val uint64) (*Number, error) {
//	mantissa := []byte(strconv.FormatUint(val, 10))
//	exponent := len(mantissa) - 1
//	ret := new(Number)
//	err := ret.encode(mantissa, exponent, false)
//	if err != nil {
//		return nil, err
//	}
//	return ret, nil
//}
//func NewNumberFromFloat(val float64) (*Number, error) {
//	if val == 0.0 {
//		return &Number{bValue: []byte{128}}, nil
//	}
//	var (
//		exponent int
//		err      error
//	)
//	mantissa := []byte(strconv.FormatFloat(val, 'e', -1, 64))
//	if i := bytes.Index(mantissa, []byte{'e'}); i >= 0 {
//		exponent, err = strconv.Atoi(string(mantissa[i+1:]))
//		if err != nil {
//			return nil, err
//		}
//		mantissa = mantissa[:i]
//	}
//
//	negative := mantissa[0] == '-'
//	if negative {
//		mantissa = mantissa[1:]
//	}
//
//	if i := bytes.Index(mantissa, []byte{'.'}); i >= 0 {
//		mantissa = append(mantissa[:i], mantissa[i+1:]...)
//	}
//	ret := new(Number)
//	err = ret.encode(mantissa, exponent, negative)
//	if err != nil {
//		return nil, err
//	}
//	return ret, nil
//}

func (number *Number) encode(mantissa []byte, exp int, negative bool) ([]byte, error) {
	trailingZeros := 0
	for i := len(mantissa) - 1; i >= 0 && mantissa[i] == '0'; i-- {
		trailingZeros++
	}
	mantissa = mantissa[:len(mantissa)-trailingZeros]
	if len(mantissa) == 0 {
		return []byte{0x80}, nil
	}
	if exp%2 == 0 {
		mantissa = append([]byte{'0'}, mantissa...)
	}
	mantissaLen := len(mantissa)
	size := 1 + (mantissaLen+1)/2
	if negative && mantissaLen < 21 {
		size++
	}
	bValue := make([]byte, size)

	for i := 0; i < mantissaLen; i += 2 {
		b := 10 * (mantissa[i] - '0')
		if i < mantissaLen-1 {
			b += mantissa[i+1] - '0'
		}
		if negative {
			b = 100 - b
		}
		bValue[1+i/2] = b + 1
	}

	if negative && mantissaLen < 21 {
		bValue[len(bValue)-1] = 0x66
	}

	if exp < 0 {
		exp--
	}
	exp = (exp / 2) + 1
	if negative {
		bValue[0] = byte(exp+64) ^ 0x7f
	} else {
		bValue[0] = byte(exp+64) | 0x80
	}
	return bValue, nil
}

func (number *Number) decode() (strNum string, exp int, negative bool, err error) {
	if len(number.bValue) == 0 {
		err = fmt.Errorf("invalid NUMBER")
		return
	}
	if number.isZero() {
		strNum = "0"
		return
	}
	negative = number.bValue[0]&0x80 == 0
	if negative {
		exp = int(number.bValue[0]^0x7F) - 64
	} else {
		exp = int(number.bValue[0]&0x7F) - 64
	}

	if _isPosInf(number.bValue) || _isNegInf(number.bValue) {
		strNum = "Infinity"
		exp = 0
		return
	}

	buf := number.bValue[1:]
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

func NewBinaryDouble(number float64) *Number {
	ret := new(Number)
	_ = ret.encodeFloat64(number)
	return ret
}

func NewBinaryFloat(number float32) *Number {
	ret := new(Number)
	_ = ret.encodeFloat32(number)
	return ret
}
func NewNumber(n interface{}) (*Number, error) {
	var err error
	ret := new(Number)
	err = ret.SetValue(n, NUMBER)
	return ret, err
	//n, err = utils.GetValue(n)
	//if err != nil {
	//	return nil, err
	//}
	//if n == nil {
	//	return nil, nil
	//}
	//rType := reflect.TypeOf(n)
	//rValue := reflect.ValueOf(n)
	//if utils.IsSigned(rType) {
	//	return NewNumberFromInt64(rValue.Int())
	//}
	//if utils.IsUnsigned(rType) {
	//	return NewNumberFromUint64(rValue.Uint())
	//}
	////if f32, ok := col.(float32); ok {
	////	return strconv.ParseFloat(fmt.Sprint(f32), 64)
	////}
	//if utils.IsFloat(rType) {
	//	return NewNumberFromFloat(rValue.Float())
	//}
	//if rType == reflect.TypeOf((*Number)(nil)).Elem() {
	//	if num, ok := n.(Number); ok {
	//		return &num, nil
	//	}
	//	return nil, errors.New("conversion of unsupported type to number")
	//}
	//if rType == utils.TyBytes {
	//	return &Number{bValue: rValue.Bytes()}, nil
	//}
	//switch rType.Kind() {
	//case reflect.Bool:
	//	if rValue.Bool() {
	//		return NewNumberFromInt64(1)
	//	}
	//
	//	return NewNumberFromInt64(0)
	//case reflect.String:
	//	return NewNumberFromString(rValue.String())
	//default:
	//	return nil, errors.New("conversion of unsupported type to number")
	//}
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

//func (number *Number) Encode(input interface{}, typeId uint16) ([]byte, error) {
//	var temp *Number
//	var err error
//	switch typeId {
//	case IBDOUBLE:
//		switch number := input.(type) {
//		case float64:
//			temp = NewBinaryDouble(number)
//		case float32:
//			temp = NewBinaryDouble(float64(number))
//		default:
//			temp, err = NewNumber(input)
//		}
//	case IBFLOAT:
//		switch number := input.(type) {
//		case float64:
//			temp = NewBinaryFloat(float32(number))
//		case float32:
//			temp = NewBinaryFloat(number)
//		default:
//			temp, err = NewNumber(input)
//		}
//	default:
//		temp, err = NewNumber(input)
//	}
//	if err != nil {
//		return nil, err
//	}
//	return temp.bValue, nil
//}
//func (number *Number) Decode(input []byte, typeId uint16) (interface{}, error) {
//	number.bValue = input
//	switch typeId {
//	case IBDOUBLE:
//		return number.decodeDouble()
//	case IBFLOAT:
//		return number.decodeFloat()
//	}
//	return number.String()
//}

func (number *Number) decodeDouble() (float64, error) {
	if len(number.bValue) < 8 {
		return 0, fmt.Errorf("error decoding binary double, supplied buffer length: %d and required length: %d", len(number.bValue), 8)
	}
	if number.bValue[0]&128 != 0 {
		number.bValue[0] = number.bValue[0] & 127
	} else {
		xorBuffer(number.bValue, 8)
	}
	return math.Float64frombits(binary.BigEndian.Uint64(number.bValue)), nil
}

func (number *Number) decodeFloat() (float32, error) {
	if len(number.bValue) < 4 {
		return 0, fmt.Errorf("error decoding binary float, supplied buffer length: %d and required length: %d", len(number.bValue), 4)
	}
	if number.bValue[0]&128 != 0 {
		number.bValue[0] = number.bValue[0] & 127
	} else {
		xorBuffer(number.bValue, 4)
	}
	return math.Float32frombits(binary.BigEndian.Uint32(number.bValue)), nil
}

func (number *Number) Value(typeId uint16) (interface{}, error) {
	if len(number.bValue) == 0 {
		return nil, nil
	}
	switch typeId {
	case IBDOUBLE:
		return number.decodeDouble()
	case IBFLOAT:
		return number.decodeFloat()
	default:
		return number.String()
	}
}

func (number *Number) Bytes() []byte {
	return number.bValue
}

func (number *Number) encodeInt(input int64) error {
	mantissa := []byte(strconv.FormatInt(input, 10))
	negative := mantissa[0] == '-'
	if negative {
		mantissa = mantissa[1:]
	}
	var err error
	exp := len(mantissa) - 1
	number.bValue, err = number.encode(mantissa, exp, negative)
	return err
	//ret := new(Number)
	//err := ret.encode(mantissa, exp, negative)
	//if err != nil {
	//	return nil, err
	//}
	//return ret, nil
}
func (number *Number) encodeUint(input uint64) error {
	mantissa := []byte(strconv.FormatUint(input, 10))
	exponent := len(mantissa) - 1
	var err error
	number.bValue, err = number.encode(mantissa, exponent, false)
	return err
}
func (number *Number) encodeFloat64(input float64) error {
	number.bValue = make([]byte, 8)
	temp := math.Float64bits(input)
	binary.BigEndian.PutUint64(number.bValue, temp)
	if input > 0 {
		number.bValue[0] = number.bValue[0] | 128
	} else {
		xorBuffer(number.bValue, 8)
	}
	return nil
}
func (number *Number) encodeFloat32(input float32) error {
	number.bValue = make([]byte, 4)
	temp := math.Float32bits(input)
	binary.BigEndian.PutUint32(number.bValue, temp)
	if input > 0 {
		number.bValue[0] = number.bValue[0] | 128
	} else {
		xorBuffer(number.bValue, 4)
	}
	return nil
}
func (number *Number) encodeFloat(input float64) error {
	if input == 0.0 {
		number.bValue = []byte{128}
		return nil
	}
	var (
		exponent int
		err      error
	)
	mantissa := []byte(strconv.FormatFloat(input, 'e', -1, 64))
	if i := bytes.Index(mantissa, []byte{'e'}); i >= 0 {
		exponent, err = strconv.Atoi(string(mantissa[i+1:]))
		if err != nil {
			return err
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
	number.bValue, err = number.encode(mantissa, exponent, negative)
	return err
}
func (number *Number) encodeString(input string) error {
	mantissa := []byte(input)
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
			return err
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
	number.bValue, err = number.encode(mantissa, exp, negative)
	return err
}
func (number *Number) SetValue(input interface{}, typeId uint16) error {
	var err error
	input, err = utils.GetValue(input)
	if err != nil {
		return err
	}
	if input == nil {
		number.bValue = nil
		return nil
	}
	rType := reflect.TypeOf(input)
	rValue := reflect.ValueOf(input)

	if rType == TyNumber {
		if num, ok := input.(Number); ok {
			number.bValue = num.bValue
		} else {
			return fmt.Errorf("conversion of unsupported type %T to number", input)
		}
		return nil
	}
	var tempValue interface{}
	//if utils.IsSigned(rType) {
	//
	//	//return number.encodeInt()
	//}
	//if utils.IsUnsigned(rType) {
	//	tempValue = rValue.Uint()
	//	//return number.encodeUint()
	//}
	//if utils.IsFloat(rType) {
	//	tempValue = rValue.Float()
	//	//return number.encodeFloat()
	//}
	switch rType {
	case TyInt, TyInt8, TyInt16, TyInt32, TyInt64:
		tempValue = rValue.Int()
	case TyUint, TyUint8, TyUint16, TyUint32, TyUint64:
		tempValue = rValue.Uint()
	case TyFloat32, TyFloat64:
		tempValue = rValue.Float()
	case TyString:
		tempValue = rValue.String()
	case TyNullByte:
		temp := input.(sql.NullByte)
		if !temp.Valid {
			number.bValue = nil
			return nil
		}
		tempValue = int64(temp.Byte)
	case TyNullInt16:
		temp := input.(sql.NullInt16)
		if !temp.Valid {
			number.bValue = nil
			return nil
		}
		tempValue = int64(temp.Int16)
	case TyNullInt32:
		temp := input.(sql.NullInt32)
		if !temp.Valid {
			number.bValue = nil
			return nil
		}
		tempValue = int64(temp.Int32)
	case TyNullInt64:
		temp := input.(sql.NullInt64)
		if !temp.Valid {
			number.bValue = nil
			return nil
		}
		tempValue = temp.Int64
	case TyNullFloat64:
		temp := input.(sql.NullFloat64)
		if !temp.Valid {
			number.bValue = nil
			return nil
		}
		tempValue = temp.Float64
	default:
		return fmt.Errorf("conversion of unsupported type %T to number", input)
	}
	switch temp := tempValue.(type) {
	case int64:
		switch typeId {
		case IBFLOAT:
			err = number.encodeFloat32(float32(temp))
		case IBDOUBLE:
			err = number.encodeFloat64(float64(temp))
		default:
			err = number.encodeInt(temp)
		}
	case uint64:
		switch typeId {
		case IBFLOAT:
			err = number.encodeFloat32(float32(temp))
		case IBDOUBLE:
			err = number.encodeFloat64(float64(temp))
		default:
			err = number.encodeUint(temp)
		}
	case float64:
		switch typeId {
		case IBFLOAT:
			err = number.encodeFloat32(float32(temp))
		case IBDOUBLE:
			err = number.encodeFloat64(temp)
		default:
			err = number.encodeFloat(temp)
		}
	case string:
		switch typeId {
		case IBFLOAT:
			value, err := strconv.ParseFloat(rValue.String(), 64)
			if err != nil {
				return err
			}
			err = number.encodeFloat32(float32(value))
		case IBDOUBLE:
			value, err := strconv.ParseFloat(rValue.String(), 64)
			if err != nil {
				return err
			}
			err = number.encodeFloat64(value)
		default:
			err = number.encodeString(temp)
		}
	default:
		err = fmt.Errorf("conversion of unsupported type %T to number", temp)
	}
	return err

	//switch rType {
	//case TyString:
	//
	//case TyNullByte:
	//	temp := input.(sql.NullByte)
	//	if temp.Valid {
	//		err = number.encodeFloat(float64(temp.Byte))
	//	} else {
	//		number.bValue = nil
	//	}
	//case TyNullInt16:
	//	temp := input.(sql.NullInt16)
	//	if temp.Valid {
	//		err = number.encodeFloat(float64(temp.Int16))
	//	} else {
	//		number.bValue = nil
	//	}
	//case TyNullInt32:
	//	temp := input.(sql.NullInt32)
	//	if temp.Valid {
	//		err = number.encodeFloat(float64(temp.Int32))
	//	} else {
	//		number.bValue = nil
	//	}
	//case TyNullInt64:
	//	temp := input.(sql.NullInt64)
	//	if temp.Valid {
	//		err = number.encodeFloat(float64(temp.Int64))
	//	} else {
	//		number.bValue = nil
	//	}
	//case TyNullFloat64:
	//	temp := input.(sql.NullFloat64)
	//	if temp.Valid {
	//		err = number.encodeFloat(temp.Float64)
	//	} else {
	//		number.bValue = nil
	//	}
	//
	//default:
	//	err = fmt.Errorf("conversion of unsupported type: %v to binary double", rType)
	//}
	//return err
	//
	//switch typeId {
	//case IBDOUBLE:
	//
	//case IBFLOAT:
	//	if rType == TyFloat32 {
	//		if value, ok := input.(float32); ok {
	//			err = number.encodeFloat32(value)
	//		} else {
	//			err = fmt.Errorf("conversion of unsupported type: %v to binary float", rType)
	//		}
	//	} else if rType == TyFloat64 {
	//		err = number.encodeFloat32(float32(rValue.Float()))
	//	} else if rType == TyString {
	//		value, err := strconv.ParseFloat(rValue.String(), 32)
	//		if err != nil {
	//			return err
	//		}
	//		err = number.encodeFloat32(float32(value))
	//	} else {
	//		err = fmt.Errorf("conversion of unsupported type: %v to binary float", rType)
	//	}
	//	return err
	//}
	//if rType == utils.TyBytes {
	//	return &Number{bValue: rValue.Bytes()}, nil
	//}
	//switch rType.Kind() {
	//case reflect.Bool:
	//	if rValue.Bool() {
	//		err = number.encodeInt(1)
	//	} else {
	//		err = number.encodeInt(0)
	//	}
	//	if err != nil {
	//		return err
	//	}
	//case reflect.String:
	//	err = number.encodeString(rValue.String())
	//default:
	//	return errors.New("conversion of unsupported type to number")
	//}
	//return nil
}

func (number *Number) SetBytes(input []byte) {
	number.bValue = input
}

func (number *Number) Scan(value interface{}) error {
	return number.SetValue(value, NUMBER)
	//temp, err := NewNumber(value)
	//if err != nil {
	//	return err
	//}
	//if temp == nil {
	//	number.bValue = nil
	//	return nil
	//}
	//number.bValue = temp.bValue
	//return nil
}

func (number *Number) CopyTo(dest interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return errors.New("dest must be a pointer")
	}

	if number.bValue == nil {
		destValue.Elem().Set(reflect.Zero(destValue.Elem().Type()))
		return nil
	}

	strVal, err := number.String()
	if err != nil {
		return err
	}

	switch dst := dest.(type) {
	case *string:
		*dst = strVal
	case *int:
		v, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return err
		}
		*dst = int(v)
	case *int8:
		v, err := strconv.ParseInt(strVal, 10, 8)
		if err != nil {
			return err
		}
		*dst = int8(v)
	case *int16:
		v, err := strconv.ParseInt(strVal, 10, 16)
		if err != nil {
			return err
		}
		*dst = int16(v)
	case *int32:
		v, err := strconv.ParseInt(strVal, 10, 32)
		if err != nil {
			return err
		}
		*dst = int32(v)
	case *int64:
		v, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return err
		}
		*dst = v
	case *uint:
		v, err := strconv.ParseUint(strVal, 10, 64)
		if err != nil {
			return err
		}
		*dst = uint(v)
	case *uint8:
		v, err := strconv.ParseUint(strVal, 10, 8)
		if err != nil {
			return err
		}
		*dst = uint8(v)
	case *uint16:
		v, err := strconv.ParseUint(strVal, 10, 16)
		if err != nil {
			return err
		}
		*dst = uint16(v)
	case *uint32:
		v, err := strconv.ParseUint(strVal, 10, 32)
		if err != nil {
			return err
		}
		*dst = uint32(v)
	case *uint64:
		v, err := strconv.ParseUint(strVal, 10, 64)
		if err != nil {
			return err
		}
		*dst = v
	case *float32:
		v, err := strconv.ParseFloat(strVal, 32)
		if err != nil {
			return err
		}
		*dst = float32(v)
	case *float64:
		v, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return err
		}
		*dst = v
	case *sql.NullString:
		dst.String = strVal
		dst.Valid = true
	case *sql.NullByte:
		v, err := strconv.ParseUint(strVal, 10, 8)
		if err != nil {
			return err
		}
		dst.Valid = true
		dst.Byte = uint8(v)
	case *sql.NullInt64:
		v, err := strconv.ParseInt(strVal, 10, 64)
		if err != nil {
			return err
		}
		dst.Int64 = v
		dst.Valid = true
	case *sql.NullFloat64:
		v, err := strconv.ParseFloat(strVal, 64)
		if err != nil {
			return err
		}
		dst.Float64 = v
		dst.Valid = true
	case *Number:
		*dst = *number
	default:
		return fmt.Errorf("cannot copy Number to variable of type %T", dest)
	}
	return nil
}

//func (number *Number) Value() (driver.Value, error) {
//	return number.Int64()
//}
