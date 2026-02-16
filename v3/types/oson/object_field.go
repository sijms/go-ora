package oson

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/sijms/go-ora/v3/types"
)

type ObjectField struct {
	value map[string]interface{}
	structField
}

func (obj *ObjectField) Value() interface{} {
	//TODO implement me
	return obj.value
}

func NewObjectField(value map[string]interface{}, header *Header) (*ObjectField, error) {
	ret := new(ObjectField)
	ret.header = header
	var err error
	for keyName, value := range value {
		var field Field
		if value == nil {
			field = &NullField{}
		} else {
			rValue := reflect.ValueOf(value)
			//rType := reflect.TypeOf(value)
			switch rValue.Kind() {
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				fallthrough
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				fallthrough
			case reflect.Float32, reflect.Float64:
				temp, err := types.NewNumber(value)
				if err != nil {
					return nil, err
				}
				field = &NumberField{value: *temp}
			case reflect.String:
				field = NewStringField(rValue.String())
			case reflect.Bool:
				field = &BooleanField{value: rValue.Bool()}
			case reflect.Slice, reflect.Array:
				field, err = NewArrayField(value, header)
				if err != nil {
					return nil, err
				}
			case reflect.Map:
				if temp, ok := value.(map[string]interface{}); ok {
					field, err = NewObjectField(temp, header)
					if err != nil {
						return nil, err
					}
				} else {
					return nil, fmt.Errorf("invalid JSON object at key: %s not decoded as Map[string]Any", keyName)
				}
			default:
				return nil, fmt.Errorf("unsupported type: %s at key: %s", rValue.Type(), keyName)
			}
		}
		if field == nil {
			return nil, fmt.Errorf("no value for key %s", keyName)
		}
		field.SetKeyIndex(header.keys.Index(keyName) + 1)
		ret.children = append(ret.children, field)
	}
	return ret, nil
}

func (obj *ObjectField) Encode() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	childBuffer := bytes.NewBuffer(nil)
	var err error
	childLen := len(obj.children)
	var flag uint8 = 0x84
	offset := obj.offset
	flag, err = obj.modifyFlag(flag)
	if err != nil {
		return nil, err
	}
	// get key information
	keys := make([]int, childLen)
	for index, child := range obj.children {
		keys[index] = child.KeyIndex()
	}
	// create object header
	objHeader := newObjectHeader(flag, keys, offset)
	var selectedHeader *objectHeader
	offset, selectedHeader, err = objHeader.write(buffer, obj.header)
	if childLen == 0 {
		return buffer.Bytes(), nil
	}
	// encode and calculate offsets
	if flag&0x20 > 0 {
		offset += childLen * 4
	} else {
		offset += childLen * 2
	}
	if selectedHeader != nil {
		for _, index := range selectedHeader.keysIndex {
			for i, child := range obj.children {
				if index == child.KeyIndex() {
					// calculate offset
					obj.children[i].SetOffset(offset + childBuffer.Len())
					child.SetOffset(offset + childBuffer.Len())
					if flag&0x20 > 0 {
						err = binary.Write(buffer, binary.BigEndian, uint32(offset+childBuffer.Len()))
					} else {
						err = binary.Write(buffer, binary.BigEndian, uint16(offset+childBuffer.Len()))
					}
					data, err := child.Encode()
					if err != nil {
						return nil, err
					}
					_, err = childBuffer.Write(data)
					if err != nil {
						return nil, err
					}
					break
				}
			}
		}
	} else {
		for i, child := range obj.children {
			// calculate offset
			obj.children[i].SetOffset(offset + childBuffer.Len())
			if flag&0x20 > 0 {
				err = binary.Write(buffer, binary.BigEndian, uint32(offset+childBuffer.Len()))
			} else {
				err = binary.Write(buffer, binary.BigEndian, uint16(offset+childBuffer.Len()))
			}
			data, err := child.Encode()
			if err != nil {
				return nil, err
			}
			_, err = childBuffer.Write(data)
			if err != nil {
				return nil, err
			}
		}
	}

	_, err = childBuffer.WriteTo(buffer)
	return buffer.Bytes(), err
}
