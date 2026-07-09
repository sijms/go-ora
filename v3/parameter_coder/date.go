package parameter_coder

import (
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type DateParameter struct {
	BasicParameter
}

func (param *DateParameter) Copy() OracleParameterCoder {
	ret := new(DateParameter)
	*ret = *param
	return ret
}

func (param *DateParameter) Init() {
	param.SetDefault()
}

func (param *DateParameter) Encode(input interface{}, conn IConnection) error {
	param.Init()
	encoder := types.Date{}
	encoder.AsUTC = conn.SendTimeZoneAsUTC()
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
	//if len(param.BValue) > 0 {
	//	param.MaxLen = int64(len(param.BValue))
	//}
	return nil
}

func (param *DateParameter) Decode(_ IConnection) (interface{}, error) {
	decoder := types.Date{}
	decoder.SetBytes(param.BValue)
	decoder.SetDataType(param.DataType)
	return decoder.Value()
}

func (param *DateParameter) Read(session network.SessionReader) error {
	var err error
	param.BValue, err = param.BasicRead(session)
	return err
}
