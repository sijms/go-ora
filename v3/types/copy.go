package types

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"
)

var (
	tyFloat64     = reflect.TypeOf((*float64)(nil)).Elem()
	tyFloat32     = reflect.TypeOf((*float32)(nil)).Elem()
	tyInt64       = reflect.TypeOf((*int64)(nil)).Elem()
	tyBool        = reflect.TypeOf((*bool)(nil)).Elem()
	tyBytes       = reflect.TypeOf((*[]byte)(nil)).Elem()
	tyString      = reflect.TypeOf((*string)(nil)).Elem()
	tyTime        = reflect.TypeOf((*time.Time)(nil)).Elem()
	tyNullByte    = reflect.TypeOf((*sql.NullByte)(nil)).Elem()
	tyNullInt16   = reflect.TypeOf((*sql.NullInt16)(nil)).Elem()
	tyNullInt32   = reflect.TypeOf((*sql.NullInt32)(nil)).Elem()
	tyNullInt64   = reflect.TypeOf((*sql.NullInt64)(nil)).Elem()
	tyNullFloat64 = reflect.TypeOf((*sql.NullFloat64)(nil)).Elem()
	tyNullBool    = reflect.TypeOf((*sql.NullBool)(nil)).Elem()
	tyNullString  = reflect.TypeOf((*sql.NullString)(nil)).Elem()
	tyNullTime    = reflect.TypeOf((*sql.NullTime)(nil)).Elem()
	tyScanner     = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
)

func tSigned(input reflect.Type) bool {
	switch input.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}
func tUnsigned(input reflect.Type) bool {
	switch input.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}
func tInteger(input reflect.Type) bool {
	return tSigned(input) || tUnsigned(input)
}
func tFloat(input reflect.Type) bool {
	return input.Kind() == reflect.Float32 || input.Kind() == reflect.Float64
}
func tNumber(input reflect.Type) bool {
	return tInteger(input) || tFloat(input) || input == tyBool
}

func defaultCopy(dst, src any) error {
	dstValue := reflect.ValueOf(dst)
	dstType := reflect.TypeOf(dst)
	// if the destination implements the scanner interface, use it
	if dstType.Elem().Implements(tyScanner) {
		return dstValue.Elem().Interface().(sql.Scanner).Scan(src)
	}
	if src == nil {
		setNil(dst)
		//dstValue.Elem().Set(reflect.Zero(dstType.Elem()))
		return nil
	}
	// if the source implements the oracle type interface, use it
	if temp, ok := src.(OracleType); ok {
		return temp.CopyTo(dst)
	}
	if dstValue.Elem().Kind() == reflect.Interface {
		dstValue.Elem().Set(reflect.ValueOf(src))
		return nil
	}
	return fmt.Errorf("can't assign value of type %T into type %T", src, dst)
}

