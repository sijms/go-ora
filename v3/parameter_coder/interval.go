package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/types"
)

type IntervalParameter struct {
	BasicParameter
}

func (param *IntervalParameter) Encode(input interface{}, _ converters.StringCoder, _ types.LobStreamer) error {
	param.SetDefault()
	encoder := &types.Interval{}
	err := encoder.SetValue(input, param.DataType)
	if err != nil {
		return err
	}
	param.BValue = encoder.Bytes()
	return nil
}
