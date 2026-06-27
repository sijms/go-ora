package oson

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type header struct {
	version              int
	flags                uint16
	flags2               uint16
	keys                 keyCollection
	keyDataLen           int
	nodeDataLen          int
	tinyNodeCount        uint16
	nodeOffset           int64
	currentObjBaseOffset []int64
	objectHeaders        []objectHeader
}

func (h *header) write(buffer *bytes.Buffer) error {
	var err error
	_, err = buffer.Write([]byte{0xff, 0x4a, 0x5a, 0x1})
	if err != nil {
		return err
	}
	keyLen := len(h.keys)
	if keyLen > 0xFFFF {
		h.flags |= 8
	} else if keyLen > 0xFF {
		h.flags |= 0x400
	}
	if h.keyDataLen > 0xFFFF {
		h.flags |= 0x800
	}
	if h.nodeDataLen > 0xFFFF {
		h.flags |= 0x1000
	}
	err = binary.Write(buffer, binary.BigEndian, h.flags)
	if err != nil {
		return err
	}
	err = h.writeKeyCount(buffer)
	if h.flags&0x800 != 0 {
		err = binary.Write(buffer, binary.BigEndian, uint32(h.keyDataLen))
	} else {
		err = binary.Write(buffer, binary.BigEndian, uint16(h.keyDataLen))
	}
	if err != nil {
		return err
	}
	if h.flags&0x1000 != 0 {
		err = binary.Write(buffer, binary.BigEndian, uint32(h.nodeDataLen))
	} else {
		err = binary.Write(buffer, binary.BigEndian, uint16(h.nodeDataLen))
	}
	if err != nil {
		return err
	}
	err = binary.Write(buffer, binary.BigEndian, h.tinyNodeCount)
	if err != nil {
		return err
	}
	// write key length
	// write keys
	return nil
}
func (h *header) read(buffer *bytes.Reader) error {
	var num uint32
	err := binary.Read(buffer, binary.BigEndian, &num)
	if err != nil {
		return err
	}
	if num != 4283062785 {
		return errors.New("corrupted json data")
	}
	h.version = int(num & 0xFF)
	if h.version < 1 || h.version > 4 {
		return fmt.Errorf("unsupported json version %d", h.version)
	}
	err = binary.Read(buffer, binary.BigEndian, &h.flags)
	if err != nil {
		return err
	}
	if h.flags&2 == 0 || h.flags&4 == 0 {
		return errors.New("corrupted json data")
	}
	keyCount, err := h.readKeyCount(buffer)
	if err != nil {
		return err
	}
	if h.flags&0x800 != 0 {
		var temp uint32
		err = binary.Read(buffer, binary.BigEndian, &temp)
		if err != nil {
			return err
		}
		h.keyDataLen = int(temp)
	} else {
		var temp uint16
		err = binary.Read(buffer, binary.BigEndian, &temp)
		if err != nil {
			return err
		}
		h.keyDataLen = int(temp)
	}
	// if version == 3
	// read 2 bytes flags2 ==> read 4 bytes keysCount2 ==> read 4 bytes key data2
	if h.version == 3 {
		err = binary.Read(buffer, binary.BigEndian, &h.flags2)
		if err != nil {
			return err
		}
		var temp uint16
		err = binary.Read(buffer, binary.BigEndian, &temp)
		if err != nil {
			return err
		}
		keyCount = int(temp)
		var temp2 uint32
		err = binary.Read(buffer, binary.BigEndian, &temp2)
		if err != nil {
			return err
		}
		h.keyDataLen = int(temp2)
	}
	if h.flags&0x1000 != 0 {
		var temp uint32
		err = binary.Read(buffer, binary.BigEndian, &temp)
		if err != nil {
			return err
		}
		h.nodeDataLen = int(temp)
	} else {
		var temp uint16
		err = binary.Read(buffer, binary.BigEndian, &temp)
		if err != nil {
			return err
		}
		h.nodeDataLen = int(temp)
	}
	err = binary.Read(buffer, binary.BigEndian, &h.tinyNodeCount)
	if err != nil {
		return err
	}
	err = h.readKeys(buffer, keyCount)
	if err != nil {
		return err
	}
	// at the end of reading key the cursor is on the start of the key buffer
	// so seek from current amount equal to key data length
	h.nodeOffset, err = buffer.Seek(int64(h.keyDataLen), io.SeekCurrent)
	return err
}
func (h *header) writeKeyCount(buffer *bytes.Buffer) error {
	var err error
	if h.flags&0x8 != 0 {
		err = binary.Write(buffer, binary.BigEndian, uint32(len(h.keys)))
	} else if h.flags&0x400 != 0 {
		err = binary.Write(buffer, binary.BigEndian, uint16(len(h.keys)))
	} else {
		err = buffer.WriteByte(uint8(len(h.keys)))
	}
	return err
}
func (h *header) readKeyCount(buffer *bytes.Reader) (int, error) {
	if h.flags&0x8 != 0 {
		var keyCount uint32
		err := binary.Read(buffer, binary.BigEndian, &keyCount)
		return int(keyCount), err
	}
	if h.flags&0x400 != 0 {
		var keyCount uint16
		err := binary.Read(buffer, binary.BigEndian, &keyCount)
		return int(keyCount), err
	}
	keyCount, err := buffer.ReadByte()
	return int(keyCount), err
}
func (h *header) readKeys(buffer *bytes.Reader, keysCount int) error {
	var err error
	var index int
	h.keys = make(keyCollection, keysCount)
	for index = 0; index < keysCount; index++ {
		h.keys[index].hash, err = buffer.ReadByte()
		if err != nil {
			return err
		}
	}
	var offset uint16
	for index = 0; index < keysCount; index++ {
		err = binary.Read(buffer, binary.BigEndian, &offset)
		h.keys[index].offset = int(offset)
	}
	var baseOffset int64
	baseOffset, err = buffer.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	var stringLen uint8
	for index = 0; index < keysCount; index++ {
		// seek to the start of the key
		_, err = buffer.Seek(baseOffset+int64(h.keys[index].offset), io.SeekStart)
		if err != nil {
			return err
		}
		stringLen, err = buffer.ReadByte()
		if err != nil {
			return err
		}
		nameBytes := make([]byte, stringLen)
		_, err = buffer.Read(nameBytes)
		if err != nil {
			return err
		}
		h.keys[index].name = string(nameBytes)
	}
	// seek to the base offset
	_, err = buffer.Seek(baseOffset, io.SeekStart)
	return err
}
func (h *header) isRelativeOffsetUsed() bool {
	return h.flags&1 != 0
}
func (h *header) pushBaseOffset(offset int64) {
	h.currentObjBaseOffset = append(h.currentObjBaseOffset, offset)
}
func (h *header) pushCurrentOffset(buffer *bytes.Reader) error {
	var baseOffset int64
	var err error
	baseOffset, err = buffer.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	h.pushBaseOffset(baseOffset)
	return nil
}
func (h *header) popBaseOffset() int64 {
	length := len(h.currentObjBaseOffset)
	if length == 0 {
		return 0
	}
	result := h.currentObjBaseOffset[length-1]
	h.currentObjBaseOffset = h.currentObjBaseOffset[:length-1]
	return result
}
func (h *header) currentBaseOffset() int64 {
	length := len(h.currentObjBaseOffset)
	if length == 0 {
		return 0
	}
	return h.currentObjBaseOffset[length-1]
}
func (h *header) absoluteOffset(offset int64) int64 {
	if h.isRelativeOffsetUsed() {
		return h.currentBaseOffset() + offset
	} else {
		return h.nodeOffset + offset
	}
}
