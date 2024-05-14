package go_ora

import (
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/sijms/go-ora/v2/configurations"
	"github.com/sijms/go-ora/v2/converters"
	"github.com/sijms/go-ora/v2/network"
)

type TNSType int
type ParameterDirection int

// func (n *NVarChar) ConvertValue(v interface{}) (driver.Value, error) {
//	return driver.Value(string(*n)), nil
// }

const (
	Input  ParameterDirection = 1
	Output ParameterDirection = 2
	InOut  ParameterDirection = 3
	//RetVal ParameterDirection = 9
)

type Out struct {
	Dest driver.Value
	Size int
	In   bool
}

// internal enum BindDirection
// {
// Output = 16,
// Input = 32,
// InputOutput = 48,
// }

//go:generate stringer -type=TNSType

const (
	NCHAR                     TNSType = 1
	NUMBER                    TNSType = 2
	BInteger                  TNSType = 3
	FLOAT                     TNSType = 4
	NullStr                   TNSType = 5
	VarNum                    TNSType = 6
	PDN                       TNSType = 7
	LONG                      TNSType = 8
	VARCHAR                   TNSType = 9
	ROWID                     TNSType = 11
	DATE                      TNSType = 12
	VarRaw                    TNSType = 15
	BFloat                    TNSType = 21
	BDouble                   TNSType = 22
	RAW                       TNSType = 23
	LongRaw                   TNSType = 24
	TNS_JSON_TYPE_DATE        TNSType = 60
	TNS_JSON_TYPE_INTERVAL_YM TNSType = 61
	TNS_JSON_TYPE_INTERVAL_DS TNSType = 62
	UINT                      TNSType = 68
	LongVarChar               TNSType = 94
	LongVarRaw                TNSType = 95
	CHAR                      TNSType = 96
	CHARZ                     TNSType = 97
	IBFloat                   TNSType = 100
	IBDouble                  TNSType = 101
	REFCURSOR                 TNSType = 102
	OCIXMLType                TNSType = 108
	XMLType                   TNSType = 109
	OCIRef                    TNSType = 110
	OCIClobLocator            TNSType = 112
	OCIBlobLocator            TNSType = 113
	OCIFileLocator            TNSType = 114
	ResultSet                 TNSType = 116
	JSON                      TNSType = 119
	TNS_DATA_TYPE_OAC122      TNSType = 120
	OCIString                 TNSType = 155
	OCIDate                   TNSType = 156
	TimeStampDTY              TNSType = 180
	TimeStampTZ_DTY           TNSType = 181
	IntervalYM_DTY            TNSType = 182
	IntervalDS_DTY            TNSType = 183
	TimeTZ                    TNSType = 186
	TIMESTAMP                 TNSType = 187
	TIMESTAMPTZ               TNSType = 188
	IntervalYM                TNSType = 189
	IntervalDS                TNSType = 190
	UROWID                    TNSType = 208
	TimeStampLTZ_DTY          TNSType = 231
	TimeStampeLTZ             TNSType = 232
	Boolean                   TNSType = 0xFC
)

//type ParameterType int

//const (
//	Number ParameterType = 1
//	String ParameterType = 2
//)

type ParameterInfo struct {
	Name                 string
	TypeName             string
	SchemaName           string
	DomainSchema         string
	DomainName           string
	Direction            ParameterDirection
	IsNull               bool
	AllowNull            bool
	IsJson               bool
	ColAlias             string
	DataType             TNSType
	IsXmlType            bool
	Flag                 uint8
	Precision            uint8
	Scale                uint8
	MaxLen               int
	MaxCharLen           int
	MaxNoOfArrayElements int
	ContFlag             int
	ToID                 []byte
	Version              int
	CharsetID            int
	CharsetForm          int
	BValue               []byte
	Value                driver.Value
	iPrimValue           driver.Value
	oPrimValue           driver.Value
	OutputVarPtr         interface{}
	getDataFromServer    bool
	oaccollid            int
	cusType              *customType
	parent               *ParameterInfo
	Annotations          map[string]string
}

