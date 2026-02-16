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
	"github.com/sijms/go-ora/v3/type_coder"
	"github.com/sijms/go-ora/v3/types"
)

type OracleDriver struct {
	dataCollected bool
	typeDecoder   map[uint16]type_coder.OracleTypeDecoder
	cusTyp        map[string]customType
	sessionParam  map[string]string
	mu            sync.Mutex
	sStrConv      converters.IStringConverter
	nStrConv      converters.IStringConverter
	UserId        string
	connOption    *configurations.ConnectionConfig
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
		typeDecoder:  make(map[uint16]type_coder.OracleTypeDecoder),
		cusTyp:       map[string]customType{},
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
	driver.typeDecoder[types.NUMBER] = &type_coder.NumberCoder{}

	driver.typeDecoder[types.NCHAR] = &type_coder.StringCoder{}
	driver.typeDecoder[types.CHAR] = &type_coder.StringCoder{}
	driver.typeDecoder[types.LONG] = &type_coder.StringCoder{}
	driver.typeDecoder[types.LongVarChar] = &type_coder.StringCoder{}

	driver.typeDecoder[types.RAW] = &type_coder.RawCoder{}
	driver.typeDecoder[types.LongRaw] = &type_coder.RawCoder{}
	driver.typeDecoder[types.LongVarRaw] = &type_coder.RawCoder{}

	driver.typeDecoder[types.IBFLOAT] = &type_coder.DoubleCoder{}
	driver.typeDecoder[types.BFLOAT] = &type_coder.DoubleCoder{}
	driver.typeDecoder[types.IBDOUBLE] = &type_coder.DoubleCoder{}
	driver.typeDecoder[types.BDOUBLE] = &type_coder.DoubleCoder{}

	driver.typeDecoder[types.DATE] = &type_coder.DateCoder{}
	driver.typeDecoder[types.TIMESTAMP] = &type_coder.DateCoder{}
	driver.typeDecoder[types.TimeStampDTY] = &type_coder.DateCoder{}
	driver.typeDecoder[types.TIMESTAMPTZ] = &type_coder.DateCoder{}
	driver.typeDecoder[types.TimeStampeLTZ] = &type_coder.DateCoder{}
	driver.typeDecoder[types.TimeStampTZ_DTY] = &type_coder.DateCoder{}

	driver.typeDecoder[types.ROWID] = &type_coder.RowIDCoder{}
	driver.typeDecoder[types.UROWID] = &type_coder.RowIDCoder{}

	driver.typeDecoder[types.BOOLEAN] = &type_coder.BoolCoder{}

	driver.typeDecoder[types.VECTOR] = type_coder.NewVectorDecoder()
	driver.typeDecoder[types.OCIBlobLocator] = type_coder.NewBlobDecoder()
	driver.typeDecoder[types.OCIClobLocator] = type_coder.NewClobDecoder()
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

func RegisterTypeDecoder(db GetDriverInterface, oracleType uint16, decoder type_coder.OracleTypeDecoder) error {
	if drv, ok := db.Driver().(*OracleDriver); ok {
		drv.mu.Lock()
		drv.typeDecoder[oracleType] = decoder
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

func RegisterType(conn *sql.DB, typeName, arrayTypeName string, typeObj interface{}) error {
	// ping first to avoid error when calling register type after open connection
	err := conn.Ping()
	if err != nil {
		return err
	}

	if drv, ok := conn.Driver().(*OracleDriver); ok {
		return RegisterTypeWithOwner(conn, drv.UserId, typeName, arrayTypeName, typeObj)
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

	cust := customType{
		owner: owner,
		name:  typeName,
		// arrayTypeName: arrayTypeName,
		typ: typ,
	}
	arrayCust := customType{owner: owner, name: arrayTypeName, isArray: true}
	var err error
	arrayParam := ParameterInfo{
		Direction: Input,
		TypeInfo: type_coder.TypeInfo{
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
			param := ParameterInfo{Direction: Input, TypeInfo: type_coder.TypeInfo{Flag: 3}}
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
						param.cusType = new(customType)
						*param.cusType = value
						param.ToID = value.toid
						break
					}
					// if strings.EqualFold(value.arrayTypeName, attTypeName.String) {
					// 	found = true
					// 	param.cusType = new(customType)
					// 	param.DataType = XMLType
					// 	*param.cusType = value
					// 	param.cusType.isArray = true
					// 	param.ToID = value.arrayTOID
					// 	break
					// }
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
		arrayParam.cusType = new(customType)
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
