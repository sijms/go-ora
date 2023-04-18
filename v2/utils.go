package go_ora

import (
	"regexp"
	"strconv"
	"strings"
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
	tag = strings.TrimSpace(tag)
	direction = Input
	extractKeyValue := func(text string) bool {
		parts := strings.Split(text, "=")
		if len(parts) != 2 {
			return false
		}
		id := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		switch id {
		case "name":
			name = value
		case "type":
			_type = value
		case "size":
			temp, _ := strconv.ParseInt(value, 10, 32)
			size = int(temp)
		case "direction":
			fallthrough
		case "dir":
			value = strings.ToLower(value)
			switch value {
			case "in":
				fallthrough
			case "input":
				direction = Input
			case "out":
				fallthrough
			case "output":
				direction = Output
			case "inout":
				direction = InOut
			}
		}
		return true
	}
	if len(tag) == 0 {
		return
	}
	tagFields := strings.Split(tag, ",")
	if len(tagFields) > 0 {
		if !extractKeyValue(tagFields[0]) {
			name = tagFields[0]
		}
	}
	if len(tagFields) > 1 {
		if !extractKeyValue(tagFields[1]) {
			_type = tagFields[1]
		}
	}
	if len(tagFields) > 2 {
		if !extractKeyValue(tagFields[2]) {
			temp, _ := strconv.ParseInt(tagFields[2], 10, 32)
			size = int(temp)
		}
	}
	if len(tagFields) > 3 {
		if !extractKeyValue(tagFields[3]) {
			dir := strings.ToLower(tagFields[3])
			switch dir {
			case "in":
				fallthrough
			case "input":
				direction = Input
			case "out":
				fallthrough
			case "output":
				direction = Output
			case "inout":
				direction = InOut
			}
		}
	}
	return
}
