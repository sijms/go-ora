package go_ora

import (
	"reflect"
	"testing"

	"github.com/sijms/go-ora/v3/types"
)

// TestColumnTypeScanTypeInlineLOB verifies that ColumnTypeScanType returns
// the correct Go type for columns whose DataType was mutated to LongVarChar
// or LongRaw by writeDefine() when LOB mode is INLINE. This is issue #719.
func TestColumnTypeScanTypeInlineLOB(t *testing.T) {
	cases := []struct {
		name     string
		dataType uint16
		want     reflect.Type
	}{
		{"CLOB_inline", types.LongVarChar, types.TyString},
		{"BLOB_inline", types.LongRaw, types.TyBytes},
		{"LongVarRaw", types.LongVarRaw, types.TyBytes},
		{"CLOB_locator", types.OCIClobLocator, types.TyString},
		{"BLOB_locator", types.OCIBlobLocator, types.TyBytes},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			col := ParameterInfo{}
			col.DataType = c.dataType
			cols := &[]ParameterInfo{col}
			rs := ResultSet{cols: cols}
			got := rs.ColumnTypeScanType(0)
			if got != c.want {
				t.Errorf("ColumnTypeScanType(%v) = %v, want %v", c.dataType, got, c.want)
			}
		})
	}
}
