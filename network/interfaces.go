package network

import "bytes"

type BufferWriter interface {
	WriteInt(buffer *bytes.Buffer, number interface{}, size uint8, bigEndian, compress bool)
	WriteUint(buffer *bytes.Buffer, number interface{}, size uint8, bigEndian, compress bool)
	WriteBytes(buffer *bytes.Buffer, data ...byte)
	WriteClr(buffer *bytes.Buffer, data []byte)
	WriteKeyVal(buffer *bytes.Buffer, key []byte, val []byte, num uint8)
	WriteKeyValString(buffer *bytes.Buffer, key string, val string, num uint8)
}
type Encoder interface {
	Encode(writer BufferWriter) ([]byte, error)
}

type Decoder interface {
	Decode(writer BufferWriter, data []byte) error
}
