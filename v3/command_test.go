package go_ora

import (
	"testing"

	"github.com/sijms/go-ora/v3/configurations"
	"github.com/sijms/go-ora/v3/converters"
	"github.com/sijms/go-ora/v3/network"
	"github.com/sijms/go-ora/v3/trace"
	"github.com/sijms/go-ora/v3/types"
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
	cases := []struct {
		name             string
		dataType         uint16
		describeDataType uint16
		charsetForm      int
		udsFlags         int
		isXmlType        bool
		schemaName       string
		typeName         string
		want             string
	}{
		{name: "UNTYPED_NULL", dataType: 0, want: ""},
		{name: "VARCHAR2_COL", dataType: types.NCHAR, charsetForm: 1, want: "VARCHAR2"},
		{name: "VARCHAR2_DEFAULT_CS", dataType: types.NCHAR, charsetForm: 0, want: "VARCHAR2"},
		{name: "VARCHAR_WIRE", dataType: types.VARCHAR, charsetForm: 1, want: "VARCHAR2"},
		{name: "OCISTRING_WIRE", dataType: types.OCIString, charsetForm: 1, want: "VARCHAR2"},
		{name: "NULLSTR_WIRE", dataType: types.NullStr, charsetForm: 1, want: "VARCHAR2"},
		{name: "NVARCHAR2_COL", dataType: types.NCHAR, charsetForm: 2, want: "NVARCHAR2"},
		{name: "NVARCHAR_WIRE", dataType: types.VARCHAR, charsetForm: 2, want: "NVARCHAR2"},
		{name: "CHAR_COL", dataType: types.CHAR, charsetForm: 1, want: "CHAR"},
		{name: "CHAR_DEFAULT_CS", dataType: types.CHAR, charsetForm: 0, want: "CHAR"},
		{name: "CHARZ_COL", dataType: types.CHARZ, charsetForm: 1, want: "CHAR"},
		{name: "NCHAR_COL", dataType: types.CHAR, charsetForm: 2, want: "NCHAR"},
		{name: "NCHARZ_COL", dataType: types.CHARZ, charsetForm: 2, want: "NCHAR"},
		{name: "NUMBER_COL", dataType: types.NUMBER, want: "NUMBER"},
		{name: "VARNUM_COL", dataType: types.VarNum, want: "NUMBER"},
		{name: "BINTEGER_COL", dataType: types.BInteger, want: "NUMBER"},
		{name: "UINT_COL", dataType: types.UINT, want: "NUMBER"},
		{name: "PDN_COL", dataType: types.PDN, want: "NUMBER"},
		{name: "FLOAT_COL", dataType: types.FLOAT, want: "FLOAT"},
		{name: "IBFLOAT_COL", dataType: types.IBFLOAT, want: "BINARY_FLOAT"},
		{name: "BFLOAT_COL", dataType: types.BFLOAT, want: "BINARY_FLOAT"},
		{name: "IBDOUBLE_COL", dataType: types.IBDOUBLE, want: "BINARY_DOUBLE"},
		{name: "BDOUBLE_COL", dataType: types.BDOUBLE, want: "BINARY_DOUBLE"},
		{name: "DATE_COL", dataType: types.DATE, want: "DATE"},
		{name: "OCIDATE_COL", dataType: types.OCIDate, want: "DATE"},
		{name: "TIMESTAMP_DTY_COL", dataType: types.TimeStampDTY, want: "TIMESTAMP"},
		{name: "TIMESTAMP_COL", dataType: types.TIMESTAMP, want: "TIMESTAMP"},
		{name: "TIMESTAMPTZ_DTY_COL", dataType: types.TimeStampTZ_DTY, want: "TIMESTAMP WITH TIME ZONE"},
		{name: "TIMESTAMPTZ_COL", dataType: types.TIMESTAMPTZ, want: "TIMESTAMP WITH TIME ZONE"},
		{name: "TIMESTAMPLTZ_DTY_COL", dataType: types.TimeStampLTZ_DTY, want: "TIMESTAMP WITH LOCAL TIME ZONE"},
		{name: "TIMESTAMPLTZ_COL", dataType: types.TimeStampLTZ, want: "TIMESTAMP WITH LOCAL TIME ZONE"},
		{name: "INTERVAL_YM_DTY_COL", dataType: types.INTERVALYM_DTY, want: "INTERVAL YEAR TO MONTH"},
		{name: "INTERVAL_YM_COL", dataType: types.IntervalYM, want: "INTERVAL YEAR TO MONTH"},
		{name: "INTERVAL_DS_DTY_COL", dataType: types.INTERVALDS_DTY, want: "INTERVAL DAY TO SECOND"},
		{name: "INTERVAL_DS_COL", dataType: types.IntervalDS, want: "INTERVAL DAY TO SECOND"},
		{name: "RAW_COL", dataType: types.RAW, want: "RAW"},
		{name: "VARRAW_COL", dataType: types.VarRaw, want: "RAW"},
		{name: "LONG_COL", dataType: types.LONG, want: "LONG"},
		{name: "LONGVARCHAR_COL", dataType: types.LongVarChar, want: "LONG"},
		{name: "LONGRAW_COL", dataType: types.LongRaw, describeDataType: types.LongRaw, want: "LONG RAW"},
		{name: "LONGVARRAW_COL", dataType: types.LongVarRaw, want: "LONG RAW"},
		{name: "ROWID_COL", dataType: types.ROWID, want: "ROWID"},
		{name: "UROWID_COL", dataType: types.UROWID, want: "UROWID"},
		{name: "CLOB_LOCATOR_COL", dataType: types.OCIClobLocator, charsetForm: 1, want: "CLOB"},
		{name: "CLOB_INLINE_COL", dataType: types.LongVarChar, describeDataType: types.OCIClobLocator, charsetForm: 1, want: "CLOB"},
		{name: "CLOB_IS_JSON_COL", dataType: types.OCIClobLocator, describeDataType: types.OCIClobLocator, charsetForm: 1, udsFlags: 0x100, want: "CLOB"},
		{name: "NCLOB_LOCATOR_COL", dataType: types.OCIClobLocator, charsetForm: 2, want: "NCLOB"},
		{name: "NCLOB_INLINE_COL", dataType: types.LongVarChar, describeDataType: types.OCIClobLocator, charsetForm: 2, want: "NCLOB"},
		{name: "BLOB_LOCATOR_COL", dataType: types.OCIBlobLocator, want: "BLOB"},
		{name: "BLOB_INLINE_COL", dataType: types.LongRaw, describeDataType: types.OCIBlobLocator, want: "BLOB"},
		{name: "BLOB_IS_JSON_CONSTRAINT_COL", dataType: types.OCIBlobLocator, describeDataType: types.OCIBlobLocator, udsFlags: 0x100, want: "BLOB"},
		{name: "JSON_DOWNGRADED_BLOB_COL", dataType: types.OCIBlobLocator, describeDataType: types.OCIBlobLocator, udsFlags: 0x400, want: "JSON"},
		{name: "JSON_DOWNGRADED_BOTH_FLAGS_COL", dataType: types.OCIBlobLocator, describeDataType: types.OCIBlobLocator, udsFlags: 0x500, want: "JSON"},
		{name: "JSON_NATIVE_COL", dataType: types.JSON, want: "JSON"},
		{name: "BFILE_COL", dataType: types.OCIFileLocator, want: "BFILE"},
		{name: "VECTOR_LOCATOR_COL", dataType: types.VECTOR, want: "VECTOR"},
		{name: "VECTOR_INLINE_COL", dataType: types.LongRaw, describeDataType: types.VECTOR, want: "VECTOR"},
		{name: "BOOLEAN_COL", dataType: types.BOOLEAN, want: "BOOLEAN"},
		{name: "REFCURSOR_COL", dataType: types.REFCURSOR, want: "REFCURSOR"},
		{name: "RESULTSET_COL", dataType: types.RESULTSET, want: "REFCURSOR"},
		{name: "REF_COL", dataType: types.OCIRef, want: "REF"},
		{name: "INT_REF_COL", dataType: TNS_DATA_TYPE_INT_REF, want: "REF"},
		{name: "OCIXMLTYPE_COL", dataType: types.OCIXMLType, want: "XMLTYPE"},
		{name: "XMLTYPE_COL", dataType: types.XMLType, describeDataType: types.XMLType, isXmlType: true, schemaName: "SYS", typeName: "XMLTYPE", want: "XMLTYPE"},
		{name: "XMLTYPE_DEFAULT_COL", dataType: types.XMLType, want: "XMLTYPE"},
		{name: "OBJECT_COL", dataType: types.XMLType, schemaName: "MDSYS", typeName: "SDO_GEOMETRY", want: "MDSYS.SDO_GEOMETRY"},
		{name: "OBJECT_NAMED_BLOB_COL", dataType: types.XMLType, schemaName: "APP", typeName: "BLOB", want: "APP.BLOB"},
		{name: "OBJECT_NO_SCHEMA_COL", dataType: types.XMLType, typeName: "CUSTOM_OBJ", want: "CUSTOM_OBJ"},
		{name: "FALLBACK_TIMETZ_COL", dataType: types.TimeTZ, want: "TimeTZ"},
		{name: "FALLBACK_UNKNOWN_COL", dataType: 999, want: ""},
	}
	cols := make([]ParameterInfo, len(cases))
	for i, c := range cases {
		cols[i].Name = c.name
		cols[i].DataType = c.dataType
		cols[i].describeDataType = c.describeDataType
		cols[i].CharsetForm = c.charsetForm
		cols[i].udsFlags = c.udsFlags
		cols[i].IsXmlType = c.isXmlType
		cols[i].SchemaName = c.schemaName
		cols[i].TypeName = c.typeName
	}
	rs := ResultSet{cols: &cols}
	ds := DataSet{resultSets: []ResultSet{rs}}

	for i, c := range cases {
		if got := rs.ColumnTypeDatabaseTypeName(i); got != c.want {
			t.Errorf("ResultSet.ColumnTypeDatabaseTypeName(%d) (%s) = %q, want %q", i, c.name, got, c.want)
		}
		if got := ds.ColumnTypeDatabaseTypeName(i); got != c.want {
			t.Errorf("DataSet.ColumnTypeDatabaseTypeName(%d) (%s) = %q, want %q", i, c.name, got, c.want)
		}
	}
}

