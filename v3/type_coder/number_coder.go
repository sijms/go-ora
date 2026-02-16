package type_coder

import (
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type NumberCoder struct {
	TypeInfo
}

func NewNumberCoder(number *types.Number) (*NumberCoder, error) {
	ret := &NumberCoder{}
	ret.BValue = number.Data
	// set type info
	ret.Flag = 3
	ret.DataType = types.NUMBER
	ret.MaxLen = 0x16
	return ret, nil
}
func (coder *NumberCoder) Write(session network.SessionWriter) error {
	session.PutClr(coder.BValue)
	return nil
}

func (coder *NumberCoder) Decode(data []byte) (interface{}, error) {
	if data == nil {
		return nil, nil
	}
	ret := types.Number{Data: data}
	return ret.String()
}

func (coder *NumberCoder) Read(session network.SessionReader) (interface{}, error) {
	bValue, err := coder.basicRead(session)
	if err != nil {
		return nil, err
	}
	return coder.Decode(bValue)
}
