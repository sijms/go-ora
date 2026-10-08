package go_ora

import (
	"math"
	"reflect"
	"testing"

	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/trace"
	"github.com/sijms/go-ora/v3/types"
)

// stageColumnMetadata populates the metadata staging fields that load() and
// captureColumnMetadata() maintain in production, for hand-built test columns.
func stageColumnMetadata(col *ParameterInfo, fetch configurations.LobFetch, readMode configurations.LobReadMode) {
	col.lobFetch = fetch
	col.lobFetchKnown = true
	col.lobReadMode = readMode
	col.lobReadModeKnown = true
	if !col.originalDataTypeKnown {
		col.originalDataType = col.DataType
		col.originalDataTypeKnown = true
	}
	col.nullabilityKnown = true
}

func metadataColumn(dataType uint16) ParameterInfo {
	col := ParameterInfo{}
	col.DataType = dataType
	stageColumnMetadata(&col, configurations.STREAM, configurations.LobReadMode_NONE)
	return col
}

// metadataResultSet builds a ResultSet from columns through the production
// descriptor constructor used by defaultStmt.read().
func metadataResultSet(cols ...ParameterInfo) *ResultSet {
	rs := &ResultSet{cols: &cols}
	rs.buildColumnDescriptors()
	return rs
}

// descriptorResultSet wraps pre-built descriptors into a ResultSet.
func descriptorResultSet(descs ...columnDescriptor) *ResultSet {
	return &ResultSet{descriptors: descs}
}

func TestResultSetColumnsNilSafe(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Columns() panicked on nil cols: %v", r)
		}
	}()
	rs := ResultSet{}
	if got := rs.Columns(); got != nil {
		t.Errorf("expected nil columns for uninitialized ResultSet, got %v", got)
	}
}

// TestColumnTypeScanTypeInlineLOB verifies that ColumnTypeScanType returns
// the correct Go type for columns whose wire DataType was replaced by a LONG
// define type when LOB mode is INLINE. This is issue #719.
func TestColumnTypeScanTypeInlineLOB(t *testing.T) {
	cases := []struct {
		name     string
		dataType uint16
		want     reflect.Type
	}{
		{"CLOB_inline", types.LongVarChar, types.TyString},
		{"BLOB_inline", types.LongRaw, types.TyBytes},
		// wire 95 has no registered result decoder, so nil is the honest
		// scan type (upstream reported []byte optimistically)
		{"LongVarRaw", types.LongVarRaw, nil},
		{"CLOB_locator", types.OCIClobLocator, types.TyString},
		{"BLOB_locator", types.OCIBlobLocator, types.TyBytes},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			col := ParameterInfo{}
			col.DataType = c.dataType
			rs := descriptorResultSet(newColumnDescriptor(&col))
			got := rs.ColumnTypeScanType(0)
			if got != c.want {
				t.Errorf("ColumnTypeScanType(%v) = %v, want %v", c.dataType, got, c.want)
			}
		})
	}
}

func TestResultSetColumnMetadataInvalidState(t *testing.T) {
	cols := []ParameterInfo{metadataColumn(types.NCHAR)}
	cases := []struct {
		name string
		rs   *ResultSet
	}{
		{"nil_result_set", nil},
		{"nil_columns", &ResultSet{}},
		{"empty_columns", metadataResultSet()},
		{"columns_without_descriptors", &ResultSet{cols: &cols}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, index := range []int{-1, 0, 1} {
				if got := tc.rs.ColumnTypeDatabaseTypeName(index); got != "" {
					t.Errorf("DatabaseTypeName(%d) = %q", index, got)
				}
				if length, ok := tc.rs.ColumnTypeLength(index); length != 0 || ok {
					t.Errorf("Length(%d) = (%d,%v)", index, length, ok)
				}
				if nullable, ok := tc.rs.ColumnTypeNullable(index); nullable || ok {
					t.Errorf("Nullable(%d) = (%v,%v)", index, nullable, ok)
				}
				if precision, scale, ok := tc.rs.ColumnTypePrecisionScale(index); precision != 0 || scale != 0 || ok {
					t.Errorf("PrecisionScale(%d) = (%d,%d,%v)", index, precision, scale, ok)
				}
				if got := tc.rs.ColumnTypeScanType(index); got != nil {
					t.Errorf("ScanType(%d) = %v", index, got)
				}
			}
		})
	}
}

