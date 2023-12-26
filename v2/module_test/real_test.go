package v2_test

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

var sqlColNamesTestDataTypeReal = []string{"id",
	"data_float", "data_double", "data_real",
	"data_decimal", "data_number", "data_numeric",
	"data_float32", "data_float64",
	"data_float4", "data_float8"}

func TestDataType_Real(t *testing.T) {
	testName := "TestDataType_Real"
	db, err := newDbConnection()
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}
	if db == nil {
		t.Skipf("%s skipped", testName)
	}
	defer func() { _ = db.Close() }()

	tblName := "test_real"
	colNameList := sqlColNamesTestDataTypeReal
	colTypes := []string{"NVARCHAR2(8)",
		"FLOAT", "DOUBLE PRECISION", "REAL",
		"DECIMAL(38,6)", "NUMBER(38,6)", "NUMERIC(38,6)",
		"BINARY_FLOAT", "BINARY_DOUBLE",
		"BINARY_FLOAT", "BINARY_DOUBLE"}
	type Row struct {
		id          string
		dataFloat   float64
		dataDouble  float64
		dataReal    float64
		dataDecimal float64
		dataNumber  float64
		dataNumeric float64
		dataFloat32 float32
		dataFloat64 float64
		dataFloat4  float32
		dataFloat8  float64
	}

	// init table
	_, _ = db.Exec(fmt.Sprintf("DROP TABLE %s", tblName))
	sql := fmt.Sprintf("CREATE TABLE %s (", tblName)
	for i := range colNameList {
		sql += colNameList[i] + " " + colTypes[i] + ","
	}
	sql += fmt.Sprintf("PRIMARY KEY(%s))", colNameList[0])
	if _, err := db.Exec(sql); err != nil {
		t.Fatalf("%s failed: %s\n%s", testName, err, sql)
	}

	rowArr := make([]Row, 0)
	numRows := 100
	// insert some rows
	sql = fmt.Sprintf("INSERT INTO %s (", tblName)
	sql += strings.Join(colNameList, ",")
	sql += ") VALUES ("
	sql += generatePlaceholders(len(colNameList)) + ")"
	for i := 1; i <= numRows; i++ {
		vReal := 0.123456789
		row := Row{
			id:          fmt.Sprintf("%03d", i),
			dataFloat:   math.Round(vReal*1e6) / 1e6,
			dataDouble:  math.Round(vReal*1e6) / 1e6,
			dataReal:    math.Round(vReal*1e6) / 1e6,
			dataDecimal: math.Round(vReal*1e6) / 1e6,
			dataNumber:  math.Round(vReal*1e6) / 1e6,
			dataNumeric: math.Round(vReal*1e6) / 1e6,
			dataFloat32: float32(math.Round(vReal*1e4) / 1e4),
			dataFloat64: math.Round(vReal*1e8) / 1e8,
			dataFloat4:  float32(math.Round(vReal*1e4) / 1e4),
			dataFloat8:  math.Round(vReal*1e8) / 1e8,
		}
		rowArr = append(rowArr, row)
		params := []interface{}{row.id, row.dataFloat, row.dataDouble, row.dataReal,
			row.dataDecimal, row.dataNumeric, row.dataNumeric,
			row.dataFloat32, row.dataFloat64, row.dataFloat4, row.dataFloat8}
		_, err := db.Exec(sql, params...)
		if err != nil {
			fmt.Printf("%s\n%#v\n", sql, params)
			t.Fatalf("%s failed: row: %d / error: %s", testName, i, err)
		}
	}

	// query some rows
	sql = fmt.Sprintf("SELECT * FROM %s ORDER BY id", tblName)
	dbRows, err := db.Query(sql)
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}
	rows, err := fetchAllRowsColumnLowerCased(dbRows)
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}

	for i, row := range rows {
		expected := rowArr[i]
		{
			f := "id"
			e := expected.id
			v, ok := row[f].(string)
			if !ok || v != e {
				t.Fatalf("%s failed: [%s] expected %#v but received %#v", testName, f, e, row[f])
			}
		}
		{
			e := expected.dataFloat
			f := colNameList[1]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataDouble
			f := colNameList[2]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataReal
			f := colNameList[3]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataDecimal
			f := colNameList[4]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataNumber
			f := colNameList[5]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataNumeric
			f := colNameList[6]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataFloat32
			f := colNameList[7]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataFloat64
			f := colNameList[8]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataFloat4
			f := colNameList[9]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataFloat8
			f := colNameList[10]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
	}
}

