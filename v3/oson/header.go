package oson

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Header struct {
	version              int
	flags                uint16
	flags2               uint16
	keys                 KeyCollection
	keyDataLen           int
	nodeDataLen          int
	tinyNodeCount        uint16
	nodeOffset           int64
	currentObjBaseOffset []int64
	objectHeaders        []objectHeader
}

func (header *Header) write(buffer *bytes.Buffer) error {
	var err error
	_, err = buffer.Write([]byte{0xff, 0x4a, 0x5a, 0x1})
	if err != nil {
		return err
	}
	keyLen := len(header.keys)
	if keyLen > 0xFFFF {
		header.flags |= 8
	} else if keyLen > 0xFF {
		header.flags |= 0x400
	}
	if header.keyDataLen > 0xFFFF {
		header.flags |= 0x800
	}
	if header.nodeDataLen > 0xFFFF {
		header.flags |= 0x1000
	}
	err = binary.Write(buffer, binary.BigEndian, header.flags)
	if err != nil {
		return err
	}
	err = header.writeKeyCount(buffer)
	if header.flags&0x800 != 0 {
		err = binary.Write(buffer, binary.BigEndian, uint32(header.keyDataLen))
	} else {
		err = binary.Write(buffer, binary.BigEndian, uint16(header.keyDataLen))
	}
	if err != nil {
		return err
	}
	if header.flags&0x1000 != 0 {
		err = binary.Write(buffer, binary.BigEndian, uint32(header.nodeDataLen))
	} else {
		err = binary.Write(buffer, binary.BigEndian, uint16(header.nodeDataLen))
	}
	if err != nil {
		return err
	}
	err = binary.Write(buffer, binary.BigEndian, header.tinyNodeCount)
	if err != nil {
		return err
	}
	// write key length
	// write keys
	return nil
}
func (header *Header) read(buffer *bytes.Reader) error {
	var num uint32
	err := binary.Read(buffer, binary.BigEndian, &num)
	if err != nil {
		return err
	}
	if num != 4283062785 {
		return errors.New("corrupted json data")
	}
	header.version = int(num & 0xFF)
	if header.version < 1 || header.version > 4 {
		return fmt.Errorf("unsupported json version %d", header.version)
	}
	err = binary.Read(buffer, binary.BigEndian, &header.flags)
	if err != nil {
		return err
	}
	if header.flags&2 == 0 || header.flags&4 == 0 {
		return errors.New("corrupted json data")
	}
	keyCount, err := header.readKeyCount(buffer)
	if err != nil {
		return err
	}
	if header.flags&0x800 != 0 {
		var temp uint32
		err = binary.Read(buffer, binary.BigEndian, &temp)
		if err != nil {
			return err
		}
		header.keyDataLen = int(temp)
	} else {
		var temp uint16
		err = binary.Read(buffer, binary.BigEndian, &temp)
		if err != nil {
			return err
		}
		header.keyDataLen = int(temp)
	}
	// if version == 3
	// read 2 bytes flags2 ==> read 4 bytes keysCount2 ==> read 4 bytes key data2
	if header.version == 3 {
		err = binary.Read(buffer, binary.BigEndian, &header.flags2)
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
		header.keyDataLen = int(temp2)
	}
	if header.flags&0x1000 != 0 {
		var temp uint32
		err = binary.Read(buffer, binary.BigEndian, &temp)
		if err != nil {
			return err
		}
		header.nodeDataLen = int(temp)
	} else {
		var temp uint16
		err = binary.Read(buffer, binary.BigEndian, &temp)
		if err != nil {
			return err
		}
		header.nodeDataLen = int(temp)
	}
	err = binary.Read(buffer, binary.BigEndian, &header.tinyNodeCount)
	if err != nil {
		return err
	}
	err = header.readKeys(buffer, keyCount)
	if err != nil {
		return err
	}
	// at the end of reading key the cursor is on the start of the key buffer
	// so seek from current amount equal to key data length
	header.nodeOffset, err = buffer.Seek(int64(header.keyDataLen), io.SeekCurrent)
	return err
}
func (header *Header) writeKeyCount(buffer *bytes.Buffer) error {
	var err error
	if header.flags&0x8 != 0 {
		err = binary.Write(buffer, binary.BigEndian, uint32(len(header.keys)))
	} else if header.flags&0x400 != 0 {
		err = binary.Write(buffer, binary.BigEndian, uint16(len(header.keys)))
	} else {
		err = buffer.WriteByte(uint8(len(header.keys)))
	}
	return err
}
func (header *Header) readKeyCount(buffer *bytes.Reader) (int, error) {
	if header.flags&0x8 != 0 {
		var keyCount uint32
		err := binary.Read(buffer, binary.BigEndian, &keyCount)
		return int(keyCount), err
	}
	if header.flags&0x400 != 0 {
		var keyCount uint16
		err := binary.Read(buffer, binary.BigEndian, &keyCount)
		return int(keyCount), err
	}
	keyCount, err := buffer.ReadByte()
	return int(keyCount), err
}
func (header *Header) readKeys(buffer *bytes.Reader, keysCount int) error {
	var err error
	var index int
	header.keys = make(KeyCollection, keysCount)
	for index = 0; index < keysCount; index++ {
		header.keys[index].hash, err = buffer.ReadByte()
		if err != nil {
			return err
		}
	}
	var offset uint16
	for index = 0; index < keysCount; index++ {
		err = binary.Read(buffer, binary.BigEndian, &offset)
		header.keys[index].offset = int(offset)
	}
	var baseOffset int64
	baseOffset, err = buffer.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	var stringLen uint8
	for index = 0; index < keysCount; index++ {
		// seek to the start of the key
		_, err = buffer.Seek(baseOffset+int64(header.keys[index].offset), io.SeekStart)
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
		header.keys[index].name = string(nameBytes)
	}
	// seek to the base offset
	_, err = buffer.Seek(baseOffset, io.SeekStart)
	return err
}
func (header *Header) isRelativeOffsetUsed() bool {
	return header.flags&1 != 0
}
func (header *Header) pushBaseOffset(offset int64) {
	header.currentObjBaseOffset = append(header.currentObjBaseOffset, offset)
}
func (header *Header) pushCurrentOffset(buffer *bytes.Reader) error {
	var baseOffset int64
	var err error
	baseOffset, err = buffer.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	header.pushBaseOffset(baseOffset)
	return nil
}
func (header *Header) popBaseOffset() int64 {
	length := len(header.currentObjBaseOffset)
	if length == 0 {
		return 0
	}
	result := header.currentObjBaseOffset[length-1]
	header.currentObjBaseOffset = header.currentObjBaseOffset[:length-1]
	return result
}
func (header *Header) currentBaseOffset() int64 {
	length := len(header.currentObjBaseOffset)
	if length == 0 {
		return 0
	}
	return header.currentObjBaseOffset[length-1]
}
func (header *Header) absoluteOffset(offset int64) int64 {
	if header.isRelativeOffsetUsed() {
		return header.currentBaseOffset() + offset
	} else {
		return header.nodeOffset + offset
	}
}
