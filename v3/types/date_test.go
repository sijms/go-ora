package types

import (
	"testing"
	"time"
)

func TestDateSetValueNullDateWithDifferentAsUTC(t *testing.T) {
	src := NewTimeStamp(time.Now())
	if err := src.SetValue(nil); err != nil {
		t.Fatal(err)
	}
	for _, input := range []interface{}{*src, src} {
		dst := Date{AsUTC: !src.AsUTC}
		if err := dst.SetValue(input); err != nil {
			t.Fatalf("%T: %v", input, err)
		}
		if dst.bValue != nil || dst.dataType != TIMESTAMP {
			t.Errorf("%T: expected null TIMESTAMP and got %v with type %d", input, dst.bValue, dst.dataType)
		}
	}
}
