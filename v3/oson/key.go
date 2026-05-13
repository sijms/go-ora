package oson

import (
	"bytes"
	"reflect"
	"sort"
	"strconv"
)

type Key struct {
	name   string
	size   int
	hash   uint8
	offset int
}

type KeyCollection []Key

func NewKey(name string) *Key {
	ret := new(Key)
	ret.name = name
	input := []rune(name)
	var num uint64 = 0xFFFFFFFF811C9DC5
	ret.size = len(input)
	for i := 0; i < len(input); i++ {
		c := input[i]
		var temp []byte
		if c <= 0x7ff {
			temp = []byte(string(c))
			for j := 0; j < len(temp); j++ {
				num = (num ^ uint64(temp[j])) * uint64(0x1000193)
			}
		} else {
			temp = []byte(strconv.Itoa(int(c)))
			for x := 0; x < len(temp); x++ {
				num = (num ^ uint64(temp[x])) * uint64(0x1000193)
			}
			if len(temp) == 2 {
				i++
			}
		}
		ret.size = ret.size + len(temp) - 1
	}
	ret.hash = uint8(num)
	return ret
}

//	func Hash(input string) (hash uint8, length int) {
//		key := []rune(input)
//		var num uint64 = 0xFFFFFFFF811C9DC5
//		length = len(key)
//		for i := 0; i < len(key); i++ {
//			c := key[i]
//			if c <= 0x7ff {
//				temp := []byte(string(c))
//				for j := 0; j < len(temp); j++ {
//					num = (num ^ uint64(temp[j])) * uint64(0x1000193)
//				}
//				length = length + len(temp) - 1
//			} else {
//				temp := strconv.Itoa(int(c))
//				for x := 0; x < len(temp); x++ {
//					num = (num ^ uint64(temp[x])) * uint64(0x1000193)
//				}
//				if len(temp) == 2 {
//					i++
//				}
//				length = length + len(temp) - 1
//			}
//		}
//		hash = uint8(num)
//		return
//	}

func (collection *KeyCollection) encode() ([]byte, error) {
	keyBuffer := bytes.NewBuffer(nil)
	var err error
	for x, key := range *collection {
		(*collection)[x].offset = keyBuffer.Len()
		err = keyBuffer.WriteByte(uint8(len(key.name)))
		if err != nil {
			return nil, err
		}
		_, err = keyBuffer.Write([]byte(key.name))
		if err != nil {
			return nil, err
		}
	}
	return keyBuffer.Bytes(), nil
}

func (collection *KeyCollection) decode(data []byte) error {
	//keyBuffer := bytes.NewBuffer(data)
	var err error
	return err
}
func (collection *KeyCollection) Sort() {
	sort.Slice(*collection, func(i, j int) bool {
		return (*collection)[i].hash < (*collection)[j].hash // Sort by hash ascending
	})
}

func (collection *KeyCollection) Index(keyName string) int {
	for i, key := range *collection {
		if key.name == keyName {
			return i
		}
	}
	return -1
}
func (collection *KeyCollection) Add(input *Key) {
	for _, key := range *collection {
		if key.name == input.name {
			return
		}
	}
	*collection = append(*collection, *input)
}
func (collection *KeyCollection) extractKeys(obj interface{}) {
	rValue := reflect.ValueOf(obj)
	switch rValue.Kind() {
	case reflect.Map:
		if temp, ok := obj.(map[string]interface{}); ok {
			for key, value := range temp {
				collection.Add(NewKey(key))
				collection.extractKeys(value)
			}
		}
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		for index := 0; index < rValue.Len(); index++ {
			if rValue.Index(index).CanInterface() {
				collection.extractKeys(rValue.Index(index).Interface())
			}
		}
	}
}
