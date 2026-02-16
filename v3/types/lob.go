package types

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sijms/go-ora/v3/configurations"
)

type Lob interface {
	GetLocator() []byte
	Free() error
	GetLength() (int64, error)
	Read(ctx context.Context) error
	GetReadMode() configurations.LobReadMode
	sql.Scanner
}

var errEmptyLocator = errors.New("calling Lob function: on nil or non allocated object")
var errNilStreamer = errors.New("calling Lob function: nil streamer")

type lobBase struct {
	//Locator []byte
	stream LobStreamer
}

//	func (l *lobBase) NewQuasiLocator(dataLen uint64) {
//		l.Locator = make([]byte, 0x28)
//		l.Locator[1] = 38
//		l.Locator[3] = 4
//		l.Locator[4] = 97
//		l.Locator[5] = 8
//		l.Locator[9] = 1
//		binary.BigEndian.PutUint64(l.Locator[10:], dataLen)
//	}

func (l *lobBase) GetLocator() []byte {
	if l.IsNil() {
		return nil
	}
	return l.stream.GetLocator()
}
func (l *lobBase) IsNil() bool {
	return l.stream == nil || l.stream.GetLocator() == nil
}
func (l *lobBase) IsQuasi() bool {
	if l.IsNil() {
		return true
	}
	locator := l.stream.GetLocator()
	return len(locator) > 3 && locator[3] == 4
}
func (l *lobBase) IsValueBased() bool {
	if !l.IsNil() {
		locator := l.stream.GetLocator()
		return len(locator) > 4 && locator[4]&0x20 == 0x20
	}
	return false
}
func (l *lobBase) IsTemporary() bool {
	if !l.IsNil() {
		locator := l.stream.GetLocator()
		return len(locator) > 7 && locator[7]&1 == 1 || locator[4]&0x40 == 0x40 || l.IsValueBased()
	}
	return false
}

func (l *lobBase) IsVarWidthChar() bool {
	if !l.IsNil() {
		locator := l.stream.GetLocator()
		return len(locator) > 6 && locator[6]&0x80 == 0x80
	}
	return false
}

func (l *lobBase) IsLittleEndian() bool {
	if !l.IsNil() {
		locator := l.stream.GetLocator()
		return len(locator) > 7 && locator[7]&0x40 == 0x40
	}
	return false
}

func (l *lobBase) IsReadOnly() bool {
	if !l.IsNil() {
		locator := l.stream.GetLocator()
		return len(locator) > 6 && locator[6]&1 == 1
	}
	return false
}

func (l *lobBase) Free() error {
	if !l.IsQuasi() {
		err := l.stream.FreeTemporaryLocator()
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *lobBase) GetLength() (int64, error) {
	if !l.IsQuasi() {
		return l.stream.GetSize()
	}
	return 0, nil
}

func (l *lobBase) GetReadMode() configurations.LobReadMode {
	if !l.IsQuasi() {
		return l.stream.GetLobReadMode()
	}
	return configurations.LobReadMode_NONE
}

//	func (l *lobBase) Read(ctx context.Context) ([]byte, error) {
//		return l.ReadBytesFromPos(ctx, 0, 0)
//	}
func (l *lobBase) ReadFromPos(ctx context.Context, pos int64) ([]byte, error) {
	return l.ReadBytesFromPos(ctx, pos, 0)
}

func (l *lobBase) ReadBytesFromPos(ctx context.Context, pos, count int64) ([]byte, error) {
	if !l.IsQuasi() {
		done := l.stream.StartContext(ctx)
		defer l.stream.EndContext(done)
		return l.stream.Read(pos, count)
	}
	return nil, errEmptyLocator
}

func (l *lobBase) copyFrom(input *lobBase) error {
	if input.stream != nil && input.stream.GetLocator() != nil {
		if l.IsQuasi() {
			l.stream = input.stream
			return nil
		}
		if bytes.Compare(l.GetLocator(), input.GetLocator()) == 0 {
			return nil
		}
		return fmt.Errorf("the source lob locator is not empty so copy is not permitted, please free the current lob first")
	}
	return nil
}
