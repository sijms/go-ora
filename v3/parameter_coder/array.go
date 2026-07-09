package parameter_coder

import (
	"database/sql/driver"
	"fmt"
	"reflect"

	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
	"github.com/sijms/go-ora/v3/utils"
)

type ArrayParameter struct {
	BasicParameter
	attrib OracleParameterCoder
	//items  []parameter_coder.OracleParameterCoder
}

func (param *ArrayParameter) Copy() OracleParameterCoder {
	ret := new(ArrayParameter)
	*ret = *param
	return ret
}

func (param *ArrayParameter) Init() {
	param.SetDefault()
	param.Flag = 0x43
}

func (param *ArrayParameter) FillAttrib(input interface{}, conn IConnection) error {
	var err error
	if param.attrib == nil {
		inputType := utils.GetType(input)
		if inputType != nil {
			itemType := inputType.Elem()
			if itemType != nil {
				for itemType.Kind() == reflect.Ptr {
					itemType = itemType.Elem()
				}
			}
			param.attrib, err = conn.GetParameterCoder(itemType)
		} else if param.DataType != 0 {
			param.attrib, err = conn.GetParameterCoder(param.DataType)
		} else {
			err = fmt.Errorf("to set array item type either pass non-nil value or pass data type")
		}

	}
	return err
}

func (param *ArrayParameter) Encode(input interface{}, conn IConnection) (err error) {
	param.Init()
	inputType := utils.GetType(input)
	if inputType.Kind() == reflect.Array || inputType.Kind() == reflect.Slice {
		var coder OracleParameterCoder
		err = param.FillAttrib(input, conn)
		if err != nil {
			return err
		}
		coder = param.attrib.Copy()
		coder.Init()
		param.SetParameterInfo(coder.GetParameterInfo())
		rValue := reflect.ValueOf(input)
		length := rValue.Len()
		if length > param.ArraySize {
			param.ArraySize = length
		}
		if param.ArraySize > 0 {
			session := network.NewMemorySession(nil, nil, conn.GetSession().GetProperties())
			session.PutUint(param.ArraySize, 4, true, true)
			if length > 0 {
				for i := 0; i < length; i++ {
					var item driver.Value
					if rValue.Index(i).CanInterface() {
						item, err = utils.GetValue(rValue.Index(i).Interface())
						if err != nil {
							return err
						}
						err = coder.Encode(item, conn)
					} else {
						err = coder.Encode(nil, conn)
					}
					if err != nil {
						return err
					}
					err = coder.Write(session)
					if err != nil {
						return err
					}

				}
			} else {
				// encode default value of the array parameter to fill parameter properties
				err = coder.Encode(nil, conn)
				if err != nil {
					return err
				}
			}
			param.SetParameterInfo(coder.GetParameterInfo())
			if param.DataType == types.NCHAR {
				param.MaxLen = conn.GetMaxStringLength()
				param.MaxCharLen = param.MaxLen // / converters.MaxBytePerChar(par.CharsetID)
			}
			if param.DataType == types.RAW {
				param.MaxLen = conn.GetMaxRawLength()
			}
			param.BValue = session.GetWriteBuffer()
		} else {
			param.BValue = nil
		}
		if param.ArraySize == 0 {
			param.ArraySize = 1
		}

	} else {
		err = fmt.Errorf("no encoder register for data type: %s", inputType.Name())
	}
	return
}

func (param *ArrayParameter) Decode(conn IConnection) (value interface{}, err error) {
	err = param.FillAttrib(nil, conn)
	if err != nil {
		return nil, err
	}
	session := conn.GetSession()
	if param.ArraySize > 0 {
		//decoders := make([]parameter_coder.OracleParameterCoder, param.ArraySize)
		items := make([]interface{}, param.ArraySize)
		for i := 0; i < param.ArraySize; i++ {
			decoder := param.attrib.Copy()
			decoder.SetParameterInfo(param.GetParameterInfo())
			err = decoder.Read(session)
			if err != nil {
				return nil, err
			}
			_, err = session.GetInt(2, true, true)
			if err != nil {
				return nil, err
			}
			items[i], err = decoder.Decode(conn)
			if err != nil {
				return
			}
		}
		return items, nil
	}
	return nil, nil
}

func (param *ArrayParameter) Read(session network.SessionReader) error {
	// non-xml array
	// this code should be in basic_read
	var err error
	param.ArraySize, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	//err = param.FillAttrib(nil, )
	//if err != nil {
	//	return err
	//}

	return nil
}

func (param *ArrayParameter) Write(session network.SessionWriter) (err error) {
	if len(param.BValue) == 0 {
		session.PutBytes(0)
	} else {
		session.PutBytes(param.BValue...)
	}
	return
}
