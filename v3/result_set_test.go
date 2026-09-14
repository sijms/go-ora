package go_ora

import (
	"reflect"
	"testing"

	"github.com/sijms/go-ora/v3/types"
)

// TestResultSetColumnsNilSafe verifies that Columns() does not panic when the
// ResultSet is uninitialized (cols == nil). This guards against the nil-pointer
// panic reported in issue #691, where a connection reset during _query() could
// yield a DataSet with an uninitialized ResultSet.
func TestResultSetColumnsNilSafe(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Columns() panicked on nil cols: %v", r)
		}
	}()
	rs := ResultSet{}
	if got := rs.Columns(); got != nil {
		t.Errorf("expected nil columns for uninitialized ResultSet, got %v", got)


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
