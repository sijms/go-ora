package go_ora

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/sijms/go-ora/v3/advanced_nego"
	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/parameter_coder"
	"github.com/sijms/go-ora/v3/types"
	"github.com/sijms/go-ora/v3/types/oson"
)

type OracleDriver struct {
	dataCollected   bool
	oracleTypeCoder map[uint16]parameter_coder.OracleParameterCoder
	goTypeCoder     map[reflect.Type]parameter_coder.OracleParameterCoder
	jsonEncoder     map[reflect.Type]oson.FieldEncoder
	cusTyp          map[string]Object
	sessionParam    map[string]string
	mu              sync.Mutex
	sStrConv        converters.IStringConverter
	nStrConv        converters.IStringConverter
	UserId          string
	connOption      *configurations.ConnectionConfig
	// Server    string
	// Port      int
	// Instance  string
	// Service   string
	// DBName    string

	// SessionId int
	// SerialNum int
}

func NewDriver() *OracleDriver {
	drv := &OracleDriver{
		goTypeCoder:     make(map[reflect.Type]parameter_coder.OracleParameterCoder),
		oracleTypeCoder: make(map[uint16]parameter_coder.OracleParameterCoder),
		//typeDecoder:  make(map[uint16]type_coder.OracleTypeDecoder),
		cusTyp:       map[string]Object{},
		sessionParam: map[string]string{},
	}
	drv.init()
	return drv
}

var oracleDriver = NewDriver()

type GetDriverInterface interface {
	Driver() driver.Driver
}

func init() {
	sql.Register("oracle", oracleDriver)
}

func GetDefaultDriver() *OracleDriver {
	return oracleDriver
}

func (driver *OracleDriver) init() {
	// add basic type decoders here
	driver.oracleTypeCoder[types.NCHAR] = &parameter_coder.StringParameter{}
	driver.oracleTypeCoder[types.CHAR] = &parameter_coder.StringParameter{}
	driver.oracleTypeCoder[types.LONG] = &parameter_coder.StringParameter{}
	driver.oracleTypeCoder[types.LongVarChar] = &parameter_coder.StringParameter{}

	driver.oracleTypeCoder[types.RAW] = &parameter_coder.RawParameter{}
	driver.oracleTypeCoder[types.LongRaw] = &parameter_coder.RawParameter{}

	driver.oracleTypeCoder[types.NUMBER] = &parameter_coder.NumberParameter{}
	driver.oracleTypeCoder[types.IBFLOAT] = &parameter_coder.NumberParameter{}
	driver.oracleTypeCoder[types.IBDOUBLE] = &parameter_coder.NumberParameter{}

	driver.oracleTypeCoder[types.DATE] = &parameter_coder.DateParameter{}
	driver.oracleTypeCoder[types.TIMESTAMP] = &parameter_coder.DateParameter{}
	driver.oracleTypeCoder[types.TIMESTAMPTZ] = &parameter_coder.DateParameter{}

	driver.oracleTypeCoder[types.VECTOR] = &parameter_coder.VectorParameter{}
	driver.oracleTypeCoder[types.JSON] = &parameter_coder.JsonParameter{}
	driver.oracleTypeCoder[types.OCIBlobLocator] = &parameter_coder.BlobParameter{}
	driver.oracleTypeCoder[types.OCIClobLocator] = &parameter_coder.ClobParameter{}
	driver.oracleTypeCoder[types.OCIFileLocator] = &parameter_coder.BFileParameter{}

	driver.oracleTypeCoder[types.ROWID] = &parameter_coder.RowIDParameter{}
	driver.oracleTypeCoder[types.UROWID] = &parameter_coder.RowIDParameter{}

	driver.oracleTypeCoder[types.BOOLEAN] = &parameter_coder.BoolParameter{}

	driver.goTypeCoder[types.TyString] = &parameter_coder.StringParameter{}
	driver.goTypeCoder[types.TyNullString] = &parameter_coder.StringParameter{}
	driver.goTypeCoder[types.TyVarchar] = &parameter_coder.StringParameter{}

	driver.goTypeCoder[types.TyBytes] = &parameter_coder.RawParameter{}
	driver.goTypeCoder[types.TyRaw] = &parameter_coder.RawParameter{}

	driver.goTypeCoder[types.TyInt] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyInt8] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyInt16] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyInt32] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyInt64] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyUint] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyUint8] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyUint16] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyUint32] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyUint64] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyFloat32] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyFloat64] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyNullByte] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyNullInt16] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyNullInt32] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyNullInt64] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyNullFloat64] = &parameter_coder.NumberParameter{}
	driver.goTypeCoder[types.TyNumber] = &parameter_coder.NumberParameter{}

	driver.goTypeCoder[types.TyBoolean] = &parameter_coder.BoolParameter{}

	driver.goTypeCoder[types.TyTime] = &parameter_coder.DateParameter{}
	driver.goTypeCoder[types.TyNullTime] = &parameter_coder.DateParameter{}
	driver.goTypeCoder[types.TyDate] = &parameter_coder.DateParameter{}

	driver.goTypeCoder[types.TyInterval] = &parameter_coder.IntervalParameter{}

	driver.goTypeCoder[types.TyVector] = &parameter_coder.VectorParameter{}
	driver.goTypeCoder[types.TyJson] = &parameter_coder.JsonParameter{}

	driver.goTypeCoder[types.TyClob] = &parameter_coder.ClobParameter{}
	driver.goTypeCoder[types.TyBlob] = &parameter_coder.BlobParameter{}
	driver.goTypeCoder[types.TyBFile] = &parameter_coder.BFileParameter{}

}
func (driver *OracleDriver) initFromConn(conn *Connection) error {
	driver.mu.Lock()
	defer driver.mu.Unlock()
	if !driver.dataCollected {
		driver.UserId = conn.connOption.UserID
		if driver.sStrConv == nil {
			driver.sStrConv = conn.sStrConv.Clone()
		}
		if driver.nStrConv == nil {
			driver.nStrConv = conn.nStrConv.Clone()
		}
		if driver.connOption == nil {
			driver.connOption = new(configurations.ConnectionConfig)
			*driver.connOption = *conn.connOption
		}
		driver.dataCollected = true
	}

	// update session parameters
	var err error
	for key, value := range driver.sessionParam {
		_, err = conn.Exec(fmt.Sprintf("alter session set %s=%s", key, value))
		if err != nil {
			return err
		}
	}
	return err
}

