package oson

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sijms/go-ora/v3/types"
)

// Op codes of Oracle's scalars that JSON has no type for.
const (
	opTimeStamp   = 0x39
	opBinary16    = 0x3A
	opBinary32    = 0x3B
	opDate        = 0x3C
	opIntervalYM  = 0x3D
	opIntervalDS  = 0x3E
	opExtended    = 0x7B
	opTimeStampTZ = 0x7C
	opTimeStamp7  = 0x7D
	opID          = 0x7E
)

var dateLengths = map[uint8]int{opDate: 7, opTimeStamp7: 7, opTimeStamp: 11, opTimeStampTZ: 13}

func readFixed(buffer *bytes.Reader, length int) ([]byte, error) {
	data := make([]byte, length)
	_, err := io.ReadFull(buffer, data)
	return data, err
}

// rawField is a RAW value or an ID.
type rawField struct {
	value []byte
	basicField
}

func (field *rawField) Value() (interface{}, error) {
	return field.value, nil
}

func (field *rawField) Encode() ([]byte, error) {
	return nil, errors.New("oson: encoding binary values is not supported")
}

type intervalField struct {
	value types.Interval
	basicField
}

// Value is the interval as Oracle writes it, as a column of its type reads.
func (field *intervalField) Value() (interface{}, error) {
	return field.value.Text()
}

func (field *intervalField) Encode() ([]byte, error) {
	return nil, errors.New("oson: encoding intervals is not supported")
}