// load get parameter information form network session
func (par *ParameterInfo) load(conn *Connection) error {
	session := conn.session
	par.getDataFromServer = true
	dataType, err := session.GetByte()
	if err != nil {
		return err
	}
	par.DataType = TNSType(dataType)
	par.Flag, err = session.GetByte()
	if err != nil {
		return err
	}
	par.Precision, err = session.GetByte()
	// precision, err := session.GetInt(1, false, false)
	// var scale int
	switch par.DataType {
	case NUMBER:
		fallthrough
	case TimeStampDTY:
		fallthrough
	case TimeStampTZ_DTY:
		fallthrough
	case IntervalDS_DTY:
		fallthrough
	case TIMESTAMP:
		fallthrough
	case TIMESTAMPTZ:
		fallthrough
	case IntervalDS:
		fallthrough
	case TimeStampLTZ_DTY:
		fallthrough
	case TimeStampeLTZ:
		if scale, err := session.GetInt(2, true, true); err != nil {
			return err
		} else {
			if scale == -127 {
				par.Precision = uint8(math.Ceil(float64(par.Precision) * 0.30103))
				par.Scale = 0xFF
			} else {
				par.Scale = uint8(scale)
			}
		}
	default:
		par.Scale, err = session.GetByte()
		// scale, err = session.GetInt(1, false, false)
	}
	// if par.Scale == uint8(-127) {
	//
	// }
	if par.DataType == NUMBER && par.Precision == 0 && (par.Scale == 0 || par.Scale == 0xFF) {
		par.Precision = 38
		par.Scale = 0xFF
	}

	// par.Scale = uint16(scale)
	// par.Precision = uint16(precision)
	par.MaxLen, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	switch par.DataType {
	case ROWID:
		par.MaxLen = 128
	case DATE:
		par.MaxLen = converters.MAX_LEN_DATE
	case IBFloat:
		par.MaxLen = 4
	case IBDouble:
		par.MaxLen = 8
	case TimeStampTZ_DTY:
		par.MaxLen = converters.MAX_LEN_TIMESTAMP
	case IntervalYM_DTY:
		fallthrough
	case IntervalDS_DTY:
		fallthrough
	case IntervalYM:
		fallthrough
	case IntervalDS:
		par.MaxLen = 11
	}
	par.MaxNoOfArrayElements, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	if session.TTCVersion >= 10 {
		par.ContFlag, err = session.GetInt(8, true, true)
	} else {
		par.ContFlag, err = session.GetInt(4, true, true)
	}
	if err != nil {
		return err
	}
	par.ToID, err = session.GetDlc()
	par.Version, err = session.GetInt(2, true, true)
	if err != nil {
		return err
	}
	par.CharsetID, err = session.GetInt(2, true, true)
	if err != nil {
		return err
	}
	par.CharsetForm, err = session.GetInt(1, false, false)
	if err != nil {
		return err
	}
	par.MaxCharLen, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	if session.TTCVersion >= 8 {
		par.oaccollid, err = session.GetInt(4, true, true)
	}
	num1, err := session.GetInt(1, false, false)
	if err != nil {
		return err
	}
	par.AllowNull = num1 > 0
	_, err = session.GetByte() //  v7 length of name
	if err != nil {
		return err
	}
	bName, err := session.GetDlc()
	if err != nil {
		return err
	}
	par.Name = session.StrConv.Decode(bName)
	bName, err = session.GetDlc() // schema name
	if err != nil {
		return err
	}
	par.SchemaName = strings.ToUpper(session.StrConv.Decode(bName))
	bName, err = session.GetDlc()
	if err != nil {
		return err
	}
	par.TypeName = strings.ToUpper(session.StrConv.Decode(bName))
	if par.DataType == XMLType && par.TypeName != "XMLTYPE" {
		for typName, cusTyp := range conn.cusTyp {
			if typName == par.TypeName {
				par.cusType = new(customType)
				*par.cusType = cusTyp
			}
		}
	}
	if par.TypeName == "XMLTYPE" {
		par.DataType = XMLType
		par.IsXmlType = true
	}
	if session.TTCVersion < 3 {
		return nil
	}
	_, err = session.GetInt(2, true, true)
	if session.TTCVersion < 6 {
		return nil
	}
	var uds_flags int
	uds_flags, err = session.GetInt(4, true, true)
	par.IsJson = (uds_flags & 0x500) > 0
	if session.TTCVersion < 17 {
		return nil
	}
	bName, err = session.GetDlc()
	if err != nil {
		return err
	}
	par.DomainSchema = strings.ToUpper(session.StrConv.Decode(bName))
	bName, err = session.GetDlc()
	if err != nil {
		return err
	}
	par.DomainName = strings.ToUpper(session.StrConv.Decode(bName))
	if session.TTCVersion < 20 {
		return nil
	}
	numAnnotations, err := session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	if numAnnotations > 0 {
		par.Annotations = make(map[string]string)
		_, err = session.GetByte()
		if err != nil {
			return err
		}
		numAnnotations, err = session.GetInt(4, true, true)
		if err != nil {
			return err
		}
		_, err = session.GetByte()
		if err != nil {
			return err
		}
		for i := 0; i < numAnnotations; i++ {
			bKey, bValue, _, err := session.GetKeyVal()
			if err != nil {
				return err
			}
			key := session.StrConv.Decode(bKey)
			value := session.StrConv.Decode(bValue)
			par.Annotations[key] = value
		}
		_, err = session.GetInt(4, true, true)
		if err != nil {
			return err
		}
	}
	return nil
}