// SetStringConverter this function is used to set a custom string converter interface
// that will be used to encode and decode strings and bytearrays
// passing nil will use driver string converter for supported langs
func SetStringConverter(db GetDriverInterface, charset, nCharset converters.IStringConverter) {
	if drv, ok := db.Driver().(*OracleDriver); ok {
		drv.sStrConv = charset
		drv.nStrConv = nCharset
	}
}

func DelSessionParam(db *sql.DB, key string) {
	if drv, ok := db.Driver().(*OracleDriver); ok {
		drv.mu.Lock()
		defer drv.mu.Unlock()
		delete(drv.sessionParam, key)
	}
}

func AddSessionParam(db *sql.DB, key, value string) error {
	_, err := db.Exec(fmt.Sprintf("alter session set %s=%s", key, value))
	if err != nil {
		return err
	}
	if drv, ok := db.Driver().(*OracleDriver); ok {
		drv.mu.Lock()
		defer drv.mu.Unlock()
		drv.sessionParam[key] = value
	}
	return nil
}

//func RegisterJsonFieldEncoder(db GetDriverInterface, key reflect.Type, encoder oson.FieldEncoder) error {
//	if drv, ok := db.Driver().(*OracleDriver); ok {
//		drv.mu.Lock()
//		drv.jsonEncoder[key] = encoder
//		drv.mu.Unlock()
//		return nil
//	}
//	return errors.New("the driver used is not a go-ora driver type")
//}

//	func RegisterJsonFieldDecoder(db GetDriverInterface, opCode int, decoder oson.FieldDecoder) error {
//		if drv, ok := db.Driver().(*OracleDriver); ok {
//			drv.mu.Lock()
//			drv.jsonDecoder[opCode] = decoder
//			drv.mu.Unlock()
//			return nil
//		}
//		return errors.New("the driver used is not a go-ora driver type")
//	}
//func RegisterParameterEncoder(db *sql.DB, _type reflect.Type, encoder parameter_coder.OracleParameterEncoder) error {
//	if drv, ok := db.Driver().(*OracleDriver); ok {
//		drv.mu.Lock()
//		drv.goTypeCoder[_type] = encoder
//		drv.mu.Unlock()
//	}
//	return errors.New("the driver used is not a go-ora driver type")
//}
//func RegisterParameterDecoder(db *sql.DB, typeId uint16, decoder parameter_coder.OracleParameterDecoder) error {
//	if drv, ok := db.Driver().(*OracleDriver); ok {
//		drv.mu.Lock()
//		drv.oracleTypeCoder[typeId] = decoder
//		drv.mu.Unlock()
//		return nil
//	}
//	return errors.New("the driver used is not a go-ora driver type")
//}

