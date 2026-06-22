package go_ora

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/converters"
	oraTypes "github.com/sijms/go-ora/v3/types"
)

var truthy = []string{"true", "1"}

// set null value from supported types
func setNull(value reflect.Value) error {
	if value.Kind() == reflect.Ptr && value.IsNil() {
		return nil
	}
	// if the value is support scanner interface call it
	if err := setWithScanner(value, nil); err == nil {
		return nil
	}
	value.Set(reflect.Zero(value.Type()))
	// value.SetZero()
	return nil
}

// set number value from supported types
func setNumber(value reflect.Value, input *Number) error {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setNumber(value.Elem(), input)
	}
	if tSigned(value.Type()) {
		temp, err := input.Int64()
		if err != nil {
			return err
		}
		value.SetInt(temp)
		return nil
	}
	if tUnsigned(value.Type()) {
		temp, err := input.Uint64()
		if err != nil {
			return err
		}
		value.SetUint(temp)
		return nil
	}
	if tFloat(value.Type()) {
		temp, err := input.Float64()
		if err != nil {
			return err
		}
		value.SetFloat(temp)
		return nil
	}
	switch value.Type() {
	case oraTypes.TyBool:
		value.SetBool(!input.isZero())
	case oraTypes.TyString:
		temp, err := input.String()
		if err != nil {
			return err
		}
		value.SetString(temp)
	case oraTypes.TyNullString:
		temp, err := input.String()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullString{temp, true}))
	case oraTypes.TyNullByte:
		temp, err := input.Int64()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullByte{uint8(temp), true}))
	case oraTypes.TyNullInt16:
		temp, err := input.Int64()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullInt16{int16(temp), true}))
	case oraTypes.TyNullInt32:
		temp, err := input.Int64()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullInt32{int32(temp), true}))
	case oraTypes.TyNullInt64:
		temp, err := input.Int64()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullInt64{temp, true}))
	case oraTypes.TyNullFloat64:
		temp, err := input.Float64()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullFloat64{temp, true}))
	case oraTypes.TyNullBool:
		value.Set(reflect.ValueOf(sql.NullBool{!input.isZero(), true}))
	//case oraTypes.TyNullNVarChar:
	//	temp, err := input.String()
	//	if err != nil {
	//		return err
	//	}
	//	value.Set(reflect.ValueOf(NullNVarChar{NVarChar(temp), true}))
	default:
		if temp, ok := value.Interface().(sql.Scanner); ok {
			if temp != nil && !reflect.ValueOf(temp).IsNil() {
				return temp.Scan(input)
			}
		}
		if value.CanAddr() {
			if temp, ok := value.Addr().Interface().(sql.Scanner); ok {
				err := temp.Scan(input)
				return err
			}
		}
		return fmt.Errorf("can not assign number to type: %v", value.Type().Name())
	}
	return nil
}

