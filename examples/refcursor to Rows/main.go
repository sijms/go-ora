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
	sqlText := `CREATE TABLE TEMP_TABLE_316(
	VISIT_ID	number(10)	NOT NULL,
	NAME		VARCHAR(200),
	VAL			number(10,2),
	VISIT_DATE	date,
	PRIMARY KEY(VISIT_ID)
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table GOORA_TEMP_VISIT :", time.Now().Sub(t))
	return nil
}

func insertData(conn *sql.DB) error {
	t := time.Now()
	index := 1
	stmt, err := conn.Prepare(`INSERT INTO TEMP_TABLE_316(VISIT_ID, NAME, VAL, VISIT_DATE) 
VALUES(:1, :2, :3, :4)`)
	if err != nil {
		return err
	}
	defer func() {
		_ = stmt.Close()
	}()
	nameText := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	val := 1.1
	for index = 1; index <= 100; index++ {
		_, err = stmt.Exec(index, nameText, val, time.Now())
		if err != nil {
			return err
		}
		val += 1.1
	}
	fmt.Println("100 rows inserted: ", time.Now().Sub(t))
	return nil
}

func dropTable(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop table TEMP_TABLE_316 purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

func createRefCursorProc(conn *sql.DB) error {
	sqlText := `CREATE OR REPLACE FUNCTION TEMP_FUNC_316
(
	LVISIT_ID IN NUMBER
) RETURN SYS_REFCURSOR AS
    L_CURSOR SYS_REFCURSOR;
BEGIN
	OPEN L_CURSOR FOR SELECT VISIT_ID, NAME, VAL, VISIT_DATE FROM TEMP_TABLE_316 WHERE VISIT_ID > LVISIT_ID ORDER BY VISIT_ID;
	RETURN L_CURSOR;
END TEMP_FUNC_316;`
	t := time.Now()
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create refcursor function: ", time.Now().Sub(t))
	return nil
}

func dropRefCursorProc(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("DROP FUNCTION TEMP_FUNC_316")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop refcursor procedure: ", time.Now().Sub(t))
	return nil
}

func UseCursor(conn *sql.DB) error {
	t := time.Now()
	var cursor sql.Rows
	//sqlText := `BEGIN :1 := GOORA_TEMP_GET_VISIT(:2); END;`
	sqlText := `SELECT TEMP_FUNC_316(10) from dual`
	rows, err := conn.Query(sqlText)
	if err != nil {
		return err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			fmt.Println("can't close rows: ", err)
		}
	}(rows)
	for rows.Next() {
		err = rows.Scan(&cursor)
		if err != nil {
			return err
		}
		var (
			id   int64
			name string
			val  float64
			date time.Time
		)
		for cursor.Next() {
			err = cursor.Scan(&id, &name, &val, &date)
			if err != nil {
				return err
			}
			fmt.Println("ID: ", id, "\tName: ", name, "\tval: ", val, "\tDate: ", date)
		}
	}
	fmt.Println("Finish query cursor: ", time.Now().Sub(t))
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
		fmt.Println("Can't create table", err)
	}
	defer func() {
		err = dropTable(conn)
		if err != nil {
			fmt.Println("Can't drop table", err)
		}
	}()
	err = insertData(conn)
	if err != nil {
		fmt.Println("Can't insert data", err)
		return
	}
	err = createRefCursorProc(conn)
	if err != nil {
		fmt.Println("Can't create RefCursor", err)
		return
	}
	defer func() {
		err = dropRefCursorProc(conn)
		if err != nil {
			fmt.Println("Can't drop RefCursor", err)
		}
	}()
	err = UseCursor(conn)
	if err != nil {
		fmt.Println("can't use cursor: ", err)
		return
	}
}
