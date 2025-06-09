package go_ora

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// ======== get primitive data from original data types ========//

// get value to bypass pointer and sql.Null* values
func getValue(origVal driver.Value) (driver.Value, error) {
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

// get string value from supported types
func getString(col interface{}) string {
	col, _ = getValue(col)
	if col == nil {
		return ""
	}
	switch val := col.(type) {
	case Clob:
		return val.String
	case NClob:
		return val.String
	}
	if temp, ok := col.(string); ok {
		return temp
	} else {
		return fmt.Sprintf("%v", col)
	}
}

// get bool value from supported types
func getBool(col interface{}) (bool, error) {
	col, err := getValue(col)
	if err != nil {
		return false, err
	}
	if col == nil {
		return false, nil
	}
	rValue := reflect.ValueOf(col)
	return rValue.Bool(), nil
}

// get int64 value from supported types
func getInt(col interface{}) (int64, error) {
	var err error
	col, err = getValue(col)
	if err != nil {
		return 0, err
	}
	if col == nil {
		return 0, nil
	}
	rType := reflect.TypeOf(col)
	rValue := reflect.ValueOf(col)
	if tInteger(rType) {
		return rValue.Int(), nil
	}
	if tFloat(rType) {
		return int64(rValue.Float()), nil
	}
	switch rType.Kind() {
	case reflect.String:
		tempInt, err := strconv.ParseInt(rValue.String(), 10, 64)
		if err != nil {
			return 0, err
		}
		return tempInt, nil
	case reflect.Bool:
		if rValue.Bool() {
			return 1, nil
		} else {
			return 0, nil
		}
	default:
		return 0, errors.New("conversion of unsupported type to int")
	}
}

// get time.Time from supported types
func getDate(col interface{}) (time.Time, error) {
	var err error
	col, err = getValue(col)
	if err != nil {
		return time.Time{}, err
	}
	if col == nil {
		return time.Time{}, nil
	}
	switch val := col.(type) {
	case time.Time:
		return val, nil
	case TimeStamp:
		return time.Time(val), nil
	case TimeStampTZ:
		return time.Time(val), nil
	case string:
		return time.Parse(time.RFC3339, val)
	default:
		return time.Time{}, errors.New("conversion of unsupported type to time.Time")
	}
}

// get []byte from supported types
func getBytes(col interface{}) ([]byte, error) {
	var err error
	col, err = getValue(col)
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, nil
	}
	switch val := col.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	case Blob:
		return val.Data, nil
	default:
		return nil, errors.New("conversion of unsupported type to []byte")
	}
}

// get lob from supported types
func getLob(col interface{}, conn *Connection) (*Lob, error) {
	var err error
	col, err = getValue(col)
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, nil
	}
	charsetID := conn.getDefaultCharsetID()
	charsetForm := 1
	stringVar := ""
	var byteVar []byte
	switch val := col.(type) {
	case string:
		stringVar = val
	case Clob:
		if !val.Valid {
			return nil, nil
		}
		stringVar = val.String
	case NVarChar:
		stringVar = string(val)
		charsetForm = 2
		charsetID = conn.tcpNego.ServernCharset
	case NClob:
		charsetForm = 2
		charsetID = conn.tcpNego.ServernCharset
		if !val.Valid {
			return nil, nil
		}
		stringVar = val.String
	case []byte:
		byteVar = val
	case Blob:
		byteVar = val.Data
	case Vector:
		byteVar, err = val.encode()
		if err != nil {
			return nil, err
		}
	}
	if len(stringVar) > 0 {
		lob := newLob(conn)
		err = lob.createTemporaryClob(charsetID, charsetForm)
		if err != nil {
			return nil, err
		}
		err = lob.putString(stringVar)
		return lob, err
	}
	if len(byteVar) > 0 {
		lob := newLob(conn)
		err = lob.createTemporaryBLOB()
		if err != nil {
			return nil, err
		}
		err = lob.putData(byteVar)
		return lob, err
	}
	return nil, nil
}
