package main

import (
	"flag"
	"fmt"
	go_ora "github.com/sijms/go-ora/v2"
	"os"
)

//func dieOnError(msg string, err error) {
//	if err != nil {
//		fmt.Println(msg, err)
//		os.Exit(1)
//	}
//}

func usage() {
	fmt.Println()
	fmt.Println("exec")
	fmt.Println("  execute DML oracle statement.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println(`  exec -server server_url sql_text`)
	flag.PrintDefaults()
	fmt.Println()
	fmt.Println("Example:")
	fmt.Println(`  exec -server "oracle://user:pass@server/service_name" "update my_table set col_1 = val1 where id = val2"`)
	fmt.Println()
}

func main() {
	var (
		server  string
		sqlText string
	)

	flag.StringVar(&server, "server", "", "Server's URL, oracle://user:pass@server/service_name")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Missing sql text")
		usage()
		os.Exit(1)
	}

	sqlText = flag.Arg(0)
	connStr := os.ExpandEnv(server)
	if connStr == "" {
		fmt.Println("Missing -server option")
		usage()
		os.Exit(1)
	}

	DB, err := go_ora.NewConnection(connStr)
	if err != nil {
		fmt.Println("Can't create new connection", err)
		return
	}
	err = DB.Open()
	if err != nil {
		fmt.Println("Can't open the connection:", err)
		return
	}

	defer func() {
		err = DB.Close()
		if err != nil {
			fmt.Println("Can't close connection", err)
		}
	}()

	stmt := go_ora.NewStmt(sqlText, DB)

	defer func() {
		err = stmt.Close()
		if err != nil {
			fmt.Println("Can't close stmt", err)
		}
	}()

	result, err := stmt.Exec(nil)
	if err != nil {
		fmt.Println("Can't execute sql", err)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Println("rows affected: ", rowsAffected)
}
