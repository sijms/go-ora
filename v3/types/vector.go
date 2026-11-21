package types

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type VectorDataType int
type VectorType int

// const (
//
//	VECTOR_NIL VectorDataType = iota
//	VECTOR_UINT8
//	VECTOR_FL32
//	VECTOR_FL64
//
// )
const (
	VECTOR_SPARSE VectorType = iota
	VECTOR_DENSE
)

type Vector struct {
	Basic
	loc Locator
	lobBase
}

func (vector *Vector) GetLocator() Locator {
	if vector.lobBase.GetLocator() != nil {
		return vector.lobBase.GetLocator()
	}
	return vector.loc
}
func (vector *Vector) Upload() error {
	return vector.uploadData(vector.bValue, 0, 0)
}

func (vector *Vector) SetValue(input interface{}) error {
	if input == nil {
		vector.bValue = nil
		return nil
	}
	var (
		header        = []byte{219, 0}
		flag   uint16 = 2 | 16
		format byte
		length uint32
		err    error
		buffer = &bytes.Buffer{}
		data   interface{}
	)
	switch value := input.(type) {
	case Vector:
		*vector = value
		return nil
	case *Vector:
		*vector = *value
		return nil
	case []uint8:
		format = 4
		length = uint32(len(value))
		data = value
		//v.Type = VECTOR_UINT8
		//v.data = value
	case *[]uint8:
		format = 4
		length = uint32(len(*value))
		data = *value
		//v.Type = VECTOR_UINT8
		//v.data = *value
	case []*uint8:
		temp := make([]byte, 0, len(value))
		for _, val := range value {
			temp = append(temp, *val)
		}
		format = 4
		length = uint32(len(value))
		data = temp
		//v.Type = VECTOR_UINT8
		//v.data = temp
	case []int8:
		// an INT8 vector, as Value returns it
		temp := make([]byte, len(value))
		for i, val := range value {
			temp[i] = byte(val)
		}
		format = vectorFormatInt8
		length = uint32(len(value))
		data = temp
	case []float32:
		format = 2
		length = uint32(len(value))
		data = value
		//v.Type = VECTOR_FL32
		//v.data = value
	case *[]float32:
		format = 2
		length = uint32(len(*value))
		data = *value
		//v.Type = VECTOR_FL32
		//v.data = *value
	case []*float32:
		temp := make([]float32, 0, len(value))
		for _, val := range value {
			temp = append(temp, *val)
		}
		format = 2
		length = uint32(len(value))
		data = temp
		//v.Type = VECTOR_FL32
		//v.data = temp
	case []float64:
		format = 3
		length = uint32(len(value))
		data = value
		//v.Type = VECTOR_FL64
		//v.data = value
	case *[]float64:
		format = 3
		length = uint32(len(*value))
		data = *value
		//v.Type = VECTOR_FL64
		//v.data = *value
	case []*float64:
		temp := make([]float64, 0, len(value))
		for _, val := range value {
			temp = append(temp, *val)
		}
		format = 3
		length = uint32(len(value))
		data = temp
		//v.Type = VECTOR_FL64
		//v.data = temp
	default:
		return vectorTypeError
	}

	_, err = buffer.Write(header)
	if err != nil {
		return err
	}
	err = binary.Write(buffer, binary.BigEndian, flag)
	if err != nil {
		return err
	}
	err = buffer.WriteByte(format)
	if err != nil {
		return err
	}
	err = binary.Write(buffer, binary.BigEndian, length)
	if err != nil {
		return err
	}
	_, err = buffer.Write(bytes.Repeat([]byte{0}, 8))
	if err != nil {
		return err
	}
	switch value := data.(type) {
	case []uint8:
		_, err = buffer.Write(value)
		if err != nil {
			return err
		}
	case []float32:
		for _, val := range value {
			n := Number{}
			n.SetDataType(IBFLOAT)
			err = n.SetValue(val)
			if err != nil {
				return err
			}
			//temp := NewDoubleCoderFromFloat(val)
			_, err = buffer.Write(n.Bytes())
			if err != nil {
				return err
			}
		}
	case []float64:
		for _, val := range value {
			n := Number{}
			n.SetDataType(IBDOUBLE)
			err = n.SetValue(val)
			if err != nil {
				return err
			}
			_, err = buffer.Write(n.Bytes())
			if err != nil {
				return err
			}
		}
	default:
		return vectorTypeError
	}
	vector.bValue = buffer.Bytes()
	dataLen := uint64(len(vector.bValue))
	if dataLen > 0 {
		vector.loc = NewQuasiLocator(dataLen)
	} else {
		vector.loc = nil
	}
	return nil
}

