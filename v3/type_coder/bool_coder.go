package type_coder

import (
	"bytes"
	"fmt"

	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type BoolCoder struct {
	TypeInfo
}

func NewBoolCoder(data interface{}) (*BoolCoder, error) {
	ret := new(BoolCoder)
	ret.SetDefault()
	ret.DataType = types.BOOLEAN
	if data == nil {
		ret.BValue = nil
		return ret, nil
	}
	var value bool
	switch v := data.(type) {
	case bool:
		value = v
	case *bool:
		value = *v
	default:
		return nil, fmt.Errorf("bool coder: unsupported data type: %T", data)
	}
	if value {
		ret.BValue = []byte{1, 1}
	} else {
		ret.BValue = []byte{1, 0}
	}
	return ret, nil
}

func (coder *BoolCoder) Decode(data []byte) (interface{}, error) {
	if data == nil {
		return nil, nil
	}
	return bytes.Equal(data, []byte{1, 1}), nil
}

func (coder *BoolCoder) Read(session network.SessionReader) (interface{}, error) {
	bValue, err := coder.basicRead(session)
	if err != nil {
		return nil, err
	}
	return coder.Decode(bValue)
}

func (coder *BoolCoder) Write(session network.SessionWriter) error {
	session.PutClr(coder.BValue)
	return nil
}
