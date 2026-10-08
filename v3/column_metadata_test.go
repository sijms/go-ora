package go_ora_test

import (
	"database/sql"
	"errors"
	"math"
	"os"
	"reflect"
	"testing"
	"time"

	_ "github.com/sijms/go-ora/v3"
	"github.com/sijms/go-ora/v3/network"
)

// metadataTestDB opens a connection through the public database/sql driver.
// The suite is opt-in: it skips cleanly when ORACLE_DSN is unset.
func metadataTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("ORACLE_DSN")
	if dsn == "" {
		t.Skip("ORACLE_DSN is not set")
	}
	db, err := sql.Open("oracle", dsn)
	if err != nil {
		t.Fatalf("open ORACLE_DSN: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		t.Fatalf("ping ORACLE_DSN: %v", err)
	}
	return db
}

// oracleErrorCode extracts the ORA error number from a driver error.
func oracleErrorCode(err error) (int, bool) {
	var oracleErr *network.OracleError
	if errors.As(err, &oracleErr) {
		return oracleErr.ErrCode, true
	}
	return 0, false
}

// metadataColumnTypes runs a metadata query. Each optional-feature test passes
// the narrowly scoped ORA codes documented for "feature not available" on its
// Oracle version; any other error fails the test.
func metadataColumnTypes(t *testing.T, db *sql.DB, query string, unsupportedCodes ...int) (*sql.Rows, []*sql.ColumnType) {
	t.Helper()
	rows, err := db.Query(query)
	if err != nil {
		if code, ok := oracleErrorCode(err); ok {
			for _, expected := range unsupportedCodes {
				if code == expected {
					t.Skipf("optional metadata query unsupported by this Oracle version: ORA-%05d", code)
				}
			}
		}
		t.Fatalf("query failed: %v", err)
	}
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		rows.Close()
		t.Fatalf("ColumnTypes() failed: %v", err)
	}
	return rows, columnTypes
}

