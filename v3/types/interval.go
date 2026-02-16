package types

import (
	"fmt"
)

type Interval struct {
	Year        int
	Month       int
	Day         int
	Hour        int
	Minute      int
	Second      int
	Microsecond int
}

func NewYearMonthInterval(year, month int) *Interval {
	return &Interval{
		Year:  year,
		Month: month,
	}
}

func NewDaySecondInterval(day, hour, minute, second, nanosecond int) *Interval {
	return &Interval{
		Day:         day,
		Hour:        hour,
		Minute:      minute,
		Second:      second,
		Microsecond: nanosecond,
	}
}

func (interval *Interval) Scan(value interface{}) error {
	switch v := value.(type) {
	case *Interval:
		*interval = *v
	case Interval:
		*interval = v
	case string:
		// TODO: parse string to interval
		return nil
	default:
		return fmt.Errorf("interval column type require string value")
	}
	return nil
}
