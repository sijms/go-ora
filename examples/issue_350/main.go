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
	sqlText := `CREATE TABLE TEMP_TABLE_350(
	ID	number(10)	NOT NULL,
	NAME		VARCHAR2(500),
	NAME2       NVARCHAR2(500),
	VAL			number(10,2),
	VAL2        number(10),
	LDATE   		date,
	LDATE2          TIMESTAMP,
	LDATE3          TIMESTAMP WITH TIME ZONE,
	PRIMARY KEY(ID)
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("finish create table: ", time.Now().Sub(t))
	return nil
}

func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop table TEMP_TABLE_350 purge")
	if err != nil {
		return err
	}
	fmt.Println("finish drop table: ", time.Now().Sub(t))
	return nil
}

func query2(conn *sql.DB) error {
	rows, err := conn.Query(`select OWNER, TABLE_NAME, COLUMN_NAME, DATA_TYPE, COLUMN_ID, DATA_DEFAULT, NULLABLE,
       COLLATION from ALL_TAB_COLUMNS WHERE TABLE_NAME = 'TEMP_TABLE_35'`)
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
		schemaName, tableName, columnName, columnType, defaultStr, nullable, collation sql.NullString
		columnPosition                                                                 int
	)
	for rows.Next() {
		err = rows.Scan(&schemaName, &tableName, &columnName, &columnType, &columnPosition, &defaultStr, &nullable, &collation)
		if err != nil {
			return err
		}
	}
	return nil
}
func query(conn *sql.DB) error {
	rows, err := conn.Query(`SELECT ID FROM TEMP_TABLE_350`)
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("can't close rows: ", err)
		}
	}()
	var id int
	for rows.Next() {
		err = rows.Scan(&id)
		if err != nil {
			return err
		}
		fmt.Println(id)
	}
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
			return
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
	err = query2(conn)
	if err != nil {
		fmt.Println("can't query: ", err)
		return
	}
}
