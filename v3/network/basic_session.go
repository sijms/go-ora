package network

import (
	"bytes"
	"encoding/binary"
	"errors"
	"strconv"
)

type SessionProperties struct {
	UseBigClrChunks bool
	UseBigScn       bool
	ClrChunkSize    int
}
type basicSession struct {
	inBuffer  *bytes.Buffer
	outBuffer *bytes.Buffer
	SessionProperties
	terminal TerminalReader
}

// GetByte read one uint8 from input buffer
func (session *basicSession) GetByte() (uint8, error) {
	rb, err := session.terminal.read(1)
	if err != nil {
		return 0, err
	}
	return rb[0], nil
}

// GetBytes read specified number of bytes from input buffer
func (session *basicSession) GetBytes(length int) ([]byte, error) {
	return session.terminal.read(length)
}

// GetInt64 read int64 number from input buffer.
// you should specify the size of the int and either compress or not and stored as big endian or not
func (session *basicSession) GetInt64(size int, compress bool, bigEndian bool) (int64, error) {
	var ret int64
	negFlag := false
	if compress {
		rb, err := session.GetByte()
		if err != nil {
			return 0, err
		}
		size = int(rb)
		if size&0x80 > 0 {
			negFlag = true
			size = size & 0x7F
		}
		bigEndian = true
	}
	if size == 0 {
		return 0, nil
	} else if size > 8 {
		// When compress is true, "size" may be a value greater than 8 in some cases
		return 0, errors.New("invalid size for GetInt64: " + strconv.Itoa(size))
	}
	rb, err := session.GetBytes(size)
	if err != nil {
		return 0, err
	}
	temp := make([]byte, 8)
	if bigEndian {
		copy(temp[8-size:], rb)
		ret = int64(binary.BigEndian.Uint64(temp))
	} else {
		copy(temp[:size], rb)
		// temp = append(pck.buffer[pck.index: pck.index + size], temp...)
		ret = int64(binary.LittleEndian.Uint64(temp))
	}
	if negFlag {
		ret = ret * -1
	}
	return ret, nil
}

// GetInt read int number from input buffer.
// you should specify the size of the int and either compress or not and stored as big endian or not
func (session *basicSession) GetInt(size int, compress bool, bigEndian bool) (int, error) {
	temp, err := session.GetInt64(size, compress, bigEndian)
	if err != nil {
		return 0, err
	}
	return int(temp), nil
}

// GetClr reed variable length bytearray from input buffer
func (session *basicSession) GetClr() (output []byte, err error) {
	var nb byte
	nb, err = session.GetByte()
	if err != nil {
		return
	}
	if nb == 0 || nb == 0xFF || nb == 0xFD {
		output = nil
		err = nil
		return
	}
	chunkSize := int(nb)
	var chunk []byte
	var tempBuffer bytes.Buffer
	if chunkSize == 0xFE {
		for chunkSize > 0 {
			//if session.IsBreak() {
			//	break
			//}
			if session.UseBigClrChunks {
				chunkSize, err = session.GetInt(4, true, true)
			} else {
				nb, err = session.GetByte()
				chunkSize = int(nb)
			}
			if err != nil {
				return
			}
			chunk, err = session.GetBytes(chunkSize)
			if err != nil {
				return
			}
			tempBuffer.Write(chunk)
		}
	} else {
		chunk, err = session.GetBytes(chunkSize)
		if err != nil {
			return
		}
		tempBuffer.Write(chunk)
	}
	output = tempBuffer.Bytes()
	return
	// var size uint8
	// var rb []byte
	// size, err = session.GetByte()
	// if err != nil {
	// 	return
	// }
	// if size == 0 || size == 0xFF {
	// 	output = nil
	// 	err = nil
	// 	return
	// }
	// if size != 0xFE {
	// 	output, err = session.read(int(size))
	// 	return
	// }
	//
	// for {
	// 	var size1 int
	// 	if session.UseBigClrChunks {
	// 		size1, err = session.GetInt(4, true, true)
	// 	} else {
	// 		size, err = session.GetByte()
	// 		size1 = int(size)
	// 	}
	// 	if err != nil || size1 == 0 {
	// 		break
	// 	}
	// 	rb, err = session.read(size1)
	// 	if err != nil {
	// 		return
	// 	}
	// 	tempBuffer.Write(rb)
	// }
	// output = tempBuffer.Bytes()
	// return
}

