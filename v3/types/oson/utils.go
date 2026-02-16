package oson

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/types"
)

/*
	json operations:

flag = 0x80 = start object
flag = 0xC0 = start array
if object size or array size > 0xFFFF flag | 32
then write byte flag | 4

string:
if length < 31 op = length
if length < 0x100 op = 51
if length < 0x10000 op = 55
else: op = 56

double op = 54 followed by value
bool op = 49 if true and op=50 if false

decimal
convert to array
if array length <= 8 then op = (length - 1) | 0x60
if length < 0x100 op = 0x74 followed by length followed by bytes

int32
op = array length | 0x40

int64
op = length | 0x50

number as string
convert to string length should be <= 256 then op = 53

bytes
if length < 0x10000 op = 58 else op = 59
then length + value

Id
length should be <= 16
op = 0x7E

float
convert to bytes
op = 7F

null op = 48

date
op = 60 and length of buffer = 7

timestamp
op = 57 with fraction size = 11
op = 0x7D without fraction size = 7

timestamp TZ
op = 0x7C size = 13

interval ym
op = 61 and length = 5

interval DS
op = 62 and length = 11
*/

/*
	operation:

1- determine type of json input like string, dictionary and objects
2- parse input
*/

//func ReadJsonString(input []byte) error {
//	var output = make(map[string]interface{})
//	err := json.Unmarshal(input, &output)
//	fmt.Println(output)
//	return err
//	//inObject := false
//	//inArray := false
//	//inExp := false
//	//keys := make([][]byte, 0, 20)
//	//values := make([][]byte, 0, 20)
//	//exp := make([]byte, 0, 50)
//	//for i := 0; i < len(input); i++ {
//	//	c := input[i]
//	//	switch c {
//	//	case '{':
//	//		inObject = true
//	//	case '}':
//	//		inObject = false
//	//
//	//	case '[':
//	//		inArray = true
//	//	case ']':
//	//		inArray = false
//	//	case '"':
//	//		inExp = !inExp
//	//	case ':':
//	//		keys = append(keys, exp)
//	//		exp = nil
//	//	case ',':
//	//		values = append(values, exp)
//	//		exp = nil
//	//	case '%':
//	//	case '\\':
//	//	default:
//	//		if inExp {
//	//			exp = append(exp, c)
//	//		}
//	//	}
//	//}
//}

func Encode(mainObj interface{}) ([]byte, error) {
	header := &Header{
		flags: 0x2106,
		keys:  KeyCollection{},
	}
	var err error
	header.keys.extractKeys(mainObj)
	keyBuffer, err := header.keys.encode()
	if err != nil {
		return nil, err
	}
	header.keyDataLen = len(keyBuffer)
	var objectData []byte
	if dict, ok := mainObj.(map[string]interface{}); ok {
		obj, err := NewObjectField(dict, header)
		if err != nil {
			return nil, err
		}
		objectData, err = obj.Encode()
		if err != nil {
			return nil, err
		}
	}
	if array, ok := mainObj.([]interface{}); ok {
		obj, err := NewArrayField(array, header)
		if err != nil {
			return nil, err
		}
		objectData, err = obj.Encode()
		if err != nil {
			return nil, err
		}
	}
	header.nodeDataLen = len(objectData)
	//if _sort {
	//	keys.Sort()
	//}

	buffer := bytes.NewBuffer(nil)
	err = header.write(buffer)
	if err != nil {
		return nil, err
	}
	for _, key := range header.keys {
		err = buffer.WriteByte(key.hash)
		if err != nil {
			return nil, err
		}
	}
	for _, key := range header.keys {
		if header.flags&0x800 != 0 {
			err = binary.Write(buffer, binary.BigEndian, uint32(key.offset))
		} else {
			err = binary.Write(buffer, binary.BigEndian, uint16(key.offset))
		}
		if err != nil {
			return nil, err
		}
	}
	_, err = buffer.Write(keyBuffer)
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write(objectData)
	return buffer.Bytes(), err
}

