package go_ora

import (
	"encoding/json"
	"errors"

	"github.com/sijms/go-ora/v2/oson"
)

type Json struct {
	Value  map[string]interface{}
	bValue []byte
	lob    Lob
}

var jsonTypeError = errors.New("unexpected data for json type")

func NewJson(input string) (*Json, error) {
	output := new(Json)
	output.Value = make(map[string]interface{})
	err := json.Unmarshal([]byte(input), &output.Value)
	if err != nil {
		return nil, err
	}
	//return Encode(output, _sort)
	//return new(Json), nil
	return output, nil
}

func (j *Json) decode(value string) error {
	return nil
}

func (j *Json) encode() error {
	var err error
	j.bValue, err = oson.Encode(j.Value, false)
	return err
}

func (j *Json) setNil() {

}

func (j *Json) Scan(input interface{}) error {
	return nil
}

func (j Json) SetDataType(conn *Connection, par *ParameterInfo) error {
	par.DataType = JSON
	return nil
}
