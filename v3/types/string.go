package types

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/sijms/go-ora/v3/converters"
)

type String struct {
	bValue      []byte
	Conv        converters.IStringConverter
	UseNCharset bool
}

func (str *String) Value(_ uint16) (interface{}, error) {
	if str.bValue == nil {
		return nil, nil
	}
	if str.Conv == nil {
		str.Conv = converters.NewStringConverter(0x7D0)
	}
	return str.Conv.Decode(str.bValue), nil
}

func (str *String) SetValue(input interface{}, _ uint16) error {
	if input == nil {
		str.bValue = nil
		return nil
	}
	if str.Conv == nil {
		str.Conv = converters.NewStringConverter(0x7D0)
	}
	switch data := input.(type) {
	case String:
		if str.Conv.GetLangID() == data.Conv.GetLangID() {
			*str = data
		} else {
			temp, err := data.Value(0)
			if err != nil {
				return err
			}
			return str.SetValue(temp, 0)
		}

	case *String:
		if str.Conv.GetLangID() == data.Conv.GetLangID() {
			*str = *data
		} else {
			temp, err := data.Value(0)
			if err != nil {
				return err
			}
			return str.SetValue(temp, 0)
		}

	case string:
		str.bValue = str.Conv.Encode(data)
	case *string:
		str.bValue = str.Conv.Encode(*data)
	case sql.NullString:
		if data.Valid {
			str.bValue = str.Conv.Encode(data.String)
		} else {
			str.bValue = nil
		}
	case *sql.NullString:
		if data.Valid {
			str.bValue = str.Conv.Encode(data.String)
		} else {
			str.bValue = nil
		}
	default:
		return fmt.Errorf("cannot set value of type %T into string", input)
	}
	return nil
}
func (str *String) Bytes() []byte {
	return str.bValue
}
func (str *String) SetBytes(input []byte) {
	str.bValue = input
}

func (str *String) Scan(input interface{}) error {
	return str.SetValue(input, 0)
}

//type String interface {
//	OracleType
//}

//type oracleString struct {
//	data  string
//	valid bool
//}

func (str *String) CopyTo(dest driver.Value) (err error) {
	value, err := str.Value(0)
	if err != nil {
		return err
	}
	switch dst := dest.(type) {
	case *string:
		if value != nil {
			*dst = value.(string)
		} else {
			*dst = ""
		}
	case *sql.NullString:
		if value != nil {
			dst.String = value.(string)
			dst.Valid = true
		} else {
			dst.Valid = false
		}
	case *[]byte:
		if value != nil {
			*dst = []byte(value.(string))
		} else {
			*dst = nil
		}
	case *String:
		*dst = *str
	default:
		return fmt.Errorf("cannot copy String to variable of type %T", dest)
	}
	return nil
}
