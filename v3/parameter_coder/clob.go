package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type ClobParameter struct {
	//UseNClob  bool
	lobParameter
}

func (param *ClobParameter) Encode(input interface{}, strConv converters.StringCoder, stream types.LobStreamer) (err error) {
	param.SetDefault()
	param.DataType = types.OCIClobLocator
	param.CharsetForm = 1
	switch input := input.(type) {
	case types.String:
		if input.UseNCharset {
			param.CharsetForm = 2
		}
	case *types.String:
		if input.UseNCharset {
			param.CharsetForm = 2
		}
	case types.Clob:
		if input.UseNCharset {
			param.CharsetForm = 2
		}
	case *types.Clob:
		if input.UseNCharset {
			param.CharsetForm = 2
		}
	}
	encoder := &types.Clob{}
	encoder.Conv, err = strConv.GetStringCoder(param.CharsetID, param.CharsetForm)
	if err != nil {
		return
	}
	param.CharsetID = encoder.Conv.GetLangID()
	err = encoder.SetValue(input, 0)
	if err != nil {
		return
	}
	if !encoder.IsDataUploaded() && len(encoder.Bytes()) > 0 {
		encoder.SetStreamer(stream)
		err = encoder.Upload()
		if err != nil {
			return
		}
	}
	param.BValue = encoder.GetLocator()
	return
	//param.CharsetForm = 1
	//switch temp := input.(type) {
	//case types.Clob:
	//	if temp.UseNCharset {
	//		param.CharsetForm = 2
	//	}
	//case *types.Clob:
	//	if temp.UseNCharset {
	//		param.CharsetForm = 2
	//	}
	//default:
	//	err = fmt.Errorf("clob parameter must be of type *Clob")
	//}
	////encoder := &types.Clob{}
	//encoder.Conv, err = strConv.GetStringCoder(param.CharsetID, param.CharsetForm)
	//if err != nil {
	//	return err
	//}
	//if param.CharsetID == 0 {
	//	param.CharsetID = encoder.Conv.GetLangID()
	//}
	//err = encoder.SetValue(input, param.DataType)
	//if err != nil {
	//	return err
	//}
	//if err != nil {
	//	return err
	//}
	//param.BValue = encoder.GetLocator()
	//param.streamer = encoder.GetStreamer()
}

func (param *ClobParameter) Decode(strConv converters.StringCoder) (interface{}, error) {
	decoder := &types.Clob{}
	decoder.SetStreamer(param.streamer)
	decoder.SetBytes(param.BValue)
	locator := param.streamer.GetLocator()
	var err error
	if locator.IsVarWidthChar() {
		if param.streamer.DatabaseVersionNumber() < 10200 && locator.IsLittleEndian() {
			decoder.Conv, err = strConv.GetStringCoder(2002, 0)
		} else {
			decoder.Conv, err = strConv.GetStringCoder(2000, 0)
		}
	} else {
		decoder.Conv, err = strConv.GetStringCoder(param.CharsetID, param.CharsetForm)
	}
	return decoder, err
}

func (param *ClobParameter) Read(session network.SessionReader) error {
	return param.read(session)
}