// write parameter information to network session
func (par *ParameterInfo) write(session *network.Session) error {
	session.PutBytes(uint8(par.DataType), par.Flag, par.Precision, par.Scale)
	session.PutUint(par.MaxLen, 4, true, true)
	// MaxNoOfArrayElements should be 0 in case of XML type
	session.PutInt(par.MaxNoOfArrayElements, 4, true, true)
	if session.TTCVersion >= 10 {
		session.PutInt(par.ContFlag, 8, true, true)
	} else {
		session.PutInt(par.ContFlag, 4, true, true)
	}
	if par.ToID == nil {
		session.PutBytes(0)
		// session.PutInt(0, 1, false, false)
	} else {
		session.PutInt(len(par.ToID), 4, true, true)
		session.PutClr(par.ToID)
	}
	session.PutUint(par.Version, 2, true, true)
	session.PutUint(par.CharsetID, 2, true, true)
	session.PutBytes(uint8(par.CharsetForm))
	// session.PutUint(par.CharsetForm, 1, false, false)
	session.PutUint(par.MaxCharLen, 4, true, true)
	if session.TTCVersion >= 8 {
		session.PutInt(par.oaccollid, 4, true, true)
	}
	return nil
}

func (par *ParameterInfo) clone() ParameterInfo {
	tempPar := ParameterInfo{}
	tempPar.DataType = par.DataType
	tempPar.cusType = par.cusType
	tempPar.TypeName = par.TypeName
	tempPar.MaxLen = par.MaxLen
	tempPar.MaxCharLen = par.MaxCharLen
	tempPar.CharsetID = par.CharsetID
	tempPar.CharsetForm = par.CharsetForm
	tempPar.Scale = par.Scale
	tempPar.Precision = par.Precision
	return tempPar
}

func (par *ParameterInfo) collectLocators() [][]byte {
	switch value := par.iPrimValue.(type) {
	case *Lob:
		if value != nil && value.sourceLocator != nil {
			return [][]byte{value.sourceLocator}
		}
	case *BFile:
		if value != nil && value.lob.sourceLocator != nil {
			return [][]byte{value.lob.sourceLocator}
		}
	case []ParameterInfo:
		output := make([][]byte, 0, 10)
		for _, temp := range value {
			output = append(output, temp.collectLocators()...)
		}
		return output
	}
	return [][]byte{}
}

