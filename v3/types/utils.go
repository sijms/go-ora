package types

import (
	"database/sql"
	"encoding/binary"
	"reflect"
	"time"
)

var (
	TyFloat64     = reflect.TypeOf((*float64)(nil)).Elem()
	TyFloat32     = reflect.TypeOf((*float32)(nil)).Elem()
	TyInt         = reflect.TypeOf((*int)(nil)).Elem()
	TyInt8        = reflect.TypeOf((*int8)(nil)).Elem()
	TyInt16       = reflect.TypeOf((*int16)(nil)).Elem()
	TyInt32       = reflect.TypeOf((*int32)(nil)).Elem()
	TyInt64       = reflect.TypeOf((*int64)(nil)).Elem()
	TyUint        = reflect.TypeOf((*uint)(nil)).Elem()
	TyUint8       = reflect.TypeOf((*uint8)(nil)).Elem()
	TyUint16      = reflect.TypeOf((*uint16)(nil)).Elem()
	TyUint32      = reflect.TypeOf((*uint32)(nil)).Elem()
	TyUint64      = reflect.TypeOf((*uint64)(nil)).Elem()
	TyBool        = reflect.TypeOf((*bool)(nil)).Elem()
	TyBytes       = reflect.TypeOf((*[]byte)(nil)).Elem()
	TyString      = reflect.TypeOf((*string)(nil)).Elem()
	TyTime        = reflect.TypeOf((*time.Time)(nil)).Elem()
	TyNullByte    = reflect.TypeOf((*sql.NullByte)(nil)).Elem()
	TyNullInt16   = reflect.TypeOf((*sql.NullInt16)(nil)).Elem()
	TyNullInt32   = reflect.TypeOf((*sql.NullInt32)(nil)).Elem()
	TyNullInt64   = reflect.TypeOf((*sql.NullInt64)(nil)).Elem()
	TyNullFloat64 = reflect.TypeOf((*sql.NullFloat64)(nil)).Elem()
	TyNullBool    = reflect.TypeOf((*sql.NullBool)(nil)).Elem()
	TyNullString  = reflect.TypeOf((*sql.NullString)(nil)).Elem()
	TyNullTime    = reflect.TypeOf((*sql.NullTime)(nil)).Elem()
	TyScanner     = reflect.TypeOf((*sql.Scanner)(nil)).Elem()

	TyNumber   = reflect.TypeOf((*Number)(nil)).Elem()
	TyBoolean  = reflect.TypeOf((*Bool)(nil)).Elem()
	TyVarchar  = reflect.TypeOf((*string)(nil)).Elem()
	TyDate     = reflect.TypeOf((*Date)(nil)).Elem()
	TyInterval = reflect.TypeOf((*Interval)(nil)).Elem()
	TyRaw      = reflect.TypeOf((*Raw)(nil)).Elem()
	TyVector   = reflect.TypeOf((*Vector)(nil)).Elem()
	TyJson     = reflect.TypeOf((*Json)(nil)).Elem()
	TyClob     = reflect.TypeOf((*Clob)(nil)).Elem()
	TyBlob     = reflect.TypeOf((*Blob)(nil)).Elem()
	TyBFile    = reflect.TypeOf((*BFile)(nil)).Elem()
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
