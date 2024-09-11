package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	go_ora "github.com/sijms/go-ora/v2"
)

type test1 struct {
	Id   int64  `udt:"test_id"`
	Name string `udt:"test_name"`
}

func createUDT(conn *sql.DB) error {
	t := time.Now()
	sqlText := `create or replace TYPE TEST_TYPE1 IS OBJECT 
(
    TEST_ID NUMBER(10, 0),
    TEST_NAME VARCHAR2(10)
)`
	_, err := conn.Exec(sqlText)
	if err != nil {
		return err
	}
	fmt.Println("Finish create UDT: ", time.Now().Sub(t))
	return nil
}

func queryUDT(conn *sql.DB) error {
	t := time.Now()
	var test test1
	err := conn.QueryRow("SELECT TEST_TYPE1(10, 'NAME') FROM DUAL").Scan(&test)
	if err != nil {
		return err
	}
	fmt.Println("Finish query UDT: ", time.Now().Sub(t))
	return nil
}

func dropUDT(conn *sql.DB) error {
	t := time.Now()
	_, err := conn.Exec("drop type TEST_TYPE1")
	if err != nil {
		return err
	}
	fmt.Println("Finish drop UDT: ", time.Now().Sub(t))
	return nil
}

func usage() {
	fmt.Println()
	fmt.Println("user defined type")
	fmt.Println("  a complete code of user defined type, query then drop it.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  UDT -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  UDT -server "oracle://user:pass@server/service_name"`)
	fmt.Println()
}

func main() {
	var server string

	flag.StringVar(&server, "server", "", "Server's URL, oracle://user:pass@server/service_name")
	flag.Parse()

	connStr := os.ExpandEnv(server)
	if connStr == "" {
		fmt.Println("Missing -server option")
		usage()
		os.Exit(1)
	}
	fmt.Println("Connection string: ", connStr)
	conn, err := sql.Open("oracle", connStr)
	if err != nil {
		fmt.Println("Can't open driver", err)
		return
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection", err)
		}
	}()
	err = createUDT(conn)
	if err != nil {
		fmt.Println("Can't create UDT", err)
		return
	}
	defer func() {
		err = dropUDT(conn)
		if err != nil {
			fmt.Println("Can't drop UDT", err)
		}
	}()
	err = go_ora.RegisterType(conn, "TEST_TYPE1", "", test1{})
	if err != nil {
		fmt.Println("Can't register UDT", err)
		return
	}

	fmt.Println("UDT registered successfully")
	err = queryUDT(conn)
	if err != nil {
		fmt.Println("Can't query UDT", err)
		return
	}
}
