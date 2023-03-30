package main

import (
	"database/sql"
	"fmt"
	"time"

	go_ora "github.com/sijms/go-ora/v2"
)

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE TEMP_TABLE_333(
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
	_, err := conn.Exec("drop table TEMP_TABLE_333 purge")
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
	_, err := conn.Exec(`INSERT INTO TEMP_TABLE_333(ID, NAME, TEAM_NAME, ONBOARD_DATE)
VALUES(:1, :2, :3, :4) RETURNING NAME, TEAM_NAME INTO :5, :6`, id, name, teamName, date,
		go_ora.Out{Dest: &ret, Size: 100}, go_ora.Out{Dest: &ret2, Size: 100})
	if err != nil {
		return err
	}
	fmt.Println("return: ", ret, "\t", ret2)
	fmt.Println("Finish insert: ", time.Now().Sub(t))
	return nil
}

func queryData(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Query(`SELECT * FROM TEMP_TABLE_333`)

	if err != nil {
		return err
	}
	fmt.Println("Finish query: ", time.Now().Sub(t))
	return nil
}

func main() {
	urlOptions := map[string]string{
		"FAILOVER":   "5",
		"RETRYTIME":  "10",
		"TRACE FILE": "trace.log",
	}

	databaseUrl := go_ora.BuildUrl("localhost", 1521, "ORCLCDB.localdomain", "db_user1", "db_user_pass", urlOptions)
	fmt.Println("connection string: ", databaseUrl)

	conn, err := sql.Open("oracle", databaseUrl)
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

	err = insertData(conn)
	if err != nil {
		fmt.Println("can't insert: ", err)
	}

	err = queryData(conn)
	if err != nil {
		fmt.Println("can't query table: ", err)
	}

	fmt.Println("waiting for restart oracledb - first attempt")
	time.Sleep(1 * time.Minute) // Wait for the connection reconnect

	err = queryData(conn)
	if err != nil {
		fmt.Println("can't query table: ", err)
	}

	fmt.Println("waiting for restart oracledb - second attempt")
	time.Sleep(90 * time.Second)
	err = queryData(conn)
	if err != nil {
		fmt.Println("can't query table: ", err)
	}

	defer func() {
		err = dropTable(conn)
		if err != nil {
			fmt.Println("can't drop table: ", err)
		}
	}()
}
