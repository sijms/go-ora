package go_ora

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/sijms/go-ora/v3/lazy_init"
	oraTypes "github.com/sijms/go-ora/v3/types"

	"github.com/sijms/go-ora/v3/network"
)

var (
	//tyFloat64  = reflect.TypeOf((*float64)(nil)).Elem()
	//tyFloat32  = reflect.TypeOf((*float32)(nil)).Elem()
	//tyInt      = reflect.TypeOf((*int)(nil)).Elem()
	//tyInt8     = reflect.TypeOf((*int8)(nil)).Elem()
	//tyInt16    = reflect.TypeOf((*int16)(nil)).Elem()
	//tyInt32    = reflect.TypeOf((*int32)(nil)).Elem()
	//tyInt64    = reflect.TypeOf((*int64)(nil)).Elem()
	//tyUint     = reflect.TypeOf((*uint)(nil)).Elem()
	//tyUint8    = reflect.TypeOf((*uint8)(nil)).Elem()
	//tyUint16   = reflect.TypeOf((*uint16)(nil)).Elem()
	//tyUint32   = reflect.TypeOf((*uint32)(nil)).Elem()
	//tyUint64   = reflect.TypeOf((*uint64)(nil)).Elem()
	//tyBool     = reflect.TypeOf((*bool)(nil)).Elem()
	//tyBytes    = oraTypes.TyBytes
	//tyString   = oraTypes.TyString
	//tyNVarChar = reflect.TypeOf((*NVarChar)(nil)).Elem()
	//tyTime     = oraTypes.TyTime
	//tyTimeStamp       = reflect.TypeOf((*TimeStamp)(nil)).Elem()
	//tyTimeStampTZ = reflect.TypeOf((*TimeStampTZ)(nil)).Elem()
	tyClob = reflect.TypeOf((*Clob)(nil)).Elem()
	//tyNClob       = reflect.TypeOf((*NClob)(nil)).Elem()
	tyBlob = reflect.TypeOf((*Blob)(nil)).Elem()
	//tyBFile = oraTypes.TyBFile
	//tyVector       = reflect.TypeOf((*oraTypes.vector)(nil)).Elem()
	//tyNullByte     = oraTypes.TyNullByte
	//tyNullInt16    = oraTypes.TyNullInt16
	//tyNullInt32    = oraTypes.TyNullInt32
	//tyNullInt64    = oraTypes.TyNullInt64
	//tyNullFloat64  = oraTypes.TyNullFloat64
	//tyNullBool     = oraTypes.TyNullBool
	//tyNullString   = oraTypes.TyNullString
	//tyNullNVarChar = reflect.TypeOf((*NullNVarChar)(nil)).Elem()
	//tyNullTime     = oraTypes.TyNullTime
	//tyNullTimeStamp   = reflect.TypeOf((*NullTimeStamp)(nil)).Elem()
	//tyNullTimeStampTZ = reflect.TypeOf((*NullTimeStampTZ)(nil)).Elem()
	//tyRefCursor       = reflect.TypeOf((*RefCursor)(nil)).Elem()
	//tyPLBool          = reflect.TypeOf((*PLBool)(nil)).Elem()
	//tyObject          = reflect.TypeOf((*Object)(nil)).Elem()
	//tyNumber          = reflect.TypeOf((*Number)(nil)).Elem()
	//tyFloat32Array    = reflect.TypeOf((*[]float32)(nil)).Elem()
	//tyUint8Array      = reflect.TypeOf((*[]uint8)(nil)).Elem()
	//tyFloat64Array    = reflect.TypeOf((*[]float64)(nil)).Elem()
	//tyDictionary      = reflect.TypeOf((*map[string]interface{})(nil)).Elem()
)

func refineSqlText(text string) string {
	index := 0
	length := len(text)
	inSingleQuote := false
	inDoubleQuote := false
	skip := false
	lineComment := false
	textBuffer := make([]byte, 0, len(text))
	for ; index < length; index++ {
		ch := text[index]
		switch ch {
		case '\\':
			// bypass next character
			continue
		case '/':
			if !inDoubleQuote && !inSingleQuote {
				if index+1 < length && text[index+1] == '*' {
					index += 1
					skip = true
				}
			}
		case '*':
			if !inDoubleQuote && !inSingleQuote {
				if index+1 < length && text[index+1] == '/' {
					index += 1
					skip = false
				}
			}
		case '\'':
			if !skip && !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
		case '"':
			if !skip && !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
		case '-':
			if !skip {
				if index+1 < length && text[index+1] == '-' {
					index += 1
					lineComment = true
				}
			}
		case '\n':
			//if lineComment {
			//	lineComment = false
			//}
			if lineComment {
				lineComment = false
			} else {
				textBuffer = append(textBuffer, ch) // oheurtel : keep the line feed character
			}
		default:
			if skip || lineComment || inSingleQuote || inDoubleQuote {
				continue
			}
			textBuffer = append(textBuffer, text[index])
		}
	}
	return strings.TrimSpace(string(textBuffer))
}