// The vector image: a magic byte and a version, flags, the element format, the
// number of dimensions, 8 bytes of a norm where a flag says so, and the elements;
// a sparse vector's elements are its count of values that are not zero, their
// indexes, then the values.
const (
	vectorMagic         = 0xDB
	vectorVersionSparse = 2
	vectorFlagNorm      = 0x02
	vectorFlagNormSpace = 0x10
	vectorFlagSparse    = 0x20
	vectorFormatFloat32 = 2
	vectorFormatFloat64 = 3
	vectorFormatInt8    = 4
	vectorFormatBinary  = 5
)

// SparseVector is a sparse VECTOR: its number of dimensions, and the indexes,
// from 0, and values of the dimensions that are not zero. Values is a []float32,
// a []float64, or a []int8.
type SparseVector struct {
	Dimensions int
	Indices    []uint32
	Values     interface{}
}

// Value returns a dense vector's elements as a []float32, a []float64, a []int8,
// or, for a BINARY vector, its bits packed in a []uint8; and a sparse vector as a
// SparseVector.
func (vector *Vector) Value() (interface{}, error) {
	if len(vector.bValue) == 0 {
		return nil, nil
	}
	buffer := bytes.NewReader(vector.bValue)
	var header struct {
		Magic   uint8
		Version uint8
		Flags   uint16
		Format  uint8
		Count   uint32
	}
	if err := binary.Read(buffer, binary.BigEndian, &header); err != nil {
		return nil, vectorTypeError
	}
	if header.Magic != vectorMagic {
		return nil, vectorTypeError
	}
	if header.Version > vectorVersionSparse {
		return nil, fmt.Errorf("vector version (%d) not supported", header.Version)
	}
	if header.Flags&(vectorFlagNorm|vectorFlagNormSpace) != 0 {
		if _, err := buffer.Seek(8, io.SeekCurrent); err != nil {
			return nil, vectorTypeError
		}
	}
	if header.Flags&vectorFlagSparse != 0 {
		var count uint16
		if err := binary.Read(buffer, binary.BigEndian, &count); err != nil {
			return nil, vectorTypeError
		}
		indices := make([]uint32, count)
		if err := binary.Read(buffer, binary.BigEndian, indices); err != nil {
			return nil, vectorTypeError
		}
		values, err := vectorElements(buffer, header.Format, int(count))
		if err != nil {
			return nil, err
		}
		return SparseVector{Dimensions: int(header.Count), Indices: indices, Values: values}, nil
	}
	count := int(header.Count)
	if header.Format == vectorFormatBinary {
		count /= 8
	}
	return vectorElements(buffer, header.Format, count)
}

func vectorElements(buffer *bytes.Reader, format uint8, count int) (interface{}, error) {
	size := map[uint8]int{vectorFormatFloat32: 4, vectorFormatFloat64: 8, vectorFormatInt8: 1, vectorFormatBinary: 1}[format]
	if size == 0 {
		return nil, fmt.Errorf("unsupported format (%d) for vector type", format)
	}
	data := make([]byte, count*size)
	if _, err := io.ReadFull(buffer, data); err != nil {
		return nil, vectorTypeError
	}
	switch format {
	case vectorFormatFloat32:
		values := make([]float32, count)
		for i := range values {
			n := Number{}
			n.SetBytes(data[i*4 : i*4+4])
			n.SetDataType(IBFLOAT)
			value, err := n.Value()
			if err != nil {
				return nil, err
			}
			values[i] = value.(float32)
		}
		return values, nil
	case vectorFormatFloat64:
		values := make([]float64, count)
		for i := range values {
			n := Number{}
			n.SetBytes(data[i*8 : i*8+8])
			n.SetDataType(IBDOUBLE)
			value, err := n.Value()
			if err != nil {
				return nil, err
			}
			values[i] = value.(float64)
		}
		return values, nil
	case vectorFormatInt8:
		values := make([]int8, count)
		for i, b := range data {
			values[i] = int8(b)
		}
		return values, nil
	default:
		return data, nil
	}
}