// set string value from supported types
func setString(value reflect.Value, input string) error {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setString(value.Elem(), input)
	}
	var intErr, floatErr, timeErr error
	tempInt, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		intErr = fmt.Errorf(`can't assign string "%v" to int variable`, input)
	}
	tempFloat, err := strconv.ParseFloat(input, 64)
	if err != nil {
		floatErr = fmt.Errorf(`can't assign string "%v" to float variable`, input)
	}
	tempTime, err := time.Parse(time.RFC3339, input)
	if err != nil {
		timeErr = fmt.Errorf(`can't assign string "%v" to time.Time variable`, input)
	}
	if tSigned(value.Type()) {
		if intErr == nil {
			value.SetInt(tempInt)
		}
		return intErr
	}
	if tUnsigned(value.Type()) {
		if intErr == nil {
			value.SetUint(uint64(tempInt))
		}
		return intErr
	}
	if tFloat(value.Type()) {
		if floatErr == nil {
			value.SetFloat(tempFloat)
		}
		return floatErr
	}
	switch value.Type() {
	case oraTypes.TyNumber:
		tempNum, err := NewNumberFromString(input)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(*tempNum))
	case oraTypes.TyBool:
		if slices.Contains(truthy, strings.ToLower(input)) {
			value.SetBool(true)
		} else {
			value.SetBool(false)
		}
	case oraTypes.TyString:
		value.SetString(input)
	case oraTypes.TyNullString:
		value.Set(reflect.ValueOf(sql.NullString{String: input, Valid: true}))
	case oraTypes.TyNullByte:
		if intErr == nil {
			value.Set(reflect.ValueOf(sql.NullByte{Byte: uint8(tempInt), Valid: true}))
		}
		return intErr
	case oraTypes.TyNullInt16:
		if intErr == nil {
			value.Set(reflect.ValueOf(sql.NullInt16{Int16: int16(tempInt), Valid: true}))
		}
		return intErr
	case oraTypes.TyNullInt32:
		if intErr == nil {
			value.Set(reflect.ValueOf(sql.NullInt32{Int32: int32(tempInt), Valid: true}))
		}
		return intErr
	case oraTypes.TyNullInt64:
		if intErr == nil {
			value.Set(reflect.ValueOf(sql.NullInt64{Int64: tempInt, Valid: true}))
		}
		return intErr
	case oraTypes.TyNullFloat64:
		if floatErr == nil {
			value.Set(reflect.ValueOf(sql.NullFloat64{Float64: tempFloat, Valid: true}))
		}
		return floatErr
	case oraTypes.TyNullBool:
		temp := slices.Contains(truthy, strings.ToLower(input))
		value.Set(reflect.ValueOf(sql.NullBool{Bool: temp, Valid: true}))
	//case tyNVarChar:
	//	value.Set(reflect.ValueOf(NVarChar(input)))
	//case tyNullNVarChar:
	//	value.Set(reflect.ValueOf(NullNVarChar{NVarChar: NVarChar(input), Valid: true}))
	case oraTypes.TyTime:
		if timeErr == nil {
			value.Set(reflect.ValueOf(tempTime))
		}
		return timeErr
	case oraTypes.TyNullTime:
		if timeErr == nil {
			value.Set(reflect.ValueOf(sql.NullTime{Time: tempTime, Valid: true}))
		}
		return timeErr
	//case tyTimeStamp:
	//	if timeErr == nil {
	//		value.Set(reflect.ValueOf(TimeStamp(tempTime)))
	//	}
	//	return timeErr
	//case tyNullTimeStamp:
	//	if timeErr == nil {
	//		value.Set(reflect.ValueOf(NullTimeStamp{TimeStamp: TimeStamp(tempTime), Valid: true}))
	//	}
	//	return timeErr
	//case tyTimeStampTZ:
	//	if timeErr == nil {
	//		value.Set(reflect.ValueOf(TimeStampTZ(tempTime)))
	//	}
	//	return timeErr
	//case tyNullTimeStampTZ:
	//	if timeErr == nil {
	//		value.Set(reflect.ValueOf(NullTimeStampTZ{TimeStampTZ: TimeStampTZ(tempTime), Valid: true}))
	//	}
	//	return timeErr
	//case tyClob:
	//	value.Set(reflect.ValueOf(Clob{String: input, Valid: true}))
	//case tyNClob:
	//	value.Set(reflect.ValueOf(NClob{String: input, Valid: true}))
	default:
		return setWithScanner(value, tempInt)
	}
	return nil
}

// set []byte value from supported types
func setBytes(value reflect.Value, input []byte) error {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setBytes(value.Elem(), input)
	}
	switch value.Type() {
	case oraTypes.TyString:
		value.SetString(string(input))
	case oraTypes.TyBytes:
		value.SetBytes(input)
	//case tyNVarChar:
	//	value.Set(reflect.ValueOf(NVarChar(input)))
	//case tyBlob:
	//	value.Set(reflect.ValueOf(Blob{Data: input}))
	//case tyClob:
	//	value.Set(reflect.ValueOf(Clob{String: string(input), Valid: true}))
	//case tyNClob:
	//	value.Set(reflect.ValueOf(NClob{String: string(input), Valid: true}))
	case oraTypes.TyNullString:
		value.Set(reflect.ValueOf(sql.NullString{string(input), true}))
	//case tyNullNVarChar:
	//	value.Set(reflect.ValueOf(NullNVarChar{NVarChar(input), true}))
	default:
		return setWithScanner(value, input)
		//if temp, ok := value.Interface().(sql.Scanner); ok {
		//	if temp != nil && !reflect.ValueOf(temp).IsNil() {
		//		return temp.Scan(input)
		//	}
		//}
		//if value.CanAddr() {
		//	if temp, ok := value.Addr().Interface().(sql.Scanner); ok {
		//		err := temp.Scan(input)
		//		return err
		//	}
		//}
		//return fmt.Errorf("can not assign []byte to type: %v", value.Type().Name())
	}
	return nil
}

