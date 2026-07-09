package parameter_coder

import (
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type NumberParameter struct {
	BasicParameter
}

func (param *NumberParameter) Copy() OracleParameterCoder {
	ret := new(NumberParameter)
	*ret = *param
	return ret
}

func (param *NumberParameter) Init() {
	param.SetDefault()
	if param.MaxLen == 0 {
		param.MaxLen = 0x16
	}
	//param.DataType = types.NUMBER
	//param.MaxLen = 0x16
}

func (param *NumberParameter) Encode(input interface{}, _ IConnection) error {
	param.Init()
	encoder := types.Number{}
	//encoder.SetDataType(param.DataType)
	err := encoder.SetValue(input)
	if err != nil {
		return err
	}
	if dt := encoder.GetDataType(); dt != 0 {
		param.DataType = dt
	}
	if param.MaxLen < encoder.GetMaxLen() {
		param.MaxLen = encoder.GetMaxLen()
	}
	param.BValue = encoder.Bytes()
	return nil
}

func (param *NumberParameter) Decode(_ IConnection) (interface{}, error) {
	decoder := types.Number{}
	decoder.SetBytes(param.BValue)
	decoder.SetDataType(param.DataType)
	return decoder.Value()
}

func (param *NumberParameter) Read(session network.SessionReader) error {
	var err error
	param.BValue, err = param.BasicRead(session)
	return err
}
