package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type IntervalParameter struct {
	BasicParameter
}

func (param *IntervalParameter) Encode(input interface{}, _ converters.StringCoder, _ types.LobStreamer) error {
	param.SetDefault()
	encoder := &types.Interval{}
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

func (param *IntervalParameter) Decode(_ converters.StringCoder) (interface{}, error) {
	decoder := &types.Interval{}
	decoder.SetBytes(param.BValue)
	return decoder.Value()
}
func (param *IntervalParameter) Write(session network.SessionWriter) error {
	session.PutClr(param.BValue)
	return nil
}

func (param *IntervalParameter) Read(session network.SessionReader) error {
	var err error
	param.BValue, err = param.basicRead(session)
	return err
}
