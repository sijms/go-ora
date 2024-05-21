package go_ora

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/sijms/go-ora/v2/converters"
)

func (par *ParameterInfo) setDataType(goType reflect.Type, value driver.Value, conn *Connection) error {
	if goType == nil {
		par.DataType = NCHAR
		return nil
	}
	for goType.Kind() == reflect.Ptr {
		goType = goType.Elem()
	}
	if goType == tyObject {
		val, err := getValue(value)
		if err != nil {
			return err
		}
		if obj, ok := val.(Object); ok {
			par.DataType = XMLType
			par.Value = obj.Value
			// set custom type
			for name, cusTyp := range conn.cusTyp {
				if strings.EqualFold(name, obj.Name) {
					par.cusType = new(customType)
					*par.cusType = cusTyp
					par.ToID = cusTyp.toid
					if cusTyp.isArray {
						par.MaxNoOfArrayElements = 1
					} else {
						par.Version = 1
					}
					break
				}
			}
			if par.cusType == nil {
				return fmt.Errorf("type %s is not created or not registered", obj.Name)
			}
			return nil
		}
	}
	if goType != tyBytes && (goType.Kind() == reflect.Array || goType.Kind() == reflect.Slice) {
		val, err := getValue(value)
		if err != nil {
			return err
		}
		var inVal driver.Value = nil
		if val != nil {
			rValue := reflect.ValueOf(val)
			size := rValue.Len()
			if size > 0 && rValue.Index(0).CanInterface() {
				inVal = rValue.Index(0).Interface()
			}
		}
		par.Flag = 0x43
		err = par.setDataType(goType.Elem(), inVal, conn)
		if err != nil {
			return err
		}
		if par.DataType == XMLType {
			// par.cusType is for item I should get that of array
			found := false
			for _, cust := range conn.cusTyp {
				if cust.isArray && len(cust.attribs) > 0 {
					if par.cusType.name == cust.attribs[0].cusType.name {
						found = true
						//par.TypeName = name
						par.ToID = cust.toid
						*par.cusType = cust
						par.Flag = 0x3
						break
					}
				}
			}
			if !found {
				return fmt.Errorf("can't get the collection of type %s", par.cusType.name)
			}
		}
		par.MaxNoOfArrayElements = 1
		return nil
	}
	if tNumber(goType) || tNullNumber(goType) {
		par.DataType = NUMBER
		par.MaxLen = converters.MAX_LEN_NUMBER
		return nil
	}

	switch goType {
	case tyNumber:
		par.DataType = NUMBER
		par.MaxLen = converters.MAX_LEN_NUMBER
	case tyPLBool:
		par.DataType = Boolean
		par.MaxLen = converters.MAX_LEN_BOOL
	case tyString, tyNullString:
		par.DataType = NCHAR
		par.CharsetForm = 1
		par.ContFlag = 16
		par.CharsetID = conn.tcpNego.ServerCharset
	case tyNVarChar, tyNullNVarChar:
		par.DataType = NCHAR
		par.CharsetForm = 2
		par.ContFlag = 16
		par.CharsetID = conn.tcpNego.ServernCharset
	case tyTime, tyNullTime:
		if par.Flag&0x40 > 0 {
			par.DataType = DATE
			par.MaxLen = converters.MAX_LEN_DATE
		} else {
			par.DataType = TimeStampTZ_DTY
			par.MaxLen = converters.MAX_LEN_TIMESTAMP
		}
	case tyTimeStamp, tyNullTimeStamp:
		//if par.Flag&0x43 > 0 {
		par.DataType = TIMESTAMP
		par.MaxLen = converters.MAX_LEN_DATE
		//} else {
		//	par.DataType = TimeStampTZ_DTY
		//	par.MaxLen = converters.MAX_LEN_TIMESTAMP
		//}
	case tyTimeStampTZ, tyNullTimeStampTZ:
		par.DataType = TimeStampTZ_DTY
		par.MaxLen = converters.MAX_LEN_TIMESTAMP
	//case tyTime, tyNullTime:
	//	if par.Direction == Input {
	//		par.DataType = TIMESTAMP
	//		par.MaxLen = converters.MAX_LEN_TIMESTAMP
	//	} else {
	//		par.DataType = DATE
	//		par.MaxLen = converters.MAX_LEN_DATE
	//	}
	//case tyTimeStamp, tyNullTimeStamp:
	//	if par.Direction == Input {
	//		par.DataType = TIMESTAMP
	//		par.MaxLen = converters.MAX_LEN_TIMESTAMP
	//	} else {
	//		par.DataType = TIMESTAMP
	//		par.MaxLen = converters.MAX_LEN_DATE
	//	}
	//case tyTimeStampTZ, tyNullTimeStampTZ:
	//	par.DataType = TimeStampTZ_DTY
	//	par.MaxLen = converters.MAX_LEN_TIMESTAMP
	case tyBytes:
		par.DataType = RAW
	case tyClob:
		par.DataType = OCIClobLocator
		par.CharsetForm = 1
		par.CharsetID = conn.tcpNego.ServerCharset
	case tyNClob:
		par.DataType = OCIClobLocator
		par.CharsetForm = 2
		par.CharsetID = conn.tcpNego.ServernCharset
	case tyBlob:
		par.DataType = OCIBlobLocator
	case tyBFile:
		par.DataType = OCIFileLocator
	case tyRefCursor:
		par.DataType = REFCURSOR
	default:
		rOriginal := reflect.ValueOf(value)
		if value != nil && !(rOriginal.Kind() == reflect.Ptr && rOriginal.IsNil()) {
			proVal := reflect.Indirect(rOriginal)
			if valuer, ok := proVal.Interface().(driver.Valuer); ok {
				val, err := valuer.Value()
				if err != nil {
					return err
				}
				if val == nil {
					par.DataType = NCHAR
					return nil
				}
				if val != value {
					return par.setDataType(reflect.TypeOf(val), val, conn)
				}
			}
		}

		//val, err := getValue(value)
		//if err != nil {
		//	return err
		//}
		//if val == nil {
		//	par.DataType = NCHAR
		//	return nil
		//}

		if goType.Kind() == reflect.Struct {
			// see if the struct is support valuer interface

			for _, cusTyp := range conn.cusTyp {
				if goType == cusTyp.typ {
					par.cusType = new(customType)
					*par.cusType = cusTyp
					par.ToID = cusTyp.toid
					//par.TypeName = cusTyp.name
				}
			}
			if par.cusType == nil {
				return errors.New("call register type before use user defined type (UDT)")
			}
			par.Version = 1
			par.DataType = XMLType
			par.MaxLen = 2000
		} else {
			return fmt.Errorf("unsupported go type: %v", goType.Name())
		}
	}
	return nil
}

