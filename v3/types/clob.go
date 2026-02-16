package types

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sijms/go-ora/v3/converters"
)

type Clob interface {
	Data() sql.NullString
	Charset() (charsetID int, charsetForm int)
	ReadFromPos(ctx context.Context, pos int64) (sql.NullString, error)
	ReadBytesFromPos(ctx context.Context, pos, count int64) (sql.NullString, error)
	Lob
}
type ClobDecoder interface {
	DecodeClob(data []byte) (Clob, error)
}
type clob struct {
	Valid            bool
	String           string
	decoder          ClobDecoder
	charsetID        int
	charsetForm      int
	charsetConverter converters.IStringConverter
	lobBase
}

func (l *clob) createLocatorAndLoadData(ctx context.Context, useNClob bool) error {
	var err error
	if l.stream == nil {
		return errNilStreamer
	}
	if l.stream.GetLocator() == nil {
		if l.charsetForm == 0 {
			l.charsetForm = 1
			if useNClob {
				l.charsetForm = 2
			}
		}
		l.charsetConverter, err = l.stream.GetStringCoder().GetStringCoder(l.charsetID, l.charsetForm)
		if err != nil {
			return err
		}
		if l.charsetID == 0 {
			l.charsetID = l.charsetConverter.GetLangID()
		}
		if !l.Valid || len(l.String) == 0 {
			return nil
		}
		_, err = l.stream.CreateTemporaryLocator(l.charsetID, l.charsetForm)
		if err != nil {
			return err
		}
		if l.IsVarWidthChar() {
			if l.IsLittleEndian() {
				l.charsetConverter, err = l.stream.GetStringCoder().GetStringCoder(2002, 0)
			} else {
				l.charsetConverter, err = l.stream.GetStringCoder().GetStringCoder(2000, 0)
			}
		}
		bytes := l.charsetConverter.Encode(l.String)
		done := l.stream.StartContext(ctx)
		defer l.stream.EndContext(done)
		err = l.stream.Write(bytes)
	}
	return err
}
func CreateClob(db *sql.DB, ctx context.Context, data sql.NullString, useNClob bool) (Clob, error) {
	var err error
	ret := &clob{Valid: data.Valid, String: data.String}
	_, err = db.ExecContext(context.Background(), "--CREATE-LOB-STREAM--", &ret.stream)
	if err != nil {
		return nil, err
	}
	err = ret.createLocatorAndLoadData(ctx, useNClob)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
func NewClob(streamer LobStreamer, decoder ClobDecoder, data sql.NullString, useNClob bool) (Clob, error) {
	ret := &clob{
		Valid:   data.Valid,
		String:  data.String,
		decoder: decoder,
		lobBase: lobBase{stream: streamer},
	}
	err := ret.createLocatorAndLoadData(context.Background(), useNClob)
	return ret, err
}

//	func NewClob(streamer LobStreamer, charsetID, charsetForm int, data []byte) (Clob, error) {
//		ret := &clob{}
//		var err error
//		ret.stream = streamer
//		if ret.IsVarWidthChar() {
//			if ret.stream.DatabaseVersionNumber() < 10200 && ret.IsLittleEndian() {
//				ret.charsetConverter, err = streamer.GetStringCoder().GetStringCoder(2002, 0)
//			} else {
//				ret.charsetConverter, err = streamer.GetStringCoder().GetStringCoder(2000, 0)
//			}
//		} else {
//			ret.charsetConverter, err = streamer.GetStringCoder().GetStringCoder(charsetID, charsetForm)
//		}
//		if err != nil {
//			return nil, err
//		}
//		if len(data) == 0 {
//			ret.Valid = false
//		} else {
//			ret.Valid = true
//			ret.String = ret.charsetConverter.Decode(data)
//		}
//		return ret, nil
//	}
func (l *clob) Charset() (charsetID int, charsetForm int) {
	return l.charsetID, l.charsetForm
}
func (l *clob) Scan(value interface{}) error {
	var err error
	if value == nil {
		if l.stream != nil {
			l.stream.SetLocator(nil)
		}
		l.String = ""
		l.Valid = false
		return nil
	}
	switch v := value.(type) {
	case *clob:
		err = l.lobBase.copyFrom(&v.lobBase)
		if err != nil {
			return err
		}
		l.Valid = v.Valid
		l.String = v.String
		l.charsetForm = v.charsetForm
		l.charsetID = v.charsetID
	case string:
		l.String = v
		l.Valid = true
	case sql.NullString:
		l.String = v.String
		l.Valid = v.Valid
	case *sql.NullString:
		l.String = v.String
		l.Valid = v.Valid
	case []byte:
		if l.decoder != nil {
			var temp Clob
			temp, err = l.decoder.DecodeClob(v)
			if err != nil {
				return err
			}
			l.String = temp.Data().String
			l.Valid = temp.Data().Valid
		} else {
			l.String = string(v)
			l.Valid = v != nil
		}
	default:
		return fmt.Errorf("value of type %T cannot scanned into Clob", value)
	}
	return nil
}

func (l *clob) Data() sql.NullString {
	//if l.IsNil() || !l.Valid {
	//	return sql.NullString{String: l.String, Valid: l.Valid}, nil
	//}
	//err := l.Read(ctx)
	//if err != nil {
	//	return sql.NullString{}, err
	//}
	//if free {
	//	err = l.Free()
	//	if err != nil {
	//		return sql.NullString{}, err
	//	}
	//}
	return sql.NullString{String: l.String, Valid: l.Valid}
}

func (l *clob) ReadBytesFromPos(ctx context.Context, pos, count int64) (sql.NullString, error) {
	if l.decoder == nil {
		return sql.NullString{}, fmt.Errorf("no decoder defined for clob type")
	}
	done := l.stream.StartContext(ctx)
	defer l.stream.EndContext(done)
	data, err := l.lobBase.ReadBytesFromPos(ctx, pos, count)
	if err != nil {
		return sql.NullString{}, err
	}
	temp, err := l.decoder.DecodeClob(data)
	if err != nil {
		return sql.NullString{}, err
	}
	return temp.Data(), nil
}

func (l *clob) ReadFromPos(ctx context.Context, pos int64) (sql.NullString, error) {
	return l.ReadBytesFromPos(ctx, pos, 0)
}
func (l *clob) Read(ctx context.Context) error {
	// if there is data don't read to avoid multiple read
	if l.Valid || len(l.String) > 0 {
		return nil
	}
	temp, err := l.ReadFromPos(ctx, 0)
	if err != nil {
		return err
	}
	l.Valid = temp.Valid
	l.String = temp.String
	return nil
}

//
//func NewClob(data sql.NullString, useNCharset bool) *Clob {
//	charsetForm := 1
//	if useNCharset {
//		charsetForm = 2
//	}
//	return &Clob{
//		String: String{
//			//TypeInfo: type_coder.TypeInfo{
//			//	CharsetForm: charsetForm,
//			//},
//			//Data: data,
//		},
//		//TypeInfo: type_coder.TypeInfo{
//		//	DataType: OCIClobLocator,
//		//},
//	}
//}
//
//func (clob *Clob) Read(session network.SessionReader, tnsType uint16, isUDTPar bool) error {
//	var err error
//	var bValue []byte
//	if clob.isInline {
//		maxSize, err := session.GetInt(4, true, true)
//		if err != nil {
//			return err
//		}
//		if maxSize > 0 {
//			/*size*/ _, err = session.GetInt64(8, true, true)
//			if err != nil {
//				return err
//			}
//			/*chunkSize*/ _, err = session.GetInt(4, true, true)
//			if err != nil {
//				return err
//			}
//			var flag uint8
//			flag, err = session.GetByte()
//			if err != nil {
//				return err
//			}
//			clob.TypeInfo.CharsetID = 0
//			if flag == 1 {
//				clob.TypeInfo.CharsetID, err = session.GetInt(2, true, true)
//				if err != nil {
//					return err
//				}
//			}
//			var temp uint8
//			temp, err = session.GetByte()
//			if err != nil {
//				return err
//			}
//			clob.TypeInfo.CharsetForm = int(temp)
//			bValue, err = session.GetClr()
//			if err != nil {
//				return err
//			}
//			clob.locator, err = session.GetClr()
//			if err != nil {
//				return err
//			}
//
//		} else {
//			clob.locator = nil
//		}
//		_, err = clob.Decode(bValue, tnsType)
//		return err
//	}
//
//	if isUDTPar {
//		clob.locator, err = session.GetFixedClr()
//	} else {
//		clob.locator, err = session.GetClr()
//	}
//	return err
//}
//
//func (clob *Clob) Write(session network.SessionWriter) error {
//	var err error
//	var bValue []byte
//	bValue, err = clob.Encode()
//	if err != nil {
//		return err
//	}
//	return clob.write(session, bValue, false)
//}
//
