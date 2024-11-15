package tns

import (
	"fmt"
	"strconv"
	"strings"
)

type parser struct {
	tns           []rune
	pos           int // next position to read
	lowerCaseKeys bool
}

// consume consumes TNS names
func (p *parser) consume() (any, error) {
	// Iterate over runes in TNS string
	var entries []any
	for {
		if p.isEndReached() {
			break
		}
		// Consume a TNS mapping entry
		entry, err := p.consumeMappingEntry()
		if err != nil {
			return nil, fmt.Errorf("failed to consume a mapping entry: %w", err)
		}
		entries = append(entries, entry)
	}
	// Return a TNS-like object with TNS mapping entries
	return map[string]any{"entries": entries}, nil
}

// consumeMappingEntry consumes a "net service name" -> "connect descriptor" entry
func (p *parser) consumeMappingEntry() (any, error) {
	// Consume whitespace
	_ = p.consumeWhitespace()
	// Consume a net service name
	netServiceName, err := p.consumeStringValue('=')
	if err != nil {
		return nil, fmt.Errorf("failed to consume a net service name: %w", err)
	}
	// Consume '='
	if r := p.consumeRune(); r != '=' {
		return nil, fmt.Errorf("expected '=', got '%s'", string(r))
	}
	// Consume a connect descriptor
	connDesc, err := p.consumeValue()
	if err != nil {
		return nil, fmt.Errorf("failed to consume a connect descriptor: %w", err)
	}
	// Return a TNS mapping entry
	return map[string]any{
		"net_service_name":   netServiceName,
		"connect_descriptor": connDesc,
	}, nil
}

// consumeValue consumes a TNS value in brackets '()'.
//
// Such value can be a whole connect descriptor or a parameter entry.
func (p *parser) consumeValue() (any, error) {
	_ = p.consumeWhitespace()
	switch p.lookAhead() {
	case '(':
		return p.consumeObjectLiteral()
	default:
		return p.consumeScalarValue(')')
	}
}

// consumeStringValue consumes string literal until terminator rune, not consuming the terminator rune itself.
//
// Initial and trailing whitespace is ignored.
//
// It returns the string and an error if the string is not terminated.
func (p *parser) consumeStringValue(terminator rune) (string, error) {
	var str strings.Builder
	for {
		r := p.lookAhead()
		if r == terminator {
			break
		}
		if r == stopRune {
			return "", fmt.Errorf("unexpected end of string")
		}
		_ = p.consumeRune()
		str.WriteRune(r)
	}
	// ignore whitespace
	return strings.TrimSpace(str.String()), nil
}

// consumeScalarValue consumes string literal until terminator rune, not consuming the terminator rune itself
//
// Initial and trailing whitespace is ignored.
//
// The type is automatically inferred:
// - abc -> string("abc")
// - 123 -> int(123)
// - 123.45 -> float64(123.45)
func (p *parser) consumeScalarValue(terminator rune) (any, error) {
	data, err := p.consumeStringValue(terminator)
	if err != nil {
		return nil, err
	}
	// try to parse as int
	if i, err := strconv.Atoi(data); err == nil {
		return i, nil
	}
	// try to parse as float
	if f, err := strconv.ParseFloat(data, 64); err == nil {
		return f, nil
	}
	// not a number —> it's a string
	return data, nil
}

// consumeObjectLiteral consumes object and returns it as a map
//
// Here are the rules:
// - Single key: (key=value) -> map[key] = value
// - Different keys make a map: (key1=value1)(key2=value2) -> map[key1] = value1, map[key2] = value2
// - Repetition of the same key makes array of values: (key1=value1)(key1=value2) -> map[key1] = [value1, value2]
//
// Whitespace is ignored.
//
// Value can be either string or another object. Thus it can be recursive.
func (p *parser) consumeObjectLiteral() (map[string]any, error) {
	obj := make(map[string][]any)

	for {
		_ = p.consumeWhitespace()

		// initial rune of the object
		r := p.lookAhead()
		if r != '(' {
			return p.compressObject(obj), nil
		}

		// ok, we have a key-value pair
		_ = p.consumeRune()

		// skip leading whitespace
		_ = p.consumeWhitespace()

		key, err := p.consumeStringValue('=')
		if err != nil {
			return nil, err
		}

		if key == "" {
			return nil, fmt.Errorf("empty key")
		}

		if p.lowerCaseKeys {
			key = strings.ToLower(key)
		}

		if p.consumeRune() != '=' {
			return nil, fmt.Errorf("expected '='")
		}

		value, err := p.consumeValue()
		if err != nil {
			return nil, err
		}

		if p.consumeRune() != ')' {
			return nil, fmt.Errorf("expected ')'")
		}

		obj[key] = append(obj[key], value)
	}
}

// FIXME: this is a temporary solution, does not take into account hierarchy
var alwaysArray = map[string]struct{}{
	"address":      {},
	"address_list": {},
}

func (p *parser) compressObject(in map[string][]any) map[string]any {
	out := make(map[string]any)
	for k, v := range in {
		if _, ok := alwaysArray[strings.ToLower(k)]; ok {
			out[k] = v
			continue
		}
		if len(v) == 1 {
			out[k] = v[0]
		} else {
			out[k] = v
		}
	}
	return out
}
