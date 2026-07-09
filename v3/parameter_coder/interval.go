package parameter_coder

import (
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type IntervalParameter struct {
	BasicParameter
}

func (param *IntervalParameter) Copy() OracleParameterCoder {
	ret := new(IntervalParameter)
	*ret = *param
	return ret
}

func (param *IntervalParameter) Init() {
	param.SetDefault()
}

func (param *IntervalParameter) Encode(input interface{}, _ IConnection) error {
	param.Init()
	encoder := &types.Interval{}
	encoder.SetDataType(param.DataType)
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

func (param *IntervalParameter) Decode(_ IConnection) (interface{}, error) {
	decoder := &types.Interval{}
	decoder.SetBytes(param.BValue)
	return decoder.Value()
}

func (param *IntervalParameter) Read(session network.SessionReader) error {
	var err error
	param.BValue, err = param.BasicRead(session)
	return err
}
