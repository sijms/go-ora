package type_coder

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type IntervalCoder struct {
	TypeInfo
}

func NewIntervalCoder(data types.Interval) (*IntervalCoder, error) {
	ret := &IntervalCoder{}
	buffer := &bytes.Buffer{}
	if data.Year != 0 || data.Month != 0 {
		ret.DataType = types.INTERVALYM_DTY
		ret.MaxLen = 0x5
		err := binary.Write(buffer, binary.BigEndian, uint32(data.Year+0x80000000))
		if err != nil {
			return nil, err
		}
		err = buffer.WriteByte(uint8(data.Month + 60))
		if err != nil {
			return nil, err
		}
	} else {
		ret.DataType = types.INTERVALDS_DTY
		ret.MaxLen = 0xB
		err := binary.Write(buffer, binary.BigEndian, uint32(data.Day+0x80000000))
		if err != nil {
			return nil, err
		}
		err = buffer.WriteByte(uint8(data.Hour + 60))
		if err != nil {
			return nil, err
		}
		err = buffer.WriteByte(uint8(data.Minute + 60))
		if err != nil {
			return nil, err
		}
		err = buffer.WriteByte(uint8(data.Second + 60))
		if err != nil {
			return nil, err
		}
		err = binary.Write(buffer, binary.BigEndian, uint32(data.Microsecond+0x80000000))
		if err != nil {
			return nil, err
		}
	}
	ret.BValue = buffer.Bytes()
	return ret, nil
}

func (coder *IntervalCoder) Decode(data []byte) (interface{}, error) {
	if data == nil {
		return nil, nil
	}
	interval := types.Interval{}
	switch coder.DataType {
	case types.INTERVALYM_DTY:
		if len(data) < 5 {
			return nil, fmt.Errorf("error decoding interval year to month buffer length: %d and required length: %d", len(data), 5)
		}

		interval.Year = int(binary.BigEndian.Uint32(data)) - 0x80000000
		interval.Month = int(data[4] - 60)
		return interval, nil
	case types.INTERVALDS_DTY:
		if len(data) < 11 {
			return nil, fmt.Errorf("error decoding interval day to second buffer length: %d and required length: %d", len(data), 5)
		}
		interval.Day = int(binary.BigEndian.Uint32(data) - 0x80000000)
		interval.Hour = int(data[4] - 60)
		interval.Minute = int(data[5] - 60)
		interval.Second = int(data[6] - 60)
		interval.Microsecond = int(binary.BigEndian.Uint32(data[7:]) - 0x80000000)
		return interval, nil
	default:
		return nil, fmt.Errorf("call interval decoder with unsupported type: %v", coder.DataType)
	}
}
func (coder *IntervalCoder) Read(session network.SessionReader) (interface{}, error) {
	bValue, err := coder.basicRead(session)
	if err != nil {
		return nil, err
	}
	return coder.Decode(bValue)
}

func (coder *IntervalCoder) Write(session network.SessionWriter) error {
	session.PutClr(coder.BValue)
	return nil
}
