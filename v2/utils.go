package go_ora

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
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
	switch rType.Kind() {
	case reflect.Float32, reflect.Float64:
		return rValue.Float(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rValue.Int()), nil
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
	switch rType.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rValue.Int(), nil
	case reflect.Float32, reflect.Float64:
		return int64(rValue.Float()), nil
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
