package parameter_coder

import (
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
	"github.com/sijms/go-ora/v3/types/oson"
)

//type jsonCoder struct {
//	streamer types.LobStreamer
//}

//func (c *jsonCoder) EncodeJson(input interface{}) ([]byte, error) {
//	return oson.Encode(input)
//}
//
//func (c *jsonCoder) DecodeJson(data []byte) (*types.Json, error) {
//	var value interface{}
//	var err error
//	if data != nil {
//		value, err = oson.Decode(data)
//		if err != nil {
//			return nil, err
//		}
//	}
//	return types.NewJson(value, c.streamer, c)
//}

type JsonParameter struct {
	locator types.Locator
	lobParameter
}

func (param *JsonParameter) Copy() OracleParameterCoder {
	ret := new(JsonParameter)
	*ret = *param
	return ret
}

func (param *JsonParameter) Encode(input interface{}, _ IConnection) (err error) {
	param.SetDefault()
	param.DataType = types.JSON
	encoder := &types.Json{}
	encoder.Coder = &oson.Oson{}
	encoder.SetDataType(param.DataType)
	err = encoder.SetValue(input)
	if err != nil {
		return
	}
	if dt := encoder.GetDataType(); dt != 0 {
		param.DataType = dt
	}
	param.BValue = encoder.Bytes()
	param.BValue = encoder.GetLocator()
	return
}

func (param *JsonParameter) Decode(_ IConnection) (interface{}, error) {
	decoder := &types.Json{}
	decoder.Coder = &oson.Oson{}
	decoder.SetStreamer(param.streamer)
	decoder.SetBytes(param.BValue)
	return decoder, nil
}

func (param *JsonParameter) Write(session network.SessionWriter) error {
	if param.locator != nil {
		session.PutUint(len(param.locator), 4, true, true)
		session.PutClr(param.locator)
		session.PutClr(param.BValue)
	} else {
		session.PutClr(param.locator)
	}
	return nil
}
func (param *JsonParameter) Read(session network.SessionReader) error {
	return param.read(session)
}
