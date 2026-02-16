package oson

import "bytes"

type structField struct {
	header *Header
	basicField
}

func (field *structField) getTotalLength() (int, error) {
	objHeaders := field.header.objectHeaders
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
	field.header.objectHeaders = objHeaders
	return childBuffer.Len(), nil
}

func (field *structField) modifyFlag(basicFlag uint8) (uint8, error) {
	childLen := len(field.children)
	if childLen >= 0x100 {
		if childLen < 0x10000 {
			basicFlag |= 8
		} else {
			basicFlag |= 0x10
		}
	}
	totalLength, err := field.getTotalLength()
	if err != nil {
		return 0, err
	}
	if (field.offset + 2 + len(field.children) + (2 * len(field.children)) + totalLength) > 0xFFFF {
		basicFlag |= 0x20
	}
	return basicFlag, nil
}
