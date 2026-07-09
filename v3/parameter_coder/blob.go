package parameter_coder

import (
	"github.com/sijms/go-ora/v3/types"
)

type BlobParameter struct {
	lobParameter
}

func (param *BlobParameter) Copy() OracleParameterCoder {
	ret := new(BlobParameter)
	*ret = *param
	return ret
}
func (param *BlobParameter) Init() {
	param.SetDefault()
	param.DataType = types.OCIBlobLocator
}
func (param *BlobParameter) Encode(input interface{}, conn IConnection) (err error) {
	param.Init()
	encoder := &types.Blob{}
	encoder.SetDataType(param.DataType)
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
	if !encoder.IsDataUploaded() && len(encoder.Bytes()) > 0 {
		encoder.SetStreamer(conn.NewLobStreamer())
		err = encoder.Upload()
		if err != nil {
			return
		}
	}
	param.BValue = encoder.GetLocator()
	return
}
func (param *BlobParameter) Decode(_ IConnection) (interface{}, error) {
	decoder := &types.Blob{}
	decoder.SetStreamer(param.streamer)
	decoder.SetBytes(param.BValue)
	return decoder, nil
}
