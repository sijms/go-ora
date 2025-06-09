package go_ora

type OracleTypeInterface interface {
	SetDataType(conn *Connection, par *ParameterInfo) error
	// first you need to define an oracle type for your custom object
	// after that the driver has a definte pathway for each type to convert the value into binary data

	//SetType(conn *Connection, par *ParameterInfo) error
	//Encode(conn *Connection, par *ParameterInfo) error
}

type OracleTypeDecoder interface {
	Decode(data []byte) error
}

//type defaultType struct{}

//func (t *defaultType) SetDataType(conn *Connection, par *ParameterInfo, value driver.Value) error {
//	goType := reflect.TypeOf(value)
//	if goType == nil {
//		par.DataType = NCHAR
//		return nil
//	}
//	for goType.Kind() == reflect.Ptr {
//		goType = goType.Elem()
//	}
//	if tNumber(goType) || tNullNumber(goType) {
//		par.DataType = NUMBER
//		par.MaxLen = converters.MAX_LEN_NUMBER
//		return nil
//	}
//	switch goType {
//	case tyString, tyNullString:
//		par.DataType = NCHAR
//		par.CharsetForm = 1
//		par.ContFlag = 16
//		par.CharsetID = conn.getDefaultCharsetID()
//	case tyTime, tyNullTime:
//		if par.Flag&0x40 > 0 {
//			par.DataType = DATE
//			par.MaxLen = converters.MAX_LEN_DATE
//		} else {
//			par.DataType = TimeStampTZ_DTY
//			par.MaxLen = converters.MAX_LEN_TIMESTAMP
//		}
//	case tyBytes:
//		par.DataType = RAW
//	default:
//		return errors.ErrUnsupported
//	}
//	return nil
//}

//func (t *defaultType) setCollDataType(conn *Connection, par *ParameterInfo, goType reflect.Type) error {
//	if goType.Kind() == reflect.Array || goType.Kind() == reflect.Slice {
//		var inVal driver.Value = nil
//		rValue := reflect.ValueOf(par.Value)
//		size := rValue.Len()
//		if size > 0 && rValue.Index(0).CanInterface() {
//			inVal = rValue.Index(0).Interface()
//		}
//		par.Flag = 0x43
//		err := par.setDataType()
//		if err != nil {
//			return err
//		}
//		if par.DataType == XMLType {
//			found := false
//			for _, cust := range conn.cusTyp {
//				if cust.isArray && len(cust.attribs) > 0 {
//					if strings.EqualFold(par.cusType.name, cust.attribs[0].cusType.name) {
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
//	}
//}

