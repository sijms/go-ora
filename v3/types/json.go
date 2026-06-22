package types

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	//"github.com/sijms/go-ora/v3/oson"
)

// type Json interface {
// }
type JsonCoder interface {
	JsonEncoder
	JsonDecoder
}

type JsonDecoder interface {
	DecodeJson(data []byte) (interface{}, error)
}
type JsonEncoder interface {
	EncodeJson(input interface{}) ([]byte, error)
}
type Json struct {
	Basic
	Coder JsonCoder
	loc   Locator
	lobBase
}

var jsonTypeError = errors.New("unexpected data for json type")

func (js *Json) GetLocator() Locator {
	if js.lobBase.GetLocator() != nil {
		return js.lobBase.GetLocator()
	}
	return js.loc
}
func (js *Json) Upload() error {
	return js.uploadData(js.bValue, 0, 0)
}

func (js *Json) SetValue(input interface{}) error {
	if input == nil {
		js.bValue = nil
		return nil
	}
	var err error
	switch value := input.(type) {
	case Json:
		*js = value
	case *Json:
		*js = *value
	case string:
		value = strings.TrimSpace(value)
		if len(value) == 0 {
			js.bValue = nil
			return nil
		}
		if strings.HasPrefix(value, "[") {
			var temp []interface{}
			err = json.Unmarshal([]byte(value), &temp)
			if err != nil {
				return err
			}
			return js.SetValue(temp)
		}

		var temp = make(map[string]interface{})
		err = json.Unmarshal([]byte(value), &temp)
		if err != nil {
			return err
		}
		return js.SetValue(temp)
	default:
		js.bValue, err = js.Coder.EncodeJson(input)
		//return fmt.Errorf("cannot set value of type %T into Json", input)
	}
	if err == nil {
		dataLen := uint64(len(js.bValue))
		if dataLen > 0 {
			js.loc = NewQuasiLocator(dataLen)
		} else {
			js.loc = nil
		}
	}
	return err
}

func (js *Json) Value() (interface{}, error) {
	return js.Coder.DecodeJson(js.bValue)
}

func (js *Json) Scan(input interface{}) error {
	return js.SetValue(input)
}

func (js *Json) CopyTo(dest driver.Value) error {
	value, err := js.Value()
	if err != nil {
		return err
	}
	switch dst := dest.(type) {
	case *Json:
		*dst = *js
	case *string:
		if value == nil {
			*dst = ""
		} else {
			temp, err := json.Marshal(value)
			if err != nil {
				return err
			}
			*dst = string(temp)
		}
	case *map[string]interface{}:
		if value == nil {
			*dst = nil
		} else {
			*dst = value.(map[string]interface{})
		}

	case *[]interface{}:
		if value == nil {
			*dst = nil
		} else {
			*dst = value.([]interface{})
		}
	default:
		return fmt.Errorf("cannot copy Json to variable of type %T", dest)
	}
	return nil
}

func (js *Json) Read(ctx context.Context) error {
	var err error
	js.bValue, err = js.ReadFromPos(ctx, 0)
	return err
}
