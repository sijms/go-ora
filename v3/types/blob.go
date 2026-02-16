package types

import (
	"context"
	"database/sql"
	"fmt"
)

type Blob interface {
	Data() []byte
	//CopyFrom(blob Blob) error
	ReadFromPos(ctx context.Context, pos int64) ([]byte, error)
	ReadBytesFromPos(ctx context.Context, pos, count int64) ([]byte, error)
	Lob
}
type blob struct {
	data []byte
	lobBase
}

func CreateBlob(db *sql.DB, ctx context.Context, data []byte) (Blob, error) {
	ret := &blob{data: data}
	if data == nil {
		return ret, nil
	}
	_, err := db.ExecContext(context.Background(), "--CREATE-LOB-STREAM--", &ret.stream)
	if err != nil {
		return nil, err
	}
	err = ret.createLocatorAndUploadData(ctx)
	return ret, nil
}

func NewBlob(streamer LobStreamer, data []byte) (Blob, error) {
	ret := &blob{
		data: data,
		lobBase: lobBase{
			stream: streamer,
		},
	}
	err := ret.createLocatorAndUploadData(context.Background())
	return ret, err
}

func (l *blob) createLocatorAndUploadData(ctx context.Context) error {
	if l.stream == nil {
		return errNilStreamer
	}
	var err error
	if l.stream.GetLocator() == nil {
		if len(l.data) == 0 {
			return nil
		}
		done := l.stream.StartContext(ctx)
		defer l.stream.EndContext(done)
		_, err = l.stream.CreateTemporaryLocator(0, 0)
		if err != nil {
			return err
		}
		err = l.stream.Write(l.data)
	}
	return err
}
func (l *blob) Read(ctx context.Context) error {
	if len(l.data) > 0 {
		return nil
	}
	var err error
	l.data, err = l.ReadFromPos(ctx, 0)
	return err
}

func (l *blob) Scan(src interface{}) error {
	var err error
	if src == nil {
		if l.stream != nil {
			l.stream.SetLocator(nil)
		}
		l.data = nil
		return nil
	}
	switch v := src.(type) {
	case *blob:
		err = l.lobBase.copyFrom(&v.lobBase)
		if err != nil {
			return err
		}
		l.data = v.data
	case []byte:
		l.data = v
	default:
		return fmt.Errorf("value of type %T cannot scanned into Blob", src)
	}
	return nil
}

func (l *blob) Data() []byte {
	//if l.IsNil() || l.data != nil {
	//	return l.data, nil
	//}
	//var err error
	//err = l.Read(ctx)
	//if err != nil {
	//	return nil, err
	//}
	//if free {
	//	err = l.Free()
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	return l.data
}

//func (lob *blob) Value() (driver.Value, error) {
//	return lob.Data(context.Background(), false)
//}

//func (blob *Blob) Encode() ([]byte, error) {
//	return blob.Data, nil
//}
//func (blob *Blob) Decode(data []byte, tnsType uint16) (interface{}, error) {
//	blob.Data = data
//	return blob.Data, nil
//}
//func (blob *Blob) Read(session network.SessionReader, tnsType uint16, isUDTPar bool) error {
//	var err error
//	blob.Data, err = blob.read(session, isUDTPar)
//	if err != nil {
//		return err
//	}
//	if blob.locator == nil {
//		blob.Data = nil
//		return nil
//	}
//	// non-inline blob data should be loaded
//	return nil
//}
//
//func (blob *Blob) Write(session network.SessionWriter) error {
//	return blob.write(session, blob.Data, false)
//}
//
//func (blob *Blob) Scan(value interface{}) error {
//	switch v := value.(type) {
//	case *Blob:
//		*blob = *v
//	case Blob:
//		*blob = v
//	case []byte:
//		blob.Data = v
//	default:
//		return fmt.Errorf("Blob column type require []byte value")
//	}
//	return nil
//}
