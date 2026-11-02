package go_ora

import (
	"database/sql/driver"
	"testing"
	"time"
)

func TestCheckNamedValueKeepsTimeStamp(t *testing.T) {
	ts := TimeStamp(time.Date(2023, 5, 28, 23, 38, 11, 500000000, time.UTC))
	checkers := map[string]driver.NamedValueChecker{"Connection": &Connection{}, "Stmt": &Stmt{}}
	for name, checker := range checkers {
		for _, value := range []driver.Value{ts, NullTimeStamp{TimeStamp: ts, Valid: true}} {
			if err := checker.CheckNamedValue(&driver.NamedValue{Value: value}); err != nil {
				t.Errorf("%s with %T: expected nil so the driver encodes it as TIMESTAMP and got %v", name, value, err)
			}
		}
	}
}
