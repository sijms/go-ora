package types

import (
	"database/sql/driver"
	"reflect"
	"strings"

	"github.com/sijms/go-ora/v3/utils"
)

type Object struct {
	Basic
	Name     string
	Owner    string
	typ      reflect.Type
	toid     []byte
	fieldMap map[string]int
	isArray  bool
	//attribs
	//fields OracleTyper
}

func (obj *Object) SetValue(input interface{}) error {
	return nil
}

func (obj *Object) Value() (interface{}, error) {
	return nil, nil
}

func (obj *Object) Scan(input interface{}) error {
	return nil
}

func (obj *Object) CopyTo(dest driver.Value) error {
	return nil
}

func (obj *Object) SetType(typ reflect.Type, isArray bool) {
	obj.typ = typ
	obj.isArray = isArray
}
func (obj *Object) loadFieldMap() {
	typ := obj.typ
	for x := 0; x < typ.NumField(); x++ {
		f := typ.Field(x)
		fieldID, _, _, _ := utils.ExtractTag(f.Tag.Get("udt"))
		if len(fieldID) == 0 {
			continue
		}
		fieldID = strings.ToUpper(fieldID)
		obj.fieldMap[fieldID] = x
	}
}
