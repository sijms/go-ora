package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"io/ioutil"
	"os"
	"time"
)

func createTable(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE TABLE GOORA_TEMP_VISIT(
	VISIT_ID	number(10)	NOT NULL,
	VISIT_DATA  CLOB,
	PRIMARY KEY(VISIT_ID)
	)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create table GOORA_TEMP_VISIT :", time.Now().Sub(t))
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

func insertData(conn *sql.DB) error {
	t := time.Now()
	val, err := ioutil.ReadFile("clob.json")
	if err != nil {
		return err
	}
	_, err = conn.Exec(`INSERT INTO GOORA_TEMP_VISIT(VISIT_ID, VISIT_DATA) VALUES(1, :1)`, val)
	if err != nil {
		return err
	}
	fmt.Println("1 row inserted: ", time.Now().Sub(t))
	return nil
}
func usage() {
	fmt.Println()
	fmt.Println("clob")
	fmt.Println("  a code for using clob by create table GOORA_TEMP_VISIT then insert then drop")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  clob -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  clob -server "oracle://user:pass@server/service_name"`)
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
		fmt.Println("Can't insert data: ", err)
	}
}
