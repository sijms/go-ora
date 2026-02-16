package TestIssues

import (
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v3"
	"reflect"
	"testing"
)

func TestVector(t *testing.T) {
	var tableName = "TTB_653"

	var createTable = func(db *sql.DB) error {
		return execCmd(db, fmt.Sprintf(`CREATE TABLE %s(
    ID NUMBER(10) NOT NULL,
    V01 VECTOR(3, INT8),
    V02 VECTOR(3, FLOAT32),
    V03 VECTOR(3, FLOAT64),
    PRIMARY KEY(ID)
    )`, tableName))
	}

	var dropTable = func(db *sql.DB) error {
		return execCmd(db, fmt.Sprintf(`DROP TABLE %s PURGE`, tableName))
	}

	var insert = func(db *sql.DB) error {
		v1, err := go_ora.NewVector([]uint8{10, 20, 30})
		if err != nil {
			return err
		}
		v2, err := go_ora.NewVector([]float32{-10.1, -20.2, -30.3})
		if err != nil {
			return err
		}
		v3, err := go_ora.NewVector([]float64{10.1, 20.2, 30.3})
		if err != nil {
			return err
		}
		_, err = db.Exec(fmt.Sprintf("INSERT INTO %s(ID, V01, V02, V03) VALUES(1, :1, :2, :3)", tableName),
			v1, v2, v3)
		if err != nil {
			return err
		}
		return nil
	}

	var do_check = func(v1 []uint8, v2 []float32, v3 []float64) error {
		var exp_v1 = []uint8{10, 20, 30}
		var exp_v2 = []float32{-10.1, -20.2, -30.3}
		var exp_v3 = []float64{10.1, 20.2, 30.3}
		if !reflect.DeepEqual(v1, exp_v1) {
			return fmt.Errorf("expected: %v Got: %v", exp_v1, v1)
		}
		if !reflect.DeepEqual(v2, exp_v2) {
			return fmt.Errorf("expected: %v Got: %v", exp_v2, v2)
		}
		if !reflect.DeepEqual(v3, exp_v3) {
			return fmt.Errorf("expected: %v Got: %v", exp_v3, v3)
		}
		return nil
	}

	var queryAsVector = func(db *sql.DB) error {
		var data1, data2, data3 go_ora.Vector
		err := db.QueryRow(fmt.Sprintf("SELECT V01, V02, V03 FROM %s", tableName)).Scan(&data1, &data2, &data3)
		if err != nil {
			return err
		}
		var (
			v1 []uint8
			v2 []float32
			v3 []float64
		)
		v1, _ = data1.Data.([]uint8)
		v2, _ = data2.Data.([]float32)
		v3, _ = data3.Data.([]float64)
		return do_check(v1, v2, v3)
	}

	var queryAsArray = func(db *sql.DB) error {
		var (
			v1 []uint8
			v2 []float32
			v3 []float64
		)
		err := db.QueryRow(fmt.Sprintf("SELECT V01, V02, V03 FROM %s", tableName)).Scan(&v1, &v2, &v3)
		if err != nil {
			return err
		}
		return do_check(v1, v2, v3)
	}

	var outputAsVector = func(db *sql.DB) error {
		var data1, data2, data3 go_ora.Vector
		_, err := db.Exec(fmt.Sprintf("BEGIN SELECT V01, V02, V03 INTO :1, :2, :3 FROM %s; END;", tableName),
			go_ora.Out{Dest: &data1},
			go_ora.Out{Dest: &data2},
			go_ora.Out{Dest: &data3},
		)
		if err != nil {
			return err
		}
		var (
			v1 []uint8
			v2 []float32
			v3 []float64
		)
		v1, _ = data1.Data.([]uint8)
		v2, _ = data2.Data.([]float32)
		v3, _ = data3.Data.([]float64)
		return do_check(v1, v2, v3)
	}

	db, err := getDB()
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			t.Error(err)
		}
	}()

	err = createTable(db)
	if err != nil {
		t.Error(err)
		return
	}
	defer func() {
		err = dropTable(db)
		if err != nil {
			t.Error(err)
		}
	}()
	err = insert(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = queryAsVector(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = queryAsArray(db)
	if err != nil {
		t.Error(err)
		return
	}
	err = outputAsVector(db)
	if err != nil {
		t.Error(err)
		return
	}
}
