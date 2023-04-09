package go_ora

import (
	"regexp"
	"strings"
)

func parseSqlText(text string) ([]string, error) {
	index := 0
	length := len(text)
	skip := false
	lineComment := false
	output := make([]byte, 0, len(text))
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
			output = append(output, text[index])
		}
	}
	refinedSql := strings.TrimSpace(string(output))
	reg, err := regexp.Compile(`:\w+`)
	if err != nil {
		return nil, err
	}
	return reg.FindAllString(refinedSql, -1), nil
}
