package parameter_coder

import (
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type VectorParameter struct {
	locator types.Locator
	lobParameter
}

func (param *VectorParameter) Copy() OracleParameterCoder {
	ret := new(VectorParameter)
	*ret = *param
	return ret
}

func (param *VectorParameter) Encode(input interface{}, _ IConnection) error {
	param.SetDefault()
	param.DataType = types.VECTOR
	encoder := &types.Vector{}
	encoder.SetDataType(param.DataType)
	err := encoder.SetValue(input)
	if err != nil {
		return err
	}
	if dt := encoder.GetDataType(); dt != 0 {
		param.DataType = dt
	}
	param.BValue = encoder.Bytes()
	param.locator = encoder.GetLocator()
	// locator
	return nil
}

func (param *VectorParameter) Decode(_ IConnection) (interface{}, error) {
	decoder := &types.Vector{}
	decoder.SetStreamer(param.streamer)
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
