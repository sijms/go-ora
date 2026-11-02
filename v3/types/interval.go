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
	// leading (day or year) and fractional-second precision of the column; precision 0 means
	// unknown, as for output parameters and UDT attributes, which carry no column metadata.
	precision, scale uint8
}

// SetPrecisionScale passes the column's leading and fractional precision so Value prints the same width as Oracle.
func (interval *Interval) SetPrecisionScale(precision, scale uint8) {
	interval.precision, interval.scale = precision, scale
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
	typeId := interval.dataType
	if typeId == 0 {
		if len(interval.bValue) >= int(MaxLenIntervalDS) {
			typeId = INTERVALDS_DTY
		} else if len(interval.bValue) >= int(MaxLenIntervalYM) {
			typeId = INTERVALYM_DTY
		}
	}
	switch typeId {
	case INTERVALYM_DTY:
		if len(interval.bValue) < int(MaxLenIntervalYM) {
			return nil, fmt.Errorf("interval data length is too short")
		}
		return interval.formatYM(), nil
	case INTERVALDS_DTY:
		if len(interval.bValue) < int(MaxLenIntervalDS) {
			return nil, fmt.Errorf("interval data length is too short")
		}
		return interval.formatDS(), nil
	default:
		return nil, fmt.Errorf("unsupported type id: %d used for decoding interval", typeId)
	}
}

// leadWidth is the column's day or year precision; Oracle's default of 2 when it is unknown.
func (interval *Interval) leadWidth() int {
	if interval.precision == 0 {
		return 2
	}
	return int(interval.precision)
}

// formatYM renders INTERVAL YEAR TO MONTH as Oracle prints it ("+21-10"); a time.Time cannot hold a signed duration.
func (interval *Interval) formatYM() string {
	val := interval.bValue
	years := int64(binary.BigEndian.Uint32(val)) - 0x80000000
	if years >= 0 && val[4] >= 60 {
		return fmt.Sprintf("+%0*d-%02d", interval.leadWidth(), years, val[4]-60)
	}
	return fmt.Sprintf("-%0*d-%02d", interval.leadWidth(), -years, 60-val[4])
}

// formatDS renders INTERVAL DAY TO SECOND as Oracle prints it ("+02 12:23:34.456000").
func (interval *Interval) formatDS() string {
	val := interval.bValue
	days := int64(binary.BigEndian.Uint32(val)) - 0x80000000
	ns := int64(binary.BigEndian.Uint32(val[7:])) - 0x80000000
	sign, clock := "+", int64(val[4]-60)*3600+int64(val[5]-60)*60+int64(val[6]-60)
	if days < 0 || ns < 0 || val[4] < 60 || val[5] < 60 || val[6] < 60 {
		sign, days, ns, clock = "-", -days, -ns, int64(60-val[4])*3600+int64(60-val[5])*60+int64(60-val[6])
	}
	digits := interval.fractionDigits(ns)
	unit := int64(1)
	for i := digits; i < 9; i++ {
		unit *= 10
	}
	// Expressions such as CAST(... AS INTERVAL DAY TO SECOND(6)) still carry nine digits; Oracle rounds them.
	const nsPerDay = int64(86400) * 1e9
	subDay := (clock*1e9 + ns + unit/2) / unit * unit
	days += subDay / nsPerDay
	subDay %= nsPerDay
	clock, ns = subDay/1e9, subDay%1e9
	text := fmt.Sprintf("%s%0*d %02d:%02d:%02d", sign, interval.leadWidth(), days, clock/3600, clock/60%60, clock%60)
	if digits > 0 {
		text += fmt.Sprintf(".%0*d", digits, ns/unit)
	}
	return text
}

// fractionDigits is the column's fractional precision. Without column metadata it keeps
// Oracle's default of six digits unless that would drop nanoseconds.
func (interval *Interval) fractionDigits(ns int64) int {
	// Oracle allows 0-9 fraction digits; anything else (0xFF from a -127 describe) is unknown.
	if interval.precision != 0 && interval.scale <= 9 {
		return int(interval.scale)
	}
	if ns%1000 != 0 {
		return 9
	}
	return 6
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
	case *string:
		if value != nil {
			*dst = value.(string)
		}
	case *sql.NullString:
		if value != nil {
			*dst = sql.NullString{Valid: true, String: value.(string)}
		} else {
			*dst = sql.NullString{Valid: false}
		}
	default:
		return fmt.Errorf("cannot copy Interval to variable of type %T", dest)
	}
	return nil
}
