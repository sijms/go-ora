package utils

import (
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type (
	ParameterDirection int
)

const (
	Input  ParameterDirection = 1
	Output ParameterDirection = 2
	InOut  ParameterDirection = 3
	// RetVal ParameterDirection = 9
)

func CreateQuasiLocator(dataLength uint64) []byte {
	ret := make([]byte, 40)
	ret[1] = 38
	ret[3] = 4
	ret[4] = 97
	ret[5] = 8
	ret[9] = 1
	binary.BigEndian.PutUint64(ret[10:], dataLength)
	return ret
}

var (
	TyFloat64 = reflect.TypeOf((*float64)(nil)).Elem()
	TyFloat32 = reflect.TypeOf((*float32)(nil)).Elem()
	TyInt64   = reflect.TypeOf((*int64)(nil)).Elem()
	TyBool    = reflect.TypeOf((*bool)(nil)).Elem()
	TyBytes   = reflect.TypeOf((*[]byte)(nil)).Elem()
	TyString  = reflect.TypeOf((*string)(nil)).Elem()
	//TyNVarChar        = reflect.TypeOf((*NVarChar)(nil)).Elem()
	TyTime = reflect.TypeOf((*time.Time)(nil)).Elem()
	//TyTimeStamp       = reflect.TypeOf((*TimeStamp)(nil)).Elem()
	//TyTimeStampTZ     = reflect.TypeOf((*TimeStampTZ)(nil)).Elem()
	//TyClob            = reflect.TypeOf((*Clob)(nil)).Elem()
	//TyNClob           = reflect.TypeOf((*NClob)(nil)).Elem()
	//TyBlob            = reflect.TypeOf((*Blob)(nil)).Elem()
	//TyBFile           = reflect.TypeOf((*BFile)(nil)).Elem()
	//TyVector          = reflect.TypeOf((*Vector)(nil)).Elem()
	TyNullByte    = reflect.TypeOf((*sql.NullByte)(nil)).Elem()
	TyNullInt16   = reflect.TypeOf((*sql.NullInt16)(nil)).Elem()
	TyNullInt32   = reflect.TypeOf((*sql.NullInt32)(nil)).Elem()
	TyNullInt64   = reflect.TypeOf((*sql.NullInt64)(nil)).Elem()
	TyNullFloat64 = reflect.TypeOf((*sql.NullFloat64)(nil)).Elem()
	TyNullBool    = reflect.TypeOf((*sql.NullBool)(nil)).Elem()
	TyNullString  = reflect.TypeOf((*sql.NullString)(nil)).Elem()
	//TyNullNVarChar    = reflect.TypeOf((*NullNVarChar)(nil)).Elem()
	TyNullTime = reflect.TypeOf((*sql.NullTime)(nil)).Elem()
	//TyNullTimeStamp   = reflect.TypeOf((*NullTimeStamp)(nil)).Elem()
	//TyNullTimeStampTZ = reflect.TypeOf((*NullTimeStampTZ)(nil)).Elem()
	//TyRefCursor       = reflect.TypeOf((*RefCursor)(nil)).Elem()
	//TyPLBool          = reflect.TypeOf((*PLBool)(nil)).Elem()
	//TyObject          = reflect.TypeOf((*Object)(nil)).Elem()
	//TyNumber          = reflect.TypeOf((*Number)(nil)).Elem()
	TyFloat32Array = reflect.TypeOf((*[]float32)(nil)).Elem()
	TyUint8Array   = reflect.TypeOf((*[]uint8)(nil)).Elem()
	TyFloat64Array = reflect.TypeOf((*[]float64)(nil)).Elem()
)

func TSigned(input reflect.Type) bool {
	switch input.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func TUnsigned(input reflect.Type) bool {
	switch input.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func TInteger(input reflect.Type) bool {
	return TSigned(input) || TUnsigned(input)
}

func TFloat(input reflect.Type) bool {
	return input.Kind() == reflect.Float32 || input.Kind() == reflect.Float64
}

func TNumber(input reflect.Type) bool {
	return TInteger(input) || TFloat(input)
}

func GetValue(origVal driver.Value) (driver.Value, error) {
	if origVal == nil {
		return nil, nil
	}
	rOriginal := reflect.ValueOf(origVal)
	if rOriginal.Kind() == reflect.Ptr && rOriginal.IsNil() {
		return nil, nil
	}
	proVal := reflect.Indirect(rOriginal)
	if valuer, ok := proVal.Interface().(driver.Valuer); ok {
		return valuer.Value()
	}
	return proVal.Interface(), nil
}

func ExtractTag(tag string) (name, _type string, size int, direction ParameterDirection) {
	extractNameValue := func(input string, pos int) {
		parts := strings.Split(input, "=")
		var id, value string
		if len(parts) < 2 {
			switch pos {
			case 0:
				id = "name"
			case 1:
				id = "type"
			case 2:
				id = "size"
			case 3:
				id = "direction"
			}
			value = input
		} else {
			id = strings.TrimSpace(strings.ToLower(parts[0]))
			value = strings.TrimSpace(parts[1])
		}
		switch id {
		case "name":
			name = value
		case "type":
			_type = value
		case "size":
			tempSize, _ := strconv.ParseInt(value, 10, 32)
			size = int(tempSize)
		case "dir":
			fallthrough
		case "direction":
			switch value {
			case "in", "input":
				direction = Input
			case "out", "output":
				direction = Output
			case "inout":
				direction = InOut
			}
		}
	}
	tag = strings.TrimSpace(tag)
	if len(tag) == 0 {
		return
	}
	tagFields := strings.Split(tag, ",")
	if len(tagFields) > 0 {
		extractNameValue(tagFields[0], 0)
	}
	if len(tagFields) > 1 {
		extractNameValue(tagFields[1], 1)
	}
	if len(tagFields) > 2 {
		extractNameValue(tagFields[2], 2)
	}
	if len(tagFields) > 3 {
		extractNameValue(tagFields[3], 3)
	}
	return
}
