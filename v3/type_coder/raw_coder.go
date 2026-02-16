package type_coder

import (
	"fmt"

	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type RawCoder struct {
	TypeInfo
}

func NewRawCoder(data interface{}) (*RawCoder, error) {
	ret := new(RawCoder)
	ret.SetDefault()
	switch v := data.(type) {
	case []byte:
		ret.BValue = v
		ret.MaxLen = int64(len(v))
	case *[]byte:
		ret.BValue = *v
		ret.MaxLen = int64(len(*v))
	default:
		return nil, fmt.Errorf("raw coder: invalid data type: %T", v)
	}
	if ret.MaxLen == 0 {
		ret.MaxLen = 1
	}
	return ret, nil
}

func (coder *RawCoder) Encode(strConv converters.StringCoder, _ types.LobStreamer) error {
	if coder.MaxLen <= strConv.GetMaxStringLength() {
		coder.DataType = types.RAW
	} else {
		coder.DataType = types.LongRaw
	}
	return nil
}

func (coder *RawCoder) Decode(data []byte) (interface{}, error) {
	return data, nil
}

func (coder *RawCoder) Read(session network.SessionReader) (interface{}, error) {
	bValue, err := coder.basicRead(session)
	if err != nil {
		return nil, err
	}
	return coder.Decode(bValue)
}

func (coder *RawCoder) Write(session network.SessionWriter) error {
	session.PutClr(coder.BValue)
	return nil
}
