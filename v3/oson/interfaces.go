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
		Encode(input interface{}) ([]byte, error)
		Decode(input []byte) (interface{}, error)
		//Value() interface{}
		//Bytes() []byte
	}
)
