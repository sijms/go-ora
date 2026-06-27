package parameter_coder

import (
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type RawParameter struct {
	BasicParameter
}

func (param *RawParameter) Copy() OracleParameterCoder {
	ret := new(RawParameter)
	*ret = *param
	return ret
}

func (param *RawParameter) Encode(input interface{}, conn IConnection) error {
	param.SetDefault()
	encoder := types.Raw{}
	encoder.SetDataType(param.DataType)
	err := encoder.SetValue(input)
	if err != nil {
		return err
	}
	if dt := encoder.GetDataType(); dt != 0 {
		param.DataType = dt
	}
	param.BValue = encoder.Bytes()
	if len(param.BValue) > 0 {
		param.MaxLen = int64(len(param.BValue))
	}
	if param.MaxLen <= conn.GetMaxStringLength() {
		param.DataType = types.RAW
	} else {
		param.DataType = types.LongRaw
	}
	return nil
}

func (param *RawParameter) Decode(_ IConnection) (interface{}, error) {
	decoder := types.Raw{}
	decoder.SetBytes(param.BValue)
	decoder.SetDataType(param.DataType)
	return decoder.Value()
}

func (param *RawParameter) Write(session network.SessionWriter) error {
	session.PutClr(param.BValue)
	return nil
}

func (param *RawParameter) Read(session network.SessionReader) error {
	var err error
	param.BValue, err = param.BasicRead(session)
	return err
}
