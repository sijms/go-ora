package types

import (
	"context"
	"database/sql"
	"errors"

	"github.com/sijms/go-ora/v3/configurations"
)

type Lob interface {
	GetLocator() Locator
	Free() error
	GetLength() (int64, error)
	Read(ctx context.Context) error
	GetReadMode() configurations.LobReadMode
	sql.Scanner
}

var errEmptyLocator = errors.New("calling Lob function: on nil or non allocated object")
var errNilStreamer = errors.New("calling Lob function: nil streamer")

type lobBase struct {
	stream    LobStreamer
	UploadCtx context.Context
}

func (l *lobBase) createStreamer(db *sql.DB) error {
	var err error
	if l.stream == nil {
		_, err = db.ExecContext(context.Background(), "--CREATE-LOB-STREAM--", &l.stream)
	}
	return err
}
func (l *lobBase) GetStreamer() LobStreamer {
	return l.stream
}
func (l *lobBase) SetStreamer(input LobStreamer) {
	l.stream = input
}
func (l *lobBase) GetLocator() Locator {
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
	return l.stream.GetLocator().IsQuasi()
}
func (l *lobBase) IsDataUploaded() bool {
	return l.stream != nil && l.stream.GetLocator() != nil
}

func (lob *lobBase) uploadData(data []byte, charsetID, charsetForm int) error {
	if len(data) == 0 {
		return nil
	}
	var err error
	if !lob.IsDataUploaded() && lob.stream != nil {
		if lob.UploadCtx == nil {
			lob.UploadCtx = context.Background()
		}
		done := lob.stream.StartContext(lob.UploadCtx)
		defer lob.stream.EndContext(done)

		_, err = lob.stream.CreateTemporaryLocator(charsetID, charsetForm)
		if err != nil {
			return err
		}
		err = lob.stream.Write(data)
	}
	return err
}

//func (l *lobBase) IsTemporary() bool {
//	if !l.IsNil() {
//		locator := l.stream.GetLocator()
//		return len(locator) > 7 && locator[7]&1 == 1 || locator[4]&0x40 == 0x40 || l.IsValueBased()
//	}
//	return false
//}

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
	return nil, nil
}

//func (l *lobBase) copyFrom(input *lobBase) error {
//	if input.stream != nil && input.stream.GetLocator() != nil {
//		if l.IsQuasi() {
//			l.stream = input.stream
//			return nil
//		}
//		if bytes.Compare(l.GetLocator(), input.GetLocator()) == 0 {
//			return nil
//		}
//		return fmt.Errorf("the source lob locator is not empty so copy is not permitted, please free the current lob first")
//	}
//	return nil
//}
