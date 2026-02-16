package go_ora

import (
	"bytes"
	"context"
	"errors"
	"go/types"

	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/trace"
)

type Clob struct {
	locator []byte
	String  string
	Valid   bool
}

type NClob Clob

//type lobInterface interface {
//	getLocator() []byte
//}

type Blob struct {
	locator []byte
	Data    []byte
	// deprecated
	Valid bool
}

var errEmptyLocator = errors.New("call lob operation on an empty locator")

type LobStream struct {
	conn          *Connection
	sourceLocator []byte
	destLocator   []byte
	scn           []byte
	sourceOffset  int64
	destOffset    int64
	destLen       int
	charsetID     int
	size          int64
	data          bytes.Buffer
	bNullO2U      bool
	isNull        bool
	sendSize      bool
}

//func newLob(connection *Connection) *LobStream {
//	return &LobStream{
//		conn: connection,
//	}
//}

func (lob *LobStream) initialize() {
	lob.bNullO2U = false
	lob.sendSize = false
	lob.size = 0
	lob.sourceOffset = 0
	lob.destOffset = 0
}

// variableWidthChar if lob has variable width char or not
func (lob *LobStream) IsVarWidthChar() bool {
	return len(lob.sourceLocator) > 6 && lob.sourceLocator[6]&0x80 == 0x80
}

// IsLittleEndian if Lob is littleEndian or not
func (lob *LobStream) IsLittleEndian() bool {
	return len(lob.sourceLocator) > 7 && lob.sourceLocator[7]&0x40 == 0x40
}

func (lob *LobStream) GetLocator() []byte {
	return lob.sourceLocator
}
func (lob *LobStream) SetLocator(locator []byte) {
	lob.sourceLocator = locator
}
func (lob *LobStream) GetTracer() trace.Tracer {
	return lob.conn.tracer
}

func (lob *LobStream) DatabaseVersionNumber() int {
	return int(lob.conn.dBVersion.Number)
}

func (lob *LobStream) GetStringCoder() converters.StringCoder {
	return lob.conn
}

// getSize return lob size
func (lob *LobStream) GetSize() (size int64, err error) {
	if lob.sourceLocator == nil {
		return 0, errEmptyLocator
	}
	lob.initialize()
	lob.sendSize = true
	session := lob.conn.session
	lob.conn.tracer.Print("Read Lob Size")
	session.ResetBuffer()
	lob.writeOp(1)
	err = session.Write()
	if err != nil {
		return
	}
	err = lob.read()
	if err != nil {
		return
	}
	size = lob.size
	lob.conn.tracer.Print("Lob Size: ", size)
	return
}

func (lob *LobStream) Read(offset, count int64) (data []byte, err error) {
	if lob.sourceLocator == nil {
		return nil, errEmptyLocator
	}
	if offset == 0 && count == 0 {
		lob.conn.tracer.Print("Read Lob Data:")
	} else {
		lob.conn.tracer.Printf("Read Lob Data Position: %d, Count: %d\n", offset, count)
	}
	lob.initialize()
	lob.size = count
	lob.sourceOffset = offset + 1
	lob.sendSize = true
	lob.data.Reset()
	session := lob.conn.session
	session.ResetBuffer()
	//done := session.StartContext(ctx)
	//defer session.EndContext(done)
	lob.writeOp(2)
	err = session.Write()
	if err != nil {
		return
	}
	err = lob.read()
	if err != nil {
		return
	}
	data = lob.data.Bytes()
	return
}
func (lob *LobStream) Write(data []byte) error {
	if lob.sourceLocator == nil {
		return errEmptyLocator
	}
	lob.conn.tracer.Printf("Write Lob Data: %d bytes", len(data))
	lob.initialize()
	//lob.size = int64(len(data))
	//lob.sendSize = true
	lob.sourceOffset = 1
	lob.conn.session.ResetBuffer()
	lob.writeOp(0x40)
	lob.conn.session.PutBytes(0xE)
	lob.conn.session.PutClr(data)
	err := lob.conn.session.Write()
	if err != nil {
		return err
	}
	return lob.read()
}

// getData return lob data
func (lob *LobStream) getData() (data []byte, err error) {
	return lob.Read(0, 0)
}

func (lob *LobStream) putData(data []byte) error {
	return lob.Write(data)
}