func TestDataType_RealNonFraction(t *testing.T) {
	testName := "TestDataType_RealNonFraction"
	db, err := newDbConnection()
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}
	if db == nil {
		t.Skipf("%s skipped", testName)
	}
	defer func() { _ = db.Close() }()

	tblName := "test_real"
	colNameList := sqlColNamesTestDataTypeReal
	colTypes := []string{"NVARCHAR2(8)",
		"FLOAT", "DOUBLE PRECISION", "REAL",
		"DECIMAL(38,6)", "NUMBER(38,6)", "NUMERIC(38,6)",
		"BINARY_FLOAT", "BINARY_DOUBLE",
		"BINARY_FLOAT", "BINARY_DOUBLE"}
	type Row struct {
		id          string
		dataFloat   float64
		dataDouble  float64
		dataReal    float64
		dataDecimal float64
		dataNumber  float64
		dataNumeric float64
		dataFloat32 float32
		dataFloat64 float64
		dataFloat4  float32
		dataFloat8  float64
	}

	// init table
	_, _ = db.Exec(fmt.Sprintf("DROP TABLE %s", tblName))
	sql := fmt.Sprintf("CREATE TABLE %s (", tblName)
	for i := range colNameList {
		sql += colNameList[i] + " " + colTypes[i] + ","
	}
	sql += fmt.Sprintf("PRIMARY KEY(%s))", colNameList[0])
	if _, err := db.Exec(sql); err != nil {
		t.Fatalf("%s failed: %s\n%s", testName, err, sql)
	}

	rowArr := make([]Row, 0)
	numRows := 100
	// insert some rows
	sql = fmt.Sprintf("INSERT INTO %s (", tblName)
	sql += strings.Join(colNameList, ",")
	sql += ") VALUES ("
	sql += generatePlaceholders(len(colNameList)) + ")"
	for i := 1; i <= numRows; i++ {
		vReal := float64(12345)
		row := Row{
			id:          fmt.Sprintf("%03d", i),
			dataFloat:   math.Round(vReal*1e6) / 1e6,
			dataDouble:  math.Round(vReal*1e6) / 1e6,
			dataReal:    math.Round(vReal*1e6) / 1e6,
			dataDecimal: math.Round(vReal*1e6) / 1e6,
			dataNumber:  math.Round(vReal*1e6) / 1e6,
			dataNumeric: math.Round(vReal*1e6) / 1e6,
			dataFloat32: float32(math.Round(vReal*1e4) / 1e4),
			dataFloat64: math.Round(vReal*1e8) / 1e8,
			dataFloat4:  float32(math.Round(vReal*1e4) / 1e4),
			dataFloat8:  math.Round(vReal*1e8) / 1e8,
		}
		rowArr = append(rowArr, row)
		params := []interface{}{row.id, row.dataFloat, row.dataDouble, row.dataReal,
			row.dataDecimal, row.dataNumeric, row.dataNumeric,
			row.dataFloat32, row.dataFloat64, row.dataFloat4, row.dataFloat8}
		_, err := db.Exec(sql, params...)
		if err != nil {
			fmt.Printf("%s\n%#v\n", sql, params)
			t.Fatalf("%s failed: row: %d / error: %s", testName, i, err)
		}
	}

	// query some rows
	sql = fmt.Sprintf("SELECT * FROM %s ORDER BY id", tblName)
	dbRows, err := db.Query(sql)
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}
	rows, err := fetchAllRowsColumnLowerCased(dbRows)
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}

	for i, row := range rows {
		expected := rowArr[i]
		{
			f := "id"
			e := expected.id
			v, ok := row[f].(string)
			if !ok || v != e {
				t.Fatalf("%s failed: [%s] expected %#v but received %#v", testName, f, e, row[f])
			}
		}
		{
			e := expected.dataFloat
			f := colNameList[1]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataDouble
			f := colNameList[2]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataReal
			f := colNameList[3]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataDecimal
			f := colNameList[4]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataNumber
			f := colNameList[5]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataNumeric
			f := colNameList[6]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataFloat32
			f := colNameList[7]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataFloat64
			f := colNameList[8]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataFloat4
			f := colNameList[9]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
		{
			e := expected.dataFloat8
			f := colNameList[10]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.f", e), fmt.Sprintf("%.f", v); err != nil || vstr != estr {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, vstr, row[f], err)
			}
		}
	}
}
