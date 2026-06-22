package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type BFileParameter struct {
	lobParameter
}

func (param *BFileParameter) Encode(input interface{}, strConv converters.StringCoder, stream types.LobStreamer) (err error) {
	param.SetDefault()
	param.DataType = types.OCIFileLocator
	encoder := &types.BFile{}
	encoder.Conv, err = strConv.GetDefaultStringCoder()
	if err != nil {
		return
	}
	encoder.SetDataType(param.DataType)
	encoder.SetStreamer(stream)
	err = encoder.SetValue(input)
	if dt := encoder.GetDataType(); dt != 0 {
		param.DataType = dt
	}
	if err != nil {
		return
	}
	param.BValue = encoder.GetLocator()
	return
}

func (param *BFileParameter) Decode(strConv converters.StringCoder) (output interface{}, err error) {
	decoder := &types.BFile{}
	decoder.Conv, err = strConv.GetDefaultStringCoder()
	if err != nil {
		return
	}
	decoder.SetStreamer(param.streamer)
	decoder.SetBytes(param.BValue)
	return decoder, nil
}

func (param *BFileParameter) Read(session network.SessionReader) error {
	return param.read(session)
}
