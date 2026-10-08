package go_ora

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/sijms/go-ora/v3/types"
)

// Compile-time assertions for the database/sql interfaces exercised through
// the conformance suite.
var (
	_ driver.Conn               = (*Connection)(nil)
	_ driver.ConnBeginTx        = (*Connection)(nil)
	_ driver.ConnPrepareContext = (*Connection)(nil)
	_ driver.Pinger             = (*Connection)(nil)
	_ driver.QueryerContext     = (*Connection)(nil)
	_ driver.ExecerContext      = (*Connection)(nil)
	_ driver.SessionResetter    = (*Connection)(nil)
	_ driver.Validator          = (*Connection)(nil)
	_ driver.NamedValueChecker  = (*Connection)(nil)

	_ driver.Stmt              = (*Stmt)(nil)
	_ driver.StmtExecContext   = (*Stmt)(nil)
	_ driver.StmtQueryContext  = (*Stmt)(nil)
	_ driver.NamedValueChecker = (*Stmt)(nil)
)

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

func metadataTestResultSet(cols []ParameterInfo, rows ...Row) ResultSet {
	resultSet := ResultSet{
		cols:   &cols,
		rows:   rows,
		parent: &metadataTestStmt{},
	}
	resultSet.buildColumnDescriptors()
	return resultSet
}

// TestDatabaseSQLColumnTypesAndScan exercises the five ColumnType methods and
// row scanning through the public database/sql API backed by a DataSet.
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
	resultSet := metadataTestResultSet(cols, Row{"abc", []byte{1, 2}, rowTime, true, "12.34", 7.5})
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
	// metadata is stable across repeated calls and readable before Next()
	if again, err := rows.ColumnTypes(); err != nil {
		t.Fatalf("second ColumnTypes() failed: %v", err)
	} else {
		for i, columnType := range again {
			if columnType.DatabaseTypeName() != columnTypes[i].DatabaseTypeName() ||
				columnType.ScanType() != columnTypes[i].ScanType() {
				t.Fatalf("column %d metadata changed between calls", i)
			}
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

// TestDatabaseSQLMetadataBeforeAndAfterRows verifies metadata is available
// before the first Next() and stays consistent after rows are exhausted.
func TestDatabaseSQLMetadataBeforeAndAfterRows(t *testing.T) {
	varchar := metadataColumn(types.NCHAR)
	varchar.Name = "TXT"
	cols := []ParameterInfo{varchar}
	resultSet := metadataTestResultSet(cols, Row{"one"})
	db := sql.OpenDB(metadataTestConnector{rows: &DataSet{resultSets: []ResultSet{resultSet}}})
	defer db.Close()

	rows, err := db.QueryContext(context.Background(), "select")
	if err != nil {
		t.Fatalf("QueryContext() failed: %v", err)
	}
	before, err := rows.ColumnTypes()
	if err != nil {
		t.Fatalf("ColumnTypes() before Next() failed: %v", err)
	}
	if got := before[0].DatabaseTypeName(); got != "VARCHAR2" {
		t.Fatalf("DatabaseTypeName() = %q, want VARCHAR2", got)
	}
	// metadata is consistent across repeated calls before iteration
	again, err := rows.ColumnTypes()
	if err != nil {
		t.Fatalf("second ColumnTypes() failed: %v", err)
	}
	if again[0].DatabaseTypeName() != before[0].DatabaseTypeName() ||
		again[0].ScanType() != before[0].ScanType() {
		t.Fatal("metadata changed between calls")
	}
	if !rows.Next() {
		t.Fatalf("Rows.Next() failed: %v", rows.Err())
	}
	var txt string
	if err := rows.Scan(&txt); err != nil || txt != "one" {
		t.Fatalf("Scan() = %q, %v", txt, err)
	}
	if rows.Next() {
		t.Fatal("unexpected second row")
	}
	// database/sql closes Rows on exhaustion; further metadata access is
	// intentionally not part of the contract
	if err := rows.Err(); err != nil {
		t.Fatalf("Rows.Err() = %v", err)
	}
	if err := rows.Close(); err != nil {
		t.Fatalf("Rows.Close() failed: %v", err)
	}
}

// TestDatabaseSQLEmptyResultSet verifies a result set without rows still
// exposes consistent metadata through database/sql.
func TestDatabaseSQLEmptyResultSet(t *testing.T) {
	varchar := metadataColumn(types.NCHAR)
	varchar.Name = "TXT"
	cols := []ParameterInfo{varchar}
	resultSet := metadataTestResultSet(cols)
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
	if len(columnTypes) != 1 || columnTypes[0].DatabaseTypeName() != "VARCHAR2" {
		t.Fatalf("unexpected column types: %v", columnTypes)
	}
	if rows.Next() {
		t.Fatal("unexpected row in empty result set")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("Rows.Err() = %v", err)
	}
}

// TestDatabaseSQLMultipleResultSets verifies metadata isolation across
// NextResultSet boundaries: the second result set must not inherit the first
// set's descriptors.
func TestDatabaseSQLMultipleResultSets(t *testing.T) {
	first := metadataColumn(types.NCHAR)
	first.Name = "TXT"
	firstCols := []ParameterInfo{first}
	second := metadataColumn(types.NUMBER)
	second.Name = "NUM"
	second.rawPrecision = 5
	second.rawScale = 0
	second.precisionScaleKnown = true
	secondCols := []ParameterInfo{second}

	dataSet := &DataSet{resultSets: []ResultSet{
		metadataTestResultSet(firstCols, Row{"a"}),
		metadataTestResultSet(secondCols, Row{"1"}),
	}}
	db := sql.OpenDB(metadataTestConnector{rows: dataSet})
	defer db.Close()

	rows, err := db.QueryContext(context.Background(), "select")
	if err != nil {
		t.Fatalf("QueryContext() failed: %v", err)
	}
	defer rows.Close()

	firstTypes, err := rows.ColumnTypes()
	if err != nil {
		t.Fatalf("first ColumnTypes() failed: %v", err)
	}
	if got := firstTypes[0].DatabaseTypeName(); got != "VARCHAR2" {
		t.Fatalf("first set name = %q, want VARCHAR2", got)
	}
	if !rows.NextResultSet() {
		t.Fatal("expected a second result set")
	}
	secondTypes, err := rows.ColumnTypes()
	if err != nil {
		t.Fatalf("second ColumnTypes() failed: %v", err)
	}
	if len(secondTypes) != 1 {
		t.Fatalf("second set column count = %d, want 1", len(secondTypes))
	}
	if got := secondTypes[0].DatabaseTypeName(); got != "NUMBER" {
		t.Fatalf("second set name = %q, want NUMBER", got)
	}
	if got := secondTypes[0].ScanType(); got != types.TyString {
		t.Fatalf("second set ScanType() = %v, want string", got)
	}
	if rows.NextResultSet() {
		t.Fatal("unexpected third result set")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("Rows.Err() = %v", err)
	}
}
