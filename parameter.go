package go_ora

import (
	"database/sql/driver"
	"math"
	"strings"

	"github.com/sijms/go-ora/network"
)

type OracleType int
type ParameterDirection int

const (
	Input  ParameterDirection = 1
	Output ParameterDirection = 2
	InOut  ParameterDirection = 3
	RetVal ParameterDirection = 9
)

//internal enum BindDirection
//{
//Output = 16,
//Input = 32,
//InputOutput = 48,
//}

//go:generate stringer -type=OracleType

const (
	NCHAR            OracleType = 1
	NUMBER           OracleType = 2
	SB1              OracleType = 3
	SB2              OracleType = 3
	SB4              OracleType = 3
	FLOAT            OracleType = 4
	NullStr          OracleType = 5
	VarNum           OracleType = 6
	LONG             OracleType = 8
	VARCHAR          OracleType = 9
	ROWID            OracleType = 11
	DATE             OracleType = 12
	VarRaw           OracleType = 15
	BFloat           OracleType = 21
	BDouble          OracleType = 22
	RAW              OracleType = 23
	LongRaw          OracleType = 24
	UINT             OracleType = 68
	LongVarChar      OracleType = 94
	LongVarRaw       OracleType = 95
	CHAR             OracleType = 96
	CHARZ            OracleType = 97
	IBFloat          OracleType = 100
	IBDouble         OracleType = 101
	REFCURSOR        OracleType = 102
	NOT              OracleType = 108
	XMLType          OracleType = 108
	OCIRef           OracleType = 110
	OCIClobLocator   OracleType = 112
	OCIBlobLocator   OracleType = 113
	OCIFileLocator   OracleType = 114
	ResultSet        OracleType = 116
	OCIString        OracleType = 155
	OCIDate          OracleType = 156
	TimeStampDTY     OracleType = 180
	TimeStampTZ_DTY  OracleType = 181
	IntervalYM_DTY   OracleType = 182
	IntervalDS_DTY   OracleType = 183
	TimeTZ           OracleType = 186
	TimeStamp        OracleType = 187
	TimeStampTZ      OracleType = 188
	IntervalYM       OracleType = 189
	IntervalDS       OracleType = 190
	UROWID           OracleType = 208
	TimeStampLTZ_DTY OracleType = 231
	TimeStampeLTZ    OracleType = 232
)

type ParameterType int

const (
	Number ParameterType = 1
	String ParameterType = 2
)

type ParameterInfo struct {
	Name                 string
	Direction            ParameterDirection
	IsNull               bool
	AllowNull            bool
	ColAlias             string
	DataType             OracleType
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
	getDataFromServer    bool
}