func TestParameterInfoCloneAndInit(t *testing.T) {
	orig := ParameterInfo{
		describeDataType: types.OCIBlobLocator,
		udsFlags:         0x400,
		IsJson:           true,
		IsXmlType:        true,
		SchemaName:       "MDSYS",
	}
	orig.DataType = types.LongRaw
	orig.TypeName = "SDO_GEOMETRY"
	orig.CharsetForm = 1

	cloned := orig.clone()
	if cloned.DataType != types.LongRaw || cloned.describeDataType != types.OCIBlobLocator || cloned.udsFlags != 0x400 || !cloned.IsJson || !cloned.IsXmlType || cloned.SchemaName != "MDSYS" || cloned.TypeName != "SDO_GEOMETRY" || cloned.CharsetForm != 1 {
		t.Errorf("clone() = %+v, want metadata preserved from %+v", cloned, orig)
	}

	cloned.init()
	if cloned.DataType != 0 || cloned.describeDataType != 0 || cloned.udsFlags != 0 {
		t.Errorf("init() left non-zero wire metadata: got %+v, want DataType=0, describeDataType=0, udsFlags=0", cloned)
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

	makeCol := func(name string, dataType uint16, charsetForm int, isJson bool, udsFlags int) ParameterInfo {
		var p ParameterInfo
		p.Name = name
		p.DataType = dataType
		p.describeDataType = dataType
		p.CharsetForm = charsetForm
		p.IsJson = isJson
		p.udsFlags = udsFlags
		return p
	}

	stmt := defaultStmt{
		connection: conn,
		columns: []ParameterInfo{
			makeCol("BLOB_COL", types.OCIBlobLocator, 0, false, 0),
			makeCol("CLOB_COL", types.OCIClobLocator, 1, false, 0),
			makeCol("NCLOB_COL", types.OCIClobLocator, 2, false, 0),
			makeCol("VECTOR_COL", types.VECTOR, 0, false, 0),
			makeCol("JSON_NATIVE_COL", types.JSON, 0, false, 0),
			makeCol("BLOB_IS_JSON_COL", types.OCIBlobLocator, 0, true, 0x100),
			makeCol("JSON_OSON_COL", types.OCIBlobLocator, 0, true, 0x400),
		},
	}
	if err := stmt.writeDefine(); err != nil {
		t.Fatalf("writeDefine() failed: %v", err)
	}

	rs := ResultSet{cols: &stmt.columns}
	want := []struct {
		wireType uint16
		sqlType  string
	}{
		{wireType: types.LongRaw, sqlType: "BLOB"},
		{wireType: types.LongVarChar, sqlType: "CLOB"},
		{wireType: types.LongVarChar, sqlType: "NCLOB"},
		{wireType: types.LongRaw, sqlType: "VECTOR"},
		{wireType: types.LongRaw, sqlType: "JSON"},
		{wireType: types.OCIBlobLocator, sqlType: "BLOB"},
		{wireType: types.OCIBlobLocator, sqlType: "JSON"},
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
	dataType    uint16
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
	if d.dataType == types.NUMBER {
		b = append(b, compressedInt(d.scale)...)
	} else {
		b = append(b, byte(d.scale))
	}
	b = append(b, 0, 0, 0) // MaxLen, ArraySize, ContFlag
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
		wantDataType uint16
		wantTypeName string
	}{
		{
			name:         "NUMBER_UNCONSTRAINED",
			desc:         describeColumn{dataType: types.NUMBER, scale: -127, name: "N"},
			wantDataType: types.NUMBER,
			wantTypeName: "NUMBER",
		},
		{
			name:         "NUMBER_PRECISION_SCALE",
			desc:         describeColumn{dataType: types.NUMBER, precision: 10, scale: 2, name: "N"},
			wantDataType: types.NUMBER,
			wantTypeName: "NUMBER",
		},
		{
			name:         "FLOAT_126",
			desc:         describeColumn{dataType: types.NUMBER, precision: 126, scale: -127, name: "F"},
			wantDataType: types.NUMBER,
			wantTypeName: "FLOAT",
		},
		{
			name:         "REAL",
			desc:         describeColumn{dataType: types.NUMBER, precision: 63, scale: -127, name: "R"},
			wantDataType: types.NUMBER,
			wantTypeName: "FLOAT",
		},
		{
			name:         "BLOB_INLINE_PREFETCH",
			desc:         describeColumn{dataType: types.OCIBlobLocator, name: "B"},
			inlineDefine: true,
			wantDataType: types.LongRaw,
			wantTypeName: "BLOB",
		},
		{
			name:         "NCLOB_INLINE_PREFETCH",
			desc:         describeColumn{dataType: types.OCIClobLocator, charsetForm: 2, name: "NC"},
			inlineDefine: true,
			wantDataType: types.LongVarChar,
			wantTypeName: "NCLOB",
		},
		{
			name:         "BLOB_IS_JSON_CONSTRAINT",
			desc:         describeColumn{dataType: types.OCIBlobLocator, name: "BJ", udsFlags: 0x100},
			inlineDefine: true,
			wantDataType: types.OCIBlobLocator,
			wantTypeName: "BLOB",
		},
		{
			name:         "JSON_DOWNGRADED_BLOB",
			desc:         describeColumn{dataType: types.OCIBlobLocator, name: "J", udsFlags: 0x400},
			inlineDefine: true,
			wantDataType: types.OCIBlobLocator,
			wantTypeName: "JSON",
		},
		{
			name:         "XMLTYPE",
			desc:         describeColumn{dataType: types.XMLType, name: "X", schemaName: "SYS", typeName: "XMLTYPE"},
			wantDataType: types.XMLType,
			wantTypeName: "XMLTYPE",
		},
		{
			name:         "OBJECT",
			desc:         describeColumn{dataType: types.XMLType, name: "G", schemaName: "mdsys", typeName: "sdo_geometry"},
			wantDataType: types.XMLType,
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