func (par *ParameterInfo) encodeWithType(connection *Connection) error {
	var err error
	var val driver.Value
	val, err = getValue(par.Value)
	if err != nil {
		return err
	}
	if val == nil {
		par.IsNull = true
		par.iPrimValue = nil
		return nil
	}
	// check if array
	//if par.MaxNoOfArrayElements > 0 && par.cusType == nil {
	if par.MaxNoOfArrayElements > 0 {
		if !isArrayValue(val) {
			return fmt.Errorf("parameter %s require array value", par.Name)
		}
		var size int
		rValue := reflect.ValueOf(val)
		if isArrayValue(val) {
			size = rValue.Len()
		}
		if size == 0 {
			par.IsNull = true
			par.iPrimValue = nil
			return nil
		}
		if size > par.MaxNoOfArrayElements {
			par.MaxNoOfArrayElements = size
		}
		pars := make([]ParameterInfo, 0, size)
		var tempPar ParameterInfo
		for x := 0; x < size; x++ {
			if par.cusType != nil && par.cusType.isArray {
				tempPar = par.cusType.attribs[0].clone()
			} else {
				tempPar = par.clone()
			}
			if rValue.Index(x).CanInterface() {
				tempPar.Value = rValue.Index(x).Interface()
			}
			err = tempPar.encodeWithType(connection)
			if err != nil {
				return err
			}
			pars = append(pars, tempPar)
		}
		par.iPrimValue = pars
		return nil
	}
	switch par.DataType {
	case Boolean:
		par.iPrimValue, err = getBool(val)
		if err != nil {
			return err
		}
	case NUMBER:
		par.iPrimValue, err = NewNumber(val)
		if err != nil {
			return err
		}
	case NCHAR:
		tempString := getString(val)
		length := len(tempString)
		par.MaxCharLen = length
		par.iPrimValue = tempString
		if length > connection.maxLen.varchar {
			par.DataType = LongVarChar
		}
	case DATE:
		fallthrough
	case TIMESTAMP:
		fallthrough
	case TimeStampTZ_DTY:
		par.iPrimValue, err = getDate(val)
		if err != nil {
			return err
		}
	case RAW:
		var tempByte []byte
		tempByte, err = getBytes(val)
		if err != nil {
			return err
		}
		par.MaxLen = len(tempByte)
		par.iPrimValue = tempByte
		if par.MaxLen == 0 {
			par.MaxLen = 1
		}
		if par.MaxLen > connection.maxLen.raw {
			par.DataType = LongRaw
		}
	case OCIClobLocator:
		fallthrough
	case OCIBlobLocator:
		var temp *Lob
		temp, err = getLob(val, connection)
		if err != nil {
			return err
		}
		par.iPrimValue = temp
		if temp == nil {
			//if par.Direction == Input {
			//	par.DataType = NCHAR
			//}
			par.MaxLen = 1
			par.iPrimValue = nil
			par.IsNull = true
		}
	case OCIFileLocator:
		if value, ok := val.(BFile); ok {
			if value.Valid {
				if par.Direction == Input && !value.isInit() {

					return errors.New("BFile should be initialized first")
				}
				par.iPrimValue = &value
			} else {
				par.iPrimValue = nil
				par.IsNull = true
			}
		}
	case REFCURSOR:
		par.iPrimValue = nil
		par.IsNull = true
	case XMLType:
		rValue := reflect.ValueOf(val)
		pars := make([]ParameterInfo, 0, 10)
		// if value is null or value is not struct ==> pass null for the object
		if !rValue.IsValid() || rValue.Kind() != reflect.Struct || (rValue.Kind() == reflect.Ptr && rValue.IsNil()) {
			par.IsNull = true
			par.iPrimValue = nil
			return nil
		}
		for _, attrib := range par.cusType.attribs {
			attrib.Direction = par.Direction
			attrib.parent = par
			if fieldIndex, ok := par.cusType.fieldMap[attrib.Name]; ok {
				if rValue.Field(fieldIndex).CanInterface() {
					attrib.Value = rValue.Field(fieldIndex).Interface()
				}
				if attrib.cusType != nil && attrib.cusType.isArray {
					attrib.MaxNoOfArrayElements = 1
				}
				err = attrib.encodeWithType(connection)
				if err != nil {
					return err
				}
				pars = append(pars, attrib)
			}
		}
		par.iPrimValue = pars
	}
	return nil
}

