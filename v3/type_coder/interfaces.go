package type_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type (
	OracleTyperCoder interface {
		OracleTypeEncoder
		OracleTypeDecoder
	}
	OracleTypeEncoder interface {
		Encode(coder converters.StringCoder, lobStreamer types.LobStreamer) error
		network.ValueStreamWriter
		GetTypeInfo() TypeInfo
		SetTypeInfo(data TypeInfo)
		//Encode() ([]byte, error)
	}
	OracleTypeDecoder interface {
		network.ValueStreamReader
		SetTypeInfo(data TypeInfo)
		Decode(data []byte) (interface{}, error)
		SetLobStreamer(lobStreamer types.LobStreamer)
		SetCharsetCoder(strCoder converters.StringCoder)
	}
)
