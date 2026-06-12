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
	//Year        int
	//Month       int
	//Day         int
	//Hour        int
	//Minute      int
	//Second      int
	//Microsecond int
	bValue []byte
}

//func NewYearMonthInterval(year, month int) *Interval {
//	return &Interval{
//		Year:  year,
//		Month: month,
//	}
//}

//func NewDaySecondInterval(day, hour, minute, second, nanosecond int) *Interval {
//	return &Interval{
//		Day:         day,
//		Hour:        hour,
//		Minute:      minute,
//		Second:      second,
//		Microsecond: nanosecond,
//	}
//}

func (interval *Interval) Value(typeId uint16) (interface{}, error) {
	if len(interval.bValue) == 0 {
		return nil, nil
	}
	var (
		year, month, day, hour, minute, second, mSec int
	)
	if typeId == 0 {
		if len(interval.bValue) >= 0xB {
			typeId = INTERVALDS_DTY
		}
		if len(interval.bValue) >= 5 {
			typeId = INTERVALYM_DTY
		}
	}
	switch typeId {
	case INTERVALYM_DTY:
		if len(interval.bValue) < 5 {
			return nil, fmt.Errorf("interval data length is too short")
		}
		year = int(binary.BigEndian.Uint32(interval.bValue)) - 0x80000000
		month = int(interval.bValue[4] - 60)
		return time.Date(year, time.Month(month), 0, 0, 0, 0, 0, time.UTC), nil
	case INTERVALDS_DTY:
		if len(interval.bValue) < 0xB {
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

func (interval *Interval) encode(input time.Time, typeId uint16) error {
	buffer := new(bytes.Buffer)
	switch typeId {
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
			return interval.SetValue(interval, INTERVALYM_DTY)
		}
		return interval.SetValue(interval, INTERVALDS_DTY)
	}
	interval.bValue = buffer.Bytes()
	return nil
}
func (interval *Interval) SetValue(input interface{}, typeId uint16) error {
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
		return interval.encode(data, typeId)
	case *time.Time:
		return interval.encode(*data, typeId)
	default:
		return fmt.Errorf("cannot set value of type %T into interval", input)
	}
	return nil
}

func (interval *Interval) Bytes() []byte {
	return interval.bValue
}

func (interval *Interval) SetBytes(input []byte) {
	interval.bValue = input
}

func (interval *Interval) Scan(value interface{}) error {
	return interval.SetValue(value, 0)
}

func (interval *Interval) CopyTo(dest driver.Value) error {
	value, err := interval.Value(0)
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
