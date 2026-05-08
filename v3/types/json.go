package types

import (
	"encoding/json"
	"errors"
	"strings"
)

//type Json interface {
//}

type JsonDecoder interface {
	DecodeJson(data []byte) (*Json, error)
}
type Json struct {
	Data    interface{}
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
			j.Data = temp
		} else {
			var temp = make(map[string]interface{})
			err = json.Unmarshal([]byte(value), &temp)
			j.Data = temp
		}
		if err != nil {
			return nil, err
		}
	case map[string]interface{}:
		j.Data = value
	case []map[string]interface{}:
		j.Data = value
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
	var err error
	if input == nil {
		if j.stream != nil {
			j.stream.SetLocator(nil)
		}
		j.Data = nil
		return nil
	}
	switch value := input.(type) {
	case *Json:
		err = j.lobBase.copyFrom(&value.lobBase)
		if err != nil {
			return nil
		}
		j.Data = value.Data
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
