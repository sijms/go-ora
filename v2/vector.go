package go_ora

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/sijms/go-ora/v2/converters"
)

type Vector struct {
	version int
	format  int
	flag    int
	Count   int
	Data    interface{}
	lob     Lob
}

var vectorTypeError = errors.New("unexpected data for vector type")

// NewVector : create vector from supported array type: uint8, float32 and float64
func NewVector(array interface{}) (*Vector, error) {
	v := new(Vector)
	v.flag = 2 | 16
	switch value := array.(type) {
	case []uint8:
		v.format = 4
		v.Count = len(value)
		v.Data = value
	case *[]uint8:
		v.format = 4
		v.Count = len(*value)
		v.Data = *value
	case []*uint8:
		v.format = 4
		temp := make([]byte, len(value))
		for _, val := range value {
			temp = append(temp, *val)
		}
		v.Count = len(temp)
		v.Data = temp
	case []float32:
		v.format = 2
		v.Count = len(value)
		v.Data = value
	case *[]float32:
		v.format = 2
		v.Count = len(*value)
		v.Data = *value
	case []*float32:
		v.format = 2
		temp := make([]float32, len(value))
		for _, val := range value {
			temp = append(temp, *val)
		}
		v.Count = len(temp)
		v.Data = temp
	case []float64:
		v.format = 3
		v.Count = len(value)
		v.Data = value
	case *[]float64:
		v.format = 3
		v.Count = len(*value)
		v.Data = *value
	case []*float64:
		v.format = 3
		temp := make([]float64, len(value))
		for _, val := range value {
			temp = append(temp, *val)
		}
		v.Count = len(temp)
		v.Data = temp
	default:
		return nil, vectorTypeError
	}
	return v, nil
}

func (v *Vector) decode(value []byte) error {
	if len(value) == 0 {
		v.setNil()
		return nil
	}
	magicNumber, index, err := v.read(value, 0, 1)
	if err != nil {
		return err
	}
	if magicNumber != 219 {
		return vectorTypeError
	}
	v.version, index, err = v.read(value, index, 1)
	if err != nil {
		return err
	}
	if v.version != 0 {
		return fmt.Errorf("vector version (%d) not supported", value[index])
	}
	v.flag, index, err = v.read(value, index, 2)
	if err != nil {
		return err
	}
	v.format, index, err = v.read(value, index, 1)
	if err != nil {
		return err
	}
	if v.flag&1 > 0 {
		v.Count, index, err = v.read(value, index, 1)
	} else if v.flag&2 > 0 {
		v.Count, index, err = v.read(value, index, 4)
	} else {
		v.Count, index, err = v.read(value, index, 2)
	}
	if err != nil {
		return err
	}
	if v.flag&0x10 > 0 {
		//temp := converters.ConvertBinaryDouble(value[index : index+8])
		//fmt.Println(temp)
		//index += 8
		rem := len(value) - index
		cnt := 8
		if cnt > rem {
			cnt = rem
		}
		index = index + cnt
	}
	switch v.format {
	case 2:
		// float32 array
		elementSize := 4
		data := make([]float32, 0)
		for i := 0; i < v.Count; i++ {
			number := converters.ConvertBinaryFloat(value[index : index+elementSize])
			data = append(data, number)
			index += elementSize
		}
		v.Data = data
	case 3:
		// float64 array
		elementSize := 8
		data := make([]float64, 0)
		for i := 0; i < v.Count; i++ {
			number := converters.ConvertBinaryDouble(value[index : index+elementSize])
			data = append(data, number)
			index += elementSize
		}
		v.Data = data
	case 4:
		// int8 array
		elementSize := 1
		data := make([]uint8, 0)
		for i := 0; i < v.Count; i++ {
			data = append(data, value[index])
			index += elementSize
		}
		v.Data = data
	default:
		return fmt.Errorf("unsupported format (%d)", v.format)
	}
	return nil
}

// load will retrieve data from Lob and then call decode
func (v *Vector) load() error {
	if v.lob.connection != nil && len(v.lob.sourceLocator) > 0 {
		value, err := v.lob.getData()
		if err != nil {
			return err
		}
		return v.decode(value)
	}
	return nil
}

func (v *Vector) encode() ([]byte, error) {
	if v.flag == 0 || v.format == 0 || v.Data == nil {
		return nil, nil
	}
	buffer := new(bytes.Buffer)
	var err error
	_, err = buffer.Write([]byte{219, 0})
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.BigEndian, uint16(v.flag))
	if err != nil {
		return nil, err
	}
	err = buffer.WriteByte(byte(v.format))
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.BigEndian, uint32(v.Count))
	if err != nil {
		return nil, err
	}
	if v.flag&0x10 > 0 {
		_, err = buffer.Write(bytes.Repeat([]byte{0}, 8))
		if err != nil {
			return nil, err
		}
	}
	switch value := v.Data.(type) {
	case []uint8:
		_, err = buffer.Write(value)
		if err != nil {
			return nil, err
		}
	case []float32:
		for _, val := range value {
			_, err = buffer.Write(converters.EncodeFloat32(val))
			if err != nil {
				return nil, err
			}
		}
	case []float64:
		for _, val := range value {
			_, err = buffer.Write(converters.EncodeFloat64(val))
			if err != nil {
				return nil, err
			}
		}
	default:
		return nil, vectorTypeError
	}
	return buffer.Bytes(), nil
}

func (v *Vector) read(buffer []byte, index, length int) (result, idx int, err error) {
	result = -1
	idx = index + length
	err = nil
	if index+length > len(buffer) {
		err = vectorTypeError
		return
	}
	switch length {
	case 1:
		result = int(buffer[index])
	case 2:
		temp := binary.BigEndian.Uint16(buffer[index : index+2])
		result = int(temp)
	case 4:
		temp := binary.BigEndian.Uint32(buffer[index : index+4])
		result = int(temp)
	default:
		err = vectorTypeError
	}
	return
}

func (v *Vector) setNil() {
	v.format = 0
	v.flag = 0
	v.Count = 0
	v.Data = nil
}

func (v *Vector) Scan(input interface{}) error {
	if input == nil {
		v.setNil()
		return nil
	}
	switch value := input.(type) {
	case Vector:
		*v = value
	case *Vector:
		*v = *value
	default:
		temp, err := NewVector(value)
		if err != nil {
			return err
		}
		*v = *temp
	}
	return nil
}

func (v Vector) SetDataType(conn *Connection, par *ParameterInfo) error {
	par.DataType = VECTOR
	return nil
}
