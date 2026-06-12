package types

import (
	"encoding/json"
	"errors"
	"strings"
	//"github.com/sijms/go-ora/v3/oson"
)

//type Json interface {
//}

type JsonDecoder interface {
	DecodeJson(data []byte) (*Json, error)
}
type Json struct {
	Value   interface{}
	bValue  []byte
	decoder JsonDecoder
	lobBase
}

var jsonTypeError = errors.New("unexpected data for json type")

func CreateJson(input interface{}) (*Json, error) {
	j := new(Json)
	var err error
	if input == nil {
		return j, nil
	}
	switch value := input.(type) {
	case string:
		value = strings.TrimSpace(value)
		if len(value) == 0 {
			return j, nil
		}
		if strings.HasPrefix(value, "[") {
			var temp []interface{}
			err = json.Unmarshal([]byte(value), &temp)
			j.Value = temp
		} else {
			var temp = make(map[string]interface{})
			err = json.Unmarshal([]byte(value), &temp)
			j.Value = temp
		}
		if err != nil {
			return nil, err
		}
	case map[string]interface{}:
		j.Value = value
	case []map[string]interface{}:
		j.Value = value
	default:
		return nil, jsonTypeError
	}
	return j, nil
}

func NewJson(input interface{}, stream LobStreamer, decoder JsonDecoder) (*Json, error) {
	j, err := CreateJson(input)
	if err != nil {
		return nil, err
	}
	j.stream = stream
	j.decoder = decoder
	return j, nil
}
func (j *Json) Scan(input interface{}) error {
	//var err error
	if input == nil {
		if j.stream != nil {
			j.stream.SetLocator(nil)
		}
		j.Value = nil
		return nil
	}
	switch value := input.(type) {
	case *Json:
		*j = *value
		//err = j.lobBase.copyFrom(&value.lobBase)
		//if err != nil {
		//	return nil
		//}
		//j.Value = value.Value
	default:
		temp, err := CreateJson(value)
		if err != nil {
			return err
		}
		err = j.Scan(temp)
		if err != nil {
			return err
		}
	}
	return nil
}

func (j *Json) decode() error {
	var err error
	//j.Value, err = oson.Decode(j.bValue)
	return err
}

func (j *Json) encode() error {
	var err error
	//j.bValue, err = oson.Encode(j.Value)
	return err
}

func (j *Json) setNil() {
	j.Value = nil
	j.bValue = nil
}
