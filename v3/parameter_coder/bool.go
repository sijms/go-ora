package parameter_coder

import (
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type BoolParameter struct {
	BasicParameter
}

func (param *BoolParameter) Copy() OracleParameterCoder {
	ret := new(BoolParameter)
	*ret = *param
	return ret
}

func (param *BoolParameter) Encode(input interface{}, _ IConnection) error {
	param.SetDefault()
	param.DataType = types.BOOLEAN
	encoder := &types.Bool{}
	encoder.SetDataType(param.DataType)
	err := encoder.SetValue(input)
	if err != nil {
		return err
	}
	if dt := encoder.GetDataType(); dt != 0 {
		param.DataType = dt
	}
	param.BValue = encoder.Bytes()
	return nil
}

func (param *BoolParameter) Decode(_ IConnection) (interface{}, error) {
	decoder := &types.Bool{}
	decoder.SetBytes(param.BValue)
	decoder.SetDataType(param.DataType)
	return decoder.Value()
}

func (param *BoolParameter) Write(session network.SessionWriter) error {
	session.PutClr(param.BValue)
	return nil
}

func (param *BoolParameter) Read(session network.SessionReader) error {
	var err error
	param.BValue, err = param.BasicRead(session)
	return err
}