// GetDlc read variable length bytearray from input buffer
func (session *basicSession) GetDlc() (output []byte, err error) {
	var length int
	length, err = session.GetInt(4, true, true)
	if err != nil {
		return
	}
	if length > 0 {
		output, err = session.GetClr()
		if len(output) > length {
			output = output[:length]
		}
	}
	return
}

func (session *basicSession) GetFixedClr() ([]byte, error) {
	nb, err := session.GetByte()
	if err != nil {
		return nil, err
	}
	var size int
	switch nb {
	case 0, 0xFD, 0xFF:
		return nil, nil
	case 0xFE:
		size, err = session.GetInt(4, false, true)
		if err != nil {
			return nil, err
		}
	default:
		size = int(nb)
	}
	return session.GetBytes(size)
}

// GetString read a string data from input buffer
func (session *basicSession) GetString(length int) (string, error) {
	ret, err := session.GetClr()
	return string(ret[:length]), err
}

// GetNullTermString read a null terminated string from input buffer
func (session *basicSession) GetNullTermString() (result string, err error) {
	return session.inBuffer.ReadString(0)
}

// GetKeyVal read key, value (in form of bytearray), a number flag from input buffer
func (session *basicSession) GetKeyVal() (key []byte, val []byte, num int, err error) {
	key, err = session.GetDlc()
	if err != nil {
		return
	}
	val, err = session.GetDlc()
	if err != nil {
		return
	}
	num, err = session.GetInt(4, true, true)
	return
}

// PutInt write int number with size entered either use bigEndian or not and use compression or not to
func (session *basicSession) PutInt(number interface{}, size uint8, bigEndian bool, compress bool) {
	var num int64
	switch number := number.(type) {
	case int64:
		num = number
	case int32:
		num = int64(number)
	case int16:
		num = int64(number)
	case int8:
		num = int64(number)
	case uint64:
		num = int64(number)
	case uint32:
		num = int64(number)
	case uint16:
		num = int64(number)
	case uint8:
		num = int64(number)
	case uint:
		num = int64(number)
	case int:
		num = int64(number)
	default:
		panic("you need to pass an integer to this function")
	}

	if compress {
		temp := make([]byte, 8)
		binary.BigEndian.PutUint64(temp, uint64(num))
		temp = bytes.TrimLeft(temp, "\x00")
		if size > uint8(len(temp)) {
			size = uint8(len(temp))
		}
		if size == 0 {
			session.outBuffer.WriteByte(0)
			// session.OutBuffer = append(session.OutBuffer, 0)
		} else {
			//if num < 0 {
			//	num = num * -1
			//	size = size & 0x80
			//}
			session.outBuffer.WriteByte(size)
			session.outBuffer.Write(temp[:size])
		}
	} else {
		if size == 1 {
			session.outBuffer.WriteByte(uint8(num))
		} else {
			temp := make([]byte, size)
			if bigEndian {
				switch size {
				case 2:
					binary.BigEndian.PutUint16(temp, uint16(num))
				case 4:
					binary.BigEndian.PutUint32(temp, uint32(num))
				case 8:
					binary.BigEndian.PutUint64(temp, uint64(num))
				}
			} else {
				switch size {
				case 2:
					binary.LittleEndian.PutUint16(temp, uint16(num))
				case 4:
					binary.LittleEndian.PutUint32(temp, uint32(num))
				case 8:
					binary.LittleEndian.PutUint64(temp, uint64(num))
				}
			}
			session.outBuffer.Write(temp)
		}
	}
}

