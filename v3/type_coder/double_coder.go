package type_coder

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type DoubleCoder struct {
	TypeInfo
}

func NewDoubleCoderFromDouble(data float64) *DoubleCoder {
	ret := &DoubleCoder{}
	ret.BValue = make([]byte, 8)
	temp := math.Float64bits(data)
	binary.BigEndian.PutUint64(ret.BValue, temp)
	if data > 0 {
		ret.BValue[0] = ret.BValue[0] | 128
	} else {
		xorBuffer(ret.BValue, 8)
	}
	ret.DataType = types.IBDOUBLE
	ret.MaxLen = 8
	return ret
}

func NewDoubleCoderFromFloat(data float32) *DoubleCoder {
	ret := &DoubleCoder{}
	ret.BValue = make([]byte, 4)
	temp := math.Float32bits(data)
	binary.BigEndian.PutUint32(ret.BValue, temp)
	if data > 0 {
		ret.BValue[0] = ret.BValue[0] | 128
	} else {
		xorBuffer(ret.BValue, 4)
	}
	ret.DataType = types.IBFLOAT
	ret.MaxLen = 4
	return ret
}

func (coder *DoubleCoder) DecodeFloat(data []byte) (float32, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("error decoding binary float, supplied buffer length: %d and required length: %d", len(data), 4)
	}
	if data[0]&128 != 0 {
		data[0] = data[0] & 127
	} else {
		xorBuffer(data, 4)
	}
	return math.Float32frombits(binary.BigEndian.Uint32(data)), nil
}

func (coder *DoubleCoder) DecodeDouble(data []byte) (float64, error) {
	if len(data) < 8 {
		return 0, fmt.Errorf("error decoding binary double, supplied buffer length: %d and required length: %d", len(data), 8)
	}
	if data[0]&128 != 0 {
		data[0] = data[0] & 127
	} else {
		xorBuffer(data, 8)
	}
	return math.Float64frombits(binary.BigEndian.Uint64(data)), nil
}

func (coder *DoubleCoder) Decode(data []byte) (interface{}, error) {
	if coder.DataType == types.BFLOAT {
		return coder.DecodeFloat(data)
	}
	return coder.DecodeDouble(data)
}

func (coder *DoubleCoder) Read(session network.SessionReader) (interface{}, error) {
	bValue, err := coder.basicRead(session)
	if err != nil {
		return 0, err
	}
	return coder.Decode(bValue)
}

func (coder *DoubleCoder) Write(session network.SessionWriter) error {
	session.PutClr(coder.BValue)
	return nil
}
