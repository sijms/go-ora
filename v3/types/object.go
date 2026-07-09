package types

import (
	"database/sql/driver"
)

type Object struct {
	Basic
	Name  string
	Owner string
	Value driver.Value
	//typ      reflect.Type
	//toid     []byte
	//fieldMap map[string]int
	//isArray  bool
	//attribs
	//fields OracleTyper
}

//func (obj *Object) SetValue(input interface{}) error {
//	return nil
//}

//func (obj *Object) Value() (interface{}, error) {
//	return nil, nil
//}

//func (obj *Object) Scan(input interface{}) error {
//	return nil
//}
//
//func (obj *Object) CopyTo(dest driver.Value) error {
//	return nil
//}
