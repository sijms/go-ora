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
		typeId = interval.guessDataType()
	}
	switch typeId {
	case INTERVALYM_DTY, IntervalYM:
		if len(interval.bValue) < int(MaxLenIntervalYM) {
			return nil, fmt.Errorf("interval data length is too short")
		}
		year = int(int64(binary.BigEndian.Uint32(interval.bValue)) - 0x80000000)
		month = int(interval.bValue[4]) - 60
		return time.Date(year, time.Month(month), 0, 0, 0, 0, 0, time.UTC), nil
	case INTERVALDS_DTY, IntervalDS:
		if len(interval.bValue) < int(MaxLenIntervalDS) {
			return nil, fmt.Errorf("interval data length is too short")
		}
		day = int(int64(binary.BigEndian.Uint32(interval.bValue)) - 0x80000000)
		hour = int(interval.bValue[4]) - 60
		minute = int(interval.bValue[5]) - 60
		second = int(interval.bValue[6]) - 60
		mSec = int(int64(binary.BigEndian.Uint32(interval.bValue[7:])) - 0x80000000)
		return time.Date(0, 0, day, hour, minute, second, mSec*1000, time.UTC), nil
	default:
		return nil, fmt.Errorf("unsupported type id: %d used for decoding interval", typeId)
	}
}

// guessDataType tells the two interval types apart by their encoded length.
func (interval *Interval) guessDataType() uint16 {
	if len(interval.bValue) >= int(MaxLenIntervalDS) {
		return INTERVALDS_DTY
	}
	return INTERVALYM_DTY
}

// Text returns the interval as Oracle writes it: "+02-03" for INTERVAL YEAR TO
// MONTH and "+04 05:06:07.000000" for INTERVAL DAY TO SECOND, with nine
// fractional digits where the value has nanoseconds. It returns nil for NULL.
//
// INTERVAL_xxx encoding described at https://www.orafaq.com/wiki/Interval
func (interval *Interval) Text() (interface{}, error) {
	val := interval.bValue
	if len(val) == 0 {
		return nil, nil
	}
	typeId := interval.dataType
	if typeId == 0 {
		typeId = interval.guessDataType()
	}
	switch typeId {
	case INTERVALYM_DTY, IntervalYM:
		if len(val) < int(MaxLenIntervalYM) {
			return nil, fmt.Errorf("interval data length is too short")
		}
		// years + 0x80000000, then months + 60
		years := int64(binary.BigEndian.Uint32(val)) - 0x80000000
		months := int64(val[4]) - 60
		sign := "+"
		if years < 0 || months < 0 {
			sign, years, months = "-", -years, -months
		}
		return fmt.Sprintf("%s%02d-%02d", sign, years, months), nil
	case INTERVALDS_DTY, IntervalDS:
		if len(val) < int(MaxLenIntervalDS) {
			return nil, fmt.Errorf("interval data length is too short")
		}
		// days + 0x80000000, hours + 60, minutes + 60, seconds + 60, then
		// nanoseconds + 0x80000000
		days := int64(binary.BigEndian.Uint32(val)) - 0x80000000
		hours := int64(val[4]) - 60
		minutes := int64(val[5]) - 60
		seconds := int64(val[6]) - 60
		nanoseconds := int64(binary.BigEndian.Uint32(val[7:])) - 0x80000000
		sign := "+"
		if days < 0 || hours < 0 || minutes < 0 || seconds < 0 || nanoseconds < 0 {
			sign = "-"
			days, hours, minutes, seconds, nanoseconds = -days, -hours, -minutes, -seconds, -nanoseconds
		}
		fraction := fmt.Sprintf("%06d", nanoseconds/1000)
		if nanoseconds%1000 != 0 {
			fraction = fmt.Sprintf("%09d", nanoseconds)
		}
		return fmt.Sprintf("%s%02d %02d:%02d:%02d.%s", sign, days, hours, minutes, seconds, fraction), nil
	default:
		return nil, fmt.Errorf("unsupported type id: %d used for decoding interval", typeId)
	}
}

func (interval *Interval) encode(input time.Time) error {
	buffer := new(bytes.Buffer)
	switch interval.dataType {
	case INTERVALYM_DTY:
		err := binary.Write(buffer, binary.BigEndian, uint32(int64(input.Year())+0x80000000))
		if err != nil {
			return err
		}
		err = buffer.WriteByte(uint8(input.Month() + 60))
		if err != nil {
			return err
		}
	case INTERVALDS_DTY:
		err := binary.Write(buffer, binary.BigEndian, uint32(int64(input.Day())+0x80000000))
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
		err = binary.Write(buffer, binary.BigEndian, uint32(int64(input.Nanosecond()/1000)+0x80000000))
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
