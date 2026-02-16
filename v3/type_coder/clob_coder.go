package type_coder

import (
	"database/sql"

	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type ClobCoder struct {
	LobCoder
}

func NewClobEncoder(clob types.Clob) OracleTypeEncoder {
	ret := &ClobCoder{}
	ret.BValue = clob.GetLocator()
	ret.TypeInfo.SetDefault()
	ret.DataType = types.OCIClobLocator
	ret.CharsetID, ret.CharsetForm = clob.Charset()
	return ret
	//ret := new(ClobCoder)
	//ret.DataType = types.OCIClobLocator
	//ret.CharsetForm = 1
	//if useNCharset {
	//	ret.CharsetForm = 2
	//}
	//ret.data = data
	//return ret
}
func (coder *ClobCoder) copy() *ClobCoder {
	ret := &ClobCoder{}
	*ret = *coder
	return ret
}
func NewClobDecoder() OracleTypeDecoder {
	return &ClobCoder{}
}

func (coder *ClobCoder) Encode(strConv converters.StringCoder, lobStream types.LobStreamer) error {
	//if !coder.data.Valid || len(coder.data.String) == 0 {
	//	coder.data.Locator = nil
	//	coder.bValue = nil
	//	coder.MaxLen = 1
	//	return nil
	//}
	//conv, err := strConv.GetStringCoder(coder.CharsetID, coder.CharsetForm)
	//if err != nil {
	//	return err
	//}
	//if coder.CharsetID == 0 {
	//	coder.CharsetID = conv.GetLangID()
	//}
	//data := conv.Encode(coder.data.String)
	//coder.bValue, err = lobStream.CreateTemporaryLocator(coder.CharsetID, coder.CharsetForm)
	//if err != nil {
	//	return err
	//}
	//coder.data.Locator = coder.bValue
	//err = lobStream.Write(data)
	//if err != nil {
	//	return err
	//}
	return nil
}

//func (coder *ClobCoder) DecodeString(data []byte) (types.Clob, error) {
//	ret := types.Clob{}
//	if coder.Coder == nil {
//		return ret, errors.New("string coder should be set first before encoding/decoding strings")
//	}
//	conv, err := coder.Coder.GetStringCoder(coder.CharsetID, coder.CharsetForm)
//	if err != nil {
//		return ret, err
//	}
//	if data != nil {
//		ret.Valid = true
//		ret.String = conv.Decode(data)
//	}
//	return ret, nil
//}

func (coder *ClobCoder) DecodeClob(data []byte) (types.Clob, error) {
	//ret := types.NewClob(coder.streamer, coder.CharsetID, coder.CharsetForm, data)
	if len(data) == 0 {
		return types.NewClob(coder.streamer, coder.copy(), sql.NullString{Valid: false, String: ""}, false)
	}
	var charsetConv converters.IStringConverter
	var err error
	if coder.streamer.IsVarWidthChar() {
		if coder.streamer.DatabaseVersionNumber() < 10200 && coder.streamer.IsLittleEndian() {
			charsetConv, err = coder.streamer.GetStringCoder().GetStringCoder(2002, 0)
		} else {
			charsetConv, err = coder.streamer.GetStringCoder().GetStringCoder(2000, 0)
		}
	} else {
		charsetConv, err = coder.streamer.GetStringCoder().GetStringCoder(coder.CharsetID, coder.CharsetForm)
	}
	//charsetConv, err = coder.streamer.GetStringCoder().GetStringCoder(coder.CharsetID, coder.CharsetForm)
	if err != nil {
		return nil, err
	}
	return types.NewClob(coder.streamer, coder.copy(), sql.NullString{Valid: true, String: charsetConv.Decode(data)}, false)
}
func (coder *ClobCoder) Decode(data []byte) (interface{}, error) {
	if coder.streamer.GetLocator() == nil {
		return nil, nil
	}
	return coder.DecodeClob(data)

	//clob, err := coder.DecodeString(data)
	//if err != nil {
	//	return nil, err
	//}
	//clob.Locator = coder.locator
	//return &clob, nil
}

func (coder *ClobCoder) Read(session network.SessionReader) (interface{}, error) {
	bValue, err := coder.read(session)
	if err != nil {
		return nil, err
	}
	return coder.Decode(bValue)
}
