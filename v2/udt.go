package go_ora

import (
	"reflect"
	"strings"
)

type customType struct {
	owner string
	name  string
	//arrayTypeName string
	attribs []ParameterInfo
	typ     reflect.Type
	toid    []byte // type oid
	//arrayTOID     []byte
	isArray  bool
	fieldMap map[string]int
}

// RegisterType register user defined type with owner equal to user id
//func (conn *Connection) RegisterType(typeName, arrayTypeName string, typeObj interface{}) error {
//	return conn.RegisterTypeWithOwner(conn.connOption.UserID, typeName, arrayTypeName, typeObj)
//}

// RegisterTypeWithOwner take typename, owner and go type object and make an information
// structure that used to create a new type during query and store values in it
//
// DataType of UDT field that can be manipulated by this function are: NUMBER,
// VARCHAR2, NVARCHAR2, TIMESTAMP, DATE AND RAW
//func (conn *Connection) RegisterTypeWithOwner(owner, typeName, arrayTypeName string, typeObj interface{}) error {
//	if typeObj == nil {
//		return errors.New("type object cannot be nil")
//	}
//	typ := reflect.TypeOf(typeObj)
//	switch typ.Kind() {
//	case reflect.Ptr:
//		return errors.New("unsupported type object: Ptr")
//	case reflect.Array:
//		return errors.New("unsupported type object: Array")
//	case reflect.Chan:
//		return errors.New("unsupported type object: Chan")
//	case reflect.Map:
//		return errors.New("unsupported type object: Map")
//	case reflect.Slice:
//		return errors.New("unsupported type object: Slice")
//	}
//	if typ.Kind() != reflect.Struct {
//		return errors.New("type object should be of structure type")
//	}
//	cust := customType{
//		owner:         owner,
//		name:          typeName,
//		arrayTypeName: arrayTypeName,
//		typ:           typ,
//		fieldMap:      map[string]int{},
//	}
//	var err error
//	cust.toid, err = getTOID(conn, cust.owner, cust.name)
//	if err != nil {
//		return err
//	}
//	if len(cust.arrayTypeName) > 0 {
//		cust.arrayTOID, err = getTOID(conn, cust.owner, cust.arrayTypeName)
//		if err != nil {
//			return err
//		}
//	}
//	sqlText := `SELECT ATTR_NAME, ATTR_TYPE_NAME, LENGTH, ATTR_NO
//FROM ALL_TYPE_ATTRS WHERE UPPER(OWNER)=:1 AND UPPER(TYPE_NAME)=:2`
//
//	stmt := NewStmt(sqlText, conn)
//	defer func(stmt *Stmt) {
//		_ = stmt.Close()
//	}(stmt)
//	rows, err := stmt.Query_([]driver.NamedValue{{Value: strings.ToUpper(owner)}, {Value: strings.ToUpper(typeName)}})
//	if err != nil {
//		return err
//	}
//	var (
//		attName     sql.NullString
//		attOrder    int64
//		attTypeName sql.NullString
//		length      sql.NullInt64
//	)
//	for rows.Next_() {
//		err = rows.Scan(&attName, &attTypeName, &length, &attOrder)
//		if err != nil {
//			return err
//		}
//
//		for int(attOrder) > len(cust.attribs) {
//			cust.attribs = append(cust.attribs, ParameterInfo{
//				Direction: Input,
//				Flag:      3,
//			})
//		}
//		param := &cust.attribs[attOrder-1]
//		param.Name = attName.String
//		param.TypeName = attTypeName.String
//		switch strings.ToUpper(attTypeName.String) {
//		case "NUMBER":
//			param.DataType = NUMBER
//			param.MaxLen = converters.MAX_LEN_NUMBER
//		case "VARCHAR2":
//			param.DataType = NCHAR
//			param.CharsetForm = 1
//			param.ContFlag = 16
//			param.MaxCharLen = int(length.Int64)
//			param.CharsetID = conn.tcpNego.ServerCharset
//			param.MaxLen = int(length.Int64) * converters.MaxBytePerChar(param.CharsetID)
//		case "NVARCHAR2":
//			param.DataType = NCHAR
//			param.CharsetForm = 2
//			param.ContFlag = 16
//			param.MaxCharLen = int(length.Int64)
//			param.CharsetID = conn.tcpNego.ServernCharset
//			param.MaxLen = int(length.Int64) * converters.MaxBytePerChar(param.CharsetID)
//		case "TIMESTAMP":
//			fallthrough
//		case "DATE":
//			param.DataType = DATE
//			param.MaxLen = 11
//		case "RAW":
//			param.DataType = RAW
//			param.MaxLen = int(length.Int64)
//		case "BLOB":
//			param.DataType = OCIBlobLocator
//			param.MaxLen = int(length.Int64)
//		case "CLOB":
//			param.DataType = OCIClobLocator
//			param.CharsetForm = 1
//			param.ContFlag = 16
//			param.CharsetID = conn.tcpNego.ServerCharset
//			param.MaxCharLen = int(length.Int64)
//			param.MaxLen = int(length.Int64) * converters.MaxBytePerChar(param.CharsetID)
//		case "NCLOB":
//			param.DataType = OCIClobLocator
//			param.CharsetForm = 2
//			param.ContFlag = 16
//			param.CharsetID = conn.tcpNego.ServernCharset
//			param.MaxCharLen = int(length.Int64)
//			param.MaxLen = int(length.Int64) * converters.MaxBytePerChar(param.CharsetID)
//		default:
//			// search for type in registered types
//			found := false
//			for name, value := range conn.cusTyp {
//				if name == strings.ToUpper(attTypeName.String) {
//					found = true
//					param.cusType = new(customType)
//					*param.cusType = value
//					param.ToID = value.toid
//					break
//				}
//				if value.arrayTypeName == strings.ToUpper(attTypeName.String) {
//					found = true
//					param.cusType = new(customType)
//					*param.cusType = value
//					param.ToID = value.toid
//					break
//				}
//			}
//			if !found {
//				return fmt.Errorf("unsupported attribute type: %s", attTypeName.String)
//			}
//		}
//	}
//for {
//	err = rows.Next(values)
//	if err != nil {
//		if errors.Is(err, io.EOF) {
//			break
//		}
//		return err
//	}
//	if attName, ok = values[0].(string); !ok {
//		return errors.New(fmt.Sprint("error reading attribute properties for type: ", typeName))
//	}
//	if attTypeName, ok = values[1].(string); !ok {
//		return errors.New(fmt.Sprint("error reading attribute properties for type: ", typeName))
//	}
//	if values[2] == nil {
//		length = 0
//	} else {
//		if length, ok = values[2].(int64); !ok {
//			return fmt.Errorf("error reading attribute properties for type: %s", typeName)
//		}
//	}
//	if attOrder, ok = values[3].(int64); !ok {
//		return fmt.Errorf("error reading attribute properties for type: %s", typeName)
//	}
//
//}
//	if len(cust.attribs) == 0 {
//		return fmt.Errorf("unknown or empty type: %s", typeName)
//	}
//	cust.loadFieldMap()
//	conn.cusTyp[strings.ToUpper(typeName)] = cust
//	return nil
//}

