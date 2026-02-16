package type_coder

import (
	"database/sql"
	"errors"

	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type StringCoder struct {
	data sql.NullString
	TypeInfo
}

func NewStringCoder(data interface{}, useNCharset bool) (*StringCoder, error) {
	ret := &StringCoder{}
	ret.SetDefault()
	ret.ContFlag = 0x10
	ret.CharsetForm = 1
	if useNCharset {
		ret.CharsetForm = 2
	}
	if data == nil {
		ret.data = sql.NullString{}
		return ret, nil
	}
	switch v := data.(type) {
	case string:
		ret.data = sql.NullString{String: v, Valid: true}
	case *string:
		ret.data = sql.NullString{String: *v, Valid: true}
	case sql.NullString:
		ret.data = v
	case *sql.NullString:
		ret.data = *v
	default:
		return nil, errors.New("data is not a string")
	}
	return ret, nil
}

func (coder *StringCoder) Encode(strConv converters.StringCoder, _ types.LobStreamer) error {
	conv, err := strConv.GetStringCoder(coder.CharsetID, coder.CharsetForm)
	if err != nil {
		return err
	}
	if coder.CharsetID == 0 {
		coder.CharsetID = conv.GetLangID()
	}
	if coder.data.Valid && len(coder.data.String) > 0 {
		coder.BValue = conv.Encode(coder.data.String)
	}
	coder.MaxLen = int64(len(coder.BValue))
	coder.MaxCharLen = int64(len(coder.BValue))
	maxLen := strConv.GetMaxStringLength()
	if coder.MaxLen > maxLen {
		coder.DataType = types.LongVarChar
	} else {
		coder.DataType = types.NCHAR
	}
	if coder.MaxLen == 0 {
		coder.MaxLen = 1
	}
	return nil
}

func (coder *StringCoder) Write(session network.SessionWriter) error {
	session.PutClr(coder.BValue)
	return nil
}
func (coder *StringCoder) DecodeString(data []byte) (sql.NullString, error) {
	ret := sql.NullString{}
	if coder.Coder == nil {
		return ret, errors.New("string coder should be set first before encoding/decoding strings")
	}
	conv, err := coder.Coder.GetStringCoder(coder.CharsetID, coder.CharsetForm)
	if err != nil {
		return ret, err
	}
	if data != nil {
		ret.Valid = true
		ret.String = conv.Decode(data)
	}
	return ret, nil
}

func (coder *StringCoder) Decode(data []byte) (interface{}, error) {
	if data == nil {
		return nil, nil
	}
	ret, err := coder.DecodeString(data)
	if err != nil {
		return nil, err
	}
	if !ret.Valid {
		return nil, err
	}
	return ret.String, err
}

func (coder *StringCoder) Read(session network.SessionReader) (interface{}, error) {
	bValue, err := coder.basicRead(session)
	if err != nil {
		return nil, err
	}
	return coder.Decode(bValue)
}
