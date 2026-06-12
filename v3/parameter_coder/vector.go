package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type VectorParameter struct {
	locator types.Locator
	lobParameter
}

func (param *VectorParameter) Encode(input interface{}, _ converters.StringCoder, _ types.LobStreamer) error {
	param.SetDefault()
	param.DataType = types.VECTOR
	encoder := &types.Vector{}
	err := encoder.SetValue(input, param.DataType)
	if err != nil {
		return err
	}
	param.BValue = encoder.Bytes()
	param.locator = encoder.GetLocator()
	// locator
	return nil
}

func (param *VectorParameter) Decode(_ converters.StringCoder) (interface{}, error) {
	decoder, err := types.NewVector(nil, param.streamer)
	if err != nil {
		return nil, err
	}
	//decoder := &types.Vector{}
	decoder.SetBytes(param.BValue)
	return decoder, nil
}

func (param *VectorParameter) Write(session network.SessionWriter) error {
	if param.locator != nil {
		session.PutUint(len(param.locator), 4, true, true)
		session.PutClr(param.locator)
		session.PutClr(param.BValue)
	} else {
		session.PutClr(param.locator)
	}
	return nil
}

func (param *VectorParameter) Read(session network.SessionReader) error {
	return param.read(session)
}
