package main

import (
	"database/sql/driver"
	"errors"
	"flag"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"io"
	"os"
)

type test1 struct {
	Id   int64  `oracle:"name:test_id"`
	Name string `oracle:"name:test_name"`
}

func dieOnError(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
		os.Exit(1)
	}
}

func createUDT(conn *go_ora.Connection) {
	fmt.Println("creating UDT")
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
	dieOnError("Can't create TEST_TYPE1", err)
	fmt.Println("Finish create UDT")
}

func queryUDT(conn *go_ora.Connection) {
	fmt.Println("Query UDT")
	stmt := go_ora.NewStmt("SELECT TEST_TYPE1(10, 'NAME') FROM DUAL", conn)
	defer func() {
		_ = stmt.Close()
	}()
	rows, err := stmt.Query(nil)
	dieOnError("Can't Query UDT", err)
	var (
		test   test1
		ok     bool
		values = make([]driver.Value, 1)
	)
	for {
		err = rows.Next(values)
		if errors.Is(err, io.EOF) {
			break
		}
		dieOnError("Can't scan rows", err)
		if test, ok = values[0].(test1); !ok {
			dieOnError("Can't convert value to object", errors.New("value conversion error"))
		}
	}
	fmt.Println(test)
	fmt.Println("Finish query UDT")
}
func dropUDT(conn *go_ora.Connection) {
	fmt.Println("dropping UDT")
	stmt := go_ora.NewStmt("drop type TEST_TYPE1", conn)
	defer func() {
		_ = stmt.Close()
	}()
	_, err := stmt.Exec(nil)
	dieOnError("Can't drop UDT", err)
	fmt.Println("Finish drop UDT")
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
	dieOnError("Can't open driver", err)

	err = conn.Open()
	dieOnError("Can't open connection", err)

	defer func() {
		_ = conn.Close()
	}()
	createUDT(conn)
	defer dropUDT(conn)
	err = conn.RegisterType("TEST_TYPE1", test1{})
	dieOnError("Can't register UDT", err)
	fmt.Println("UDT registered successfully")
	queryUDT(conn)
}