var parameterNameRegexp = lazy_init.NewLazyInit(func() (interface{}, error) {
	return regexp.Compile(`:(\w+)`)
})

func parseQueryParametersNames(text string) (names []string, err error) {
	refinedSql := refineSqlText(text)

	var parameterNameRegexpAny interface{}
	parameterNameRegexpAny, err = parameterNameRegexp.GetValue()
	if err != nil {
		return nil, err
	}

	names = make([]string, 0, 10)
	matches := parameterNameRegexpAny.(*regexp.Regexp).FindAllStringSubmatch(refinedSql, -1)
	for _, match := range matches {
		if len(match) > 1 {
			names = append(names, match[1])
		}
	}
	return names, nil
}

//func copy_(dest, source any) error {
//	var err error
//	dst := reflect.ValueOf(dest)
//	if dst.Kind() != reflect.Ptr {
//		return fmt.Errorf("source is not a pointer")
//	}
//	// do pre-copy operation
//	if temp, ok := source.(oraTypes.Lob); ok {
//		if temp.GetReadMode() == configurations.LobReadMode_AUTO {
//			err = temp.Read(context.Background())
//			if err != nil {
//				return err
//			}
//		}
//	}
//
//	// actual copy will be done through scanner interface
//	if temp, ok := dest.(sql.Scanner); ok {
//		err = temp.Scan(source)
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}