func TestColumnMetadataCoreTypes(t *testing.T) {
	db := metadataTestDB(t)
	defer db.Close()

	rows, columnTypes := metadataColumnTypes(t, db, `SELECT
		CAST('x' AS CHAR(4)) AS C_CHAR,
		CAST('x' AS NCHAR(4)) AS C_NCHAR,
		CAST('x' AS VARCHAR2(16)) AS C_VARCHAR2,
		CAST('x' AS NVARCHAR2(16)) AS C_NVARCHAR2,
		CAST(123 AS NUMBER(10,-2)) AS C_NUMBER_NEG_SCALE,
		CAST(123 AS NUMBER(8,0)) AS C_NUMBER,
		CAST(NULL AS NUMBER) AS C_NUMBER_UNCONSTRAINED,
		CAST(1.5 AS BINARY_FLOAT) AS C_BINARY_FLOAT,
		CAST(1.5 AS BINARY_DOUBLE) AS C_BINARY_DOUBLE,
		CAST(SYSDATE AS DATE) AS C_DATE,
		CAST(SYSTIMESTAMP AS TIMESTAMP) AS C_TIMESTAMP,
		CAST(LOCALTIMESTAMP AS TIMESTAMP WITH LOCAL TIME ZONE) AS C_TIMESTAMP_LTZ,
		CAST(INTERVAL '1-2' YEAR TO MONTH AS INTERVAL YEAR TO MONTH) AS C_INTERVAL_YM,
		CAST(INTERVAL '1 02:03:04' DAY TO SECOND AS INTERVAL DAY TO SECOND) AS C_INTERVAL_DS,
		UTL_RAW.CAST_TO_RAW('test') AS C_RAW
	FROM dual`)
	defer rows.Close()

	wantNames := []string{
		"CHAR", "NCHAR", "VARCHAR2", "NVARCHAR2",
		"NUMBER", "NUMBER", "NUMBER",
		"BINARY_FLOAT", "BINARY_DOUBLE",
		"DATE", "TIMESTAMP", "TIMESTAMP WITH LOCAL TIME ZONE",
		"INTERVAL YEAR TO MONTH", "INTERVAL DAY TO SECOND",
		"RAW",
	}
	if len(columnTypes) != len(wantNames) {
		t.Fatalf("got %d columns, want %d", len(columnTypes), len(wantNames))
	}
	for i, want := range wantNames {
		if got := columnTypes[i].DatabaseTypeName(); got != want {
			t.Errorf("column %d DatabaseTypeName() = %q, want %q", i, got, want)
		}
		if _, ok := columnTypes[i].Nullable(); !ok {
			t.Errorf("column %d nullability is unknown", i)
		}
	}
	for _, index := range []int{0, 1} {
		if length, ok := columnTypes[index].Length(); length != 0 || ok {
			t.Errorf("fixed character column %d Length() = (%d,%v), want (0,false)", index, length, ok)
		}
	}
	for _, index := range []int{2, 3} {
		if length, ok := columnTypes[index].Length(); length != 16 || !ok {
			t.Errorf("variable character column %d Length() = (%d,%v), want (16,true)", index, length, ok)
		}
	}
	if precision, scale, ok := columnTypes[4].DecimalSize(); precision != 10 || scale != -2 || !ok {
		t.Errorf("negative-scale NUMBER DecimalSize() = (%d,%d,%v)", precision, scale, ok)
	}
	if precision, scale, ok := columnTypes[5].DecimalSize(); precision != 8 || scale != 0 || !ok {
		t.Errorf("NUMBER DecimalSize() = (%d,%d,%v)", precision, scale, ok)
	}
	if precision, scale, ok := columnTypes[6].DecimalSize(); precision != math.MaxInt64 || scale != math.MaxInt64 || !ok {
		t.Errorf("unconstrained NUMBER DecimalSize() = (%d,%d,%v)", precision, scale, ok)
	}
	if length, ok := columnTypes[14].Length(); length <= 0 || !ok {
		t.Errorf("RAW Length() = (%d,%v)", length, ok)
	}

	if !rows.Next() {
		t.Fatalf("no core row returned: %v", rows.Err())
	}
	values := make([]interface{}, len(columnTypes))
	dest := make([]interface{}, len(values))
	standardScanDestinations := map[int]bool{
		0: true, 1: true, 2: true, 3: true, 4: true, 5: true,
		8: true, 9: true, 10: true, 11: true, 12: true, 13: true, 14: true,
	}
	for i := range values {
		if standardScanDestinations[i] {
			if scanType := columnTypes[i].ScanType(); scanType != nil {
				dest[i] = reflect.New(scanType).Interface()
				continue
			}
		}
		dest[i] = &values[i]
	}
	if err := rows.Scan(dest...); err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}
	for i := range values {
		if standardScanDestinations[i] {
			values[i] = reflect.ValueOf(dest[i]).Elem().Interface()
		}
	}
	wantRuntimeTypes := []reflect.Type{
		reflect.TypeOf(""), reflect.TypeOf(""), reflect.TypeOf(""), reflect.TypeOf(""),
		reflect.TypeOf(""), reflect.TypeOf(""), nil,
		reflect.TypeOf(float32(0)), reflect.TypeOf(float64(0)),
		reflect.TypeOf(time.Time{}), reflect.TypeOf(time.Time{}), reflect.TypeOf(time.Time{}),
		reflect.TypeOf(time.Time{}), reflect.TypeOf(time.Time{}),
		reflect.TypeOf([]byte{}),
	}
	for i, want := range wantRuntimeTypes {
		if actual := reflect.TypeOf(values[i]); actual != want {
			t.Errorf("column %d runtime type = %v, want %v", i, actual, want)
		}
		if actual := reflect.TypeOf(values[i]); actual != nil && actual != columnTypes[i].ScanType() {
			t.Errorf("column %d ScanType() = %v, runtime type %v", i, columnTypes[i].ScanType(), actual)
		}
	}
}

