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
    DATA        TIMESTAMP WITH TIME ZONE
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

func insertData(conn *sql.DB, loc string) error {
	data := time.Now()
	if len(loc) > 0 {
		zoneLoc, err := time.LoadLocation(loc)
		if err != nil {
			return err
		}
		data = data.In(zoneLoc)
	}

	t := time.Now()
	_, err := conn.Exec(`INSERT INTO TEMP_TABLE_322(ID, DATA) VALUES (:1 , :2)`,
		1, go_ora.TimeStampTZ(data))
	if err != nil {
		return err
	}
	fmt.Println("Finish insert data: ", time.Now().Sub(t))
	return nil
}

func queryOutputPar(conn *sql.DB) error {
	t := time.Now()
	//var data sql.NullTime
	var data go_ora.NullTimeStampTZ
	//var data go_ora.TimeStampTZ
	_, err := conn.Exec(`BEGIN SELECT DATA INTO :1 FROM TEMP_TABLE_322 WHERE ID=1; END;`, go_ora.Out{Dest: &data})
	if err != nil {
		return err
	}
	//fmt.Println(data, "\tLocation: ", data.Time.Location())
	//fmt.Println(time.Time(data), "\tLocation: ", time.Time(data).Location())
	fmt.Println(time.Time(data.TimeStampTZ), "\tLocation: ", time.Time(data.TimeStampTZ).Location())
	fmt.Println("Finish query data: ", time.Now().Sub(t))
	return nil
}
func queryData(conn *sql.DB) error {
	t := time.Now()
	var data time.Time
	rows, err := conn.Query(`SELECT DATA FROM TEMP_TABLE_322`)
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("can't close rows: ", err)
		}
	}()
	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			return err
		}
		fmt.Println(data, "\tLocation: ", data.Location())
	}
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
	err = insertData(conn, "Asia/Kolkata")
	//err = insertData(conn, "")
	if err != nil {
		fmt.Println("can't insert data: ", err)
		return
	}
	err = queryData(conn)
	if err != nil {
		fmt.Println("can't query data: ", err)
		return
	}
	err = queryOutputPar(conn)
	if err != nil {
		fmt.Println("can't query output: ", err)
		return
	}
}
