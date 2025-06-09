package go_ora

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/sijms/go-ora/v2/lazy_init"

	"github.com/sijms/go-ora/v2/network"
)

var (
	tyFloat64         = reflect.TypeOf((*float64)(nil)).Elem()
	tyFloat32         = reflect.TypeOf((*float32)(nil)).Elem()
	tyInt64           = reflect.TypeOf((*int64)(nil)).Elem()
	tyBool            = reflect.TypeOf((*bool)(nil)).Elem()
	tyBytes           = reflect.TypeOf((*[]byte)(nil)).Elem()
	tyString          = reflect.TypeOf((*string)(nil)).Elem()
	tyNVarChar        = reflect.TypeOf((*NVarChar)(nil)).Elem()
	tyTime            = reflect.TypeOf((*time.Time)(nil)).Elem()
	tyTimeStamp       = reflect.TypeOf((*TimeStamp)(nil)).Elem()
	tyTimeStampTZ     = reflect.TypeOf((*TimeStampTZ)(nil)).Elem()
	tyClob            = reflect.TypeOf((*Clob)(nil)).Elem()
	tyNClob           = reflect.TypeOf((*NClob)(nil)).Elem()
	tyBlob            = reflect.TypeOf((*Blob)(nil)).Elem()
	tyBFile           = reflect.TypeOf((*BFile)(nil)).Elem()
	tyVector          = reflect.TypeOf((*Vector)(nil)).Elem()
	tyNullByte        = reflect.TypeOf((*sql.NullByte)(nil)).Elem()
	tyNullInt16       = reflect.TypeOf((*sql.NullInt16)(nil)).Elem()
	tyNullInt32       = reflect.TypeOf((*sql.NullInt32)(nil)).Elem()
	tyNullInt64       = reflect.TypeOf((*sql.NullInt64)(nil)).Elem()
	tyNullFloat64     = reflect.TypeOf((*sql.NullFloat64)(nil)).Elem()
	tyNullBool        = reflect.TypeOf((*sql.NullBool)(nil)).Elem()
	tyNullString      = reflect.TypeOf((*sql.NullString)(nil)).Elem()
	tyNullNVarChar    = reflect.TypeOf((*NullNVarChar)(nil)).Elem()
	tyNullTime        = reflect.TypeOf((*sql.NullTime)(nil)).Elem()
	tyNullTimeStamp   = reflect.TypeOf((*NullTimeStamp)(nil)).Elem()
	tyNullTimeStampTZ = reflect.TypeOf((*NullTimeStampTZ)(nil)).Elem()
	tyRefCursor       = reflect.TypeOf((*RefCursor)(nil)).Elem()
	tyPLBool          = reflect.TypeOf((*PLBool)(nil)).Elem()
	tyObject          = reflect.TypeOf((*Object)(nil)).Elem()
	tyNumber          = reflect.TypeOf((*Number)(nil)).Elem()
	tyFloat32Array    = reflect.TypeOf((*[]float32)(nil)).Elem()
	tyUint8Array      = reflect.TypeOf((*[]uint8)(nil)).Elem()
	tyFloat64Array    = reflect.TypeOf((*[]float64)(nil)).Elem()
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
			if index+1 < length && text[index+1] == '*' {
				index += 1
				skip = true
			}
		case '*':
			if index+1 < length && text[index+1] == '/' {
				index += 1
				skip = false
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

func extractTag(tag string) (name, _type string, size int, direction ParameterDirection) {
	extractNameValue := func(input string, pos int) {
		parts := strings.Split(input, "=")
		var id, value string
		if len(parts) < 2 {
			switch pos {
			case 0:
				id = "name"
			case 1:
				id = "type"
			case 2:
				id = "size"
			case 3:
				id = "direction"
			}
			value = input
		} else {
			id = strings.TrimSpace(strings.ToLower(parts[0]))
			value = strings.TrimSpace(parts[1])
		}
		switch id {
		case "name":
			name = value
		case "type":
			_type = value
		case "size":
			tempSize, _ := strconv.ParseInt(value, 10, 32)
			size = int(tempSize)
		case "dir":
			fallthrough
		case "direction":
			switch value {
			case "in", "input":
				direction = Input
			case "out", "output":
				direction = Output
			case "inout":
				direction = InOut
			}
		}
	}
	tag = strings.TrimSpace(tag)
	if len(tag) == 0 {
		return
	}
	tagFields := strings.Split(tag, ",")
	if len(tagFields) > 0 {
		extractNameValue(tagFields[0], 0)
	}
	if len(tagFields) > 1 {
		extractNameValue(tagFields[1], 1)
	}
	if len(tagFields) > 2 {
		extractNameValue(tagFields[2], 2)
	}
	if len(tagFields) > 3 {
		extractNameValue(tagFields[3], 3)
	}
	return
}

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

func tNumber(input reflect.Type) bool {
	return tInteger(input) || tFloat(input) || input == tyBool
}

func tNullNumber(input reflect.Type) bool {
	switch input {
	case tyNullBool, tyNullByte, tyNullInt16, tyNullInt32, tyNullInt64:
		fallthrough
	case tyNullFloat64:
		return true
	}
	return false
}

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

func getTOID(conn *Connection, owner, typeName string) ([]byte, error) {
	sqlText := `SELECT type_oid FROM ALL_TYPES WHERE UPPER(OWNER)=:1 AND UPPER(TYPE_NAME)=:2`
	stmt := NewStmt(sqlText, conn)
	defer func(stmt *Stmt) {
		_ = stmt.Close()
	}(stmt)
	var ret []byte
	rows, err := stmt.Query_([]driver.NamedValue{
		{Value: strings.ToUpper(owner)},
		{Value: strings.ToUpper(typeName)},
	})
	if err != nil {
		return nil, err
	}
	if rows.Next_() {
		err = rows.Scan(&ret)
		if err != nil {
			return nil, err
		}
	}
	if len(ret) == 0 {
		return nil, fmt.Errorf("unknown type: %s", typeName)
	}
	return ret, rows.Err()
}

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

func putUDTAttributes(input *customType, pars []ParameterInfo, index int) ([]ParameterInfo, int) {
	oPrimValue := make([]ParameterInfo, 0, len(input.attribs))
	for _, attrib := range input.attribs {
		if attrib.cusType != nil && !attrib.cusType.isArray {
			var tempValue []ParameterInfo
			tempValue, index = putUDTAttributes(attrib.cusType, pars, index)
			attrib.oPrimValue = tempValue
			oPrimValue = append(oPrimValue, attrib)
		} else {
			oPrimValue = append(oPrimValue, pars[index])
			index++
		}
	}
	return oPrimValue, index
}

func getUDTAttributes(input *customType, value reflect.Value) []ParameterInfo {
	output := make([]ParameterInfo, 0, 10)
	for _, attrib := range input.attribs {
		fieldValue := reflect.Value{}
		if value.IsValid() && value.Kind() == reflect.Struct {
			if fieldIndex, ok := input.fieldMap[attrib.Name]; ok {
				fieldValue = value.Field(fieldIndex)
			}
		}
		// if attribute is a nested type and not array
		if attrib.cusType != nil && !attrib.cusType.isArray {
			output = append(output, getUDTAttributes(attrib.cusType, fieldValue)...)
		} else {
			if isArrayValue(fieldValue) {
				attrib.MaxNoOfArrayElements = 1
			}
			if fieldValue.IsValid() {
				attrib.Value = fieldValue.Interface()
			}

			output = append(output, attrib)
		}
	}
	return output
}

func isArrayValue(val interface{}) bool {
	tyVal := reflect.TypeOf(val)
	if tyVal == nil {
		return false
	}
	for tyVal.Kind() == reflect.Ptr {
		tyVal = tyVal.Elem()
	}
	if tyVal != tyBytes && (tyVal.Kind() == reflect.Array || tyVal.Kind() == reflect.Slice) {
		return true
	}
	return false
}

func decodeObject(conn *Connection, parent *ParameterInfo, temporaryLobs *[][]byte) error {
	session := conn.session
	if parent.parent == nil {
		newState := network.SessionState{InBuffer: parent.BValue}
		session.SaveState(&newState)
		defer session.LoadState()
		objectType, err := session.GetByte()
		if err != nil {
			return err
		}
		ctl, err := session.GetInt(4, true, true)
		if err != nil {
			return err
		}
		if ctl == 0xFE {
			_, err = session.GetInt(4, false, true)
			if err != nil {
				return err
			}
		}
		switch objectType {
		case 0x88:
			_ /*attribsLen*/, err := session.GetInt(2, true, true)
			if err != nil {
				return err
			}

			itemsLen, err := session.GetInt(2, false, true)
			if err != nil {
				return err
			}
			if itemsLen == 0xFE {
				itemsLen, err = session.GetInt(4, false, true)
				if err != nil {
					return err
				}
			}
			pars := make([]ParameterInfo, 0, itemsLen)
			for x := 0; x < itemsLen; x++ {
				tempPar := parent.cusType.attribs[0]
				// if parent.cusType.isRegularArray() {
				//
				// } else {
				// 	tempPar = parent.clone()
				// }

				tempPar.Direction = parent.Direction
				if tempPar.DataType == XMLType {
					ctlByte, err := session.GetByte()
					if err != nil {
						return err
					}
					var objectBufferSize int
					if ctlByte == 0xFE {
						objectBufferSize, err = session.GetInt(4, false, true)
						if err != nil {
							return err
						}
					} else {
						objectBufferSize = int(ctlByte)
					}
					tempPar.BValue, err = session.GetBytes(objectBufferSize)
					if err != nil {
						return err
					}
					err = decodeObject(conn, &tempPar, temporaryLobs)
					if err != nil {
						return err
					}
				} else {
					err = tempPar.decodePrimValue(conn, temporaryLobs, true)
					if err != nil {
						return err
					}
				}
				if err != nil {
					return err
				}
				pars = append(pars, tempPar)
			}
			parent.oPrimValue = pars
		case 0x85: // xmltype
			_, err = session.GetByte() // represent 1
			if err != nil {
				return err
			}
			dataType, err := session.GetInt(4, false, true) // represent 0x14
			if err != nil {
				return err
			}
			value, err := session.GetBytes(len(parent.BValue) - 8)
			if err != nil {
				return err
			}
			switch dataType {
			case 0x14:
				conv, err := conn.getDefaultStrConv()
				if err != nil {
					return err
				}
				parent.oPrimValue = conv.Decode(value)
			case 0x11:
				lob := Lob{
					sourceLocator: value,
					sourceLen:     len(value),
					connection:    conn,
					charsetID:     conn.getDefaultCharsetID(),
				}
				var strValue string
				err = setLob(reflect.ValueOf(&strValue), lob)
				if err != nil {
					return err
				}
				parent.oPrimValue = strValue
			}
		case 0x84:
			// pars := make([]ParameterInfo, 0, len(parent.cusType.attribs))
			// collect all attributes in one list
			// pars := getUDTAttributes(parent.cusType, reflect.Value{})
			pars := make([]ParameterInfo, 0, 10)
			for _, attrib := range parent.cusType.attribs {
				attrib.Direction = parent.Direction
				attrib.parent = parent
				// check if this an object or array and comming value is nil
				if attrib.DataType == XMLType {
					temp, err := session.Peek(1)
					if err != nil {
						return err
					}
					if temp[0] == 0xFD || temp[0] == 0xFF {
						_, err = session.GetByte()
						if err != nil {
							return err
						}
					} else {
						if attrib.cusType.isArray {
							attrib.parent = nil
							attrib.BValue, err = session.GetFixedClr()
						}
						err = decodeObject(conn, &attrib, temporaryLobs)
						if err != nil {
							return err
						}
					}
				} else {
					err := attrib.decodePrimValue(conn, temporaryLobs, true)
					if err != nil {
						return err
					}
				}
				// err = attrib.decodePrimValue(conn, temporaryLobs, true)
				// if err != nil {
				// 	return err
				// }
				pars = append(pars, attrib)
			}
			parent.oPrimValue = pars
			// for index, _ := range pars {
			// 	pars[index].Direction = parent.Direction
			// 	pars[index].parent = parent
			// 	// if we get 0xFD this means null object
			// 	err = pars[index].decodePrimValue(conn, temporaryLobs, true)
			// 	if err != nil {
			// 		return err
			// 	}
			// }
			// fill pars in its place in sub types
			// parent.oPrimValue, _ = putUDTAttributes(parent.cusType, pars, 0)
		}
	} else {
		pars := make([]ParameterInfo, 0, 10)
		for _, attrib := range parent.cusType.attribs {
			attrib.Direction = parent.Direction
			attrib.parent = parent
			// check if this an object or array and comming value is nil
			if attrib.DataType == XMLType {
				temp, err := session.Peek(1)
				if err != nil {
					return err
				}
				if temp[0] == 0xFD || temp[0] == 0xFF {
					_, err = session.GetByte()
					if err != nil {
						return err
					}
				} else {
					if attrib.cusType.isArray {
						attrib.parent = nil
						nb, err := session.GetByte()
						if err != nil {
							return err
						}
						var size int
						switch nb {
						case 0:
							size = 0
						case 0xFE:
							size, err = session.GetInt(4, false, true)
							if err != nil {
								return err
							}
						default:
							size = int(nb)
						}
						if size > 0 {
							attrib.BValue, err = session.GetBytes(size)
							if err != nil {
								return err
							}
						}
					}
					err = decodeObject(conn, &attrib, temporaryLobs)
					if err != nil {
						return err
					}
				}
			} else {
				err := attrib.decodePrimValue(conn, temporaryLobs, true)
				if err != nil {
					return err
				}
			}

			pars = append(pars, attrib)
		}
		parent.oPrimValue = pars
	}

	return nil
}

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
		// var fieldVal float64
		tempPar.Value, err = NewNumber(fieldValue.Interface()) // getFloat(fieldValue.Interface())
		if err != nil {
			err = typeErr
			return
		}
	case "varchar":
		tempPar.Value = getString(fieldValue.Interface())
	case "nvarchar":
		tempPar.Value = NVarChar(getString(fieldValue.Interface()))
	case "date":
		tempPar.Value, err = getDate(fieldValue.Interface())
		if err != nil {
			err = typeErr
			return
		}
	case "timestamp":
		var fieldVal time.Time
		fieldVal, err = getDate(fieldValue.Interface())
		if err != nil {
			err = typeErr
			return
		}
		tempPar.Value = TimeStamp(fieldVal)
	case "timestamptz":
		var fieldVal time.Time
		fieldVal, err = getDate(fieldValue.Interface())
		if err != nil {
			err = typeErr
			return
		}
		tempPar.Value = TimeStampTZ(fieldVal)
	case "raw":
		tempPar.Value, err = getBytes(fieldValue.Interface())
		if err != nil {
			err = typeErr
			return
		}
	case "clob":
		fieldVal := getString(fieldValue.Interface())
		if len(fieldVal) == 0 {
			tempPar.Value = Clob{Valid: false}
		} else {
			tempPar.Value = Clob{String: fieldVal, Valid: true}
		}
	case "nclob":
		fieldVal := getString(fieldValue.Interface())
		if len(fieldVal) == 0 {
			tempPar.Value = NClob{Valid: false}
		} else {
			tempPar.Value = NClob{String: fieldVal, Valid: true}
		}
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
	return t1.Equal(t2)
}
