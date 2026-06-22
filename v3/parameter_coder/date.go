package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type DateParameter struct {
	BasicParameter
}

func (param *DateParameter) Encode(input interface{}, _ converters.StringCoder, _ types.LobStreamer) error {
	param.SetDefault()
	param.DataType = types.DATE
	encoder := types.Date{}
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
	return nil
}

func (param *DateParameter) Decode(_ converters.StringCoder) (interface{}, error) {
	decoder := types.Date{}
	decoder.SetBytes(param.BValue)
	decoder.SetDataType(param.DataType)
	return decoder.Value()
}

func (param *DateParameter) Write(session network.SessionWriter) error {
	session.PutClr(param.BValue)
	return nil
}

func (param *DateParameter) Read(session network.SessionReader) error {
	var err error
	param.BValue, err = param.basicRead(session)
	return err
}