func (par *ParameterInfo) isLongType() bool {
	return par.DataType == LONG || par.DataType == LongRaw || par.DataType == LongVarChar || par.DataType == LongVarRaw
}

func (par *ParameterInfo) isLobType() bool {
	return par.DataType == OCIBlobLocator || par.DataType == OCIClobLocator || par.DataType == OCIFileLocator
}

func (par *ParameterInfo) decodePrimValue(conn *Connection, temporaryLobs *[][]byte, udt bool) error {
	session := conn.session
	var err error
	par.oPrimValue = nil
	par.BValue = nil
	if par.MaxNoOfArrayElements > 0 {
		size, err := session.GetInt(4, true, true)
		if err != nil {
			return err
		}
		if size > 0 {
			par.MaxNoOfArrayElements = size
			pars := make([]ParameterInfo, 0, size)
			for x := 0; x < size; x++ {
				tempPar := par.clone()
				err = tempPar.decodeParameterValue(conn, temporaryLobs)
				if err != nil {
					return err
				}
				//, err = tempPar.decodeValue(stmt.connection, false)
				//if x < size-1 {
				_, err = session.GetInt(2, true, true)
				if err != nil {
					return err
				}
				//}
				pars = append(pars, tempPar)
			}
			par.oPrimValue = pars
		}
		return nil
	}
	if par.DataType == XMLType && par.parent == nil {
		if par.TypeName == "XMLTYPE" {
			return errors.New("unsupported data type: XMLTYPE")
		}
		if par.cusType == nil {
			return fmt.Errorf("unregister custom type: %s. call RegisterType first", par.TypeName)
		}
		_, err = session.GetDlc() // contain toid and some 0s
		if err != nil {
			return err
		}
		_, err = session.GetBytes(3) // 3 0s
		if err != nil {
			return err
		}
		var size int
		size, err = session.GetInt(4, true, true)
		if err != nil {
			return err
		}
		_, err = session.GetBytes(2) // 0x1 0x1
		if err != nil {
			return err
		}
		if size == 0 {
			// the object is null
			_, err = session.GetBytes(2) // 0x81 0x01
			if err != nil {
				return err
			}
			par.oPrimValue = nil
			par.IsNull = true
			return nil
		} else {
			par.IsNull = false
		}
	}
	if par.DataType == ROWID {
		rowid, err := newRowID(session)
		if err != nil {
			return err
		}
		if rowid != nil {
			par.oPrimValue = string(rowid.getBytes())
		}
		return nil
	}
	if par.DataType == UROWID {
		rowid, err := newURowID(session)
		if err != nil {
			return err
		}
		if rowid != nil {
			par.oPrimValue = string(rowid.getBytes())
		}
		return nil
	}
	if (par.DataType == NCHAR || par.DataType == CHAR) && par.MaxCharLen == 0 {
		return nil
	}
	if par.DataType == RAW && par.MaxLen == 0 {
		return nil
	}
	par.BValue, err = session.GetClr()
	if err != nil {
		return err
	}
	if par.BValue == nil {
		return nil
	}
	//}
	switch par.DataType {
	case NCHAR, CHAR, LONG, LongVarChar:
		strConv, err := conn.getStrConv(par.CharsetID)
		if err != nil {
			return err
		}
		par.oPrimValue = strConv.Decode(par.BValue)
	case Boolean:
		par.oPrimValue = converters.DecodeBool(par.BValue)
	case RAW, LongRaw:
		par.oPrimValue = par.BValue
	case NUMBER:
		num := Number{data: par.BValue}
		//strNum, err := num.String()
		//if err != nil {
		//	return err
		//}
		//if strings.Contains(strNum, ".") {
		//	par.oPrimValue, err = num.Float64()
		//} else {
		//	par.oPrimValue, err = num.Uint64()
		//}
		par.oPrimValue, err = num.String()
		if err != nil {
			return err
		}
		//if par.Scale == 0 && par.Precision == 0 {
		//	var tempFloat string
		//	tempFloat, err = converters.NumberToString(par.BValue)
		//	if err != nil {
		//		return err
		//	}
		//	if strings.Contains(tempFloat, ".") {
		//		par.oPrimValue, err = strconv.ParseFloat(tempFloat, 64)
		//	} else {
		//		par.oPrimValue, err = strconv.ParseInt(tempFloat, 10, 64)
		//	}
		//} else if par.Scale == 0 && par.Precision <= 18 {
		//	par.oPrimValue, err = converters.NumberToInt64(par.BValue)
		//	if err != nil {
		//		return err
		//	}
		//} else if par.Scale == 0 && (converters.CompareBytes(par.BValue, converters.Int64MaxByte) > 0 &&
		//	converters.CompareBytes(par.BValue, converters.Uint64MaxByte) < 0) {
		//	par.oPrimValue, err = converters.NumberToUInt64(par.BValue)
		//	if err != nil {
		//		return err
		//	}
		//} else if par.Scale > 0 {
		//	//par.oPrimValue, err = converters.NumberToString(par.BValue)
		//	var tempFloat string
		//	tempFloat, err = converters.NumberToString(par.BValue)
		//	if err != nil {
		//		return err
		//	}
		//	if strings.Contains(tempFloat, ".") {
		//		par.oPrimValue, err = strconv.ParseFloat(tempFloat, 64)
		//	} else {
		//		if strings.Contains(tempFloat, "-") {
		//			par.oPrimValue, err = strconv.ParseInt(tempFloat, 10, 64)
		//		} else {
		//			par.oPrimValue, err = strconv.ParseUint(tempFloat, 10, 64)
		//		}
		//	}
		//} else {
		//	par.oPrimValue = converters.DecodeNumber(par.BValue)
		//}
	case DATE, TIMESTAMP, TimeStampDTY:
		tempTime, err := converters.DecodeDate(par.BValue)
		if err != nil {
			return err
		}

		if conn.dbTimeZone != time.UTC {
			par.oPrimValue = time.Date(tempTime.Year(), tempTime.Month(), tempTime.Day(),
				tempTime.Hour(), tempTime.Minute(), tempTime.Second(), tempTime.Nanosecond(), conn.dbServerTimeZone)
		} else {
			par.oPrimValue = tempTime
		}

	case TIMESTAMPTZ, TimeStampTZ_DTY:
		tempTime, err := converters.DecodeDate(par.BValue)
		if err != nil {
			return err
		}
		par.oPrimValue = tempTime
	case TimeStampeLTZ, TimeStampLTZ_DTY:
		tempTime, err := converters.DecodeDate(par.BValue)
		if err != nil {
			return err
		}
		par.oPrimValue = tempTime
		if conn.dbTimeZone != time.UTC {
			par.oPrimValue = time.Date(tempTime.Year(), tempTime.Month(), tempTime.Day(),
				tempTime.Hour(), tempTime.Minute(), tempTime.Second(), tempTime.Nanosecond(), conn.dbTimeZone)
		}
	//case TimeStampDTY, TimeStampeLTZ, TimeStampLTZ_DTY, TIMESTAMPTZ, TimeStampTZ_DTY:
	//	fallthrough
	//case TIMESTAMP, DATE:
	//	tempTime, err := converters.DecodeDate(par.BValue)
	//	if err != nil {
	//		return err
	//	}
	//	if (par.DataType == DATE || par.DataType == TIMESTAMP || par.DataType == TimeStampDTY) && conn.dbTimeLoc != time.UTC {
	//
	//	} else {
	//
	//	}
	case OCIClobLocator, OCIBlobLocator:
		var locator []byte
		if !udt {
			locator, err = session.GetClr()
		} else {
			locator = par.BValue

		}
		if err != nil {
			return err
		}
		lob := Lob{
			sourceLocator: locator,
			sourceLen:     len(locator),
			connection:    conn,
			charsetID:     par.CharsetID,
		}
		if lob.isTemporary() {
			*temporaryLobs = append(*temporaryLobs, locator)
		}
		par.oPrimValue = lob
	case OCIFileLocator:
		var locator []byte
		if !udt {
			locator, err = session.GetClr()
		} else {
			locator = par.BValue
		}
		if err != nil {
			return err
		}
		var dirName, fileName string
		if len(locator) > 16 {
			index := 16
			length := int(binary.BigEndian.Uint16(locator[index : index+2]))
			index += 2
			dirName = conn.sStrConv.Decode(locator[index : index+length])
			index += length
			length = int(binary.BigEndian.Uint16(locator[index : index+2]))
			index += 2
			fileName = conn.sStrConv.Decode(locator[index : index+length])
			index += length
		}
		par.oPrimValue = BFile{
			dirName:  dirName,
			fileName: fileName,
			Valid:    len(locator) > 0,
			isOpened: false,
			lob: Lob{
				sourceLocator: locator,
				sourceLen:     len(locator),
				connection:    conn,
				charsetID:     par.CharsetID,
			},
		}
	case IBFloat:
		par.oPrimValue = converters.ConvertBinaryFloat(par.BValue)
	case IBDouble:
		par.oPrimValue = converters.ConvertBinaryDouble(par.BValue)
	case IntervalYM_DTY:
		par.oPrimValue = converters.ConvertIntervalYM_DTY(par.BValue)
	case IntervalDS_DTY:
		par.oPrimValue = converters.ConvertIntervalDS_DTY(par.BValue)
	case XMLType:
		err = decodeObject(conn, par, temporaryLobs)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("unable to decode oracle type %v to its primitive value", par.DataType)
	}
	return nil
}