//	type Vector interface {
//		VectorType() VectorDataType
//		Data() interface{}
//		Length() int
//		Lob
//	}
//type VectorDecoder interface {
//	DecodeVector(data []byte) (Vector, error)
//}
//type vector struct {
//	//version int
//	//format  int
//	//flag int
//	//Count int
//	Type    VectorDataType
//	data    interface{}
//	decoder VectorDecoder
//	lobBase
//}

var vectorTypeError = errors.New("unexpected data for vector type")

// CreateVector : create vector from supported array type: uint8, float32 and float64
func CreateVector(array interface{}) (*Vector, error) {
	v := new(Vector)
	return v, v.SetValue(array)
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

//func (v *vector) Length() int {
//	switch v.Type {
//	case VECTOR_NIL:
//		return 0
//	case VECTOR_UINT8:
//		if temp, ok := v.data.([]uint8); ok {
//			return len(temp)
//		}
//	case VECTOR_FL32:
//		if temp, ok := v.data.([]float32); ok {
//			return len(temp)
//		}
//	case VECTOR_FL64:
//		if temp, ok := v.data.([]float64); ok {
//			return len(temp)
//		}
//	}
//	return 0
//}

//func (v *vector) VectorType() VectorDataType {
//	return v.Type
//}

//func (v *vector) Data() interface{} {
//	return v.data
//}

// Truncated reports whether the vector holds less than the database's, which
// Read fetches.
func (vector *Vector) Truncated() bool {
	return vector.length > int64(len(vector.bValue))
}

func (vector *Vector) Read(ctx context.Context) error {
	var err error
	vector.bValue, err = vector.ReadFromPos(ctx, 0)
	return err
}

//func (v *vector) Read(ctx context.Context) error {
//	if v.Type != VECTOR_NIL || v.data != nil {
//		return nil
//	}
//	if v.IsQuasi() {
//		return nil
//	}
//	if v.decoder == nil {
//		return fmt.Errorf("no decoder defined for vector type")
//	}
//	temp, err := v.lobBase.ReadFromPos(ctx, 0)
//	if err != nil {
//		return err
//	}
//	if v.decoder == nil {
//		return fmt.Errorf("no decoder defined for vector type")
//	}
//	tempVector, err := v.decoder.DecodeVector(temp)
//	if err != nil {
//		return err
//	}
//	return v.Scan(tempVector)
//}
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

func (vector *Vector) Scan(input interface{}) error {
	return vector.SetValue(input)
	//var err error
	//if input == nil {
	//	if v.stream != nil {
	//		v.stream.SetLocator(nil)
	//	}
	//	v.data = nil
	//	v.Type = VECTOR_NIL
	//	return nil
	//}
	//switch value := input.(type) {
	//case *vector:
	//	err = v.lobBase.copyFrom(&value.lobBase)
	//	if err != nil {
	//		return nil
	//	}
	//	v.Type = value.Type
	//	v.data = value.data
	//	v.decoder = value.decoder
	//default:
	//	temp, err := CreateVector(value)
	//	if err != nil {
	//		return err
	//	}
	//	err = v.Scan(temp)
	//	if err != nil {
	//		return err
	//	}
	//}
	//return nil
}
func (vector *Vector) CopyTo(dest driver.Value) error {
	val, err := vector.Value()
	//switch value := val.(type) {
	//case nil:
	//case []uint8:
	//case []float32:
	//case []float64:
	//}
	if err != nil {
		return err
	}
	switch dst := dest.(type) {
	case *Vector:
		*dst = *vector
		return nil
	case *[]uint8:
		if val == nil {
			*dst = nil
			return nil
		}
		if v, ok := val.([]uint8); ok {
			*dst = v
			return nil
		}
	case *[]float32:
		if val == nil {
			*dst = nil
			return nil
		}
		if v, ok := val.([]float32); ok {
			*dst = v
			return nil
		}
	case *[]float64:
		if val == nil {
			*dst = nil
			return nil
		}
		if v, ok := val.([]float64); ok {
			*dst = v
			return nil
		}
	case *[]int8:
		if val == nil {
			*dst = nil
			return nil
		}
		if v, ok := val.([]int8); ok {
			*dst = v
			return nil
		}
	case *SparseVector:
		if val == nil {
			*dst = SparseVector{}
			return nil
		}
		if v, ok := val.(SparseVector); ok {
			*dst = v
			return nil
		}
	}
	return fmt.Errorf("cannot copy Vector to variable of type %T", dest)
}
