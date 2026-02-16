package type_coder

import (
	"encoding/binary"
	"fmt"

	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type RowIDCoder struct {
	rba         int64
	partitionID int64
	filter      byte
	blockNumber int64
	slotNumber  int64
	data        []byte
	TypeInfo
}
type RowID struct {
}

func (coder *RowIDCoder) decode() []byte {
	output := make([]byte, 0, 18)
	output = append(output, convertRowIDToByte(coder.rba, 6)...)
	output = append(output, convertRowIDToByte(coder.partitionID, 3)...)
	output = append(output, convertRowIDToByte(coder.blockNumber, 6)...)
	output = append(output, convertRowIDToByte(coder.slotNumber, 3)...)
	return output
}

func (coder *RowIDCoder) Decode(data []byte) (interface{}, error) {
	switch coder.DataType {
	case types.ROWID:
		return coder.decode(), nil
	case types.UROWID:
		if data[0] == 1 {
			return string(coder.physicalRawIDToByteArray(data)), nil
		}
		return string(coder.logicalRawIDToByteArray(data)), nil
	}
	return nil, fmt.Errorf("ROWID decoder called with unsupported data type: %d", coder.DataType)
}
func (coder *RowIDCoder) Read(session network.SessionReader) (interface{}, error) {
	switch coder.DataType {
	case types.ROWID:
		length, err := session.GetByte()
		if err != nil {
			return nil, err
		}
		if length == 0 {
			return nil, nil
		}
		coder.rba, err = session.GetInt64(4, true, true)
		if err != nil {
			return nil, err
		}
		coder.partitionID, err = session.GetInt64(2, true, true)
		if err != nil {
			return nil, err
		}
		num, err := session.GetByte()
		if err != nil {
			return nil, err
		}
		coder.blockNumber, err = session.GetInt64(4, true, true)
		if err != nil {
			return nil, err
		}
		coder.slotNumber, err = session.GetInt64(2, true, true)
		if err != nil {
			return nil, err
		}
		if coder.rba == 0 && coder.partitionID == 0 && num == 0 && coder.blockNumber == 0 && coder.slotNumber == 0 {
			return nil, nil
		}
		return coder.Decode(nil)
	case types.UROWID:
		length, err := session.GetInt(4, true, true)
		if err != nil {
			return nil, err
		}
		var data []byte
		if length > 0 {
			data, err = session.GetClr()
			if err != nil {
				return nil, err
			}
		} else {
			data = nil
		}
		return coder.Decode(data)
	}
	return nil, fmt.Errorf("ROWID decoder called with unsupported data type: %d", coder.DataType)

}

func (coder *RowIDCoder) physicalRawIDToByteArray(data []byte) []byte {
	// physical
	temp32 := binary.BigEndian.Uint32(data[1:5])
	coder.rba = int64(temp32)
	temp16 := binary.BigEndian.Uint16(data[5:7])
	coder.partitionID = int64(temp16)
	temp32 = binary.BigEndian.Uint32(data[7:11])
	coder.blockNumber = int64(temp32)
	temp16 = binary.BigEndian.Uint16(data[11:13])
	coder.slotNumber = int64(temp16)
	if coder.rba == 0 {
		return []byte(fmt.Sprintf("%08X.%04X.%04X", coder.blockNumber, coder.slotNumber, coder.partitionID))
	}
	return coder.decode()
}

func (coder *RowIDCoder) logicalRawIDToByteArray(data []byte) []byte {
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

//func (row *RowID) Scan(value interface{}) error {
//	switch v := value.(type) {
//	case *RowID:
//		*row = *v
//	case RowID:
//		*row = v
//	case string:
//		row.Data = []byte(v)
//	case []byte:
//		row.Data = v
//	default:
//		return fmt.Errorf("row id column type require string or []byte value")
//	}
//	return nil
//}
