package v2_test

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"testing"
)

var sqlColNamesTestDataTypeMoney = []string{"id", "data_money2", "data_money4", "data_money6", "data_money8"}

func TestDataType_Money(t *testing.T) {
	testName := "TestDataType_Money"
	db, err := newDbConnection()
	if err != nil {
		t.Fatalf("%s failed: %s", testName, err)
	}
	if db == nil {
		t.Skipf("%s skipped", testName)
	}
	defer func() { _ = db.Close() }()

	tblName := "test_money"
	colNameList := sqlColNamesTestDataTypeMoney
	colTypes := []string{"NVARCHAR2(8)", "NUMERIC(24,2)", "DECIMAL(28,4)", "DEC(32,6)", "NUMERIC(36,8)"}
	type Row struct {
		id         string
		dataMoney2 float64
		dataMoney4 float64
		dataMoney6 float64
		dataMoney8 float64
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
		//vMoneySmall := float64(rand.Intn(65536)) + rand.Float64()
		//vMoneyLarge := float64(rand.Int31()) + rand.Float64()
		vMoneySmall := float64(rand.Intn(65536))
		vMoneyLarge := float64(rand.Int31())
		row := Row{
			id:         fmt.Sprintf("%03d", i),
			dataMoney2: math.Round(vMoneySmall*1e2) / 1e2,
			dataMoney4: math.Round(vMoneySmall*1e4) / 1e4,
			dataMoney6: math.Round(vMoneyLarge*1e6) / 1e6,
			dataMoney8: math.Round(vMoneyLarge*1e8) / 1e8,
		}
		rowArr = append(rowArr, row)
		params := []interface{}{row.id, row.dataMoney2, row.dataMoney4, row.dataMoney6, row.dataMoney8}
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
			e := expected.dataMoney2
			f := colNameList[1]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.2f", e), fmt.Sprintf("%.2f", v); err != nil || vstr != estr {
				fmt.Printf("DEBUG: Row #%s / From db table: %#v(%.10f) / Expected: %#v(%.10f)\n", row["id"], row[f], row[f], e, e)
				t.Fatalf("%s failed: [%s] expected %#v/%.10f(%T) but received %#v/%.10f(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, e, vstr, v, row[f], err)
			}
		}
		{
			e := expected.dataMoney4
			f := colNameList[2]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.4f", e), fmt.Sprintf("%.4f", v); err != nil || vstr != estr {
				fmt.Printf("DEBUG: Row #%s / From db table: %#v(%.10f) / Expected: %#v(%.10f)\n", row["id"], row[f], row[f], e, e)
				t.Fatalf("%s failed: [%s] expected %#v/%.10f(%T) but received %#v/%.10f(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, e, vstr, v, row[f], err)
			}
		}
		{
			e := expected.dataMoney6
			f := colNameList[3]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.6f", e), fmt.Sprintf("%.6f", v); err != nil || vstr != estr {
				fmt.Printf("DEBUG: Row #%s / From db table: %#v(%.10f) / Expected: %#v(%.10f)\n", row["id"], row[f], row[f], e, e)
				t.Fatalf("%s failed: [%s] expected %#v/%.10f(%T) but received %#v/%.10f(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, e, vstr, v, row[f], err)
			}
		}
		{
			e := expected.dataMoney8
			f := colNameList[4]
			v, err := toFloatIfReal(row[f])
			if estr, vstr := fmt.Sprintf("%.8f", e), fmt.Sprintf("%.8f", v); err != nil || vstr != estr {
				fmt.Printf("DEBUG: Row #%s / From db table: %#v (%.10f) / Expected: %#v (%.10f)\n", row["id"], row[f], row[f], e, e)
				t.Fatalf("%s failed: [%s] expected %#v/%.10f(%T) but received %#v/%.10f(%T) (error: %s)", testName, row["id"].(string)+"/"+f, estr, e, e, vstr, v, row[f], err)
			}
		}
	}
}