func (par *ParameterInfo) decodeParameterValue(connection *Connection, temporaryLobs *[][]byte) error {
	return par.decodePrimValue(connection, temporaryLobs, false)
}

func (par *ParameterInfo) decodeColumnValue(connection *Connection, temporaryLobs *[][]byte, udt bool) error {
	//var err error
	if !udt && connection.connOption.Lob == configurations.INLINE && (par.DataType == OCIBlobLocator || par.DataType == OCIClobLocator) {
		session := connection.session
		maxSize, err := session.GetInt(4, true, true)
		if err != nil {
			return err
		}
		if maxSize > 0 {
			/*size*/ _, err = session.GetInt(8, true, true)
			if err != nil {
				return err
			}
			/*chunkSize*/ _, err := session.GetInt(4, true, true)
			if err != nil {
				return err
			}
			if par.DataType == OCIClobLocator {
				flag, err := session.GetByte()
				if err != nil {
					return err
				}
				par.CharsetID = 0
				if flag == 1 {
					par.CharsetID, err = session.GetInt(2, true, true)
					if err != nil {
						return err
					}
				}
				tempByte, err := session.GetByte()
				if err != nil {
					return err
				}
				par.CharsetForm = int(tempByte)
				if par.CharsetID == 0 {
					if par.CharsetForm == 1 {
						par.CharsetID = connection.tcpNego.ServerCharset
					} else {
						par.CharsetID = connection.tcpNego.ServernCharset
					}
				}
			}
			par.BValue, err = session.GetClr()
			if err != nil {
				return err
			}
			if par.DataType == OCIClobLocator {
				strConv, err := connection.getStrConv(par.CharsetID)
				if err != nil {
					return err
				}
				par.oPrimValue = strConv.Decode(par.BValue)
			} else {
				par.oPrimValue = par.BValue
			}
			_ /*locator*/, err = session.GetClr()
			if err != nil {
				return err
			}
		} else {
			par.oPrimValue = nil
		}
		return nil
	}
	//par.Value, err = par.decodeValue(connection, udt)
	return par.decodePrimValue(connection, temporaryLobs, udt)
}
