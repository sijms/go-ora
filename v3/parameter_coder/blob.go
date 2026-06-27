package parameter_coder

import (
	"github.com/sijms/go-ora/v3/network"
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
func (param *BlobParameter) Encode(input interface{}, conn IConnection) (err error) {
	param.SetDefault()
	param.DataType = types.OCIBlobLocator
	encoder := &types.Blob{}
	encoder.SetDataType(param.DataType)
	err = encoder.SetValue(input)
	if dt := encoder.GetDataType(); dt != 0 {
		param.DataType = dt
	}
	if err != nil {
		return
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

func (param *BlobParameter) Read(session network.SessionReader) error {
	return param.read(session)
}
