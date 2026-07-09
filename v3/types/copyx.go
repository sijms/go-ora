package types

import (
	"database/sql"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

var truthy = []string{"true", "1"}

func setNil(dest reflect.Value) error {
	if dest.Kind() == reflect.Ptr && dest.IsNil() {
		return nil
	}
	// if the value is support scanner interface call it
	processed, err := copyWithScanner(dest, nil)
	if processed {
		return err
	}
	dest.Set(reflect.Zero(dest.Type()))
	return nil
}

func copyTime(dest reflect.Value, src time.Time) error {
	switch dest.Type() {
	case TyString:
		dest.SetString(src.Format(time.RFC3339))
	case TyTime:
		dest.Set(reflect.ValueOf(src))
	case TyNullString:
		dest.Set(reflect.ValueOf(sql.NullString{String: src.Format(time.RFC3339), Valid: true}))
	case TyNullTime:
		dest.Set(reflect.ValueOf(sql.NullTime{Time: src, Valid: true}))
	default:
		return defaultCopy(dest, src)
	}
	return nil
}

func copyBytes(dest reflect.Value, src []byte) error {
	switch dest.Type() {
	case TyString:
		dest.SetString(string(src))
	case TyBytes:
		dest.SetBytes(src)
	case TyNullString:
		dest.Set(reflect.ValueOf(sql.NullString{String: string(src), Valid: src != nil}))
	default:
		return defaultCopy(dest, src)
	}
	return nil
}
func copyString(dest reflect.Value, src string) error {
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

	if tSigned(dest.Type()) {
		if intErr == nil {
			dest.SetInt(tempInt)
		}
		return intErr
	}
	if tUnsigned(dest.Type()) {
		if intErr == nil {
			dest.SetUint(uint64(tempInt))
		}
		return intErr
	}
	if tFloat(dest.Type()) {
		if floatErr == nil {
			dest.SetFloat(tempFloat)
		}
		return floatErr
	}
	switch dest.Type() {
	case TyBool:
		if slices.Contains(truthy, strings.ToLower(src)) {
			dest.SetBool(true)
		} else {
			dest.SetBool(false)
		}
	case TyString:
		dest.SetString(src)
	case TyNullString:
		if src == "" {
			dest.Set(reflect.ValueOf(sql.NullString{}))
		} else {
			dest.Set(reflect.ValueOf(sql.NullString{String: src, Valid: true}))
		}
	case TyNullByte:
		if src == "" {
			dest.Set(reflect.ValueOf(sql.NullByte{}))
		} else {
			if intErr == nil {
				dest.Set(reflect.ValueOf(sql.NullByte{Byte: uint8(tempInt), Valid: true}))
			}
			return intErr
		}
	case TyNullInt16:
		if src == "" {
			dest.Set(reflect.ValueOf(sql.NullInt16{}))
		} else {
			if intErr == nil {
				dest.Set(reflect.ValueOf(sql.NullInt16{Int16: int16(tempInt), Valid: true}))
			}
			return intErr
		}
	case TyNullInt32:
		if src == "" {
			dest.Set(reflect.ValueOf(sql.NullInt32{}))
		} else {
			if intErr == nil {
				dest.Set(reflect.ValueOf(sql.NullInt32{Int32: int32(tempInt), Valid: true}))
			}
			return intErr
		}
	case TyNullInt64:
		if src == "" {
			dest.Set(reflect.ValueOf(sql.NullInt64{}))
		} else {
			if intErr == nil {
				dest.Set(reflect.ValueOf(sql.NullInt64{Int64: tempInt, Valid: true}))
			}
			return intErr
		}

	case TyNullFloat64:
		if src == "" {
			dest.Set(reflect.ValueOf(sql.NullFloat64{}))
		} else {
			if floatErr == nil {
				dest.Set(reflect.ValueOf(sql.NullFloat64{Float64: tempFloat, Valid: true}))
			}
			return floatErr
		}

	case TyNullBool:
		if src == "" {
			dest.Set(reflect.ValueOf(sql.NullBool{}))
		} else {
			temp := slices.Contains(truthy, strings.ToLower(src))
			dest.Set(reflect.ValueOf(sql.NullBool{Bool: temp, Valid: true}))
		}

	case TyTime:
		if timeErr == nil {
			dest.Set(reflect.ValueOf(tempTime))
		}
		return timeErr
	case TyNullTime:
		if src == "" {
			dest.Set(reflect.ValueOf(sql.NullTime{}))
		} else {
			if timeErr == nil {
				dest.Set(reflect.ValueOf(sql.NullTime{Time: tempTime, Valid: true}))
			}
			return timeErr
		}

	default:
		return defaultCopy(dest, src)
	}
	return nil
}

func copyArray(dest reflect.Value, src []interface{}) error {
	if dest.Kind() != reflect.Slice {
		return defaultCopy(dest, src)
	}
	var err error
	tempSlice := reflect.MakeSlice(dest.Type(), len(src), len(src))
	for index, item := range src {
		err = RCopy(tempSlice.Index(index), item)
		if err != nil {
			return err
		}
	}
	dest.Set(tempSlice)
	return nil
}
