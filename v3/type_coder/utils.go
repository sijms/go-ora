package type_coder

import (
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"reflect"
	"time"

	"github.com/sijms/go-ora/v3/types"
	"github.com/sijms/go-ora/v3/utils"
)

func xorBuffer(buffer []byte, length int) {
	if len(buffer) < length {
		length = len(buffer)
	}
	for i := 0; i < length; i++ {
		buffer[i] = ^buffer[i]
	}
}

func putDate(buffer []byte, ti *time.Time) {
	buffer[0] = uint8(ti.Year()/100 + 100)
	buffer[1] = uint8(ti.Year()%100 + 100)
	buffer[2] = uint8(ti.Month())
	buffer[3] = uint8(ti.Day())
	buffer[4] = uint8(ti.Hour() + 1)
	buffer[5] = uint8(ti.Minute() + 1)
	buffer[6] = uint8(ti.Second() + 1)
}

func putTimestamp(buffer []byte, ti *time.Time) {
	putDate(buffer, ti)
	binary.BigEndian.PutUint32(buffer[7:11], uint32(ti.Nanosecond()))
}

func convertRowIDToByte(number int64, size int) []byte {
	buffer := []byte{
		65, 66, 67, 68, 69, 70, 71, 72,
		73, 74, 75, 76, 77, 78, 79, 80,
		81, 82, 83, 84, 85, 86, 87, 88,
		89, 90, 97, 98, 99, 100, 101, 102,
		103, 104, 105, 106, 107, 108, 109, 110,
		111, 112, 113, 114, 115, 116, 117, 118,
		119, 120, 121, 122, 48, 49, 50, 51,
		52, 53, 54, 55, 56, 57, 43, 47,
	}
	output := make([]byte, size)
	for x := size; x > 0; x-- {
		output[x-1] = buffer[number&0x3F]
		if number >= 0 {
			number = number >> 6
		} else {
			number = (number >> 6) + (2 << (32 + ^6))
		}
	}
	return output
}

func createQuasiLocator(dataLen uint64) []byte {
	ret := make([]byte, 40)
	ret[1] = 38
	ret[3] = 4
	ret[4] = 97
	ret[5] = 8
	ret[9] = 1
	binary.BigEndian.PutUint64(ret[10:], dataLen)
	return ret
}

func Encode(input any) (OracleTypeEncoder, error) {
	// if nil use string converter
	if input == nil {
		return NewStringCoder(sql.NullString{}, false)
	}
	// check if the type is actually implements OracleTypeEncoder so return it
	if temp, ok := input.(OracleTypeEncoder); ok {
		return temp, nil
	}
	vInput := reflect.ValueOf(input)
	for vInput.Kind() == reflect.Ptr {
		vInput = vInput.Elem()
	}
	// switch special types
	switch v := input.(type) {
	case types.Blob:
		return NewBlobEncoder(v), nil
	case *types.Blob:
		return NewBlobEncoder(*v), nil
	case types.Clob:
		return NewClobEncoder(v), nil
	case *types.Clob:
		return NewClobEncoder(*v), nil
	case types.Vector:
		return NewVectorEncoder(v)
	case *types.Vector:
		return NewVectorEncoder(*v)
		// BFile
		// Json
	}
	// switch main types
	if utils.IsNumber(vInput.Type()) || utils.IsNumber(vInput.Type()) {
		number, err := types.NewNumber(input)
		if err != nil {
			return nil, err
		}
		return NewNumberCoder(number)
	}
	switch vInput.Type() {
	case utils.TyString, utils.TyNullString:
		return NewStringCoder(input, false)
	case utils.TyBool:
		return NewBoolCoder(input)
	case utils.TyTime, utils.TyNullTime:
		return NewDate(input)
	case utils.TyBytes:
		return NewRawCoder(input)
	}
	// switch complex types types
	// objects
	// arrays

	// last thing see if the input support valuer interface
	if temp, ok := input.(driver.Valuer); ok {
		v, err := temp.Value()
		if err != nil {
			return nil, err
		}
		return Encode(v)
	}
	return nil, fmt.Errorf("unsupported type: %v", vInput.Type())
}
