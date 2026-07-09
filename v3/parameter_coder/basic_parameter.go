package parameter_coder

import (
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type BasicParameter struct {
	TypeName    string
	Flag        uint8
	ContFlag    int
	DataType    uint16
	CharsetID   int
	CharsetForm int
	MaxLen      int64
	MaxCharLen  int64
	ToID        []byte
	ArraySize   int
	//Coder        converters.StringCoder
	IsUDTPar     bool
	IsArrayPar   bool
	BValue       []byte
	Version      int
	VectorDim    int
	VectorFormat uint8
	VectorFlag   uint8
	VectorType   types.VectorType
	PSession     network.SessionReadWriter
}

func (basic *BasicParameter) Write(session network.SessionWriter) error {
	if basic.IsUDTPar {
		session.PutFixedClr(basic.BValue)
	} else {
		if basic.IsArrayPar && len(basic.BValue) == 0 {
			session.PutBytes(0xFF)
		} else {
			session.PutClr(basic.BValue)
		}
	}
	return nil
}
func (basic *BasicParameter) BasicRead(session network.SessionReader) ([]byte, error) {
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
	//basic.DataType = 0
	if basic.Flag == 0 {
		basic.Flag = 3
	}
	basic.ContFlag = 0
	basic.CharsetID = 0
	//basic.CharsetForm = 0
	if basic.MaxLen == 0 {
		basic.MaxLen = 1
	}
	basic.BValue = nil
	basic.VectorDim = 0
	basic.VectorFormat = 0
	basic.VectorFlag = 0
	basic.VectorType = 0
}

func (basic *BasicParameter) Bytes() []byte {
	return basic.BValue
}
func (basic *BasicParameter) GetParameterInfo() BasicParameter {
	return *basic
}

func (basic *BasicParameter) SetParameterInfo(data BasicParameter) {
	if len(basic.TypeName) == 0 {
		basic.TypeName = data.TypeName
	}
	if basic.Flag == 0 {
		basic.Flag = data.Flag
	}
	basic.ContFlag = data.ContFlag
	if basic.DataType == 0 {
		basic.DataType = data.DataType
	}
	basic.CharsetID = data.CharsetID
	if data.CharsetForm != 0 {
		basic.CharsetForm = data.CharsetForm
	}
	if basic.MaxLen < data.MaxLen {
		basic.MaxLen = data.MaxLen
	}
	if basic.MaxCharLen < data.MaxCharLen {
		basic.MaxCharLen = data.MaxCharLen
	}
	if len(basic.ToID) == 0 {
		basic.ToID = data.ToID
	}
	if basic.ArraySize < data.ArraySize {
		basic.ArraySize = data.ArraySize
	}
	basic.IsUDTPar = data.IsUDTPar
	basic.BValue = data.BValue
	basic.VectorDim = data.VectorDim
	basic.VectorFormat = data.VectorFormat
	basic.VectorFlag = data.VectorFlag
	basic.VectorType = data.VectorType
	//*basic = data
}

// func (basic *BasicParameter) UpdateParameterInfo() {}
func (basic *BasicParameter) SetLobStreamer(lobStreamer types.LobStreamer) {

}

func (basic *BasicParameter) SetAsUDTPar() {
	basic.IsUDTPar = true
}

func (basic *BasicParameter) SetAsArrayPar() {
	basic.IsArrayPar = true
}

func (basic *BasicParameter) SetAQMessage() {

}
func (basic *BasicParameter) SetParentSession(input network.SessionReadWriter) {
	basic.PSession = input
}

func (basic *BasicParameter) SetBytes(data []byte) {
	basic.BValue = data
}

//func (basic *BasicParameter) IsChild() bool {
//	return basic.isChild
//}
