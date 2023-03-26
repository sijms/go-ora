package main

import (
	"database/sql"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"os"
	"strings"
	"time"
)

var longText = strings.Repeat("*", 0x3FFF)

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE GOORA_TEMP_LONG(
	ID	number(10)	NOT NULL,
	NAME		VARCHAR(200),
	NAME_LONG   LONG,
	VAL			number(10,2),
	VISIT_DATE	date,
	PRIMARY KEY(ID)
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table :", time.Now().Sub(t))
	return nil
}

func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop table GOORA_TEMP_LONG purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

func insertData(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec(`INSERT INTO GOORA_TEMP_LONG(ID, NAME, NAME_LONG, VAL, VISIT_DATE) VALUES(:1, :2, :3, :4, :5)`,
		1, "test name", longText, 1.1, time.Now())
	if err != nil {
		return err
	}
	fmt.Println("Finish insert one row: ", time.Now().Sub(t))
	return nil

}

func queryData(conn *sql.DB) error {
	t := time.Now()
	var (
		name, nameLong string
		val            float64
		date           time.Time
	)
	err := conn.QueryRow("SELECT NAME, NAME_LONG, VAL, VISIT_DATE FROM GOORA_TEMP_LONG WHERE ID = 1").Scan(&name,
		&nameLong, &val, &date)
	if err != nil {
		return err
	}
	fmt.Println("name: ", name)
	fmt.Println("val: ", val)
	fmt.Println("date: ", date)
	if longText == nameLong {
		fmt.Println("long text equal")
	} else {
		fmt.Println("long text not equal")
	}
	fmt.Println("Finish query: ", time.Now().Sub(t))
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
		fmt.Println("can't query: ", err)
	}
}
