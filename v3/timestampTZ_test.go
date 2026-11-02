package go_ora

import (
	"database/sql/driver"
	"testing"
	"time"
)

func TestTimeStampTZValueIsTime(t *testing.T) {
	timeVal := time.Date(2023, 5, 28, 23, 38, 11, 500000000, time.UTC)
	tz := TimeStampTZ(timeVal)
	for name, input := range map[string]interface{}{
		"TimeStampTZ":     tz,
		"*TimeStampTZ":    &tz,
		"NullTimeStampTZ": NullTimeStampTZ{TimeStampTZ: tz, Valid: true},
	} {
		got, err := driver.DefaultParameterConverter.ConvertValue(input)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if gotTime, ok := got.(time.Time); !ok || !gotTime.Equal(timeVal) {
			t.Errorf("%s: expected %v and got %#v", name, timeVal, got)
		}
	}
}
