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
	sqlText := `CREATE TABLE TEMP_TABLE_330(
    ID          NUMBER(10),
    NAME VARCHAR2(100),
    TEAM_NAME VARCHAR2(100),
    int1		number(10),
    text1		varchar2(100),
    ONBOARD_DATE DATE
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
	_, err := conn.Exec("drop table TEMP_TABLE_330 purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

func insertData(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec(`INSERT INTO TEMP_TABLE_330(ID, NAME, ONBOARD_DATE, TEAM_NAME, int1, text1)
VALUES(:p1, :TEXT, :p2, :TEXT, :p1, :TEXT)`, sql.Named("p1", 1), sql.Named("p2", time.Now()),
		sql.Named("TEXT", "test"))
	if err != nil {
		return err
	}
	fmt.Println("Finish insert data: ", time.Now().Sub(t))
	return nil
}

func queryData(conn *sql.DB) error {
	t := time.Now()
	var (
		name, team, text1 string
		number            int64
		date              time.Time
	)
	err := conn.QueryRow(`SELECT NAME, ONBOARD_DATE, TEAM_NAME, int1, text1 FROM TEMP_TABLE_330 WHERE ID=1`).Scan(
		&name, &date, &team, &number, &text1)
	if err != nil {
		return err
	}
	fmt.Println("Name: ", name, "\tDate: ", date, "\tTeam: ", team, "\tNumber: ", number, "\tText: ", text1)
	fmt.Println("Finish query data: ", time.Now().Sub(t))
	return nil
}
func main() {
	conn, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("can't connect: ", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("can't close: ", err)
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
