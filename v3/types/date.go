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
	asUTC bool
}

func (date *Date) Value() (interface{}, error) {
	if len(date.bValue) == 0 {
		return nil, nil
	}
	return date.decode()
}
func (date *Date) encode(input time.Time, typeId uint16) (bytes []byte, err error) {
	switch typeId {
	case DATE:
		bytes = make([]byte, 7)
		putDate(bytes, &input)
	case TIMESTAMP:
		bytes = make([]byte, 0xB)
		putTimestamp(bytes, &input)
	case TIMESTAMPTZ:
		bytes = make([]byte, 0xD)
		putTimestamp(bytes, &input)
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
		if date.asUTC {
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
	if input == nil {
		date.bValue = nil
		return
	}
	switch data := input.(type) {
	case Date:
		*date = data
	case *Date:
		*date = *data
	case time.Time:
		date.bValue, err = date.encode(data, date.dataType)
	case *time.Time:
		date.bValue, err = date.encode(*data, date.dataType)
	case sql.NullTime:
		if data.Valid {
			date.bValue, err = date.encode(data.Time, date.dataType)
		} else {
			date.bValue = nil
		}
	case *sql.NullTime:
		if data.Valid {
			date.bValue, err = date.encode(data.Time, date.dataType)
		} else {
			date.bValue = nil
		}
	default:
		err = fmt.Errorf("cannot set value of type %T into date/time", input)
	}
	return
}
func (date *Date) decode() (output time.Time, err error) {
	if len(date.bValue) < 7 {
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
	if tzHour == 0 && tzMin == 0 {
		return
	}

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
		return
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