// set time.Time value from supported types
func setTime(value reflect.Value, input time.Time) error {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setTime(value.Elem(), input)
	}
	switch value.Type() {
	case oraTypes.TyString:
		value.SetString(input.Format(time.RFC3339))
	case oraTypes.TyTime:
		value.Set(reflect.ValueOf(input))
	//case tyTimeStamp:
	//	value.Set(reflect.ValueOf(TimeStamp(input)))
	//case oraTypes.TyTimeStampTZ:
	//	value.Set(reflect.ValueOf(TimeStampTZ(input)))
	case oraTypes.TyNullString:
		value.Set(reflect.ValueOf(sql.NullString{input.Format(time.RFC3339), true}))
	case oraTypes.TyNullTime:
		value.Set(reflect.ValueOf(sql.NullTime{input, true}))
	//case tyNullTimeStamp:
	//	value.Set(reflect.ValueOf(NullTimeStamp{TimeStamp(input), true}))
	//case tyNullTimeStampTZ:
	//	value.Set(reflect.ValueOf(NullTimeStampTZ{TimeStampTZ(input), true}))
	default:
		if temp, ok := value.Interface().(sql.Scanner); ok {
			if temp != nil && !reflect.ValueOf(temp).IsNil() {
				return temp.Scan(input)
			}
		}
		if value.CanAddr() {
			if temp, ok := value.Addr().Interface().(sql.Scanner); ok {
				err := temp.Scan(input)
				return err
			}
		}
		return fmt.Errorf("can not assign time to type: %v", value.Type().Name())
	}
	return nil
}

