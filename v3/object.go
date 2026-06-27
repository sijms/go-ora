package go_ora

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/sijms/go-ora/v3/parameter_coder"
	"github.com/sijms/go-ora/v3/utils"
)

type Object struct {
	Owner        string
	Name         string
	fields       []string
	attribs      map[string]parameter_coder.OracleParameterCoder
	typ          reflect.Type
	toid         []byte // type oid
	isArray      bool
	activeFields map[string]int
}

func (obj *Object) loadActiveFields() {
	typ := obj.typ
	for x := 0; x < typ.NumField(); x++ {
		f := typ.Field(x)
		fieldID, _, _, _ := utils.ExtractTag(f.Tag.Get("udt"))
		if len(fieldID) == 0 {
			continue
		}
		fieldID = strings.ToUpper(fieldID)
		obj.activeFields[fieldID] = x
	}
}

func (obj *Object) loadObjectTypeInfo(db *sql.DB) (err error) {
	if obj.typ == nil {
		err = errors.New("type object cannot be nil")
		return
	}
	if obj.typ.Kind() != reflect.Struct {
		err = errors.New("type object should be of structure type")
		return
	}
	obj.activeFields = make(map[string]int)
	obj.toid, err = getTOID2(db, obj.Owner, obj.Name)
	if err != nil {
		return
	}
	sqlText := `SELECT ATTR_NAME, ATTR_TYPE_NAME, LENGTH, ATTR_NO 
					FROM ALL_TYPE_ATTRS 
					WHERE UPPER(OWNER)=:1 AND UPPER(TYPE_NAME)=:2
					ORDER BY ATTR_NO`
	var rows *sql.Rows
	rows, err = db.Query(sqlText, strings.ToUpper(obj.Owner), strings.ToUpper(obj.Name))
	if err != nil {
		return
	}
	var (
		attName     sql.NullString
		attOrder    int64
		attTypeName sql.NullString
		length      sql.NullInt64
	)
	drv := db.Driver().(*OracleDriver)
	obj.attribs = make(map[string]parameter_coder.OracleParameterCoder)
	for rows.Next() {
		err = rows.Scan(&attName, &attTypeName, &length, &attOrder)
		if err != nil {
			return
		}
		// use fields to preserve order because go map is unordered
		obj.fields = append(obj.fields, attName.String)
		par, ok := drv.nameTypeCoder[strings.ToUpper(attTypeName.String)]
		if ok {
			obj.attribs[strings.ToUpper(attName.String)] = par.Copy()
			obj.attribs[strings.ToUpper(attName.String)].SetAsUDTPar()
		} else {
			err = fmt.Errorf("unsupported attribute type: %s", attTypeName.String)
			return
		}

		//var par parameter_coder.OracleParameterCoder
		//switch strings.ToUpper(attTypeName.String) {
		//case "NUMBER":
		//	par = &parameter_coder.NumberParameter{}
		//case "VARCHAR2":
		//	par = &parameter_coder.StringParameter{}
		//case "NVARCHAR2":
		//	par = &parameter_coder.StringParameter{BasicParameter: parameter_coder.BasicParameter{CharsetForm: 2}}
		//case "TIMESTAMP":
		//	par = &parameter_coder.DateParameter{}
		//case "DATE":
		//	par = &parameter_coder.DateParameter{}
		//case "TIMESTAMP WITH LOCAL TIME ZONE":
		//	par = &parameter_coder.DateParameter{}
		//case "RAW":
		//	par = &parameter_coder.RawParameter{}
		//case "BLOB":
		//	par = &parameter_coder.BlobParameter{}
		//case "CLOB":
		//	par = &parameter_coder.ClobParameter{}
		//case "NCLOB":
		//	temp := &parameter_coder.ClobParameter{}
		//	temp.CharsetForm = 2
		//	par = temp
		//default:
		//	found := false
		//
		//	for name, value := range drv.cusTyp {
		//		if strings.EqualFold(name, attTypeName.String) {
		//			found = true
		//			temp := &ObjectParameter{
		//				obj: value,
		//			}
		//			temp.ToID = value.toid
		//			par = temp
		//			break
		//		}
		//	}
		//	if !found {
		//		return fmt.Errorf("unsupported attribute type: %s", attTypeName.String)
		//	}
		//}
		//obj.attribs = append(obj.attribs, par)

	}
	if len(obj.attribs) == 0 {
		err = fmt.Errorf("unknown or empty type: %s", obj.Name)
	}
	obj.loadActiveFields()
	return
}

//func (obj *Object) loadArrayTypeInfo(db *sql.DB, elementTypeName string) (err error) {
//	drv := db.Driver().(*OracleDriver)
//
//	return
//	//var arrayParam parameter_coder.OracleParameterCoder
//	//switch strings.ToUpper(obj.Name) {
//	//case "NUMBER":
//	//	arrayParam = &parameter_coder.NumberParameter{}
//	//case "VARCHAR2":
//	//	arrayParam = &parameter_coder.StringParameter{}
//	//case "NVARCHAR2":
//	//	arrayParam = &parameter_coder.StringParameter{BasicParameter: parameter_coder.BasicParameter{CharsetForm: 2}}
//	//case "TIMESTAMP":
//	//	arrayParam = &parameter_coder.DateParameter{
//	//		BasicParameter: parameter_coder.BasicParameter{
//	//			DataType: types.TimeStampDTY,
//	//		},
//	//	}
//	//case "DATE":
//	//	arrayParam = &parameter_coder.DateParameter{
//	//		BasicParameter: parameter_coder.BasicParameter{DataType: types.DATE},
//	//	}
//	//case "TIMESTAMP WITH LOCAL TIME ZONE":
//	//	arrayParam = &parameter_coder.DateParameter{
//	//		BasicParameter: parameter_coder.BasicParameter{DataType: types.TimeStampLTZ_DTY},
//	//	}
//	//case "RAW":
//	//	arrayParam = &parameter_coder.RawParameter{}
//	//case "BLOB":
//	//	arrayParam = &parameter_coder.BlobParameter{}
//	//case "CLOB":
//	//	arrayParam = &parameter_coder.ClobParameter{}
//	//case "NCLOB":
//	//	temp := &parameter_coder.ClobParameter{}
//	//	temp.CharsetForm = 2
//	//	arrayParam = temp
//	//default:
//	//	temp := &ObjectParameter{}
//	//	childObj := &Object{}
//	//
//	//	err = childObj.loadObjectTypeInfo(db, input)
//	//	if err != nil {
//	//		return
//	//	}
//	//	temp.obj = *childObj
//	//	arrayParam = temp
//	//}
//	////obj.attribs = append(obj.attribs, arrayParam)
//	//return
//}