func TestColumnDescriptorDatabaseTypeName(t *testing.T) {
	cases := []struct {
		name        string
		dataType    uint16
		charsetForm int
		typeName    string
		isJSON      bool
		want        string
	}{
		{"CHAR", types.CHAR, 1, "", false, "CHAR"},
		{"NCHAR", types.CHAR, 2, "", false, "NCHAR"},
		{"VARCHAR2", types.NCHAR, 1, "", false, "VARCHAR2"},
		{"NVARCHAR2_NCHAR_wire", types.NCHAR, 2, "", false, "NVARCHAR2"},
		{"VARCHAR2_VARCHAR_wire", types.VARCHAR, 1, "", false, "VARCHAR2"},
		{"NVARCHAR2_VARCHAR_wire", types.VARCHAR, 2, "", false, "NVARCHAR2"},
		{"LONG", types.LONG, 1, "", false, "LONG"},
		{"LONG_define", types.LongVarChar, 1, "", false, "LONG"},
		{"NUMBER", types.NUMBER, 0, "", false, "NUMBER"},
		{"FLOAT", types.FLOAT, 0, "", false, "FLOAT"},
		{"BINARY_FLOAT", types.IBFLOAT, 0, "", false, "BINARY_FLOAT"},
		{"BINARY_DOUBLE", types.IBDOUBLE, 0, "", false, "BINARY_DOUBLE"},
		{"DATE", types.DATE, 0, "", false, "DATE"},
		{"TIMESTAMP", types.TIMESTAMP, 0, "", false, "TIMESTAMP"},
		{"TIMESTAMP_TZ", types.TIMESTAMPTZ, 0, "", false, "TIMESTAMP WITH TIME ZONE"},
		{"TIMESTAMP_LTZ", types.TimeStampLTZ, 0, "", false, "TIMESTAMP WITH LOCAL TIME ZONE"},
		{"INTERVAL_YM", types.IntervalYM, 0, "", false, "INTERVAL YEAR TO MONTH"},
		{"INTERVAL_DS", types.IntervalDS, 0, "", false, "INTERVAL DAY TO SECOND"},
		{"RAW", types.RAW, 0, "", false, "RAW"},
		{"LONG_RAW", types.LongRaw, 0, "", false, "LONG RAW"},
		{"ROWID", types.ROWID, 0, "", false, "ROWID"},
		{"UROWID", types.UROWID, 0, "", false, "UROWID"},
		{"CLOB", types.OCIClobLocator, 1, "", false, "CLOB"},
		{"NCLOB", types.OCIClobLocator, 2, "", false, "NCLOB"},
		{"BLOB", types.OCIBlobLocator, 0, "", false, "BLOB"},
		{"BFILE", types.OCIFileLocator, 0, "", false, "BFILE"},
		{"JSON", types.JSON, 0, "", false, "JSON"},
		{"JSON_semantic_CLOB", types.OCIClobLocator, 1, "", true, "JSON"},
		{"VECTOR", types.VECTOR, 0, "", false, "VECTOR"},
		{"XMLTYPE", types.XMLType, 0, "XMLTYPE", false, "XMLTYPE"},
		{"UDT", types.XMLType, 0, "HR.CUSTOM_TYPE", false, "HR.CUSTOM_TYPE"},
		{"BOOLEAN", types.BOOLEAN, 0, "", false, "BOOLEAN"},
		{"REF_CURSOR", types.REFCURSOR, 0, "", false, "REF CURSOR"},
		{"unknown", 0xEEEE, 0, "", false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			col := metadataColumn(tc.dataType)
			col.CharsetForm = tc.charsetForm
			col.TypeName = tc.typeName
			col.IsJson = tc.isJSON
			got := metadataResultSet(col).ColumnTypeDatabaseTypeName(0)
			if got != tc.want {
				t.Fatalf("DatabaseTypeName() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestColumnDescriptorLength(t *testing.T) {
	cases := []struct {
		name       string
		col        ParameterInfo
		wantLength int64
		wantOK     bool
	}{
		{name: "VARCHAR2_chars", col: func() ParameterInfo {
			col := metadataColumn(types.NCHAR)
			col.CharsetForm = 1
			col.MaxCharLen = 32
			col.MaxLen = 64
			return col
		}(), wantLength: 32, wantOK: true},
		{name: "NVARCHAR2_chars", col: func() ParameterInfo {
			col := metadataColumn(types.NCHAR)
			col.CharsetForm = 2
			col.MaxCharLen = 16
			col.MaxLen = 64
			return col
		}(), wantLength: 16, wantOK: true},
		{name: "VARCHAR2_byte_fallback", col: func() ParameterInfo {
			col := metadataColumn(types.NCHAR)
			col.MaxLen = 20
			return col
		}(), wantLength: 20, wantOK: true},
		{name: "CHAR_fixed", col: func() ParameterInfo {
			col := metadataColumn(types.CHAR)
			col.MaxCharLen = 10
			col.MaxLen = 30
			return col
		}(), wantLength: 0, wantOK: false},
		{name: "RAW", col: func() ParameterInfo {
			col := metadataColumn(types.RAW)
			col.MaxLen = 16
			return col
		}(), wantLength: 16, wantOK: true},
		{name: "LONG", col: metadataColumn(types.LONG), wantLength: math.MaxInt64, wantOK: true},
		{name: "LONG_RAW", col: metadataColumn(types.LongRaw), wantLength: math.MaxInt64, wantOK: true},
		{name: "CLOB", col: metadataColumn(types.OCIClobLocator), wantLength: math.MaxInt64, wantOK: true},
		{name: "BLOB", col: metadataColumn(types.OCIBlobLocator), wantLength: math.MaxInt64, wantOK: true},
		{name: "BFILE", col: metadataColumn(types.OCIFileLocator), wantLength: math.MaxInt64, wantOK: true},
		{name: "JSON", col: metadataColumn(types.JSON), wantLength: math.MaxInt64, wantOK: true},
		{name: "XMLTYPE", col: metadataColumn(types.XMLType), wantLength: math.MaxInt64, wantOK: true},
		{name: "VECTOR", col: metadataColumn(types.VECTOR), wantLength: math.MaxInt64, wantOK: true},
		{name: "NUMBER", col: metadataColumn(types.NUMBER), wantLength: 0, wantOK: false},
		{name: "DATE", col: metadataColumn(types.DATE), wantLength: 0, wantOK: false},
		{name: "UDT", col: func() ParameterInfo {
			col := metadataColumn(types.XMLType)
			col.TypeName = "CUSTOM_TYPE"
			col.MaxLen = 10
			return col
		}(), wantLength: 0, wantOK: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			length, ok := metadataResultSet(tc.col).ColumnTypeLength(0)
			if length != tc.wantLength || ok != tc.wantOK {
				t.Fatalf("Length() = (%d,%v), want (%d,%v)", length, ok, tc.wantLength, tc.wantOK)
			}
		})
	}
}

func TestColumnDescriptorPrecisionScale(t *testing.T) {
	cases := []struct {
		name          string
		dataType      uint16
		rawPrecision  int64
		rawScale      int64
		known         bool
		unconstrained bool
		wantPrecision int64
		wantScale     int64
		wantOK        bool
	}{
		{name: "negative_scale", dataType: types.NUMBER, rawPrecision: 10, rawScale: -2, known: true, wantPrecision: 10, wantScale: -2, wantOK: true},
		{name: "zero_scale", dataType: types.NUMBER, rawPrecision: 5, rawScale: 0, known: true, wantPrecision: 5, wantScale: 0, wantOK: true},
		{name: "unconstrained", dataType: types.NUMBER, rawPrecision: 0, rawScale: -127, known: true, unconstrained: true, wantPrecision: math.MaxInt64, wantScale: math.MaxInt64, wantOK: true},
		{name: "precision_zero_not_unconstrained", dataType: types.NUMBER, rawPrecision: 0, rawScale: 0, known: true, wantPrecision: 0, wantScale: 0, wantOK: true},
		{name: "unknown", dataType: types.NUMBER, rawPrecision: 10, rawScale: 2, known: false, wantOK: false},
		{name: "non_number", dataType: types.NCHAR, rawPrecision: 10, rawScale: 2, known: true, wantOK: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			col := metadataColumn(tc.dataType)
			col.rawPrecision = tc.rawPrecision
			col.rawScale = tc.rawScale
			col.precisionScaleKnown = tc.known
			col.unconstrainedNumber = tc.unconstrained
			precision, scale, ok := metadataResultSet(col).ColumnTypePrecisionScale(0)
			if precision != tc.wantPrecision || scale != tc.wantScale || ok != tc.wantOK {
				t.Fatalf("PrecisionScale() = (%d,%d,%v), want (%d,%d,%v)", precision, scale, ok, tc.wantPrecision, tc.wantScale, tc.wantOK)
			}
		})
	}
}

func TestColumnDescriptorNullable(t *testing.T) {
	knownFalse := metadataColumn(types.NCHAR)
	knownFalse.AllowNull = false
	knownTrue := metadataColumn(types.NCHAR)
	knownTrue.AllowNull = true
	unknown := metadataColumn(types.NCHAR)
	unknown.nullabilityKnown = false

	cases := []struct {
		name         string
		col          ParameterInfo
		wantNullable bool
		wantOK       bool
	}{
		{"not_null", knownFalse, false, true},
		{"nullable", knownTrue, true, true},
		{"unknown", unknown, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nullable, ok := metadataResultSet(tc.col).ColumnTypeNullable(0)
			if nullable != tc.wantNullable || ok != tc.wantOK {
				t.Fatalf("Nullable() = (%v,%v), want (%v,%v)", nullable, ok, tc.wantNullable, tc.wantOK)
			}
		})
	}
}

func TestColumnDescriptorScanType(t *testing.T) {
	jsonType := reflect.TypeOf((*types.Json)(nil))
	clobType := reflect.TypeOf((*types.Clob)(nil))
	blobType := reflect.TypeOf((*types.Blob)(nil))
	bfileType := reflect.TypeOf((*types.BFile)(nil))
	vectorType := reflect.TypeOf((*types.Vector)(nil))
	registeredUDT := reflect.TypeOf(struct{ Value string }{})

	cases := []struct {
		name string
		col  ParameterInfo
		want reflect.Type
	}{
		{"string", metadataColumn(types.NCHAR), types.TyString},
		{"fixed_string", metadataColumn(types.CHAR), types.TyString},
		{"number", metadataColumn(types.NUMBER), types.TyString},
		{"binary_float", metadataColumn(types.IBFLOAT), types.TyFloat32},
		{"binary_double", metadataColumn(types.IBDOUBLE), types.TyFloat64},
		{"date", metadataColumn(types.DATE), types.TyTime},
		{"timestamp", metadataColumn(types.TIMESTAMP), types.TyTime},
		{"timestamp_ltz_alias", metadataColumn(types.TimeStampLTZ), types.TyTime},
		{"timestamp_ltz_dty", metadataColumn(types.TimeStampLTZ_DTY), types.TyTime},
		{"interval_ym", metadataColumn(types.INTERVALYM_DTY), types.TyTime},
		{"interval_ds", metadataColumn(types.INTERVALDS_DTY), types.TyTime},
		{"alternate_interval_ym", metadataColumn(types.IntervalYM), nil},
		{"alternate_interval_ds", metadataColumn(types.IntervalDS), nil},
		{"raw", metadataColumn(types.RAW), types.TyBytes},
		{"varchar_alias", metadataColumn(types.VARCHAR), nil},
		{"var_raw_alias", metadataColumn(types.VarRaw), nil},
		{"binary_float_alias", metadataColumn(types.BFLOAT), nil},
		{"binary_double_alias", metadataColumn(types.BDOUBLE), nil},
		{"charz_alias", metadataColumn(types.CHARZ), nil},
		{"long", metadataColumn(types.LONG), types.TyString},
		{"long_raw", metadataColumn(types.LongRaw), types.TyBytes},
		{"long_var_raw_alias", metadataColumn(types.LongVarRaw), nil},
		{"rowid", metadataColumn(types.ROWID), types.TyString},
		{"boolean", metadataColumn(types.BOOLEAN), types.TyBool},
		{"json", metadataColumn(types.JSON), jsonType},
		{"bfile", metadataColumn(types.OCIFileLocator), bfileType},
		{"ref_cursor", metadataColumn(types.REFCURSOR), nil},
		{"unknown", metadataColumn(0xEEEE), nil},
		{"clob_inline", func() ParameterInfo {
			col := metadataColumn(types.OCIClobLocator)
			col.DataType = types.LongVarChar
			col.lobFetch = configurations.INLINE
			return col
		}(), types.TyString},
		{"clob_auto", func() ParameterInfo {
			col := metadataColumn(types.OCIClobLocator)
			col.lobReadMode = configurations.LobReadMode_AUTO
			return col
		}(), types.TyString},
		{"clob_explicit", metadataColumn(types.OCIClobLocator), clobType},
		{"blob_inline", func() ParameterInfo {
			col := metadataColumn(types.OCIBlobLocator)
			col.DataType = types.LongRaw
			col.lobFetch = configurations.INLINE
			return col
		}(), types.TyBytes},
		{"blob_auto", func() ParameterInfo {
			col := metadataColumn(types.OCIBlobLocator)
			col.lobReadMode = configurations.LobReadMode_AUTO
			return col
		}(), types.TyBytes},
		{"blob_explicit", metadataColumn(types.OCIBlobLocator), blobType},
		{"json_semantic_clob_auto", func() ParameterInfo {
			col := metadataColumn(types.OCIClobLocator)
			col.IsJson = true
			col.lobReadMode = configurations.LobReadMode_AUTO
			return col
		}(), types.TyString},
		{"json_inline", func() ParameterInfo {
			col := metadataColumn(types.JSON)
			col.DataType = types.LongRaw
			col.lobFetch = configurations.INLINE
			return col
		}(), types.TyBytes},
		{"vector_inline", func() ParameterInfo {
			col := metadataColumn(types.VECTOR)
			col.DataType = types.LongRaw
			col.lobFetch = configurations.INLINE
			return col
		}(), types.TyBytes},
		{"vector_float32_auto", func() ParameterInfo {
			col := metadataColumn(types.VECTOR)
			col.VectorFormat = 2
			col.lobReadMode = configurations.LobReadMode_AUTO
			return col
		}(), reflect.TypeOf((*[]float32)(nil)).Elem()},
		{"vector_float64_auto", func() ParameterInfo {
			col := metadataColumn(types.VECTOR)
			col.VectorFormat = 3
			col.lobReadMode = configurations.LobReadMode_AUTO
			return col
		}(), reflect.TypeOf((*[]float64)(nil)).Elem()},
		{"vector_bytes_auto", func() ParameterInfo {
			col := metadataColumn(types.VECTOR)
			col.VectorFormat = 4
			col.lobReadMode = configurations.LobReadMode_AUTO
			return col
		}(), types.TyBytes},
		{"vector_unknown_auto", func() ParameterInfo {
			col := metadataColumn(types.VECTOR)
			col.VectorFormat = 9
			col.lobReadMode = configurations.LobReadMode_AUTO
			return col
		}(), nil},
		{"vector_explicit", func() ParameterInfo {
			col := metadataColumn(types.VECTOR)
			col.VectorFormat = 2
			return col
		}(), vectorType},
		{"registered_udt", func() ParameterInfo {
			col := metadataColumn(types.XMLType)
			col.TypeName = "CUSTOM_TYPE"
			col.udtScanType = registeredUDT
			return col
		}(), registeredUDT},
		{"unregistered_udt", func() ParameterInfo {
			col := metadataColumn(types.XMLType)
			col.TypeName = "CUSTOM_TYPE"
			return col
		}(), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := metadataResultSet(tc.col).ColumnTypeScanType(0)
			if got != tc.want {
				t.Fatalf("ScanType() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestColumnDescriptorNilParameter(t *testing.T) {
	desc := newColumnDescriptor(nil)
	if desc.databaseTypeName != "" || desc.scanType != nil || desc.lengthKnown ||
		desc.precisionScaleOK || desc.nullabilityKnown {
		t.Fatalf("nil ParameterInfo produced a non-zero descriptor: %+v", desc)
	}
}

// TestColumnDescriptorSnapshotsConnectionOptions verifies that the descriptor
// uses the LOB options staged on the column, not the live connection config.
func TestColumnDescriptorSnapshotsConnectionOptions(t *testing.T) {
	cfg := &configurations.ConnectionConfig{
		Lob:         configurations.STREAM,
		LobReadMode: configurations.LobReadMode_AUTO,
	}
	conn := &Connection{connOption: cfg}
	col := metadataColumn(types.OCIClobLocator)
	col.captureColumnMetadata(conn)

	cfg.Lob = configurations.INLINE
	cfg.LobReadMode = configurations.LobReadMode_NONE

	rs := metadataResultSet(col)
	if got := rs.ColumnTypeScanType(0); got != types.TyString {
		t.Fatalf("ScanType() reread mutable connection options: %v", got)
	}
}

// TestWriteDefinePreservesOriginalLobMetadata covers the inline RefCursor
// ordering: staging was captured by load(), writeDefine() mutates DataType
// before the descriptor is built, and the descriptor still reports the
// original semantic type.
func TestWriteDefinePreservesOriginalLobMetadata(t *testing.T) {
	cases := []struct {
		name         string
		dataType     uint16
		charsetForm  int
		isJSON       bool
		fetch        configurations.LobFetch
		readMode     configurations.LobReadMode
		wantWireType uint16
		wantName     string
		wantScanType reflect.Type
	}{
		{"inline_clob", types.OCIClobLocator, 1, false, configurations.INLINE, configurations.LobReadMode_AUTO, types.LongVarChar, "CLOB", types.TyString},
		{"inline_nclob", types.OCIClobLocator, 2, false, configurations.INLINE, configurations.LobReadMode_AUTO, types.LongVarChar, "NCLOB", types.TyString},
		{"inline_blob", types.OCIBlobLocator, 0, false, configurations.INLINE, configurations.LobReadMode_AUTO, types.LongRaw, "BLOB", types.TyBytes},
		{"inline_vector", types.VECTOR, 0, false, configurations.INLINE, configurations.LobReadMode_AUTO, types.LongRaw, "VECTOR", types.TyBytes},
		{"inline_json_wire", types.JSON, 0, false, configurations.INLINE, configurations.LobReadMode_AUTO, types.LongRaw, "JSON", types.TyBytes},
		{"json_semantic_stays_locator", types.JSON, 0, true, configurations.INLINE, configurations.LobReadMode_AUTO, types.JSON, "JSON", reflect.TypeOf((*types.Json)(nil))},
		{"stream_blob_auto", types.OCIBlobLocator, 0, false, configurations.STREAM, configurations.LobReadMode_AUTO, types.OCIBlobLocator, "BLOB", types.TyBytes},
		{"stream_blob_explicit", types.OCIBlobLocator, 0, false, configurations.STREAM, configurations.LobReadMode_NONE, types.OCIBlobLocator, "BLOB", reflect.TypeOf((*types.Blob)(nil))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &configurations.ConnectionConfig{
				SessionInfo: configurations.SessionInfo{
					SessionDataUnitSize:   0xFFFF,
					TransportDataUnitSize: 0xFFFF,
				},
				Lob:         tc.fetch,
				LobReadMode: tc.readMode,
			}
			conn := &Connection{
				connOption: cfg,
				session:    network.NewSession(cfg, trace.NilTracer()),
			}
			col := metadataColumn(tc.dataType)
			col.CharsetForm = tc.charsetForm
			col.IsJson = tc.isJSON
			stmt := &defaultStmt{connection: conn, columns: []ParameterInfo{col}}
			if err := stmt.writeDefine(); err != nil {
				t.Fatalf("writeDefine() failed: %v", err)
			}
			if got := stmt.columns[0].DataType; got != tc.wantWireType {
				t.Fatalf("wire DataType = %d, want %d", got, tc.wantWireType)
			}
			// RefCursor ordering: the descriptor is built after writeDefine()
			// mutated DataType and must still report the staged original type.
			rs := metadataResultSet(stmt.columns...)
			if got := rs.ColumnTypeDatabaseTypeName(0); got != tc.wantName {
				t.Fatalf("DatabaseTypeName() = %q, want %q", got, tc.wantName)
			}
			if got := rs.ColumnTypeScanType(0); got != tc.wantScanType {
				t.Fatalf("ScanType() = %v, want %v", got, tc.wantScanType)
			}
		})
	}
}

func descriptorOutputs(rs *ResultSet) (name string, length int64, lengthOK bool,
	precision int64, scale int64, psOK bool, nullable bool, nullOK bool, scan reflect.Type) {
	name = rs.ColumnTypeDatabaseTypeName(0)
	length, lengthOK = rs.ColumnTypeLength(0)
	precision, scale, psOK = rs.ColumnTypePrecisionScale(0)
	nullable, nullOK = rs.ColumnTypeNullable(0)
	scan = rs.ColumnTypeScanType(0)
	return
}

// TestDescriptorBuiltBeforeWriteDefine covers the ordinary statement
// ordering: load() captured staging, read() built the descriptor, and a later
// inline writeDefine() mutation plus rebuild attempt must leave it unchanged.
func TestDescriptorBuiltBeforeWriteDefine(t *testing.T) {
	cfg := &configurations.ConnectionConfig{
		SessionInfo: configurations.SessionInfo{
			SessionDataUnitSize:   0xFFFF,
			TransportDataUnitSize: 0xFFFF,
		},
		Lob:         configurations.INLINE,
		LobReadMode: configurations.LobReadMode_AUTO,
	}
	conn := &Connection{
		connOption: cfg,
		session:    network.NewSession(cfg, trace.NilTracer()),
	}
	// stage the column exactly as load() would have captured it against this
	// connection (LOB fetch INLINE, read mode AUTO)
	col := ParameterInfo{}
	col.DataType = types.OCIClobLocator
	col.CharsetForm = 1
	col.AllowNull = true
	stageColumnMetadata(&col, configurations.INLINE, configurations.LobReadMode_AUTO)
	stmt := &defaultStmt{connection: conn, columns: []ParameterInfo{col}}

	rs := &ResultSet{cols: &stmt.columns}
	rs.buildColumnDescriptors()
	beforeName, beforeLen, beforeLenOK, beforeP, beforeS, beforePSOK,
		beforeNull, beforeNullOK, beforeScan := descriptorOutputs(rs)

	if err := stmt.writeDefine(); err != nil {
		t.Fatalf("writeDefine() failed: %v", err)
	}
	if stmt.columns[0].DataType != types.LongVarChar {
		t.Fatalf("writeDefine() did not mutate DataType: %d", stmt.columns[0].DataType)
	}
	cfg.Lob = configurations.STREAM
	cfg.LobReadMode = configurations.LobReadMode_NONE

	// repeated read()/fetch() cycles must not rebuild the snapshot
	rs.buildColumnDescriptors()
	name, length, lengthOK, precision, scale, psOK, nullable, nullOK, scan := descriptorOutputs(rs)
	if name != beforeName || length != beforeLen || lengthOK != beforeLenOK ||
		precision != beforeP || scale != beforeS || psOK != beforePSOK ||
		nullable != beforeNull || nullOK != beforeNullOK || scan != beforeScan {
		t.Fatal("descriptor changed after writeDefine()/config mutation")
	}
	if name != "CLOB" || scan != types.TyString {
		t.Fatalf("descriptor = (%q,%v), want (CLOB,string)", name, scan)
	}
}

// TestDescriptorFromStagedSyntheticColumn covers the synthetic equivalent of
// the production orderings: a hand-built ParameterInfo staged via
// stageColumnMetadata, mutated by writeDefine(), then snapshotted.
func TestDescriptorFromStagedSyntheticColumn(t *testing.T) {
	cfg := &configurations.ConnectionConfig{
		SessionInfo: configurations.SessionInfo{
			SessionDataUnitSize:   0xFFFF,
			TransportDataUnitSize: 0xFFFF,
		},
		Lob:         configurations.INLINE,
		LobReadMode: configurations.LobReadMode_AUTO,
	}
	conn := &Connection{
		connOption: cfg,
		session:    network.NewSession(cfg, trace.NilTracer()),
	}
	col := ParameterInfo{}
	col.DataType = types.OCIClobLocator
	col.CharsetForm = 1
	stageColumnMetadata(&col, configurations.INLINE, configurations.LobReadMode_AUTO)
	stmt := &defaultStmt{connection: conn, columns: []ParameterInfo{col}}
	if err := stmt.writeDefine(); err != nil {
		t.Fatalf("writeDefine() failed: %v", err)
	}
	if stmt.columns[0].DataType != types.LongVarChar {
		t.Fatalf("wire DataType = %d, want %d", stmt.columns[0].DataType, types.LongVarChar)
	}
	rs := metadataResultSet(stmt.columns...)
	if got := rs.ColumnTypeDatabaseTypeName(0); got != "CLOB" {
		t.Fatalf("DatabaseTypeName() = %q, want CLOB", got)
	}
	if got := rs.ColumnTypeScanType(0); got != types.TyString {
		t.Fatalf("ScanType() = %v, want string", got)
	}
}
