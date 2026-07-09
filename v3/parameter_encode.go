package go_ora

import (
	"database/sql"
	"reflect"

	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/parameter_coder"
	oraTypes "github.com/sijms/go-ora/v3/types"
)

// func (par *ParameterInfo) enocde_(conn *Connection, value interface{}) error {
//
// }
//func (par *ParameterInfo) setDataType(conn *Connection, goType reflect.Type, data driver.Value) error {
//	if par.DataType > 0 {
//		return nil
//	}
//	// step to find the data type
//	// 1- check for nil
//	if goType == nil {
//		par.DataType = oraTypes.NCHAR
//		return nil
//	}
//	for goType.Kind() == reflect.Ptr {
//		goType = goType.Elem()
//	}
//	// 2- check for common types
//	if tNumber(goType) || tNullNumber(goType) {
//		par.DataType = oraTypes.NUMBER
//		par.MaxLen = int64(converters.MAX_LEN_NUMBER)
//		return nil
//	}
//	switch goType {
//	case tyString, tyNullString:
//		par.DataType = oraTypes.NCHAR
//		par.CharsetForm = 1
//		par.ContFlag = 16
//		par.CharsetID = conn.getDefaultCharsetID()
//		return nil
//	case tyTime, tyNullTime:
//		if par.Flag&0x40 > 0 {
//			par.DataType = oraTypes.DATE
//			par.MaxLen = int64(converters.MAX_LEN_DATE)
//		} else {
//			par.DataType = oraTypes.TimeStampTZ_DTY
//			par.MaxLen = int64(converters.MAX_LEN_TIMESTAMP)
//		}
//		return nil
//	case tyBytes:
//		par.DataType = oraTypes.RAW
//		return nil
//	}
//	// 3- call getValue
//	vData, err := getValue(data)
//	if err != nil {
//		return err
//	}
//	// 4- call setType again
//	if reflect.TypeOf(vData) != reflect.TypeOf(data) {
//		return par.setDataType(conn, reflect.TypeOf(vData), vData)
//	}
//	value := reflect.ValueOf(data)
//	if value.Kind() == reflect.Ptr && value.IsNil() {
//		data = reflect.New(goType).Interface()
//	}
//	//if temp, ok := data.(type_coder.OracleTypeInterface); ok {
//	//	fmt.Println(temp)
//	//	//err := temp.SetDataType(conn, par)
//	//	return nil
//	//}
//	switch goType.Kind() {
//	case reflect.Array, reflect.Slice:
//		var inVal driver.Value = nil
//		var err error
//		rValue := reflect.ValueOf(data)
//		size := rValue.Len()
//		if size > 0 && rValue.Index(0).CanInterface() {
//			inVal, err = getValue(rValue.Index(0).Interface())
//			if err != nil {
//				return err
//			}
//		}
//		par.Flag = 0x43
//		err = par.setDataType(conn, goType.Elem(), inVal)
//		if err != nil {
//			return err
//		}
//		if par.DataType == oraTypes.XMLType {
//			// par.cusType is for item I should get that of array
//			found := false
//			for _, cust := range conn.cusTyp {
//				if cust.isArray && len(cust.attribs) > 0 {
//					if par.cusType.name == cust.attribs[0].cusType.name {
//						found = true
//						// par.TypeName = name
//						par.ToID = cust.toid
//						*par.cusType = cust
//						par.Flag = 0x3
//						break
//					}
//				}
//			}
//			if !found {
//				return fmt.Errorf("can't get the collection of type %s", par.cusType.name)
//			}
//		}
//		par.MaxNoOfArrayElements = 1
//		return nil
//	case reflect.Struct:
//		// see if the struct is support valuer interface
//
//		for _, cusTyp := range conn.cusTyp {
//			if goType == cusTyp.typ {
//				par.cusType = new(customType)
//				*par.cusType = cusTyp
//				par.ToID = cusTyp.toid
//				// par.TypeName = cusTyp.name
//			}
//		}
//		if par.cusType == nil {
//			return errors.New("call register type before use user defined type (UDT)")
//		}
//		par.Version = 1
//		par.DataType = oraTypes.XMLType
//		par.MaxLen = 2000
//	default:
//		return fmt.Errorf("unsupported go type: %v", goType.Name())
//	}
//	//temp := defaultType{}
//	//err := temp.SetDataType(conn, par)
//	//if err != nil {
//	//	return err
//	//}
//
//	//if goType == tyObject {
//	//	val, err := getValue(value)
//	//	if err != nil {
//	//		return err
//	//	}
//	//	if obj, ok := val.(Object); ok {
//	//		par.DataType = XMLType
//	//		par.Value = obj.Value
//	//		// set custom type
//	//		for name, cusTyp := range conn.cusTyp {
//	//			if strings.EqualFold(name, obj.Name) {
//	//				par.cusType = new(customType)
//	//				*par.cusType = cusTyp
//	//				par.ToID = cusTyp.toid
//	//				if cusTyp.isArray {
//	//					par.MaxNoOfArrayElements = 1
//	//				} else {
//	//					par.Version = 1
//	//				}
//	//				break
//	//			}
//	//		}
//	//		if par.cusType == nil {
//	//			return fmt.Errorf("type %s is not created or not registered", obj.Name)
//	//		}
//	//		return nil
//	//	}
//	//}
//	//if goType != tyBytes && (goType.Kind() == reflect.Array || goType.Kind() == reflect.Slice) {
//	//val, err := getValue(value)
//	//if err != nil {
//	//	return err
//	//}
//	//var inVal driver.Value = nil
//	//if val != nil {
//	//	rValue := reflect.ValueOf(val)
//	//	size := rValue.Len()
//	//	if size > 0 && rValue.Index(0).CanInterface() {
//	//		inVal = rValue.Index(0).Interface()
//	//	}
//	//}
//	//par.Flag = 0x43
//	//err = par.setDataType(goType.Elem(), inVal, conn)
//	//if err != nil {
//	//	return err
//	//}
//	//if par.DataType == XMLType {
//	//	// par.cusType is for item I should get that of array
//	//	found := false
//	//	for _, cust := range conn.cusTyp {
//	//		if cust.isArray && len(cust.attribs) > 0 {
//	//			if par.cusType.name == cust.attribs[0].cusType.name {
//	//				found = true
//	//				// par.TypeName = name
//	//				par.ToID = cust.toid
//	//				*par.cusType = cust
//	//				par.Flag = 0x3
//	//				break
//	//			}
//	//		}
//	//	}
//	//	if !found {
//	//		return fmt.Errorf("can't get the collection of type %s", par.cusType.name)
//	//	}
//	//}
//	//par.MaxNoOfArrayElements = 1
//	//return nil
//	//}
//	//if tNumber(goType) || tNullNumber(goType) {
//	//	par.DataType = NUMBER
//	//	par.MaxLen = converters.MAX_LEN_NUMBER
//	//	return nil
//	//}
//
//	//switch goType {
//	//case tyNumber:
//	//	par.DataType = NUMBER
//	//	par.MaxLen = converters.MAX_LEN_NUMBER
//	//case tyPLBool:
//	//	par.DataType = Boolean
//	//	par.MaxLen = converters.MAX_LEN_BOOL
//	//case tyString, tyNullString:
//	//	par.DataType = NCHAR
//	//	par.CharsetForm = 1
//	//	par.ContFlag = 16
//	//	par.CharsetID = conn.getDefaultCharsetID()
//	//case tyNVarChar, tyNullNVarChar:
//	//	par.DataType = NCHAR
//	//	par.CharsetForm = 2
//	//	par.ContFlag = 16
//	//	par.CharsetID = conn.tcpNego.ServernCharset
//	//case tyTime, tyNullTime:
//	//	if par.Flag&0x40 > 0 {
//	//		par.DataType = DATE
//	//		par.MaxLen = converters.MAX_LEN_DATE
//	//	} else {
//	//		par.DataType = TimeStampTZ_DTY
//	//		par.MaxLen = converters.MAX_LEN_TIMESTAMP
//	//	}
//	//case tyTimeStamp, tyNullTimeStamp:
//	//	// if par.Flag&0x43 > 0 {
//	//	par.DataType = TIMESTAMP
//	//	par.MaxLen = converters.MAX_LEN_DATE
//	//} else {
//	//	par.DataType = TimeStampTZ_DTY
//	//	par.MaxLen = converters.MAX_LEN_TIMESTAMP
//	//}
//	//case tyTimeStampTZ, tyNullTimeStampTZ:
//	//	par.DataType = TimeStampTZ_DTY
//	//	par.MaxLen = converters.MAX_LEN_TIMESTAMP
//	// case tyTime, tyNullTime:
//	//	if par.Direction == Input {
//	//		par.DataType = TIMESTAMP
//	//		par.MaxLen = converters.MAX_LEN_TIMESTAMP
//	//	} else {
//	//		par.DataType = DATE
//	//		par.MaxLen = converters.MAX_LEN_DATE
//	//	}
//	// case tyTimeStamp, tyNullTimeStamp:
//	//	if par.Direction == Input {
//	//		par.DataType = TIMESTAMP
//	//		par.MaxLen = converters.MAX_LEN_TIMESTAMP
//	//	} else {
//	//		par.DataType = TIMESTAMP
//	//		par.MaxLen = converters.MAX_LEN_DATE
//	//	}
//	// case tyTimeStampTZ, tyNullTimeStampTZ:
//	//	par.DataType = TimeStampTZ_DTY
//	//	par.MaxLen = converters.MAX_LEN_TIMESTAMP
//	//case tyBytes:
//	//	par.DataType = RAW
//	//case tyClob:
//	//	par.DataType = OCIClobLocator
//	//	par.CharsetForm = 1
//	//	par.CharsetID = conn.getDefaultCharsetID()
//	//case tyNClob:
//	//	par.DataType = OCIClobLocator
//	//	par.CharsetForm = 2
//	//	par.CharsetID = conn.tcpNego.ServernCharset
//	//case tyBlob:
//	//	par.DataType = OCIBlobLocator
//	//case tyVector:
//	//	par.DataType = VECTOR
//	//case tyBFile:
//	//	par.DataType = OCIFileLocator
//	//case tyRefCursor:
//	//	par.DataType = REFCURSOR
//	//default:
//	//	rOriginal := reflect.ValueOf(value)
//	//	if value != nil && !(rOriginal.Kind() == reflect.Ptr && rOriginal.IsNil()) {
//	//		proVal := reflect.Indirect(rOriginal)
//	//		if valuer, ok := proVal.Interface().(driver.Valuer); ok {
//	//			val, err := valuer.Value()
//	//			if err != nil {
//	//				return err
//	//			}
//	//			if val == nil {
//	//				par.DataType = NCHAR
//	//				return nil
//	//			}
//	//			if val != value {
//	//				return par.setDataType(reflect.TypeOf(val), val, conn)
//	//			}
//	//		}
//	//	}
//
//	return nil
//}
//func (par *ParameterInfo) encodeWithType(connection *Connection) error {
//	var err error
//	var val driver.Value
//	val, err = getValue(par.Value)
//	if err != nil {
//		return err
//	}
//	if val == nil {
//		par.IsNull = true
//		par.iPrimValue = nil
//		return nil
//	}
//	// check if array
//	// if par.MaxNoOfArrayElements > 0 && par.cusType == nil {
//	if par.MaxNoOfArrayElements > 0 {
//		if !isArrayValue(val) {
//			return fmt.Errorf("parameter %s require array value", par.Name)
//		}
//		var size int
//		rValue := reflect.ValueOf(val)
//		if isArrayValue(val) {
//			size = rValue.Len()
//		}
//		if size == 0 {
//			par.IsNull = true
//			par.iPrimValue = nil
//			return nil
//		}
//		if size > par.MaxNoOfArrayElements {
//			par.MaxNoOfArrayElements = size
//		}
//		pars := make([]ParameterInfo, 0, size)
//		var tempPar ParameterInfo
//		for x := 0; x < size; x++ {
//			if par.cusType != nil && par.cusType.isArray {
//				tempPar = par.cusType.attribs[0].clone()
//			} else {
//				tempPar = par.clone()
//			}
//			if rValue.Index(x).CanInterface() {
//				tempPar.Value = rValue.Index(x).Interface()
//			}
//			err = tempPar.encodeWithType(connection)
//			if err != nil {
//				return err
//			}
//			pars = append(pars, tempPar)
//		}
//		par.iPrimValue = pars
//		return nil
//	}
//	switch par.DataType {
//	//case oraTypes.BOOLEAN:
//	//	par.iPrimValue, err = getBool(val)
//	//	if err != nil {
//	//		return err
//	//	}
//	//case oraTypes.NUMBER:
//	//	par.iPrimValue, err = NewNumber(val)
//	//	if err != nil {
//	//		return err
//	//	}
//	//case oraTypes.NCHAR:
//	//	tempString := getString(val)
//	//	length := int64(len(tempString))
//	//	par.MaxCharLen = length
//	//	par.iPrimValue = tempString
//	//	if length > connection.maxLen.varchar {
//	//		par.DataType = oraTypes.LongVarChar
//	//	}
//	//case oraTypes.DATE:
//	//	fallthrough
//	//case oraTypes.TIMESTAMP:
//	//	fallthrough
//	//case oraTypes.TimeStampTZ_DTY:
//	//	par.iPrimValue, err = getDate(val)
//	//	if err != nil {
//	//		return err
//	//	}
//	//case oraTypes.RAW:
//	//	var tempByte []byte
//	//	tempByte, err = getBytes(val)
//	//	if err != nil {
//	//		return err
//	//	}
//	//	par.MaxLen = int64(len(tempByte))
//	//	par.iPrimValue = tempByte
//	//	if par.MaxLen == 0 {
//	//		par.MaxLen = 1
//	//	}
//	//	if par.MaxLen > int64(connection.maxLen.raw) {
//	//		par.DataType = oraTypes.LongRaw
//	//	}
//	//case oraTypes.OCIClobLocator:
//	//	fallthrough
//	//case oraTypes.OCIBlobLocator:
//	//	var temp *LobStream
//	//	temp, err = getLob(val, connection)
//	//	if err != nil {
//	//		return err
//	//	}
//	//	par.iPrimValue = temp
//	//	if temp == nil {
//	//		par.MaxLen = 1
//	//		par.iPrimValue = nil
//	//		par.IsNull = true
//	//	}
//	//case oraTypes.VECTOR:
//	//	temp, err := getVector(val)
//	//	if err != nil {
//	//		return err
//	//	}
//	//	par.iPrimValue = temp
//	//	if temp == nil {
//	//		par.MaxLen = 1
//	//		par.IsNull = true
//	//	}
//	//case oraTypes.JSON:
//	//	temp, err := getJson(val)
//	//	if err != nil {
//	//		return err
//	//	}
//	//	par.iPrimValue = temp
//	//	if temp == nil {
//	//		par.MaxLen = 1
//	//		par.IsNull = true
//	//	}
//	//case oraTypes.OCIFileLocator:
//	//	if value, ok := val.(oraTypes.BFile); ok {
//	//		if value.Valid {
//	//			if par.Direction == Input && !value.IsInit() {
//	//				return errors.New("BFile should be initialized first")
//	//			}
//	//			par.iPrimValue = &value
//	//		} else {
//	//			par.iPrimValue = nil
//	//			par.IsNull = true
//	//		}
//	//	}
//	case oraTypes.REFCURSOR:
//		par.iPrimValue = nil
//		par.IsNull = true
//	case oraTypes.XMLType:
//		rValue := reflect.ValueOf(val)
//		pars := make([]ParameterInfo, 0, 10)
//		// if value is null or value is not struct ==> pass null for the object
//		if !rValue.IsValid() || rValue.Kind() != reflect.Struct || (rValue.Kind() == reflect.Ptr && rValue.IsNil()) {
//			par.IsNull = true
//			par.iPrimValue = nil
//			return nil
//		}
//		for _, attrib := range par.cusType.attribs {
//			attrib.Direction = par.Direction
//			attrib.parent = par
//			if fieldIndex, ok := par.cusType.fieldMap[attrib.Name]; ok {
//				if rValue.Field(fieldIndex).CanInterface() {
//					attrib.Value = rValue.Field(fieldIndex).Interface()
//				}
//				if attrib.cusType != nil && attrib.cusType.isArray {
//					attrib.MaxNoOfArrayElements = 1
//				}
//				err = attrib.encodeWithType(connection)
//				if err != nil {
//					return err
//				}
//				pars = append(pars, attrib)
//			}
//		}
//		par.iPrimValue = pars
//	}
//	return nil
//}
//func (par *ParameterInfo) encodePrimValue(conn *Connection) error {
//	var err error
//	switch value := par.iPrimValue.(type) {
//	case nil:
//		if par.DataType == oraTypes.XMLType && par.IsNull {
//			if par.cusType.isArray {
//				par.BValue = []byte{0xFF}
//			} else {
//				par.BValue = []byte{0xFD}
//			}
//			par.MaxNoOfArrayElements = 0
//			par.Flag = 0x3
//		} else {
//			par.BValue = nil
//		}
//	// case float64:
//	//	par.BValue, err = converters.EncodeDouble(value)
//	//	if err != nil {
//	//		return err
//	//	}
//	// case int64:
//	//	par.BValue = converters.EncodeInt64(value)
//	// case uint64:
//	//	par.BValue = converters.EncodeUint64(value)
//	case *Number:
//		par.BValue = value.data
//	//case bool:
//	//	par.BValue = converters.EncodeBool(value)
//	case string:
//		conv, err := conn.GetStringCoder(par.CharsetID, par.CharsetForm)
//		if err != nil {
//			return err
//		}
//		par.BValue = conv.Encode(value)
//		par.MaxLen = int64(len(par.BValue))
//		if par.MaxLen == 0 {
//			par.MaxLen = 1
//		}
//	//case time.Time:
//	//	switch par.DataType {
//	//	case oraTypes.DATE:
//	//		par.BValue = converters.EncodeDate(value)
//	//	case TIMESTAMP:
//	//		par.BValue = converters.EncodeTimeStamp(value, false, true)
//	//	case TimeStampTZ_DTY:
//	//		par.BValue = converters.EncodeTimeStamp(value, true, conn.dataNego.serverTZVersion > 0 && conn.dataNego.clientTZVersion != conn.dataNego.serverTZVersion)
//	//	}
//	case *LobStream:
//		par.BValue = value.sourceLocator
//	//case *Vector:
//	//	buffer := bytes.Buffer{}
//	//	conn.session.WriteUint(&buffer, len(value.lob.sourceLocator), 4, true, true)
//	//	conn.session.WriteClr(&buffer, value.lob.sourceLocator)
//	//	conn.session.WriteClr(&buffer, value.bValue)
//	//	par.BValue = buffer.Bytes()
//	//case *Json:
//	//	buffer := bytes.Buffer{}
//	//	conn.session.WriteUint(&buffer, len(value.lob.sourceLocator), 4, true, true)
//	//	//conn.session.WriteClr(&buffer, value.lob.sourceLocator)
//	//	conn.session.WriteBytes(&buffer, value.lob.sourceLocator...)
//	//	conn.session.WriteClr(&buffer, value.bValue)
//	//	par.BValue = buffer.Bytes()
//	//case *BFile:
//	//	par.BValue = value.lob.sourceLocator
//	case []byte:
//		par.BValue = value
//	case []ParameterInfo:
//		session := conn.session
//		if par.MaxNoOfArrayElements > 0 {
//
//			if len(value) > 0 {
//				arrayBuffer := bytes.Buffer{}
//				if par.DataType == oraTypes.XMLType {
//					arrayBuffer.Write([]byte{1, 3})
//					if par.MaxNoOfArrayElements > 0xFC {
//						session.WriteUint(&arrayBuffer, 0xFE, 2, true, false)
//						session.WriteUint(&arrayBuffer, par.MaxNoOfArrayElements, 4, true, false)
//					} else {
//						session.WriteUint(&arrayBuffer, par.MaxNoOfArrayElements, 2, true, false)
//					}
//				} else {
//					session.WriteUint(&arrayBuffer, par.MaxNoOfArrayElements, 4, true, true)
//				}
//				for _, attrib := range value {
//					attrib.parent = nil
//					err = attrib.encodePrimValue(conn)
//					if err != nil {
//						return err
//					}
//					if attrib.DataType == oraTypes.XMLType {
//						session.WriteFixedClr(&arrayBuffer, attrib.BValue)
//					} else {
//						if attrib.IsNull && par.DataType == oraTypes.XMLType {
//							arrayBuffer.WriteByte(0xff)
//						} else {
//							session.WriteClr(&arrayBuffer, attrib.BValue)
//						}
//					}
//					if par.MaxCharLen < attrib.MaxCharLen {
//						par.MaxCharLen = attrib.MaxCharLen
//					}
//					if par.MaxLen < attrib.MaxLen {
//						par.MaxLen = attrib.MaxLen
//					}
//				}
//				par.BValue = arrayBuffer.Bytes()
//			}
//			if par.DataType == oraTypes.NCHAR {
//				par.MaxLen = int64(conn.maxLen.nvarchar)
//				par.MaxCharLen = par.MaxLen // / converters.MaxBytePerChar(par.CharsetID)
//			}
//			if par.DataType == oraTypes.RAW {
//				par.MaxLen = conn.maxLen.raw
//			}
//			if par.DataType == oraTypes.XMLType {
//				par.BValue = encodeObject(session, par.BValue, true)
//				par.MaxNoOfArrayElements = 0
//				par.Flag = 3
//			}
//		} else {
//			var objectBuffer bytes.Buffer
//			for _, attrib := range value {
//				err = attrib.encodePrimValue(conn)
//				if err != nil {
//					return err
//				}
//				if attrib.DataType == oraTypes.OCIFileLocator && attrib.MaxLen == 0 {
//					attrib.MaxLen = 4000
//				}
//				switch attrib.DataType {
//				case oraTypes.XMLType:
//					if attrib.cusType.isArray {
//						session.WriteFixedClr(&objectBuffer, attrib.BValue)
//					} else {
//						objectBuffer.Write(attrib.BValue)
//					}
//				//case NCHAR, CHAR, LONG, LongVarChar:
//				//	session.WriteFixedClr(&objectBuffer, attrib.BValue)
//				default:
//					session.WriteFixedClr(&objectBuffer, attrib.BValue)
//					//session.WriteClr(&objectBuffer, attrib.BValue)
//				}
//			}
//			if par.parent == nil {
//				par.BValue = encodeObject(session, objectBuffer.Bytes(), false)
//			} else {
//				par.BValue = objectBuffer.Bytes()
//			}
//		}
//	default:
//		return fmt.Errorf("unsupported primitive type: %v", reflect.TypeOf(par.iPrimValue).Name())
//	}
//	return nil
//}

