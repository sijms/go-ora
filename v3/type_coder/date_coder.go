package type_coder

import (
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/types"
)

type DateCoder struct {
	TypeInfo
}

func NewDate(data interface{}) (*DateCoder, error) {
	ret := new(DateCoder)
	ret.SetDefault()
	ret.MaxLen = 0x7
	ret.DataType = types.DATE
	value := sql.NullTime{}
	switch v := data.(type) {
	case time.Time:
		value = sql.NullTime{Time: v, Valid: true}
	case *time.Time:
		value = sql.NullTime{Time: *v, Valid: true}
	case sql.NullTime:
		value = v
	case *sql.NullTime:
		value = *v
	default:
		return nil, fmt.Errorf("time coder: unsupported type %T", data)
	}
	if value.Valid {
		ret.BValue = make([]byte, 7)
		putDate(ret.BValue, &value.Time)
	} else {
		ret.BValue = nil
	}
	return ret, nil
}

func NewTimestamp(data sql.NullTime) *DateCoder {
	ret := new(DateCoder)
	ret.SetDefault()
	ret.MaxLen = 0xB
	ret.DataType = types.TIMESTAMP
	if data.Valid {
		ret.BValue = make([]byte, 11)
		putTimestamp(ret.BValue, &data.Time)
	}
	return ret
}
func NewTimestampTZ(data sql.NullTime, asUTC bool) *DateCoder {
	ret := new(DateCoder)
	ret.SetDefault()
	ret.MaxLen = 0xD
	ret.DataType = types.TIMESTAMPTZ
	if data.Valid {
		value := data.Time
		ret.BValue = make([]byte, 13)
		putTimestamp(ret.BValue, &value)
		zoneLoc := value.Location()
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
			ret.BValue = append(ret.BValue, zone1, zone2)

		} else {
			_, offset := value.Zone()
			zone1 := uint8(offset/3600) + 20
			zone2 := uint8((offset/60)%60) + 60
			ret.BValue = append(ret.BValue, zone1, zone2)
		}
		if asUTC {
			if ret.BValue[11]&0x80 != 0 {
				ret.BValue[12] |= 1
				if value.IsDST() {
					ret.BValue[12] |= 2
				}
			} else {
				ret.BValue[11] |= 0x40
			}
		}
	}
	return ret
}

//func (coder *DateCoder) SendAsUTC(value bool) {
//	coder.asUTC = value
//}

//func (date *Date) Encode() ([]byte, error) {
//
//	if !date.Data.Valid {
//		return nil, nil
//	}
//	value := date.Data.Time
//	if date.asUTC {
//		value = date.Data.Time.UTC()
//	}
//	switch date.DataType {
//	case DATE:
//
//	case TIMESTAMP:
//
//	case TIMESTAMPTZ:
//
//	default:
//		return nil, errors.New("stored type is note date/time type")
//	}
//}

func (coder *DateCoder) DecodeDate(data []byte) (sql.NullTime, error) {
	ret := sql.NullTime{}
	if len(data) == 0 {
		return ret, nil
	}
	if len(data) < 7 {
		return ret, errors.New("abnormal data representation for date/time")
	}
	year := (int(data[0]) - 100) * 100
	year += int(data[1]) - 100
	nanoSec := 0
	tzHour := 0
	tzMin := 0
	if len(data) > 10 {
		nanoSec = int(binary.BigEndian.Uint32(data[7:11]))
	}
	if len(data) > 11 {
		tzHour = int(data[11]&0x3F) - 20
	}
	if len(data) > 12 {
		tzMin = int(data[12]) - 60
	}
	if tzHour == 0 && tzMin == 0 {
		ret.Valid = true
		ret.Time = time.Date(year, time.Month(data[2]), int(data[3]),
			int(data[4]-1), int(data[5]-1), int(data[6]-1), nanoSec, time.UTC)
		return ret, nil
	}
	var zone *time.Location
	var timeInZone bool
	if data[11]&0x80 != 0 {
		regionCode := (int(data[11]) & 0x7F) << 6
		regionCode += (int(data[12]) & 0xFC) >> 2
		timeInZone = data[12]&0x1 == 1
		name, found := oracleZones[regionCode]
		if found {
			zone, _ = time.LoadLocation(name)
		}
	} else {
		timeInZone = data[11]&0x40 == 0x40
	}
	if zone == nil {
		zone = time.FixedZone(fmt.Sprintf("%+03d:%02d", tzHour, tzMin), tzHour*60*60+tzMin*60)
		// timeInZone = true
	}
	ret.Valid = true
	if timeInZone {
		ret.Time = time.Date(year, time.Month(data[2]), int(data[3]),
			int(data[4]-1), int(data[5]-1), int(data[6]-1), nanoSec, zone)
		return ret, nil
	}
	temp := time.Date(year, time.Month(data[2]), int(data[3]),
		int(data[4]-1), int(data[5]-1), int(data[6]-1), nanoSec, time.UTC)
	ret.Time = temp.In(zone)
	return ret, nil

	// for DATE, TIMESTAMP, TimeStampDTY
	//	if !isEqualLoc(conn.dbServerTimeZone, time.UTC) {
	//		par.oPrimValue = time.Date(tempTime.Year(), tempTime.Month(), tempTime.Day(),
	//			tempTime.Hour(), tempTime.Minute(), tempTime.Second(), tempTime.Nanosecond(), conn.dbServerTimeZone)
	//	} else {
	//		par.oPrimValue = tempTime
	//	}

	// for TimeStampeLTZ, TimeStampLTZ_DTY
	//	if !isEqualLoc(conn.dbTimeZone, time.UTC) {
	//		par.oPrimValue = time.Date(tempTime.Year(), tempTime.Month(), tempTime.Day(),
	//			tempTime.Hour(), tempTime.Minute(), tempTime.Second(), tempTime.Nanosecond(), conn.dbTimeZone)
	//	}
}
func (coder *DateCoder) Decode(data []byte) (interface{}, error) {
	if data == nil {
		return nil, nil
	}
	ret, err := coder.DecodeDate(data)
	if err != nil {
		return nil, err
	}
	if !ret.Valid {
		return nil, nil
	}
	return ret.Time, nil
}

func (coder *DateCoder) Read(session network.SessionReader) (interface{}, error) {
	bValue, err := coder.basicRead(session)
	if err != nil {
		return nil, err
	}
	return coder.Decode(bValue)
}

func (coder *DateCoder) Write(session network.SessionWriter) error {
	session.PutClr(coder.BValue)
	return nil
}

//func (date *Date) Scan(value interface{}) error {
//	switch v := value.(type) {
//	case *Date:
//		*date = *v
//	case Date:
//		*date = v
//	case time.Time:
//		date.Data.Time = v
//		date.Data.Valid = true
//	default:
//		return fmt.Errorf("Date column type require time.Time value")
//	}
//	return nil
//}
