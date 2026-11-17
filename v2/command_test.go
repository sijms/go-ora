package go_ora

import (
	"testing"

	"github.com/sijms/go-ora/v2/configurations"
	"github.com/sijms/go-ora/v2/converters"
	"github.com/sijms/go-ora/v2/network"
	"github.com/sijms/go-ora/v2/trace"
)

func TestNewStmt_WithComments(t *testing.T) {
	t.Run("SELECT", func(t *testing.T) {
		querySelectWithComments := `
-- comment #1
  -- comment #2
/* comment #3 */
  /* comment #4 */ select * from dual
`

		stmt := NewStmt(querySelectWithComments, nil)
		if stmt == nil {
			t.Errorf("no stmt returned")
		} else if stmt.stmtType != SELECT {
			t.Errorf("expected stmt.stmtType to be %v but was %v", SELECT, stmt.stmtType)
		}
	})

	t.Run("UPDATE", func(t *testing.T) {
		querySelectWithComments := `
-- comment #1
  -- comment #2
/* comment #3 */
  /* comment #4 */ update foo set bar = 1 where baz = 1
`

		stmt := NewStmt(querySelectWithComments, nil)
		if stmt == nil {
			t.Errorf("no stmt returned")
		} else if stmt.stmtType != DML {
			t.Errorf("expected stmt.stmtType to be %v but was %v", DML, stmt.stmtType)
		}
	})

	t.Run("DECLARE", func(t *testing.T) {
		querySelectWithComments := `
-- comment #1
  -- comment #2
/* comment #3 */
  /* comment #4 */ 
DECLARE
   foo NUMBER := 42;
BEGIN
   INSERT INTO bar VALUES (foo);
END;
`

		stmt := NewStmt(querySelectWithComments, nil)
		if stmt == nil {
			t.Errorf("no stmt returned")
		} else if stmt.stmtType != PLSQL {
			t.Errorf("expected stmt.stmtType to be %v but was %v", PLSQL, stmt.stmtType)
		}
	})
}

