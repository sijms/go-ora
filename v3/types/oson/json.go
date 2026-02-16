package oson

import (
	"encoding/json"
	"errors"
	"strings"
)

type Json struct {
	Value  interface{}
	bValue []byte
}

var jsonTypeError = errors.New("unexpected data for json type")

func NewJsonBytes(input []byte) (*Json, error) {
	output := new(Json)
	output.bValue = input
	if len(input) == 0 {
		output.setNil()
		return output, nil
	}
	err := output.decode()
	return output, err
}
func NewJsonString(input string) (*Json, error) {
	input = strings.TrimSpace(input)
	var err error
	output := new(Json)
	if len(input) == 0 {
		output.setNil()
		return output, nil
	}
	if strings.HasPrefix(input, "[") {
		var temp []interface{}
		err = json.Unmarshal([]byte(input), &temp)
		output.Value = temp

	} else {
		var temp = make(map[string]interface{})
		err = json.Unmarshal([]byte(input), &temp)
		output.Value = temp
	}
	if err != nil {
		return nil, err
	}
	err = output.encode()
	return output, err
}
func (j *Json) String() (string, error) {
	jsonBytes, err := json.Marshal(j.Value)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
func (j *Json) Bytes() []byte {
	return j.bValue
}
func (j *Json) decode() error {
	var err error
	j.Value, err = Decode(j.bValue)
	return err
}

func (j *Json) encode() error {
	var err error
	j.bValue, err = Encode(j.Value)
	return err
}

func (j *Json) setNil() {
	j.Value = nil
	j.bValue = nil
}

func (j *Json) Scan(input interface{}) error {
	return nil
}

// func (j Json) SetDataType(conn *Connection, par *ParameterInfo) error {
// 	par.DataType = JSON
// 	return nil
// }
