package oson

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type StringField struct {
	value  string
	bValue []byte
	basicField
}

func (field *StringField) Value() interface{} {
	return field.value
}

func NewStringField(value string) *StringField {
	return &StringField{
		value: value,
	}
}
func (field *StringField) Encode() ([]byte, error) {
	buffer := bytes.NewBuffer(nil)
	var err error
	length := len(field.value)
	if length < 0x1F {
		field.opCode = uint8(length)
		err = buffer.WriteByte(field.opCode)
		if err != nil {
			return nil, err
		}
	} else if length < 0x100 {
		field.opCode = 51
		err = buffer.WriteByte(field.opCode)
		if err != nil {
			return nil, err
		}
		err = buffer.WriteByte(uint8(length))
	} else if length < 0x10000 {
		field.opCode = 55
		err = buffer.WriteByte(field.opCode)
		if err != nil {
			return nil, err
		}
		err = binary.Write(buffer, binary.BigEndian, uint16(length))
	} else {
		field.opCode = 56
		err = buffer.WriteByte(field.opCode)
		if err != nil {
			return nil, err
		}
		err = binary.Write(buffer, binary.BigEndian, uint32(length))
	}
	if err != nil {
		return nil, err
	}
	_, err = buffer.WriteString(field.value)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func (field *StringField) Decode(input []byte) error {
	buffer := bytes.NewBuffer(input)
	var err error
	field.opCode, err = buffer.ReadByte()
	if err != nil {
		return err
	}
	var data []byte
	if field.opCode < 0x1F {
		data = make([]byte, field.opCode)
	} else {
		switch field.opCode {
		case 51:
			var length uint8
			length, err = buffer.ReadByte()
			if err != nil {
				return err
			}
			data = make([]byte, length)
		case 55:
			var length uint16
			err = binary.Read(buffer, binary.BigEndian, &length)
			if err != nil {
				return err
			}
			data = make([]byte, length)
		case 56:
			var length uint32
			err = binary.Read(buffer, binary.BigEndian, &length)
			if err != nil {
				return err
			}
			data = make([]byte, length)
		default:
			return fmt.Errorf("invalid opcode (%d) for string field", field.opCode)
		}
	}

	var read int
	read, err = buffer.Read(data)
	if read != len(data) {
		return errors.New("invalid data length")
	}
	field.value = string(data)
	return nil
}
