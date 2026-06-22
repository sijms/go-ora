package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type NumberParameter struct {
	BasicParameter
}

func (param *NumberParameter) Encode(input interface{}, _ converters.StringCoder, _ types.LobStreamer) error {
	param.SetDefault()
	param.DataType = types.NUMBER
	param.MaxLen = 0x16
	encoder := types.Number{}
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

func (param *NumberParameter) Decode(_ converters.StringCoder) (interface{}, error) {
	decoder := types.Number{}
	decoder.SetBytes(param.BValue)
	decoder.SetDataType(param.DataType)
	return decoder.Value()
}

func (param *NumberParameter) Write(session network.SessionWriter) error {
	session.PutClr(param.BValue)
	return nil
}

func (param *NumberParameter) Read(session network.SessionReader) error {
	var err error
	param.BValue, err = param.basicRead(session)
	return err
}
