package types

import (
	"context"
	"database/sql"
	"database/sql/driver"

	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/trace"
)

type LobStreamer interface {
	StartContext(ctx context.Context) chan struct{}
	EndContext(done chan struct{})
	GetLocator() Locator
	SetLocator(locator Locator)
	DatabaseVersionNumber() int
	GetStringCoder() converters.StringCoder
	GetLobStreamMode() configurations.LobFetch
	GetLobReadMode() configurations.LobReadMode
	GetTracer() trace.Tracer
	GetSize() (int64, error)
	Exists() (bool, error)
	CreateTemporaryLocator(charsetID, charsetForm int) (Locator, error)
	FreeTemporaryLocator() error
	Open(mode, opID int) error
	Read(offset, count int64) ([]byte, error)
	Write(data []byte) error
	Close(opID int) error
}

type OracleType interface {
	SetValue(input interface{}) error
	Value() (interface{}, error)
	CopyTo(dest driver.Value) error
	SetDataType(dt uint16)
	GetDataType() uint16
	GetMaxLen() int64
	sql.Scanner
}

//type LobTyper interface {
//	SetLobStream(stream LobStreamer)
//}
//
//type StringTyper interface{}
