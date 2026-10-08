package go_ora

import (
	"bytes"
	"context"
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

// TestColumnTypeScanTypeMatchesCoderOutputs verifies at the driver layer that
// the descriptor ScanType agrees with the concrete driver.Value type each
// coder produces for a non-NULL value.
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
		SessionInfo: configurations.SessionInfo{
			SessionDataUnitSize:   0xFFFF,
			TransportDataUnitSize: 0xFFFF,
		},
		Lob:         fetch,
		LobReadMode: read,
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

// TestDecodePrimScanTypeConsistency exercises the real decode path and checks
// that the concrete runtime type matches the descriptor ScanType.
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

func TestDataSetMetadataResultSetSwitching(t *testing.T) {
	first := metadataColumn(types.NCHAR)
	first.Name = "FIRST"
	second := metadataColumn(types.RAW)
	second.Name = "SECOND"
	dataSet := &DataSet{resultSets: []ResultSet{
		*metadataResultSet(first),
		*metadataResultSet(second),
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