func TestColumnTypeDatabaseTypeName(t *testing.T) {
	tests := []struct {
		col  ParameterInfo
		want string
	}{
		{col: ParameterInfo{Name: "UNTYPED_NULL", DataType: 0}, want: ""},
		{col: ParameterInfo{Name: "VARCHAR2_COL", DataType: NCHAR, CharsetForm: 1}, want: "VARCHAR2"},
		{col: ParameterInfo{Name: "VARCHAR2_DEFAULT_CS", DataType: NCHAR, CharsetForm: 0}, want: "VARCHAR2"},
		{col: ParameterInfo{Name: "VARCHAR_WIRE", DataType: VARCHAR, CharsetForm: 1}, want: "VARCHAR2"},
		{col: ParameterInfo{Name: "OCISTRING_WIRE", DataType: OCIString, CharsetForm: 1}, want: "VARCHAR2"},
		{col: ParameterInfo{Name: "NULLSTR_WIRE", DataType: NullStr, CharsetForm: 1}, want: "VARCHAR2"},
		{col: ParameterInfo{Name: "NVARCHAR2_COL", DataType: NCHAR, CharsetForm: 2}, want: "NVARCHAR2"},
		{col: ParameterInfo{Name: "NVARCHAR_WIRE", DataType: VARCHAR, CharsetForm: 2}, want: "NVARCHAR2"},
		{col: ParameterInfo{Name: "CHAR_COL", DataType: CHAR, CharsetForm: 1}, want: "CHAR"},
		{col: ParameterInfo{Name: "CHAR_DEFAULT_CS", DataType: CHAR, CharsetForm: 0}, want: "CHAR"},
		{col: ParameterInfo{Name: "CHARZ_COL", DataType: CHARZ, CharsetForm: 1}, want: "CHAR"},
		{col: ParameterInfo{Name: "NCHAR_COL", DataType: CHAR, CharsetForm: 2}, want: "NCHAR"},
		{col: ParameterInfo{Name: "NCHARZ_COL", DataType: CHARZ, CharsetForm: 2}, want: "NCHAR"},
		{col: ParameterInfo{Name: "NUMBER_COL", DataType: NUMBER}, want: "NUMBER"},
		{col: ParameterInfo{Name: "VARNUM_COL", DataType: VarNum}, want: "NUMBER"},
		{col: ParameterInfo{Name: "BINTEGER_COL", DataType: BInteger}, want: "NUMBER"},
		{col: ParameterInfo{Name: "UINT_COL", DataType: UINT}, want: "NUMBER"},
		{col: ParameterInfo{Name: "PDN_COL", DataType: PDN}, want: "NUMBER"},
		{col: ParameterInfo{Name: "FLOAT_COL", DataType: FLOAT}, want: "FLOAT"},
		{col: ParameterInfo{Name: "IBFLOAT_COL", DataType: IBFloat}, want: "BINARY_FLOAT"},
		{col: ParameterInfo{Name: "BFLOAT_COL", DataType: BFloat}, want: "BINARY_FLOAT"},
		{col: ParameterInfo{Name: "IBDOUBLE_COL", DataType: IBDouble}, want: "BINARY_DOUBLE"},
		{col: ParameterInfo{Name: "BDOUBLE_COL", DataType: BDouble}, want: "BINARY_DOUBLE"},
		{col: ParameterInfo{Name: "DATE_COL", DataType: DATE}, want: "DATE"},
		{col: ParameterInfo{Name: "OCIDATE_COL", DataType: OCIDate}, want: "DATE"},
		{col: ParameterInfo{Name: "TIMESTAMP_DTY_COL", DataType: TimeStampDTY}, want: "TIMESTAMP"},
		{col: ParameterInfo{Name: "TIMESTAMP_COL", DataType: TIMESTAMP}, want: "TIMESTAMP"},
		{col: ParameterInfo{Name: "TIMESTAMPTZ_DTY_COL", DataType: TimeStampTZ_DTY}, want: "TIMESTAMP WITH TIME ZONE"},
		{col: ParameterInfo{Name: "TIMESTAMPTZ_COL", DataType: TIMESTAMPTZ}, want: "TIMESTAMP WITH TIME ZONE"},
		{col: ParameterInfo{Name: "TIMESTAMPLTZ_DTY_COL", DataType: TimeStampLTZ_DTY}, want: "TIMESTAMP WITH LOCAL TIME ZONE"},
		{col: ParameterInfo{Name: "TIMESTAMPLTZ_COL", DataType: TimeStampeLTZ}, want: "TIMESTAMP WITH LOCAL TIME ZONE"},
		{col: ParameterInfo{Name: "INTERVAL_YM_DTY_COL", DataType: IntervalYM_DTY}, want: "INTERVAL YEAR TO MONTH"},
		{col: ParameterInfo{Name: "INTERVAL_YM_COL", DataType: IntervalYM}, want: "INTERVAL YEAR TO MONTH"},
		{col: ParameterInfo{Name: "INTERVAL_DS_DTY_COL", DataType: IntervalDS_DTY}, want: "INTERVAL DAY TO SECOND"},
		{col: ParameterInfo{Name: "INTERVAL_DS_COL", DataType: IntervalDS}, want: "INTERVAL DAY TO SECOND"},
		{col: ParameterInfo{Name: "RAW_COL", DataType: RAW}, want: "RAW"},
		{col: ParameterInfo{Name: "VARRAW_COL", DataType: VarRaw}, want: "RAW"},
		{col: ParameterInfo{Name: "LONG_COL", DataType: LONG}, want: "LONG"},
		{col: ParameterInfo{Name: "LONGVARCHAR_COL", DataType: LongVarChar}, want: "LONG"},
		{col: ParameterInfo{Name: "LONGRAW_COL", DataType: LongRaw, describeDataType: LongRaw}, want: "LONG RAW"},
		{col: ParameterInfo{Name: "LONGVARRAW_COL", DataType: LongVarRaw}, want: "LONG RAW"},
		{col: ParameterInfo{Name: "ROWID_COL", DataType: ROWID}, want: "ROWID"},
		{col: ParameterInfo{Name: "UROWID_COL", DataType: UROWID}, want: "UROWID"},
		{col: ParameterInfo{Name: "CLOB_LOCATOR_COL", DataType: OCIClobLocator, CharsetForm: 1}, want: "CLOB"},
		{col: ParameterInfo{Name: "CLOB_INLINE_COL", DataType: LongVarChar, describeDataType: OCIClobLocator, CharsetForm: 1}, want: "CLOB"},
		{col: ParameterInfo{Name: "CLOB_IS_JSON_COL", DataType: OCIClobLocator, describeDataType: OCIClobLocator, CharsetForm: 1, IsJson: true, udsFlags: 0x100}, want: "CLOB"},
		{col: ParameterInfo{Name: "NCLOB_LOCATOR_COL", DataType: OCIClobLocator, CharsetForm: 2}, want: "NCLOB"},
		{col: ParameterInfo{Name: "NCLOB_INLINE_COL", DataType: LongVarChar, describeDataType: OCIClobLocator, CharsetForm: 2}, want: "NCLOB"},
		{col: ParameterInfo{Name: "BLOB_LOCATOR_COL", DataType: OCIBlobLocator}, want: "BLOB"},
		{col: ParameterInfo{Name: "BLOB_INLINE_COL", DataType: LongRaw, describeDataType: OCIBlobLocator}, want: "BLOB"},
		{col: ParameterInfo{Name: "BLOB_IS_JSON_CONSTRAINT_COL", DataType: OCIBlobLocator, describeDataType: OCIBlobLocator, IsJson: true, udsFlags: 0x100}, want: "BLOB"},
		{col: ParameterInfo{Name: "JSON_DOWNGRADED_BLOB_COL", DataType: OCIBlobLocator, describeDataType: OCIBlobLocator, IsJson: true, udsFlags: 0x400}, want: "JSON"},
		{col: ParameterInfo{Name: "JSON_DOWNGRADED_BOTH_FLAGS_COL", DataType: OCIBlobLocator, describeDataType: OCIBlobLocator, IsJson: true, udsFlags: 0x500}, want: "JSON"},
		{col: ParameterInfo{Name: "JSON_NATIVE_COL", DataType: JSON}, want: "JSON"},
		{col: ParameterInfo{Name: "BFILE_COL", DataType: OCIFileLocator}, want: "BFILE"},
		{col: ParameterInfo{Name: "VECTOR_LOCATOR_COL", DataType: VECTOR}, want: "VECTOR"},
		{col: ParameterInfo{Name: "VECTOR_INLINE_COL", DataType: LongRaw, describeDataType: VECTOR}, want: "VECTOR"},
		{col: ParameterInfo{Name: "BOOLEAN_COL", DataType: Boolean}, want: "BOOLEAN"},
		{col: ParameterInfo{Name: "REFCURSOR_COL", DataType: REFCURSOR}, want: "REFCURSOR"},
		{col: ParameterInfo{Name: "RESULTSET_COL", DataType: RESULTSET}, want: "REFCURSOR"},
		{col: ParameterInfo{Name: "REF_COL", DataType: OCIRef}, want: "REF"},
		{col: ParameterInfo{Name: "INT_REF_COL", DataType: TNSType(TNS_DATA_TYPE_INT_REF)}, want: "REF"},
		{col: ParameterInfo{Name: "OCIXMLTYPE_COL", DataType: OCIXMLType}, want: "XMLTYPE"},
		{col: ParameterInfo{Name: "XMLTYPE_COL", DataType: XMLType, describeDataType: XMLType, IsXmlType: true, SchemaName: "SYS", TypeName: "XMLTYPE"}, want: "XMLTYPE"},
		{col: ParameterInfo{Name: "XMLTYPE_DEFAULT_COL", DataType: XMLType}, want: "XMLTYPE"},
		{col: ParameterInfo{Name: "OBJECT_COL", DataType: XMLType, SchemaName: "MDSYS", TypeName: "SDO_GEOMETRY"}, want: "MDSYS.SDO_GEOMETRY"},
		{col: ParameterInfo{Name: "OBJECT_NAMED_BLOB_COL", DataType: XMLType, SchemaName: "APP", TypeName: "BLOB"}, want: "APP.BLOB"},
		{col: ParameterInfo{Name: "OBJECT_NO_SCHEMA_COL", DataType: XMLType, TypeName: "CUSTOM_OBJ"}, want: "CUSTOM_OBJ"},
		{col: ParameterInfo{Name: "FALLBACK_TIMETZ_COL", DataType: TimeTZ}, want: "TimeTZ"},
	}

	cols := make([]ParameterInfo, len(tests))
	for i, tc := range tests {
		cols[i] = tc.col
	}
	rs := ResultSet{cols: &cols}
	ds := DataSet{resultSets: []ResultSet{rs}}

	for i, tc := range tests {
		if got := rs.ColumnTypeDatabaseTypeName(i); got != tc.want {
			t.Errorf("ResultSet.ColumnTypeDatabaseTypeName(%d) (%s) = %q, want %q", i, tc.col.Name, got, tc.want)
		}
		if got := ds.ColumnTypeDatabaseTypeName(i); got != tc.want {
			t.Errorf("DataSet.ColumnTypeDatabaseTypeName(%d) (%s) = %q, want %q", i, tc.col.Name, got, tc.want)
		}
	}
}

