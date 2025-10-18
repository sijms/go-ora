package oson

import (
	"bytes"
	"fmt"

	"github.com/sijms/go-ora/v2/types"
)

type NumberField struct {
	Value types.Number
	basicField
}

func (field *NumberField) Encode() ([]byte, error) {
	data := field.Value.Data()
	length := len(data)
	buffer := bytes.NewBuffer(nil)
	var err error
	if length <= 8 {
		field.opCode = uint8(0x60 | (length - 1))
		err = buffer.WriteByte(field.opCode)
		if err != nil {
			return nil, err
		}
	} else if length < 0x100 {
		field.opCode = 0x74
		err = buffer.WriteByte(field.opCode)
		if err != nil {
			return nil, err
		}
		err = buffer.WriteByte(uint8(length))
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("unsupported number length (%d) >= 256", length)
	}
	var written int
	written, err = buffer.Write(data)
	if written != length {
		return nil, fmt.Errorf("invalid buffer write: expected to write %d, got %d", length, written)
	}
	return buffer.Bytes(), err
}
func (field *NumberField) Decode(input []byte) error {
	buffer := bytes.NewBuffer(input)
	var err error
	field.opCode, err = buffer.ReadByte()
	if err != nil {
		return err
	}
	var length int
	if field.opCode == 0x74 {
		var temp uint8
		temp, err = buffer.ReadByte()
		if err != nil {
			return err
		}
		length = int(temp)
	}
	if field.opCode&0x60 > 0 {
		length = int(field.opCode) & ^0x60
		if length > 8 {
			return fmt.Errorf("invalid number length (%d) for NumberField", length)
		}
	} else {
		return fmt.Errorf("invalid opCode(%d) for NumberField", field.opCode)
	}
	data := make([]byte, length)
	var read int
	read, err = buffer.Read(data)
	if read != length {
		return fmt.Errorf("invalid buffer read: expected to read %d, got %d", length, read)
	}
	var temp *types.Number
	temp, err = types.NewNumber(data)
	if err != nil {
		return err
	}
	field.Value = *temp
	return nil
}