func Copy(dst, src any) error {
	var dstValue reflect.Value
	var dstType reflect.Type
	//if dst == nil {
	//	dstValue = reflect.ValueOf(&dst).Elem()
	//	dstType = reflect.TypeOf(&dst).Elem()
	//} else {
	//	dstValue = reflect.ValueOf(dst)
	//	dstType = reflect.TypeOf(dst)
	//}
	dstValue = reflect.ValueOf(dst)
	dstType = reflect.TypeOf(dst)
	var err error
	if dstType.Kind() != reflect.Ptr && dstType.Kind() != reflect.Pointer {
		return errors.New("dst must be a pointer")
	}
	// first check if the destination is empty
	if dstValue.Elem().Kind() == reflect.Interface && dstValue.Elem().IsNil() {
		err = createNewType(dstValue.Elem(), dstType.Elem())
		if err != nil {
			return err
		}
	}
	//// if the destination implements the scanner interface, use it
	//if dstType.Elem().Implements(tyScanner) {
	//	return dstValue.Elem().Interface().(sql.Scanner).Scan(src)
	//}
	//// if the source implements the oracle type interface, use it
	//if temp, ok := src.(OracleType); ok {
	//	return temp.CopyTo(dst)
	//}
	//if src == nil {
	//	setNil(dst)
	//	//dstValue.Elem().Set(reflect.Zero(dstType.Elem()))
	//	return nil
	//}
	// if destination is interface and is not nil, set it to src
	//if dstValue.Elem().Kind() == reflect.Interface {
	//	dstValue.Elem().Set(reflect.ValueOf(src))
	//	return nil
	//}

	switch src := src.(type) {
	case string:
		// update destination with src
		err = copyString(dst, src)
	//if dstType.Elem() == tyString || dstType.Elem().Kind() == reflect.String {
	//	dstValue.Elem().SetString(src)
	//} else if reflect.TypeOf(src).ConvertibleTo(dstType.Elem()) {
	//	dstValue.Elem().Set(reflect.ValueOf(src).Convert(dstType.Elem()))
	//}
	case sql.NullString:
		if src.Valid {
			err = copyString(dst, src.String)
		} else {
			setNil(dst)
		}
	case []byte:
		err = copyBytes(dst, src)
	//case int, int8, int16, int32, int64:
	//	if dstType.Elem() == tyString || dstType.Elem().Kind() == reflect.String {
	//		dstValue.Elem().SetString(fmt.Sprintf("%d", src))
	//	} else if reflect.TypeOf(src).ConvertibleTo(dstType.Elem()) {
	//		dstValue.Elem().Set(reflect.ValueOf(src).Convert(dstType.Elem()))
	//	}
	//case uint, uint8, uint16, uint32, uint64:
	//	if dstType.Elem() == tyString || dstType.Elem().Kind() == reflect.String {
	//		dstValue.Elem().SetString(fmt.Sprintf("%d", src))
	//	} else if reflect.TypeOf(src).ConvertibleTo(dstType.Elem()) {
	//		dstValue.Elem().Set(reflect.ValueOf(src).Convert(dstType.Elem()))
	//	}
	//case float32, float64:
	//	if dstType.Elem() == tyString || dstType.Elem().Kind() == reflect.String {
	//		dstValue.Elem().SetString(fmt.Sprintf("%f", src))
	//	} else if reflect.TypeOf(src).ConvertibleTo(dstType.Elem()) {
	//		dstValue.Elem().Set(reflect.ValueOf(src).Convert(dstType.Elem()))
	//	}
	case bool:
		if dstType.Elem() == tyString || dstType.Elem().Kind() == reflect.String {
			dstValue.Elem().SetString(fmt.Sprintf("%t", src))
		} else if reflect.TypeOf(src).ConvertibleTo(dstType.Elem()) {
			dstValue.Elem().Set(reflect.ValueOf(src).Convert(dstType.Elem()))
		}
	case time.Time:
		err = copyTime(dst, src)
		//if dstType.Elem() == tyString || dstType.Elem().Kind() == reflect.String {
		//	dstValue.Elem().SetString(src.Format(time.RFC3339))
		//} else if reflect.TypeOf(src).ConvertibleTo(dstType.Elem()) {
		//	dstValue.Elem().Set(reflect.ValueOf(src).Convert(dstType.Elem()))
		//}
	case sql.NullTime:
		if src.Valid {
			err = copyTime(dst, src.Time)
		} else {
			setNil(dst)
		}
	default:
		err = defaultCopy(dst, src)
		//if reflect.TypeOf(src).ConvertibleTo(dstType.Elem()) {
		//	dstValue.Elem().Set(reflect.ValueOf(src).Convert(dstType.Elem()))
		//} else {
		//	return fmt.Errorf("cannot copy value of type %T to type %T", src, dstType.Elem())
		//}
	}
	return err
}

func createNewType(dstValue reflect.Value, dstType reflect.Type) error {
	// special types
	if dstType == reflect.TypeOf((*Blob)(nil)).Elem() {
		dstValue.Set(reflect.ValueOf(&blob{}))
		return nil
	}
	if dstType == reflect.TypeOf((*Clob)(nil)).Elem() {
		dstValue.Set(reflect.ValueOf(&clob{}))
		return nil
	}
	if dstType == reflect.TypeOf((*Vector)(nil)).Elem() {
		dstValue.Set(reflect.ValueOf(&vector{}))
		return nil
	}
	if dstType.Kind() != reflect.Ptr {
		return errors.New("dst must be a pointer")
	}
	// bfile also here

	// return default null type
	dstValue.Set(reflect.New(dstType.Elem()))
	return nil
}

//func addSomeData(dst any) {
//	if temp, ok := dst.(*blob); ok {
//		temp.data = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
//	}
//}
