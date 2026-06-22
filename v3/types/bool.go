package types

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"fmt"
)

type Bool struct {
	Basic
}

func (bl *Bool) SetValue(input interface{}) error {
	if input == nil {
		bl.bValue = nil
		return nil
	}
	switch value := input.(type) {
	case Bool:
		*bl = value
	case *Bool:
		*bl = *value
	case bool:
		if value {
			bl.bValue = []byte{1, 1}
		} else {
			bl.bValue = []byte{1, 0}
		}
	case *bool:
		if *value {
			bl.bValue = []byte{1, 1}
		} else {
			bl.bValue = []byte{1, 0}
		}
	default:
		return fmt.Errorf("cannot set value of type %T into Bool", input)
	}
	return nil
}

func (bl *Bool) Value() (interface{}, error) {
	if bl.bValue == nil {
		return nil, nil
	}
	if bytes.Equal(bl.bValue, []byte{1, 1}) {
		return true, nil
	}
	return false, nil
}

func (bl *Bool) Scan(input interface{}) error {
	return bl.SetValue(input)
}

func (bl *Bool) CopyTo(dest driver.Value) error {
	val, err := bl.Value()
	if err != nil {
		return err
	}
	switch dst := dest.(type) {
	case *Bool:
		*dst = *bl
	case *bool:
		if val != nil {
			*dst = val.(bool)
		} else {
			*dst = false
		}
	case *sql.NullBool:
		if val != nil {
			dst.Bool = val.(bool)
			dst.Valid = true
		} else {
			dst.Valid = false
		}
	case *int:
		if val != nil && val.(bool) {
			*dst = 1
		} else {
			*dst = 0
		}
	case *int64:
		if val != nil && val.(bool) {
			*dst = 1
		} else {
			*dst = 0
		}
	case *string:
		if val != nil && val.(bool) {
			*dst = "true"
		} else {
			*dst = "false"
		}
	default:
		return fmt.Errorf("cannot copy Bool to variable of type %T", dest)
	}
	return nil
}
