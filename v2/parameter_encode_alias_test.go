package go_ora

import (
	"reflect"
	"testing"
)

// TestSetDataTypeTypeAlias verifies that setDataType handles type aliases
// (e.g. type StringAlias string) by falling back to the underlying Kind.
// This is issue #699.
func TestSetDataTypeTypeAlias(t *testing.T) {
	type StringAlias string
	type BytesAlias []byte

	conn := &Connection{tcpNego: &TCPNego{}}
	par := &ParameterInfo{}
	aliasType := reflect.TypeOf(StringAlias(""))
	if err := par.setDataType(conn, aliasType, StringAlias("")); err != nil {
		t.Fatalf("setDataType failed for string alias: %v", err)
	}
	if par.DataType != NCHAR {
		t.Errorf("string alias: expected DataType=NCHAR, got %v", par.DataType)
	}

	par2 := &ParameterInfo{}
	bytesType := reflect.TypeOf(BytesAlias(nil))
	if err := par2.setDataType(conn, bytesType, BytesAlias(nil)); err != nil {
		t.Fatalf("setDataType failed for bytes alias: %v", err)
	}
	if par2.DataType != RAW {
		t.Errorf("bytes alias: expected DataType=RAW, got %v", par2.DataType)
	}
}