func (lob *LobStream) putString(data string) error {
	if lob.sourceLocator == nil {
		return errEmptyLocator
	}
	conn := lob.conn
	conn.tracer.Printf("Put Lob String: %d character", int64(len([]rune(data))))
	lob.initialize()
	var strConv converters.IStringConverter
	if lob.IsVarWidthChar() {
		if conn.dBVersion.Number < 10200 && lob.IsLittleEndian() {
			strConv, _ = conn.GetStringCoder(2002, 0)
		} else {
			strConv, _ = conn.GetStringCoder(2000, 0)
		}
	} else {
		var err error
		strConv, err = conn.GetStringCoder(lob.charsetID, 0)
		if err != nil {
			return err
		}
	}
	lobData := strConv.Encode(data)
	// lob.size = int64(len([]rune(data)))
	// lob.sendSize = true
	lob.sourceOffset = 1
	lob.conn.session.ResetBuffer()
	lob.writeOp(0x40)
	lob.conn.session.PutBytes(0xE)
	lob.conn.session.PutClr(lobData)
	err := lob.conn.session.Write()
	if err != nil {
		return err
	}
	return lob.read()
}

func (lob *LobStream) Exists() (bool, error) {
	if lob.sourceLocator == nil {
		return false, errEmptyLocator
	}
	lob.initialize()
	lob.bNullO2U = true
	session := lob.conn.session
	session.ResetBuffer()
	lob.writeOp(0x800)
	err := session.Write()
	if err != nil {
		return false, err
	}
	err = lob.read()
	if err != nil {
		return false, err
	}
	return lob.isNull, nil
}

func (lob *LobStream) GetLobStreamMode() configurations.LobFetch {
	return lob.conn.connOption.Lob
}
func (lob *LobStream) GetLobReadMode() configurations.LobReadMode {
	return lob.conn.connOption.LobReadMode
}

// isTemporary: return true if the lob is temporary
func (lob *LobStream) isTemporary() bool {
	if len(lob.sourceLocator) > 7 {
		if lob.sourceLocator[7]&1 == 1 || lob.sourceLocator[4]&0x40 == 0x40 || lob.isValueBasedLocator() {
			return true
		}
	}
	return false
}

func (lob *LobStream) isQuasiLocator() bool {
	return lob.sourceLocator[3] == 4
}

func (lob *LobStream) isValueBasedLocator() bool {
	return lob.sourceLocator[4]&0x20 > 0
}
func (lob *LobStream) StartContext(ctx context.Context) chan struct{} {
	return lob.conn.session.StartContext(ctx)
}
func (lob *LobStream) EndContext(done chan struct{}) {
	lob.conn.session.EndContext(done)
}
func (lob *LobStream) FreeTemporaryLocator() error {
	if lob.sourceLocator == nil {
		return nil
	}
	lob.initialize()
	lob.conn.session.ResetBuffer()
	lob.writeOp(0x111)
	err := lob.conn.session.Write()
	if err != nil {
		return err
	}
	err = lob.read()
	if err != nil {
		return err
	}
	lob.sourceLocator = nil
	return nil
}
func (lob *LobStream) CreateTemporaryLocator(charsetID, charsetForm int) ([]byte, error) {
	lob.initialize()
	lob.conn.tracer.Print("Create Temporary Lob:")
	lob.sourceLocator = make([]byte, 0x28)
	lob.sourceLocator[1] = 0x54
	//lob.sourceLen = len(lob.sourceLocator)
	lob.bNullO2U = true
	lob.scn = make([]byte, 1)
	// lob.scn[0] = 1 if you need to cache otherwise 0
	// 0xA is equal to duration
	lob.destLen = 0xA
	lob.size = 0xA
	lob.sendSize = true
	if charsetForm == 0 {
		lob.destOffset = 0x71
		lob.charsetID = 1
	} else {
		lob.destOffset = 0x70
		lob.charsetID = charsetID
		lob.sourceOffset = int64(charsetForm)
	}
	session := lob.conn.session
	session.ResetBuffer()
	lob.writeOp(0x110)
	err := session.Write()
	if err != nil {
		return nil, err
	}
	err = lob.read()
	if err != nil {
		return nil, err
	}
	return lob.sourceLocator, nil
}

//func (lob *Lob) createTemporaryClob(charset, charsetForm int) error {
//	//lob.initialize()
//	//lob.conn.tracer.Print("Create Temporary CLob")
//	//lob.sourceLocator = make([]byte, 0x28)
//	//lob.sourceLocator[1] = 0x54
//	//lob.bNullO2U = true
//	//lob.scn = make([]byte, 1)
//	//lob.destLen = 0xA
//	//lob.size = 0xA
//	//lob.sendSize = true
//	//lob.charsetID = charset
//	//lob.sourceOffset = int64(charsetForm)
//	session := lob.conn.session
//	session.ResetBuffer()
//	lob.writeOp(0x110)
//	err := session.Write()
//	if err != nil {
//		return err
//	}
//	return lob.read()
//}

