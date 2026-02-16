package type_coder

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type VectorCoder struct {
	locator []byte
	LobCoder
}

var vectorTypeError = errors.New("unexpected data for vector type")

func NewVectorDecoder() OracleTypeDecoder { return &VectorCoder{} }
func NewVectorEncoder(vector types.Vector) (OracleTypeEncoder, error) {
	ret := new(VectorCoder)
	ret.SetDefault()
	ret.DataType = types.VECTOR
	if vector.VectorType() == types.VECTOR_NIL {
		return ret, nil
	}
	format := 0
	switch vector.VectorType() {
	case types.VECTOR_UINT8:
		format = 4
	case types.VECTOR_FL32:
		format = 2
	case types.VECTOR_FL64:
		format = 3
	default:
		return nil, vectorTypeError
	}
	count := vector.Length()
	flag := 2 | 16
	buffer := new(bytes.Buffer)
	var err error
	_, err = buffer.Write([]byte{219, 0})
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.BigEndian, uint16(flag))
	if err != nil {
		return nil, err
	}
	err = buffer.WriteByte(byte(format))
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.BigEndian, uint32(count))
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(bytes.Repeat([]byte{0}, 8))
	if err != nil {
		return nil, err
	}
	switch value := vector.Data().(type) {
	case []uint8:
		_, err = buffer.Write(value)
		if err != nil {
			return nil, err
		}
	case []float32:
		for _, val := range value {
			temp := NewDoubleCoderFromFloat(val)
			_, err = buffer.Write(temp.BValue)
			if err != nil {
				return nil, err
			}
		}
	case []float64:
		for _, val := range value {
			temp := NewDoubleCoderFromDouble(val)
			_, err = buffer.Write(temp.BValue)
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, vectorTypeError
	}
	ret.BValue = buffer.Bytes()
	ret.locator = createQuasiLocator(uint64(len(ret.BValue)))
	return ret, nil
}

func (coder *VectorCoder) copy() *VectorCoder {
	ret := new(VectorCoder)
	*ret = *coder
	return ret
}

func (coder *VectorCoder) Write(session network.SessionWriter) error {
	if coder.locator != nil {
		session.PutUint(len(coder.locator), 4, true, true)
		session.PutClr(coder.locator)
		session.PutClr(coder.BValue)
	} else {
		session.PutClr(coder.locator)
	}
	return nil
}

func (coder *VectorCoder) DecodeVector(data []byte) (types.Vector, error) {
	if data == nil {
		return types.NewVector(nil, coder.streamer, coder.copy())
	}
	read := func(tempBuffer *bytes.Buffer, size int) (int, error) {
		var err error
		switch size {
		case 1:
			var temp uint8
			temp, err = tempBuffer.ReadByte()
			if err != nil {
				return 0, vectorTypeError
			}
			return int(temp), nil
		case 2:
			var temp uint16
			err = binary.Read(tempBuffer, binary.BigEndian, &temp)
			if err != nil {
				return 0, vectorTypeError
			}
			return int(temp), nil
		case 4:
			var temp uint32
			err = binary.Read(tempBuffer, binary.BigEndian, &temp)
			if err != nil {
				return 0, vectorTypeError
			}
			return int(temp), nil
		default:
			return 0, vectorTypeError
		}
	}
	buffer := bytes.NewBuffer(data)
	magicNumber, err := read(buffer, 1)
	if err != nil {
		return nil, err
	}
	if magicNumber != 219 {
		return nil, vectorTypeError
	}
	var version, flag, format, count int
	version, err = read(buffer, 1)
	if err != nil {
		return nil, err
	}
	if version != 0 {
		return nil, fmt.Errorf("vector version (%d) not supported", version)
	}
	flag, err = read(buffer, 2)
	if err != nil {
		return nil, err
	}
	format, err = read(buffer, 1)
	if err != nil {
		return nil, err
	}
	if flag&1 > 0 {
		count, err = read(buffer, 1)
	} else if flag&2 > 0 {
		count, err = read(buffer, 4)
	} else {
		count, err = read(buffer, 2)
	}
	if err != nil {
		return nil, err
	}
	if flag&0x10 > 0 {
		_ = buffer.Next(8)
	}
	switch format {
	case 2:
		// float32 array
		vecData := make([]float32, count)
		for i := 0; i < count; i++ {
			temp := &DoubleCoder{}
			vecData[i], err = temp.DecodeFloat(buffer.Next(4))
			if err != nil {
				return nil, err
			}
		}
		return types.NewVector(vecData, coder.streamer, coder.copy())
	case 3:
		// float64 array
		vecData := make([]float64, count)
		for i := 0; i < count; i++ {
			temp := &DoubleCoder{}
			vecData[i], err = temp.DecodeDouble(buffer.Next(8))
			if err != nil {
				return nil, err
			}
		}
		return types.NewVector(vecData, coder.streamer, coder.copy())
	case 4:
		// int8 array
		vecData := make([]uint8, count)
		for i := 0; i < count; i++ {
			vecData[i], err = buffer.ReadByte()
		}
		return types.NewVector(vecData, coder.streamer, coder.copy())
	default:
		return nil, fmt.Errorf("unsupported format (%d)", format)
	}
}

func (coder *VectorCoder) Decode(data []byte) (interface{}, error) {
	if coder.streamer.GetLocator() == nil {
		return nil, nil
	}
	return coder.DecodeVector(data)
}

func (coder *VectorCoder) Read(session network.SessionReader) (interface{}, error) {
	bValue, err := coder.read(session)
	if err != nil {
		return nil, err
	}
	return coder.Decode(bValue)
	//if bValue != nil {
	//	return coder.Decode(bValue)
	//}
	// at this point we have locator and data will be retrieved with Lob operation ( in case
	// check if the lob is temporary to be released later
	//return vector.Decode()
}
