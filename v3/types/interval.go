package types

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"time"
)

type Interval struct {
	Basic
}

func NewYearMonthInterval(year, month int) (*Interval, error) {
	ret := &Interval{}
	ret.dataType = INTERVALYM_DTY
	input := time.Date(year, time.Month(month), 0, 0, 0, 0, 0, time.UTC)
	err := ret.SetValue(input)
	return ret, err
}

func NewDaySecondInterval(day, hour, minute, second, nanosecond int) (*Interval, error) {
	ret := &Interval{}
	ret.dataType = INTERVALDS_DTY
	input := time.Date(0, 0, day, hour, minute, second, nanosecond, time.UTC)
	err := ret.SetValue(input)
	return ret, err
}

func (interval *Interval) GetMaxLen() int64 {
	switch interval.dataType {
	case INTERVALYM_DTY:
		return MaxLenIntervalYM
	default:
		return MaxLenIntervalDS
	}
}
func (interval *Interval) Value() (interface{}, error) {
	if len(interval.bValue) == 0 {
		return nil, nil
	}
	var (
		year, month, day, hour, minute, second, mSec int
		typeId                                       = interval.dataType
	)
	if typeId == 0 {
		if len(interval.bValue) >= int(MaxLenIntervalDS) {
			typeId = INTERVALDS_DTY
		}
		if len(interval.bValue) >= int(MaxLenIntervalYM) {
			typeId = INTERVALYM_DTY
		}
	}
	switch typeId {
	case INTERVALYM_DTY:
		if len(interval.bValue) < int(MaxLenIntervalYM) {
			return nil, fmt.Errorf("interval data length is too short")
		}
		year = int(binary.BigEndian.Uint32(interval.bValue)) - 0x80000000
		month = int(interval.bValue[4] - 60)
		return time.Date(year, time.Month(month), 0, 0, 0, 0, 0, time.UTC), nil
	case INTERVALDS_DTY:
		if len(interval.bValue) < int(MaxLenIntervalDS) {
			return nil, fmt.Errorf("interval data length is too short")
		}
		day = int(binary.BigEndian.Uint32(interval.bValue)) - 0x80000000
		hour = int(interval.bValue[4] - 60)
		minute = int(interval.bValue[5] - 60)
		second = int(interval.bValue[6] - 60)
		mSec = int(binary.BigEndian.Uint32(interval.bValue[7:]) - 0x80000000)
		return time.Date(0, 0, day, hour, minute, second, mSec*1000, time.UTC), nil
	default:
		return nil, fmt.Errorf("unsupported type id: %d used for decoding interval", typeId)
	}
}

func (interval *Interval) encode(input time.Time) error {
	buffer := new(bytes.Buffer)
	switch interval.dataType {
	case INTERVALYM_DTY:
		err := binary.Write(buffer, binary.BigEndian, uint32(input.Year()+0x80000000))
		if err != nil {
			return err
		}
		err = buffer.WriteByte(uint8(input.Month() + 60))
		if err != nil {
			return err
		}
	case INTERVALDS_DTY:
		err := binary.Write(buffer, binary.BigEndian, uint32(input.Day()+0x80000000))
		if err != nil {
			return err
		}
		err = buffer.WriteByte(uint8(input.Hour() + 60))
		if err != nil {
			return err
		}
		err = buffer.WriteByte(uint8(input.Minute() + 60))
		if err != nil {
			return err
		}
		err = buffer.WriteByte(uint8(input.Second() + 60))
		if err != nil {
			return err
		}
		err = binary.Write(buffer, binary.BigEndian, uint32((input.Nanosecond()/1000)+0x80000000))
		if err != nil {
			return err
		}
	default:
		if input.Year() != 0 || input.Month() != 0 {
			interval.SetDataType(INTERVALYM_DTY)
			return interval.encode(input)
		}
		interval.SetDataType(INTERVALDS_DTY)
		return interval.encode(input)
	}
	interval.bValue = buffer.Bytes()
	return nil
}
func (interval *Interval) SetValue(input interface{}) error {
	if input == nil {
		interval.bValue = nil
		return nil
	}
	switch data := input.(type) {
	case Interval:
		*interval = data
	case *Interval:
		*interval = *data
	case time.Time:
		return interval.encode(data)
	case *time.Time:
		return interval.encode(*data)
	default:
		return fmt.Errorf("cannot set value of type %T into interval", input)
	}
	return nil
}

func (interval *Interval) Scan(value interface{}) error {
	return interval.SetValue(value)
}

func (interval *Interval) CopyTo(dest driver.Value) error {
	value, err := interval.Value()
	if err != nil {
		return err
	}
	switch dst := dest.(type) {
	case *time.Time:
		if value != nil {
			*dst = value.(time.Time)
		}
	case *sql.NullTime:
		if value != nil {
			*dst = sql.NullTime{Valid: true, Time: value.(time.Time)}
		} else {
			*dst = sql.NullTime{Valid: false}
		}
	default:
		return fmt.Errorf("cannot copy Interval to variable of type %T", dest)
	}
	return nil
}
