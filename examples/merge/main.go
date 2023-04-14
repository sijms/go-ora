package main

import (
	"database/sql"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"os"
	"strings"
	"time"
)

func _print(prefix string, val interface{}) {
	fmt.Println(prefix, val)
}
func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE TEMP_TABLE_343(
	ID	number(10)	NOT NULL,
	NAME		VARCHAR(500),
	VAL			number(10,2),
	LDATE   		date,
	PRIMARY KEY(ID)
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	_print("finish create table: ", time.Now().Sub(t))
	return nil
}

func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop table TEMP_TABLE_343 purge")
	if err != nil {
		return err
	}
	_print("finish drop table: ", time.Now().Sub(t))
	return nil
}

func insert(conn *sql.DB) error {
	t := time.Now()
	sqlText := `INSERT INTO TEMP_TABLE_343(ID, NAME, VAL, LDATE) VALUES(:ID, :NAME, :VAL, :LDATE)`
	length := 500
	ids := make([]int, length)
	names := make([]string, length+1)
	vals := make([]float32, length+2)
	dates := make([]time.Time, length)
	for x := 0; x < length; x++ {
		ids[x] = x + 1
		names[x] = strings.Repeat("*", x+1)
		vals[x] = float32(length) / float32(x+1)
		dates[x] = time.Now()
	}
	_, err := conn.Exec(sqlText, sql.Named("ID", ids),
		sql.Named("NAME", names),
		sql.Named("VAL", vals),
		sql.Named("LDATE", dates))
	if err != nil {
		return err
	}
	_print("finish insert: ", time.Now().Sub(t))
	return nil
}
func merge(conn *sql.DB) error {
	t := time.Now()
	sqlText := `MERGE INTO TEMP_TABLE_343 t1 USING(select :ID ID,:NAME NAME,:VAL VAL,:LDATE LDATE from dual) tmp  
ON (tmp.ID=t1.ID)  
	WHEN MATCHED THEN UPDATE SET NAME=:NAME, VAL=:VAL, LDATE=:LDATE 
	WHEN NOT MATCHED THEN INSERT (ID, NAME, VAL, LDATE) VALUES (:ID, :NAME, :VAL , :LDATE)`
	length := 500
	ids := make([]int, length)
	names := make([]string, length+1)
	vals := make([]float32, length+2)
	dates := make([]time.Time, length)
	for x := 0; x < length; x++ {
		ids[x] = x + 1
		names[x] = strings.Repeat("+", x+1)
		vals[x] = float32(length) / float32(x+1)
		dates[x] = time.Now()
	}
	_, err := conn.Exec(sqlText, sql.Named("ID", ids),
		sql.Named("NAME", names),
		sql.Named("VAL", vals),
		sql.Named("LDATE", dates))
	if err != nil {
		return err
	}
	_print("finish merge: ", time.Now().Sub(t))
	return nil
}
func query(conn *sql.DB) error {
	t := time.Now()
	rows, err := conn.Query(`SELECT ID, NAME, VAL, LDATE FROM TEMP_TABLE_343 WHERE ID <= 10 ORDER BY ID`)
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("can't close rows: ", err)
		}
	}()
	var (
		id   int
		name string
		val  float32
		date time.Time
	)
	for rows.Next() {
		err = rows.Scan(&id, &name, &val, &date)
		if err != nil {
			return err
		}
		fmt.Println("ID: ", id, "\tName: ", name, "\tVal: ", val, "\tDate: ", date)
	}
	fmt.Println("finish query: ", time.Now().Sub(t))
	return nil
}

func main() {
	conn, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		_print("can't open connection: ", err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			_print("can't close connection: ", err)
			return
		}
	}()
	err = createTable(conn)
	if err != nil {
		_print("can't create table: ", err)
		return
	}
	defer func() {
		err = dropTable(conn)
		if err != nil {
			_print("can't drop table: ", err)
		}
	}()
	err = insert(conn)
	if err != nil {
		_print("can't insert: ", err)
		return
	}
	err = query(conn)
	if err != nil {
		fmt.Println("can't query: ", err)
		return
	}
	err = merge(conn)
	if err != nil {
		_print("can't merge: ", err)
		return
	}
	err = query(conn)
	if err != nil {
		fmt.Println("can't query: ", err)
		return
	}
}