func isObject(opCode uint8) bool {
	return opCode&0xC0 == 0x80
}
func isArray(opCode uint8) bool {
	return opCode&0xC0 == 0xC0
}
func readObjectHeader(buffer *bytes.Reader, header *Header, opCode uint8) (childCount int, keyIndex []int, err error) {
	var useWideOffset = opCode&0x20 != 0
	switch opCode & 0x18 {
	case 0x18:
		var offset int64
		if useWideOffset {
			var temp uint32
			err = binary.Read(buffer, binary.BigEndian, &temp)
			if err != nil {
				return
			}
			offset = int64(temp)
		} else {
			var temp uint16
			err = binary.Read(buffer, binary.BigEndian, &temp)
			if err != nil {
				return
			}
			offset = int64(temp)
		}
		offset = header.absoluteOffset(offset)
		var currentOffset int64
		currentOffset, err = buffer.Seek(0, io.SeekCurrent)
		if err != nil {
			return
		}
		_, err = buffer.Seek(offset, io.SeekStart)
		if err != nil {
			return
		}
		opCode, err = buffer.ReadByte()
		if err != nil {
			return
		}
		childCount, keyIndex, err = readObjectHeader(buffer, header, opCode)
		if err != nil {
			return
		}
		_, err = buffer.Seek(currentOffset, io.SeekStart)
		return
	case 0:
		var temp uint8
		temp, err = buffer.ReadByte()
		if err != nil {
			return
		}
		childCount = int(temp)
	case 8:
		var temp uint16
		err = binary.Read(buffer, binary.BigEndian, &temp)
		if err != nil {
			return
		}
		childCount = int(temp)
	case 0x10:
		var temp uint32
		err = binary.Read(buffer, binary.BigEndian, &temp)
		if err != nil {
			return
		}
		childCount = int(temp)
	}
	if childCount > 0 && isObject(opCode) {
		keyIndex = make([]int, childCount)
		for index := 0; index < childCount; index++ {
			if header.flags&0x8 > 0 {
				var temp uint32
				err = binary.Read(buffer, binary.BigEndian, &temp)
				keyIndex[index] = int(temp)
			} else if header.flags&0x400 > 0 {
				var temp uint16
				err = binary.Read(buffer, binary.BigEndian, &temp)
				keyIndex[index] = int(temp)
			} else {
				var temp uint8
				temp, err = buffer.ReadByte()
				keyIndex[index] = int(temp)
			}
			if err != nil {
				return
			}
		}
	}
	return
}

