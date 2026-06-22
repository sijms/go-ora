package types

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
)

type Raw struct {
	Basic
}

func (raw *Raw) Value() (interface{}, error) {
	return raw.bValue, nil
}
func (raw *Raw) SetValue(input interface{}) error {
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

func (raw *Raw) Scan(input interface{}) error {
	return raw.SetValue(input)
}

func (raw *Raw) CopyTo(dest driver.Value) error {
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