func TestParameterInfoCloneAndInit(t *testing.T) {
	orig := ParameterInfo{
		DataType:         LongRaw,
		describeDataType: OCIBlobLocator,
		udsFlags:         0x400,
		IsJson:           true,
		IsXmlType:        true,
		SchemaName:       "MDSYS",
		TypeName:         "SDO_GEOMETRY",
		CharsetForm:      1,
	}
	cloned := orig.clone()
	if cloned.DataType != LongRaw || cloned.describeDataType != OCIBlobLocator || cloned.udsFlags != 0x400 || !cloned.IsJson || !cloned.IsXmlType || cloned.SchemaName != "MDSYS" || cloned.TypeName != "SDO_GEOMETRY" || cloned.CharsetForm != 1 {
		t.Errorf("clone() = %+v, want metadata preserved from %+v", cloned, orig)
	}

	cloned.init()
	if cloned.DataType != 0 || cloned.describeDataType != 0 || cloned.udsFlags != 0 || cloned.CharsetForm != 0 {
		t.Errorf("init() left non-zero wire metadata: got %+v, want DataType=0, describeDataType=0, udsFlags=0, CharsetForm=0", cloned)
	}
}

func TestWriteDefinePreservesColumnTypeDatabaseTypeName(t *testing.T) {
	cfg := &configurations.ConnectionConfig{
		Lob: configurations.INLINE,
		SessionInfo: configurations.SessionInfo{
			SessionDataUnitSize:   0xFFFF,
			TransportDataUnitSize: 0xFFFF,
		},
	}
	conn := &Connection{
		connOption: cfg,
		session:    network.NewSession(cfg, trace.NilTracer()),
	}
	stmt := defaultStmt{
		connection: conn,
		columns: []ParameterInfo{
			{Name: "BLOB_COL", DataType: OCIBlobLocator, describeDataType: OCIBlobLocator},
			{Name: "CLOB_COL", DataType: OCIClobLocator, describeDataType: OCIClobLocator, CharsetForm: 1},
			{Name: "NCLOB_COL", DataType: OCIClobLocator, describeDataType: OCIClobLocator, CharsetForm: 2},
			{Name: "VECTOR_COL", DataType: VECTOR, describeDataType: VECTOR},
			{Name: "BLOB_IS_JSON_COL", DataType: OCIBlobLocator, describeDataType: OCIBlobLocator, IsJson: true, udsFlags: 0x100},
			{Name: "JSON_OSON_COL", DataType: OCIBlobLocator, describeDataType: OCIBlobLocator, IsJson: true, udsFlags: 0x400},
		},
	}
	if err := stmt.writeDefine(); err != nil {
		t.Fatalf("writeDefine() failed: %v", err)
	}

	rs := ResultSet{cols: &stmt.columns}
	want := []struct {
		wireType TNSType
		sqlType  string
	}{
		{wireType: LongRaw, sqlType: "BLOB"},
		{wireType: LongVarChar, sqlType: "CLOB"},
		{wireType: LongVarChar, sqlType: "NCLOB"},
		{wireType: LongRaw, sqlType: "VECTOR"},
		{wireType: OCIBlobLocator, sqlType: "BLOB"},
		{wireType: OCIBlobLocator, sqlType: "JSON"},
	}
	for i, w := range want {
		if gotWire := stmt.columns[i].DataType; gotWire != w.wireType {
			t.Errorf("stmt.columns[%d] (%s) DataType after writeDefine = %v, want %v", i, stmt.columns[i].Name, gotWire, w.wireType)
		}
		if gotSQL := rs.ColumnTypeDatabaseTypeName(i); gotSQL != w.sqlType {
			t.Errorf("ResultSet.ColumnTypeDatabaseTypeName(%d) (%s) after writeDefine = %q, want %q", i, stmt.columns[i].Name, gotSQL, w.sqlType)
		}
	}
}

