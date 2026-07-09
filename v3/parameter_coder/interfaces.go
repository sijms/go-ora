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
	//GetGoParameterCoder(goType reflect.Type) OracleParameterCoder
	//GetOracleParameterCoder(oracleType uint16) OracleParameterCoder
	//GetNameParameterCoder(nameType string) OracleParameterCoder
	GetParameterCoder(input interface{}) (OracleParameterCoder, error)
	SendTimeZoneAsUTC() bool
	GetMaxRawLength() int64
}

type (
	OracleParameterCoder interface {
		OracleParameterEncoder
		OracleParameterDecoder
		Init()
		Bytes() []byte
		Copy() OracleParameterCoder
		SetAsUDTPar()
		SetAsArrayPar()
		SetAQMessage()
		SetParentSession(input network.SessionReadWriter)
		//SetChild(bool)
		//IsChild() bool
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
		SetBytes(data []byte)
	}
)
