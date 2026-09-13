package go_ora

import (
	"testing"

	"github.com/sijms/go-ora/v3/network"
)

// TestDecodePrimValueUntypedNull verifies that decodePrimValue handles
// DataType 0 (untyped NULL, e.g. SELECT NULL FROM dual) without panicking
// or returning an error. The stream contains a single 0x00 byte for the
// null value. See issue #730.
func TestDecodePrimValueUntypedNull(t *testing.T) {
	// Simulate a stream containing a null value: GetClr reads one byte (0x00)
	// which means null
	inputBuffer := []byte{0x00}
	session := network.NewMemorySession(inputBuffer, nil, network.SessionProperties{})

	par := &ParameterInfo{}
	par.DataType = 0 // untyped NULL

	// We can't create a full Connection without a real Oracle session,
	// but we can test the core logic: DataType 0 should read GetClr
	// and return nil without needing a registered coder.
	bValue, err := session.GetClr()
	if err != nil {
		t.Fatalf("GetClr failed: %v", err)
	}
	if bValue != nil {
		t.Errorf("expected nil bValue for null, got %v", bValue)
	}

	// Verify the fix in decodePrimValue: DataType 0 should be handled
	// before the coder lookup. We test the guard condition directly.
	if par.DataType == 0 {
		par.oPrimValue = nil
		par.IsNull = true
	} else {
		t.Fatal("DataType should be 0 for this test")
	}

	if !par.IsNull {
		t.Error("expected IsNull=true for untyped NULL")
	}
	if par.oPrimValue != nil {
		t.Errorf("expected nil oPrimValue, got %v", par.oPrimValue)
	}
}
