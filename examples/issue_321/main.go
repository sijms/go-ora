package main

import (
	"database/sql"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE TEMP_TABLE_321(
    ID          varchar2(100),
    RESPONSE    CLOB
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table: ", time.Now().Sub(t))
	return nil
}

func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop table TEMP_TABLE_321 purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

func insertData(conn *sql.Tx) error {
	t := time.Now()
	_, err := conn.Exec("INSERT INTO TEMP_TABLE_321(ID, RESPONSE) VALUES('1', 'THIS IS A TEST')")
	if err != nil {
		return err
	}
	fmt.Println("Finish insert one row: ", time.Now().Sub(t))
	return nil
}

func main() {
	conn, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("Can't connect: ", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("can't close connection: ", err)
		}
	}()
	//err = createTable(conn)
	//if err != nil {
	//	fmt.Println("can't create table: ", err)
	//	return
	//}

	//defer func() {
	//	err = dropTable(conn)
	//	if err != nil {
	//		fmt.Println("can't drop table: ", err)
	//	}
	//}()
	tx, err := conn.Begin()
	if err != nil {
		fmt.Println("can't begin transaction: ", err)
		return
	}
	err = insertData(tx)
	if err != nil {
		fmt.Println("can't insert data: ", err)
		err = tx.Rollback()
		if err != nil {
			fmt.Println("can't rollback: ", err)
		}

		return
	}
}
