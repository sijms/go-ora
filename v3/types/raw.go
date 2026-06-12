package types

import (
	"database/sql"
	"fmt"
)

type Raw struct {
	bValue []byte
}

func (raw *Raw) Value(_ uint16) (interface{}, error) {
	return raw.bValue, nil
}
func (raw *Raw) SetValue(input interface{}, _ uint16) error {
	if input == nil {
		raw.bValue = nil
		return nil
	}
	switch data := input.(type) {
	case []byte:
		raw.bValue = data
	case *[]byte:
		raw.bValue = *data
	case string:
		raw.bValue = []byte(data)
	case *string:
		raw.bValue = []byte(*data)
	default:
		return fmt.Errorf("cannot set value of type %T into Blob", input)
	}
	return nil
}

func (raw *Raw) Bytes() []byte {
	return raw.bValue
}
func (raw *Raw) SetBytes(input []byte) {
	raw.bValue = input
}

func (raw *Raw) Scan(input interface{}) error {
	return raw.SetValue(input, 0)
}

func (raw *Raw) CopyTo(dest interface{}) error {
	switch dst := dest.(type) {
	case *[]byte:
		*dst = raw.bValue
	case *string:
		*dst = string(raw.bValue)
	case *sql.NullString:
		if raw.bValue == nil {
			*dst = sql.NullString{Valid: false}
		} else {
			*dst = sql.NullString{String: string(raw.bValue), Valid: true}
		}
	case *Raw:
		*dst = *raw
	default:
		return fmt.Errorf("cannot copy Raw to variable of type %T", dest)
	}
	return nil
}