func tSigned(input reflect.Type) bool {
	switch input.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func tUnsigned(input reflect.Type) bool {
	switch input.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func tInteger(input reflect.Type) bool {
	return tSigned(input) || tUnsigned(input)
}

func tFloat(input reflect.Type) bool {
	return input.Kind() == reflect.Float32 || input.Kind() == reflect.Float64
}

//func tNumber(input reflect.Type) bool {
//	return tInteger(input) || tFloat(input) || input == tyBool
//}

//func tNullNumber(input reflect.Type) bool {
//	switch input {
//	case tyNullBool, tyNullByte, tyNullInt16, tyNullInt32, tyNullInt64:
//		fallthrough
//	case tyNullFloat64:
//		return true
//	}
//	return false
//}

func processReset(err error, conn *Connection) error {
	if errors.Is(err, network.ErrConnReset) {
		err = conn.read()
	}
	if err != nil && isBadConn(err) {
		conn.setBad()
	}
	return err
}

func isBadConn(err error) bool {
	var opError *net.OpError
	var oraError *network.OracleError
	switch {
	case errors.Is(err, io.EOF):
		return true
	case errors.Is(err, syscall.EPIPE):
		return true
	case errors.As(err, &opError):
		if opError.Net == "tcp" && opError.Op == "write" {
			return true
		}
	case errors.As(err, &oraError):
		return oraError.Bad()
	}
	return false
}

func getTOID2(conn *sql.DB, owner, typeName string) ([]byte, error) {
	var toid []byte
	err := conn.QueryRow(`SELECT type_oid FROM ALL_TYPES WHERE UPPER(OWNER)=:1 AND UPPER(TYPE_NAME)=:2`,
		strings.ToUpper(owner), strings.ToUpper(typeName)).Scan(&toid)
	if errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("type: %s is not present or wrong type name", typeName)
	}
	return toid, err
}

//func getTOID(conn *Connection, owner, typeName string) ([]byte, error) {
//	sqlText := `SELECT type_oid FROM ALL_TYPES WHERE UPPER(OWNER)=:1 AND UPPER(TYPE_NAME)=:2`
//	stmt := NewStmt(sqlText, conn)
//	defer func(stmt *Stmt) {
//		_ = stmt.Close()
//	}(stmt)
//	var ret []byte
//	rows, err := stmt.Query_([]driver.NamedValue{
//		{Value: strings.ToUpper(owner)},
//		{Value: strings.ToUpper(typeName)},
//	})
//	if err != nil {
//		return nil, err
//	}
//	if rows.Next_() {
//		err = rows.Scan(&ret)
//		if err != nil {
//			return nil, err
//		}
//	}
//	if len(ret) == 0 {
//		return nil, fmt.Errorf("unknown type: %s", typeName)
//	}
//	return ret, rows.Err()
//}

func encodeObject(session *network.Session, objectData []byte, isArray bool) []byte {
	size := len(objectData)
	fieldsData := bytes.Buffer{}
	if isArray {
		fieldsData.Write([]byte{0x88, 0x1})
	} else {
		fieldsData.Write([]byte{0x84, 0x1})
	}
	if (size + 7) < 0xfe {
		size += 3
		fieldsData.Write([]byte{uint8(size)})
	} else {
		size += 7
		fieldsData.Write([]byte{0xfe})
		session.WriteInt(&fieldsData, size, 4, true, false)
	}
	fieldsData.Write(objectData)
	return fieldsData.Bytes()
}

//func putUDTAttributes(input *Object, pars []ParameterInfo, index int) ([]ParameterInfo, int) {
//	oPrimValue := make([]ParameterInfo, 0, len(input.attribs))
//	for _, attrib := range input.attribs {
//		if attrib.cusType != nil && !attrib.cusType.isArray {
//			var tempValue []ParameterInfo
//			tempValue, index = putUDTAttributes(attrib.cusType, pars, index)
//			attrib.oPrimValue = tempValue
//			oPrimValue = append(oPrimValue, attrib)
//		} else {
//			oPrimValue = append(oPrimValue, pars[index])
//			index++
//		}
//	}
//	return oPrimValue, index
//}

//func getUDTAttributes(input *Object, value reflect.Value) []ParameterInfo {
//	output := make([]ParameterInfo, 0, 10)
//	for _, attrib := range input.attribs {
//		fieldValue := reflect.Value{}
//		if value.IsValid() && value.Kind() == reflect.Struct {
//			if fieldIndex, ok := input.activeFields[attrib.Name]; ok {
//				fieldValue = value.Field(fieldIndex)
//			}
//		}
//		// if attribute is a nested type and not array
//		if attrib.cusType != nil && !attrib.cusType.isArray {
//			output = append(output, getUDTAttributes(attrib.cusType, fieldValue)...)
//		} else {
//			if isArrayValue(fieldValue) {
//				attrib.MaxNoOfArrayElements = 1
//			}
//			if fieldValue.IsValid() {
//				attrib.Value = fieldValue.Interface()
//			}
//
//			output = append(output, attrib)
//		}
//	}
//	return output
//}

func isArrayValue(val interface{}) bool {
	tyVal := reflect.TypeOf(val)
	if tyVal == nil {
		return false
	}
	for tyVal.Kind() == reflect.Ptr {
		tyVal = tyVal.Elem()
	}
	if tyVal != oraTypes.TyBytes && (tyVal.Kind() == reflect.Array || tyVal.Kind() == reflect.Slice) {
		return true
	}
	return false
}

//func decodeObject(conn *Connection, parent *ParameterInfo, temporaryLobs *[][]byte) error {
//	session := conn.session
//	if parent.parent == nil {
//		newState := network.SessionState{InBuffer: bytes.NewBuffer(parent.BValue)}
//		session.SaveState(&newState)
//		defer session.LoadState()
//		objectType, err := session.GetByte()
//		if err != nil {
//			return err
//		}
//		ctl, err := session.GetInt(4, true, true)
//		if err != nil {
//			return err
//		}
//		if ctl == 0xFE {
//			_, err = session.GetInt(4, false, true)
//			if err != nil {
//				return err
//			}
//		}
//		switch objectType {
//		case 0x88:
//			_ /*attribsLen*/, err := session.GetInt(2, true, true)
//			if err != nil {
//				return err
//			}
//
//			itemsLen, err := session.GetInt(2, false, true)
//			if err != nil {
//				return err
//			}
//			if itemsLen == 0xFE {
//				itemsLen, err = session.GetInt(4, false, true)
//				if err != nil {
//					return err
//				}
//			}
//			pars := make([]ParameterInfo, 0, itemsLen)
//			for x := 0; x < itemsLen; x++ {
//				tempPar := parent.cusType.attribs[""]
//				// if parent.cusType.isRegularArray() {
//				//
//				// } else {
//				// 	tempPar = parent.clone()
//				// }
//
//				tempPar.Direction = parent.Direction
//				if tempPar.DataType == oraTypes.XMLType {
//					ctlByte, err := session.GetByte()
//					if err != nil {
//						return err
//					}
//					var objectBufferSize int
//					if ctlByte == 0xFE {
//						objectBufferSize, err = session.GetInt(4, false, true)
//						if err != nil {
//							return err
//						}
//					} else {
//						objectBufferSize = int(ctlByte)
//					}
//					tempPar.BValue, err = session.GetBytes(objectBufferSize)
//					if err != nil {
//						return err
//					}
//					err = decodeObject(conn, &tempPar, temporaryLobs)
//				} else {
//					err = tempPar.decodePrimValue(conn, true)
//				}
//				if err != nil {
//					return err
//				}
//				pars = append(pars, tempPar)
//			}
//			parent.oPrimValue = pars
//		case 0x85: // xmltype
//			_, err = session.GetByte() // represent 1
//			if err != nil {
//				return err
//			}
//			dataType, err := session.GetInt(4, false, true) // represent 0x14
//			if err != nil {
//				return err
//			}
//			value, err := session.GetBytes(len(parent.BValue) - 8)
//			if err != nil {
//				return err
//			}
//			switch dataType {
//			case 0x14:
//				conv, err := conn.GetDefaultStringCoder()
//				if err != nil {
//					return err
//				}
//				parent.oPrimValue = conv.Decode(value)
//			case 0x11:
//				lob := LobStream{
//					sourceLocator: value,
//					conn:          conn,
//					charsetID:     conn.getDefaultCharsetID(),
//				}
//				var strValue string
//				err = setLob(reflect.ValueOf(&strValue), lob)
//				if err != nil {
//					return err
//				}
//				parent.oPrimValue = strValue
//			}
//		case 0x84:
//			// pars := make([]ParameterInfo, 0, len(parent.cusType.attribs))
//			// collect all attributes in one list
//			// pars := getUDTAttributes(parent.cusType, reflect.Value{})
//			pars := make([]ParameterInfo, 0, 10)
//			for _, attrib := range parent.cusType.attribs {
//				attrib.Direction = parent.Direction
//				attrib.parent = parent
//				// check if this an object or array and coming value is nil
//				if attrib.DataType == oraTypes.XMLType {
//					temp, err := session.Peek()
//					if err != nil {
//						return err
//					}
//					if temp == 0xFD || temp == 0xFF {
//						_, err = session.GetByte()
//						if err != nil {
//							return err
//						}
//					} else {
//						if attrib.cusType.isArray {
//							attrib.parent = nil
//							attrib.BValue, err = session.GetFixedClr()
//						}
//						err = decodeObject(conn, &attrib, temporaryLobs)
//						if err != nil {
//							return err
//						}
//					}
//				} else {
//					err := attrib.decodePrimValue(conn, true)
//					if err != nil {
//						return err
//					}
//				}
//				// err = attrib.decodePrimValue(conn, temporaryLobs, true)
//				// if err != nil {
//				// 	return err
//				// }
//				pars = append(pars, attrib)
//			}
//			parent.oPrimValue = pars
//			// for index, _ := range pars {
//			// 	pars[index].Direction = parent.Direction
//			// 	pars[index].parent = parent
//			// 	// if we get 0xFD this means null object
//			// 	err = pars[index].decodePrimValue(conn, temporaryLobs, true)
//			// 	if err != nil {
//			// 		return err
//			// 	}
//			// }
//			// fill pars in its place in sub types
//			// parent.oPrimValue, _ = putUDTAttributes(parent.cusType, pars, 0)
//		}
//	} else {
//		pars := make([]ParameterInfo, 0, 10)
//		for _, attrib := range parent.cusType.attribs {
//			attrib.Direction = parent.Direction
//			attrib.parent = parent
//			// check if this an object or array and coming value is nil
//			if attrib.DataType == oraTypes.XMLType {
//				temp, err := session.Peek()
//				if err != nil {
//					return err
//				}
//				if temp == 0xFD || temp == 0xFF {
//					_, err = session.GetByte()
//					if err != nil {
//						return err
//					}
//				} else {
//					if attrib.cusType.isArray {
//						attrib.parent = nil
//						nb, err := session.GetByte()
//						if err != nil {
//							return err
//						}
//						var size int
//						switch nb {
//						case 0:
//							size = 0
//						case 0xFE:
//							size, err = session.GetInt(4, false, true)
//							if err != nil {
//								return err
//							}
//						default:
//							size = int(nb)
//						}
//						if size > 0 {
//							attrib.BValue, err = session.GetBytes(size)
//							if err != nil {
//								return err
//							}
//						}
//					}
//					err = decodeObject(conn, &attrib, temporaryLobs)
//					if err != nil {
//						return err
//					}
//				}
//			} else {
//				err := attrib.decodePrimValue(conn, true)
//				if err != nil {
//					return err
//				}
//			}
//
//			pars = append(pars, attrib)
//		}
//		parent.oPrimValue = pars
//	}
//
//	return nil
//}

func parseInputField(structValue reflect.Value, name, _type string, fieldIndex int) (tempPar *ParameterInfo, err error) {
	tempPar = &ParameterInfo{
		Name:      name,
		Direction: Input,
	}
	fieldValue := structValue.Field(fieldIndex)
	for fieldValue.Type().Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			tempPar.Value = nil
			return
		}
		fieldValue = fieldValue.Elem()
	}
	if !fieldValue.IsValid() {
		tempPar.Value = nil
		return
	}
	if fieldValue.CanInterface() && fieldValue.Interface() == nil {
		tempPar.Value = nil
		return
	}
	typeErr := fmt.Errorf("error passing field %s as type %s", fieldValue.Type().Name(), _type)
	switch _type {
	case "number":
		tempPar.Value = getString(fieldValue.Interface())
		// var fieldVal float64
		//tempPar.Value, err = NewNumber(fieldValue.Interface()) // getFloat(fieldValue.Interface())
		//if err != nil {
		//	err = typeErr
		//	return
		//}
	case "varchar":
		tempPar.Value = getString(fieldValue.Interface())
	case "nvarchar":
		temp := &oraTypes.String{}
		temp.UseNCharset = true
		err = temp.SetValue(getString(fieldValue.Interface()))
		if err != nil {
			return nil, err
		}
		tempPar.Value = temp
	case "date":
		tempPar.Value, err = getDate(fieldValue.Interface())
		if err != nil {
			err = typeErr
			return
		}
	case "timestamp":
		var fieldVal = &oraTypes.Date{}
		var temp time.Time
		temp, err = getDate(fieldValue.Interface())
		if err != nil {
			err = typeErr
			return
		}
		fieldVal.SetDataType(oraTypes.TIMESTAMP)
		err = fieldVal.SetValue(temp)
		tempPar.Value = fieldVal
	//case "timestamp":
	//	var fieldVal time.Time
	//	fieldVal, err = getDate(fieldValue.Interface())
	//	if err != nil {
	//		err = typeErr
	//		return
	//	}
	//	tempPar.Value = TimeStamp(fieldVal)
	case "timestamptz":
		var fieldVal = &oraTypes.Date{}
		var temp time.Time
		temp, err = getDate(fieldValue.Interface())
		if err != nil {
			err = typeErr
			return
		}
		fieldVal.SetDataType(oraTypes.TIMESTAMPTZ)
		err = fieldVal.SetValue(temp)
		tempPar.Value = fieldVal
	case "raw":
		tempPar.Value, err = getBytes(fieldValue.Interface())
		if err != nil {
			err = typeErr
			return
		}
	case "clob":
		fieldVal := getString(fieldValue.Interface())
		clob := &oraTypes.Clob{}
		if len(fieldVal) == 0 {
			err = clob.SetValue(nil)
			//tempPar.Value, err = oraTypes.CreateClob()
			//tempPar.Value = Clob{Valid: false}
		} else {
			err = clob.SetValue(fieldVal)
			//tempPar.Value = Clob{String: fieldVal, Valid: true}
		}
		if err != nil {
			return nil, err
		}
		tempPar.Value = clob
	case "nclob":
		fieldVal := getString(fieldValue.Interface())
		clob := &oraTypes.Clob{}
		clob.UseNCharset = true
		if len(fieldVal) == 0 {
			err = clob.SetValue(nil)
			//tempPar.Value = NClob{Valid: false}
		} else {
			err = clob.SetValue(fieldVal)
			//tempPar.Value = NClob{String: fieldVal, Valid: true}
		}
		if err != nil {
			return nil, err
		}
		tempPar.Value = clob
	case "blob":
		var fieldVal []byte
		fieldVal, err = getBytes(fieldValue.Interface())
		if err != nil {
			err = typeErr
			return
		}
		tempPar.Value = Blob{Data: fieldVal}
	case "":
		tempPar.Value = fieldValue.Interface()
	default:
		err = typeErr
	}
	return
}

func isEqualLoc(zone1, zone2 *time.Location) bool {
	t := time.Now()
	t1 := t.In(zone1)
	t2 := t.In(zone2)
	//return t1.Equal(t2)
	name1, offset1 := t1.Zone()
	name2, offset2 := t2.Zone()
	return name1 == name2 && offset1 == offset2
}