// describeColumn holds the fields of a TTC column describe payload decoded by
// ParameterInfo.load for TTC version 6.
type describeColumn struct {
	dataType    TNSType
	precision   uint8
	scale       int
	charsetForm uint8
	name        string
	schemaName  string
	typeName    string
	udsFlags    int
}

// compressedInt encodes v in the TTC compressed integer format read by
// network.Session.GetInt(size, true, true).
func compressedInt(v int) []byte {
	if v == 0 {
		return []byte{0}
	}
	neg := v < 0
	if neg {
		v = -v
	}
	var b []byte
	for v > 0 {
		b = append([]byte{byte(v)}, b...)
		v >>= 8
	}
	n := byte(len(b))
	if neg {
		n |= 0x80
	}
	return append([]byte{n}, b...)
}

// dlc encodes s in the format read by network.Session.GetDlc.
func dlc(s string) []byte {
	if s == "" {
		return []byte{0}
	}
	b := compressedInt(len(s))
	b = append(b, byte(len(s)))
	return append(b, s...)
}

func (d describeColumn) bytes() []byte {
	b := []byte{byte(d.dataType), 0, d.precision}
	if d.dataType == NUMBER {
		b = append(b, compressedInt(d.scale)...)
	} else {
		b = append(b, byte(d.scale))
	}
	b = append(b, 0, 0, 0) // MaxLen, MaxNoOfArrayElements, ContFlag
	b = append(b, 0)       // ToID
	b = append(b, 0, 0)    // Version, CharsetID
	b = append(b, d.charsetForm)
	b = append(b, 0)    // MaxCharLen
	b = append(b, 1, 0) // AllowNull, v7 name length
	b = append(b, dlc(d.name)...)
	b = append(b, dlc(d.schemaName)...)
	b = append(b, dlc(d.typeName)...)
	b = append(b, 0) // reserved 2-byte integer
	return append(b, compressedInt(d.udsFlags)...)
}