//func (cust *customType) getFieldIndex(name string) int {
//	for x := 0; x < cust.typ.NumField(); x++ {
//		fieldID, _, _, _ := extractTag(cust.typ.Field(x).Tag.Get("udt"))
//		if strings.ToUpper(fieldID) == strings.ToUpper(name) {
//			return x
//		}
//	}
//	return -1
//}

// loadFieldMap read struct tag that supplied with golang type object passed in RegisterType
// function
func (cust *customType) loadFieldMap() {
	typ := cust.typ
	for x := 0; x < typ.NumField(); x++ {
		f := typ.Field(x)
		fieldID, _, _, _ := extractTag(f.Tag.Get("udt"))
		if len(fieldID) == 0 {
			continue
		}
		fieldID = strings.ToUpper(fieldID)
		cust.fieldMap[fieldID] = x
	}
}

//func (cust *customType) isRegularArray() bool {
//	return cust.isArray && len(cust.attribs) > 0 && cust.attribs[0].DataType != XMLType
//}

type Object struct {
	Owner       string
	Name        string
	Value       interface{}
	itemMaxSize int
}

// NewObject call this function to wrap oracle object types or arrayso
func NewObject(owner, name string, value interface{}) *Object {
	return &Object{
		Owner:       owner,
		Name:        name,
		Value:       value,
		itemMaxSize: 0,
	}
}

// NewArrayObject call this function for creation of array of regular type
// for example sql: create or replace TYPE SLICE AS TABLE OF varchar2(500)
// in this case itemMaxSize will be 500.
// if you use this function to define array of object type itemMaxSize is not used
//func NewArrayObject(owner, name string, itemMaxSize int, value interface{}) *Object {
//	return &Object{
//		Owner:       owner,
//		Name:        name,
//		Value:       value,
//		itemMaxSize: itemMaxSize,
//	}
//}
