package parameter_coder

import (
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
	"github.com/sijms/go-ora/v3/utils"
)

type StringParameter struct {
	BasicParameter
}

func (param *StringParameter) Copy() OracleParameterCoder {
	ret := new(StringParameter)
	*ret = *param
	return ret
}

func (param *StringParameter) Init() {
	param.SetDefault()
	param.DataType = types.NCHAR
	param.ContFlag = 0x10
	if param.CharsetForm == 0 {
		param.CharsetForm = 1
	}
}

func (param *StringParameter) Encode(input interface{}, conn IConnection) error {
	param.Init()
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
	conv, err := conn.GetStringCoder(param.CharsetID, param.CharsetForm)
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
	length := int64(len(param.BValue))
	length = utils.Max(length, param.MaxLen, param.MaxCharLen)
	param.MaxLen = length
	param.MaxCharLen = length
	length = conn.GetMaxStringLength()
	if param.MaxLen > length {
		param.DataType = types.LongVarChar
	} else {
		param.DataType = types.NCHAR
	}
	if param.MaxLen < encoder.GetMaxLen() {
		param.MaxLen = encoder.GetMaxLen()
	}
	return nil
}

func (param *StringParameter) Decode(conn IConnection) (interface{}, error) {
	conv, err := conn.GetStringCoder(param.CharsetID, param.CharsetForm)
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

func (param *StringParameter) Read(session network.SessionReader) error {
	var err error
	param.BValue, err = param.BasicRead(session)
	return err
}
