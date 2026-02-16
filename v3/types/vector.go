package types

import (
	"context"
	"errors"
	"fmt"
)

type VectorDataType int
type VectorType int

const (
	VECTOR_NIL VectorDataType = iota
	VECTOR_UINT8
	VECTOR_FL32
	VECTOR_FL64
)
const (
	VECTOR_SPARSE VectorType = iota
	VECTOR_DENSE
)

type Vector interface {
	VectorType() VectorDataType
	Data() interface{}
	Length() int
	Lob
}
type VectorDecoder interface {
	DecodeVector(data []byte) (Vector, error)
}
type vector struct {
	//version int
	//format  int
	//flag int
	//Count int
	Type    VectorDataType
	data    interface{}
	decoder VectorDecoder
	lobBase
}

var vectorTypeError = errors.New("unexpected data for vector type")

// CreateVector : create vector from supported array type: uint8, float32 and float64
func CreateVector(array interface{}) (Vector, error) {
	v := new(vector)
	if array == nil {
		v.Type = VECTOR_NIL
		v.data = nil
		return v, nil
	}
	switch value := array.(type) {
	case []uint8:
		v.Type = VECTOR_UINT8
		v.data = value
	case *[]uint8:
		v.Type = VECTOR_UINT8
		v.data = *value
	case []*uint8:
		temp := make([]byte, len(value))
		for _, val := range value {
			temp = append(temp, *val)
		}
		v.Type = VECTOR_UINT8
		v.data = temp
	case []float32:
		v.Type = VECTOR_FL32
		v.data = value
	case *[]float32:
		v.Type = VECTOR_FL32
		v.data = *value
	case []*float32:
		temp := make([]float32, len(value))
		for _, val := range value {
			temp = append(temp, *val)
		}
		v.Type = VECTOR_FL32
		v.data = temp
	case []float64:
		v.Type = VECTOR_FL64
		v.data = value
	case *[]float64:
		v.Type = VECTOR_FL64
		v.data = *value
	case []*float64:
		temp := make([]float64, len(value))
		for _, val := range value {
			temp = append(temp, *val)
		}
		v.Type = VECTOR_FL64
		v.data = temp
	default:
		return nil, vectorTypeError
	}
	return v, nil
}

//	func NewVectorFromBytes(data []byte) (Vector, error) {
//		read := func(tempBuffer *bytes.Buffer, size int) (int, error) {
//			var err error
//			switch size {
//			case 1:
//				var temp uint8
//				temp, err = tempBuffer.ReadByte()
//				if err != nil {
//					return 0, vectorTypeError
//				}
//				return int(temp), nil
//			case 2:
//				var temp uint16
//				err = binary.Read(tempBuffer, binary.BigEndian, &temp)
//				if err != nil {
//					return 0, vectorTypeError
//				}
//				return int(temp), nil
//			case 4:
//				var temp uint32
//				err = binary.Read(tempBuffer, binary.BigEndian, &temp)
//				if err != nil {
//					return 0, vectorTypeError
//				}
//				return int(temp), nil
//			default:
//				return 0, vectorTypeError
//			}
//		}
//		buffer := bytes.NewBuffer(data)
//		magicNumber, err := read(buffer, 1)
//		if err != nil {
//			return nil, err
//		}
//		if magicNumber != 219 {
//			return nil, vectorTypeError
//		}
//		var version, flag, format, count int
//		version, err = read(buffer, 1)
//		if err != nil {
//			return nil, err
//		}
//		if version != 0 {
//			return nil, fmt.Errorf("vector version (%d) not supported", version)
//		}
//		flag, err = read(buffer, 2)
//		if err != nil {
//			return nil, err
//		}
//		format, err = read(buffer, 1)
//		if err != nil {
//			return nil, err
//		}
//		if flag&1 > 0 {
//			count, err = read(buffer, 1)
//		} else if flag&2 > 0 {
//			count, err = read(buffer, 4)
//		} else {
//			count, err = read(buffer, 2)
//		}
//		if err != nil {
//			return nil, err
//		}
//		if flag&0x10 > 0 {
//			_ = buffer.Next(8)
//		}
//		switch format {
//		case 2:
//			// float32 array
//			vecData := make([]float32, count)
//			for i := 0; i < count; i++ {
//				temp := &DoubleCoder{}
//				vecData[i], err = temp.DecodeFloat(buffer.Next(4))
//				if err != nil {
//					return nil, err
//				}
//			}
//			return NewVector(vecData, nil)
//		case 3:
//		case 4:
//
//		}
//	}
func NewVector(array interface{}, stream LobStreamer, decoder VectorDecoder) (Vector, error) {
	v, err := CreateVector(array)
	if err != nil {
		return nil, err
	}
	v.(*vector).stream = stream
	v.(*vector).decoder = decoder
	return v, nil
}
func (v *vector) Length() int {
	switch v.Type {
	case VECTOR_NIL:
		return 0
	case VECTOR_UINT8:
		if temp, ok := v.data.([]uint8); ok {
			return len(temp)
		}
	case VECTOR_FL32:
		if temp, ok := v.data.([]float32); ok {
			return len(temp)
		}
	case VECTOR_FL64:
		if temp, ok := v.data.([]float64); ok {
			return len(temp)
		}
	}
	return 0
}

func (v *vector) VectorType() VectorDataType {
	return v.Type
}

func (v *vector) Data() interface{} {
	return v.data
}

func (v *vector) Read(ctx context.Context) error {
	if v.Type != VECTOR_NIL || v.data != nil {
		return nil
	}
	if v.IsQuasi() {
		return nil
	}
	if v.decoder == nil {
		return fmt.Errorf("no decoder defined for vector type")
	}
	temp, err := v.lobBase.ReadFromPos(ctx, 0)
	if err != nil {
		return err
	}
	if v.decoder == nil {
		return fmt.Errorf("no decoder defined for vector type")
	}
	tempVector, err := v.decoder.DecodeVector(temp)
	if err != nil {
		return err
	}
	return v.Scan(tempVector)
}

//func (v *vector) CopyFrom(source Vector) error {
//	if temp, ok := source.(*vector); ok {
//		err := v.lobBase.CopyFrom(&temp.lobBase)
//		if err != nil {
//			return err
//		}
//		//v.flag = temp.flag
//		v.Type = temp.Type
//		v.data = temp.data
//	}
//	return nil
//}

// load will retrieve data from Lob and then call decode
//func (vector *Vector) load(stream LobStreamer) error {
//	if len(vector.locator) > 0 {
//		bValue, err := stream.Read(context.Background(), 0, 0)
//		if err != nil {
//			return err
//		}
//		_, err = vector.Decode(bValue, VECTOR)
//		return err
//	}
//	return nil
//}

func (v *vector) Scan(input interface{}) error {
	var err error
	if input == nil {
		if v.stream != nil {
			v.stream.SetLocator(nil)
		}
		v.data = nil
		v.Type = VECTOR_NIL
		return nil
	}
	switch value := input.(type) {
	case *vector:
		err = v.lobBase.copyFrom(&value.lobBase)
		if err != nil {
			return nil
		}
		v.Type = value.Type
		v.data = value.data
	default:
		temp, err := CreateVector(value)
		if err != nil {
			return err
		}
		err = v.Scan(temp)
		if err != nil {
			return err
		}
	}
	return nil
}

//func (vector *Vector) Value() (driver.Value, error) {
//	return vector.Data, nil
//}
