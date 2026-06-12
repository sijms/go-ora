package oson

import (
	"bytes"
	"fmt"

	"github.com/sijms/go-ora/v3/types"
)

type DateField struct {
	value types.Date
	basicField
}

func (field *DateField) Value() (interface{}, error) {
	return field.value.Value(0)
}

func (field *DateField) Encode() ([]byte, error) {
	var err error
	data := field.value.Bytes()
	length := len(data)
	buffer := bytes.NewBuffer(nil)
	// here encode field
	var written int
	written, err = buffer.Write(data)
	if written != length {
		return nil, fmt.Errorf("invalid buffer write: expected to write %d, got %d", length, written)
	}
	return buffer.Bytes(), err
}

func (field *DateField) Decode(input []byte) error {
	buffer := bytes.NewBuffer(input)
	var err error
	field.opCode, err = buffer.ReadByte()
	if err != nil {
		return err
	}
	var length int
	switch field.opCode {
	case 60:
	case 0x7D:
		length = 7
	case 57:
		length = 11
	case 0x7C:
		length = 13
	default:
		return fmt.Errorf("invalid opCode (%d) for DateField", field.opCode)
	}
	data := make([]byte, length)
	var read int
	read, err = buffer.Read(data)
	if read != length {
		return fmt.Errorf("invalid buffer read: expected to read %d, got %d", length, read)
	}
	if err != nil {
		return err
	}
	field.value.SetBytes(data)
	return nil
}
