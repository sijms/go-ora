package v2_test

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/sijms/go-ora/v3/types"
)

// TestColumnValues reads the values v3 read wrong or not at all: an interval
// as a date (#785), an XMLTYPE, a JSON value as text, past its prefetch, and
// NULL, a declared VECTOR column of each format, and NUMBER's smallest value.
// VECTOR needs Oracle 23ai.
func TestColumnValues(t *testing.T) {
	db, err := newDbConnection()
	if err != nil {
		t.Fatal(err)
	}
	if db == nil {
		t.Skip("ORACLE_DSN is not set")
	}
	defer func() {
		_ = db.Close()
	}()
	exec := func(statements ...string) error {
		for _, statement := range statements {
			if _, err := db.Exec(statement); err != nil {
				return fmt.Errorf("%s: %w", statement, err)
			}
		}
		return nil
	}
	_ = exec("DROP TABLE GOORA_T_VALUES PURGE")
	err = exec(`CREATE TABLE GOORA_T_VALUES (
		id NUMBER, ym INTERVAL YEAR TO MONTH, ds INTERVAL DAY(2) TO SECOND(9), x XMLTYPE, j JSON,
		f32 VECTOR(3, FLOAT32), f64 VECTOR(2, FLOAT64), i8 VECTOR(3, INT8), bin VECTOR(16, BINARY),
		sp VECTOR(5, FLOAT32, SPARSE), tiny NUMBER)`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = exec("DROP TABLE GOORA_T_VALUES PURGE")
	}()
	err = exec(`INSERT INTO GOORA_T_VALUES VALUES (1, INTERVAL '-2-3' YEAR TO MONTH,
		INTERVAL '4 05:06:07.123456789' DAY(2) TO SECOND(9), XMLTYPE('<a>ü</a>'), JSON('{"b":[1,2.5,null],"a":"x"}'),
		'[1,2.5,3]', '[0.1,-1e300]', '[1,-2,127]', '[255,0]', '[5,[1,3],[1.5,2.5]]', 1e-130)`,
		"INSERT INTO GOORA_T_VALUES (id) VALUES (2)")
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]struct {
		column string
		want   interface{}
	}{
		"year to month":    {"ym", "-02-03"},
		"day to second":    {"ds", "+04 05:06:07.123456789"},
		"xml":              {"x", "<a>ü</a>\n"},
		"json":             {"j", `{"b":[1,2.5,null],"a":"x"}`},
		"float32 vector":   {"f32", []float32{1, 2.5, 3}},
		"float64 vector":   {"f64", []float64{0.1, -1e300}},
		"int8 vector":      {"i8", []int8{1, -2, 127}},
		"binary vector":    {"bin", []uint8{255, 0}},
		"sparse vector":    {"sp", types.SparseVector{Dimensions: 5, Indices: []uint32{1, 3}, Values: []float32{1.5, 2.5}}},
		"number's tiniest": {"tiny", "0." + strings.Repeat("0", 129) + "1"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			rows, err := db.Query("SELECT " + c.column + " FROM GOORA_T_VALUES ORDER BY id")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				_ = rows.Close()
			}()
			var got []interface{}
			for rows.Next() {
				var value interface{}
				if err := rows.Scan(&value); err != nil {
					t.Fatal(err)
				}
				if j, ok := value.(*types.Json); ok {
					value = j.String()
				}
				got = append(got, value)
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if len(got) != 2 || !reflect.DeepEqual(got[0], c.want) || got[1] != nil {
				t.Errorf("got %#v, want %#v and NULL", got, c.want)
			}
		})
	}

	// A large XMLTYPE is sent as a CLOB, and JSON and a VECTOR past the
	// prefetch are read in full.
	large := map[string]struct {
		query string
		size  int
	}{
		"xml":    {"SELECT XMLTYPE(TO_CLOB('<a>') || RPAD(TO_CLOB('x'), 40000, 'y') || TO_CLOB('</a>')) FROM dual", 40007},
		"json":   {"SELECT JSON_ARRAYAGG(RPAD('x', 100, 'y') RETURNING JSON) FROM dual CONNECT BY level <= 1000", 103001},
		"vector": {"SELECT TO_VECTOR(JSON_ARRAYAGG(level RETURNING CLOB), 8192, FLOAT64) FROM dual CONNECT BY level <= 8192", 8192},
	}
	for name, c := range large {
		t.Run("large "+name, func(t *testing.T) {
			var value interface{}
			if err := db.QueryRow(c.query).Scan(&value); err != nil {
				t.Fatal(err)
			}
			var size int
			switch value := value.(type) {
			case string:
				size = len(value)
			case *types.Json:
				size = len(value.String())
			case []float64:
				size = len(value)
				if value[size-1] != 8192 {
					t.Errorf("the last dimension is %v", value[size-1])
				}
			default:
				t.Fatalf("a %T", value)
			}
			if size != c.size {
				t.Errorf("size %d, want %d", size, c.size)
			}
		})
	}

	var text sql.NullString
	if err := db.QueryRow("SELECT JSON_SERIALIZE(j) FROM GOORA_T_VALUES WHERE id = 1").Scan(&text); err != nil {
		t.Fatal(err)
	}
	var j types.Json
	if err := db.QueryRow("SELECT j FROM GOORA_T_VALUES WHERE id = 1").Scan(&j); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprint(&j); got != text.String {
		t.Errorf("JSON text %s, JSON_SERIALIZE %s", got, text.String)
	}
}
