package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type BasicParameter struct {
	Flag         uint8
	ContFlag     int
	DataType     uint16
	CharsetID    int
	CharsetForm  int
	MaxLen       int64
	MaxCharLen   int64
	ToID         []byte
	ArraySize    int
	Coder        converters.StringCoder
	IsUDTPar     bool
	BValue       []byte
	VectorDim    int
	VectorFormat uint8
	VectorFlag   uint8
	VectorType   types.VectorType
}

func (basic *BasicParameter) basicRead(session network.SessionReader) ([]byte, error) {
	if (basic.DataType == types.NCHAR || basic.DataType == types.CHAR) && basic.MaxLen == 0 {
		return nil, nil
	}
	if basic.DataType == types.RAW && basic.MaxLen == 0 {
		return nil, nil
	}
	var err error
	var bValue []byte
	if basic.IsUDTPar {
		bValue, err = session.GetFixedClr()
	} else {
		bValue, err = session.GetClr()
	}
	return bValue, err
}

func (basic *BasicParameter) SetDefault() {
	basic.DataType = 0
	basic.Flag = 3
	basic.ContFlag = 0
	basic.CharsetID = 0
	basic.CharsetForm = 0
	basic.MaxLen = 1
	basic.MaxCharLen = 0
	basic.ArraySize = 0
	basic.BValue = nil
	basic.VectorDim = 0
	basic.VectorFormat = 0
	basic.VectorFlag = 0
	basic.VectorType = 0
}

func (basic *BasicParameter) GetParameterInfo() BasicParameter {
	return *basic
}

func (basic *BasicParameter) SetParameterInfo(data BasicParameter) {
	*basic = data
}

func (basic *BasicParameter) SetLobStreamer(lobStreamer types.LobStreamer) {

}
