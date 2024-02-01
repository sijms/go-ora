package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/sijms/go-ora/v2"
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

	cancel(conn)
	timeout(conn)

	dbCancel(db)
	dbTimeout(db)
}

func cancel(conn *sql.Conn) {
	execCtx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	_, err := conn.ExecContext(execCtx, "begin DBMS_LOCK.sleep(60); end;")
	if err != nil {
		fmt.Println("Can't execute: ", err)
		if !strings.Contains(err.Error(), "ORA-01013") {
			panic(err)
		}
	}
	if err == nil {
		panic("should not happen")
	}
}

func timeout(conn *sql.Conn) {
	execCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := conn.ExecContext(execCtx, "begin DBMS_LOCK.sleep(60); end;")
	if err != nil {
		fmt.Println("Can't execute: ", err)
		if !strings.Contains(err.Error(), "ORA-01013") {
			panic(err)
		}
	}
	if err == nil {
		panic("should not happen")
	}
}

func dbCancel(db *sql.DB) {
	execCtx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	_, err := db.ExecContext(execCtx, "begin DBMS_LOCK.sleep(60); end;")
	if err != nil {
		fmt.Println("Can't execute: ", err)
		if !strings.Contains(err.Error(), "ORA-01013") {
			panic(err)
		}
	}
	if err == nil {
		panic("should not happen")
	}
}

func dbTimeout(db *sql.DB) {
	execCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := db.ExecContext(execCtx, "begin DBMS_LOCK.sleep(60); end;")
	if err != nil {
		fmt.Println("Can't execute: ", err)
		if !strings.Contains(err.Error(), "ORA-01013") {
			panic(err)
		}
	}
	if err == nil {
		panic("should not happen")
	}
}
