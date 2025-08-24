package json

import (
	"bytes"
	"fmt"
)

type Field interface {
	Encode() ([]byte, error)
	Decode(input []byte) error
	OpCode() uint8
	KeyIndex() int
	SetKeyIndex(int)
	Offset() int
	SetOffset(int)
	Children() []Field
}

type basicField struct {
	opCode   uint8
	keyIndex int
	offset   int
	children []Field
}

func (field *basicField) OpCode() uint8 {
	return field.opCode
}
func (field *basicField) KeyIndex() int {
	return field.keyIndex
}
func (field *basicField) Children() []Field {
	return field.children
}
func (field *basicField) SetKeyIndex(index int) {
	field.keyIndex = index
}
func (field *basicField) Offset() int {
	return field.offset
}
func (field *basicField) SetOffset(offset int) {
	field.offset = offset
}

func (field *basicField) GetTotalLength() (int, error) {
	childBuffer := bytes.NewBuffer(nil)
	for _, child := range field.children {
		data, err := child.Encode()
		if err != nil {
			return 0, err
		}
		_, err = childBuffer.Write(data)
		if err != nil {
			return 0, err
		}
	}
	return childBuffer.Len(), nil
}

func (field *basicField) modifyFlag(basicFlag uint8) (uint8, error) {
	childLen := len(field.children)
	if childLen >= 0x100 {
		if childLen < 0x10000 {
			basicFlag |= 8
		} else {
			basicFlag |= 0x10
		}
	}
	totalLength, err := field.GetTotalLength()
	if err != nil {
		return 0, err
	}
	if (2 + len(field.children) + (2 * len(field.children)) + totalLength) > 0xFFFF {
		basicFlag |= 0x20
	}
	return basicFlag, nil
}

//type basicKey struct {
//	keyName   string
//	keyHash   uint8
//	keySize   int
//	keyOffset int
//}
//
//func (key *basicKey) KeyName() string {
//	return key.keyName
//}
//func (key *basicKey) KeyHash() uint8 {
//	return key.keyHash
//}
//func (key *basicKey) KeySize() int {
//	return key.keySize
//}
//func (key *basicKey) KeyOffset() int {
//	return key.keyOffset
//}

type NullField struct {
	basicField
}

func (field *NullField) Encode() ([]byte, error) {
	field.opCode = 48
	return []byte{field.opCode}, nil
}
func (field *NullField) Decode(input []byte) error {
	if input[0] != 48 {
		return fmt.Errorf("invalid null field with opCode (%d)", input[0])

	}
	return nil
}

type BooleanField struct {
	Value bool
	basicField
}

func (field *BooleanField) Encode() ([]byte, error) {
	if field.Value {
		field.opCode = 49
	} else {
		field.opCode = 50
	}
	return []byte{field.opCode}, nil
}

func (field *BooleanField) Decode(input []byte) error {
	field.opCode = input[0]
	switch field.opCode {
	case 49:
		field.Value = true
	case 50:
		field.Value = false
	default:
		return fmt.Errorf("invalid opcode(%d) for BooleanField", field.opCode)
	}
	return nil
}
