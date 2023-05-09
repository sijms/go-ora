package go_ora

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"errors"
	"reflect"
	"time"

	"github.com/sijms/go-ora/v2/converters"
)

func (par *ParameterInfo) setForNull() {
	par.DataType = NCHAR
	par.BValue = nil
	par.ContFlag = 0
	par.MaxCharLen = 0
	par.MaxLen = 1
	par.CharsetForm = 1
}

func (par *ParameterInfo) setForNumber() {
	par.DataType = NUMBER
	par.ContFlag = 0
	par.MaxCharLen = 0
	par.MaxLen = converters.MAX_LEN_NUMBER
	par.CharsetForm = 0
	par.CharsetID = 0
}
func (par *ParameterInfo) setForTime() {
	par.DataType = DATE
	par.ContFlag = 0
	par.MaxLen = converters.MAX_LEN_DATE
	par.CharsetID = 0
	par.CharsetForm = 0
}
func (par *ParameterInfo) setForRefCursor() {
	par.BValue = nil
	par.MaxCharLen = 0
	par.MaxLen = 1
	par.DataType = REFCURSOR
	par.ContFlag = 0
	par.CharsetForm = 0
}
func (par *ParameterInfo) setForUDT() {
	par.Flag = 3
	par.Version = 1
	par.DataType = XMLType
	par.CharsetID = 0
	par.CharsetForm = 0
	par.MaxLen = 2000
}

func (par *ParameterInfo) encodeInt(value int64) {
	par.setForNumber()
	par.BValue = converters.EncodeInt64(value)
}

func (par *ParameterInfo) encodeFloat(value float64) error {
	par.setForNumber()
	var err error
	par.BValue, err = converters.EncodeDouble(value)
	return err
}

func (par *ParameterInfo) encodeString(value string, converter converters.IStringConverter, size int) {
	par.DataType = NCHAR
	par.ContFlag = 16
	par.MaxCharLen = len([]rune(value))
	if len(value) == 0 {
		par.BValue = nil
	} else {
		par.BValue = converter.Encode(value)
	}
	if size > len(value) {
		par.MaxCharLen = size
	}
	if par.Direction == Input {
		if par.BValue == nil {
			par.MaxLen = 1
		} else {
			par.MaxLen = len(par.BValue)
			par.MaxCharLen = par.MaxLen
		}
	} else {
		par.MaxLen = par.MaxCharLen * converters.MaxBytePerChar(par.CharsetID)
	}
}

func (par *ParameterInfo) encodeTime(value time.Time) {
	par.setForTime()
	par.BValue = converters.EncodeDate(value)
}

func (par *ParameterInfo) encodeTimeStampTZ(value TimeStampTZ, conn *Connection) {
	par.setForTime()
	par.DataType = TimeStampTZ_DTY
	par.MaxLen = converters.MAX_LEN_TIMESTAMP
	temp := converters.EncodeTimeStamp(time.Time(value), true)
	if conn.dataNego.clientTZVersion != conn.dataNego.serverTZVersion {
		if temp[11]&0x80 != 0 {
			temp[12] |= 1
			if time.Time(value).IsDST() {
				temp[12] |= 2
			}
		} else {
			temp[11] |= 0x40
		}
	}
	par.BValue = temp
}
func (par *ParameterInfo) encodeTimeStamp(value TimeStamp) {
	par.setForTime()
	par.DataType = TIMESTAMP
	par.BValue = converters.EncodeTimeStamp(time.Time(value), false)
}

func (par *ParameterInfo) encodeRaw(value []byte, size int) {
	par.BValue = value
	par.DataType = RAW
	par.MaxLen = len(value)
	if size > par.MaxLen {
		par.MaxLen = size
	}
	if par.MaxLen == 0 {
		par.MaxLen = 1
	}
	par.ContFlag = 0
	par.MaxCharLen = 0
	par.CharsetForm = 0
	par.CharsetID = 0
}

