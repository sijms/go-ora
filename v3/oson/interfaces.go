package oson

type (
	NumberEncoder interface {
		EncodeNumber(input interface{}) ([]byte, error)
	}
	NumberDecoder interface {
		DecodeNumber(input []byte) (string, error)
	}
	BinaryDoubleEncoder interface {
		EncodeBinaryDouble(input float64) ([]byte, error)
	}
	BinaryDoubleDecoder interface {
		DecodeBinaryDouble(input []byte) (float64, error)
	}
	BinaryFloatEncoder interface {
		EncodeBinaryFloat(input float32) ([]byte, error)
	}
	BinaryFloatDecoder interface {
		DecodeBinaryFloat(input []byte) (float32, error)
	}
	OracleType interface {
		SetValue(v interface{}, typeId uint16) error
		SetBytes(b []byte)
		Value(typeId uint16) (interface{}, error)
		Bytes() []byte

		//Encode(typeId uint16) error
		//Decode(typeId uint16) error
		//Value() interface{}
		//Bytes() []byte
	}
)
