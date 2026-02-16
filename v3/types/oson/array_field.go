package oson

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/sijms/go-ora/v3/types"
)

type ArrayField struct {
	value []interface{}
	structField
}

func (array *ArrayField) Value() interface{} {
	return array.value
}

func NewArrayField(value interface{}, header *Header) (*ArrayField, error) {
	ret := new(ArrayField)
	ret.header = header
	var err error
	rValue := reflect.ValueOf(value)
	length := rValue.Len()
	for i := 0; i < length; i++ {
		var field Field
		item := rValue.Index(i)
		if item.IsNil() || !item.IsValid() {
			field = &NullField{}
		} else {
			item = reflect.ValueOf(item.Interface())
			switch item.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fallthrough
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				fallthrough
			case reflect.Float32, reflect.Float64:
				temp, err := types.NewNumber(item.Interface())
				if err != nil {
					return nil, err
				}
				field = &NumberField{value: *temp}
			case reflect.String:
				field = NewStringField(item.String())
			case reflect.Bool:
				field = &BooleanField{value: item.Bool()}
			case reflect.Slice, reflect.Array:
				field, err = NewArrayField(item.Interface(), header)
				if err != nil {
					return nil, err
				}
			case reflect.Map:
				if temp, ok := item.Interface().(map[string]interface{}); ok {
					field, err = NewObjectField(temp, header)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf("invalid JSON Object at index: %d", i)
				}
			default:
				return nil, fmt.Errorf("unsupported type: %s at index: %d", rValue.Type(), i)
			}
		}
		if field == nil {
			return nil, fmt.Errorf("no value for index %d", i)
		}
		ret.children = append(ret.children, field)
	}

	return ret, nil
}

func (array *ArrayField) Encode() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	childBuffer := bytes.NewBuffer(nil)
	var err error
	childLen := len(array.children)
	var flag uint8 = 0xC0
	offset := array.offset
	flag, err = array.modifyFlag(flag)
	if err != nil {
		return nil, err
	}
	err = buffer.WriteByte(flag)
	if err != nil {
		return nil, err
	}
	offset += 1
	// write child count
	if childLen < 0x100 {
		err = buffer.WriteByte(uint8(childLen))
		offset += 1
	} else if childLen < 0x10000 {
		err = binary.Write(buffer, binary.BigEndian, uint16(childLen))
		offset += 2
	} else {
		err = binary.Write(buffer, binary.BigEndian, uint32(childLen))
		offset += 4
	}
	if err != nil {
		return nil, err
	}
	if childLen == 0 {
		return buffer.Bytes(), nil
	}
	if flag&0x20 > 0 {
		offset += childLen * 4
	} else {
		offset += childLen * 2
	}
	for i, _ := range array.children {
		array.children[i].SetOffset(offset + childBuffer.Len())
		if flag&0x20 > 0 {
			err = binary.Write(buffer, binary.BigEndian, uint32(offset+childBuffer.Len()))
		} else {
			err = binary.Write(buffer, binary.BigEndian, uint16(offset+childBuffer.Len()))
		}
		data, err := array.children[i].Encode()
		if err != nil {
			return nil, err
		}
		_, err = childBuffer.Write(data)
		if err != nil {
			return nil, err
		}
	}
	_, err = childBuffer.WriteTo(buffer)
	return buffer.Bytes(), err
}
