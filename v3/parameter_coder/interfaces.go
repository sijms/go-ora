package parameter_coder

import (
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type (
	OracleParameterCoder interface {
		OracleParameterEncoder
		OracleParameterDecoder
		Bytes() []byte
	}
	OracleParameterEncoder interface {
		Encode(input interface{}, strConv converters.StringCoder, _ types.LobStreamer) error
		network.ValueStreamWriter
		GetParameterInfo() BasicParameter
		SetParameterInfo(data BasicParameter)
	}

	OracleParameterDecoder interface {
		Read(session network.SessionReader) error
		Decode(strConv converters.StringCoder) (interface{}, error)
		SetLobStreamer(lobStreamer types.LobStreamer)
		SetParameterInfo(data BasicParameter)
	}
)
