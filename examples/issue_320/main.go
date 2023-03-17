package main

import (
	"context"
	"database/sql"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

type Mat struct {
	Id       sql.NullString
	Response go_ora.Clob
}

func MatCol(colname string, mat *Mat) interface{} {
	switch colname {
	case "ID":
		return &mat.Id
	case "RESPONSE":
		return &mat.Response
	default:
		return new(string)
	}
}

func read(conn *sql.DB) error {
	rows, err := conn.Query("SELECT ID, RESPONSE FROM TEMP_TABLE_320")
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("Can't close rows: ", err)
		}
	}()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	for rows.Next() {
		mat := Mat{}
		values := make([]interface{}, len(columns))
		for i, v := range columns {
			values[i] = MatCol(v, &mat)
		}
		err = rows.Scan(values...)
		if err != nil {
			return err
		}
		fmt.Println(values)
	}
	return nil
}

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE TEMP_TABLE_320(
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
	_, err := conn.Exec("drop table TEMP_TABLE_320 purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

func insertData(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("INSERT INTO TEMP_TABLE_320(ID, RESPONSE) VALUES('1', 'THIS IS A TEST')")
	if err != nil {
		return err
	}
	fmt.Println("Finish insert one row: ", time.Now().Sub(t))
	return nil
}

func createRefCursorProc(conn *sql.DB) error {
	sqlText := `CREATE OR REPLACE FUNCTION TEMP_PROC_320 RETURN SYS_REFCURSOR AS
    L_CURSOR SYS_REFCURSOR;
BEGIN
	OPEN L_CURSOR FOR SELECT ID, RESPONSE FROM TEMP_TABLE_320;
	return L_CURSOR;
END TEMP_PROC_320;`
	t := time.Now()
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create refcursor procedure: ", time.Now().Sub(t))
	return nil
}

func dropRefCursorProc(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("DROP FUNCTION TEMP_PROC_320")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop refcursor procedure: ", time.Now().Sub(t))
	return nil
}

func readWithRefCursor(conn *sql.DB) error {
	var cursor go_ora.RefCursor
	var err error
	_, err = conn.ExecContext(context.Background(), `BEGIN :1 := TEMP_PROC_320(); END;`, sql.Out{Dest: &cursor})
	if err != nil {
		return err
	}
	defer func() {
		err = cursor.Close()
		if err != nil {
			fmt.Println("can't close cursor: ", err)
		}
	}()
	rows, err := cursor.Query()
	if err != nil {
		return err
	}
	defer func() {
		err = rows.Close()
		if err != nil {
			fmt.Println("can't close rows: ", err)
		}
	}()
	columns := rows.Columns()

	for rows.Next_() {
		mat := Mat{}
		values := make([]interface{}, len(columns))
		for i, v := range columns {
			values[i] = MatCol(v, &mat)
		}
		err = rows.Scan(values...)
		if err != nil {
			return err
		}
		fmt.Println(values)
	}
	return nil
}
func main() {
	conn, err := sql.Open("oracle", os.Getenv("DSN"))
	if err != nil {
		fmt.Println("Can't connect: ", err)
		return
	}
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
	err = createRefCursorProc(conn)
	if err != nil {
		fmt.Println("can't create refcursor proc: ", err)
		return
	}
	defer func() {
		err = dropRefCursorProc(conn)
		if err != nil {
			fmt.Println("can't drop refcursor proc: ", err)
		}
	}()
	err = insertData(conn)
	if err != nil {
		fmt.Println("can't insert data: ", err)
		return
	}
	err = read(conn)
	if err != nil {
		fmt.Println("can't read data: ", err)
		return
	}

	err = readWithRefCursor(conn)
	if err != nil {
		fmt.Println("can't read data with refcursor: ", err)
		return
	}
}
