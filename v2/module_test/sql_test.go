package v2_test

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func generatePlaceholders(n int) string {
	result := ""
	for i := 1; i < n; i++ {
		result += ":" + strconv.Itoa(i) + ","
	}
	result += ":" + strconv.Itoa(n)
	return result
}

func newDbConnection() (*sql.DB, error) {
	driver := "oracle"
	dsn := os.Getenv("ORACLE_DSN")
	if dsn == "" {
		return nil, nil
	}
	return sql.Open(driver, dsn)
}

func fetchOneRow(rows *sql.Rows, colsAndTypes []*sql.ColumnType) (map[string]interface{}, error) {
	numCols := len(colsAndTypes)
	vals := make([]interface{}, numCols)
	scanVals := make([]interface{}, numCols)
	for i := 0; i < numCols; i++ {
		scanVals[i] = &vals[i]
	}
	if err := rows.Scan(scanVals...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	result := map[string]interface{}{}
	for i, col := range colsAndTypes {
		result[strings.ToLower(col.Name())] = vals[i]
	}
	return result, nil
}

func fetchAllRowsColumnLowerCased(dbrows *sql.Rows) ([]map[string]interface{}, error) {
	defer func() { _ = dbrows.Close() }()

	colTypes, err := dbrows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0)
	for dbrows.Next() {
		rowData, err := fetchOneRow(dbrows, colTypes)
		if err != nil {
			return nil, err
		}
		result = append(result, rowData)
	}
	return result, dbrows.Err()
}

func toIntIfInteger(v interface{}) (int64, error) {
	if v == nil {
		return 0, errors.New("input is nil")
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(rv.Uint()), nil
	default:
		fmt.Printf("[DEBUG] - toIntIfInteger: %#v(%T)\n", v, v)
		return 0, errors.New("input is not integer")
	}
}

func toFloatIfReal(v interface{}) (float64, error) {
	if v == nil {
		return 0, errors.New("input is nil")
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	default:
		fmt.Printf("[DEBUG] toFloatIfReal: %#v(%T)\n", v, v)
		return 0, errors.New("input is not real number")
	}
}

func toFloatIfNumber(v interface{}) (float64, error) {
	if v == nil {
		return 0, errors.New("input is nil")
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(rv.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(rv.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return rv.Float(), nil
	default:
		fmt.Printf("[DEBUG] toFloatIfNumber: %#v(%T)\n", v, v)
		return 0, errors.New("input is not valid number")
	}
}

const (
	timezoneSql  = "Asia/Kabul"
	timezoneSql2 = "Europe/Rome"
)

func TestPing(t *testing.T) {
	testName := "TestPing"
	db, err := newDbConnection()
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}
	if db == nil {
		t.Skipf("%s skipped", testName)
	}
	defer func() { _ = db.Close() }()

	err = db.Ping()
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}
}