func (par *ParameterInfo) init() {
	par.SetDefault()
	par.Flag = 0
	par.iPrimValue = nil
	par.oPrimValue = nil
}

func (par *ParameterInfo) encodeValue(size int64, connection *Connection) error {
	par.init()
	var err error
	par.MaxLen = size
	if par.Value == nil {
		par.Value = sql.NullString{}
	}
	tempType := getType(par.Value)
	tempValue, err := getValue(par.Value)
	if err != nil {
		return err
	}

	if value, ok := tempValue.(oraTypes.Object); ok {
		par.encoder, err = connection.GetParameterCoder(value.Name)
		if err != nil {
			return err
		}
		par.Value = value.Value
		tempValue, err = getValue(par.Value)
		if err != nil {
			return err
		}
	} else {
		if par.encoder, err = connection.GetParameterCoder(tempType); err == nil {
		} else {
			connection.tracer.Printf("Error get parameter encoder for type %v: %v", tempType, err)
			if reflect.TypeOf(tempValue) != nil && tempType != reflect.TypeOf(tempValue) {
				par.encoder, err = connection.GetParameterCoder(reflect.TypeOf(tempValue))
				if err != nil {
					connection.tracer.Printf("Error get parameter encoder for type %v: %v", reflect.TypeOf(tempValue), err)
				}
			}
			if par.encoder == nil {
				if tempType.Kind() == reflect.Array || tempType.Kind() == reflect.Slice {
					// get array item and ensure it is registered in parameter coders
					itemType := tempType.Elem()
					if itemType != nil {
						for itemType.Kind() == reflect.Ptr {
							itemType = itemType.Elem()
						}
					}
					tempCoder, err := connection.GetParameterCoder(itemType)
					if err != nil {
						return err
					}
					tempCoder.Init()
					// if item is of type xml so search for slice of item in the registered parameter coders
					if tempCoder.GetParameterInfo().DataType == oraTypes.XMLType {
						par.encoder, err = connection.GetParameterCoder(reflect.SliceOf(itemType))
					} else {
						// if max length > 1 (mean it is set with size) so transfer size to array size
						if par.MaxLen > 1 {
							par.ArraySize = int(par.MaxLen)
							par.MaxLen = 1
						}
						// if item is not xml type go for arrays
						par.encoder = &parameter_coder.ArrayParameter{}
					}
				} else {
					return err
				}
			}
		}
	}
	par.encoder.SetParameterInfo(par.GetParameterInfo())
	err = par.encoder.Encode(tempValue, connection)
	if err != nil {
		return err
	}
	par.SetParameterInfo(par.encoder.GetParameterInfo())
	if par.MaxLen < size {
		par.MaxLen = size
	}

	//else {
	//
	//	switch tempType.Kind() {
	//	case reflect.Array, reflect.Slice:
	//		var inVal driver.Value = nil
	//		rValue := reflect.ValueOf(tempValue)
	//		size := rValue.Len()
	//		if size > 0 && rValue.Index(0).CanInterface() {
	//			inVal, err = getValue(rValue.Index(0).Interface())
	//			if err != nil {
	//				return err
	//			}
	//		}
	//		tempType = getType(inVal)
	//		tempValue, err = getValue(inVal)
	//		if err != nil {
	//			return err
	//		}
	//		if coder, ok = connection.goTypeCoder[tempType]; ok {
	//			par.encoder = coder.Copy()
	//			err = par.encoder.Encode(tempValue, connection)
	//			if err != nil {
	//				return err
	//			}
	//			par.SetParameterInfo(par.encoder.GetParameterInfo())
	//			if par.MaxLen < int64(size) {
	//				par.MaxLen = int64(size)
	//			}
	//			if par.DataType != oraTypes.XMLType {
	//				par.Flag = 0x43
	//			}
	//			par.ArraySize = 1
	//		} else {
	//			// return err
	//		}
	//
	//		return nil
	//	default:
	//		return fmt.Errorf("no encoder register for data type: %T", par.Value)
	//	}
	//}

	//if par.encoder != nil {
	//	err = par.encoder.Encode(connection, nil)
	//	if err != nil {
	//		return err
	//	}
	//	//par.SetParameterInfo(par.encoder.GetTypeInfo())
	//	if par.MaxLen < size {
	//		par.MaxLen = size
	//	}
	//	if par.DataType == oraTypes.NCHAR {
	//		par.MaxCharLen = par.MaxLen
	//	}
	//}
	if par.DataType == oraTypes.OCIFileLocator {
		par.MaxLen = int64(size)
		if par.MaxLen == 0 {
			par.MaxLen = oraTypes.MaxLenBFile
		}
	}
	if par.Direction == Output && !(par.DataType == oraTypes.XMLType) {
		par.BValue = nil
	}
	//if par.MaxLen < int64(size) {
	//	par.MaxLen = int64(size)
	//}
	//if par.DataType == oraTypes.NCHAR {
	//	if par.MaxCharLen < int64(size) {
	//		par.MaxCharLen = int64(size)
	//	}
	//}
	return nil

	//if par.MaxNoOfArrayElements > 0 && par.MaxNoOfArrayElements < int(size) {
	//	par.MaxNoOfArrayElements = int(size)
	//}
	//err = par.encodeWithType(connection)
	//if err != nil {
	//	return err
	//}
	//err = par.encodePrimValue(connection)
	//if err != nil {
	//	return err
	//}

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

	if par.Direction != Input {
		if par.DataType == oraTypes.NCHAR {
			if par.MaxCharLen < int64(size) {
				par.MaxCharLen = int64(size)
			}
			conv, err := connection.GetStringCoder(par.CharsetID, par.CharsetForm)
			if err != nil {
				return err
			}
			par.MaxLen = par.MaxCharLen * int64(converters.MaxBytePerChar(conv.GetLangID()))
		}
		if par.DataType == oraTypes.RAW {
			if par.MaxLen < int64(size) {
				par.MaxLen = int64(size)
			}
		}
	}

	if par.Direction == Output && !(par.DataType == oraTypes.XMLType) {
		par.BValue = nil
		// fix max size for each array item (non-xml arrays)
		//if par.MaxNoOfArrayElements > 0 {
		//	switch par.DataType {
		//	case oraTypes.NCHAR:
		//		par.MaxLen = connection.maxLen.varchar
		//		par.MaxCharLen = connection.maxLen.varchar
		//	case oraTypes.RAW:
		//		par.MaxLen = connection.maxLen.raw
		//	}
		//}
	}
	return nil
}
