package network

type TerminalReader interface {
	read(num int) ([]byte, error)
}
type SessionReader interface {
	GetByte() (uint8, error)
	GetBytes(length int) ([]byte, error)
	GetInt64(size int, compress, bigEndian bool) (int64, error)
	GetInt(size int, compress, bigEndian bool) (int, error)
	GetClr() ([]byte, error)
	GetDlc() ([]byte, error)
	GetFixedClr() ([]byte, error)
	GetString(length int) (string, error)
	GetNullTermString() (string, error)
	GetKeyVal() (key, val []byte, num int, err error)
}
type SessionWriter interface {
	PutInt(number interface{}, size uint8, bigEndian, compress bool)
	PutUint(number interface{}, size uint8, bigEndian, compress bool)
	PutBytes(data ...byte)
	PutClr(data []byte)
	//PutFixedClr(data []byte)
	PutString(data string)
	PutKeyVal(key, val []byte, num uint8)
	PutKeyValString(key, val string, num uint8)
}
type SessionReadWriter interface {
	ResetBuffer()
	SessionReader
	SessionWriter
}

type ValueStreamReader interface {
	Read(session SessionReader) (interface{}, error)
}
type ValueStreamWriter interface {
	Write(session SessionWriter) error
}
type ValueStreamer interface {
	// Read value from stream
	ValueStreamReader
	// Write value to stream
	ValueStreamWriter
}

type Lob struct{}

func (lob *Lob) Read(reader SessionReader) ([]byte, error) {
	// extract []byte data from network by reading flags
	var value []byte
	maxSize, err := reader.GetInt(4, true, true)
	if err != nil {
		return nil, err
	}
	if maxSize > 0 {

		value, err = reader.GetClr()
		if err != nil {
			return nil, err
		}
		/*locator*/ _, err = reader.GetClr()
		if err != nil {
			return nil, err
		}

	}

	// at this point we have []byte data as value we should convert it to original value

	//	v := Vector{}
	//	err = v.decode(par.BValue)
	//	if err != nil {
	//		return err
	//	}
	//	par.oPrimValue = v.Data
	//	_ /*locator*/, err = session.GetClr()
	//	if err != nil {
	//		return err
	//	}
	//}
	return value, nil
}

func (lob *Lob) Write(writer SessionWriter, data []byte) error {
	return nil
}
