package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Date struct {
	Basic
	AsUTC            bool
	DBTimeZone       *time.Location
	DBServerTimeZone *time.Location
}

func NewTimeStamp(t time.Time) *Date {
	ret := new(Date)
	ret.dataType = TIMESTAMP
	_ = ret.SetValue(t)
	return ret
}
func NewTimeStampTZ(t time.Time) *Date {
	ret := new(Date)
	ret.dataType = TimeStampTZ_DTY
	_ = ret.SetValue(t)
	return ret
}

//	func NewTimeStampLTZ(t time.Time) *Date {
//		ret := new(Date)
//		ret.dataType = TimeStampLTZ_DTY
//		_ = ret.SetValue(t)
//		return ret
//
// }
func (date *Date) GetMaxLen() int64 {
	switch date.dataType {
	case DATE:
		return MaxLenDate
	case TIMESTAMP:
		return MaxLenTimeStamp
	default:
		return MaxLenTimeStampTZ
	}
}
func (date *Date) Value() (interface{}, error) {
	if len(date.bValue) == 0 {
		return nil, nil
	}
	return date.decode()
}
func (date *Date) encode(input time.Time) (bytes []byte, err error) {
	if date.dataType == 0 {
		date.dataType = DATE
	}
	switch date.dataType {
	case DATE:
		bytes = make([]byte, MaxLenDate)
		putDate(bytes, &input)
	case TIMESTAMP:
		bytes = make([]byte, MaxLenTimeStamp)
		putTimestamp(bytes, &input)
	//case TimeStampLTZ, TimeStampLTZ_DTY:
	//	//input = input.UTC()
	//	bytes = make([]byte, MaxLenTimeStampTZ)
	//	putTimestamp(bytes, &input)
	case TIMESTAMPTZ, TimeStampTZ_DTY, TimeStampLTZ, TimeStampLTZ_DTY:
		var val time.Time
		if date.AsUTC {
			val = input.UTC()
		} else {
			val = input
		}
		bytes = make([]byte, MaxLenTimeStamp)
		putTimestamp(bytes, &val)
		zoneLoc := input.Location()
		zoneID := 0
		for key, val := range oracleZones {
			if strings.EqualFold(zoneLoc.String(), val) {
				zoneID = key
				break
			}
		}
		if zoneID > 0 {
			zone1 := uint8((zoneID&0x1FC0)>>6) | 0x80
			zone2 := uint8((zoneID & 0x3F) << 2)
			bytes = append(bytes, zone1, zone2)

		} else {
			_, offset := input.Zone()
			zone1 := uint8(offset/3600) + 20
			zone2 := uint8((offset/60)%60) + 60
			bytes = append(bytes, zone1, zone2)
		}
		if !date.AsUTC {
			if bytes[11]&0x80 != 0 {
				bytes[12] |= 1
				if input.IsDST() {
					bytes[12] |= 2
				}
			} else {
				bytes[11] |= 0x40
			}
		}
	}
	return
}

func (date *Date) SetValue(input interface{}) (err error) {
	defer func(input *Date) {
		if input.dataType == 0 {
			input.dataType = DATE
		}
	}(date)
	if input == nil {
		date.bValue = nil
		return
	}
	switch data := input.(type) {
	case Date:
		if date.AsUTC == data.AsUTC {
			*date = data
		} else {
			temp, err := data.decode()
			if err != nil {
				return err
			}
			date.dataType = data.dataType
			return date.SetValue(temp)
		}
	case *Date:
		if date.AsUTC == data.AsUTC {
			*date = *data
		} else {
			temp, err := data.decode()
			if err != nil {
				return err
			}
			date.dataType = data.dataType
			return date.SetValue(temp)
		}
	case time.Time:
		date.bValue, err = date.encode(data)
	case *time.Time:
		date.bValue, err = date.encode(*data)
	case sql.NullTime:
		if data.Valid {
			date.bValue, err = date.encode(data.Time)
		} else {
			date.bValue = nil
		}
	case *sql.NullTime:
		if data.Valid {
			date.bValue, err = date.encode(data.Time)
		} else {
			date.bValue = nil
		}
	default:
		err = fmt.Errorf("cannot set value of type %T into date/time", input)
	}

	return
}
func (date *Date) decode() (output time.Time, err error) {
	if len(date.bValue) < int(MaxLenDate) {
		err = errors.New("abnormal data representation for date/time")
		return
	}
	year := (int(date.bValue[0]) - 100) * 100
	year += int(date.bValue[1]) - 100
	nanoSec := 0
	tzHour := 0
	tzMin := 0
	if len(date.bValue) > 10 {
		nanoSec = int(binary.BigEndian.Uint32(date.bValue[7:11]))
	}
	if len(date.bValue) > 11 {
		tzHour = int(date.bValue[11]&0x3F) - 20
	}
	if len(date.bValue) > 12 {
		tzMin = int(date.bValue[12]) - 60
	}
	output = time.Date(year, time.Month(date.bValue[2]), int(date.bValue[3]),
		int(date.bValue[4]-1), int(date.bValue[5]-1), int(date.bValue[6]-1), nanoSec, time.UTC)
	if tzHour != 0 || tzMin != 0 {

		var zone *time.Location
		var timeInZone bool
		if date.bValue[11]&0x80 != 0 {
			regionCode := (int(date.bValue[11]) & 0x7F) << 6
			regionCode += (int(date.bValue[12]) & 0xFC) >> 2
			timeInZone = date.bValue[12]&0x1 == 1
			name, found := oracleZones[regionCode]
			if found {
				zone, _ = time.LoadLocation(name)
			}
		} else {
			timeInZone = date.bValue[11]&0x40 == 0x40
		}
		if zone == nil {
			zone = time.FixedZone(fmt.Sprintf("%+03d:%02d", tzHour, tzMin), tzHour*60*60+tzMin*60)
		}
		if timeInZone {
			output = time.Date(year, time.Month(date.bValue[2]), int(date.bValue[3]),
				int(date.bValue[4]-1), int(date.bValue[5]-1), int(date.bValue[6]-1), nanoSec, zone)
		} else {
			output = output.In(zone)
		}
	}
	switch date.dataType {
	case DATE, TIMESTAMP, TimeStampDTY:
		if date.DBServerTimeZone != nil && !isEqualLoc(date.DBServerTimeZone, time.UTC) {
			output = time.Date(output.Year(), output.Month(), output.Day(),
				output.Hour(), output.Minute(), output.Second(), output.Nanosecond(), date.DBServerTimeZone)
		}
	case TimeStampLTZ, TimeStampLTZ_DTY:
		if date.DBTimeZone != nil && !isEqualLoc(date.DBTimeZone, time.UTC) {
			output = time.Date(output.Year(), output.Month(), output.Day(),
				output.Hour(), output.Minute(), output.Second(), output.Nanosecond(), date.DBTimeZone)
		}

	}
	return
}
func (date *Date) Scan(value interface{}) (err error) {
	return date.SetValue(value)
}

func (date *Date) CopyTo(dest driver.Value) (err error) {
	value, err := date.Value()
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
		return fmt.Errorf("cannot copy Date to variable of type %T", dest)
	}
	return nil
}
