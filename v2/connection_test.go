package go_ora

import (
	"context"
	"database/sql/driver"
	"testing"
)

func TestValidatorInterface(t *testing.T) {
	// Test that Connection implements driver.Validator interface
	var conn interface{} = &Connection{}
	if _, ok := conn.(driver.Validator); !ok {
		t.Error("Connection does not implement driver.Validator interface")
	}
}

func TestSessionResetterInterface(t *testing.T) {
	// Test that Connection implements driver.SessionResetter interface  
	var conn interface{} = &Connection{}
	if _, ok := conn.(driver.SessionResetter); !ok {
		t.Error("Connection does not implement driver.SessionResetter interface")
	}
}

func TestIsValidMethod(t *testing.T) {
	// Create a basic connection instance
	conn := &Connection{
		State: Closed,
		bad:   false,
	}

	// Test closed connection is invalid
	if conn.IsValid() {
		t.Error("Expected closed connection to be invalid")
	}

	// Test opened connection is valid
	conn.State = Opened
	if !conn.IsValid() {
		t.Error("Expected opened connection to be valid")
	}

	// Test bad connection is invalid even when opened
	conn.bad = true
	if conn.IsValid() {
		t.Error("Expected bad connection to be invalid")
	}

	// Test connection is valid when opened and not bad
	conn.bad = false
	conn.State = Opened
	if !conn.IsValid() {
		t.Error("Expected opened, non-bad connection to be valid")
	}
}

func TestResetSessionMethod(t *testing.T) {
	ctx := context.Background()

	// Test reset session on good connection
	conn := &Connection{
		State: Opened,
		bad:   false,
	}

	err := conn.ResetSession(ctx)
	if err != nil {
		t.Errorf("Expected no error on good connection, got: %v", err)
	}

	// Test reset session on bad connection returns ErrBadConn
	conn.bad = true
	err = conn.ResetSession(ctx)
	if err != driver.ErrBadConn {
		t.Errorf("Expected driver.ErrBadConn on bad connection, got: %v", err)
	}
}
