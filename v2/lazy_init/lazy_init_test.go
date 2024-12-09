package lazy_init

import (
	"fmt"
	"testing"
)

func TestNewLazyInit_SuccessInitialize(t *testing.T) {
	lazy := NewLazyInit(func() (interface{}, error) {
		return 123, nil
	})
	v, err := lazy.GetValue()
	if err != nil {
		t.Fatalf("error should be nil")
	}
	if v != 123 {
		t.Fatalf("value should be 123")
	}

	v2, err := lazy.GetValue()
	if err != nil {
		t.Fatalf("error should be nil")
	}
	if v2 != 123 {
		t.Fatalf("value should be 123")
	}
}

func TestNewLazyInit_FailInitialize(t *testing.T) {
	lazy := NewLazyInit(func() (interface{}, error) {
		return 0, fmt.Errorf("error")
	})
	v, err := lazy.GetValue()
	if err == nil {
		t.Fatalf("error should be set on first call")
	}
	if v != 0 {
		t.Fatalf("value should be 0  on first call")
	}

	v2, err := lazy.GetValue()
	if err == nil {
		t.Fatalf("error should be set on second call")
	}
	if v2 != 0 {
		t.Fatalf("value should be 0  on second call")
	}
}
