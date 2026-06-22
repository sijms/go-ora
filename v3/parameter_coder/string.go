package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type StringParameter struct {
	BasicParameter
}

func (param *StringParameter) Encode(input interface{}, strConv converters.StringCoder, _ types.LobStreamer) error {
	param.SetDefault()
	param.ContFlag = 0x10
	param.CharsetForm = 1
	switch temp := input.(type) {
	case types.String:
		if temp.UseNCharset {
			param.CharsetForm = 2
		}
	case *types.String:
		if temp.UseNCharset {
			param.CharsetForm = 2
		}
	}
	conv, err := strConv.GetStringCoder(param.CharsetID, param.CharsetForm)
	if err != nil {
		return err
	}
	if param.CharsetID == 0 {
		param.CharsetID = conv.GetLangID()
	}
	encoder := &types.String{
		Conv: conv,
	}
	encoder.SetDataType(param.DataType)
	err = encoder.SetValue(input)
	if err != nil {
		return err
	}
	if dt := encoder.GetDataType(); dt != 0 {
		param.DataType = dt
	}
	param.BValue = encoder.Bytes()
	param.MaxLen = int64(len(param.BValue))
	param.MaxCharLen = int64(len(param.BValue))
	maxLen := strConv.GetMaxStringLength()
	if param.MaxLen > maxLen {
		param.DataType = types.LongVarChar
	} else {
		param.DataType = types.NCHAR
	}
	if param.MaxLen == 0 {
		param.MaxLen = 1
	}
	return nil
}

func (param *StringParameter) Decode(strConv converters.StringCoder) (interface{}, error) {
	conv, err := strConv.GetStringCoder(param.CharsetID, param.CharsetForm)
	if err != nil {
		return nil, err
	}
	if param.CharsetID == 0 {
		param.CharsetID = conv.GetLangID()
	}
	decoder := &types.String{
		Conv: conv,
	}
	decoder.SetBytes(param.BValue)
	decoder.SetDataType(param.DataType)
	return decoder.Value()
}

func (param *StringParameter) Write(session network.SessionWriter) error {
	session.PutClr(param.BValue)
	return nil
}

func (param *StringParameter) Read(session network.SessionReader) error {
	var err error
	param.BValue, err = param.basicRead(session)
	return err
}
