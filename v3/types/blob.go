package types

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
)

type Blob struct {
	Raw
	lobBase
	UploadCtx context.Context
}

func (blob *Blob) Upload() error {
	return blob.uploadData(blob.bValue, 0, 0)
}

func (blob *Blob) SetValue(input interface{}) error {
	var err error
	switch input := input.(type) {
	case Blob:
		*blob = input
		return nil
	case *Blob:
		*blob = *input
		return nil
	default:
		err = blob.Raw.SetValue(input)
		if err != nil {
			return err
		}
	}
	return blob.Upload()
}

//	type Blob interface {
//		Data() []byte
//		//CopyFrom(blob Blob) error
//		ReadFromPos(ctx context.Context, pos int64) ([]byte, error)
//		ReadBytesFromPos(ctx context.Context, pos, count int64) ([]byte, error)
//		Lob
//	}
//type blob struct {
//	data []byte
//	lobBase
//	UploadCtx context.Context
//}

func CreateBlob(db *sql.DB, uploadCtx context.Context, input interface{}) (*Blob, error) {
	var err error
	ret := &Blob{}
	ret.UploadCtx = uploadCtx
	err = ret.createStreamer(db)
	if err != nil {
		return nil, err
	}
	return ret, ret.SetValue(input)
}

func NewBlob(stream LobStreamer, uploadCtx context.Context, input interface{}) (*Blob, error) {
	ret := &Blob{}
	ret.UploadCtx = uploadCtx
	ret.stream = stream
	return ret, ret.SetValue(input)
}

//	func (blob *Blob) createLocatorAndUploadData(ctx context.Context) error {
//		if l.stream == nil {
//			return errNilStreamer
//		}
//		var err error
//		if len(blob.bValue) == 0 {
//			return nil
//		}
//		done := blob.stream.StartContext(ctx)
//		defer blob.stream.EndContext(done)
//		_, err = blob.stream.CreateTemporaryLocator(0, 0)
//		if err != nil {
//			return err
//		}
//		err = blob.stream.Write(blob.bValue)
//		return err
//	}
func (blob *Blob) Read(ctx context.Context) error {
	var err error
	blob.bValue, err = blob.ReadFromPos(ctx, 0)
	return err
}
func (blob *Blob) CopyTo(dest driver.Value) error {
	switch dst := dest.(type) {
	case *[]byte:
		*dst = blob.bValue
	case *string:
		*dst = string(blob.bValue)
	case *sql.NullString:
		if blob.bValue == nil {
			*dst = sql.NullString{Valid: false}
		} else {
			*dst = sql.NullString{String: string(blob.bValue), Valid: true}
		}
	default:
		return fmt.Errorf("cannot copy blob to variable of type %T ", dest)
	}
	return nil
}
func (blob *Blob) Scan(src interface{}) error {
	return blob.SetValue(src)
}

//func (l *blob) Data() []byte {
//	//if l.IsNil() || l.data != nil {
//	//	return l.data, nil
//	//}
//	//var err error
//	//err = l.Read(ctx)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//if free {
//	//	err = l.Free()
//	//	if err != nil {
//	//		return nil, err
//	//	}
//	//}
//	return l.data
//}
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
