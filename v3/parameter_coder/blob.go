package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type BlobParameter struct {
	lobParameter
}

func (param *BlobParameter) Encode(input interface{}, _ converters.StringCoder, stream types.LobStreamer) (err error) {
	param.SetDefault()
	param.DataType = types.OCIBlobLocator
	encoder := &types.Blob{}
	encoder.SetDataType(param.DataType)
	err = 	encoder.SetValue(input)
	if dt := encoder.GetDataType(); dt != 0 {
		param.DataType = dt
	}
	if err != nil {
		return
	}
	//switch temp := input.(type) {
	//case types.Blob:
	//	param.BValue = temp.GetLocator()
	//case *types.Blob:
	//	param.BValue = temp.GetLocator()
	//default:
	//
	//	//err = fmt.Errorf("blob parameter must be of type *Blob")
	//}
	if !encoder.IsDataUploaded() && len(encoder.Bytes()) > 0 {
		encoder.SetStreamer(stream)
		err = encoder.Upload()
		if err != nil {
			return
		}
	}
	param.BValue = encoder.GetLocator()
	return
}
func (param *BlobParameter) Decode(_ converters.StringCoder) (interface{}, error) {
	decoder := &types.Blob{}
	decoder.SetStreamer(param.streamer)
	decoder.SetBytes(param.BValue)
	return decoder, nil
}

func (param *BlobParameter) Read(session network.SessionReader) error {
	return param.read(session)
}