func (par *ParameterInfo) encodePrimValue(conn *Connection) error {
	var err error
	switch value := par.iPrimValue.(type) {
	case nil:
		if par.DataType == XMLType && par.IsNull {
			if par.cusType.isArray {
				par.BValue = []byte{0xFF}
			} else {
				par.BValue = []byte{0xFD}
			}
			par.MaxNoOfArrayElements = 0
			par.Flag = 0x3
		} else {
			par.BValue = nil
		}
	//case float64:
	//	par.BValue, err = converters.EncodeDouble(value)
	//	if err != nil {
	//		return err
	//	}
	//case int64:
	//	par.BValue = converters.EncodeInt64(value)
	//case uint64:
	//	par.BValue = converters.EncodeUint64(value)
	case *Number:
		par.BValue = value.data
	case bool:
		par.BValue = converters.EncodeBool(value)
	case string:
		conv, err := conn.getStrConv(par.CharsetID)
		if err != nil {
			return err
		}
		par.BValue = conv.Encode(value)
		par.MaxLen = len(par.BValue)
		if par.MaxLen == 0 {
			par.MaxLen = 1
		}
	case time.Time:
		switch par.DataType {
		case DATE:
			par.BValue = converters.EncodeDate(value)
		case TIMESTAMP:
			par.BValue = converters.EncodeTimeStamp(value, false, true)
		case TimeStampTZ_DTY:
			par.BValue = converters.EncodeTimeStamp(value, true, conn.dataNego.serverTZVersion > 0 && conn.dataNego.clientTZVersion != conn.dataNego.serverTZVersion)
		}
	case *Lob:
		par.BValue = value.sourceLocator
	case *BFile:
		par.BValue = value.lob.sourceLocator
	case []byte:
		par.BValue = value
	case []ParameterInfo:
		session := conn.session
		if par.MaxNoOfArrayElements > 0 {

			if len(value) > 0 {
				arrayBuffer := bytes.Buffer{}
				if par.DataType == XMLType {
					arrayBuffer.Write([]byte{1, 3})
					if par.MaxNoOfArrayElements > 0xFC {
						session.WriteUint(&arrayBuffer, 0xFE, 2, true, false)
						session.WriteUint(&arrayBuffer, par.MaxNoOfArrayElements, 4, true, false)
					} else {
						session.WriteUint(&arrayBuffer, par.MaxNoOfArrayElements, 2, true, false)
					}
				} else {
					session.WriteUint(&arrayBuffer, par.MaxNoOfArrayElements, 4, true, true)
				}
				for _, attrib := range value {
					attrib.parent = nil
					err = attrib.encodePrimValue(conn)
					if err != nil {
						return err
					}
					if attrib.DataType == XMLType {
						session.WriteFixedClr(&arrayBuffer, attrib.BValue)
					} else {
						if attrib.IsNull && par.DataType == XMLType {
							arrayBuffer.WriteByte(0xff)
						} else {
							session.WriteClr(&arrayBuffer, attrib.BValue)
						}

					}
					if par.MaxCharLen < attrib.MaxCharLen {
						par.MaxCharLen = attrib.MaxCharLen
					}
					if par.MaxLen < attrib.MaxLen {
						par.MaxLen = attrib.MaxLen
					}
				}
				par.BValue = arrayBuffer.Bytes()
			}
			if par.DataType == NCHAR {
				par.MaxLen = conn.maxLen.nvarchar
				par.MaxCharLen = par.MaxLen // / converters.MaxBytePerChar(par.CharsetID)
			}
			if par.DataType == RAW {
				par.MaxLen = conn.maxLen.raw
			}
			if par.DataType == XMLType {
				par.BValue = encodeObject(session, par.BValue, true)
				par.MaxNoOfArrayElements = 0
				par.Flag = 3
			}
		} else {
			var objectBuffer bytes.Buffer
			for _, attrib := range value {
				err = attrib.encodePrimValue(conn)
				if err != nil {
					return err
				}
				if attrib.DataType == OCIFileLocator && attrib.MaxLen == 0 {
					attrib.MaxLen = 4000
				}
				if attrib.DataType == XMLType {
					if attrib.cusType.isArray {
						session.WriteFixedClr(&objectBuffer, attrib.BValue)
					} else {
						objectBuffer.Write(attrib.BValue)
					}
				} else {
					session.WriteClr(&objectBuffer, attrib.BValue)
				}
			}
			if par.parent == nil {
				par.BValue = encodeObject(session, objectBuffer.Bytes(), false)
			} else {
				par.BValue = objectBuffer.Bytes()
			}
		}
	default:
		return fmt.Errorf("unsupported primitive type: %v", reflect.TypeOf(par.iPrimValue).Name())
	}
	return nil
}

