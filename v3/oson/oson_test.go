package oson

import (
	"encoding/json"
	"testing"
)

func TestOson(t *testing.T) {
	var err error
	input := `{
	"id": 5, 
	"name": "this is a test",
	"time": "2026-05-21T10:59:59"
    }`
	var obj map[string]interface{}
	err = json.Unmarshal([]byte(input), &obj)
	if err != nil {
		t.Error(err)
		return
	}
	var outputBytes []byte
	outputBytes, err = Encode(obj)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(outputBytes)
	output, err := Decode(outputBytes)
	if err != nil {
		t.Error(err)
	}
	t.Log(output)
}
