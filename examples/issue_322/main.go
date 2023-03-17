package main

import (
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE TEMP_TABLE_322(
    ID          NUMBER(10),
    DATA        TIMESTAMP
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
	_, err := conn.Exec("drop table TEMP_TABLE_322 purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

func insertData(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec(`INSERT INTO TEMP_TABLE_322(ID, DATA) VALUES (:1 , :2)`,
		1, go_ora.TimeStamp(time.Now()))
	if err != nil {
		return err
	}
	fmt.Println("Finish insert data: ", time.Now().Sub(t))
	return nil
}

func queryData(conn *sql.DB) error {
	t := time.Now()
	var data time.Time
	err := conn.QueryRow(`SELECT DATA FROM TEMP_TABLE_322 WHERE ID=1`).Scan(&data)
	if err != nil {
		return err
	}
	fmt.Println(data)
	fmt.Println("Finish query data: ", time.Now().Sub(t))
	return nil
}
func main() {
	conn, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("can't open connection: ", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("can't close connection: ", err)
		}
	}()

	err = createTable(conn)
	if err != nil {
		fmt.Println("can't create table: ", err)
		return
	}

	defer func() {
		err = dropTable(conn)
		if err != nil {
			fmt.Println("can't drop table: ", err)
		}
	}()
	err = insertData(conn)
	if err != nil {
		fmt.Println("can't insert data: ", err)
		return
	}
	err = queryData(conn)
	if err != nil {
		fmt.Println("can't query data: ", err)
		return
	}
}
