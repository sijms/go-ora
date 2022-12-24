package main

import (
	"database/sql"
	"flag"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

func usage() {
	fmt.Println()
	fmt.Println("inout")
	fmt.Println("  a complete code of create proc, use inout parameters then drop proc.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  inout -server server_url`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  inout -server "oracle://user:pass@server/service_name"`)
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
	conn, err := sql.Open("oracle", connStr)
	if err != nil {
		fmt.Println("Can't create connection: ", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection: ", err)
		}
	}()
	err = createProc(conn)
	if err != nil {
		fmt.Println("Can't create proc: ", err)
		return
	}
	defer func() {
		err = dropProc(conn)
		if err != nil {
			fmt.Println("Can't drop proc: ", err)
		}
	}()
	err = useProc(conn)
	if err != nil {
		fmt.Println("Can't use proc: ", err)
		return
	}
}

func createProc(conn *sql.DB) error {
	t := time.Now()
	sqlText := `CREATE PROCEDURE test_in_out_param (
		io_param IN OUT NVARCHAR2
	) AS
	BEGIN
		IF io_param IS NULL THEN
			io_param := 'changed by procedure because it was empty';
		ELSE
			io_param := io_param || ' ' || io_param;
		END IF;
	END;`
	err := execSql(conn, sqlText)
	fmt.Println("finish create proc: ", time.Now().Sub(t))
	return err
}
func execSql(conn *sql.DB, sqlText string) error {
	_, err := conn.Exec(sqlText)
	return err
}
func dropProc(conn *sql.DB) error {
	t := time.Now()
	err := execSql(conn, "DROP PROCEDURE test_in_out_param")
	fmt.Println("finish drop proc: ", time.Now().Sub(t))
	return err
}

func useProc(conn *sql.DB) error {
	p := "hi github"
	_, err := conn.Exec("BEGIN test_in_out_param(:p); END;", go_ora.Out{Dest: &p, Size: 100, In: true})
	fmt.Println("finish use proc with result: ", p)
	return err
}
