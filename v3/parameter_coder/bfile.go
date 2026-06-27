package parameter_coder

import (
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type BFileParameter struct {
	lobParameter
}

func (param *BFileParameter) Copy() OracleParameterCoder {
	ret := new(BFileParameter)
	*ret = *param
	return ret
}
func (param *BFileParameter) Encode(input interface{}, conn IConnection) (err error) {
	param.SetDefault()
	param.DataType = types.OCIFileLocator
	encoder := &types.BFile{}
	encoder.Conv, err = conn.GetDefaultStringCoder()
	if err != nil {
		return
	}
	encoder.SetDataType(param.DataType)
	encoder.SetStreamer(conn.NewLobStreamer())
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

func (param *BFileParameter) Decode(conn IConnection) (output interface{}, err error) {
	decoder := &types.BFile{}
	decoder.Conv, err = conn.GetDefaultStringCoder()
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