// set field value
func setFieldValue(fieldValue reflect.Value, cust *Object, input interface{}) error {
	// input should be one of primitive values
	if input == nil {
		return setNull(fieldValue)
	}
	if fieldValue.Kind() == reflect.Ptr && fieldValue.Elem().Kind() == reflect.Interface {
		fieldValue.Elem().Set(reflect.ValueOf(input))
		return nil
	}
	if fieldValue.Kind() == reflect.Interface {
		fieldValue.Set(reflect.ValueOf(input))
		return nil
	}
	//if fieldValue.CanAddr() {
	//	if scan, ok := fieldValue.Addr().Interface().(sql.Scanner); ok {
	//		return scan.Scan(input)
	//	}
	//} else {
	//	if scan, ok := fieldValue.Interface().(sql.Scanner); ok {
	//		return scan.Scan(input)
	//	}
	//}

	switch val := input.(type) {
	case int64, float64:
		num, err := NewNumber(val)
		if err != nil {
			return err
		}
		return setNumber(fieldValue, num)
	// case float64:
	//	return setNumber(fieldValue, val)
	case string:
		return setString(fieldValue, val)
	case time.Time:
		return setTime(fieldValue, val)
	case []byte:
		return setBytes(fieldValue, val)
	case LobStream:
		return setLob(fieldValue, val)
	case oraTypes.Clob:
		return setClob(fieldValue, val)
	case oraTypes.Blob:
		return setBlob(fieldValue, val)
	//case BFile:
	//	return setBFile(fieldValue, val)
	//case Vector:
	//	return setVector(fieldValue, val)
	case []ParameterInfo:
		return setUDTObject(fieldValue, cust, val)
	default:
		return setWithScanner(fieldValue, input)
		//if temp, ok := fieldValue.Interface().(sql.Scanner); ok {
		//	if temp != nil && !reflect.ValueOf(temp).IsNil() {
		//		return temp.Scan(input)
		//	}
		//}
		//if fieldValue.CanAddr() {
		//	if temp, ok := fieldValue.Addr().Interface().(sql.Scanner); ok {
		//		err := temp.Scan(input)
		//		return err
		//	}
		//}
		//return fmt.Errorf("unsupported primitive type: %s", fieldValue.Type().Name())
	}
}
func setClob(dest reflect.Value, input oraTypes.Clob) error {
	if dest.Kind() == reflect.Ptr {
		if dest.IsNil() {
			dest.Set(reflect.New(dest.Type().Elem()))
		}
		return setClob(dest.Elem(), input)
	}
	// read data
	var err error
	if input.GetReadMode() == configurations.LobReadMode_AUTO {
		err = input.Read(context.Background())
		if err != nil {
			return err
		}
	}
	temp, err := input.Value()
	if err != nil {
		return err
	}
	if temp == nil {
		dest.Set(reflect.Zero(dest.Type()))
	} else {
		switch dest.Type() {
		case oraTypes.TyString:
			dest.SetString(temp.(string))
		case oraTypes.TyNullString:
			dest.Set(reflect.ValueOf(sql.NullString{String: temp.(string), Valid: true}))
		default:
			return setWithScanner(dest, input)
		}
	}

	return nil
}
func setWithScanner(dest reflect.Value, input interface{}) error {
	if temp, ok := dest.Interface().(sql.Scanner); ok {
		if temp != nil && !reflect.ValueOf(temp).IsNil() {
			return temp.Scan(input)
		}
	}
	if dest.CanAddr() {
		if temp, ok := dest.Addr().Interface().(sql.Scanner); ok {
			err := temp.Scan(input)
			return err
		}
	}
	return fmt.Errorf("can't set %T to type: %v", input, dest.Type().Name())
}
func setBlob(dest reflect.Value, input oraTypes.Blob) error {
	if dest.Kind() == reflect.Ptr {
		if dest.IsNil() {
			dest.Set(reflect.New(dest.Type().Elem()))
		}
		return setBlob(dest.Elem(), input)
	}
	// read data
	var err error
	if input.GetReadMode() == configurations.LobReadMode_AUTO {
		err = input.Read(context.Background())
		if err != nil {
			return err
		}
	}
	temp, err := input.Value()
	if err != nil {
		return err
	}
	if temp == nil {
		dest.Set(reflect.Zero(dest.Type()))
	} else {
		switch dest.Type() {
		case oraTypes.TyBytes:
			dest.SetBytes(temp.([]byte))
		default:
			return setWithScanner(dest, input)
		}
	}
	return nil
}
func setLob(value reflect.Value, input LobStream) error {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setLob(value.Elem(), input)
	}
	// dataSize, err := input.getSize()
	// if err != nil {
	// 	return err
	// }
	if input.conn == nil || len(input.sourceLocator) == 0 {
		return setNull(value)
	}
	lobData, err := input.getData()
	if err != nil {
		return err
	}
	conn := input.conn
	if len(lobData) == 0 {
		return setNull(value)
	}
	getStrConv := func() (converters.IStringConverter, error) {
		var ret converters.IStringConverter
		if input.GetLocator().IsVarWidthChar() {
			if conn.dBVersion.Number < 10200 && input.GetLocator().IsLittleEndian() {
				ret, _ = conn.GetStringCoder(2002, 0)
			} else {
				ret, _ = conn.GetStringCoder(2000, 0)
			}
		} else {
			ret, err = conn.GetStringCoder(input.charsetID, 0)
			if err != nil {
				return nil, err
			}
		}
		return ret, nil
	}
	var strConv converters.IStringConverter
	switch value.Type() {
	case oraTypes.TyString:
		strConv, err = getStrConv()
		if err != nil {
			return err
		}
		value.SetString(strConv.Decode(lobData))
	case oraTypes.TyNullString:
		strConv, err = getStrConv()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullString{strConv.Decode(lobData), true}))
	//case tyNVarChar:
	//	strConv, err = getStrConv()
	//	if err != nil {
	//		return err
	//	}
	//	value.Set(reflect.ValueOf(NVarChar(strConv.Decode(lobData))))
	//case tyNullNVarChar:
	//	strConv, err = getStrConv()
	//	if err != nil {
	//		return err
	//	}
	//	value.Set(reflect.ValueOf(NullNVarChar{NVarChar(strConv.Decode(lobData)), true}))
	//case tyClob:
	//	strConv, err = getStrConv()
	//	if err != nil {
	//		return err
	//	}
	//	value.Set(reflect.ValueOf(Clob{
	//		String:  strConv.Decode(lobData),
	//		Valid:   true,
	//		locator: input.sourceLocator,
	//	}))
	//case tyNClob:
	//	strConv, err = getStrConv()
	//	if err != nil {
	//		return err
	//	}
	//	value.Set(reflect.ValueOf(NClob{
	//		String:  strConv.Decode(lobData),
	//		Valid:   true,
	//		locator: input.sourceLocator,
	//	}))
	case tyBlob:
		value.Set(reflect.ValueOf(Blob{
			Data:    lobData,
			locator: input.sourceLocator,
		}))
	case oraTypes.TyBytes:
		value.Set(reflect.ValueOf(lobData))
	default:
		return setWithScanner(value, input)
	}
	return nil
}

