package json

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/sijms/go-ora/v2/types"
	"reflect"
)

type ObjectField struct {
	value map[string]interface{}
	basicField
}

func NewObjectField(value map[string]interface{}, keys KeyCollection) (*ObjectField, error) {
	ret := new(ObjectField)
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
				field = &NumberField{Value: *temp}
			case reflect.String:
				field = NewStringField(rValue.String())
			case reflect.Bool:
				field = &BooleanField{Value: rValue.Bool()}
			case reflect.Slice, reflect.Array:
				field, err = NewArrayField(value, keys)
				if err != nil {
					return nil, err
				}
			case reflect.Map:
				if temp, ok := value.(map[string]interface{}); ok {
					field, err = NewObjectField(temp, keys)
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
		field.SetKeyIndex(keys.Index(keyName) + 1)
		ret.children = append(ret.children, field)
	}
	return ret, nil
}

func (obj *ObjectField) Encode() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	childBuffer := bytes.NewBuffer(nil)
	var err error
	// get total value length
	var totalLength int
	totalLength, err = obj.GetTotalLength()
	if err != nil {
		return nil, err
	}
	childLen := len(obj.children)
	useWideOffset := (2 + childLen + (2 * childLen) + totalLength) > 0xFFFF
	var flag uint8 = 0x84
	offset := obj.offset
	flag, err = obj.modifyFlag(flag)
	if err != nil {
		return nil, err
	}

	// write flag
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
	// write key index
	for _, child := range obj.children {
		err = buffer.WriteByte(uint8(child.KeyIndex()))
		if err != nil {
			return nil, err
		}
	}
	offset += childLen
	// encode and calculate offsets
	if useWideOffset {
		offset += childLen * 4
	} else {
		offset += childLen * 2
	}
	for i, child := range obj.children {
		// calculate offset

		obj.children[i].SetOffset(offset + childBuffer.Len())
		if useWideOffset {
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
	_, err = childBuffer.WriteTo(buffer)
	return buffer.Bytes(), err
}

func (obj *ObjectField) Decode(input []byte) error {
	return nil
}
