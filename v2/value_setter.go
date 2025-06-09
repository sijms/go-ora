package go_ora

import (
	"database/sql"
	"fmt"
	"github.com/sijms/go-ora/v2/converters"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// set null value from supported types
func setNull(value reflect.Value) error {
	if value.Kind() == reflect.Ptr && value.IsNil() {
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
	case tyBool:
		value.SetBool(!input.isZero())
	case tyString:
		temp, err := input.String()
		if err != nil {
			return err
		}
		value.SetString(temp)
	case tyNullString:
		temp, err := input.String()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullString{temp, true}))
	case tyNullByte:
		temp, err := input.Int64()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullByte{uint8(temp), true}))
	case tyNullInt16:
		temp, err := input.Int64()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullInt16{int16(temp), true}))
	case tyNullInt32:
		temp, err := input.Int64()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullInt32{int32(temp), true}))
	case tyNullInt64:
		temp, err := input.Int64()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullInt64{temp, true}))
	case tyNullFloat64:
		temp, err := input.Float64()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullFloat64{temp, true}))
	case tyNullBool:
		value.Set(reflect.ValueOf(sql.NullBool{!input.isZero(), true}))
	case tyNullNVarChar:
		temp, err := input.String()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(NullNVarChar{NVarChar(temp), true}))
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
		floatErr = fmt.Errorf(`can't assign string "%v" to float variablle`, input)
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
	case tyNumber:
		tempNum, err := NewNumberFromString(input)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(*tempNum))
	case tyBool:
		if strings.ToLower(input) == "true" {
			value.SetBool(true)
		} else {
			value.SetBool(false)
		}
	case tyString:
		value.SetString(input)
	case tyNullString:
		value.Set(reflect.ValueOf(sql.NullString{String: input, Valid: true}))
	case tyNullByte:
		if intErr == nil {
			value.Set(reflect.ValueOf(sql.NullByte{Byte: uint8(tempInt), Valid: true}))
		}
		return intErr
	case tyNullInt16:
		if intErr == nil {
			value.Set(reflect.ValueOf(sql.NullInt16{Int16: int16(tempInt), Valid: true}))
		}
		return intErr
	case tyNullInt32:
		if intErr == nil {
			value.Set(reflect.ValueOf(sql.NullInt32{Int32: int32(tempInt), Valid: true}))
		}
		return intErr
	case tyNullInt64:
		if intErr == nil {
			value.Set(reflect.ValueOf(sql.NullInt64{Int64: tempInt, Valid: true}))
		}
		return intErr
	case tyNullFloat64:
		if floatErr == nil {
			value.Set(reflect.ValueOf(sql.NullFloat64{Float64: tempFloat, Valid: true}))
		}
		return floatErr
	case tyNullBool:
		temp := strings.ToLower(input) == "true"
		value.Set(reflect.ValueOf(sql.NullBool{Bool: temp, Valid: true}))
	case tyNVarChar:
		value.Set(reflect.ValueOf(NVarChar(input)))
	case tyNullNVarChar:
		value.Set(reflect.ValueOf(NullNVarChar{NVarChar: NVarChar(input), Valid: true}))
	case tyTime:
		if timeErr == nil {
			value.Set(reflect.ValueOf(tempTime))
		}
		return timeErr
	case tyNullTime:
		if timeErr == nil {
			value.Set(reflect.ValueOf(sql.NullTime{Time: tempTime, Valid: true}))
		}
		return timeErr
	case tyTimeStamp:
		if timeErr == nil {
			value.Set(reflect.ValueOf(TimeStamp(tempTime)))
		}
		return timeErr
	case tyNullTimeStamp:
		if timeErr == nil {
			value.Set(reflect.ValueOf(NullTimeStamp{TimeStamp: TimeStamp(tempTime), Valid: true}))
		}
		return timeErr
	case tyTimeStampTZ:
		if timeErr == nil {
			value.Set(reflect.ValueOf(TimeStampTZ(tempTime)))
		}
		return timeErr
	case tyNullTimeStampTZ:
		if timeErr == nil {
			value.Set(reflect.ValueOf(NullTimeStampTZ{TimeStampTZ: TimeStampTZ(tempTime), Valid: true}))
		}
		return timeErr
	case tyClob:
		value.Set(reflect.ValueOf(Clob{String: input, Valid: true}))
	case tyNClob:
		value.Set(reflect.ValueOf(NClob{String: input, Valid: true}))
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
		return fmt.Errorf("can not assign string to type: %v", value.Type().Name())
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
	case tyString:
		value.SetString(string(input))
	case tyBytes:
		value.SetBytes(input)
	case tyNVarChar:
		value.Set(reflect.ValueOf(NVarChar(input)))
	case tyBlob:
		value.Set(reflect.ValueOf(Blob{Data: input}))
	case tyClob:
		value.Set(reflect.ValueOf(Clob{String: string(input), Valid: true}))
	case tyNClob:
		value.Set(reflect.ValueOf(NClob{String: string(input), Valid: true}))
	case tyNullString:
		value.Set(reflect.ValueOf(sql.NullString{string(input), true}))
	case tyNullNVarChar:
		value.Set(reflect.ValueOf(NullNVarChar{NVarChar(input), true}))
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
		return fmt.Errorf("can not assign []byte to type: %v", value.Type().Name())
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
	case tyString:
		value.SetString(input.Format(time.RFC3339))
	case tyTime:
		value.Set(reflect.ValueOf(input))
	case tyTimeStamp:
		value.Set(reflect.ValueOf(TimeStamp(input)))
	case tyTimeStampTZ:
		value.Set(reflect.ValueOf(TimeStampTZ(input)))
	case tyNullString:
		value.Set(reflect.ValueOf(sql.NullString{input.Format(time.RFC3339), true}))
	case tyNullTime:
		value.Set(reflect.ValueOf(sql.NullTime{input, true}))
	case tyNullTimeStamp:
		value.Set(reflect.ValueOf(NullTimeStamp{TimeStamp(input), true}))
	case tyNullTimeStampTZ:
		value.Set(reflect.ValueOf(NullTimeStampTZ{TimeStampTZ(input), true}))
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
func setFieldValue(fieldValue reflect.Value, cust *customType, input interface{}) error {
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
	case Lob:
		return setLob(fieldValue, val)
	case BFile:
		return setBFile(fieldValue, val)
	case Vector:
		return setVector(fieldValue, val)
	case []ParameterInfo:
		return setUDTObject(fieldValue, cust, val)
	default:
		if temp, ok := fieldValue.Interface().(sql.Scanner); ok {
			if temp != nil && !reflect.ValueOf(temp).IsNil() {
				return temp.Scan(input)
			}
		}
		if fieldValue.CanAddr() {
			if temp, ok := fieldValue.Addr().Interface().(sql.Scanner); ok {
				err := temp.Scan(input)
				return err
			}
		}
		return fmt.Errorf("unsupported primitive type: %s", fieldValue.Type().Name())
	}
}

