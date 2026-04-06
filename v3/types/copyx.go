package types

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func setNil(dest any) {
	dstValue := reflect.ValueOf(dest).Elem()
	dstType := reflect.TypeOf(dest).Elem()
	//if dstValue.Kind() == reflect.Ptr && dstValue.IsNil() {
	//	return
	//}
	dstValue.Set(reflect.Zero(dstType))
}
func copyTime(dest any, src time.Time) error {
	dstValue := reflect.ValueOf(dest).Elem()
	dstType := reflect.TypeOf(dest).Elem()
	switch dstType {
	case tyString:
		dstValue.SetString(src.Format(time.RFC3339))
	case tyTime:
		dstValue.Set(reflect.ValueOf(src))
	case tyNullString:
		dstValue.Set(reflect.ValueOf(sql.NullString{String: src.Format(time.RFC3339), Valid: true}))
	case tyNullTime:
		dstValue.Set(reflect.ValueOf(sql.NullTime{Time: src, Valid: true}))
	default:
		return defaultCopy(dest, src)
	}
	return nil
}
func copyBytes(dest any, src []uint8) error {
	dstValue := reflect.ValueOf(dest).Elem()
	dstType := reflect.TypeOf(dest).Elem()
	switch dstType {
	case tyString:
		dstValue.SetString(string(src))
	case tyBytes:
		dstValue.SetBytes(src)
	case tyNullString:
		dstValue.Set(reflect.ValueOf(sql.NullString{String: string(src), Valid: src != nil}))
	default:
		return defaultCopy(dest, src)
	}
	return nil
}
func copyString(dest any, src string) error {
	var intErr, floatErr, timeErr error
	tempInt, err := strconv.ParseInt(src, 10, 64)
	if err != nil {
		intErr = fmt.Errorf(`can't assign string "%v" to int variable`, src)
	}
	tempFloat, err := strconv.ParseFloat(src, 64)
	if err != nil {
		floatErr = fmt.Errorf(`can't assign string "%v" to float variable`, src)
	}
	tempTime, err := time.Parse(time.RFC3339, src)
	if err != nil {
		timeErr = fmt.Errorf(`can't assign string "%v" to time.Time variable`, src)
	}
	dstValue := reflect.ValueOf(dest).Elem()
	dstType := reflect.TypeOf(dest).Elem()

	if tSigned(dstType) {
		if intErr == nil {
			dstValue.SetInt(tempInt)
		}
		return intErr
	}
	if tUnsigned(dstType) {
		if intErr == nil {
			dstValue.SetUint(uint64(tempInt))
		}
		return intErr
	}
	if tFloat(dstType) {
		if floatErr == nil {
			dstValue.SetFloat(tempFloat)
		}
		return floatErr
	}
	switch dstType {
	case tyBool:
		dstValue.SetBool(strings.ToLower(src) == "true")
	case tyString:
		dstValue.SetString(src)
	case tyNullString:
		if src == "" {
			dstValue.Set(reflect.ValueOf(sql.NullString{}))
		} else {
			dstValue.Set(reflect.ValueOf(sql.NullString{String: src, Valid: true}))
		}
	case tyNullByte:
		if src == "" {
			dstValue.Set(reflect.ValueOf(sql.NullByte{}))
		} else {
			if intErr == nil {
				dstValue.Set(reflect.ValueOf(sql.NullByte{Byte: uint8(tempInt), Valid: true}))
			}
			return intErr
		}
	case tyNullInt16:
		if src == "" {
			dstValue.Set(reflect.ValueOf(sql.NullInt16{}))
		} else {
			if intErr == nil {
				dstValue.Set(reflect.ValueOf(sql.NullInt16{Int16: int16(tempInt), Valid: true}))
			}
			return intErr
		}
	case tyNullInt32:
		if src == "" {
			dstValue.Set(reflect.ValueOf(sql.NullInt32{}))
		} else {
			if intErr == nil {
				dstValue.Set(reflect.ValueOf(sql.NullInt32{Int32: int32(tempInt), Valid: true}))
			}
			return intErr
		}
	case tyNullInt64:
		if src == "" {
			dstValue.Set(reflect.ValueOf(sql.NullInt64{}))
		} else {
			if intErr == nil {
				dstValue.Set(reflect.ValueOf(sql.NullInt64{Int64: tempInt, Valid: true}))
			}
			return intErr
		}

	case tyNullFloat64:
		if src == "" {
			dstValue.Set(reflect.ValueOf(sql.NullFloat64{}))
		} else {
			if floatErr == nil {
				dstValue.Set(reflect.ValueOf(sql.NullFloat64{Float64: tempFloat, Valid: true}))
			}
			return floatErr
		}

	case tyNullBool:
		if src == "" {
			dstValue.Set(reflect.ValueOf(sql.NullBool{}))
		} else {
			temp := strings.ToLower(src) == "true"
			dstValue.Set(reflect.ValueOf(sql.NullBool{Bool: temp, Valid: true}))
		}

	case tyTime:
		if timeErr == nil {
			dstValue.Set(reflect.ValueOf(tempTime))
		}
		return timeErr
	case tyNullTime:
		if src == "" {
			dstValue.Set(reflect.ValueOf(sql.NullTime{}))
		} else {
			if timeErr == nil {
				dstValue.Set(reflect.ValueOf(sql.NullTime{Time: tempTime, Valid: true}))
			}
			return timeErr
		}

	default:
		return defaultCopy(dest, src)
	}
	return nil
}