//func setVector(value reflect.Value, input Vector) error {
//	if value.Kind() == reflect.Ptr {
//		if value.IsNil() {
//			value.Set(reflect.New(value.Type().Elem()))
//		}
//		return setVector(value.Elem(), input)
//	}
//	err := input.load()
//	if err != nil {
//		return err
//	}
//	switch value.Type() {
//	case tyVector:
//		value.Set(reflect.ValueOf(input))
//	case tyUint8Array, tyFloat32Array, tyFloat64Array:
//		value.Set(reflect.ValueOf(input.Data))
//	default:
//		if temp, ok := value.Interface().(sql.Scanner); ok {
//			if temp != nil && !reflect.ValueOf(temp).IsNil() {
//				return temp.Scan(input)
//			}
//		}
//		if value.CanAddr() {
//			if temp, ok := value.Addr().Interface().(sql.Scanner); ok {
//				err := temp.Scan(input)
//				return err
//			}
//		}
//		return fmt.Errorf("can't assign Vector to type: %v", value.Type().Name())
//	}
//	return nil
//}
//func setBFile(value reflect.Value, input BFile) error {
//	if value.Kind() == reflect.Ptr {
//		if value.IsNil() {
//			value.Set(reflect.New(value.Type().Elem()))
//		}
//		return setBFile(value.Elem(), input)
//	}
//	switch value.Type() {
//	case tyBFile:
//		value.Set(reflect.ValueOf(input))
//	default:
//		if temp, ok := value.Interface().(sql.Scanner); ok {
//			if temp != nil && !reflect.ValueOf(temp).IsNil() {
//				return temp.Scan(input)
//			}
//		}
//		if value.CanAddr() {
//			if temp, ok := value.Addr().Interface().(sql.Scanner); ok {
//				err := temp.Scan(input)
//				return err
//			}
//		}
//		return fmt.Errorf("can't assign BFILE to type: %v", value.Type().Name())
//	}
//	return nil
//}

func setArray(value reflect.Value, input []ParameterInfo) error {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setArray(value.Elem(), input)
	}
	tValue := value.Type()
	tempSlice := reflect.MakeSlice(tValue, 0, len(input))
	for _, par := range input {
		tempObj := reflect.New(tValue.Elem())
		err := setFieldValue(tempObj.Elem(), par.cusType, par.oPrimValue)
		if err != nil {
			return err
		}
		tempSlice = reflect.Append(tempSlice, tempObj.Elem())
	}
	value.Set(tempSlice)
	return nil
}

func setUDTObject(value reflect.Value, cust *Object, input []ParameterInfo) error {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setUDTObject(value.Elem(), cust, input)
	}
	if value.Kind() == reflect.Slice || value.Kind() == reflect.Array {
		return setArray(value, input)
		// if cust.isRegularArray() {
		//
		// } else {
		// 	arrayObj := reflect.MakeSlice(reflect.SliceOf(cust.typ), 0, len(input))
		// 	for _, par := range input {
		// 		if temp, ok := par.oPrimValue.([]ParameterInfo); ok {
		// 			tempObj2 := reflect.New(cust.typ)
		// 			err := setFieldValue(tempObj2.Elem(), par.cusType, temp)
		// 			if err != nil {
		// 				return err
		// 			}
		// 			arrayObj = reflect.Append(arrayObj, tempObj2.Elem())
		// 		}
		// 	}
		// 	value.Set(arrayObj)
		// }
	} else {
		tempObj := reflect.New(cust.typ)
		for _, par := range input {
			if fieldIndex, ok := cust.fieldMap[par.Name]; ok {
				err := setFieldValue(tempObj.Elem().Field(fieldIndex), par.cusType, par.oPrimValue)
				if err != nil {
					return err
				}
			}
		}
		value.Set(tempObj.Elem())
	}
	return nil
}
