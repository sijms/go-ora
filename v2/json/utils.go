package json

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
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

func ReadJsonString(input []byte) error {
	var output = make(map[string]interface{})
	err := json.Unmarshal(input, &output)
	fmt.Println(output)
	return err
	//inObject := false
	//inArray := false
	//inExp := false
	//keys := make([][]byte, 0, 20)
	//values := make([][]byte, 0, 20)
	//exp := make([]byte, 0, 50)
	//for i := 0; i < len(input); i++ {
	//	c := input[i]
	//	switch c {
	//	case '{':
	//		inObject = true
	//	case '}':
	//		inObject = false
	//
	//	case '[':
	//		inArray = true
	//	case ']':
	//		inArray = false
	//	case '"':
	//		inExp = !inExp
	//	case ':':
	//		keys = append(keys, exp)
	//		exp = nil
	//	case ',':
	//		values = append(values, exp)
	//		exp = nil
	//	case '%':
	//	case '\\':
	//	default:
	//		if inExp {
	//			exp = append(exp, c)
	//		}
	//	}
	//}
}

func readAllKeys(obj map[string]interface{}, keys KeyCollection) (KeyCollection, error) {
	var err error
	for key, value := range obj {
		keys = append(keys, *NewKey(key))
		rValue := reflect.ValueOf(value)
		if rValue.Kind() == reflect.Map {
			if temp, ok := value.(map[string]interface{}); ok {
				keys, err = readAllKeys(temp, keys)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return keys, nil
}

func encodeMainObject(mainObj map[string]interface{}, _sort bool) ([]byte, error) {
	keys, err := readAllKeys(mainObj, nil)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(nil)
	_, err = buffer.Write([]byte{0xff, 0x4a, 0x5a, 0x1})
	if err != nil {
		return nil, err
	}
	_, err = buffer.Write([]byte{0x21, 0x6})
	if err != nil {
		return nil, err
	}
	err = buffer.WriteByte(uint8(len(keys)))
	if err != nil {
		return nil, err
	}
	keyBuffer, err := keys.encode()
	if err != nil {
		return nil, err
	}
	if _sort {
		keys.Sort()
	}
	obj, err := NewObjectField(mainObj, keys)
	if err != nil {
		return nil, err
	}
	objectData, err := obj.Encode()
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.BigEndian, uint16(len(keyBuffer)))
	if err != nil {
		return nil, err
	}
	err = binary.Write(buffer, binary.BigEndian, uint16(len(objectData)))
	buffer.Write([]byte{0, 0})
	for _, key := range keys {
		err = buffer.WriteByte(key.hash)
		if err != nil {
			return nil, err
		}
	}
	for _, key := range keys {
		err = binary.Write(buffer, binary.BigEndian, uint16(key.offset))
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
func Encode(mainObj map[string]interface{}, _sort bool) ([]byte, error) {
	return encodeMainObject(mainObj, _sort)
}

func EncodeJsonString(input string, _sort bool) ([]byte, error) {
	var output = make(map[string]interface{})
	err := json.Unmarshal([]byte(input), &output)
	if err != nil {
		return nil, err
	}
	return Encode(output, _sort)
}
