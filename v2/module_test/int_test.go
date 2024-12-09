package v2_test

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

var sqlColNamesTestDataTypeInt = []string{
	"id",
	"data_int", "data_integer", "data_decimal", "data_number", "data_numeric",
	"data_tinyint", "data_smallint", "data_mediumint", "data_bigint",
	"data_int1", "data_int2", "data_int4", "data_int8",
}

func TestDataType_Int(t *testing.T) {
	testName := "TestDataType_Int"
	db, err := newDbConnection()
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}
	if db == nil {
		t.Skipf("%s skipped", testName)
	}
	defer func() { _ = db.Close() }()

	tblName := "test_int"
	colNameList := sqlColNamesTestDataTypeInt
	colTypes := []string{
		"NVARCHAR2(8)",
		"INT", "INTEGER", "NUMERIC(38,0)", "NUMBER(38,0)", "DECIMAL(38,0)",
		"NUMERIC(3,0)", "SMALLINT", "DECIMAL(19,0)", "DEC(38,0)",
		"DEC(4,0)", "NUMBER(8,0)", "DECIMAL(16,0)", "NUMERIC(32,0)",
	}
	type Row struct {
		id            string
		dataInt       int
		dataInteger   int
		dataDecimal   int
		dataNumber    int
		dataNumeric   int
		dataTinyInt   int8
		dataSmallInt  int16
		dataMediumInt int32
		dataBigInt    int64
		dataInt1      int8
		dataInt2      int16
		dataInt4      int32
		dataInt8      int64
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
		vInt := rand.Int63()
		row := Row{
			id:            fmt.Sprintf("%03d", i),
			dataInt:       int(vInt%(2^32)) + 1,
			dataInteger:   int(vInt%(2^32)) + 2,
			dataDecimal:   int(vInt%(2^32)) + 3,
			dataNumber:    int(vInt%(2^32)) + 4,
			dataNumeric:   int(vInt%(2^32)) + 5,
			dataTinyInt:   int8(vInt%(2^8)) + 6,
			dataSmallInt:  int16(vInt%(2^16)) + 7,
			dataMediumInt: int32(vInt%(2^24)) + 8,
			dataBigInt:    vInt - 1,
			dataInt1:      int8(vInt%(2^8)) + 9,
			dataInt2:      int16(vInt%(2^16)) + 10,
			dataInt4:      int32(vInt%(2^24)) + 11,
			dataInt8:      vInt - 2,
		}
		rowArr = append(rowArr, row)
		params := []interface{}{
			row.id,
			row.dataInt, row.dataInteger, row.dataDecimal, row.dataNumber, row.dataNumeric,
			row.dataTinyInt, row.dataSmallInt, row.dataMediumInt, row.dataBigInt,
			row.dataInt1, row.dataInt2, row.dataInt4, row.dataInt8,
		}
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
			e := int64(expected.dataInt)
			f := colNameList[1]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
		{
			e := int64(expected.dataInteger)
			f := colNameList[2]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
		{
			e := int64(expected.dataDecimal)
			f := colNameList[3]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
		{
			e := int64(expected.dataNumber)
			f := colNameList[4]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
		{
			e := int64(expected.dataNumeric)
			f := colNameList[5]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
		{
			e := int64(expected.dataTinyInt)
			f := colNameList[6]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
		{
			e := int64(expected.dataSmallInt)
			f := colNameList[7]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
		{
			e := int64(expected.dataMediumInt)
			f := colNameList[8]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
		{
			e := int64(expected.dataBigInt)
			f := colNameList[9]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
		{
			e := int64(expected.dataInt1)
			f := colNameList[10]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
		{
			e := int64(expected.dataInt2)
			f := colNameList[11]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
		{
			e := int64(expected.dataInt4)
			f := colNameList[12]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
		{
			e := int64(expected.dataInt8)
			f := colNameList[13]
			v, err := toIntIfInteger(row[f])
			if err != nil || v != e {
				t.Fatalf("%s failed: [%s] expected %#v(%T) but received %#v(%T) (error: %s)", testName, row["id"].(string)+"/"+f, e, e, row[f], row[f], err)
			}
		}
	}
}
