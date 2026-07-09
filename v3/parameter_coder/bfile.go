package parameter_coder

import (
	"github.com/sijms/go-ora/v3/types"
)

type BFileParameter struct {
	lobParameter
}

func (param *BFileParameter) Init() {
	param.SetDefault()
	param.DataType = types.OCIFileLocator
}
func (param *BFileParameter) Copy() OracleParameterCoder {
	ret := new(BFileParameter)
	*ret = *param
	return ret
}
func (param *BFileParameter) Encode(input interface{}, conn IConnection) (err error) {
	param.Init()
	encoder := &types.BFile{}
	encoder.Conv, err = conn.GetDefaultStringCoder()
	if err != nil {
		return
	}
	encoder.SetDataType(param.DataType)
	encoder.SetStreamer(conn.NewLobStreamer())
	err = encoder.SetValue(input)
	if err != nil {
		return
	}
	if dt := encoder.GetDataType(); dt != 0 {
		param.DataType = dt
	}
	if param.MaxLen < encoder.GetMaxLen() {
		param.MaxLen = encoder.GetMaxLen()
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
