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
	if len(tag) == 0 {
		return
	}
	tagFields := strings.Split(tag, ",")
	if len(tagFields) > 0 {
		name = tagFields[0]
	}
	if len(tagFields) > 1 {
		_type = tagFields[1]
	}
	if len(tagFields) > 2 {
		temp, _ := strconv.ParseInt(tagFields[2], 10, 32)
		size = int(temp)
	}
	if len(tagFields) > 3 {
		dir := strings.ToLower(tagFields[3])
		if dir == "in" || dir == "input" {
			direction = Input
		} else if dir == "out" || dir == "output" {
			direction = Output
		} else if dir == "inout" {
			direction = InOut
		}
	}
	return
}
