package parameter_coder

import (
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type CursorParameter struct {
	BasicParameter
}

func (param *CursorParameter) Copy() OracleParameterCoder {
	ret := new(CursorParameter)
	*ret = *param
	return ret
}

func (param *CursorParameter) Encode(input interface{}, _ IConnection) error {
	param.SetDefault()
	param.DataType = types.REFCURSOR
	param.BValue = nil
	return nil
}

func (param *CursorParameter) Write(session network.SessionWriter) error {
	session.PutBytes(1, 0)
	return nil
}

func (param *CursorParameter) Read(session network.SessionReader) error {
	return nil
}

func (param *CursorParameter) Decode(_ IConnection) (interface{}, error) {
	return nil, nil
}