func (par *ParameterInfo) load(session *network.Session) error {
	par.getDataFromServer = true
	dataType, err := session.GetByte()
	if err != nil {
		return err
	}
	par.DataType = OracleType(dataType)
	par.Flag, err = session.GetByte()
	if err != nil {
		return err
	}
	par.Precision, err = session.GetByte()
	//precision, err := session.GetInt(1, false, false)
	//var scale int
	switch par.DataType {
	case NUMBER:
		fallthrough
	case TimeStampDTY:
		fallthrough
	case TimeStampTZ_DTY:
		fallthrough
	case IntervalDS_DTY:
		fallthrough
	case TimeStamp:
		fallthrough
	case TimeStampTZ:
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
		//scale, err = session.GetInt(1, false, false)
	}
	//if par.Scale == uint8(-127) {
	//
	//}
	if par.DataType == NUMBER && par.Precision == 0 && (par.Scale == 0 || par.Scale == 0xFF) {
		par.Precision = 38
		par.Scale = 0xFF
	}

	//par.Scale = uint16(scale)
	//par.Precision = uint16(precision)
	par.MaxLen, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	switch par.DataType {
	case ROWID:
		par.MaxLen = 128
	case DATE:
		par.MaxLen = 7
	case IBFloat:
		par.MaxLen = 4
	case IBDouble:
		par.MaxLen = 8
	case TimeStampTZ_DTY:
		par.MaxLen = 13
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
	par.ContFlag, err = session.GetInt(4, true, true)
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
	num1, err := session.GetInt(1, false, false)
	if err != nil {
		return err
	}
	par.AllowNull = num1 > 0
	_, err = session.GetByte() //  session.GetInt(1, false, false)
	if err != nil {
		return err
	}
	bName, err := session.GetDlc()
	if err != nil {
		return err
	}
	par.Name = session.StrConv.Decode(bName)
	_, err = session.GetDlc()
	bName, err = session.GetDlc()
	if err != nil {
		return err
	}
	if strings.ToUpper(string(bName)) == "XMLTYPE" {
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
	_, err = session.GetInt(4, true, true)
	return nil
}
func (par *ParameterInfo) write(session *network.Session) error {
	session.PutBytes(uint8(par.DataType), par.Flag, par.Precision, par.Scale)
	//session.PutUint(int(par.DataType), 1, false, false)
	//session.PutUint(par.Flag, 1, false, false)
	//session.PutUint(par.Precision, 1, false, false)
	//session.PutUint(par.Scale, 1, false, false)
	session.PutUint(par.MaxLen, 4, true, true)
	session.PutInt(par.MaxNoOfArrayElements, 4, true, true)
	session.PutInt(par.ContFlag, 4, true, true)
	if par.ToID == nil {
		session.PutBytes(0)
		//session.PutInt(0, 1, false, false)
	} else {
		session.PutInt(len(par.ToID), 4, true, true)
		session.PutClr(par.ToID)
	}
	session.PutUint(par.Version, 2, true, true)
	session.PutUint(par.CharsetID, 2, true, true)
	session.PutBytes(uint8(par.CharsetForm))
	//session.PutUint(par.CharsetForm, 1, false, false)
	session.PutUint(par.MaxCharLen, 4, true, true)
	return nil
}

//func NewIntegerParameter(name string, val int, direction ParameterDirection) *ParameterInfo {
//	ret := ParameterInfo{
//		Name:        name,
//		Direction:   direction,
//		flag:        3,
//		ContFlag:    0,
//		DataType:    NUMBER,
//		MaxCharLen:  22,
//		MaxLen:      22,
//		CharsetID:   871,
//		CharsetForm: 1,
//		BValue:       converters.EncodeInt(val),
//	}
//	return &ret
//}
//func NewStringParameter(name string, val string, size int, direction ParameterDirection) *ParameterInfo {
//	ret := ParameterInfo{
//		Name:        name,
//		Direction:   direction,
//		flag:        3,
//		ContFlag:    16,
//		DataType:    NCHAR,
//		MaxCharLen:  size,
//		MaxLen:      size,
//		CharsetID:   871,
//		CharsetForm: 1,
//		BValue:       []byte(val),
//	}
//	return &ret
//}

//func NewParamInfo(name string, parType ParameterType, size int, direction ParameterDirection) *ParameterInfo {
//	ret := new(ParameterInfo)
//	ret.Name = name
//	ret.Direction = direction
//	ret.flag = 3
//	//ret.DataType = dataType
//	switch parType {
//	case String:
//		ret.ContFlag = 16
//	default:
//		ret.ContFlag = 0
//	}
//	switch parType {
//	case Number:
//		ret.DataType = NUMBER
//		ret.MaxLen = 22
//	case String:
//		ret.CharsetForm = 1
//		ret.DataType = NCHAR
//		ret.MaxCharLen = size
//		ret.MaxLen = size
//	}
//	//ret.MaxCharLen = 0 // number of character to write
//	//ret.MaxLen = ret.MaxCharLen * 1 // number of character * byte per character
//	ret.CharsetID = 871
//	return ret
//	// if duplicateBind ret.flag = 128 else ret.flag = 3
//	// if collection type is assocative array ret.Flat |= 64
//
//	//num3 := 0
//	//switch dataType {
//	//case LONG:
//	//	fallthrough
//	//case LongRaw:
//	//	fallthrough
//	//case CHAR:
//	//	fallthrough
//	//case RAW:
//	//	fallthrough
//	//case NCHAR:
//	//	num3 = 1
//	//default:
//	//	num3 = 0
//	//}
//	//if num3 != 0 {
//	//
//	//}
//	//return ret
//}
