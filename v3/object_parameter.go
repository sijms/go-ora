package go_ora

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/parameter_coder"
	"github.com/sijms/go-ora/v3/types"
)

type ObjectParameter struct {
	parameter_coder.BasicParameter
	//lobParameter
	//children []parameter_coder.OracleParameterCoder
	obj Object
}

func (param *ObjectParameter) Copy() parameter_coder.OracleParameterCoder {
	ret := new(ObjectParameter)
	*ret = *param
	return ret
}

func (param *ObjectParameter) Encode(input interface{}, conn parameter_coder.IConnection) (err error) {
	param.SetDefault()
	param.TypeName = param.obj.Name
	param.DataType = types.XMLType
	if input == nil {
		if param.obj.isArray {
			param.BValue = []byte{0xFF}
		} else {
			param.BValue = []byte{0xFD}
		}
		return
	}
	param.MaxLen = 2000
	// loop for all parameters
	inputType := reflect.TypeOf(input)
	for inputType.Kind() == reflect.Ptr {
		inputType = inputType.Elem()
	}
	// this code is work for object struct
	if inputType != param.obj.typ {
		return fmt.Errorf("wrong parameter mapping for type %s", inputType.String())
	}
	rValue := reflect.ValueOf(input)
	//fieldsBuffer := &bytes.Buffer{}
	fieldsBuffer := network.NewMemorySession(nil, nil, conn.GetSession().GetProperties())
	// call encode
	// use ordered field
	for _, field := range param.obj.fields {
		attrib := param.obj.attribs[field]
		if idx, ok := param.obj.activeFields[field]; ok {
			field := rValue.Field(idx)
			if field.CanInterface() {
				err = attrib.Encode(field.Interface(), conn)
			} else {
				err = attrib.Encode(nil, conn)
			}
		} else {
			err = attrib.Encode(nil, conn)
		}
		if err != nil {
			return
		}
		if attrib.GetParameterInfo().DataType == types.XMLType {
			fieldsBuffer.PutBytes(attrib.Bytes()...)
		} else {
			fieldsBuffer.PutFixedClr(attrib.Bytes())
		}
	}
	// save BValue to buffer
	param.BValue = fieldsBuffer.Write(true)
	return
}
func (param *ObjectParameter) encapsulate() (err error) {
	// the following code is for top level object only
	size := len(param.BValue)
	objectBuffer := &bytes.Buffer{}
	objectBuffer.Write([]byte{0x84, 0x1})
	if (size + 7) < 0xfe {
		size += 3
		objectBuffer.Write([]byte{uint8(size)})
	} else {
		size += 7
		objectBuffer.Write([]byte{0xfe})
		err = binary.Write(objectBuffer, binary.BigEndian, uint32(size))
		if err != nil {
			return
		}
	}
	// at the end save buffer
	objectBuffer.Write(param.BValue)
	param.BValue = objectBuffer.Bytes()
	return
}
func (param *ObjectParameter) Decode(conn parameter_coder.IConnection) (interface{}, error) {
	// need main session to get its properties
	session := network.NewMemorySession(param.BValue, nil, conn.GetSession().GetProperties())
	objectType, err := session.GetByte()
	if err != nil {
		return nil, err
	}
	ctl, err := session.GetInt64(4, true, true)
	if err != nil {
		return nil, err
	}
	if ctl == 0xFE {
		_, err = session.GetInt(4, false, true)
		if err != nil {
			return nil, err
		}
	}
	switch objectType {
	case 0x88:
		return nil, nil
	case 0x85:
		return nil, nil
	case 0x84:
		// used predefined attribs
		retObj := reflect.New(param.obj.typ)
		var value interface{}
		for _, field := range param.obj.fields {
			attrib := param.obj.attribs[field]
			temp, err := session.Peek()
			if err != nil {
				return nil, err
			}
			if temp == 0xFD || temp == 0xFF {
				_, err = session.GetByte()
				if err != nil {
					return nil, err
				}
				value = nil
			} else {
				err = attrib.Read(session)
				if err != nil {
					return nil, err
				}
				value, err = attrib.Decode(conn)
				if err != nil {
					return nil, err
				}
			}
			if idx, ok := param.obj.activeFields[field]; ok {
				err = types.RCopy(retObj.Elem().Field(idx), value)
				if err != nil {
					return nil, err
				}
			}
		}
		return retObj.Elem().Interface(), nil
	default:
		return nil, fmt.Errorf("unknown object type: %v", objectType)
	}

}

func (param *ObjectParameter) Read(session network.SessionReader) error {
	var err error
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
		param.BValue = nil
		return nil
	}
	param.BValue, err = param.BasicRead(session)
	return err
}

func (param *ObjectParameter) Write(session network.SessionWriter) error {
	err := param.encapsulate()
	if err != nil {
		return err
	}
	session.PutBytes(0, 0, 0, 0)
	session.PutUint(len(param.BValue), 4, true, true)
	session.PutBytes(1, 1)
	session.PutClr(param.BValue)
	return nil

}
