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
	sqlText := `CREATE TABLE TEMP_TABLE_329(
    ID          NUMBER(10),
    NAME VARCHAR2(100),
    TEAM_NAME VARCHAR2(100),
    ONBOARD_DATE DATE
	)`
	// WITH TIME ZONE
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table: ", time.Now().Sub(t))
	return nil
}
func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop table TEMP_TABLE_329 purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

func insertData(conn *sql.DB) error {
	t := time.Now()
	var (
		id       = 1
		name     = "test"
		teamName = "team"
		date     = time.Now()
		ret      string
		ret2     string
	)
	_, err := conn.Exec(`INSERT INTO TEMP_TABLE_329(ID, NAME, TEAM_NAME, ONBOARD_DATE)
VALUES(:1, :2, :3, :4) RETURNING NAME, TEAM_NAME INTO :5, :6`, id, name, teamName, date,
		go_ora.Out{Dest: &ret, Size: 100}, go_ora.Out{Dest: &ret2, Size: 100})
	if err != nil {
		return err
	}
	fmt.Println("return: ", ret, "\t", ret2)
	fmt.Println("Finish insert: ", time.Now().Sub(t))
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
		fmt.Println("can't insert: ", err)
	}
}
