package go_ora

import "strings"

func parseSqlText(text string) []string {
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
	split := func(r rune) bool {
		return r == ' ' || r == '\t' || r == ',' || r == '+' || r == '-' ||
			r == '*' || r == '/' || r == '<' || r == '>' ||
			r == '=' || r == '|' || r == '(' || r == ')'
	}
	pars := make([]string, 0, 10)
	words := strings.FieldsFunc(refinedSql, split)
	for _, word := range words {
		if len(word) > 1 && word[0] == ':' {
			pars = append(pars, word[1:])
		}
	}
	return pars
}
