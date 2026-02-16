package oson

import (
	"bytes"
	"encoding/binary"
	"slices"
)

type objectHeader struct {
	flag             uint8
	keysIndex        []int
	orderedKeysIndex []int
	offset           int
}

func newObjectHeader(flag uint8, keyIndex []int, offset int) *objectHeader {
	// get a copy of key index and then sort and store this copy
	ret := &objectHeader{
		flag:      flag,
		keysIndex: keyIndex,
		offset:    offset,
	}
	ret.orderedKeysIndex = make([]int, len(keyIndex))
	copy(ret.orderedKeysIndex, keyIndex)
	slices.Sort(ret.orderedKeysIndex)
	return ret
}
func (objHeader *objectHeader) isEqual(input objectHeader) bool {
	return slices.Equal(objHeader.orderedKeysIndex, input.orderedKeysIndex)
}

func (objHeader *objectHeader) write(buffer *bytes.Buffer, header *Header) (offset int, selectedHeader *objectHeader, err error) {
	offset = objHeader.offset
	selectedHeader = nil
	// search in stored object header for similarity
	for _, tempHeader := range header.objectHeaders {
		// if find similar just write the offset of similar object
		if objHeader.isEqual(tempHeader) {
			selectedHeader = &tempHeader
			objHeader.flag |= 0x18
			err = buffer.WriteByte(objHeader.flag)
			if err != nil {
				return
			}
			offset += 1
			if objHeader.flag&0x20 > 0 { // use wide offset
				err = binary.Write(buffer, binary.BigEndian, uint32(tempHeader.offset))
				offset += 4
			} else {
				err = binary.Write(buffer, binary.BigEndian, uint16(tempHeader.offset))
				offset += 2
			}
			return
		}
	}
	// if not find write object header and save it to object header list
	err = buffer.WriteByte(objHeader.flag)
	if err != nil {
		return
	}
	offset += 1
	childLen := len(objHeader.keysIndex)
	// write child count
	switch objHeader.flag & 0x18 {
	case 0:
		err = buffer.WriteByte(uint8(childLen))
		offset += 1
	case 8:
		err = binary.Write(buffer, binary.BigEndian, uint16(childLen))
		offset += 2
	case 0x10:
		err = binary.Write(buffer, binary.BigEndian, uint32(childLen))
		offset += 4
	}
	if err != nil {
		return
	}
	if childLen == 0 {
		return
	}
	// write key index
	for _, keyIndex := range objHeader.keysIndex {
		if header.flags&0x8 > 0 {
			err = binary.Write(buffer, binary.BigEndian, uint32(keyIndex))
			offset += 4
		} else if header.flags&0x400 > 0 {
			err = binary.Write(buffer, binary.BigEndian, uint16(keyIndex))
			offset += 2
		} else {
			err = buffer.WriteByte(uint8(keyIndex))
			offset += 1
		}
		if err != nil {
			return
		}
	}
	header.objectHeaders = append(header.objectHeaders, *objHeader)
	return
}
