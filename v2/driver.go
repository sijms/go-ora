package go_ora

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/sijms/go-ora/v2/advanced_nego"
	"github.com/sijms/go-ora/v2/configurations"
	"reflect"
	"strings"
	"sync"

	"github.com/sijms/go-ora/v2/converters"
)

type OracleDriver struct {
	dataCollected bool
	cusTyp        map[string]customType
	sessionParam  map[string]string
	mu            sync.Mutex
	sStrConv      converters.IStringConverter
	nStrConv      converters.IStringConverter
	UserId        string
	connOption    *configurations.ConnectionConfig
	//Server    string
	//Port      int
	//Instance  string
	//Service   string
	//DBName    string

	//SessionId int
	//SerialNum int
}

var oracleDriver = &OracleDriver{
	cusTyp:       map[string]customType{},
	sessionParam: map[string]string{},
}

type GetDriverInterface interface {
	Driver() driver.Driver
}

func init() {
	sql.Register("oracle", oracleDriver)
}

func GetDefaultDriver() *OracleDriver {
	return oracleDriver
}

func NewDriver() *OracleDriver {
	return &OracleDriver{
		cusTyp:       map[string]customType{},
		sessionParam: map[string]string{},
	}
}

func (driver *OracleDriver) init(conn *Connection) error {
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
	if driver, ok := db.Driver().(*OracleDriver); ok {
		driver.sStrConv = charset
		driver.nStrConv = nCharset
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

//func RegisterRegularTypeArray(conn *sql.DB, regularTypeName, arrayTypeName string, itemMaxSize int) error {
//	err := conn.Ping()
//	if err != nil {
//		return err
//	}
//
//	if drv, ok := conn.Driver().(*OracleDriver); ok {
//		return RegisterRegularTypeArrayWithOwner(conn, drv.UserId, regularTypeName, arrayTypeName, itemMaxSize)
//	}
//	return errors.New("the driver used is not a go-ora driver type")
//}
//func RegisterRegularTypeArrayWithOwner(conn *sql.DB, owner, regularTypeName, arrayTypeName string, itemMaxSize int) error {
//	drv := conn.Driver().(*OracleDriver)
//	regularTypeName = strings.TrimSpace(regularTypeName)
//	arrayTypeName = strings.TrimSpace(arrayTypeName)
//	if len(regularTypeName) == 0 {
//		return errors.New("typeName shouldn't be empty")
//	}
//	if len(arrayTypeName) == 0 {
//		return errors.New("array type name shouldn't be empty")
//	}
//	cust := customType{
//		owner:         owner,
//		name:          regularTypeName,
//		arrayTypeName: arrayTypeName,
//		isArray:       true,
//	}
//	var err error
//	cust.arrayTOID, err = getTOID2(conn, owner, arrayTypeName)
//	if err != nil {
//		return err
//	}
//	param := ParameterInfo{Direction: Input, Flag: 3, TypeName: regularTypeName, MaxLen: 1, MaxCharLen: 1}
//	switch strings.ToUpper(regularTypeName) {
//	case "NUMBER":
//		param.DataType = NUMBER
//	case "VARCHAR2":
//		param.DataType = NCHAR
//		param.CharsetForm = 1
//		param.ContFlag = 16
//		param.CharsetID = drv.sStrConv.GetLangID()
//	case "NVARCHAR2":
//		param.DataType = NCHAR
//		param.CharsetForm = 2
//		param.ContFlag = 16
//		param.CharsetID = drv.nStrConv.GetLangID()
//	case "TIMESTAMP":
//		param.DataType = TimeStampDTY
//	case "DATE":
//		param.DataType = DATE
//	case "TIMESTAMP WITH LOCAL TIME ZONE":
//		param.DataType = TimeStampLTZ_DTY
//	//case "TIMESTAMP WITH TIME ZONE":
//	//	param.DataType = TimeStampTZ_DTY
//	//	param.MaxLen = converters.MAX_LEN_TIMESTAMP
//	case "RAW":
//		param.DataType = RAW
//	case "BLOB":
//		param.DataType = OCIBlobLocator
//	case "CLOB":
//		param.DataType = OCIClobLocator
//		param.CharsetForm = 1
//		param.ContFlag = 16
//		param.CharsetID = drv.sStrConv.GetLangID()
//	case "NCLOB":
//		param.DataType = OCIClobLocator
//		param.CharsetForm = 2
//		param.ContFlag = 16
//		param.CharsetID = drv.nStrConv.GetLangID()
//	default:
//		return fmt.Errorf("unsupported regular type: %s", regularTypeName)
//	}
//	cust.attribs = append(cust.attribs, param)
//	drv.mu.Lock()
//	defer drv.mu.Unlock()
//	drv.cusTyp[strings.ToUpper(arrayTypeName)] = cust
//	return nil
//}

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
		//arrayTypeName: arrayTypeName,
		typ: typ,
	}
	arrayCust := customType{owner: owner, name: arrayTypeName, isArray: true}
	var err error
	arrayParam := ParameterInfo{Direction: Input, Flag: 3, TypeName: typeName, MaxLen: 1, MaxCharLen: 1}
	switch strings.ToUpper(typeName) {
	case "NUMBER":
		arrayParam.DataType = NUMBER
	case "VARCHAR2":
		arrayParam.DataType = NCHAR
		arrayParam.CharsetForm = 1
		arrayParam.ContFlag = 16
		arrayParam.CharsetID = drv.sStrConv.GetLangID()
	case "NVARCHAR2":
		arrayParam.DataType = NCHAR
		arrayParam.CharsetForm = 2
		arrayParam.ContFlag = 16
		arrayParam.CharsetID = drv.nStrConv.GetLangID()
	case "TIMESTAMP":
		arrayParam.DataType = TimeStampDTY
	case "DATE":
		arrayParam.DataType = DATE
	case "TIMESTAMP WITH LOCAL TIME ZONE":
		arrayParam.DataType = TimeStampLTZ_DTY
	//case "TIMESTAMP WITH TIME ZONE":
	//	arrayParam.DataType = TimeStampTZ_DTY
	//	arrayParam.MaxLen = converters.MAX_LEN_TIMESTAMP
	case "RAW":
		arrayParam.DataType = RAW
	case "BLOB":
		arrayParam.DataType = OCIBlobLocator
	case "CLOB":
		arrayParam.DataType = OCIClobLocator
		arrayParam.CharsetForm = 1
		arrayParam.ContFlag = 16
		arrayParam.CharsetID = drv.sStrConv.GetLangID()
	case "NCLOB":
		arrayParam.DataType = OCIClobLocator
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
			param := ParameterInfo{Direction: Input, Flag: 3}
			param.Name = attName.String
			param.TypeName = attTypeName.String
			switch strings.ToUpper(attTypeName.String) {
			case "NUMBER":
				param.DataType = NUMBER
				param.MaxLen = converters.MAX_LEN_NUMBER
			case "VARCHAR2":
				param.DataType = NCHAR
				param.CharsetForm = 1
				param.ContFlag = 16
				param.MaxCharLen = int(length.Int64)
				param.CharsetID = drv.sStrConv.GetLangID()
				param.MaxLen = int(length.Int64) * converters.MaxBytePerChar(param.CharsetID)
			case "NVARCHAR2":
				param.DataType = NCHAR
				param.CharsetForm = 2
				param.ContFlag = 16
				param.MaxCharLen = int(length.Int64)
				param.CharsetID = drv.nStrConv.GetLangID()
				param.MaxLen = int(length.Int64) * converters.MaxBytePerChar(param.CharsetID)
			case "TIMESTAMP":
				param.DataType = TimeStampDTY
				param.MaxLen = converters.MAX_LEN_DATE
			case "DATE":
				param.DataType = DATE
				param.MaxLen = converters.MAX_LEN_DATE
			case "TIMESTAMP WITH LOCAL TIME ZONE":
				param.DataType = TimeStampLTZ_DTY
				param.MaxLen = converters.MAX_LEN_DATE
			case "TIMESTAMP WITH TIME ZONE":
				param.DataType = TimeStampTZ_DTY
				param.MaxLen = converters.MAX_LEN_TIMESTAMP
			case "RAW":
				param.DataType = RAW
				param.MaxLen = int(length.Int64)
			case "BLOB":
				param.DataType = OCIBlobLocator
				param.MaxLen = int(length.Int64)
			case "CLOB":
				param.DataType = OCIClobLocator
				param.CharsetForm = 1
				param.ContFlag = 16
				param.CharsetID = drv.sStrConv.GetLangID()
				param.MaxCharLen = int(length.Int64)
				param.MaxLen = int(length.Int64) * converters.MaxBytePerChar(param.CharsetID)
			case "NCLOB":
				param.DataType = OCIClobLocator
				param.CharsetForm = 2
				param.ContFlag = 16
				param.CharsetID = drv.nStrConv.GetLangID()
				param.MaxCharLen = int(length.Int64)
				param.MaxLen = int(length.Int64) * converters.MaxBytePerChar(param.CharsetID)
			default:
				found := false
				for name, value := range drv.cusTyp {
					if strings.EqualFold(name, attTypeName.String) {
						found = true
						param.DataType = XMLType
						param.cusType = new(customType)
						*param.cusType = value
						param.ToID = value.toid
						break
					}
					//if strings.EqualFold(value.arrayTypeName, attTypeName.String) {
					//	found = true
					//	param.cusType = new(customType)
					//	param.DataType = XMLType
					//	*param.cusType = value
					//	param.cusType.isArray = true
					//	param.ToID = value.arrayTOID
					//	break
					//}
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
		arrayParam.DataType = XMLType
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
	//connStr, err := newConnectionStringFromUrl(dsn)
	//if err != nil {
	//	return nil, err
	//}
	//return &connStr.connOption, nil
}

func RegisterConnConfig(config *configurations.ConnectionConfig) {
	oracleDriver.connOption = config

}
