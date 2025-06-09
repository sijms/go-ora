package go_ora

import (
	"database/sql/driver"
	"encoding/json"
	"github.com/sijms/go-ora/v2/converters"
)

type ValueEncoder interface {
	EncodeValue(param *ParameterInfo, connection *Connection) error
}

type (
	PLBool   bool
	NVarChar string
	batch    struct {
		array driver.Value
	}
)

func NewBatch(array driver.Value) driver.Value {
	return &batch{array: array}
}
func (val PLBool) SetDataType(_ *Connection, par *ParameterInfo) error {
	par.DataType = Boolean
	par.MaxLen = converters.MAX_LEN_BOOL
	return nil
}

func (val NVarChar) SetDataType(conn *Connection, par *ParameterInfo) error {
	par.DataType = NCHAR
	par.CharsetForm = 2
	par.ContFlag = 16
	par.CharsetID = conn.tcpNego.ServernCharset
	return nil
}

func (val NullNVarChar) SetDataType(conn *Connection, par *ParameterInfo) error {
	par.DataType = NCHAR
	par.CharsetForm = 2
	par.ContFlag = 16
	par.CharsetID = conn.tcpNego.ServernCharset
	return nil
}

func (val TimeStamp) SetDataType(_ *Connection, par *ParameterInfo) error {
	par.DataType = TIMESTAMP
	par.MaxLen = converters.MAX_LEN_TIMESTAMP
	return nil
}
func (val NullTimeStamp) SetDataType(_ *Connection, par *ParameterInfo) error {
	par.DataType = TIMESTAMP
	par.MaxLen = converters.MAX_LEN_TIMESTAMP
	return nil
}
func (val TimeStampTZ) SetDataType(_ *Connection, par *ParameterInfo) error {
	par.DataType = TimeStampTZ_DTY
	par.MaxLen = converters.MAX_LEN_TIMESTAMP
	return nil
}
func (val NullTimeStampTZ) SetDataType(_ *Connection, par *ParameterInfo) error {
	par.DataType = TimeStampTZ_DTY
	par.MaxLen = converters.MAX_LEN_TIMESTAMP
	return nil
}
func (val *NVarChar) Value() (driver.Value, error) {
	return val, nil
}

func (val *NVarChar) Scan(value interface{}) error {
	*val = NVarChar(getString(value))
	return nil
}

func (val NVarChar) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(val))
}

func (val *NVarChar) UnmarshalJSON(data []byte) error {
	var temp string
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	*val = NVarChar(temp)
	return nil
}

//func (val *NVarChar) ValueDecoder(buffer []byte) error {
//
//}
//func (val *NVarChar) EncodeValue(param *ParameterInfo, connection *Connection) error {
//	strVal := string(*val)
//	param.DataType = NCHAR
//	param.CharsetForm = 2
//	param.ContFlag = 0x10
//	param.CharsetID = connection.tcpNego.ServernCharset
//	param.MaxCharLen = len([]rune(strVal))
//	if len(string(*val)) == 0 {
//		param.BValue = nil
//	}
//	if param.CharsetID != connection.strConv.GetLangID() {
//		tempCharset := connection.strConv.SetLangID(param.CharsetID)
//		defer connection.strConv.SetLangID(tempCharset)
//		param.BValue = connection.strConv.Encode(strVal)
//	} else {
//		param.BValue = connection.strConv.Encode(strVal)
//	}
//	if param.MaxLen < len(param.BValue) {
//		param.MaxLen = len(param.BValue)
//	}
//	if param.BValue == nil && param.MaxLen == 0 {
//		param.MaxLen = 1
//	}
//	if (param.Direction == Output && param.MaxLen == 0) || param.MaxLen > converters.MAX_LEN_NVARCHAR2 {
//		param.MaxLen = converters.MAX_LEN_NVARCHAR2
//	}
//	return nil
//}
//func (val *TimeStamp) EncodeValue(param *ParameterInfo, _ *Connection) error {
//	param.setForTime()
//	param.DataType = TIMESTAMP
//	param.BValue = converters.EncodeTimeStamp(time.Time(*val))
//	return nil
//}

type NullNVarChar struct {
	NVarChar NVarChar
	Valid    bool
}

func (val NullNVarChar) Value() (driver.Value, error) {
	if val.Valid {
		return val.NVarChar.Value()
	} else {
		return nil, nil
	}
}

func (val *NullNVarChar) Scan(value interface{}) error {
	if value == nil {
		val.Valid = false
		return nil
	}
	val.Valid = true
	return val.NVarChar.Scan(value)
}

func (val NullNVarChar) MarshalJSON() ([]byte, error) {
	if val.Valid {
		return json.Marshal(string(val.NVarChar))
	}
	return json.Marshal(nil)
}

func (val *NullNVarChar) UnmarshalJSON(data []byte) error {
	temp := new(string)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return err
	}
	if temp == nil {
		val.Valid = false
	} else {
		val.Valid = true
		val.NVarChar = NVarChar(*temp)
	}
	return nil
}

func (val NClob) MarshalJSON() ([]byte, error) {
	if val.Valid {
		return json.Marshal(val.String)
	}
	return json.Marshal(nil)
}

func (val *NClob) UnmarshalJSON(data []byte) error {
	temp := new(string)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return err
	}
	if temp == nil {
		val.Valid = false
	} else {
		val.Valid = true
		val.String = *temp
	}
	return nil
}
func (val Clob) SetDataType(conn *Connection, par *ParameterInfo) error {
	par.DataType = OCIClobLocator
	par.CharsetForm = 1
	par.CharsetID = conn.getDefaultCharsetID()
	return nil
}
func (val NClob) SetDataType(conn *Connection, par *ParameterInfo) error {
	par.DataType = OCIClobLocator
	par.CharsetForm = 2
	par.CharsetID = conn.tcpNego.ServernCharset
	return nil
}
func (val Blob) SetDataType(conn *Connection, par *ParameterInfo) error {
	par.DataType = OCIBlobLocator
	return nil
}

func (val Clob) MarshalJSON() ([]byte, error) {
	if val.Valid {
		return json.Marshal(val.String)
	}
	return json.Marshal(nil)
}

func (val *Clob) UnmarshalJSON(data []byte) error {
	temp := new(string)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return err
	}
	if temp == nil {
		val.Valid = false
	} else {
		val.Valid = true
		val.String = *temp
	}
	return nil
}

func (val Blob) MarshalJSON() ([]byte, error) {
	return json.Marshal(val.Data)
}

func (val *Blob) UnmarshalJSON(data []byte) error {
	temp := new([]byte)
	err := json.Unmarshal(data, temp)
	if err != nil {
		return err
	}
	val.Data = *temp
	return nil
}