func (par *ParameterInfo) encodeValue(val driver.Value, size int, connection *Connection) error {
	var err error
	par.Value = val
	if val == nil {
		par.setForNull()
		return nil
	}
	// put common values
	par.Flag = 3
	par.CharsetID = connection.tcpNego.ServerCharset
	par.CharsetForm = 1
	par.BValue = nil

	tempType := reflect.TypeOf(val)
	if tempType.Kind() == reflect.Ptr {
		if reflect.ValueOf(val).IsNil() && par.Direction == Input {
			par.setForNull()
			return nil
		}
		tempType = tempType.Elem()
	}
	if tempType != reflect.TypeOf([]byte{}) {
		if tempType.Kind() == reflect.Array || tempType.Kind() == reflect.Slice {
			par.Flag = 0x43

			par.MaxNoOfArrayElements = reflect.Indirect(reflect.ValueOf(val)).Len()
			if par.MaxNoOfArrayElements == 0 {
				par.MaxNoOfArrayElements = 1
			}
			return par.encodeArrayValue(val, size, connection)
		}
	}

	if temp, ok := val.(driver.Valuer); ok {
		if temp == nil || (reflect.ValueOf(temp).Kind() == reflect.Ptr && reflect.ValueOf(temp).IsNil()) {
			// bypass nil pointer
		} else {
			tempVal, err := temp.Value()
			if err != nil {
				return err
			}
			if tempVal == nil {
				switch val.(type) {
				case sql.NullInt32:
					par.setForNumber()
				case sql.NullBool:
					par.setForNumber()
				case sql.NullTime:
					par.setForTime()
				case sql.NullByte:
					par.setForNumber()
				case sql.NullFloat64:
					par.setForNumber()
				case sql.NullInt16:
					par.setForNumber()
				case sql.NullInt64:
					par.setForNumber()
				case sql.NullString:
					par.encodeString("", nil, size)
				case NullNVarChar:
					par.CharsetForm = 2
					par.CharsetID = connection.tcpNego.ServernCharset
					par.encodeString("", nil, size)
				case NullTimeStamp:
					par.setForTime()
					par.DataType = TIMESTAMP
				case NullTimeStampTZ:
					par.setForTime()
					par.DataType = TimeStampTZ_DTY
					par.MaxLen = converters.MAX_LEN_TIMESTAMP
				case *sql.NullInt32:
					par.setForNumber()
				case *sql.NullBool:
					par.setForNumber()
				case *sql.NullTime:
					par.setForTime()
				case *sql.NullByte:
					par.setForNumber()
				case *sql.NullFloat64:
					par.setForNumber()
				case *sql.NullInt16:
					par.setForNumber()
				case *sql.NullInt64:
					par.setForNumber()
				case *sql.NullString:
					par.encodeString("", nil, size)
				case *NullNVarChar:
					par.CharsetForm = 2
					par.CharsetID = connection.tcpNego.ServernCharset
					par.encodeString("", nil, size)
				case *NullTimeStamp:
					par.setForTime()
					par.DataType = TIMESTAMP
				case *NullTimeStampTZ:
					par.setForTime()
					par.DataType = TimeStampTZ_DTY
					par.MaxLen = converters.MAX_LEN_TIMESTAMP
				default:
					par.encodeString("", nil, size)
				}
				return nil
			} else {
				val = tempVal
			}
		}
	}
	switch value := val.(type) {
	case bool:
		if value {
			par.encodeInt(1)
		} else {
			par.encodeInt(0)
		}
	case int:
		par.encodeInt(int64(value))
	case int8:
		par.encodeInt(int64(value))
	case int16:
		par.encodeInt(int64(value))
	case int32:
		par.encodeInt(int64(value))
	case int64:
		par.encodeInt(value)
	case *bool:
		if value == nil {
			par.setForNumber()
		} else {
			if *value {
				par.encodeInt(1)
			} else {
				par.encodeInt(0)
			}
		}

	case *int:
		if value == nil {
			par.setForNumber()
		} else {
			par.encodeInt(int64(*value))
		}
	case *int8:
		if value == nil {
			par.setForNumber()
		} else {
			par.encodeInt(int64(*value))
		}
	case *int16:
		if value == nil {
			par.setForNumber()
		} else {
			par.encodeInt(int64(*value))
		}
	case *int32:
		if value == nil {
			par.setForNumber()
		} else {
			par.encodeInt(int64(*value))
		}
	case *int64:
		if value == nil {
			par.setForNumber()
		} else {
			par.encodeInt(*value)
		}
	case uint:
		par.encodeInt(int64(value))
	case uint8:
		par.encodeInt(int64(value))
	case uint16:
		par.encodeInt(int64(value))
	case uint32:
		par.encodeInt(int64(value))
	case uint64:
		par.encodeInt(int64(value))
	case *uint:
		if value == nil {
			par.setForNumber()
		} else {
			par.encodeInt(int64(*value))
		}
	case *uint8:
		if value == nil {
			par.setForNumber()
		} else {
			par.encodeInt(int64(*value))
		}
	case *uint16:
		if value == nil {
			par.setForNumber()
		} else {
			par.encodeInt(int64(*value))
		}
	case *uint32:
		if value == nil {
			par.setForNumber()
		} else {
			par.encodeInt(int64(*value))
		}
	case *uint64:
		if value == nil {
			par.setForNumber()
		} else {
			par.encodeInt(int64(*value))
		}
	case float32:
		err = par.encodeFloat(float64(value))
		if err != nil {
			return err
		}
	case float64:
		err = par.encodeFloat(value)
		if err != nil {
			return err
		}
	case *float32:
		if value == nil {
			par.setForNumber()
		} else {
			err = par.encodeFloat(float64(*value))
			if err != nil {
				return err
			}
		}
	case *float64:
		if value == nil {
			par.setForNumber()
		} else {
			err = par.encodeFloat(*value)
			if err != nil {
				return err
			}
		}
	case time.Time:
		par.encodeTime(value)
	case *time.Time:
		par.encodeTime(*value)
	case TimeStamp:
		par.encodeTimeStamp(value)
	case *TimeStamp:
		if value == nil {
			par.setForTime()
			par.DataType = TIMESTAMP
		} else {
			par.encodeTimeStamp(*value)
		}
	case TimeStampTZ:
		par.encodeTimeStampTZ(value, connection)
	case *TimeStampTZ:
		if value == nil {
			par.setForTime()
			par.MaxLen = converters.MAX_LEN_TIMESTAMP
			par.DataType = TimeStampTZ_DTY
		} else {
			par.encodeTimeStampTZ(*value, connection)
		}
	case NClob:
		par.CharsetForm = 2
		par.CharsetID = connection.tcpNego.ServernCharset
		strConv, _ := connection.getStrConv(par.CharsetID)
		par.encodeString(value.String, strConv, size)
		if par.Direction == Output {
			par.DataType = OCIClobLocator
		} else {
			if par.MaxLen >= connection.maxLen.nvarchar {
				par.DataType = OCIClobLocator
				lob := newLob(connection)
				err = lob.createTemporaryClob(connection.tcpNego.ServernCharset, 2)
				if err != nil {
					return err
				}
				err = lob.putString(value.String, connection.tcpNego.ServernCharset)
				if err != nil {
					return err
				}
				value.locator = lob.sourceLocator
				par.BValue = lob.sourceLocator
				par.Value = value
			}
		}
	case *NClob:
		par.CharsetForm = 2
		par.CharsetID = connection.tcpNego.ServernCharset
		strConv, _ := connection.getStrConv(par.CharsetID)
		par.encodeString(value.String, strConv, size)
		if par.Direction == Output {
			par.DataType = OCIClobLocator
		} else {
			if par.MaxLen >= connection.maxLen.nvarchar {
				par.DataType = OCIClobLocator
				lob := newLob(connection)
				err = lob.createTemporaryClob(connection.tcpNego.ServernCharset, 2)
				if err != nil {
					return err
				}
				err = lob.putString(value.String, connection.tcpNego.ServernCharset)
				if err != nil {
					return err
				}
				value.locator = lob.sourceLocator
				par.BValue = lob.sourceLocator
			}
		}
	case Clob:
		strConv, _ := connection.getStrConv(par.CharsetID)
		par.encodeString(value.String, strConv, size)
		if par.Direction == Output {
			par.DataType = OCIClobLocator
		} else {
			if par.MaxLen >= connection.maxLen.varchar {
				// here we need to use clob
				par.DataType = OCIClobLocator
				lob := newLob(connection)
				err = lob.createTemporaryClob(connection.tcpNego.ServerCharset, 1)
				if err != nil {
					return err
				}
				err = lob.putString(value.String, connection.tcpNego.ServerCharset)
				if err != nil {
					return err
				}
				value.locator = lob.sourceLocator
				par.BValue = lob.sourceLocator
				par.Value = value
			}
		}
	case *Clob:
		if value == nil {
			par.encodeString("", nil, size)
		} else {
			strConv, _ := connection.getStrConv(par.CharsetID)
			par.encodeString(value.String, strConv, size)
		}
		if par.Direction == Output {
			par.DataType = OCIClobLocator
		} else {
			if par.MaxLen >= connection.maxLen.varchar {
				par.DataType = OCIClobLocator
				lob := newLob(connection)
				err = lob.createTemporaryClob(connection.tcpNego.ServerCharset, 1)
				if err != nil {
					return err
				}
				err = lob.putString(value.String, connection.tcpNego.ServerCharset)
				if err != nil {
					return err
				}
				value.locator = lob.sourceLocator
				par.BValue = lob.sourceLocator
			}
		}
	case BFile:
		par.encodeRaw(nil, size)
		if par.MaxLen == 0 {
			par.MaxLen = 4000
		}
		par.DataType = OCIFileLocator
		if par.Direction == Input {
			if !value.isInit() {
				return errors.New("BFile must be initialized")
			}
			par.BValue = value.lob.sourceLocator
		}
	case *BFile:
		par.encodeRaw(nil, size)
		if par.MaxLen == 0 {
			par.MaxLen = 4000
		}
		par.DataType = OCIFileLocator
		if par.Direction == Input {
			if !value.isInit() {
				return errors.New("BFile must be initialized")
			}
			par.BValue = value.lob.sourceLocator
		}
	case Blob:
		par.encodeRaw(value.Data, size)
		if par.MaxLen == 0 {
			par.MaxLen = 1
		}
		if par.Direction == Output {
			par.DataType = OCIBlobLocator
		} else {
			if len(value.Data) >= connection.maxLen.raw {
				par.DataType = OCIBlobLocator
				lob := newLob(connection)
				err = lob.createTemporaryBLOB()
				if err != nil {
					return err
				}
				err = lob.putData(value.Data)
				if err != nil {
					return err
				}
				value.locator = lob.sourceLocator
				par.BValue = lob.sourceLocator
				par.Value = value
			}
		}
	case *Blob:
		if value == nil {
			par.encodeRaw(nil, size)
		} else {
			par.encodeRaw(value.Data, size)
		}
		if par.MaxLen == 0 {
			par.MaxLen = 1
		}
		if par.Direction == Output {
			par.DataType = OCIBlobLocator
		} else {
			if len(value.Data) >= connection.maxLen.raw {
				par.DataType = OCIBlobLocator
				lob := newLob(connection)
				err = lob.createTemporaryBLOB()
				if err != nil {
					return err
				}
				err = lob.putData(value.Data)
				if err != nil {
					return err
				}
				value.locator = lob.sourceLocator
				par.BValue = lob.sourceLocator
			}
		}
	case []byte:
		if len(value) == 0 {
			par.encodeRaw(nil, size)
		} else {
			if len(value) > connection.maxLen.raw && par.Direction == Input {
				return par.encodeValue(Blob{Valid: true, Data: value}, size, connection)
			}
			par.encodeRaw(value, size)
		}
	case *[]byte:
		if *value == nil || len(*value) == 0 {
			par.encodeRaw(nil, size)
		} else {
			if len(*value) > connection.maxLen.raw && par.Direction == Input {
				return par.encodeValue(&Blob{Valid: true, Data: *value}, size, connection)
			}
			par.encodeRaw(*value, size)
		}
	case RefCursor, *RefCursor:
		par.setForRefCursor()
	case string:
		if len(value) > connection.maxLen.nvarchar && par.Direction == Input {
			return par.encodeValue(Clob{Valid: true, String: value}, size, connection)
		}
		strConv, _ := connection.getStrConv(par.CharsetID)
		par.encodeString(value, strConv, size)

	case *string:
		if value == nil {
			par.encodeString("", nil, size)
		} else {
			if len(*value) > connection.maxLen.nvarchar && par.Direction == Input {
				return par.encodeValue(&Clob{Valid: true, String: *value}, size, connection)
			}
			strConv, _ := connection.getStrConv(par.CharsetID)
			par.encodeString(*value, strConv, size)
		}
	case NVarChar:
		par.CharsetForm = 2
		par.CharsetID = connection.tcpNego.ServernCharset
		strConv, _ := connection.getStrConv(par.CharsetID)
		par.encodeString(string(value), strConv, size)
	case *NVarChar:
		par.CharsetForm = 2
		par.CharsetID = connection.tcpNego.ServernCharset
		if value == nil {
			par.encodeString("", nil, size)
		} else {
			strConv, _ := connection.getStrConv(par.CharsetID)
			par.encodeString(string(*value), strConv, size)
		}
	case *sql.NullBool:
		par.setForNumber()
	case *sql.NullByte:
		par.setForNumber()
	case *sql.NullInt16:
		par.setForNumber()
	case *sql.NullInt32:
		par.setForNumber()
	case *sql.NullInt64:
		par.setForNumber()
	case *sql.NullTime:
		par.setForTime()
	case *sql.NullFloat64:
		par.setForNumber()
	case *sql.NullString:
		par.encodeString("", nil, size)
	case *NullNVarChar:
		par.encodeString("", nil, size)
		par.CharsetForm = 2
		par.CharsetID = connection.tcpNego.ServernCharset
	case *NullTimeStamp:
		par.setForTime()
		par.DataType = TIMESTAMP
	case *NullTimeStampTZ:
		par.setForTime()
		par.DataType = TimeStampTZ_DTY
		par.MaxLen = converters.MAX_LEN_TIMESTAMP
	default:
		custVal := reflect.ValueOf(val)
		if custVal.Kind() == reflect.Ptr {
			custVal = custVal.Elem()
		}
		if custVal.Kind() == reflect.Struct {
			par.setForUDT()
			for _, cusTyp := range connection.cusTyp {
				if custVal.Type() == cusTyp.typ {
					par.cusType = &cusTyp
					par.ToID = cusTyp.toid
				}
			}
			if par.cusType == nil {
				return errors.New("struct parameter only allowed with user defined type (UDT)")
			}
			var objectBuffer bytes.Buffer
			for _, attrib := range par.cusType.attribs {
				if fieldIndex, ok := par.cusType.filedMap[attrib.Name]; ok {
					tempPar := ParameterInfo{
						Direction:   par.Direction,
						Flag:        3,
						CharsetID:   connection.tcpNego.ServerCharset,
						CharsetForm: 1,
					}
					err = tempPar.encodeValue(custVal.Field(fieldIndex).Interface(), 0, connection)
					if err != nil {
						return err
					}
					connection.session.WriteClr(&objectBuffer, tempPar.BValue)
				}
			}
			par.BValue = objectBuffer.Bytes()
		}
	}
	return nil
}
