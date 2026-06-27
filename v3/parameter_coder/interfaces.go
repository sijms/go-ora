package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type IConnection interface {
	converters.StringCoder
	GetSession() network.SessionReadWriter
	NewLobStreamer() types.LobStreamer
}

type (
	OracleParameterCoder interface {
		OracleParameterEncoder
		OracleParameterDecoder
		Bytes() []byte
		Copy() OracleParameterCoder
		SetAsUDTPar()
	}
	OracleParameterEncoder interface {
		Encode(input interface{}, conn IConnection) error
		network.ValueStreamWriter
		GetParameterInfo() BasicParameter
		SetParameterInfo(data BasicParameter)
	}

	OracleParameterDecoder interface {
		Read(session network.SessionReader) error
		Decode(conn IConnection) (interface{}, error)
		SetLobStreamer(lobStreamer types.LobStreamer)
		SetParameterInfo(data BasicParameter)
	}
)
