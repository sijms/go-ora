package go_ora

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func parseSqlText(text string) ([]string, error) {
	index := 0
	length := len(text)
	skip := false
	lineComment := false
	textBuffer := make([]byte, 0, len(text))
	for ; index < length; index++ {
		ch := text[index]
		switch ch {
		case '\\':
			// bypass next character
			index++
			continue
		case '/':
			if index+1 < length && text[index+1] == '*' {
				index += 1
				skip = true
			}
		case '*':
			if index+1 < length && text[index+1] == '/' {
				index += 1
				skip = false
			}
		case '\'':
			skip = !skip
		case '"':
			skip = !skip
		case '-':
			if index+1 < length && text[index+1] == '-' {
				index += 1
				lineComment = true
			}
		case '\n':
			if lineComment {
				lineComment = false
			}
		default:
			if skip || lineComment {
				continue
			}
			textBuffer = append(textBuffer, text[index])
		}
	}
	refinedSql := strings.TrimSpace(string(textBuffer))
	reg, err := regexp.Compile(`:(\w+)`)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, 10)
	matches := reg.FindAllStringSubmatch(refinedSql, -1)
	for _, match := range matches {
		if len(match) > 1 {
			names = append(names, match[1])
		}
	}
	return names, nil
}

func extractTag(tag string) (name, _type string, size int, direction ParameterDirection) {
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
	return tInteger(input) || tFloat(input) || input.Kind() == reflect.Bool
}

//======== get primitive data from original data types ========//

// get value to bypass pointer and sql.Null* values
func getValue(origVal driver.Value) (driver.Value, error) {
	if origVal == nil {
		return nil, nil
	}
	proVal := reflect.Indirect(reflect.ValueOf(origVal))
	if valuer, ok := proVal.Interface().(driver.Valuer); ok {
		return valuer.Value()
	}
	return proVal.Interface(), nil
}

// get prim string value from supported types
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

func getNumber(col interface{}) (interface{}, error) {
	var err error
	col, err = getValue(col)
	if err != nil {
		return int64(0), err
	}
	if col == nil {
		return int64(0), nil
	}
	rType := reflect.TypeOf(col)
	rValue := reflect.ValueOf(col)
	if tInteger(rType) {
		return rValue.Int(), nil
	}
	if f32, ok := col.(float32); ok {
		return strconv.ParseFloat(fmt.Sprint(f32), 64)
	}
	if tFloat(rType) {
		return rValue.Float(), nil
	}
	switch rType.Kind() {
	case reflect.Bool:
		if rValue.Bool() {
			return int64(1), nil
		} else {
			return int64(0), nil
		}
	case reflect.String:
		tempFloat, err := strconv.ParseFloat(rValue.String(), 64)
		if err != nil {
			return 0, err
		}
		return tempFloat, nil
	default:
		return 0, errors.New("conversion of unsupported type to number")
	}
}

// get prim float64 from supported types
func getFloat(col interface{}) (float64, error) {
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
		return float64(rValue.Int()), nil
	}
	if f32, ok := col.(float32); ok {
		return strconv.ParseFloat(fmt.Sprint(f32), 64)
	}
	if tFloat(rType) {
		return rValue.Float(), nil
	}
	switch rType.Kind() {
	case reflect.Bool:
		if rValue.Bool() {
			return 1, nil
		} else {
			return 0, nil
		}
	case reflect.String:
		tempFloat, err := strconv.ParseFloat(rValue.String(), 64)
		if err != nil {
			return 0, err
		}
		return tempFloat, nil
	default:
		return 0, errors.New("conversion of unsupported type to float")
	}
}

// get prim int64 value from supported types
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

// get prim time.Time from supported types
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

// get prim []byte from supported types
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

// get prim lob from supported types
func getLob(col interface{}, conn *Connection) (*Lob, error) {
	var err error
	col, err = getValue(col)
	if err != nil {
		return nil, err
	}
	if col == nil {
		return nil, nil
	}
	charsetID := conn.tcpNego.ServerCharset
	charsetForm := 1
	stringVar := ""
	var byteVar []byte
	switch val := col.(type) {
	case string:
		stringVar = val
	case Clob:
		stringVar = val.String
	case NVarChar:
		stringVar = string(val)
		charsetForm = 2
		charsetID = conn.tcpNego.ServernCharset
	case NClob:
		stringVar = val.String
		charsetForm = 2
		charsetID = conn.tcpNego.ServernCharset
	case []byte:
		byteVar = val
	case Blob:
		byteVar = val.Data
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

//=============================================================//

func setBytes(value reflect.Value, input []byte) error {
	switch value.Type() {
	case reflect.TypeOf([]byte{}):
		value.SetBytes(input)
	case reflect.TypeOf(Blob{}):
		value.Set(reflect.ValueOf(Blob{Data: input, Valid: true}))
	case reflect.TypeOf(string("")):
		value.SetString(string(input))
	default:
		return fmt.Errorf("can not assign []byte to type: %v", value.Type())
	}
	return nil
}
func setTime(value reflect.Value, input time.Time) error {
	switch value.Type() {
	case reflect.TypeOf(time.Time{}):
		value.Set(reflect.ValueOf(input))
	case reflect.TypeOf(TimeStamp{}):
		value.Set(reflect.ValueOf(TimeStamp(input)))
	case reflect.TypeOf(TimeStampTZ{}):
		value.Set(reflect.ValueOf(TimeStampTZ(input)))
	default:
		return fmt.Errorf("can not assign time to type: %v", value.Type())
	}
	return nil
}
func setString(value reflect.Value, input string) error {
	switch value.Kind() {
	case reflect.String:
		value.SetString(input)
	default:
		switch value.Type() {
		case reflect.TypeOf(NVarChar("")):
			value.Set(reflect.ValueOf(NVarChar(input)))
		case reflect.TypeOf(Clob{}):
			value.Set(reflect.ValueOf(Clob{String: input, Valid: true}))
		case reflect.TypeOf(NClob{}):
			value.Set(reflect.ValueOf(NClob{String: input, Valid: true}))
		default:
			return fmt.Errorf("can not assign string to type: %v", value.Type())
		}
	}
	return nil
}
func setNumber(value reflect.Value, input float64) error {
	if tSigned(value.Type()) {
		value.SetInt(int64(input))
		return nil
	}
	if tUnsigned(value.Type()) {
		value.SetUint(uint64(input))
		return nil
	}
	if tFloat(value.Type()) {
		value.SetFloat(input)
		return nil
	}
	switch value.Kind() {
	case reflect.Bool:
		if input == 0 {
			value.SetBool(false)
		} else {
			value.SetBool(true)
		}
		return nil
	case reflect.String:
		value.SetString(fmt.Sprintf("%v", input))
		return nil
	default:
		return fmt.Errorf("can not assign number to type: %v", value.Type())
	}
}

func isBadConn(err error) bool {
	return errors.Is(err, io.EOF) || errors.Is(err, syscall.EPIPE)
}
