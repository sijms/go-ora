package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type BoolParameter struct {
	BasicParameter
}

func (param *BoolParameter) Encode(input interface{}, _ converters.StringCoder, _ types.LobStreamer) error {
	param.SetDefault()
	param.DataType = types.BOOLEAN
	encoder := &types.Bool{}
	err := encoder.SetValue(input, param.DataType)
	if err != nil {
		return err
	}
	param.BValue = encoder.Bytes()
	return nil
}

func (param *BoolParameter) Decode(_ converters.StringCoder) (interface{}, error) {
	decoder := &types.Bool{}
	decoder.SetBytes(param.BValue)
	return decoder.Value(param.DataType)
}

func (param *BoolParameter) Write(session network.SessionWriter) error {
	session.PutClr(param.BValue)
	return nil
}

func (param *BoolParameter) Read(session network.SessionReader) error {
	var err error
	param.BValue, err = param.basicRead(session)
	return err
}
