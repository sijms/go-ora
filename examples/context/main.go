package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/sijms/go-ora/v2"
	"os"
	"time"
)

func main() {
	var (
		server string
	)
	flag.StringVar(&server, "server", "", "Server's URL, oracle://user:pass@server/service_name")
	flag.Parse()

	connStr := os.ExpandEnv(server)
	if connStr == "" {
		fmt.Println("Missing -server option")
		//usage()
		os.Exit(1)
	}
	fmt.Println("Connection string: ", connStr)
	db, err := sql.Open("oracle", connStr)
	if err != nil {
		fmt.Println("Can't open database: ", err)
		return
	}

	defer func() {
		err = db.Close()
		if err != nil {
			fmt.Println("Can't close database: ", err)
		}
	}()
	connectCtx, connectCancel := context.WithTimeout(context.Background(), time.Second*3)
	defer connectCancel()
	conn, err := db.Conn(connectCtx)
	if err != nil {
		fmt.Println("Can't open connection: ", err)
		return
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			fmt.Println("Can't close connection: ", conn)
		}
	}()
	execCtx, execCancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer execCancel()
	_, err = conn.ExecContext(execCtx, "begin DBMS_LOCK.sleep(5); end;")
	if err != nil {
		fmt.Println("Can't execute: ", err)
	}
}