func TestColumnMetadataLobAndXML(t *testing.T) {
	db := metadataTestDB(t)
	defer db.Close()

	rows, columnTypes := metadataColumnTypes(t, db, `SELECT
		CAST('abc' AS CLOB) AS C_CLOB,
		CAST('abc' AS NCLOB) AS C_NCLOB,
		TO_BLOB(UTL_RAW.CAST_TO_RAW('abc')) AS C_BLOB,
		XMLTYPE('<a/>') AS C_XML
	FROM dual`)
	defer rows.Close()

	wantNames := []string{"CLOB", "NCLOB", "BLOB", "XMLTYPE"}
	for i, want := range wantNames {
		if got := columnTypes[i].DatabaseTypeName(); got != want {
			t.Errorf("column %d DatabaseTypeName() = %q, want %q", i, got, want)
		}
		if length, ok := columnTypes[i].Length(); length != math.MaxInt64 || !ok {
			t.Errorf("column %d Length() = (%d,%v), want unbounded", i, length, ok)
		}
	}
	allowedScanTypes := [][]string{
		{"string", "*types.Clob"},
		{"string", "*types.Clob"},
		{"[]uint8", "*types.Blob"},
	}
	for i, allowed := range allowedScanTypes {
		scanType := columnTypes[i].ScanType()
		name := ""
		if scanType != nil {
			name = scanType.String()
		}
		found := false
		for _, candidate := range allowed {
			if name == candidate {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("column %d ScanType() = %v, expected one of %v", i, scanType, allowed)
		}
	}
	if got := columnTypes[3].ScanType(); got != nil {
		t.Errorf("XMLTYPE ScanType() = %v, want current nil", got)
	}
}

// Optional types are gated by narrowly scoped ORA codes:
// ORA-00904 invalid identifier (unknown function/keyword on older versions),
// ORA-00900 invalid SQL statement, ORA-00923 FROM keyword not found.
func TestColumnMetadataOptionalTypes(t *testing.T) {
	db := metadataTestDB(t)
	defer db.Close()

	t.Run("JSON", func(t *testing.T) {
		rows, columnTypes := metadataColumnTypes(t, db, `SELECT JSON('{"a":1}') FROM dual`, 904, 900)
		defer rows.Close()
		if got := columnTypes[0].DatabaseTypeName(); got != "JSON" {
			t.Fatalf("DatabaseTypeName() = %q, want JSON", got)
		}
		if length, ok := columnTypes[0].Length(); length != math.MaxInt64 || !ok {
			t.Fatalf("Length() = (%d,%v), want unbounded", length, ok)
		}
	})

	t.Run("VECTOR", func(t *testing.T) {
		rows, columnTypes := metadataColumnTypes(t, db, `SELECT VECTOR('[1.5,2.5]', 2, FLOAT32) FROM dual`, 904, 900)
		defer rows.Close()
		if got := columnTypes[0].DatabaseTypeName(); got != "VECTOR" {
			t.Fatalf("DatabaseTypeName() = %q, want VECTOR", got)
		}
		if length, ok := columnTypes[0].Length(); length != math.MaxInt64 || !ok {
			t.Fatalf("Length() = (%d,%v), want unbounded", length, ok)
		}
	})

	t.Run("BOOLEAN", func(t *testing.T) {
		rows, columnTypes := metadataColumnTypes(t, db, `SELECT TRUE FROM dual`, 904, 900, 923)
		defer rows.Close()
		if got := columnTypes[0].DatabaseTypeName(); got != "BOOLEAN" {
			t.Fatalf("DatabaseTypeName() = %q, want BOOLEAN", got)
		}
	})
}

// TestColumnMetadataBareNull is the regression for the result-parsing issue
// reported upstream as #730: a bare untyped NULL column must not break the
// row parse, and its metadata must report safe defaults.
func TestColumnMetadataBareNull(t *testing.T) {
	db := metadataTestDB(t)
	defer db.Close()

	rows, columnTypes := metadataColumnTypes(t, db, `SELECT 1 AS a, NULL AS b, 2 AS c FROM dual`)
	defer rows.Close()
	if len(columnTypes) != 3 {
		t.Fatalf("got %d columns, want 3", len(columnTypes))
	}
	if got := columnTypes[0].DatabaseTypeName(); got != "NUMBER" {
		t.Errorf("column 0 DatabaseTypeName() = %q, want NUMBER", got)
	}
	if !rows.Next() {
		t.Fatalf("bare NULL row not returned: %v", rows.Err())
	}
	var a, b, c interface{}
	if err := rows.Scan(&a, &b, &c); err != nil {
		t.Fatalf("Scan() failed: %v", err)
	}
	if b != nil {
		t.Errorf("bare NULL column scanned as %#v, want nil", b)
	}
	if rows.Next() {
		t.Error("unexpected second row")
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("Rows.Err() = %v", err)
	}
}

func TestColumnMetadataMultipleResultSets(t *testing.T) {
	db := metadataTestDB(t)
	defer db.Close()

	rows, err := db.Query(`DECLARE
		first_cur sys_refcursor;
		second_cur sys_refcursor;
	BEGIN
		OPEN first_cur FOR SELECT CAST('x' AS VARCHAR2(10)) AS FIRST_VALUE FROM dual;
		dbms_sql.return_result(first_cur);
		OPEN second_cur FOR SELECT CAST(1 AS NUMBER(5,0)) AS SECOND_VALUE FROM dual;
		dbms_sql.return_result(second_cur);
	END;`)
	if err != nil {
		t.Fatalf("multiple-result query failed: %v", err)
	}
	defer rows.Close()

	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		t.Fatalf("first ColumnTypes() failed: %v", err)
	}
	if got := columnTypes[0].DatabaseTypeName(); got != "VARCHAR2" {
		t.Fatalf("first DatabaseTypeName() = %q", got)
	}
	if !rows.NextResultSet() {
		t.Fatalf("NextResultSet() failed: %v", rows.Err())
	}
	columnTypes, err = rows.ColumnTypes()
	if err != nil {
		t.Fatalf("second ColumnTypes() failed: %v", err)
	}
	if got := columnTypes[0].DatabaseTypeName(); got != "NUMBER" {
		t.Fatalf("second DatabaseTypeName() = %q", got)
	}
	if precision, scale, ok := columnTypes[0].DecimalSize(); precision != 5 || scale != 0 || !ok {
		t.Fatalf("second DecimalSize() = (%d,%d,%v)", precision, scale, ok)
	}
}