func (lob *LobStream) Open(mode, opID int) error {
	if lob.sourceLocator == nil {
		return errEmptyLocator
	}
	lob.conn.tracer.Printf("Open Lob: Mode= %d   Operation ID= %d", mode, opID)
	if lob.isTemporary() {
		if lob.sourceLocator[7]&8 == 8 {
			return errors.New("TTC Error")
		}
		if mode == 2 {
			lob.sourceLocator[7] |= 0x10
		}
		return nil
	}

	lob.initialize()
	lob.size = int64(mode)
	lob.sendSize = true
	//done := lob.conn.session.StartContext(context)
	//defer lob.conn.session.EndContext(done)
	lob.conn.session.ResetBuffer()
	lob.writeOp(opID)
	err := lob.conn.session.Write()
	if err != nil {
		return err
	}
	err = lob.read()
	return processReset(err, lob.conn)
}

func (lob *LobStream) Close(opID int) error {
	if lob.sourceLocator == nil {
		return errEmptyLocator
	}
	lob.conn.tracer.Print("Close Lob: ")
	lob.initialize()
	lob.conn.session.ResetBuffer()
	lob.writeOp(opID)
	err := lob.conn.session.Write()
	if err != nil {
		return err
	}
	err = lob.read()
	return processReset(err, lob.conn)
	//}
}

func (lob *LobStream) writeOp(operationID int) {
	session := lob.conn.session
	session.PutTTCFunc(0x60)
	if len(lob.sourceLocator) == 0 {
		session.PutBytes(0)
	} else {
		session.PutBytes(1)
	}
	session.PutUint(len(lob.sourceLocator), 4, true, true)

	if len(lob.destLocator) == 0 {
		session.PutBytes(0)
	} else {
		lob.destLen = len(lob.destLocator)
		session.PutBytes(1)
	}
	session.PutUint(lob.destLen, 4, true, true)

	// put offsets
	if session.TTCVersion < 3 {
		session.PutUint(lob.sourceOffset, 4, true, true)
		session.PutUint(lob.destOffset, 4, true, true)
	} else {
		session.PutBytes(0, 0)
	}

	if lob.charsetID != 0 {
		session.PutBytes(1)
	} else {
		session.PutBytes(0)
	}

	if lob.sendSize && session.TTCVersion < 3 {
		session.PutBytes(1)
	} else {
		session.PutBytes(0)
	}

	if lob.bNullO2U {
		session.PutBytes(1)
	} else {
		session.PutBytes(0)
	}

	session.PutInt(operationID, 4, true, true)
	if len(lob.scn) == 0 {
		session.PutBytes(0)
	} else {
		session.PutBytes(1)
	}
	session.PutUint(len(lob.scn), 4, true, true)

	if session.TTCVersion >= 3 {
		session.PutUint(lob.sourceOffset, 8, true, true)
		session.PutInt(lob.destOffset, 8, true, true)
		if lob.sendSize {
			session.PutBytes(1)
		} else {
			session.PutBytes(0)
		}
	}
	if session.TTCVersion >= 4 {
		session.PutBytes(0, 0, 0, 0, 0, 0)
	}

	if len(lob.sourceLocator) > 0 {
		session.PutBytes(lob.sourceLocator...)
	}

	if len(lob.destLocator) > 0 {
		session.PutBytes(lob.destLocator...)
	}

	if lob.charsetID != 0 {
		session.PutUint(lob.charsetID, 2, true, true)
	}
	if session.TTCVersion < 3 && lob.sendSize {
		session.PutUint(lob.size, 4, true, true)
	}
	for x := 0; x < len(lob.scn); x++ {
		session.PutUint(lob.scn[x], 4, true, true)
	}
	if session.TTCVersion >= 3 && lob.sendSize {
		session.PutUint(lob.size, 8, true, true)
	}
}