func decodeNode(buffer *bytes.Reader, header *Header) (Field, error) {
	var opCode uint8
	var err error
	err = header.pushCurrentOffset(buffer)
	if err != nil {
		return nil, err
	}
	defer header.popBaseOffset()

	// object header contain
	// 1- opCode
	// 2- child count
	// 3- key index in case of object not array
	opCode, err = buffer.ReadByte()
	if err != nil {
		return nil, err
	}
	var keyIndex []int
	if isArray(opCode) || isObject(opCode) {
		var childCount int
		childCount, keyIndex, err = readObjectHeader(buffer, header, opCode)
		if err != nil {
			return nil, err
		}
		var useWideOffset = opCode&0x20 != 0
		var index int
		var children []Field
		offset := make([]int64, childCount)
		if childCount > 0 {
			if useWideOffset {
				for index = 0; index < childCount; index++ {
					var temp uint32
					err = binary.Read(buffer, binary.BigEndian, &temp)
					if err != nil {
						return nil, err
					}
					offset[index] = int64(temp)
				}
			} else {
				for index = 0; index < childCount; index++ {
					var temp uint16
					err = binary.Read(buffer, binary.BigEndian, &temp)
					if err != nil {
						return nil, err
					}
					offset[index] = int64(temp)
				}
			}

			children = make([]Field, childCount)

			for index = 0; index < childCount; index++ {
				_, err = buffer.Seek(header.absoluteOffset(offset[index]), io.SeekStart)
				if err != nil {
					return nil, err
				}
				children[index], err = decodeNode(buffer, header)
				if err != nil {
					return nil, err
				}
				if keyIndex != nil {
					children[index].SetKeyIndex(int(keyIndex[index]))
				}
			}
		}

		if isArray(opCode) {
			value := make([]interface{}, childCount)
			if childCount > 0 {
				for index = 0; index < childCount; index++ {
					value[index] = children[index].Value()
				}
			}
			return &ArrayField{
				value: value,
				structField: structField{
					basicField: basicField{
						opCode:   opCode,
						keyIndex: 0,
						offset:   0,
						children: children,
					},
				},
			}, nil
		}
		if isObject(opCode) {

			value := make(map[string]interface{})
			if childCount > 0 {
				for index = 0; index < childCount; index++ {
					key := header.keys[children[index].KeyIndex()-1]
					value[key.name] = children[index].Value()
				}
			}
			return &ObjectField{
				value: value,
				structField: structField{
					basicField: basicField{
						opCode:   opCode,
						keyIndex: 0,
						offset:   0,
						children: children,
					},
				},
			}, nil
		}
	}
	if opCode <= 0x1F {
		temp := &StringField{}
		temp.opCode = opCode
		temp.bValue = make([]byte, opCode)
		_, err = buffer.Read(temp.bValue)
		if err != nil {
			return nil, err
		}
		temp.value = string(temp.bValue)
		return temp, nil
	}
	switch opCode {
	case 0x30:
		return &NullField{
			basicField: basicField{
				opCode: opCode,
			},
		}, nil
	case 0x31:
		return &BooleanField{
			value:      true,
			basicField: basicField{opCode: opCode},
		}, nil
	case 0x32:
		return &BooleanField{
			value:      false,
			basicField: basicField{opCode: opCode},
		}, nil
	case 0x33:
		var stringLen uint8
		stringLen, err = buffer.ReadByte()
		if err != nil {
			return nil, err
		}
		temp := &StringField{}
		temp.opCode = opCode
		temp.bValue = make([]byte, stringLen)
		_, err = buffer.Read(temp.bValue)
		if err != nil {
			return nil, err
		}
		temp.value = string(temp.bValue)
		return temp, nil
	case 0x36:
		var temp = make([]byte, 8)
		_, err = buffer.Read(temp)
		return &BinaryDoubleField{
			value: converters.ConvertBinaryDouble(temp),
			basicField: basicField{
				opCode: opCode,
			},
		}, nil
	case 0x37:
		var stringLen uint16
		err = binary.Read(buffer, binary.BigEndian, &stringLen)
		if err != nil {
			return nil, err
		}
		temp := &StringField{}
		temp.opCode = opCode
		temp.bValue = make([]byte, stringLen)
		_, err = buffer.Read(temp.bValue)
		if err != nil {
			return nil, err
		}
		temp.value = string(temp.bValue)
		return temp, nil
	case 0x38:
		var stringLen uint32
		err = binary.Read(buffer, binary.BigEndian, &stringLen)
		if err != nil {
			return nil, err
		}
		var temp = &StringField{}
		temp.opCode = opCode
		temp.bValue = make([]byte, stringLen)
		_, err = buffer.Read(temp.bValue)
		if err != nil {
			return nil, err
		}
		temp.value = string(temp.bValue)
		return temp, nil
	case 0x74:
		dataLen, err := buffer.ReadByte()
		if err != nil {
			return nil, err
		}
		numberBytes := make([]byte, dataLen)
		_, err = buffer.Read(numberBytes)
		if err != nil {
			return nil, err
		}
		number, err := types.NewNumber(numberBytes)
		if err != nil {
			return nil, err
		}
		return &NumberField{
			value:      *number,
			basicField: basicField{opCode: opCode},
		}, nil
	case 0x7F:
		var temp = make([]byte, 4)
		_, err = buffer.Read(temp)
		return &BinaryFloatField{
			value: converters.ConvertBinaryFloat(temp),
			basicField: basicField{
				opCode: opCode,
			},
		}, nil
	}
	if opCode&0x60 == 0x60 {
		dataLen := (int(opCode) & ^0x60) + 1
		numberBytes := make([]byte, dataLen)
		_, err = buffer.Read(numberBytes)
		if err != nil {
			return nil, err
		}
		number, err := types.NewNumber(numberBytes)
		if err != nil {
			return nil, err
		}
		var temp NumberField
		temp.value = *number
		temp.opCode = opCode
		return &NumberField{
			value:      *number,
			basicField: basicField{opCode: opCode},
		}, nil
	}
	return nil, fmt.Errorf("unsupported type code: %d", opCode)
}