func setLob(value reflect.Value, input Lob) error {
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
	if input.connection == nil || len(input.sourceLocator) == 0 {
		return setNull(value)
	}
	lobData, err := input.getData()
	if err != nil {
		return err
	}
	conn := input.connection
	if len(lobData) == 0 {
		return setNull(value)
	}
	getStrConv := func() (converters.IStringConverter, error) {
		var ret converters.IStringConverter
		if input.variableWidthChar() {
			if conn.dBVersion.Number < 10200 && input.littleEndianClob() {
				ret, _ = conn.getStrConv(2002)
			} else {
				ret, _ = conn.getStrConv(2000)
			}
		} else {
			ret, err = conn.getStrConv(input.charsetID)
			if err != nil {
				return nil, err
			}
		}
		return ret, nil
	}
	var strConv converters.IStringConverter
	switch value.Type() {
	case tyString:
		strConv, err = getStrConv()
		if err != nil {
			return err
		}
		value.SetString(strConv.Decode(lobData))
	case tyNullString:
		strConv, err = getStrConv()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(sql.NullString{strConv.Decode(lobData), true}))
	case tyNVarChar:
		strConv, err = getStrConv()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(NVarChar(strConv.Decode(lobData))))
	case tyNullNVarChar:
		strConv, err = getStrConv()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(NullNVarChar{NVarChar(strConv.Decode(lobData)), true}))
	case tyClob:
		strConv, err = getStrConv()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(Clob{
			String:  strConv.Decode(lobData),
			Valid:   true,
			locator: input.sourceLocator,
		}))
	case tyNClob:
		strConv, err = getStrConv()
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(NClob{
			String:  strConv.Decode(lobData),
			Valid:   true,
			locator: input.sourceLocator,
		}))
	case tyBlob:
		value.Set(reflect.ValueOf(Blob{
			Data:    lobData,
			locator: input.sourceLocator,
		}))
	case tyBytes:
		value.Set(reflect.ValueOf(lobData))
	default:
		if temp, ok := value.Interface().(sql.Scanner); ok {
			if temp != nil && !reflect.ValueOf(temp).IsNil() {
				return temp.Scan(lobData)
			}
		}
		if value.CanAddr() {
			if temp, ok := value.Addr().Interface().(sql.Scanner); ok {
				err := temp.Scan(lobData)
				return err
			}
		}
		return fmt.Errorf("can't assign LOB to type: %v", value.Type().Name())
	}
	return nil
}

func setVector(value reflect.Value, input Vector) error {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setVector(value.Elem(), input)
	}
	err := input.load()
	if err != nil {
		return err
	}
	switch value.Type() {
	case tyVector:
		value.Set(reflect.ValueOf(input))
	case tyUint8Array, tyFloat32Array, tyFloat64Array:
		value.Set(reflect.ValueOf(input.Data))
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
		return fmt.Errorf("can't assign Vector to type: %v", value.Type().Name())
	}
	return nil
}

func setBFile(value reflect.Value, input BFile) error {
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return setBFile(value.Elem(), input)
	}
	switch value.Type() {
	case tyBFile:
		value.Set(reflect.ValueOf(input))
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
		return fmt.Errorf("can't assign BFILE to type: %v", value.Type().Name())
	}
	return nil
}

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

func setUDTObject(value reflect.Value, cust *customType, input []ParameterInfo) error {
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