//func (t *defaultTypeCoder) Encode(conn *Connection, par *ParameterInfo) error {
//	// entered value should be prepared (means passed inside getValue
//	// clear parameter value
//
//	par.Value = value
//	// get the type
//	_type := reflect.TypeOf(value)
//	if _type == nil {
//		par.DataType = NCHAR
//		par.Value = value
//		par.CharsetForm = 0
//		par.MaxLen = 1
//		return nil
//	}
//
//	for _type.Kind() == reflect.Ptr {
//		_type = _type.Elem()
//	}
//	if tNumber(_type) || tNullNumber(_type) {
//		par.DataType = NUMBER
//		par.MaxLen = converters.MAX_LEN_NUMBER
//		number, err := NewNumber(value)
//		if err != nil {
//			return err
//		}
//		if number == nil {
//			par.BValue = nil
//		} else {
//			par.BValue = number.data
//		}
//	} else {
//		switch _type {
//		case tyPLBool:
//			par.DataType = Boolean
//			par.MaxLen = converters.MAX_LEN_BOOL
//			temp, err := getBool(value)
//			if err != nil {
//				return err
//			}
//			if temp != nil {
//				par.BValue = converters.EncodeBool(temp)
//			}
//		case tyString, tyNullString, tyNVarChar, tyNullNVarChar:
//			par.DataType = NCHAR
//			if _type == tyString || _type == tyNullString {
//				par.CharsetForm = 1
//				par.CharsetID = conn.getDefaultCharsetID()
//			} else {
//				par.CharsetForm = 2
//				par.CharsetID = conn.tcpNego.ServernCharset
//			}
//			par.ContFlag = 16
//			conv, err := conn.getStrConv(par.CharsetID)
//			if err != nil {
//				return err
//			}
//			temp := getString(value)
//			length := len(temp)
//			par.MaxCharLen = length
//			if length > conn.maxLen.varchar {
//				par.DataType = LongVarChar
//			}
//			par.BValue = conv.Encode(temp)
//			par.MaxLen = len(par.BValue)
//			if par.MaxLen == 0 {
//				par.MaxLen = 1
//				par.IsNull = true
//			}
//		case tyTime, tyNullTime:
//			date, err := getDate(value)
//			if err != nil {
//				return err
//			}
//			if par.Flag&0x40 > 0 {
//				par.DataType = DATE
//				par.MaxLen = converters.MAX_LEN_DATE
//				par.BValue = converters.EncodeDate(date)
//			} else {
//				par.DataType = TimeStampTZ_DTY
//				par.MaxLen = converters.MAX_LEN_TIMESTAMP
//				par.BValue = converters.EncodeTimeStamp(date, true, conn.dataNego.serverTZVersion > 0 && conn.dataNego.clientTZVersion != conn.dataNego.serverTZVersion)
//			}
//		case tyTimeStamp, tyNullTimeStamp:
//			date, err := getDate(value)
//			if err != nil {
//				return err
//			}
//			par.DataType = TIMESTAMP
//			par.MaxLen = converters.MAX_LEN_TIMESTAMP
//			par.BValue = converters.EncodeTimeStamp(date, false, true)
//		case tyTimeStampTZ, tyNullTimeStampTZ:
//			date, err := getDate(value)
//			if err != nil {
//				return err
//			}
//			par.DataType = TimeStampTZ_DTY
//			par.MaxLen = converters.MAX_LEN_TIMESTAMP
//			par.BValue = converters.EncodeTimeStamp(date, true, conn.dataNego.serverTZVersion > 0 && conn.dataNego.clientTZVersion != conn.dataNego.serverTZVersion)
//		case tyBytes:
//			par.DataType = RAW
//			temp, err := getBytes(value)
//			if err != nil {
//				return err
//			}
//			par.BValue = temp
//			par.MaxLen = len(par.BValue)
//			if par.MaxLen == 0 {
//				par.MaxLen = 1
//			}
//			if par.MaxLen > conn.maxLen.raw {
//				par.DataType = LongRaw
//			}
//		case tyClob, tyNClob, tyBlob, tyVector:
//			switch _type {
//			case tyClob:
//				par.DataType = OCIClobLocator
//				par.CharsetForm = 1
//				par.CharsetID = conn.getDefaultCharsetID()
//			case tyNClob:
//				par.DataType = OCIClobLocator
//				par.CharsetForm = 2
//				par.CharsetID = conn.tcpNego.ServernCharset
//			case tyBlob:
//				par.DataType = OCIBlobLocator
//			case tyVector:
//				par.DataType = VECTOR
//			}
//			temp, err := getLob(value, conn)
//			if err != nil {
//				return err
//			}
//			if temp == nil {
//				par.MaxLen = 1
//				par.IsNull = true
//				par.BValue = nil
//			} else {
//				par.BValue = temp.sourceLocator
//			}
//		case tyBFile:
//			par.DataType = OCIFileLocator
//			if temp, ok := value.(BFile); ok {
//				if temp.Valid {
//					if par.Direction == Input && !temp.isInit() {
//						return errors.New("BFile should be initialized first")
//					}
//					par.BValue = temp.lob.sourceLocator
//				} else {
//					par.BValue = nil
//					par.IsNull = true
//				}
//			}
//
//		case tyRefCursor:
//			par.DataType = REFCURSOR
//			par.BValue = nil
//			par.IsNull = true
//		default:
//			return errors.ErrUnsupported
//		}
//	}
//	// check par.BValue for nil and fill par fields accordingly
//	if par.BValue == nil {
//		par.IsNull = true
//		par.MaxLen = 1
//
//	}
//	return nil
//}
//
////type OracleType interface {
////	sql.Scanner
////	driver.Valuer
////	OracleTypeEncoder
////	OracleTypeDecoder
////}