func AddParameterCoder(db *sql.DB, go_type reflect.Type, oracle_type uint16, coder parameter_coder.OracleParameterCoder) error {
	if drv, ok := db.Driver().(*OracleDriver); ok {
		drv.mu.Lock()
		drv.oracleTypeCoder[oracle_type] = coder
		drv.goTypeCoder[go_type] = coder
		drv.mu.Unlock()
		return nil
	}
	return errors.New("the driver used is not a go-ora driver type")
}

// func RegisterRegularTypeArray(conn *sql.DB, regularTypeName, arrayTypeName string, itemMaxSize int) error {
// 	err := conn.Ping()
// 	if err != nil {
// 		return err
// 	}
//
// 	if drv, ok := conn.Driver().(*OracleDriver); ok {
// 		return RegisterRegularTypeArrayWithOwner(conn, drv.UserId, regularTypeName, arrayTypeName, itemMaxSize)
// 	}
// 	return errors.New("the driver used is not a go-ora driver type")
// }
// func RegisterRegularTypeArrayWithOwner(conn *sql.DB, owner, regularTypeName, arrayTypeName string, itemMaxSize int) error {
// 	drv := conn.Driver().(*OracleDriver)
// 	regularTypeName = strings.TrimSpace(regularTypeName)
// 	arrayTypeName = strings.TrimSpace(arrayTypeName)
// 	if len(regularTypeName) == 0 {
// 		return errors.New("typeName shouldn't be empty")
// 	}
// 	if len(arrayTypeName) == 0 {
// 		return errors.New("array type name shouldn't be empty")
// 	}
// 	cust := customType{
// 		owner:         owner,
// 		name:          regularTypeName,
// 		arrayTypeName: arrayTypeName,
// 		isArray:       true,
// 	}
// 	var err error
// 	cust.arrayTOID, err = getTOID2(conn, owner, arrayTypeName)
// 	if err != nil {
// 		return err
// 	}
// 	param := ParameterInfo{Direction: Input, Flag: 3, TypeName: regularTypeName, MaxLen: 1, MaxCharLen: 1}
// 	switch strings.ToUpper(regularTypeName) {
// 	case "NUMBER":
// 		param.DataType = NUMBER
// 	case "VARCHAR2":
// 		param.DataType = NCHAR
// 		param.CharsetForm = 1
// 		param.ContFlag = 16
// 		param.CharsetID = drv.sStrConv.GetLangID()
// 	case "NVARCHAR2":
// 		param.DataType = NCHAR
// 		param.CharsetForm = 2
// 		param.ContFlag = 16
// 		param.CharsetID = drv.nStrConv.GetLangID()
// 	case "TIMESTAMP":
// 		param.DataType = TimeStampDTY
// 	case "DATE":
// 		param.DataType = DATE
// 	case "TIMESTAMP WITH LOCAL TIME ZONE":
// 		param.DataType = TimeStampLTZ_DTY
// 	// case "TIMESTAMP WITH TIME ZONE":
// 	// 	param.DataType = TimeStampTZ_DTY
// 	// 	param.MaxLen = converters.MAX_LEN_TIMESTAMP
// 	case "RAW":
// 		param.DataType = RAW
// 	case "BLOB":
// 		param.DataType = OCIBlobLocator
// 	case "CLOB":
// 		param.DataType = OCIClobLocator
// 		param.CharsetForm = 1
// 		param.ContFlag = 16
// 		param.CharsetID = drv.sStrConv.GetLangID()
// 	case "NCLOB":
// 		param.DataType = OCIClobLocator
// 		param.CharsetForm = 2
// 		param.ContFlag = 16
// 		param.CharsetID = drv.nStrConv.GetLangID()
// 	default:
// 		return fmt.Errorf("unsupported regular type: %s", regularTypeName)
// 	}
// 	cust.attribs = append(cust.attribs, param)
// 	drv.mu.Lock()
// 	defer drv.mu.Unlock()
// 	drv.cusTyp[strings.ToUpper(arrayTypeName)] = cust
// 	return nil
// }

func RegisterType(db *sql.DB, typeName, arrayTypeName string, typeObj interface{}) error {
	// ping first to avoid error when calling register type after open connection
	err := db.Ping()
	if err != nil {
		return err
	}

	if drv, ok := db.Driver().(*OracleDriver); ok {
		return RegisterTypeWithOwner(db, drv.UserId, typeName, arrayTypeName, typeObj)
	}
	return errors.New("the driver used is not a go-ora driver type")
}

