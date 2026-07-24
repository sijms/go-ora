package aq

import (
	"time"

	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/parameter_coder"
	"github.com/sijms/go-ora/v3/types"
)

type (
	IConnection interface {
		converters.StringCoder
		GetSession() network.SessionReadWriter
		NewLobStreamer() types.LobStreamer
		ProcessTCCResponse(msgCode uint8) error
		//GetGoParameterCoder(goType reflect.Type) OracleParameterCoder
		//GetOracleParameterCoder(oracleType uint16) OracleParameterCoder
		//GetNameParameterCoder(nameType string) OracleParameterCoder
		GetParameterCoder(input interface{}) (parameter_coder.OracleParameterCoder, error)
		SendTimeAsUTC() bool
		GetDBTimeZone() *time.Location
		GetDBServerTimeZone() *time.Location
		TTCVersion() uint8
		GetMaxRawLength() int64
	}
)
