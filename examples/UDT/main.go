package main

import (
	"flag"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

type test1 struct {
	Id   int64  `oracle:"name:test_id"`
	Name string `oracle:"name:test_name"`
}

func createUDT(conn *go_ora.Connection) error {
	t := time.Now()
	sqlText := `create or replace TYPE TEST_TYPE1 IS OBJECT 
(
    TEST_ID NUMBER(10, 0),
    TEST_NAME VARCHAR2(10)
)`
	stmt := go_ora.NewStmt(sqlText, conn)
	defer func() {
		_ = stmt.Close()
	}()
	_, err := stmt.Exec(nil)
	if err != nil {
		return err
	}
	fmt.Println("Finish create UDT: ", time.Now().Sub(t))
	return nil
}

func queryUDT(conn *go_ora.Connection) error {
	t := time.Now()
	stmt := go_ora.NewStmt("SELECT TEST_TYPE1(10, 'NAME') FROM DUAL", conn)
	defer func() {
		_ = stmt.Close()
	}()
	rows, err := stmt.Query(nil)
	if err != nil {
		return err
	}
	var (
		test test1
	)
	if oraRows, ok := rows.(*go_ora.DataSet); ok {
		for oraRows.Next_() {
			err = oraRows.Scan(&test)
			if err != nil {
				return err
			}
		}
		fmt.Println(test)
	}
	fmt.Println("Finish query UDT: ", time.Now().Sub(t))
	return nil
}

func dropUDT(conn *go_ora.Connection) error {
	t := time.Now()
	stmt := go_ora.NewStmt("drop type TEST_TYPE1", conn)
	defer func() {
		_ = stmt.Close()
	}()
	_, err := stmt.Exec(nil)
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
	conn, err := go_ora.NewConnection(connStr)
	if err != nil {
		fmt.Println("Can't open driver", err)
		return
	}

	err = conn.Open()
	if err != nil {
		fmt.Println("Can't open connection", err)
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
	err = conn.RegisterType("TEST_TYPE1", test1{})
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