// read lob response from network session
func (lob *LobStream) read() error {
	loop := true
	session := lob.conn.session
	for loop {
		msg, err := session.GetByte()
		if err != nil {
			return err
		}
		switch msg {
		case 8:
			// read rpa message
			if len(lob.sourceLocator) > 0 {
				lob.sourceLocator, err = session.GetBytes(len(lob.sourceLocator))
				if err != nil {
					return err
				}
			}
			if len(lob.destLocator) != 0 {
				lob.destLocator, err = session.GetBytes(len(lob.destLocator))
				if err != nil {
					return err
				}
				lob.destLen = len(lob.destLocator)
			} else {
				lob.destLen = 0
			}
			if lob.charsetID != 0 {
				lob.charsetID, err = session.GetInt(2, true, true)
				if err != nil {
					return err
				}
			}
			if lob.sendSize {
				// get data size
				if session.TTCVersion < 3 {
					lob.size, err = session.GetInt64(4, true, true)
					if err != nil {
						return err
					}
				} else {
					lob.size, err = session.GetInt64(8, true, true)
					if err != nil {
						return err
					}
				}
			}
			if lob.bNullO2U {
				temp, err := session.GetInt(2, true, true)
				if err != nil {
					return err
				}
				if temp != 0 {
					lob.isNull = true
				}
			}
		case 14:
			// get the data
			err = lob.readData()
			if err != nil {
				return err
			}
		default:
			err = lob.conn.processTCCResponse(msg)
			if err != nil {
				return err
			}
			if msg == 4 || msg == 9 {
				loop = false
			}
		}
	}
	return nil
}

// read lob data chunks from network session
func (lob *LobStream) readData() error {
	session := lob.conn.session
	tempBytes, err := session.GetClr()
	if err != nil {
		return err
	}
	lob.data.Write(tempBytes)
	return nil
}

func (lob *LobStream) GetLobId() []byte {
	// BitConverter.ToString(lobLocator, 10, 10);
	if lob.sourceLocator == nil {
		return nil
	}
	return lob.sourceLocator[10 : 10+10]
}

func (lob *LobStream) append(dest []byte) error {
	if lob.sourceLocator == nil {
		return errEmptyLocator
	}
	lob.initialize()
	lob.destLocator = dest
	lob.destLen = len(dest)
	lob.conn.session.ResetBuffer()
	lob.writeOp(0x80)
	err := lob.conn.session.Write()
	if err != nil {
		return err
	}
	err = lob.read()
	return processReset(err, lob.conn)
}

func (lob *LobStream) copy(srcLocator, dstLocator []byte, srcOffset, dstOffset, length int64) error {
	lob.initialize()
	lob.sourceLocator = srcLocator
	lob.destLocator = dstLocator
	lob.destLen = len(dstLocator)
	lob.sourceOffset = srcOffset
	lob.destOffset = dstOffset
	lob.size = length
	lob.sendSize = true
	lob.conn.session.ResetBuffer()
	lob.writeOp(4)
	err := lob.conn.session.Write()
	if err != nil {
		return err
	}
	err = lob.read()
	return processReset(err, lob.conn)
}

func (val *Clob) Scan(value interface{}) error {
	val.Valid = true
	if value == nil {
		val.Valid = false
		val.String = ""
		return nil
	}
	switch temp := value.(type) {
	case Clob:
		*val = temp
	case *Clob:
		*val = *temp
	case NClob:
		*val = Clob(temp)
	case *NClob:
		*val = Clob(*temp)
	case string:
		val.String = temp
	case types.Nil:
		val.String = ""
		val.Valid = false
	default:
		return errors.New("go-ora: Clob column type require Clob or string values")
	}
	return nil
}

func (val *Blob) Scan(value interface{}) error {
	// val.Valid = true
	if value == nil {
		// val.Valid = false
		val.Data = nil
		return nil
	}
	switch temp := value.(type) {
	case Blob:
		*val = temp
	case *Blob:
		*val = *temp
	case []byte:
		val.Data = temp
	case types.Nil:
		val.Data = nil
	default:
		return errors.New("go-ora: Blob column type require Blob or []byte values")
	}
	return nil
}

func (val *NClob) Scan(value interface{}) error {
	val.Valid = true
	if value == nil {
		val.Valid = false
		val.String = ""
		return nil
	}
	switch temp := value.(type) {
	case Clob:
		*val = NClob(temp)
	case *Clob:
		*val = NClob(*temp)
	case NClob:
		*val = temp
	case *NClob:
		*val = *temp
	case string:
		val.String = temp
	case types.Nil:
		val.String = ""
		val.Valid = false
	default:
		return errors.New("go-ora: Clob column type require Clob or string values")
	}
	return nil
}

func (val Blob) getLocator() []byte {
	return val.locator
}

func (val Clob) getLocator() []byte {
	return val.locator
}

func (val NClob) getLocator() []byte {
	return val.locator
}

//func (val Clob) Value() (driver.Value, error) {
//	if val.Valid {
//		return val.String, nil
//	} else {
//		return nil, nil
//	}
//}

//
//func (val *NClob) Value() (driver.Value, error) {
//	if val.Valid {
//		return val.String, nil
//	} else {
//		return nil, nil
//	}
//}
//
//func (val *Blob) Value() (driver.Value, error) {
//	if val.Valid {
//		return val.Data, nil
//	} else {
//		return nil, nil
//	}
//}
