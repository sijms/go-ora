package types

import (
	"encoding/binary"
	"fmt"
)

type RowID struct {
	Basic
	Rba         int64
	PartitionID int64
	filter      byte
	BlockNumber int64
	SlotNumber  int64
}

func (rowid RowID) SetValue(_ interface{}) error {
	return fmt.Errorf("row id is returned from server only")
}

func (rowid RowID) Value() (interface{}, error) {
	switch rowid.dataType {
	case ROWID:
		return string(rowid.decode()), nil
	case UROWID:
		if len(rowid.bValue) > 0 && rowid.bValue[0] == 1 {
			return string(rowid.physicalRawIDToByteArray(rowid.bValue)), nil
		}
		return string(rowid.logicalRawIDToByteArray(rowid.bValue)), nil
	}
	return nil, fmt.Errorf("row id is returned from server only")
}

func (rowid *RowID) decode() []byte {
	var convertRowIDToByte = func(number int64, size int) []byte {
		buffer := []byte{
			65, 66, 67, 68, 69, 70, 71, 72,
			73, 74, 75, 76, 77, 78, 79, 80,
			81, 82, 83, 84, 85, 86, 87, 88,
			89, 90, 97, 98, 99, 100, 101, 102,
			103, 104, 105, 106, 107, 108, 109, 110,
			111, 112, 113, 114, 115, 116, 117, 118,
			119, 120, 121, 122, 48, 49, 50, 51,
			52, 53, 54, 55, 56, 57, 43, 47,
		}
		output := make([]byte, size)
		for x := size; x > 0; x-- {
			output[x-1] = buffer[number&0x3F]
			if number >= 0 {
				number = number >> 6
			} else {
				number = (number >> 6) + (2 << (32 + ^6))
			}
		}
		return output
	}
	output := make([]byte, 0, 18)
	output = append(output, convertRowIDToByte(rowid.Rba, 6)...)
	output = append(output, convertRowIDToByte(rowid.PartitionID, 3)...)
	output = append(output, convertRowIDToByte(rowid.BlockNumber, 6)...)
	output = append(output, convertRowIDToByte(rowid.SlotNumber, 3)...)
	return output
}

func (rowid *RowID) physicalRawIDToByteArray(data []byte) []byte {
	// physical
	temp32 := binary.BigEndian.Uint32(data[1:5])
	rowid.Rba = int64(temp32)
	temp16 := binary.BigEndian.Uint16(data[5:7])
	rowid.PartitionID = int64(temp16)
	temp32 = binary.BigEndian.Uint32(data[7:11])
	rowid.BlockNumber = int64(temp32)
	temp16 = binary.BigEndian.Uint16(data[11:13])
	rowid.SlotNumber = int64(temp16)
	if rowid.Rba == 0 {
		return []byte(fmt.Sprintf("%08X.%04X.%04X", rowid.BlockNumber, rowid.SlotNumber, rowid.PartitionID))
	}
	return rowid.decode()
}

func (rowid *RowID) logicalRawIDToByteArray(data []byte) []byte {
	length1 := len(data)
	num1 := length1 / 3
	num2 := length1 % 3
	num3 := num1 * 4
	num4 := 0
	if num2 > 1 {
		num4 = 3
	} else {
		num4 = num2
	}
	length2 := num3 + num4
	var output []byte = nil
	if length2 > 0 {
		KGRD_INDBYTE_CHAR := []byte{65, 42, 45, 40, 41}
		buffer := []byte{
			65, 66, 67, 68, 69, 70, 71, 72,
			73, 74, 75, 76, 77, 78, 79, 80,
			81, 82, 83, 84, 85, 86, 87, 88,
			89, 90, 97, 98, 99, 100, 101, 102,
			103, 104, 105, 106, 107, 108, 109, 110,
			111, 112, 113, 114, 115, 116, 117, 118,
			119, 120, 121, 122, 48, 49, 50, 51,
			52, 53, 54, 55, 56, 57, 43, 47,
		}
		output = make([]byte, length2)
		srcIndex := 0
		dstIndex := 1
		output[dstIndex] = KGRD_INDBYTE_CHAR[data[srcIndex]-1]
		length1 -= 1
		srcIndex++
		dstIndex++
		for length1 > 0 {
			output[dstIndex] = buffer[data[srcIndex]>>2]
			if length1 == 1 {
				output[dstIndex+1] = buffer[(data[srcIndex]&3)<<4]
				break
			}
			output[dstIndex+1] = buffer[(data[srcIndex]&3)<<4|(data[srcIndex+1]&0xF0)>>4]
			if length1 == 2 {
				output[dstIndex+2] = buffer[(data[srcIndex+1]&0xF)<<2]
				break
			}
			output[dstIndex+2] = buffer[(data[srcIndex+1]&0xF)<<2|(data[srcIndex+2]&0xC0)>>6]
			output[dstIndex+3] = buffer[data[srcIndex+2]&63]
			length1 -= 3
			srcIndex += 3
			dstIndex += 3
		}
	}
	return output
}
