package type_coder

import (
	"errors"

	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
	"github.com/sijms/go-ora/v3/types/oson"
)

type JsonCoder struct {
	locator []byte
	LobCoder
}

var jsonTypeError = errors.New("unexpected data for json type")

func NewJsonDecoder() OracleTypeDecoder { return &JsonCoder{} }

func NewJsonEncoder(input *types.Json) (OracleTypeEncoder, error) {
	ret := new(JsonCoder)
	var err error
	ret.SetDefault()
	ret.DataType = types.JSON
	if input == nil {
		return ret, nil
	}
	if input.Data == nil {
		return ret, nil
	}
	// convert json to bytes
	ret.BValue, err = oson.Encode(input.Data)
	if err != nil {
		return nil, err
	}
	ret.locator = createQuasiLocator(uint64(len(ret.BValue)))
	return ret, nil
}

func (coder *JsonCoder) copy() *JsonCoder {
	ret := new(JsonCoder)
	*ret = *coder
	return ret
}
func (coder *JsonCoder) Write(session network.SessionWriter) error {
	if coder.locator != nil {
		session.PutUint(len(coder.locator), 4, true, true)
		session.PutClr(coder.locator)
		session.PutClr(coder.BValue)
	} else {
		session.PutClr(coder.locator)
	}
	return nil
}

func (coder *JsonCoder) DecodeJson(data []byte) (*types.Json, error) {
	var value interface{}
	var err error
	if data != nil {
		value, err = oson.Decode(data)
		if err != nil {
			return nil, err
		}
	}
	return types.NewJson(value, coder.streamer, coder.copy())
}

func (coder *JsonCoder) Decode(data []byte) (interface{}, error) {
	if coder.streamer.GetLocator() == nil {
		return nil, nil
	}
	return coder.DecodeJson(data)
}

func (coder *JsonCoder) Read(session network.SessionReader) (interface{}, error) {
	bValue, err := coder.read(session)
	if err != nil {
		return nil, err
	}
	return coder.Decode(bValue)
}
