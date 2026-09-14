package go_ora

import (
	"reflect"
	"testing"
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
