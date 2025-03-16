package main

import (
	"database/sql"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

func execCmd(db *sql.DB, stmts ...string) error {
	for _, stmt := range stmts {
		if _, err := db.Exec(stmt); err != nil {
			if len(stmts) > 1 {
				return fmt.Errorf("error: %v in execuation of stmt: %s", err, stmt)
			} else {
				return err
			}
		}
	}
	return nil
}

var tableName = "TTB_653"

func createTable(db *sql.DB) error {
	t := time.Now()
	err := execCmd(db, fmt.Sprintf(`CREATE TABLE %s(
	ID	number(10)	NOT NULL,
	v01 VECTOR(3, INT8),
    v02 VECTOR(3, FLOAT32),
    v03 VECTOR(3, FLOAT64),
	PRIMARY KEY(ID)
	)`, tableName))
	if err != nil {
		return err
	}
	fmt.Println("Finish create table :", time.Since(t))
	return nil
}

func dropTable(db *sql.DB) error {
	t := time.Now()
	err := execCmd(db, fmt.Sprintf(`DROP TABLE %s purge`, tableName))
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table :", time.Since(t))
	return nil
}

func insert(db *sql.DB) error {
	t := time.Now()
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
	_, err = db.Exec(fmt.Sprintf("INSERT INTO %s(id, v01, v02, v03) VALUES(1, :1, :2, :3)", tableName), v1, v2, v3)
	if err != nil {
		return err
	}
	fmt.Println("Finish insert :", time.Since(t))
	return nil
}
func queryAsVector(db *sql.DB) error {
	t := time.Now()
	var data1, data2, data3 go_ora.Vector
	rows, err := db.Query(fmt.Sprintf("SELECT v01, v02, v03 FROM %s", tableName))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&data1, &data2, &data3)
		if err != nil {
			return err
		}
		fmt.Println(data1, data2, data3)
	}
	fmt.Println("Finish query :", time.Since(t))
	return nil
}

func queryAsArray(db *sql.DB) error {
	t := time.Now()
	var (
		data1 []uint8
		data2 []float32
		data3 []float64
	)
	rows, err := db.Query(fmt.Sprintf("SELECT v01, v02, v03 FROM %s", tableName))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&data1, &data2, &data3)
		if err != nil {
			return err
		}
		fmt.Println(data1, data2, data3)
	}
	fmt.Println("Finish query :", time.Since(t))
	return nil
}

func outputVector(db *sql.DB) error {
	t := time.Now()
	var data1, data2, data3 go_ora.Vector
	_, err := db.Exec(fmt.Sprintf("BEGIN SELECT v01, v02, v03 INTO :1, :2, :3 FROM %s WHERE ID=:4; END;", tableName),
		go_ora.Out{Dest: &data1},
		go_ora.Out{Dest: &data2},
		go_ora.Out{Dest: &data3},
		1)
	if err != nil {
		return err
	}
	fmt.Println(data1.Data, data2.Data, data3.Data)
	fmt.Println("Finish Output Vector :", time.Since(t))
	return nil
}

func main() {
	db, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("can't connect: ", err)
		return
	}
	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("can't close db:", err)
			return
		}
	}()
	err = createTable(db)
	if err != nil {
		fmt.Println("can't create table:", err)
		return
	}
	defer func() {
		err = dropTable(db)
		if err != nil {
			fmt.Println("can't drop table:", err)
		}
	}()
	err = insert(db)
	if err != nil {
		fmt.Println("can't insert data:", err)
		return
	}
	err = queryAsVector(db)
	if err != nil {
		fmt.Println("can't query as vector:", err)
		return
	}
	err = queryAsArray(db)
	if err != nil {
		fmt.Println("can't query as array:", err)
		return
	}
	err = outputVector(db)
	if err != nil {
		fmt.Println("can't output vector:", err)
		return
	}
}
