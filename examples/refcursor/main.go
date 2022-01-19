package main

import (
	"database/sql"
	"flag"
	"fmt"

	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE GOORA_TEMP_VISIT(
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
	stmt, err := conn.Prepare(`INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, NAME, VAL, VISIT_DATE) 
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
	_, err := conn.Exec("drop table GOORA_TEMP_VISIT purge")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
	return nil
}

func createRefCursorProc(conn *sql.DB) error {
	sqlText := `CREATE OR REPLACE PROCEDURE GOORA_TEMP_GET_VISIT
(
	LVISIT_ID IN NUMBER,
	L_CURSOR OUT SYS_REFCURSOR
) AS
BEGIN
	OPEN L_CURSOR FOR SELECT VISIT_ID, NAME, VAL, VISIT_DATE FROM GOORA_TEMP_VISIT WHERE VISIT_ID > LVISIT_ID;
END GOORA_TEMP_GET_VISIT;`
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
	_, err := conn.Exec("DROP PROCEDURE GOORA_TEMP_GET_VISIT")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop refcursor procedure: ", time.Now().Sub(t))
	return nil
}

func usage() {
	fmt.Println()
	fmt.Println("refcursor")
	fmt.Println("  a complete code of using refcursor by create temporary table GOORA_TEMP_VISIT and procedure GOORA_TEMP_GET_VISIT.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  refcursor -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  refcursor -server "oracle://user:pass@server/service_name"`)
	fmt.Println()
}

func queryCursor(cursor *go_ora.RefCursor) error {
	t := time.Now()
	rows, err := cursor.Query()
	if err != nil {
		return err
	}
	var (
		id   int64
		name string
		val  float64
		date time.Time
	)

	for rows.Next_() {
		err = rows.Scan(&id, &name, &val, &date)
		if err != nil {
			return err
		}
		fmt.Println("ID: ", id, "\tName: ", name, "\tval: ", val, "\tDate: ", date)
	}
	fmt.Println("Finish query RefCursor: ", time.Now().Sub(t))
	return nil
}

func main() {
	var (
		server string
	)

	flag.StringVar(&server, "server", "", "Server's URL, oracle://user:pass@server/service_name")
	flag.Parse()

	connStr := os.ExpandEnv(server)
	if connStr == "" {
		fmt.Println("Missing -server option")
		usage()
		os.Exit(1)
	}
	fmt.Println("Connection string: ", connStr)
	conn, err := sql.Open("oracle", server)
	if err != nil {
		fmt.Println("Can't open the driver", err)
		return
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection", err)
		}
	}()

	err = conn.Ping()
	if err != nil {
		fmt.Println("Can't ping connection", err)
		return
	}

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
	var cursor go_ora.RefCursor
	_, err = conn.Exec(`BEGIN GOORA_TEMP_GET_VISIT(:1, :2); END;`, 1, sql.Out{Dest: &cursor})
	if err != nil {
		fmt.Println("Can't call refcursor procedure", err)
		return
	}

	defer func() {
		err = cursor.Close()
		if err != nil {
			fmt.Println("Can't close RefCursor", err)
		}
	}()

	err = queryCursor(&cursor)
	if err != nil {
		fmt.Println("Can't query RefCursor", err)
	}
}
