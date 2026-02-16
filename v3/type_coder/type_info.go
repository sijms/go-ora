package type_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type TypeInfo struct {
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

func (info *TypeInfo) GetTypeInfo() TypeInfo {
	return *info
}
func (info *TypeInfo) SetTypeInfo(data TypeInfo) {
	*info = data
}

func (info *TypeInfo) basicRead(session network.SessionReader) ([]byte, error) {
	if (info.DataType == types.NCHAR || info.DataType == types.CHAR) && info.MaxCharLen == 0 {
		return nil, nil
	}
	if info.DataType == types.RAW && info.MaxLen == 0 {
		return nil, nil
	}
	var err error
	var bValue []byte
	if info.IsUDTPar {
		bValue, err = session.GetFixedClr()
	} else {
		bValue, err = session.GetClr()
	}
	if err != nil {
		return nil, err
	}
	return bValue, nil
}

func (info *TypeInfo) Encode(_ converters.StringCoder, _ types.LobStreamer) error {
	return nil
}

func (info *TypeInfo) SetDefault() {
	info.Flag = 3
	info.MaxLen = 1
}

func (info *TypeInfo) SetLobStreamer(lobStreamer types.LobStreamer) {

}

func (info *TypeInfo) SetCharsetCoder(strCoder converters.StringCoder) {
	info.Coder = strCoder
}