func RegisterTypeWithOwner(conn *sql.DB, owner, typeName, arrayTypeName string, typeObj interface{}) error {
	if len(owner) == 0 {
		return errors.New("owner can't be empty")
	}
	drv := conn.Driver().(*OracleDriver)

	//if typeObj == nil {
	//	return errors.New("type object cannot be nil")
	//}
	var typ reflect.Type
	if typeObj != nil {
		typ = reflect.TypeOf(typeObj)
		switch typ.Kind() {
		case reflect.Ptr:
			return errors.New("unsupported type object: Ptr")
		case reflect.Array:
			return errors.New("unsupported type object: Array")
		case reflect.Chan:
			return errors.New("unsupported type object: Chan")
		case reflect.Map:
			return errors.New("unsupported type object: Map")
		case reflect.Slice:
			return errors.New("unsupported type object: Slice")
		}
	}
	typeName = strings.TrimSpace(typeName)
	arrayTypeName = strings.TrimSpace(arrayTypeName)
	if len(typeName) == 0 {
		return errors.New("typeName shouldn't be empty")
	}

	cust := Object{
		Owner: owner,
		Name:  typeName,
		// arrayTypeName: arrayTypeName,
		typ: typ,
	}
	arrayCust := Object{Owner: owner, Name: arrayTypeName, isArray: true}
	var err error
	arrayParam := ParameterInfo{
		Direction: Input,
		BasicParameter: parameter_coder.BasicParameter{
			Flag:       3,
			MaxLen:     1,
			MaxCharLen: 1,
		},
		TypeName: typeName,
	}
	switch strings.ToUpper(typeName) {
	case "NUMBER":
		arrayParam.DataType = types.NUMBER
	case "VARCHAR2":
		arrayParam.DataType = types.NCHAR
		arrayParam.CharsetForm = 1
		arrayParam.ContFlag = 16
		arrayParam.CharsetID = drv.sStrConv.GetLangID()
	case "NVARCHAR2":
		arrayParam.DataType = types.NCHAR
		arrayParam.CharsetForm = 2
		arrayParam.ContFlag = 16
		arrayParam.CharsetID = drv.nStrConv.GetLangID()
	case "TIMESTAMP":
		arrayParam.DataType = types.TimeStampDTY
	case "DATE":
		arrayParam.DataType = types.DATE
	case "TIMESTAMP WITH LOCAL TIME ZONE":
		arrayParam.DataType = types.TimeStampLTZ_DTY
	// case "TIMESTAMP WITH TIME ZONE":
	//	arrayParam.DataType = TimeStampTZ_DTY
	//	arrayParam.MaxLen = converters.MAX_LEN_TIMESTAMP
	case "RAW":
		arrayParam.DataType = types.RAW
	case "BLOB":
		arrayParam.DataType = types.OCIBlobLocator
	case "CLOB":
		arrayParam.DataType = types.OCIClobLocator
		arrayParam.CharsetForm = 1
		arrayParam.ContFlag = 16
		arrayParam.CharsetID = drv.sStrConv.GetLangID()
	case "NCLOB":
		arrayParam.DataType = types.OCIClobLocator
		arrayParam.CharsetForm = 2
		arrayParam.ContFlag = 16
		arrayParam.CharsetID = drv.nStrConv.GetLangID()
	default:
		if cust.typ == nil {
			return errors.New("type object cannot be nil")
		}
		if typ.Kind() != reflect.Struct {
			return errors.New("type object should be of structure type")
		}
		cust.fieldMap = map[string]int{}
		cust.toid, err = getTOID2(conn, owner, typeName)
		if err != nil {
			return err
		}
		sqlText := `SELECT ATTR_NAME, ATTR_TYPE_NAME, LENGTH, ATTR_NO 
					FROM ALL_TYPE_ATTRS 
					WHERE UPPER(OWNER)=:1 AND UPPER(TYPE_NAME)=:2
					ORDER BY ATTR_NO`
		rows, err := conn.Query(sqlText, strings.ToUpper(owner), strings.ToUpper(typeName))
		if err != nil {
			return err
		}
		var (
			attName     sql.NullString
			attOrder    int64
			attTypeName sql.NullString
			length      sql.NullInt64
		)
		for rows.Next() {
			err = rows.Scan(&attName, &attTypeName, &length, &attOrder)
			if err != nil {
				return err
			}
			param := ParameterInfo{Direction: Input, BasicParameter: parameter_coder.BasicParameter{Flag: 3}}
			param.Name = attName.String
			param.TypeName = attTypeName.String
			switch strings.ToUpper(attTypeName.String) {
			case "NUMBER":
				param.DataType = types.NUMBER
				param.MaxLen = int64(converters.MAX_LEN_NUMBER)
			case "VARCHAR2":
				param.DataType = types.NCHAR
				param.CharsetForm = 1
				param.ContFlag = 16
				param.MaxCharLen = length.Int64
				param.CharsetID = drv.sStrConv.GetLangID()
				param.MaxLen = length.Int64 * int64(converters.MaxBytePerChar(param.CharsetID))
			case "NVARCHAR2":
				param.DataType = types.NCHAR
				param.CharsetForm = 2
				param.ContFlag = 16
				param.MaxCharLen = length.Int64
				param.CharsetID = drv.nStrConv.GetLangID()
				param.MaxLen = length.Int64 * int64(converters.MaxBytePerChar(param.CharsetID))
			case "TIMESTAMP":
				param.DataType = types.TimeStampDTY
				param.MaxLen = int64(converters.MAX_LEN_DATE)
			case "DATE":
				param.DataType = types.DATE
				param.MaxLen = int64(converters.MAX_LEN_DATE)
			case "TIMESTAMP WITH LOCAL TIME ZONE":
				param.DataType = types.TimeStampLTZ_DTY
				param.MaxLen = int64(converters.MAX_LEN_DATE)
			case "TIMESTAMP WITH TIME ZONE":
				param.DataType = types.TimeStampTZ_DTY
				param.MaxLen = int64(converters.MAX_LEN_TIMESTAMP)
			case "RAW":
				param.DataType = types.RAW
				param.MaxLen = length.Int64
			case "BLOB":
				param.DataType = types.OCIBlobLocator
				param.MaxLen = length.Int64
			case "CLOB":
				param.DataType = types.OCIClobLocator
				param.CharsetForm = 1
				param.ContFlag = 16
				param.CharsetID = drv.sStrConv.GetLangID()
				param.MaxCharLen = length.Int64
				param.MaxLen = length.Int64 * int64(converters.MaxBytePerChar(param.CharsetID))
			case "NCLOB":
				param.DataType = types.OCIClobLocator
				param.CharsetForm = 2
				param.ContFlag = 16
				param.CharsetID = drv.nStrConv.GetLangID()
				param.MaxCharLen = length.Int64
				param.MaxLen = length.Int64 * int64(converters.MaxBytePerChar(param.CharsetID))
			default:
				found := false
				for name, value := range drv.cusTyp {
					if strings.EqualFold(name, attTypeName.String) {
						found = true
						param.DataType = types.XMLType
						param.cusType = new(Object)
						*param.cusType = value
						param.ToID = value.toid
						break
					}
				}
				if !found {
					return fmt.Errorf("unsupported attribute type: %s", attTypeName.String)
				}
			}
			cust.attribs = append(cust.attribs, param)
		}
		if len(cust.attribs) == 0 {
			return fmt.Errorf("unknown or empty type: %s", typeName)
		}
		arrayParam.DataType = types.XMLType
		arrayParam.cusType = new(Object)
		*arrayParam.cusType = cust

		cust.loadFieldMap()
		drv.mu.Lock()
		drv.cusTyp[strings.ToUpper(typeName)] = cust
		drv.mu.Unlock()
	}
	if len(arrayTypeName) > 0 {
		var err error
		arrayCust.toid, err = getTOID2(conn, owner, arrayTypeName)
		if err != nil {
			return err
		}
		arrayCust.attribs = append(arrayCust.attribs, arrayParam)
		drv.mu.Lock()
		drv.cusTyp[strings.ToUpper(arrayTypeName)] = arrayCust
		drv.mu.Unlock()
	}

	return nil
}

func ParseConfig(dsn string) (*configurations.ConnectionConfig, error) {
	config, err := configurations.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	if len(config.OSPassword) > 0 {
		SetNTSAuth(&advanced_nego.NTSAuthHash{})
	}
	return config, nil
	// connStr, err := newConnectionStringFromUrl(dsn)
	//i f err != nil {
	// 	return nil, err
	// }
	// return &connStr.connOption, nil
}

func RegisterConnConfig(config *configurations.ConnectionConfig) {
	oracleDriver.connOption = config
}
