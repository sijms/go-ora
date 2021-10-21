package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"os"
	"strings"
	"time"
)

func dieOnError(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
		os.Exit(1)
	}
}

func createTable(conn *sql.DB) {
	fmt.Println("Creating temporary table GOORA_TEMP_VISIT")
	sqlText := `CREATE TABLE GOORA_TEMP_VISIT(
	VISIT_ID	number(10)	NOT NULL,
	NAME		VARCHAR(200),
	VAL			number(10,2),
	VISIT_DATE	date,
	PRIMARY KEY(VISIT_ID)
	)`
	_, err := conn.Exec(sqlText)
	dieOnError("Cannot create temporary table", err)
}

func insertData(conn *sql.DB) {
	fmt.Println("Inserting values in the table")
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
	fmt.Println("100 Rows inserted")
}

func dropTable(conn *sql.DB) {
	fmt.Println("Dropping table")
	_, err := conn.Exec("drop table GOORA_TEMP_VISIT purge")
	dieOnError("Can't drop table", err)
	fmt.Println("Finish drop table")
}

func usage() {
	fmt.Println()
	fmt.Println("output_par")
	fmt.Println("  a complete code of create table insert, return output parameter then drop table.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  output_par -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  output_par -server "oracle://user:pass@server/service_name"`)
	fmt.Println()
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

	insertData(conn)
	defer dropTable(conn)

	sqlText := `BEGIN
SELECT VISIT_ID, NAME, VAL, VISIT_DATE INTO :1, :2, :3, :4 FROM GOORA_TEMP_VISIT WHERE VISIT_ID = 1;
END;`
	var (
		id   int64
		name string
		val  float64
		date time.Time
	)
	name = strings.Repeat(" ", 200)
	_, err = conn.Exec(sqlText, sql.Out{Dest: &id}, sql.Out{Dest: &name},
		sql.Out{Dest: &val}, sql.Out{Dest: &date})
	dieOnError("Can't get output parameters", err)
	fmt.Println("ID: ", id)
	fmt.Println("Name: ", name)
	fmt.Println("Val: ", val)
	fmt.Println("Date: ", date)
}