func (par *ParameterInfo) init() {
	par.DataType = 0
	par.Flag = 3
	par.ContFlag = 0
	par.CharsetID = 0
	par.CharsetForm = 0
	par.MaxLen = 1
	par.MaxCharLen = 0
	par.MaxNoOfArrayElements = 0
	par.BValue = nil
	par.iPrimValue = nil
	par.oPrimValue = nil
}

func (par *ParameterInfo) encodeValue(size int, connection *Connection) error {
	par.init()
	err := par.setDataType(reflect.TypeOf(par.Value), par.Value, connection)
	if err != nil {
		return err
	}
	if par.MaxNoOfArrayElements > 0 && par.MaxNoOfArrayElements < size {
		par.MaxNoOfArrayElements = size
	}
	err = par.encodeWithType(connection)
	if err != nil {
		return err
	}
	err = par.encodePrimValue(connection)
	if err != nil {
		return err
	}

	// check if the data length beyond max length for some types
	//switch par.DataType {
	//case NCHAR:
	//	if len(par.BValue) > connection.maxLen.varchar {
	//		return fmt.Errorf("passing varchar value with size: %d bigger than max size: %d", len(par.BValue), connection.maxLen.varchar)
	//	}
	//case RAW:
	//	if len(par.BValue) > connection.maxLen.raw {
	//		return fmt.Errorf("passing raw value with size: %d bigger than max size: %d", len(par.BValue), connection.maxLen.raw)
	//	}
	//}
	if par.DataType == OCIFileLocator {
		par.MaxLen = size
		if par.MaxLen == 0 {
			par.MaxLen = 4000
		}
	}
	if par.Direction != Input {
		if par.DataType == NCHAR {
			if par.MaxCharLen < size {
				par.MaxCharLen = size
			}
			conv, err := connection.getStrConv(par.CharsetID)
			if err != nil {
				return err
			}
			par.MaxLen = par.MaxCharLen * converters.MaxBytePerChar(conv.GetLangID())
		}
		if par.DataType == RAW {
			if par.MaxLen < size {
				par.MaxLen = size
			}
		}
	}

	if par.Direction == Output && !(par.DataType == XMLType) {
		par.BValue = nil
	}
	return nil
}
