package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type RawParameter struct {
	BasicParameter
}

func (param *RawParameter) Encode(input interface{}, strConv converters.StringCoder, _ types.LobStreamer) error {
	param.SetDefault()
	encoder := types.Raw{}
	err := encoder.SetValue(input, param.DataType)
	if err != nil {
		return err
	}
	param.BValue = encoder.Bytes()
	if len(param.BValue) > 0 {
		param.MaxLen = int64(len(param.BValue))
	}
	if param.MaxLen <= strConv.GetMaxStringLength() {
		param.DataType = types.RAW
	} else {
		param.DataType = types.LongRaw
	}
	return nil
}

func (param *RawParameter) Decode(_ converters.StringCoder) (interface{}, error) {
	decoder := types.Raw{}
	decoder.SetBytes(param.BValue)
	return decoder.Value(param.DataType)
}

func (param *RawParameter) Write(session network.SessionWriter) error {
	session.PutClr(param.BValue)
	return nil
}

func (param *RawParameter) Read(session network.SessionReader) error {
	var err error
	param.BValue, err = param.basicRead(session)
	return err
}
