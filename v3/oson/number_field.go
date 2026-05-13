package oson

import (
	"bytes"
	"fmt"
	"math/big"
	"strings"
)

type NumberField struct {
	//value types.Number
	data    []byte
	decoder NumberDecoder
	basicField
}

func (field *NumberField) Value() interface{} {
	strNum, err := field.decoder.DecodeNumber(field.data)
	if err != nil {
		return "E"
	}
	if strings.Contains(strNum, ".") {
		ret, success := big.NewFloat(0).SetString(strNum)
		if !success {
			return "E"
		}
		return ret
	}

	ret, success := big.NewInt(0).SetString(strNum, 10)
	if !success {
		return "E"
	}
	return ret
}
func (field *NumberField) Encode() ([]byte, error) {
	length := len(field.data)
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
	written, err = buffer.Write(field.data)
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
	field.data = make([]byte, length)
	var read int
	read, err = buffer.Read(field.data)
	if read != length {
		return fmt.Errorf("invalid buffer read: expected to read %d, got %d", length, read)
	}
	return nil
}

type BinaryDoubleField struct {
	data    []byte
	decoder BinaryDoubleDecoder
	basicField
}

func (field *BinaryDoubleField) Value() interface{} {
	ret, _ := field.decoder.DecodeBinaryDouble(field.data)
	return ret
}
func (field *BinaryDoubleField) Encode() ([]byte, error) {
	output := []uint8{0x36}
	output = append(output, field.data...)
	return output, nil
}

type BinaryFloatField struct {
	data    []byte
	decoder BinaryFloatDecoder
	basicField
}

func (field *BinaryFloatField) Value() interface{} {
	ret, _ := field.decoder.DecodeBinaryFloat(field.data)
	return ret
}
func (field *BinaryFloatField) Encode() ([]byte, error) {
	output := []uint8{0x7F}
	output = append(output, field.data...)
	return output, nil
}