// Text returns the JSON text of an OSON image as Oracle's JSON_SERIALIZE writes
// it: members in the order the image keeps them, numbers with all their digits,
// and the scalars JSON has no type for as strings: a date or a timestamp in
// ISO 8601, an interval as an ISO 8601 duration, and binary as hexadecimal.
func Text(data []byte) ([]byte, error) {
	buffer := bytes.NewReader(data)
	header := &header{}
	err := header.read(buffer)
	if err != nil {
		return nil, err
	}
	node, err := decodeNode(buffer, header)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	err = writeText(&out, node, header)
	if err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func writeText(out *bytes.Buffer, field Field, header *header) error {
	switch field := field.(type) {
	case *objectField:
		out.WriteByte('{')
		for i, child := range field.Children() {
			if i > 0 {
				out.WriteByte(',')
			}
			index := child.KeyIndex() - 1
			if index < 0 || index >= len(header.keys) {
				return fmt.Errorf("oson: key index %d is out of range", index+1)
			}
			writeString(out, header.keys[index].name)
			out.WriteByte(':')
			if err := writeText(out, child, header); err != nil {
				return err
			}
		}
		out.WriteByte('}')
	case *arrayField:
		out.WriteByte('[')
		for i, child := range field.Children() {
			if i > 0 {
				out.WriteByte(',')
			}
			if err := writeText(out, child, header); err != nil {
				return err
			}
		}
		out.WriteByte(']')
	case *nullField:
		out.WriteString("null")
	case *booleanField:
		out.WriteString(strconv.FormatBool(field.value))
	case *stringField:
		writeString(out, field.value)
	case *numberField:
		text, err := field.value.String()
		if err != nil {
			return err
		}
		out.WriteString(jsonNumber(text))
	case *binaryDoubleField:
		value, err := field.Value()
		if err != nil {
			return err
		}
		writeFloat(out, value.(float64), 64)
	case *binaryFloatField:
		value, err := field.Value()
		if err != nil {
			return err
		}
		writeFloat(out, float64(value.(float32)), 32)
	case *dateField:
		value, err := field.Value()
		if err != nil {
			return err
		}
		writeString(out, dateText(value.(time.Time), field.opCode))
	case *intervalField:
		text, err := field.value.Text()
		if err != nil {
			return err
		}
		writeString(out, isoDuration(text.(string)))
	case *rawField:
		writeString(out, strings.ToUpper(hex.EncodeToString(field.value)))
	default:
		return fmt.Errorf("oson: no text for a field of type %T", field)
	}
	return nil
}

// jsonNumber writes a number as Oracle does in JSON: in full, but in E notation
// where its first digit is at 10^40 or above, or at 10^-41 or below.
func jsonNumber(text string) string {
	sign := ""
	if strings.HasPrefix(text, "-") {
		sign, text = "-", text[1:]
	}
	whole, fraction, _ := strings.Cut(text, ".")
	var digits string
	var exponent int
	if strings.TrimLeft(whole, "0") != "" {
		whole = strings.TrimLeft(whole, "0")
		digits, exponent = whole+fraction, len(whole)-1
	} else {
		zeros := len(fraction) - len(strings.TrimLeft(fraction, "0"))
		digits, exponent = fraction[zeros:], -(zeros + 1)
	}
	digits = strings.TrimRight(digits, "0")
	if digits == "" {
		return "0"
	}
	if exponent < 40 && exponent > -41 {
		if whole == "" {
			whole = "0"
		}
		if fraction == "" {
			return sign + whole
		}
		return sign + whole + "." + fraction
	}
	mantissa := digits[:1]
	if len(digits) > 1 {
		mantissa += "." + digits[1:]
	}
	if exponent > 0 {
		return sign + mantissa + "E+" + strconv.Itoa(exponent)
	}
	return sign + mantissa + "E" + strconv.Itoa(exponent)
}

func writeFloat(out *bytes.Buffer, value float64, bitSize int) {
	text := strconv.FormatFloat(value, 'f', -1, bitSize)
	if math.IsInf(value, 0) || math.IsNaN(value) {
		writeString(out, text)
		return
	}
	out.WriteString(text)
}

// dateText writes a timestamp's fraction to six digits, or nine where it has
// nanoseconds.
func dateText(value time.Time, opCode uint8) string {
	if opCode == opDate || opCode == opTimeStamp7 {
		return value.Format("2006-01-02T15:04:05")
	}
	layout := "2006-01-02T15:04:05.000000"
	if value.Nanosecond()%1000 != 0 {
		layout = "2006-01-02T15:04:05.000000000"
	}
	if opCode == opTimeStampTZ {
		layout += "Z07:00"
	}
	return value.Format(layout)
}

// isoDuration turns an interval as Oracle writes it, "+03-02" or
// "+04 05:06:07.890000", into ISO 8601's "P3Y2M" or "P4DT5H6M7.890000S",
// leaving out the parts that are zero.
func isoDuration(text string) string {
	sign := ""
	if strings.HasPrefix(text, "-") {
		sign = "-"
	}
	text = strings.TrimLeft(text, "+-")
	var b strings.Builder
	part := func(value, unit string) {
		if value = trimZeros(value); value != "0" {
			b.WriteString(value + unit)
		}
	}
	days, clock, ok := strings.Cut(text, " ")
	if !ok {
		years, months, _ := strings.Cut(text, "-")
		part(years, "Y")
		part(months, "M")
		if b.Len() == 0 {
			return "P0Y"
		}
		return sign + "P" + b.String()
	}
	part(days, "D")
	parts := strings.SplitN(clock, ":", 3)
	if len(parts) == 3 {
		seconds, fraction, _ := strings.Cut(parts[2], ".")
		if strings.Trim(fraction, "0") != "" {
			seconds += "." + fraction
		}
		date := b.Len()
		part(parts[0], "H")
		part(parts[1], "M")
		part(seconds, "S")
		if b.Len() > date {
			return sign + "P" + b.String()[:date] + "T" + b.String()[date:]
		}
	}
	if b.Len() == 0 {
		return "P0D"
	}
	return sign + "P" + b.String()
}

// trimZeros drops a number's leading zeros, keeping one before its point.
func trimZeros(number string) string {
	trimmed := strings.TrimLeft(number, "0")
	if trimmed == "" || trimmed[0] == '.' {
		return "0" + trimmed
	}
	return trimmed
}

// writeString writes a JSON string as Oracle does: quotes, backslashes, and
// control characters escaped, everything else as it is.
func writeString(out *bytes.Buffer, value string) {
	out.WriteByte('"')
	for i := 0; i < len(value); {
		c := value[i]
		if c >= 0x20 && c != 0x7F && c != '"' && c != '\\' {
			if c < utf8.RuneSelf {
				out.WriteByte(c)
				i++
				continue
			}
			r, size := utf8.DecodeRuneInString(value[i:])
			if r == utf8.RuneError && size == 1 {
				out.WriteString(`�`)
			} else {
				out.WriteString(value[i : i+size])
			}
			i += size
			continue
		}
		switch c {
		case '"':
			out.WriteString(`\"`)
		case '\\':
			out.WriteString(`\\`)
		case '\n':
			out.WriteString(`\n`)
		case '\r':
			out.WriteString(`\r`)
		case '\t':
			out.WriteString(`\t`)
		case '\b':
			out.WriteString(`\b`)
		case '\f':
			out.WriteString(`\f`)
		default:
			fmt.Fprintf(out, `\u%04X`, c)
		}
		i++
	}
	out.WriteByte('"')
}
