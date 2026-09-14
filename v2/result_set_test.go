package go_ora

import (
	"reflect"
	"testing"
)

// TestColumnTypeScanTypeInlineLOB verifies that ColumnTypeScanType returns
// the correct Go type for columns whose DataType was mutated to LongVarChar
// or LongRaw by writeDefine() when LOB mode is INLINE. This is issue #719.
func TestColumnTypeScanTypeInlineLOB(t *testing.T) {
	cases := []struct {
		name     string
		dataType TNSType
		want     reflect.Type
	}{
		{"CLOB_inline", LongVarChar, tyString},
		{"BLOB_inline", LongRaw, tyBytes},
		{"LongVarRaw", LongVarRaw, tyBytes},
		{"CLOB_locator", OCIClobLocator, tyString},
		{"BLOB_locator", OCIBlobLocator, tyBytes},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cols := &[]ParameterInfo{{DataType: c.dataType}}
			rs := ResultSet{cols: cols}
			got := rs.ColumnTypeScanType(0)
			if got != c.want {
				t.Errorf("ColumnTypeScanType(%v) = %v, want %v", c.dataType, got, c.want)
			}
		})
	}
}
