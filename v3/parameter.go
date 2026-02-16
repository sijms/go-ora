package go_ora

import (
	"database/sql/driver"
	"fmt"
	"math"
	"strings"

	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/type_coder"
	oraTypes "github.com/sijms/go-ora/v3/types"
)

type (
	ParameterDirection int
)

// func (n *NVarChar) ConvertValue(v interface{}) (driver.Value, error) {
//	return driver.Value(string(*n)), nil
// }

const (
	Input  ParameterDirection = 1
	Output ParameterDirection = 2
	InOut  ParameterDirection = 3
	// RetVal ParameterDirection = 9
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

// type ParameterType int

//const (
//	Number ParameterType = 1
//	String ParameterType = 2
//)

type ParameterInfo struct {
	Name         string
	TypeName     string
	SchemaName   string
	DomainSchema string
	DomainName   string
	Direction    ParameterDirection
	IsNull       bool
	AllowNull    bool
	IsJson       bool
	ColAlias     string
	//DataType             uint16
	IsXmlType bool
	//Flag                 uint8
	Precision uint8
	Scale     uint8
	//MaxLen               int
	//MaxCharLen           int
	MaxNoOfArrayElements int
	//ContFlag             int
	//ToID                 []byte
	Version int
	//CharsetID            int
	//CharsetForm          int
	//BValue            []byte
	Value             driver.Value
	encoder           type_coder.OracleTypeEncoder
	iPrimValue        driver.Value
	oPrimValue        driver.Value
	OutputVarPtr      interface{}
	getDataFromServer bool
	oaccollid         int
	cusType           *customType
	parent            *ParameterInfo
	Annotations       map[string]string
	type_coder.TypeInfo
}

// load get parameter information form network session
func (par *ParameterInfo) load(conn *Connection) error {
	session := conn.session
	par.getDataFromServer = true
	dataType, err := session.GetByte()
	if err != nil {
		return err
	}
	par.DataType = uint16(dataType)
	par.Flag, err = session.GetByte()
	if err != nil {
		return err
	}
	par.Precision, err = session.GetByte()
	// precision, err := session.GetInt(1, false, false)
	// var scale int
	switch par.DataType {
	case oraTypes.NUMBER:
		fallthrough
	case oraTypes.TimeStampDTY:
		fallthrough
	case oraTypes.TimeStampTZ_DTY:
		fallthrough
	case oraTypes.INTERVALDS_DTY:
		fallthrough
	case oraTypes.TIMESTAMP:
		fallthrough
	case oraTypes.TIMESTAMPTZ:
		fallthrough
	case oraTypes.IntervalDS:
		fallthrough
	case oraTypes.TimeStampLTZ_DTY:
		fallthrough
	case oraTypes.TimeStampeLTZ:
		scale, err := session.GetInt(2, true, true)
		if err != nil {
			return err
		}

		if scale == -127 {
			par.Precision = uint8(math.Ceil(float64(par.Precision) * 0.30103))
			par.Scale = 0xFF
		} else {
			par.Scale = uint8(scale)
		}
	default:
		par.Scale, err = session.GetByte()
		// scale, err = session.GetInt(1, false, false)
	}
	// if par.Scale == uint8(-127) {
	//
	// }
	if par.DataType == oraTypes.NUMBER && par.Precision == 0 && (par.Scale == 0 || par.Scale == 0xFF) {
		par.Precision = 38
		par.Scale = 0xFF
	}

	// par.Scale = uint16(scale)
	// par.Precision = uint16(precision)
	par.MaxLen, err = session.GetInt64(4, true, true)
	if err != nil {
		return err
	}
	switch par.DataType {
	case oraTypes.ROWID:
		par.MaxLen = 128
	case oraTypes.DATE:
		par.MaxLen = int64(converters.MAX_LEN_DATE)
	case oraTypes.IBFLOAT:
		par.MaxLen = 4
	case oraTypes.IBDOUBLE:
		par.MaxLen = 8
	case oraTypes.TimeStampTZ_DTY:
		par.MaxLen = int64(converters.MAX_LEN_TIMESTAMP)
	case oraTypes.INTERVALYM_DTY:
		fallthrough
	case oraTypes.INTERVALDS_DTY:
		fallthrough
	case oraTypes.IntervalYM:
		fallthrough
	case oraTypes.IntervalDS:
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
	par.MaxCharLen, err = session.GetInt64(4, true, true)
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
	if par.DataType == oraTypes.XMLType && par.TypeName != "XMLTYPE" {
		for typName, cusTyp := range conn.cusTyp {
			if typName == par.TypeName {
				par.cusType = new(customType)
				*par.cusType = cusTyp
			}
		}
	}
	if par.TypeName == "XMLTYPE" {
		par.DataType = oraTypes.XMLType
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
	if session.TTCVersion < 24 {
		return nil
	}
	if par.DataType == oraTypes.VECTOR {
		par.VectorDim, err = session.GetInt(4, true, true)
		if err != nil {
			return err
		}
		par.VectorFormat, err = session.GetByte()
		if err != nil {
			return err
		}
		par.VectorFlag, err = session.GetByte()

		par.VectorType = oraTypes.VECTOR_DENSE
		if par.VectorFlag&2 == 2 {
			par.VectorType = oraTypes.VECTOR_SPARSE
		}
		//colMetaData.m_vectorDim = (int)mEngine.UnmarshalUB4(bAsync, false);
		//colMetaData.m_vectorNumFormat = (VectorNumFormat)mEngine.UnmarshalUB1(bAsync, false);
		//colMetaData.m_vectorFlag = (int)((byte)mEngine.UnmarshalUB1(bAsync, false));
		//colMetaData.m_vectorType = (((colMetaData.m_vectorFlag & 2) != 0) ? OracleVectorType.Sparse : OracleVectorType.Dense);
	} else {
		_, err = session.GetInt(4, true, true)
		if err != nil {
			return err
		}
		_, err = session.GetInt(1, true, true)
		if err != nil {
			return err
		}
		_, err = session.GetInt(1, true, true)
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
	//switch value := par.iPrimValue.(type) {
	//case *Lob:
	//	if value != nil && value.sourceLocator != nil {
	//		return [][]byte{value.sourceLocator}
	//	}
	//case *BFile:
	//	if value != nil && value.lob.sourceLocator != nil {
	//		return [][]byte{value.lob.sourceLocator}
	//	}
	//case []ParameterInfo:
	//	output := make([][]byte, 0, 10)
	//	for _, temp := range value {
	//		output = append(output, temp.collectLocators()...)
	//	}
	//	return output
	//}
	return [][]byte{}
}

func (par *ParameterInfo) isLongType() bool {
	return par.DataType == oraTypes.LONG ||
		par.DataType == oraTypes.LongRaw ||
		par.DataType == oraTypes.LongVarChar ||
		par.DataType == oraTypes.LongVarRaw
}

func (par *ParameterInfo) isLobType() bool {
	return par.DataType == oraTypes.OCIBlobLocator ||
		par.DataType == oraTypes.OCIClobLocator ||
		par.DataType == oraTypes.OCIFileLocator ||
		par.DataType == oraTypes.VECTOR ||
		par.DataType == oraTypes.JSON
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
	if par.DataType == oraTypes.XMLType && par.parent == nil {
		//if par.TypeName == "XMLTYPE" {
		//	return errors.New("unsupported data type: XMLTYPE")
		//}
		//if par.cusType == nil {
		//	return fmt.Errorf("unregister custom type: %s. call RegisterType first", par.TypeName)
		//}
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
		}
		par.IsNull = false
	}
	if decoder, ok := conn.typeDecoder[par.DataType]; ok {
		par.IsUDTPar = udt
		decoder.SetTypeInfo(par.TypeInfo)
		decoder.SetCharsetCoder(conn)
		if par.isLobType() {
			streamer := &LobStream{conn: conn}
			decoder.SetLobStreamer(streamer)
		}
		par.oPrimValue, err = decoder.Read(session)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("unknown data type: %d", par.DataType)
	}

	//switch par.DataType {

	//case OCIClobLocator, OCIBlobLocator:
	//	var locator []byte
	//	if !udt {
	//		locator, err = session.GetClr()
	//	} else {
	//		locator = par.BValue
	//	}
	//	if err != nil {
	//		return err
	//	}
	//	lob := Lob{
	//		sourceLocator: locator,
	//		sourceLen:     len(locator),
	//		connection:    conn,
	//		charsetID:     par.CharsetID,
	//	}
	//	if lob.isTemporary() {
	//		*temporaryLobs = append(*temporaryLobs, locator)
	//	}
	//	par.oPrimValue = lob
	//case VECTOR:
	//	var locator []byte
	//	if !udt {
	//		locator, err = session.GetClr()
	//	} else {
	//		locator = par.BValue
	//	}
	//	if err != nil {
	//		return err
	//	}
	//	lob := Lob{
	//		sourceLocator: locator,
	//		sourceLen:     len(locator),
	//		connection:    conn,
	//		charsetID:     par.CharsetID,
	//	}
	//	v := oraTypes.Vector{}
	//	if lob.isTemporary() {
	//		*temporaryLobs = append(*temporaryLobs, locator)
	//	}
	//	par.oPrimValue = v
	//case OCIFileLocator:
	//	var locator []byte
	//	if !udt {
	//		locator, err = session.GetClr()
	//	} else {
	//		locator = par.BValue
	//	}
	//	if err != nil {
	//		return err
	//	}
	//	var dirName, fileName string
	//	if len(locator) > 16 {
	//		index := 16
	//		length := int(binary.BigEndian.Uint16(locator[index : index+2]))
	//		index += 2
	//		dirName = conn.sStrConv.Decode(locator[index : index+length])
	//		index += length
	//		length = int(binary.BigEndian.Uint16(locator[index : index+2]))
	//		index += 2
	//		fileName = conn.sStrConv.Decode(locator[index : index+length])
	//		index += length
	//	}
	//	lob := &Lob{
	//		sourceLocator: locator,
	//		sourceLen:     len(locator),
	//		connection:    conn,
	//		charsetID:     par.CharsetID,
	//	}
	//	f := oraTypes.CreateBFileFromStream(lob, dirName, fileName)
	//	//f := oraTypes.BFile{
	//	//	dirName:  dirName,
	//	//	fileName: fileName,
	//	//	Valid:    len(locator) > 0,
	//	//	isOpened: false,
	//	//	,
	//	//}
	//	if lob.isTemporary() {
	//		*temporaryLobs = append(*temporaryLobs, locator)
	//	}
	//	par.oPrimValue = f

	//case XMLType:
	//	err = decodeObject(conn, par, temporaryLobs)
	//	if err != nil {
	//		return err
	//	}
	//default:
	//	return fmt.Errorf("unable to decode oracle type %v to its primitive value", par.DataType)
	//}
	return nil
}

func (par *ParameterInfo) decodeParameterValue(connection *Connection, temporaryLobs *[][]byte) error {
	currentLobFetch := connection.connOption.Lob
	connection.connOption.Lob = configurations.STREAM
	defer func() {
		connection.connOption.Lob = currentLobFetch
	}()
	err := par.decodePrimValue(connection, temporaryLobs, false)
	if err != nil {
		return err
	}
	return nil
	//if temp, ok := par.Value.(sql.Scanner); ok {
	//	err = temp.Scan(par.oPrimValue)
	//	if err != nil {
	//		return err
	//	}
	//	return nil
	//}
	//return fmt.Errorf("the input parameter does not support scanner interface")
}

func (par *ParameterInfo) decodeColumnValue(connection *Connection, temporaryLobs *[][]byte, udt bool) error {
	// var err error
	//if !udt && connection.connOption.Lob == configurations.INLINE && (par.isLobType()) {
	//	session := connection.session
	//	maxSize, err := session.GetInt(4, true, true)
	//	if err != nil {
	//		return err
	//	}
	//	if maxSize > 0 {
	//		/*size*/ _, err = session.GetInt(8, true, true)
	//		if err != nil {
	//			return err
	//		}
	//		/*chunkSize*/ _, err = session.GetInt(4, true, true)
	//		if err != nil {
	//			return err
	//		}
	//		if par.DataType == oraTypes.OCIClobLocator {
	//			flag, err := session.GetByte()
	//			if err != nil {
	//				return err
	//			}
	//			par.CharsetID = 0
	//			if flag == 1 {
	//				par.CharsetID, err = session.GetInt(2, true, true)
	//				if err != nil {
	//					return err
	//				}
	//			}
	//			tempByte, err := session.GetByte()
	//			if err != nil {
	//				return err
	//			}
	//			par.CharsetForm = int(tempByte)
	//			if par.CharsetID == 0 {
	//				if par.CharsetForm == 1 {
	//					par.CharsetID = connection.tcpNego.ServerCharset
	//				} else {
	//					par.CharsetID = connection.tcpNego.ServernCharset
	//				}
	//			}
	//		}
	//		par.BValue, err = session.GetClr()
	//		if err != nil {
	//			return err
	//		}
	//		_ /*locator*/, err = session.GetClr()
	//		if err != nil {
	//			return err
	//		}
	//		if par.DataType == oraTypes.OCIClobLocator {
	//			strConv, err := connection.GetStringCoder(par.CharsetID, par.CharsetForm)
	//			if err != nil {
	//				return err
	//			}
	//			par.oPrimValue = strConv.Decode(par.BValue)
	//		} else if par.DataType == oraTypes.VECTOR {
	//			//v := Vector{}
	//			//err = v.decode(par.BValue)
	//			//if err != nil {
	//			//	return err
	//			//}
	//			//par.oPrimValue = v.Data
	//
	//		} else {
	//			par.oPrimValue = par.BValue
	//		}
	//
	//	} else {
	//		par.oPrimValue = nil
	//	}
	//	return nil
	//}
	// par.Value, err = par.decodeValue(connection, udt)
	return par.decodePrimValue(connection, temporaryLobs, udt)
}
