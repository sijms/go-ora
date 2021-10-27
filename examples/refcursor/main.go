package main

import (
	"database/sql"
	"flag"
	"fmt"

	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

func dieOnError(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
		os.Exit(1)
	}
}

func createTable(conn *sql.DB) {
	t := time.Now()
	sqlText := `CREATE TABLE GOORA_TEMP_VISIT(
	VISIT_ID	number(10)	NOT NULL,
	NAME		VARCHAR(200),
	VAL			number(10,2),
	VISIT_DATE	date,
	PRIMARY KEY(VISIT_ID)
	)`
	_, err := conn.Exec(sqlText)
	dieOnError("Cannot create temporary table", err)
	fmt.Println("Finish create table GOORA_TEMP_VISIT :", time.Now().Sub(t))
}

func insertData(conn *sql.DB) {
	t := time.Now()
	index := 1
	stmt, err := conn.Prepare(`INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, NAME, VAL, VISIT_DATE) 
VALUES(:1, :2, :3, :4)`)
	dieOnError("Cannot prepare stmt for insert", err)
	defer func() {
		_ = stmt.Close()
	}()
	nameText := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	val := 1.1
	for index = 1; index <= 100; index++ {
		_, err = stmt.Exec(index, nameText, val, time.Now())
		errorText := fmt.Sprintf("Error during insert at index: %d", index)
		dieOnError(errorText, err)
		val += 1.1
	}
	fmt.Println("100 Rows inserted: ", time.Now().Sub(t))
}

func dropTable(conn *sql.DB) {
	t := time.Now()
	_, err := conn.Exec("drop table GOORA_TEMP_VISIT purge")
	dieOnError("Can't drop table", err)
	fmt.Println("Finish drop table: ", time.Now().Sub(t))
}

func createRefCursorProc(conn *sql.DB) {
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
	dieOnError("Can't create refcursor procedure", err)
	fmt.Println("Finish create refcursor procedure: ", time.Now().Sub(t))
}

func dropRefCursorProc(conn *sql.DB) {
	t := time.Now()
	_, err := conn.Exec("DROP PROCEDURE GOORA_TEMP_GET_VISIT")
	dieOnError("Can't drop refcursor procedure", err)
	fmt.Println("Finish drop refcursor procedure: ", time.Now().Sub(t))
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

func queryCursor(cursor *go_ora.RefCursor) {
	rows, err := cursor.Query()
	dieOnError("Can't query cursor", err)
	var (
		id   int64
		name string
		val  float64
		date time.Time
	)

	for rows.Next_() {
		err = rows.Scan(&id, &name, &val, &date)
		dieOnError("Can't Scan row in cursor", err)
		fmt.Println("ID: ", id, "\tName: ", name, "\tval: ", val, "\tDate: ", date)
	}
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
	dieOnError("Can't open the driver:", err)

	defer func() {
		_ = conn.Close()
	}()

	err = conn.Ping()
	dieOnError("Can't ping connection", err)

	createTable(conn)
	defer dropTable(conn)
	insertData(conn)
	createRefCursorProc(conn)
	defer dropRefCursorProc(conn)
	var cursor go_ora.RefCursor
	_, err = conn.Exec(`BEGIN GOORA_TEMP_GET_VISIT(:1, :2); END;`, 1, sql.Out{Dest: &cursor})
	dieOnError("Can't call refcursor procedure", err)
	defer func() {
		_ = cursor.Close()
	}()

	queryCursor(&cursor)
}
