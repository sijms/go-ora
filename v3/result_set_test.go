package go_ora

import (
	"bytes"
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/parameter_coder"
	"github.com/sijms/go-ora/v3/trace"
	"github.com/sijms/go-ora/v3/types"
)

func metadataColumn(dataType uint16) ParameterInfo {
	col := ParameterInfo{
		nullabilityKnown:      true,
		originalDataType:      dataType,
		originalDataTypeKnown: true,
		lobFetch:              configurations.STREAM,
		lobFetchKnown:         true,
		lobReadMode:           configurations.LobReadMode_NONE,
		lobReadModeKnown:      true,
	}
	col.DataType = dataType
	return col
}

func metadataResultSet(cols ...ParameterInfo) *ResultSet {
	return &ResultSet{cols: &cols}
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

func TestResultSetColumnMetadataInvalidState(t *testing.T) {
	cases := []struct {
		name string
		rs   *ResultSet
	}{
		{"nil_result_set", nil},
		{"nil_columns", &ResultSet{}},
		{"empty_columns", metadataResultSet()},
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

func TestColumnTypeDatabaseTypeName(t *testing.T) {
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

func TestColumnTypeLength(t *testing.T) {
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

func TestColumnTypePrecisionScale(t *testing.T) {
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

func TestColumnTypeNullable(t *testing.T) {
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

func TestColumnTypeScanType(t *testing.T) {
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

func TestColumnMetadataSnapshotsConnectionOptions(t *testing.T) {
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
				SessionDataUnitSize:   0xFFFF,
				TransportDataUnitSize: 0xFFFF,
				Lob:                   tc.fetch,
				LobReadMode:           tc.readMode,
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

type metadataObject struct {
	Value string
}

func TestParameterLoadPreservesMetadata(t *testing.T) {
	conn := testParameterConnection(configurations.STREAM, configurations.LobReadMode_AUTO)
	col := loadParameterInfo(t, conn, parameterDescriptorSpec{
		dataType:  types.NUMBER,
		precision: 0,
		scale:     -127,
		nullable:  true,
		name:      "AMOUNT",
	})
	if col.Precision != 38 || col.Scale != 0xFF {
		t.Fatalf("public NUMBER fields = (%d,%d), want (38,255)", col.Precision, col.Scale)
	}
	rs := metadataResultSet(*col)
	precision, scale, ok := rs.ColumnTypePrecisionScale(0)
	if precision != math.MaxInt64 || scale != math.MaxInt64 || !ok {
		t.Fatalf("unconstrained PrecisionScale() = (%d,%d,%v)", precision, scale, ok)
	}
	nullable, known := rs.ColumnTypeNullable(0)
	if !nullable || !known {
		t.Fatalf("Nullable() = (%v,%v)", nullable, known)
	}
}

func TestParameterLoadPreservesNegativeScaleAndUDT(t *testing.T) {
	udtType := reflect.TypeOf(metadataObject{})
	conn := testParameterConnection(configurations.INLINE, configurations.LobReadMode_NONE)
	conn.nameTypeCoder["CUSTOM_TYPE"] = &ObjectParameter{typ: udtType}
	col := loadParameterInfo(t, conn, parameterDescriptorSpec{
		dataType:  types.XMLType,
		precision: 0,
		scale:     -2,
		typeName:  "custom_type",
		name:      "OBJ",
	})
	rs := metadataResultSet(*col)
	if got := rs.ColumnTypeDatabaseTypeName(0); got != "CUSTOM_TYPE" {
		t.Fatalf("DatabaseTypeName() = %q", got)
	}
	if got := rs.ColumnTypeScanType(0); got != udtType {
		t.Fatalf("ScanType() = %v, want %v", got, udtType)
	}
	if !col.lobFetchKnown || col.lobFetch != configurations.INLINE {
		t.Fatal("LOB fetch mode was not captured")
	}
}

func TestParameterLoadCollectionUDTScanType(t *testing.T) {
	registeredType := reflect.TypeOf([]metadataObject{})
	conn := testParameterConnection(configurations.STREAM, configurations.LobReadMode_AUTO)
	objectCoder := &ObjectParameter{isArray: true, typ: registeredType}
	conn.nameTypeCoder["CUSTOM_COLLECTION"] = objectCoder
	col := loadParameterInfo(t, conn, parameterDescriptorSpec{
		dataType: types.XMLType,
		typeName: "custom_collection",
		name:     "ITEMS",
	})
	rs := metadataResultSet(*col)
	if got := rs.ColumnTypeDatabaseTypeName(0); got != "CUSTOM_COLLECTION" {
		t.Fatalf("DatabaseTypeName() = %q", got)
	}
	if got, want := rs.ColumnTypeScanType(0), reflect.TypeOf([]interface{}{}); got != want {
		t.Fatalf("collection UDT ScanType() = %v, want %v", got, want)
	}

	writer := network.NewMemorySession(nil, nil, conn.GetSession().GetProperties())
	writer.PutBytes(0x88)
	writer.PutInt(0, 4, true, true)
	writer.PutInt(0, 2, true, true)
	writer.PutInt(0, 2, true, false)
	objectCoder.BValue = writer.GetWriteBuffer()
	decoded, err := objectCoder.Decode(conn)
	if err != nil {
		t.Fatalf("collection UDT Decode() failed: %v", err)
	}
	if got, want := reflect.TypeOf(decoded), reflect.TypeOf([]interface{}{}); got != want {
		t.Fatalf("collection UDT Decode() type = %v, want %v", got, want)
	}
}

func TestParameterLoadReadsLateJSONFlag(t *testing.T) {
	conn := testParameterConnection(configurations.INLINE, configurations.LobReadMode_AUTO)
	col := loadParameterInfo(t, conn, parameterDescriptorSpec{
		ttcVersion:  6,
		dataType:    types.OCIClobLocator,
		charsetForm: 1,
		udsFlags:    0x500,
		name:        "JSON_COL",
	})
	if !col.IsJson {
		t.Fatal("IsJson was not parsed")
	}
	rs := metadataResultSet(*col)
	if got := rs.ColumnTypeDatabaseTypeName(0); got != "JSON" {
		t.Fatalf("DatabaseTypeName() = %q, want JSON", got)
	}
	if got := rs.ColumnTypeScanType(0); got != types.TyString {
		t.Fatalf("ScanType() = %v, want string for automatic CLOB JSON", got)
	}
}

func TestParameterLoadPreservesConstrainedNegativeScale(t *testing.T) {
	conn := testParameterConnection(configurations.STREAM, configurations.LobReadMode_AUTO)
	col := loadParameterInfo(t, conn, parameterDescriptorSpec{
		dataType:  types.NUMBER,
		precision: 10,
		scale:     -2,
		name:      "NEGATIVE_SCALE",
	})
	precision, scale, ok := metadataResultSet(*col).ColumnTypePrecisionScale(0)
	if precision != 10 || scale != -2 || !ok {
		t.Fatalf("PrecisionScale() = (%d,%d,%v), want (10,-2,true)", precision, scale, ok)
	}
}

func TestParameterLoadPreservesRawZeroPrecisionScale(t *testing.T) {
	conn := testParameterConnection(configurations.STREAM, configurations.LobReadMode_AUTO)
	col := loadParameterInfo(t, conn, parameterDescriptorSpec{
		dataType:  types.NUMBER,
		precision: 0,
		scale:     0,
		name:      "ZERO_PRECISION",
	})
	precision, scale, ok := metadataResultSet(*col).ColumnTypePrecisionScale(0)
	if precision != 0 || scale != 0 || !ok {
		t.Fatalf("PrecisionScale() = (%d,%d,%v), want raw (0,0,true)", precision, scale, ok)
	}
}

func TestParameterLoadPrecisionReadError(t *testing.T) {
	conn := testParameterConnection(configurations.STREAM, configurations.LobReadMode_AUTO)
	conn.session.SaveState(&network.SessionState{
		InBuffer:  bytes.NewBuffer([]byte{byte(types.NUMBER), 3}),
		OutBuffer: &bytes.Buffer{},
	})
	col := ParameterInfo{}
	if err := col.load(conn); err == nil {
		t.Fatal("expected precision read error")
	}
}

func TestParameterLoadLateDescriptorErrors(t *testing.T) {
	cases := []struct {
		name string
		spec parameterDescriptorSpec
	}{
		{name: "post_type_name", spec: parameterDescriptorSpec{ttcVersion: 3, dataType: types.NCHAR, truncate: 1}},
		{name: "json_flags", spec: parameterDescriptorSpec{ttcVersion: 6, dataType: types.OCIClobLocator, truncate: 1}},
		{name: "vector_flag", spec: parameterDescriptorSpec{ttcVersion: 24, dataType: types.VECTOR, truncate: 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			conn := testParameterConnection(configurations.STREAM, configurations.LobReadMode_AUTO)
			if err := loadParameterInfoError(conn, tc.spec); err == nil {
				t.Fatal("expected truncated descriptor error")
			}
		})
	}
}

func metadataScalarBValue(t *testing.T, conn *Connection, dataType uint16, input interface{}) []byte {
	t.Helper()
	var err error
	switch dataType {
	case types.NCHAR, types.CHAR, types.LONG, types.LongVarChar:
		encoder := &types.String{Conv: conn.sStrConv}
		encoder.SetDataType(dataType)
		err = encoder.SetValue(input)
		if err == nil {
			return encoder.Bytes()
		}
	case types.NUMBER, types.IBFLOAT, types.IBDOUBLE:
		encoder := &types.Number{}
		encoder.SetDataType(dataType)
		err = encoder.SetValue(input)
		if err == nil {
			return encoder.Bytes()
		}
	case types.DATE, types.TIMESTAMP, types.TIMESTAMPTZ, types.TimeStampDTY, types.TimeStampTZ_DTY, types.TimeStampLTZ_DTY, types.TimeStampLTZ:
		encoder := &types.Date{
			DBTimeZone:       conn.GetDBTimeZone(),
			DBServerTimeZone: conn.GetDBServerTimeZone(),
		}
		encodeType := dataType
		if dataType == types.TimeStampDTY {
			encodeType = types.TIMESTAMP
		}
		encoder.SetDataType(encodeType)
		err = encoder.SetValue(input)
		if err == nil {
			return encoder.Bytes()
		}
	case types.INTERVALYM_DTY, types.INTERVALDS_DTY:
		encoder := &types.Interval{}
		encoder.SetDataType(dataType)
		err = encoder.SetValue(input)
		if err == nil {
			return encoder.Bytes()
		}
	case types.RAW, types.LongRaw:
		encoder := &types.Raw{}
		encoder.SetDataType(dataType)
		err = encoder.SetValue(input)
		if err == nil {
			return encoder.Bytes()
		}
	case types.BOOLEAN:
		encoder := &types.Bool{}
		encoder.SetDataType(dataType)
		err = encoder.SetValue(input)
		if err == nil {
			return encoder.Bytes()
		}
	default:
		t.Fatalf("no scalar encoder case for datatype %d", dataType)
	}
	if err != nil {
		t.Fatalf("scalar encode for datatype %d failed: %v", dataType, err)
	}
	return nil
}

func TestColumnTypeScanTypeMatchesCoderOutputs(t *testing.T) {
	drv := NewDriver()
	conn := testParameterConnection(configurations.STREAM, configurations.LobReadMode_AUTO)
	conn.oracleTypeCoder = drv.oracleTypeCoder

	timestamp := time.Date(2026, time.January, 2, 3, 4, 5, 123456000, time.UTC)
	cases := []struct {
		name     string
		dataType uint16
		input    interface{}
	}{
		{"varchar2", types.NCHAR, "abc"},
		{"char", types.CHAR, "x"},
		{"number", types.NUMBER, "12.34"},
		{"binary_float", types.IBFLOAT, float32(1.5)},
		{"binary_double", types.IBDOUBLE, float64(1.5)},
		{"date", types.DATE, timestamp},
		{"timestamp", types.TIMESTAMP, timestamp},
		{"timestamp_tz", types.TIMESTAMPTZ, timestamp},
		{"timestamp_dty", types.TimeStampDTY, timestamp},
		{"timestamp_tz_dty", types.TimeStampTZ_DTY, timestamp},
		{"timestamp_ltz_dty", types.TimeStampLTZ_DTY, timestamp},
		{"timestamp_ltz", types.TimeStampLTZ, timestamp},
		{"interval_ym", types.INTERVALYM_DTY, time.Date(2026, time.March, 0, 0, 0, 0, 0, time.UTC)},
		{"interval_ds", types.INTERVALDS_DTY, time.Date(0, 0, 2, 3, 4, 5, 6000000, time.UTC)},
		{"raw", types.RAW, []byte{1, 2, 3}},
		{"boolean", types.BOOLEAN, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			coder, err := conn.GetParameterCoder(tc.dataType)
			if err != nil {
				t.Fatalf("GetParameterCoder(%d) failed: %v", tc.dataType, err)
			}
			info := coder.GetParameterInfo()
			info.DataType = tc.dataType
			info.BValue = metadataScalarBValue(t, conn, tc.dataType, tc.input)
			coder.SetParameterInfo(info)
			actual, err := coder.Decode(conn)
			if err != nil {
				t.Fatalf("Decode() failed: %v", err)
			}
			if actual == nil {
				t.Fatal("Decode() returned nil for a non-NULL test value")
			}
			if got, want := metadataResultSet(metadataColumn(tc.dataType)).ColumnTypeScanType(0), reflect.TypeOf(actual); got != want {
				t.Fatalf("ScanType() = %v, decoded value %v", got, want)
			}
		})
	}
}

func TestNativeJSONAndRefCursorRuntimeTypes(t *testing.T) {
	drv := NewDriver()
	conn := testParameterConnection(configurations.STREAM, configurations.LobReadMode_AUTO)
	conn.oracleTypeCoder = drv.oracleTypeCoder

	jsonCoder, err := conn.GetParameterCoder(types.JSON)
	if err != nil {
		t.Fatalf("JSON coder unavailable: %v", err)
	}
	if err := jsonCoder.Encode(map[string]interface{}{"a": float64(1)}, conn); err != nil {
		t.Fatalf("JSON Encode() failed: %v", err)
	}
	jsonValue, err := jsonCoder.Decode(conn)
	if err != nil {
		t.Fatalf("JSON Decode() failed: %v", err)
	}
	if _, ok := jsonValue.(*types.Json); !ok {
		t.Fatalf("JSON decoded as %T, want *types.Json", jsonValue)
	}
	if got, want := metadataResultSet(metadataColumn(types.JSON)).ColumnTypeScanType(0), reflect.TypeOf(jsonValue); got != want {
		t.Fatalf("JSON ScanType() = %v, decoded value %v", got, want)
	}

	cursorCoder, err := conn.GetParameterCoder(types.REFCURSOR)
	if err != nil {
		t.Fatalf("REFCURSOR coder unavailable: %v", err)
	}
	cursorValue, err := cursorCoder.Decode(conn)
	if err != nil {
		t.Fatalf("REFCURSOR Decode() failed: %v", err)
	}
	if cursorValue != nil {
		t.Fatalf("REFCURSOR decoded as %#v, want current nil", cursorValue)
	}
	if got := metadataResultSet(metadataColumn(types.REFCURSOR)).ColumnTypeScanType(0); got != nil {
		t.Fatalf("REFCURSOR ScanType() = %v, want nil", got)
	}
}

type parameterDescriptorSpec struct {
	ttcVersion   uint8
	dataType     uint16
	flag         uint8
	precision    uint8
	scale        int
	maxLen       int64
	charsetForm  int
	maxCharLen   int64
	nullable     bool
	name         string
	schemaName   string
	typeName     string
	udsFlags     int
	vectorDim    int
	vectorFormat uint8
	vectorFlag   uint8
	truncate     int
}

func testParameterConnection(fetch configurations.LobFetch, read configurations.LobReadMode) *Connection {
	cfg := &configurations.ConnectionConfig{
		SessionDataUnitSize:   0xFFFF,
		TransportDataUnitSize: 0xFFFF,
		Lob:                   fetch,
		LobReadMode:           read,
	}
	session := network.NewSession(cfg, trace.NilTracer())
	conv := converters.NewStringConverter(873)
	session.StrConv = conv
	return &Connection{
		session:          session,
		connOption:       cfg,
		tcpNego:          &TCPNego{ServerCharset: 873, ServerNCharset: 2000},
		dataNego:         &DataTypeNego{},
		sStrConv:         conv,
		nStrConv:         conv,
		nameTypeCoder:    map[string]parameter_coder.OracleParameterCoder{},
		dbTimeZone:       time.UTC,
		dbServerTimeZone: time.UTC,
	}
}

func parameterDescriptorBytes(conn *Connection, spec parameterDescriptorSpec) []byte {
	if spec.ttcVersion == 0 {
		spec.ttcVersion = 2
	}
	writer := network.NewSession(conn.connOption, trace.NilTracer())
	writer.PutBytes(byte(spec.dataType), spec.flag, spec.precision)
	if numericDescriptor(spec.dataType) {
		if spec.scale < 0 {
			writer.PutBytes(0x81, byte(-spec.scale))
		} else {
			writer.PutInt(spec.scale, 2, true, true)
		}
	} else {
		writer.PutBytes(byte(spec.scale))
	}
	writer.PutInt(spec.maxLen, 4, true, true)
	writer.PutInt(0, 4, true, true)
	if spec.ttcVersion >= 10 {
		writer.PutInt(0, 8, true, true)
	} else {
		writer.PutInt(0, 4, true, true)
	}
	writer.PutDlc(nil)
	writer.PutInt(0, 2, true, true)
	writer.PutInt(0, 2, true, true)
	writer.PutInt(spec.charsetForm, 1, false, false)
	writer.PutInt(spec.maxCharLen, 4, true, true)
	if spec.ttcVersion >= 8 {
		writer.PutInt(0, 4, true, true)
	}
	if spec.nullable {
		writer.PutInt(1, 1, false, false)
	} else {
		writer.PutInt(0, 1, false, false)
	}
	writer.PutBytes(0)
	writer.PutDlc([]byte(spec.name))
	writer.PutDlc([]byte(spec.schemaName))
	writer.PutDlc([]byte(spec.typeName))
	if spec.ttcVersion >= 3 {
		writer.PutInt(0, 2, true, true)
	}
	if spec.ttcVersion >= 6 {
		writer.PutInt(spec.udsFlags, 4, true, true)
	}
	if spec.ttcVersion >= 17 {
		writer.PutDlc(nil)
		writer.PutDlc(nil)
	}
	if spec.ttcVersion >= 20 {
		writer.PutInt(0, 4, true, true)
	}
	if spec.ttcVersion >= 24 {
		if spec.dataType == types.VECTOR {
			writer.PutInt(spec.vectorDim, 4, true, true)
			writer.PutBytes(spec.vectorFormat, spec.vectorFlag)
		} else {
			writer.PutInt(0, 4, true, true)
			writer.PutInt(0, 1, true, true)
			writer.PutInt(0, 1, true, true)
		}
	}
	data := writer.GetWriteBuffer()
	if spec.truncate > 0 && spec.truncate < len(data) {
		data = data[:len(data)-spec.truncate]
	}
	return data
}

func loadParameterInfoError(conn *Connection, spec parameterDescriptorSpec) error {
	conn.session.TTCVersion = spec.ttcVersion
	conn.session.SaveState(&network.SessionState{
		InBuffer:  bytes.NewBuffer(parameterDescriptorBytes(conn, spec)),
		OutBuffer: &bytes.Buffer{},
	})
	return new(ParameterInfo).load(conn)
}

func loadParameterInfo(t *testing.T, conn *Connection, spec parameterDescriptorSpec) *ParameterInfo {
	t.Helper()
	data := parameterDescriptorBytes(conn, spec)
	conn.session.TTCVersion = spec.ttcVersion
	conn.session.SaveState(&network.SessionState{
		InBuffer:  bytes.NewBuffer(data),
		OutBuffer: &bytes.Buffer{},
	})
	col := new(ParameterInfo)
	if err := col.load(conn); err != nil {
		t.Fatalf("ParameterInfo.load() failed: %v", err)
	}
	return col
}

func numericDescriptor(dataType uint16) bool {
	switch dataType {
	case types.NUMBER, types.TimeStampDTY, types.TimeStampTZ_DTY, types.INTERVALDS_DTY,
		types.TIMESTAMP, types.TIMESTAMPTZ, types.IntervalDS, types.TimeStampLTZ_DTY,
		types.TimeStampLTZ:
		return true
	default:
		return false
	}
}

type metadataTestStringCoder struct {
	conv converters.IStringConverter
}

func (coder metadataTestStringCoder) GetStringCoder(int, int) (converters.IStringConverter, error) {
	return coder.conv, nil
}

func (coder metadataTestStringCoder) GetDefaultStringCoder() (converters.IStringConverter, error) {
	return coder.conv, nil
}

func (coder metadataTestStringCoder) GetServerStringCoder() converters.IStringConverter {
	return coder.conv
}

func (coder metadataTestStringCoder) GetServerNStringCoder() converters.IStringConverter {
	return coder.conv
}

func (coder metadataTestStringCoder) GetMaxStringLength() int64 { return 4000 }

type metadataTestLobStreamer struct {
	locator  types.Locator
	data     []byte
	fetch    configurations.LobFetch
	readMode configurations.LobReadMode
	conv     converters.IStringConverter
}

func (stream *metadataTestLobStreamer) StartContext(context.Context) chan struct{} {
	return make(chan struct{})
}
func (stream *metadataTestLobStreamer) EndContext(chan struct{})  {}
func (stream *metadataTestLobStreamer) GetLocator() types.Locator { return stream.locator }
func (stream *metadataTestLobStreamer) SetLocator(locator types.Locator) {
	stream.locator = locator
}
func (stream *metadataTestLobStreamer) DatabaseVersionNumber() int { return 2300 }
func (stream *metadataTestLobStreamer) GetStringCoder() converters.StringCoder {
	return metadataTestStringCoder{conv: stream.conv}
}
func (stream *metadataTestLobStreamer) GetLobStreamMode() configurations.LobFetch {
	return stream.fetch
}
func (stream *metadataTestLobStreamer) GetLobReadMode() configurations.LobReadMode {
	return stream.readMode
}
func (stream *metadataTestLobStreamer) GetTracer() trace.Tracer { return trace.NilTracer() }
func (stream *metadataTestLobStreamer) GetSize() (int64, error) {
	return int64(len(stream.data)), nil
}
func (stream *metadataTestLobStreamer) Exists() (bool, error) { return stream.locator != nil, nil }
func (stream *metadataTestLobStreamer) CreateTemporaryLocator(int, int) (types.Locator, error) {
	stream.locator = make(types.Locator, 40)
	return stream.locator, nil
}
func (stream *metadataTestLobStreamer) FreeTemporaryLocator() error { return nil }
func (stream *metadataTestLobStreamer) Open(int, int) error         { return nil }
func (stream *metadataTestLobStreamer) Read(int64, int64) ([]byte, error) {
	return stream.data, nil
}
func (stream *metadataTestLobStreamer) Write(data []byte) error {
	stream.data = append(stream.data, data...)
	return nil
}
func (stream *metadataTestLobStreamer) Close(int) error { return nil }

func decodeMetadataValue(t *testing.T, col ParameterInfo, value driver.Value) driver.Value {
	t.Helper()
	stmt := &defaultStmt{
		connection: &Connection{connOption: &configurations.ConnectionConfig{Lob: configurations.STREAM}},
		columns:    []ParameterInfo{col},
	}
	resultSet := &ResultSet{cols: &stmt.columns, rows: []Row{{value}}}
	if err := stmt.decodePrim(resultSet); err != nil {
		t.Fatalf("decodePrim() failed: %v", err)
	}
	return resultSet.rows[0][0]
}

func TestDecodePrimScanTypeConsistency(t *testing.T) {
	conv := converters.NewStringConverter(0x7D0)
	ordinaryLocator := types.Locator(make([]byte, 40))

	t.Run("clob_auto", func(t *testing.T) {
		col := metadataColumn(types.OCIClobLocator)
		col.lobReadMode = configurations.LobReadMode_AUTO
		clob := &types.Clob{}
		clob.Conv = conv
		clob.SetStreamer(&metadataTestLobStreamer{
			locator:  ordinaryLocator,
			data:     []byte("hello"),
			fetch:    configurations.STREAM,
			readMode: configurations.LobReadMode_AUTO,
			conv:     conv,
		})
		actual := decodeMetadataValue(t, col, clob)
		if got, want := metadataResultSet(col).ColumnTypeScanType(0), reflect.TypeOf(actual); got != want {
			t.Fatalf("ScanType() = %v, actual %v", got, want)
		}
	})

	t.Run("blob_auto", func(t *testing.T) {
		col := metadataColumn(types.OCIBlobLocator)
		col.lobReadMode = configurations.LobReadMode_AUTO
		blob := &types.Blob{}
		blob.SetStreamer(&metadataTestLobStreamer{
			locator:  ordinaryLocator,
			data:     []byte{1, 2, 3},
			fetch:    configurations.STREAM,
			readMode: configurations.LobReadMode_AUTO,
			conv:     conv,
		})
		actual := decodeMetadataValue(t, col, blob)
		if got, want := metadataResultSet(col).ColumnTypeScanType(0), reflect.TypeOf(actual); got != want {
			t.Fatalf("ScanType() = %v, actual %v", got, want)
		}
	})

	t.Run("blob_explicit_wrapper", func(t *testing.T) {
		col := metadataColumn(types.OCIBlobLocator)
		blob := &types.Blob{}
		blob.SetStreamer(&metadataTestLobStreamer{
			locator:  ordinaryLocator,
			data:     []byte{1, 2, 3},
			fetch:    configurations.STREAM,
			readMode: configurations.LobReadMode_NONE,
			conv:     conv,
		})
		actual := decodeMetadataValue(t, col, blob)
		if got, want := metadataResultSet(col).ColumnTypeScanType(0), reflect.TypeOf(actual); got != want {
			t.Fatalf("ScanType() = %v, actual %v", got, want)
		}
	})

	t.Run("quasi_locator_exception", func(t *testing.T) {
		col := metadataColumn(types.OCIClobLocator)
		col.lobReadMode = configurations.LobReadMode_AUTO
		clob := &types.Clob{}
		clob.Conv = conv
		clob.SetStreamer(&metadataTestLobStreamer{
			locator:  types.NewQuasiLocator(5),
			data:     []byte("hello"),
			fetch:    configurations.STREAM,
			readMode: configurations.LobReadMode_AUTO,
			conv:     conv,
		})
		actual := decodeMetadataValue(t, col, clob)
		if _, ok := actual.(*types.Clob); !ok {
			t.Fatalf("quasi locator decoded as %T, want *types.Clob", actual)
		}
		if got := metadataResultSet(col).ColumnTypeScanType(0); got != types.TyString {
			t.Fatalf("descriptor prediction = %v, want string", got)
		}
	})

	t.Run("null_lob_exception", func(t *testing.T) {
		col := metadataColumn(types.OCIBlobLocator)
		col.lobReadMode = configurations.LobReadMode_AUTO
		blob := &types.Blob{}
		blob.SetStreamer(&metadataTestLobStreamer{
			locator:  nil,
			fetch:    configurations.STREAM,
			readMode: configurations.LobReadMode_AUTO,
			conv:     conv,
		})
		actual := decodeMetadataValue(t, col, blob)
		if _, ok := actual.(*types.Blob); !ok {
			t.Fatalf("NULL locator decoded as %T, want *types.Blob", actual)
		}
		if got := metadataResultSet(col).ColumnTypeScanType(0); got != types.TyBytes {
			t.Fatalf("descriptor prediction = %v, want []byte", got)
		}
	})
}

func TestDecodePrimVectorFormats(t *testing.T) {
	cases := []struct {
		name   string
		format uint8
		value  interface{}
		want   interface{}
	}{
		{"float32", 2, []float32{1.5, -2.25}, []float32{1.5, -2.25}},
		{"float64", 3, []float64{1.5, -2.25}, []float64{1.5, -2.25}},
		{"bytes", 4, []byte{1, 2, 3}, []byte{1, 2, 3}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			encoded, err := types.CreateVector(tc.value)
			if err != nil {
				t.Fatalf("CreateVector() failed: %v", err)
			}
			col := metadataColumn(types.VECTOR)
			col.VectorFormat = tc.format
			col.lobReadMode = configurations.LobReadMode_AUTO
			vector := &types.Vector{}
			vector.SetStreamer(&metadataTestLobStreamer{
				locator:  types.Locator(make([]byte, 40)),
				data:     encoded.Bytes(),
				fetch:    configurations.STREAM,
				readMode: configurations.LobReadMode_AUTO,
			})
			actual := decodeMetadataValue(t, col, vector)
			if !reflect.DeepEqual(actual, tc.want) {
				t.Fatalf("decoded value = %#v, want %#v", actual, tc.want)
			}
			if got, want := metadataResultSet(col).ColumnTypeScanType(0), reflect.TypeOf(actual); got != want {
				t.Fatalf("ScanType() = %v, actual %v", got, want)
			}
		})
	}
}

type metadataTestStmt struct{}

func (stmt *metadataTestStmt) hasMoreRows() bool      { return false }
func (stmt *metadataTestStmt) noOfRowsToFetch() int   { return 1 }
func (stmt *metadataTestStmt) fetch(*ResultSet) error { return io.EOF }
func (stmt *metadataTestStmt) hasBLOB() bool          { return false }
func (stmt *metadataTestStmt) hasLONG() bool          { return false }
func (stmt *metadataTestStmt) read(*ResultSet) error  { return nil }
func (stmt *metadataTestStmt) Close() error           { return nil }
func (stmt *metadataTestStmt) CanAutoClose() bool     { return false }

type metadataTestDriver struct{}
type metadataTestConnector struct {
	rows driver.Rows
}
type metadataTestConn struct {
	rows driver.Rows
}

func (metadataTestDriver) Open(string) (driver.Conn, error) { return nil, driver.ErrSkip }
func (connector metadataTestConnector) Connect(context.Context) (driver.Conn, error) {
	return metadataTestConn{rows: connector.rows}, nil
}
func (connector metadataTestConnector) Driver() driver.Driver { return metadataTestDriver{} }
func (conn metadataTestConn) Prepare(string) (driver.Stmt, error) {
	return nil, driver.ErrSkip
}
func (conn metadataTestConn) Close() error { return nil }
func (conn metadataTestConn) Begin() (driver.Tx, error) {
	return nil, driver.ErrSkip
}
func (conn metadataTestConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return conn.rows, nil
}

func TestDatabaseSQLColumnTypesAndScan(t *testing.T) {
	varchar := metadataColumn(types.NCHAR)
	varchar.CharsetForm = 1
	varchar.MaxCharLen = 16
	varchar.Name = "TXT"
	raw := metadataColumn(types.RAW)
	raw.MaxLen = 4
	raw.Name = "RAW_COL"
	ts := metadataColumn(types.TIMESTAMP)
	ts.Name = "TS"
	boolean := metadataColumn(types.BOOLEAN)
	boolean.Name = "FLAG"
	number := metadataColumn(types.NUMBER)
	number.Name = "NUM"
	number.rawPrecision = 10
	number.rawScale = -2
	number.precisionScaleKnown = true
	double := metadataColumn(types.IBDOUBLE)
	double.Name = "DBL"

	cols := []ParameterInfo{varchar, raw, ts, boolean, number, double}
	rowTime := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	resultSet := ResultSet{
		cols:   &cols,
		rows:   []Row{{"abc", []byte{1, 2}, rowTime, true, "12.34", 7.5}},
		parent: &metadataTestStmt{},
	}
	db := sql.OpenDB(metadataTestConnector{rows: &DataSet{resultSets: []ResultSet{resultSet}}})
	defer db.Close()

	rows, err := db.QueryContext(context.Background(), "select")
	if err != nil {
		t.Fatalf("QueryContext() failed: %v", err)
	}
	defer rows.Close()
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		t.Fatalf("ColumnTypes() failed: %v", err)
	}
	wantNames := []string{"VARCHAR2", "RAW", "TIMESTAMP", "BOOLEAN", "NUMBER", "BINARY_DOUBLE"}
	for i, columnType := range columnTypes {
		if got := columnType.DatabaseTypeName(); got != wantNames[i] {
			t.Fatalf("column %d name = %q, want %q", i, got, wantNames[i])
		}
		if columnType.ScanType() == nil {
			t.Fatalf("column %d returned nil ScanType", i)
		}
	}
	if length, ok := columnTypes[0].Length(); length != 16 || !ok {
		t.Fatalf("VARCHAR2 length = (%d,%v)", length, ok)
	}
	if length, ok := columnTypes[1].Length(); length != 4 || !ok {
		t.Fatalf("RAW length = (%d,%v)", length, ok)
	}
	if precision, scale, ok := columnTypes[4].DecimalSize(); precision != 10 || scale != -2 || !ok {
		t.Fatalf("NUMBER decimal size = (%d,%d,%v)", precision, scale, ok)
	}
	if !rows.Next() {
		t.Fatalf("Rows.Next() failed: %v", rows.Err())
	}
	dest := make([]interface{}, len(columnTypes))
	for i, columnType := range columnTypes {
		dest[i] = reflect.New(columnType.ScanType()).Interface()
	}
	if err := rows.Scan(dest...); err != nil {
		t.Fatalf("Rows.Scan() failed: %v", err)
	}
	wantValues := []interface{}{"abc", []byte{1, 2}, rowTime, true, "12.34", 7.5}
	for i := range wantValues {
		actual := reflect.ValueOf(dest[i]).Elem().Interface()
		if !reflect.DeepEqual(actual, wantValues[i]) {
			t.Fatalf("column %d value = %#v, want %#v", i, actual, wantValues[i])
		}
	}
	if rows.Next() {
		t.Fatal("unexpected second row")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("Rows.Err() = %v", err)
	}
}

func TestDataSetMetadataResultSetSwitching(t *testing.T) {
	first := metadataColumn(types.NCHAR)
	first.Name = "FIRST"
	second := metadataColumn(types.RAW)
	second.Name = "SECOND"
	dataSet := &DataSet{resultSets: []ResultSet{
		{cols: &[]ParameterInfo{first}},
		{cols: &[]ParameterInfo{second}},
	}}
	if got := dataSet.ColumnTypeDatabaseTypeName(0); got != "VARCHAR2" {
		t.Fatalf("first DatabaseTypeName() = %q", got)
	}
	if err := dataSet.NextResultSet(); err != nil {
		t.Fatalf("NextResultSet() failed: %v", err)
	}
	if got := dataSet.ColumnTypeDatabaseTypeName(0); got != "RAW" {
		t.Fatalf("second DatabaseTypeName() = %q", got)
	}
	if err := dataSet.NextResultSet(); err != io.EOF {
		t.Fatalf("final NextResultSet() = %v, want io.EOF", err)
	}
}

func TestDataSetMetadataEmptyAndCleared(t *testing.T) {
	var nilDataSet *DataSet
	if got := nilDataSet.ColumnTypeDatabaseTypeName(0); got != "" {
		t.Fatalf("nil DataSet name = %q", got)
	}
	dataSet := &DataSet{resultSets: make([]ResultSet, 0)}
	if got := dataSet.ColumnTypeScanType(0); got != nil {
		t.Fatalf("empty DataSet ScanType() = %v", got)
	}
	dataSet.clear()
	if got := dataSet.ColumnTypeDatabaseTypeName(0); got != "" {
		t.Fatalf("cleared DataSet name = %q", got)
	}
}