// PutUint write uint number with size entered either use bigEndian or not and use compression or not to
func (session *basicSession) PutUint(number interface{}, size uint8, bigEndian, compress bool) {
	var num uint64
	switch number := number.(type) {
	case int64:
		num = uint64(number)
	case int32:
		num = uint64(number)
	case int16:
		num = uint64(number)
	case int8:
		num = uint64(number)
	case uint64:
		num = number
	case uint32:
		num = uint64(number)
	case uint16:
		num = uint64(number)
	case uint8:
		num = uint64(number)
	case uint:
		num = uint64(number)
	case int:
		num = uint64(number)
	default:
		panic("you need to pass an integer to this function")
	}
	// if the size is one byte no compression occur only one byte written
	if size == 1 {
		session.outBuffer.WriteByte(uint8(num))
		// session.OutBuffer = append(session.OutBuffer, uint8(num))
		return
	}
	if compress {
		temp := make([]byte, 8)
		binary.BigEndian.PutUint64(temp, num)
		temp = bytes.TrimLeft(temp, "\x00")
		if size > uint8(len(temp)) {
			size = uint8(len(temp))
		}
		if size == 0 {
			session.outBuffer.WriteByte(0)
			// session.OutBuffer = append(session.OutBuffer, 0)
		} else {
			session.outBuffer.WriteByte(size)
			session.outBuffer.Write(temp)
			// session.OutBuffer = append(session.OutBuffer, size)
			// session.OutBuffer = append(session.OutBuffer, temp...)
		}
	} else {
		temp := make([]byte, size)
		if bigEndian {
			switch size {
			case 2:
				binary.BigEndian.PutUint16(temp, uint16(num))
			case 4:
				binary.BigEndian.PutUint32(temp, uint32(num))
			case 8:
				binary.BigEndian.PutUint64(temp, num)
			}
		} else {
			switch size {
			case 2:
				binary.LittleEndian.PutUint16(temp, uint16(num))
			case 4:
				binary.LittleEndian.PutUint32(temp, uint32(num))
			case 8:
				binary.LittleEndian.PutUint64(temp, num)
			}
		}
		session.outBuffer.Write(temp)
		// session.OutBuffer = append(session.OutBuffer, temp...)
	}
}

// PutBytes write bytes of data to output buffer
func (session *basicSession) PutBytes(data ...byte) {
	session.outBuffer.Write(data)
}

// PutClr write variable length bytearray to output buffer
func (session *basicSession) PutClr(data []byte) {
	dataLen := len(data)
	if dataLen > 0xFC {
		session.outBuffer.WriteByte(0xFE)
		start := 0
		for start < dataLen {
			end := start + session.ClrChunkSize
			if end > dataLen {
				end = dataLen
			}
			temp := data[start:end]
			if session.UseBigClrChunks {
				session.PutInt(len(temp), 4, true, true)
			} else {
				session.outBuffer.WriteByte(uint8(len(temp)))
			}
			session.outBuffer.Write(temp)
			start += session.ClrChunkSize
		}
		session.outBuffer.WriteByte(0)
	} else if dataLen == 0 {
		session.outBuffer.WriteByte(0)
	} else {
		session.outBuffer.WriteByte(uint8(len(data)))
		session.outBuffer.Write(data)
	}
}

// PutString write a string data to output buffer
func (session *basicSession) PutString(data string) {
	session.PutClr([]byte(data))
}

// PutKeyVal write key, val (in form of bytearray) and flag number to output buffer
func (session *basicSession) PutKeyVal(key []byte, val []byte, num uint8) {
	if len(key) == 0 {
		session.outBuffer.WriteByte(0)
	} else {
		session.PutUint(len(key), 4, true, true)
		session.PutClr(key)
	}
	if len(val) == 0 {
		session.outBuffer.WriteByte(0)
	} else {
		session.PutUint(len(val), 4, true, true)
		session.PutClr(val)
	}
	session.PutInt(num, 4, true, true)
}

// PutKeyValString write key, val (in form of string) and flag number to output buffer
func (session *basicSession) PutKeyValString(key string, val string, num uint8) {
	session.PutKeyVal([]byte(key), []byte(val), num)
}
