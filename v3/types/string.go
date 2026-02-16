package types

import (
	"database/sql"
	"database/sql/driver"
)

type String interface {
	OracleType
}

type oracleString struct {
	data  string
	valid bool
}

func (os *oracleString) CopyTo(dest driver.Value) (supported bool, err error) {
	switch dst := dest.(type) {
	case *oracleString:
		*dst = *os
	case *string:
		if os.valid {
			*dst = os.data
		} else {
			*dst = ""
		}
	case *sql.NullString:
		dst.Valid = os.valid
		dst.String = os.data
	default:
		return false, nil
	}
	return true, nil
}
