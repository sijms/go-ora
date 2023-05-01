package go_ora

import (
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
	return tInteger(input) || tFloat(input)
}

// try to get string data from row field
func getString(col interface{}) string {
	if temp, ok := col.(string); ok {
		return temp
	} else {
		return fmt.Sprintf("%v", col)
	}
}

// try to get float64 data from row field
func getFloat(col interface{}) (float64, error) {
	rType := reflect.TypeOf(col)
	rValue := reflect.ValueOf(col)
	if tInteger(rType) {
		return float64(rValue.Int()), nil
	}
	if tFloat(rType) {
		return rValue.Float(), nil
	}
	switch rType.Kind() {
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

// try to get int64 value from the row field
func getInt(col interface{}) (int64, error) {
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

func getDate(col interface{}) (time.Time, error) {
	switch val := col.(type) {
	case time.Time:
		return val, nil
	case string:
		return time.Parse(time.RFC3339, val)
	default:
		return time.Time{}, errors.New("conversion of unsupported type to time.Time")
	}
}

func getBytes(col interface{}) ([]byte, error) {
	switch val := col.(type) {
	case []byte:
		return val, nil
	case string:
		return []byte(val), nil
	default:
		return nil, errors.New("conversion of unsupported type to []byte")
	}
}
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