func Decode(data []byte) (interface{}, error) {
	buffer := bytes.NewReader(data)
	header := &Header{}
	err := header.read(buffer)
	if err != nil {
		return nil, err
	}
	obj, err := decodeNode(buffer, header)
	if err != nil {
		return nil, err
	}
	return obj.Value(), nil
	//if value, ok := obj.Value().(map[string]interface{}); ok {
	//	return value, nil
	//}
	//if !isScalar(flags) {
	//	if flags&8 != 0 {
	//		fieldId := 4
	//		uniqueFields := binary.BigEndian.Uint32(data[index : index+4])
	//		index += 4
	//		fmt.Println(fieldId, uniqueFields)
	//	} else if flags&1024 != 0 {
	//		fieldId := 2
	//		uniqueFields := binary.BigEndian.Uint32(data[index : index+2])
	//		index += 2
	//		fmt.Println(fieldId, uniqueFields)
	//	} else {
	//		fieldId := 1
	//		uniqueFields := data[index]
	//		index++
	//		fmt.Println(fieldId, uniqueFields)
	//	}
	//	if flags&2048 != 0 {
	//		fieldHeapSize := binary.BigEndian.Uint32(data[index : index+4])
	//		index += 4
	//		fmt.Println(fieldHeapSize)
	//	} else {
	//		fieldHeapSize := binary.BigEndian.Uint16(data[index : index+2])
	//		index += 2
	//		fmt.Println(fieldHeapSize)
	//	}
	//	if version >= 3 {
	//		flags2 := binary.BigEndian.Uint16(data[index : index+2])
	//		index += 2
	//		uniqueFields2 := binary.BigEndian.Uint32(data[index : index+4])
	//		index += 4
	//		fieldHeapSize2 := binary.BigEndian.Uint32(data[index : index+4])
	//		index += 4
	//		fmt.Println(flags2, fieldHeapSize2, uniqueFields2)
	//	}
	//	if flags&4096 != 0 {
	//		treeSegmentSize := binary.BigEndian.Uint32(data[index : index+4])
	//		index += 4
	//		fmt.Println(treeSegmentSize)
	//	} else {
	//		treeSegmentSize := binary.BigEndian.Uint16(data[index : index+2])
	//		index += 2
	//		fmt.Println(treeSegmentSize)
	//	}
	//	tinyNodeCount := binary.BigEndian.Uint16(data[index : index+2])
	//	index += 2
	//	fmt.Println(tinyNodeCount)
	//} else {
	//	if flags&4096 != 0 {
	//		treeSegmentSize := binary.BigEndian.Uint32(data[index : index+4])
	//		index += 4
	//		fmt.Println(treeSegmentSize)
	//	} else {
	//		treeSegmentSize := binary.BigEndian.Uint16(data[index : index+2])
	//		index += 2
	//		fmt.Println(treeSegmentSize)
	//	}
	//}
	//if
	//if (this.uniqueFields > 0)
	//{
	//	this.ReadDictionary(b);
	//}
	//if (this.uniqueFields2 > 0)
	//{
	//	this.ReadDictionary2(b);
	//}
	//if (!this.IsSet(4) || !this.IsSet(2))
	//{
	//	throw new NotSupportedException();
	//}
}

//func EncodeJsonString(input string, _sort bool) ([]byte, error) {
//	input = strings.TrimSpace(input)
//	var err error
//	if strings.HasPrefix(input, "[") {
//		var output []interface{}
//		err = json.Unmarshal([]byte(input), &output)
//		if err != nil {
//			return nil, err
//		}
//		return Encode(output, _sort)
//	} else {
//		var output = make(map[string]interface{})
//		err = json.Unmarshal([]byte(input), &output)
//		if err != nil {
//			return nil, err
//		}
//		return Encode(output, _sort)
//	}
//
//}
