package go_ora

import (
	"reflect"
	"testing"

	"github.com/sijms/go-ora/v3/parameter_coder"
	"github.com/sijms/go-ora/v3/types"
)

// TestGetParameterCoderTypeAlias verifies that GetParameterCoder handles
// type aliases (e.g. type StringAlias string) by falling back to the
// underlying Kind. This is issue #699.
func TestGetParameterCoderTypeAlias(t *testing.T) {
	type StringAlias string
	type BytesAlias []byte

	conn := &Connection{
		goTypeCoder: map[reflect.Type]parameter_coder.OracleParameterCoder{
			types.TyString: &parameter_coder.StringParameter{},
			types.TyBytes:  &parameter_coder.RawParameter{},
		},
	}

	coder, err := conn.GetParameterCoder(reflect.TypeOf(StringAlias("")))
	if err != nil {
		t.Fatalf("GetParameterCoder failed for string alias: %v", err)
	}
	if coder == nil {
		t.Fatal("expected non-nil coder for string alias")
	}

	coder2, err := conn.GetParameterCoder(reflect.TypeOf(BytesAlias(nil)))
	if err != nil {
		t.Fatalf("GetParameterCoder failed for bytes alias: %v", err)
	}
	if coder2 == nil {
		t.Fatal("expected non-nil coder for bytes alias")
	}
}
