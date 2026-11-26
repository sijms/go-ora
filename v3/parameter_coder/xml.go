package parameter_coder

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

// XmlParameter reads an XMLTYPE as its text, as v2 did. The server sends the
// document inline, or a locator to a CLOB holding it when it is large; the
// CLOB is returned unread, and read with the row as any CLOB is.
type XmlParameter struct {
	BasicParameter
	streamer types.LobStreamer
}

// The image's flags: its document inline as text, or a CLOB's locator to it,
// with four bytes before either where xmlSkipNext4 is set. Other bits vary by
// release (18c sets 0x20000 where 21c does not), so each is tested alone.
const (
	xmlImage     = 0x85
	xmlString    = 0x04
	xmlLob       = 0x01
	xmlSkipNext4 = 0x100000
)

func (param *XmlParameter) Copy() OracleParameterCoder {
	ret := new(XmlParameter)
	*ret = *param
	return ret
}

func (param *XmlParameter) Init() {
	param.SetDefault()
	param.DataType = types.XMLType
}

func (param *XmlParameter) SetLobStreamer(streamer types.LobStreamer) {
	param.streamer = streamer
}

func (param *XmlParameter) Encode(_ interface{}, _ IConnection) error {
	return errors.New("an XMLTYPE parameter is not supported: bind the document as a string and call XMLTYPE() on it")
}

func (param *XmlParameter) Read(session network.SessionReader) error {
	var err error
	if param.IsUDTPar || param.IsArrayPar {
		param.BValue, err = session.GetFixedClr()
		return err
	}
	// the type's TOID, the object's OID, and its snapshot
	for i := 0; i < 3; i++ {
		_, err = session.GetDlc()
		if err != nil {
			return err
		}
	}
	_, err = session.GetInt(2, true, true) // version
	if err != nil {
		return err
	}
	var size int
	size, err = session.GetInt(4, true, true)
	if err != nil {
		return err
	}
	_, err = session.GetInt(2, true, true) // flags
	if err != nil {
		return err
	}
	param.BValue = nil
	if size > 0 {
		param.BValue, err = param.BasicRead(session)
	}
	return err
}

// Decode reads the image: 0x85, its length, a flag byte, 4 bytes of flags, then
// the document's text in the database character set or a CLOB's locator.
func (param *XmlParameter) Decode(conn IConnection) (interface{}, error) {
	image := param.BValue
	if len(image) == 0 {
		return nil, nil
	}
	if image[0] != xmlImage {
		return nil, fmt.Errorf("unexpected XMLTYPE image type: %#x", image[0])
	}
	pos := 1
	if pos >= len(image) {
		return nil, errors.New("XMLTYPE image is too short")
	}
	lengthSize := int(image[pos])
	pos++
	if pos+lengthSize > len(image) {
		return nil, errors.New("XMLTYPE image is too short")
	}
	var length uint64
	for _, b := range image[pos : pos+lengthSize] {
		length = length<<8 | uint64(b)
	}
	pos += lengthSize
	if length == 0xFE {
		pos += 4
	}
	pos++ // flag
	if pos+4 > len(image) {
		return nil, errors.New("XMLTYPE image is too short")
	}
	flags := binary.BigEndian.Uint32(image[pos:])
	pos += 4
	if flags&xmlSkipNext4 != 0 {
		pos += 4
		if pos > len(image) {
			return nil, errors.New("XMLTYPE image is too short")
		}
	}
	value := image[pos:]
	switch {
	case flags&xmlString != 0:
		conv, err := conn.GetDefaultStringCoder()
		if err != nil {
			return nil, err
		}
		return conv.Decode(value), nil
	case flags&xmlLob != 0:
		if param.streamer == nil {
			param.streamer = conn.NewLobStreamer()
		}
		param.streamer.SetLocator(value)
		clob := &types.Clob{}
		clob.SetStreamer(param.streamer)
		var err error
		locator := param.streamer.GetLocator()
		if locator.IsVarWidthChar() {
			if param.streamer.DatabaseVersionNumber() < 10200 && locator.IsLittleEndian() {
				clob.Conv, err = conn.GetStringCoder(2002, 0)
			} else {
				clob.Conv, err = conn.GetStringCoder(2000, 0)
			}
		} else {
			clob.Conv, err = conn.GetDefaultStringCoder()
		}
		return clob, err
	default:
		return nil, fmt.Errorf("unsupported XMLTYPE storage: %#x", flags)
	}
}
