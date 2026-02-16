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
)

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
	if dstType.Kind() != reflect.Ptr {
		return errors.New("dst must be a pointer")
	}
	if dstValue.Elem().Kind() == reflect.Interface && dstValue.Elem().IsNil() {
		err = createNewType(dstValue.Elem(), dstType.Elem())
		if err != nil {
			return err
		}
	}
	if dstType.Elem().Implements(reflect.TypeOf((*sql.Scanner)(nil)).Elem()) {
		return dstValue.Elem().Interface().(sql.Scanner).Scan(src)
	}
	//if scanner, ok := dstValue.Elem().Implements.(sql.Scanner); ok {
	//	return scanner.Scan(src)
	//}
	if temp, ok := src.(OracleType); ok {
		return temp.CopyTo(dst)
	}
	if src == nil {
		dstValue.Elem().Set(reflect.Zero(dstType.Elem()))
	}
	if dstValue.Elem().Kind() == reflect.Interface {
		dstValue.Elem().Set(reflect.ValueOf(src))
	}
	// all sql.Null* support scanner interface so should be resolved
	switch dstType.Elem() {
	case tyString:
		return setString(dstValue, src)
	}
	return nil
}
func setString(dst reflect.Value, src any) error {
	switch src := src.(type) {
	case string:
		dst.Elem().SetString(src)
	case int, int8, int16, int32, int64:
		dst.SetString(fmt.Sprintf("%d", src))
	case uint, uint8, uint16, uint32, uint64:
		dst.SetString(fmt.Sprintf("%d", src))
	default:
		return fmt.Errorf("cannot copy value of type %T to type %T", src, dst.Type())
	}
	return nil
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

func addSomeData(dst any) {
	if temp, ok := dst.(*blob); ok {
		temp.data = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	}
}