func TestParameterInfoLoadColumnTypeDatabaseTypeName(t *testing.T) {
	tests := []struct {
		name         string
		desc         describeColumn
		inlineDefine bool
		wantDataType TNSType
		wantTypeName string
	}{
		{
			name:         "NUMBER_UNCONSTRAINED",
			desc:         describeColumn{dataType: NUMBER, scale: -127, name: "N"},
			wantDataType: NUMBER,
			wantTypeName: "NUMBER",
		},
		{
			name:         "NUMBER_PRECISION_SCALE",
			desc:         describeColumn{dataType: NUMBER, precision: 10, scale: 2, name: "N"},
			wantDataType: NUMBER,
			wantTypeName: "NUMBER",
		},
		{
			name:         "FLOAT_126",
			desc:         describeColumn{dataType: NUMBER, precision: 126, scale: -127, name: "F"},
			wantDataType: NUMBER,
			wantTypeName: "FLOAT",
		},
		{
			name:         "REAL",
			desc:         describeColumn{dataType: NUMBER, precision: 63, scale: -127, name: "R"},
			wantDataType: NUMBER,
			wantTypeName: "FLOAT",
		},
		{
			name:         "BLOB_INLINE_PREFETCH",
			desc:         describeColumn{dataType: OCIBlobLocator, name: "B"},
			inlineDefine: true,
			wantDataType: LongRaw,
			wantTypeName: "BLOB",
		},
		{
			name:         "NCLOB_INLINE_PREFETCH",
			desc:         describeColumn{dataType: OCIClobLocator, charsetForm: 2, name: "NC"},
			inlineDefine: true,
			wantDataType: LongVarChar,
			wantTypeName: "NCLOB",
		},
		{
			name:         "BLOB_IS_JSON_CONSTRAINT",
			desc:         describeColumn{dataType: OCIBlobLocator, name: "BJ", udsFlags: 0x100},
			inlineDefine: true,
			wantDataType: OCIBlobLocator,
			wantTypeName: "BLOB",
		},
		{
			name:         "JSON_DOWNGRADED_BLOB",
			desc:         describeColumn{dataType: OCIBlobLocator, name: "J", udsFlags: 0x400},
			inlineDefine: true,
			wantDataType: OCIBlobLocator,
			wantTypeName: "JSON",
		},
		{
			name:         "XMLTYPE",
			desc:         describeColumn{dataType: XMLType, name: "X", schemaName: "SYS", typeName: "XMLTYPE"},
			wantDataType: XMLType,
			wantTypeName: "XMLTYPE",
		},
		{
			name:         "OBJECT",
			desc:         describeColumn{dataType: XMLType, name: "G", schemaName: "mdsys", typeName: "sdo_geometry"},
			wantDataType: XMLType,
			wantTypeName: "MDSYS.SDO_GEOMETRY",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			conn := &Connection{
				connOption: &configurations.ConnectionConfig{Lob: configurations.INLINE},
				session:    network.NewSessionWithInputBufferForDebug(tc.desc.bytes()),
			}
			conn.session.TTCVersion = 6
			conn.session.StrConv = converters.NewStringConverter(871)

			var par ParameterInfo
			if err := par.load(conn); err != nil {
				t.Fatalf("ParameterInfo.load() failed: %v", err)
			}
			cols := []ParameterInfo{par}
			if tc.inlineDefine {
				stmt := defaultStmt{connection: conn, columns: cols}
				if err := stmt.writeDefine(); err != nil {
					t.Fatalf("writeDefine() failed: %v", err)
				}
			}
			if got := cols[0].DataType; got != tc.wantDataType {
				t.Errorf("DataType = %v, want %v", got, tc.wantDataType)
			}
			rs := ResultSet{cols: &cols}
			if got := rs.ColumnTypeDatabaseTypeName(0); got != tc.wantTypeName {
				t.Errorf("ColumnTypeDatabaseTypeName(0) = %q, want %q", got, tc.wantTypeName)
			}
		})
	}
}
